package build

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shivamx96/leafpress/cli/internal/assets"
	"github.com/shivamx96/leafpress/cli/internal/config"
	"github.com/shivamx96/leafpress/cli/internal/content"
	"github.com/shivamx96/leafpress/cli/internal/templates"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Options configures the build process
type Options struct {
	IncludeDrafts bool
	Verbose       bool
	SkipClean     bool // Skip cleaning output directory (for hot reload)
}

// Stats contains build statistics
type Stats struct {
	PageCount    int
	WarningCount int
}

// Builder handles site generation
type Builder struct {
	cfg       *config.Config
	opts      Options
	rootDir   string
	outputDir string
	templates *templates.Templates

	// Cached state for incremental builds
	pages          []*content.Page
	pagesByPath    map[string]*content.Page   // SourcePath -> Page
	pagesBySlug    map[string]*content.Page   // Slug -> Page
	pagesBySection map[string][]*content.Page // Section -> Pages (for fast section lookups)
	pagesByTag     map[string][]*content.Page // Tag (lowercase) -> Pages (for fast tag lookups)
	linkResolver   *content.LinkResolver      // Cached link resolver
	siteData       templates.SiteData
}

// New creates a new Builder
func New(cfg *config.Config, opts Options) *Builder {
	cwd, _ := os.Getwd()
	return &Builder{
		cfg:       cfg,
		opts:      opts,
		rootDir:   cwd,
		outputDir: filepath.Join(cwd, cfg.OutputDir),
	}
}

// SetSkipClean enables or disables cleaning the output directory
func (b *Builder) SetSkipClean(skip bool) {
	b.opts.SkipClean = skip
}

// logTiming prints timing info in verbose mode with aligned formatting
func (b *Builder) logTiming(label string, d time.Duration) {
	if b.opts.Verbose {
		fmt.Printf("  %-16s %v\n", label, d.Round(time.Microsecond))
	}
}

// Build generates the static site
func (b *Builder) Build() (*Stats, error) {
	stats := &Stats{}
	var t0 time.Time

	// Initialize templates
	t0 = time.Now()
	var err error
	b.templates, err = templates.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize templates: %w", err)
	}
	b.logTiming("templates", time.Since(t0))

	// Clean output directory (skip for hot reload)
	t0 = time.Now()
	if !b.opts.SkipClean {
		if err := os.RemoveAll(b.outputDir); err != nil {
			return nil, fmt.Errorf("failed to clean output directory: %w", err)
		}
	}
	if err := os.MkdirAll(b.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}
	b.logTiming("clean", time.Since(t0))

	// Scan content
	t0 = time.Now()
	scanner := content.NewScanner(b.rootDir, b.cfg.Ignore)
	pages, err := scanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("failed to scan content: %w", err)
	}
	b.logTiming("scan", time.Since(t0))

	// Filter drafts
	if !b.opts.IncludeDrafts {
		pages = filterDrafts(pages)
	}

	// Build section index for O(1) lookups
	b.pagesBySection = buildSectionIndex(pages)

	// Build tag index for O(1) lookups
	b.pagesByTag = buildTagIndex(pages)

	// Create link resolver once (reused for backlinks, rendering, graph)
	b.linkResolver = content.NewLinkResolver(pages)

	// Build backlinks (if enabled)
	t0 = time.Now()
	if b.cfg.Backlinks {
		content.BuildBacklinks(pages, b.linkResolver)
	}
	b.logTiming("backlinks", time.Since(t0))

	// Extract basePath early for use in rendering
	basePath := extractBasePath(b.cfg.BaseURL)

	// Render markdown to HTML
	t0 = time.Now()
	warnings := content.RenderPages(pages, b.cfg.Wikilinks, b.linkResolver, basePath)
	b.logTiming("markdown", time.Since(t0))
	stats.WarningCount = len(warnings)

	if b.opts.Verbose {
		for _, w := range warnings {
			fmt.Printf("  warning: %s\n", w)
		}
	}

	// Generate site data
	siteData := templates.SiteData{
		Title:       b.cfg.Title,
		Description: b.cfg.Description,
		Author:      b.cfg.Author,
		Nav:         b.cfg.Nav,
		Theme:       b.cfg.Theme,
		BaseURL:     b.cfg.BaseURL,
		BasePath:    basePath,
		Image:       b.cfg.Image,
		TOC:         b.cfg.TOC,
		Graph:       b.cfg.Graph,
		Search:      b.cfg.Search,
		HeadExtra:   b.cfg.HeadExtra,
	}

	// Cache state for incremental builds
	b.pages = pages
	b.siteData = siteData
	b.pagesByPath = make(map[string]*content.Page)
	b.pagesBySlug = make(map[string]*content.Page)
	for _, page := range pages {
		b.pagesByPath[page.SourcePath] = page
		b.pagesBySlug[page.Slug] = page
	}

	// Render pages in parallel
	t0 = time.Now()
	stats.PageCount = len(pages)
	numWorkers := runtime.NumCPU()
	if numWorkers > len(pages) {
		numWorkers = len(pages)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	pageChan := make(chan *content.Page, len(pages))
	errChan := make(chan error, len(pages))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for page := range pageChan {
				var err error
				if page.IsIndex {
					err = b.renderSectionIndex(page, pages, siteData)
				} else {
					err = b.renderPage(page, siteData)
				}
				if err != nil {
					errChan <- fmt.Errorf("failed to render %s: %w", page.SourcePath, err)
				}
			}
		}()
	}

	// Send pages to workers
	for _, page := range pages {
		pageChan <- page
	}
	close(pageChan)

	// Wait for workers to finish
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		return nil, err
	}
	b.logTiming("render", time.Since(t0))

	// Generate auto-indexes for directories without _index.md
	t0 = time.Now()
	if err := b.generateAutoIndexes(pages, siteData); err != nil {
		return nil, fmt.Errorf("failed to generate auto indexes: %w", err)
	}
	b.logTiming("auto-indexes", time.Since(t0))

	// Generate tag pages
	t0 = time.Now()
	if err := b.generateTagPages(pages, siteData); err != nil {
		return nil, fmt.Errorf("failed to generate tag pages: %w", err)
	}
	b.logTiming("tags", time.Since(t0))

	// Copy static files
	t0 = time.Now()
	if err := b.copyStatic(); err != nil {
		return nil, fmt.Errorf("failed to copy static files: %w", err)
	}
	b.logTiming("static", time.Since(t0))

	// Generate CSS
	t0 = time.Now()
	if err := b.generateCSS(); err != nil {
		return nil, fmt.Errorf("failed to generate CSS: %w", err)
	}
	b.logTiming("css", time.Since(t0))

	// Copy favicons
	t0 = time.Now()
	if err := b.copyFavicons(); err != nil {
		return nil, fmt.Errorf("failed to copy favicons: %w", err)
	}
	b.logTiming("favicons", time.Since(t0))

	// Generate graph.json and search-index.json if enabled
	if b.cfg.Graph || b.cfg.Search {
		t0 = time.Now()
		if err := b.generateJSONFiles(pages, b.cfg.Graph, b.cfg.Search); err != nil {
			return nil, fmt.Errorf("failed to generate JSON files: %w", err)
		}
		b.logTiming("json", time.Since(t0))
	}

	// Generate robots.txt
	t0 = time.Now()
	if err := b.generateRobotsTxt(); err != nil {
		return nil, fmt.Errorf("failed to generate robots.txt: %w", err)
	}
	b.logTiming("robots.txt", time.Since(t0))

	// Generate sitemap.xml
	t0 = time.Now()
	if err := b.generateSitemap(pages); err != nil {
		return nil, fmt.Errorf("failed to generate sitemap.xml: %w", err)
	}
	b.logTiming("sitemap", time.Since(t0))

	// Generate 404.html
	t0 = time.Now()
	if err := b.generate404(siteData); err != nil {
		return nil, fmt.Errorf("failed to generate 404.html: %w", err)
	}
	b.logTiming("404", time.Since(t0))

	// Generate RSS feed
	t0 = time.Now()
	if err := b.generateRSS(pages, siteData); err != nil {
		return nil, fmt.Errorf("failed to generate RSS feed: %w", err)
	}
	b.logTiming("rss", time.Since(t0))

	return stats, nil
}

// ChangeType represents the type of file change
type ChangeType int

const (
	ChangeModify ChangeType = iota
	ChangeCreate
	ChangeDelete
)

// IncrementalStats contains incremental build statistics
type IncrementalStats struct {
	PagesRebuilt int
	TagsRebuilt  int
	FullRebuild  bool
}

// RebuildIncremental performs an incremental rebuild based on changed file
func (b *Builder) RebuildIncremental(changedPath string, changeType ChangeType) (*IncrementalStats, error) {
	stats := &IncrementalStats{}
	var t0 time.Time

	// If no cached state, do full rebuild
	if b.pages == nil {
		if _, err := b.Build(); err != nil {
			return nil, err
		}
		stats.FullRebuild = true
		return stats, nil
	}

	// Get relative path
	relPath, err := filepath.Rel(b.rootDir, changedPath)
	if err != nil {
		relPath = changedPath
	}

	// Check if it's a config change - requires full rebuild with fresh config
	if filepath.Base(relPath) == "leafpress.json" {
		// Reload config from disk
		newCfg, err := config.Load("leafpress.json")
		if err != nil {
			return nil, fmt.Errorf("failed to reload config: %w", err)
		}
		b.cfg = newCfg
		b.outputDir = filepath.Join(b.rootDir, b.cfg.OutputDir)

		// Regenerate templates
		newTemplates, err := templates.New()
		if err != nil {
			return nil, fmt.Errorf("failed to reload templates: %w", err)
		}
		b.templates = newTemplates

		b.opts.SkipClean = false // Full clean for config changes
		if _, err := b.Build(); err != nil {
			return nil, err
		}
		stats.FullRebuild = true
		return stats, nil
	}

	// Check if it's a static file (cross-platform: handle both / and \ separators)
	isStaticFile := strings.HasPrefix(relPath, "static/") || strings.HasPrefix(relPath, "static"+string(filepath.Separator))
	if isStaticFile {
		t0 = time.Now()
		if err := b.copyStatic(); err != nil {
			return nil, err
		}
		b.logTiming("static", time.Since(t0))
		return stats, nil
	}

	// Check if it's a CSS file
	if relPath == "style.css" {
		t0 = time.Now()
		if err := b.generateCSS(); err != nil {
			return nil, err
		}
		b.logTiming("css", time.Since(t0))
		return stats, nil
	}

	// Handle markdown file changes
	if filepath.Ext(relPath) == ".md" {
		return b.rebuildMarkdownFile(relPath, changeType)
	}

	return stats, nil
}

// rebuildMarkdownFile handles incremental rebuild for a markdown file change
func (b *Builder) rebuildMarkdownFile(relPath string, changeType ChangeType) (*IncrementalStats, error) {
	stats := &IncrementalStats{}
	var t0 time.Time

	// For deletions, remove the output file and rebuild affected pages
	if changeType == ChangeDelete {
		return b.handleDeletedFile(relPath)
	}

	// Check if file is in an ignored folder
	topLevel := strings.Split(relPath, string(filepath.Separator))[0]
	for _, ignored := range b.cfg.Ignore {
		if topLevel == ignored {
			return stats, nil // File is in ignored folder, skip
		}
	}

	// Parse only the changed file (not full scan)
	t0 = time.Now()
	changedPage, err := content.ParseSingleFile(b.rootDir, relPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", relPath, err)
	}

	// Skip drafts if needed
	if !b.opts.IncludeDrafts && changedPage.Draft {
		// If it was previously not a draft but now is, treat as deletion
		if oldPage := b.pagesByPath[relPath]; oldPage != nil {
			return b.handleDeletedFile(relPath)
		}
		return stats, nil
	}
	b.logTiming("parse", time.Since(t0))

	// Get the old page if it existed
	oldPage := b.pagesByPath[relPath]

	// Update the pages cache with the new/changed page
	if oldPage != nil {
		// Remove old slug mapping if slug changed
		if oldPage.Slug != changedPage.Slug {
			delete(b.pagesBySlug, oldPage.Slug)
		}
		// Replace old page with new one in the slice
		for i, p := range b.pages {
			if p.SourcePath == relPath {
				b.pages[i] = changedPage
				break
			}
		}
	} else {
		// Add new page to the slice
		b.pages = append(b.pages, changedPage)
	}
	b.pagesByPath[relPath] = changedPage
	b.pagesBySlug[changedPage.Slug] = changedPage

	// Determine what needs rebuilding
	pagesToRebuild := make(map[string]*content.Page)
	tagsToRebuild := make(map[string]bool)
	rebuildSectionIndex := false
	var sectionSlug string

	pagesToRebuild[changedPage.SourcePath] = changedPage

	// Get section for this page
	sectionSlug = filepath.Dir(changedPage.Slug)
	if sectionSlug == "." {
		sectionSlug = ""
	}

	// If old page existed, check what changed
	if oldPage != nil {
		// Rebuild pages that had backlinks to this page (their backlinks section changed)
		for _, backlinker := range oldPage.Backlinks {
			pagesToRebuild[backlinker.SourcePath] = backlinker
		}

		// If tags changed, rebuild affected tag pages
		oldTags := make(map[string]bool)
		for _, t := range oldPage.Tags {
			oldTags[strings.ToLower(t)] = true
		}
		for _, t := range changedPage.Tags {
			tLower := strings.ToLower(t)
			if !oldTags[tLower] {
				tagsToRebuild[tLower] = true // New tag
			}
			delete(oldTags, tLower)
		}
		for t := range oldTags {
			tagsToRebuild[t] = true // Removed tag
		}
	} else {
		// New file - rebuild section index
		rebuildSectionIndex = true
		// All tags are new
		for _, t := range changedPage.Tags {
			tagsToRebuild[strings.ToLower(t)] = true
		}
	}

	// Update the cached resolver with current pages
	b.linkResolver = content.NewLinkResolver(b.pages)

	// Rebuild backlinks with updated page set and fresh resolver
	t0 = time.Now()
	if b.cfg.Backlinks {
		content.BuildBacklinks(b.pages, b.linkResolver)
	}
	b.logTiming("backlinks", time.Since(t0))

	// If the changed page has new outlinks, rebuild pages it now links to
	for _, target := range changedPage.OutLinks {
		result := b.linkResolver.Resolve(target)
		if result.Page != nil {
			pagesToRebuild[result.Page.SourcePath] = result.Page
		}
	}

	// Render markdown for pages that need rebuilding
	t0 = time.Now()
	var pagesToRender []*content.Page
	for _, p := range pagesToRebuild {
		// Find the updated version from b.pages
		for _, np := range b.pages {
			if np.SourcePath == p.SourcePath {
				pagesToRender = append(pagesToRender, np)
				break
			}
		}
	}
	content.RenderPages(pagesToRender, b.cfg.Wikilinks, b.linkResolver, b.siteData.BasePath)
	b.logTiming("markdown", time.Since(t0))

	// Render the affected pages
	t0 = time.Now()
	for _, page := range pagesToRender {
		if page.IsIndex {
			if err := b.renderSectionIndex(page, b.pages, b.siteData); err != nil {
				return nil, err
			}
		} else {
			if err := b.renderPage(page, b.siteData); err != nil {
				return nil, err
			}
		}
		stats.PagesRebuilt++
	}
	b.logTiming("render", time.Since(t0))

	// Rebuild section index if needed
	if rebuildSectionIndex && sectionSlug != "" {
		t0 = time.Now()
		// Check if there's a manual _index.md
		hasManualIndex := false
		for _, p := range b.pages {
			if p.IsIndex && p.Slug == sectionSlug {
				hasManualIndex = true
				break
			}
		}
		if !hasManualIndex {
			if err := b.rebuildAutoIndex(sectionSlug, b.pages); err != nil {
				return nil, err
			}
		}
		b.logTiming("auto-index", time.Since(t0))
	}

	// Rebuild affected tag pages
	if len(tagsToRebuild) > 0 {
		t0 = time.Now()
		if err := b.rebuildTagPages(tagsToRebuild, b.pages); err != nil {
			return nil, err
		}
		stats.TagsRebuilt = len(tagsToRebuild)
		b.logTiming("tags", time.Since(t0))
	}

	// Regenerate JSON files if enabled
	if b.cfg.Graph || b.cfg.Search {
		t0 = time.Now()
		if err := b.generateJSONFiles(b.pages, b.cfg.Graph, b.cfg.Search); err != nil {
			return nil, err
		}
		b.logTiming("json", time.Since(t0))
	}

	return stats, nil
}

// handleDeletedFile handles removal of a markdown file
func (b *Builder) handleDeletedFile(relPath string) (*IncrementalStats, error) {
	stats := &IncrementalStats{}

	oldPage := b.pagesByPath[relPath]
	if oldPage == nil {
		return stats, nil // File wasn't tracked
	}

	// Remove the output HTML file
	outPath := filepath.Join(b.outputDir, oldPage.OutputPath)
	os.Remove(outPath)
	// Also try to remove the parent directory if empty
	os.Remove(filepath.Dir(outPath))

	// Rebuild pages that had backlinks to this page
	pagesToRebuild := make([]*content.Page, 0)
	for _, backlinker := range oldPage.Backlinks {
		pagesToRebuild = append(pagesToRebuild, backlinker)
	}

	// Remove from cached state
	delete(b.pagesByPath, relPath)
	delete(b.pagesBySlug, oldPage.Slug)

	// Filter out the deleted page from pages slice
	newPages := make([]*content.Page, 0, len(b.pages)-1)
	for _, p := range b.pages {
		if p.SourcePath != relPath {
			newPages = append(newPages, p)
		}
	}
	b.pages = newPages

	// Update resolver and rebuild backlinks
	b.linkResolver = content.NewLinkResolver(b.pages)
	if b.cfg.Backlinks {
		content.BuildBacklinks(b.pages, b.linkResolver)
	}

	// Re-render affected pages
	content.RenderPages(pagesToRebuild, b.cfg.Wikilinks, b.linkResolver, b.siteData.BasePath)
	for _, page := range pagesToRebuild {
		if page.IsIndex {
			if err := b.renderSectionIndex(page, b.pages, b.siteData); err != nil {
				return nil, err
			}
		} else {
			if err := b.renderPage(page, b.siteData); err != nil {
				return nil, err
			}
		}
		stats.PagesRebuilt++
	}

	// Rebuild tag pages that contained this page
	tagsToRebuild := make(map[string]bool)
	for _, t := range oldPage.Tags {
		tagsToRebuild[strings.ToLower(t)] = true
	}
	if len(tagsToRebuild) > 0 {
		if err := b.rebuildTagPages(tagsToRebuild, b.pages); err != nil {
			return nil, err
		}
		stats.TagsRebuilt = len(tagsToRebuild)
	}

	// Rebuild section index
	sectionSlug := filepath.Dir(oldPage.Slug)
	if sectionSlug != "." && sectionSlug != "" {
		if err := b.rebuildAutoIndex(sectionSlug, b.pages); err != nil {
			return nil, err
		}
	}

	// Regenerate JSON files
	if b.cfg.Graph || b.cfg.Search {
		if err := b.generateJSONFiles(b.pages, b.cfg.Graph, b.cfg.Search); err != nil {
			return nil, err
		}
	}

	return stats, nil
}

// rebuildAutoIndex rebuilds a single auto-generated index
func (b *Builder) rebuildAutoIndex(sectionSlug string, pages []*content.Page) error {
	sectionPages := b.getSectionPagesFromIndex(sectionSlug)
	sortPages(sectionPages, "date")

	outPath := filepath.Join(b.outputDir, sectionSlug, "index.html")
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	title := cases.Title(language.English).String(filepath.Base(sectionSlug))
	data := templates.IndexData{
		Site:        b.siteData,
		Title:       title,
		Pages:       sectionPages,
		ShowList:    true,
		CurrentPath: "/" + sectionSlug + "/",
	}

	return b.templates.RenderIndex(f, data)
}

// rebuildTagPages rebuilds specific tag pages
func (b *Builder) rebuildTagPages(tags map[string]bool, pages []*content.Page) error {
	// Rebuild the tag index (pages may have changed)
	b.pagesByTag = buildTagIndex(pages)

	tagsDir := filepath.Join(b.outputDir, "tags")

	for tag := range tags {
		tagDir := filepath.Join(tagsDir, tag)
		pagesForTag := b.pagesByTag[tag]

		if len(pagesForTag) == 0 {
			// Tag no longer has any pages, remove it
			os.RemoveAll(tagDir)
			continue
		}

		sortPages(pagesForTag, "date")

		if err := os.MkdirAll(tagDir, 0755); err != nil {
			return err
		}

		tagPath := filepath.Join(tagDir, "index.html")
		f, err := os.Create(tagPath)
		if err != nil {
			return err
		}

		if err := b.templates.RenderTagPage(f, templates.TagPageData{
			Site:        b.siteData,
			Tag:         tag,
			Pages:       pagesForTag,
			CurrentPath: "/tags/" + tag + "/",
		}); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	// Rebuild tag index using cached pagesByTag
	var allTags []templates.TagInfo
	for tag, taggedPages := range b.pagesByTag {
		allTags = append(allTags, templates.TagInfo{
			Name:  tag,
			Count: len(taggedPages),
		})
	}
	sort.Slice(allTags, func(i, j int) bool {
		return allTags[i].Name < allTags[j].Name
	})

	if err := os.MkdirAll(tagsDir, 0755); err != nil {
		return err
	}

	indexPath := filepath.Join(tagsDir, "index.html")
	f, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return b.templates.RenderTagIndex(f, templates.TagIndexData{
		Site:        b.siteData,
		Tags:        allTags,
		CurrentPath: "/tags/",
	})
}

// renderPage renders a single content page
func (b *Builder) renderPage(page *content.Page, siteData templates.SiteData) error {
	outPath := filepath.Join(b.outputDir, page.OutputPath)

	if b.opts.Verbose {
		fmt.Printf("  writing: %s\n", outPath)
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	// Create output file
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Extract TOC if enabled (check page override first, then site default)
	var toc []templates.TOCItem
	htmlContent := page.HTMLContent
	showTOC := siteData.TOC
	if page.TOC != nil {
		showTOC = *page.TOC
	}
	if showTOC {
		htmlContent, toc = templates.ExtractTOC(page.HTMLContent)
	}

	// Render template
	data := templates.PageData{
		Site:        siteData,
		Page:        page,
		Content:     template.HTML(htmlContent),
		TOC:         toc,
		CurrentPath: page.Permalink,
	}

	return b.templates.RenderPage(f, data)
}

// renderSectionIndex renders a section index page
func (b *Builder) renderSectionIndex(indexPage *content.Page, allPages []*content.Page, siteData templates.SiteData) error {
	// Get pages in this section
	sectionPages := b.getSectionPagesFromIndex(indexPage.Slug)

	// Sort pages
	sortPages(sectionPages, indexPage.SectionSort)

	outPath := filepath.Join(b.outputDir, indexPage.OutputPath)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Determine if we should show the list (default true if not specified)
	showList := true
	if indexPage.ShowList != nil {
		showList = *indexPage.ShowList
	}

	currentPath := "/" + indexPage.Slug
	if currentPath != "/" {
		currentPath += "/"
	}
	data := templates.IndexData{
		Site:        siteData,
		Title:       indexPage.Title,
		Pages:       sectionPages,
		Intro:       template.HTML(indexPage.HTMLContent),
		ShowList:    showList,
		CurrentPath: currentPath,
	}

	return b.templates.RenderIndex(f, data)
}

// generateAutoIndexes creates index pages for directories without _index.md
func (b *Builder) generateAutoIndexes(pages []*content.Page, siteData templates.SiteData) error {
	// Find all directories
	dirs := make(map[string]bool)
	indexedDirs := make(map[string]bool)

	for _, page := range pages {
		if page.IsIndex {
			indexedDirs[page.Slug] = true
		} else {
			dir := filepath.Dir(page.Slug)
			if dir != "." {
				dirs[dir] = true
			}
		}
	}

	// Collect directories that need auto-indexes
	var dirsToIndex []string
	for dir := range dirs {
		if !indexedDirs[dir] {
			dirsToIndex = append(dirsToIndex, dir)
		}
	}

	if len(dirsToIndex) == 0 {
		return nil
	}

	// Generate indexes in parallel
	numWorkers := runtime.NumCPU()
	if numWorkers > len(dirsToIndex) {
		numWorkers = len(dirsToIndex)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	dirChan := make(chan string, len(dirsToIndex))
	errChan := make(chan error, len(dirsToIndex))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dir := range dirChan {
				sectionPages := b.getSectionPagesFromIndex(dir)
				sortPages(sectionPages, "date")

				outPath := filepath.Join(b.outputDir, dir, "index.html")
				if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
					errChan <- err
					continue
				}

				f, err := os.Create(outPath)
				if err != nil {
					errChan <- err
					continue
				}

				title := cases.Title(language.English).String(filepath.Base(dir))
				data := templates.IndexData{
					Site:        siteData,
					Title:       title,
					Pages:       sectionPages,
					ShowList:    true,
					CurrentPath: "/" + dir + "/",
				}

				if err := b.templates.RenderIndex(f, data); err != nil {
					f.Close()
					errChan <- err
					continue
				}
				f.Close()
			}
		}()
	}

	for _, dir := range dirsToIndex {
		dirChan <- dir
	}
	close(dirChan)

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}

	return nil
}

// generateTagPages creates tag index and individual tag pages
func (b *Builder) generateTagPages(pages []*content.Page, siteData templates.SiteData) error {
	// Use cached tag index (already built during Build)
	tagPages := b.pagesByTag
	if tagPages == nil {
		// Fallback: build tag index if not cached
		tagPages = buildTagIndex(pages)
	}

	if len(tagPages) == 0 {
		return nil
	}

	// Create tags directory
	tagsDir := filepath.Join(b.outputDir, "tags")
	if err := os.MkdirAll(tagsDir, 0755); err != nil {
		return err
	}

	// Generate tag index
	var tags []templates.TagInfo
	for tag, pages := range tagPages {
		tags = append(tags, templates.TagInfo{
			Name:  tag,
			Count: len(pages),
		})
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})

	indexPath := filepath.Join(tagsDir, "index.html")
	f, err := os.Create(indexPath)
	if err != nil {
		return err
	}

	if err := b.templates.RenderTagIndex(f, templates.TagIndexData{
		Site:        siteData,
		Tags:        tags,
		CurrentPath: "/tags/",
	}); err != nil {
		f.Close()
		return err
	}
	f.Close()

	// Generate individual tag pages in parallel
	type tagJob struct {
		tag   string
		pages []*content.Page
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > len(tagPages) {
		numWorkers = len(tagPages)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	jobChan := make(chan tagJob, len(tagPages))
	errChan := make(chan error, len(tagPages))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				sortPages(job.pages, "date")

				tagDir := filepath.Join(tagsDir, job.tag)
				if err := os.MkdirAll(tagDir, 0755); err != nil {
					errChan <- err
					continue
				}

				tagPath := filepath.Join(tagDir, "index.html")
				f, err := os.Create(tagPath)
				if err != nil {
					errChan <- err
					continue
				}

				if err := b.templates.RenderTagPage(f, templates.TagPageData{
					Site:        siteData,
					Tag:         job.tag,
					Pages:       job.pages,
					CurrentPath: "/tags/" + job.tag + "/",
				}); err != nil {
					f.Close()
					errChan <- err
					continue
				}
				f.Close()
			}
		}()
	}

	// Send jobs
	for tag, pages := range tagPages {
		jobChan <- tagJob{tag: tag, pages: pages}
	}
	close(jobChan)

	// Wait for workers
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		return err
	}

	return nil
}

// copyStatic copies the static directory
func (b *Builder) copyStatic() error {
	srcDir := filepath.Join(b.rootDir, "static")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil // No static directory
	}

	dstDir := filepath.Join(b.outputDir, "static")
	return copyDir(srcDir, dstDir)
}

// copyFavicons copies favicons from user directory or uses embedded defaults
func (b *Builder) copyFavicons() error {
	favicons := []string{"favicon.ico", "favicon.svg", "favicon-96x96.png"}

	for _, name := range favicons {
		userPath := filepath.Join(b.rootDir, name)
		outPath := filepath.Join(b.outputDir, name)

		// Check if user has provided their own favicon
		if data, err := os.ReadFile(userPath); err == nil {
			// Use user's favicon
			if err := os.WriteFile(outPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", name, err)
			}
		} else {
			// Use embedded default favicon
			var defaultData []byte
			switch name {
			case "favicon.ico":
				defaultData = assets.FaviconICO
			case "favicon.svg":
				defaultData = assets.FaviconSVG
			case "favicon-96x96.png":
				defaultData = assets.FaviconPNG
			}
			if err := os.WriteFile(outPath, defaultData, 0644); err != nil {
				return fmt.Errorf("failed to write default %s: %w", name, err)
			}
		}
	}

	return nil
}

// generateCSS writes the combined stylesheet
func (b *Builder) generateCSS() error {
	// Start with default CSS
	css := templates.DefaultCSS

	// Append user CSS if exists
	userCSS := filepath.Join(b.rootDir, "style.css")
	if data, err := os.ReadFile(userCSS); err == nil {
		css += "\n\n/* User Styles */\n" + string(data)
	}

	// Write combined CSS
	outPath := filepath.Join(b.outputDir, "style.css")
	return os.WriteFile(outPath, []byte(css), 0644)
}

// generateRobotsTxt writes the robots.txt file
func (b *Builder) generateRobotsTxt() error {
	var content string
	if b.cfg.BaseURL != "" {
		content = fmt.Sprintf("User-agent: *\nAllow: /\n\nSitemap: %s/sitemap.xml\n", strings.TrimSuffix(b.cfg.BaseURL, "/"))
	} else {
		content = "User-agent: *\nAllow: /\n"
	}
	outPath := filepath.Join(b.outputDir, "robots.txt")
	return os.WriteFile(outPath, []byte(content), 0644)
}

// generateSitemap writes the sitemap.xml file
func (b *Builder) generateSitemap(pages []*content.Page) error {
	baseURL := strings.TrimSuffix(b.cfg.BaseURL, "/")

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	sb.WriteString("\n")

	for _, page := range pages {
		loc := page.Permalink
		if baseURL != "" {
			loc = baseURL + page.Permalink
		}

		// Use modified date if available, otherwise use created date
		var lastmod string
		if page.HasModified() {
			lastmod = page.Modified.Format("2006-01-02")
		} else if !page.Date.IsZero() {
			lastmod = page.Date.Format("2006-01-02")
		}

		sb.WriteString("  <url>\n")
		sb.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", loc))
		if lastmod != "" {
			sb.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", lastmod))
		}
		sb.WriteString("  </url>\n")
	}

	sb.WriteString("</urlset>\n")

	outPath := filepath.Join(b.outputDir, "sitemap.xml")
	return os.WriteFile(outPath, []byte(sb.String()), 0644)
}

// generate404 writes the 404.html file
func (b *Builder) generate404(siteData templates.SiteData) error {
	outPath := filepath.Join(b.outputDir, "404.html")
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return b.templates.RenderNotFound(f, templates.NotFoundData{
		Site: siteData,
	})
}

// generateRSS writes the feed.xml file
func (b *Builder) generateRSS(pages []*content.Page, siteData templates.SiteData) error {
	baseURL := strings.TrimSuffix(b.cfg.BaseURL, "/")

	// Sort pages by date (newest first) and filter out index pages
	var feedPages []*content.Page
	for _, p := range pages {
		if !p.IsIndex {
			feedPages = append(feedPages, p)
		}
	}
	sort.Slice(feedPages, func(i, j int) bool {
		dateI := feedPages[i].Date
		if feedPages[i].HasModified() {
			dateI = feedPages[i].Modified
		}
		dateJ := feedPages[j].Date
		if feedPages[j].HasModified() {
			dateJ = feedPages[j].Modified
		}
		return dateI.After(dateJ)
	})

	// Limit to 20 most recent items
	if len(feedPages) > 20 {
		feedPages = feedPages[:20]
	}

	// Build date for the feed (most recent item date)
	var lastBuildDate string
	if len(feedPages) > 0 {
		p := feedPages[0]
		if p.HasModified() {
			lastBuildDate = p.Modified.Format(time.RFC1123Z)
		} else if !p.Date.IsZero() {
			lastBuildDate = p.Date.Format(time.RFC1123Z)
		} else {
			lastBuildDate = time.Now().Format(time.RFC1123Z)
		}
	} else {
		lastBuildDate = time.Now().Format(time.RFC1123Z)
	}

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">`)
	sb.WriteString("\n")
	sb.WriteString("  <channel>\n")
	sb.WriteString(fmt.Sprintf("    <title>%s</title>\n", escapeXML(siteData.Title)))
	if baseURL != "" {
		sb.WriteString(fmt.Sprintf("    <link>%s</link>\n", baseURL))
		sb.WriteString(fmt.Sprintf("    <atom:link href=\"%s/feed.xml\" rel=\"self\" type=\"application/rss+xml\"/>\n", baseURL))
	}
	if siteData.Author != "" {
		sb.WriteString(fmt.Sprintf("    <description>%s's digital garden</description>\n", escapeXML(siteData.Author)))
	} else {
		sb.WriteString(fmt.Sprintf("    <description>%s</description>\n", escapeXML(siteData.Title)))
	}
	sb.WriteString(fmt.Sprintf("    <lastBuildDate>%s</lastBuildDate>\n", lastBuildDate))
	sb.WriteString("    <generator>leafpress</generator>\n")

	for _, page := range feedPages {
		link := page.Permalink
		if baseURL != "" {
			link = baseURL + page.Permalink
		}

		var pubDate string
		if page.HasModified() {
			pubDate = page.Modified.Format(time.RFC1123Z)
		} else if !page.Date.IsZero() {
			pubDate = page.Date.Format(time.RFC1123Z)
		}

		sb.WriteString("    <item>\n")
		sb.WriteString(fmt.Sprintf("      <title>%s</title>\n", escapeXML(page.Title)))
		sb.WriteString(fmt.Sprintf("      <link>%s</link>\n", link))
		sb.WriteString(fmt.Sprintf("      <guid>%s</guid>\n", link))
		if pubDate != "" {
			sb.WriteString(fmt.Sprintf("      <pubDate>%s</pubDate>\n", pubDate))
		}
		// Use plain text content as description (truncated)
		desc := page.PlainContent()
		if len(desc) > 300 {
			desc = desc[:300] + "..."
		}
		if desc != "" {
			sb.WriteString(fmt.Sprintf("      <description>%s</description>\n", escapeXML(desc)))
		}
		sb.WriteString("    </item>\n")
	}

	sb.WriteString("  </channel>\n")
	sb.WriteString("</rss>\n")

	outPath := filepath.Join(b.outputDir, "feed.xml")
	return os.WriteFile(outPath, []byte(sb.String()), 0644)
}

// escapeXML escapes special characters for XML
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// generateJSONFiles creates graph.json and search-index.json in a single pass
func (b *Builder) generateJSONFiles(pages []*content.Page, genGraph, genSearch bool) error {
	if !genGraph && !genSearch {
		return nil
	}

	// Graph types
	type GraphNode struct {
		ID     string   `json:"id"`
		Title  string   `json:"title"`
		URL    string   `json:"url"`
		Growth string   `json:"growth,omitempty"`
		Tags   []string `json:"tags,omitempty"`
	}
	type GraphEdge struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}
	type Graph struct {
		Nodes []GraphNode `json:"nodes"`
		Edges []GraphEdge `json:"edges"`
	}

	// Search index type
	type SearchEntry struct {
		Title   string   `json:"title"`
		URL     string   `json:"url"`
		Content string   `json:"content"`
		Tags    []string `json:"tags,omitempty"`
	}

	var graph Graph
	var searchIndex []SearchEntry

	// Single loop over all pages
	for _, page := range pages {
		if genGraph {
			graph.Nodes = append(graph.Nodes, GraphNode{
				ID:     page.Slug,
				Title:  page.Title,
				URL:    b.siteData.BasePath + page.Permalink,
				Growth: page.Growth,
				Tags:   page.Tags,
			})

			for _, target := range page.OutLinks {
				result := b.linkResolver.Resolve(target)
				if result.Page != nil {
					graph.Edges = append(graph.Edges, GraphEdge{
						Source: page.Slug,
						Target: result.Page.Slug,
					})
				}
			}
		}

		if genSearch && !page.IsIndex {
			searchIndex = append(searchIndex, SearchEntry{
				Title:   page.Title,
				URL:     b.siteData.BasePath + page.Permalink,
				Content: page.PlainContent(),
				Tags:    page.Tags,
			})
		}
	}

	// Write graph.json
	if genGraph {
		outPath := filepath.Join(b.outputDir, "graph.json")
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		if err := encodeJSON(f, graph); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	// Write search-index.json
	if genSearch {
		outPath := filepath.Join(b.outputDir, "search-index.json")
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		if err := encodeJSON(f, searchIndex); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}

// Helper functions

func filterDrafts(pages []*content.Page) []*content.Page {
	var result []*content.Page
	for _, p := range pages {
		if !p.Draft {
			result = append(result, p)
		}
	}
	return result
}

// buildSectionIndex creates a map of section -> pages for O(1) lookups
func buildSectionIndex(pages []*content.Page) map[string][]*content.Page {
	index := make(map[string][]*content.Page)
	for _, page := range pages {
		if page.IsIndex {
			continue
		}
		section := filepath.Dir(page.Slug)
		if section == "." {
			section = ""
		}
		index[section] = append(index[section], page)
	}
	return index
}

// buildTagIndex creates a map of tag (lowercase) -> pages for O(1) lookups
func buildTagIndex(pages []*content.Page) map[string][]*content.Page {
	index := make(map[string][]*content.Page)
	for _, page := range pages {
		for _, tag := range page.Tags {
			tagLower := strings.ToLower(tag)
			index[tagLower] = append(index[tagLower], page)
		}
	}
	return index
}

// getSectionPages returns pages in a section (falls back to linear scan if no index)
func getSectionPages(section string, allPages []*content.Page) []*content.Page {
	var result []*content.Page
	for _, page := range allPages {
		if page.IsIndex {
			continue
		}
		pageDir := filepath.Dir(page.Slug)
		if pageDir == "." {
			pageDir = ""
		}
		if pageDir == section {
			result = append(result, page)
		}
	}
	return result
}

// getSectionPagesFromIndex returns pages using the pre-built section index (O(1) lookup)
func (b *Builder) getSectionPagesFromIndex(section string) []*content.Page {
	if b.pagesBySection != nil {
		return b.pagesBySection[section]
	}
	// Fallback to linear scan if index not built
	return getSectionPages(section, b.pages)
}

func sortPages(pages []*content.Page, sortBy string) {
	switch sortBy {
	case "title":
		sort.Slice(pages, func(i, j int) bool {
			return pages[i].Title < pages[j].Title
		})
	case "growth":
		growthOrder := map[string]int{"seedling": 0, "budding": 1, "evergreen": 2, "": 3}
		sort.Slice(pages, func(i, j int) bool {
			return growthOrder[pages[i].Growth] < growthOrder[pages[j].Growth]
		})
	default: // date - use display date logic (modified if present, otherwise created)
		sort.Slice(pages, func(i, j int) bool {
			dateI := pages[i].Date
			if pages[i].HasModified() {
				dateI = pages[i].Modified
			}
			dateJ := pages[j].Date
			if pages[j].HasModified() {
				dateJ = pages[j].Modified
			}
			return dateI.After(dateJ)
		})
	}
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0644)
	})
}

func encodeJSON(f *os.File, v interface{}) error {
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// extractBasePath extracts the path portion from a URL for subdirectory hosting
// e.g., "https://user.github.io/repo" -> "/repo"
// e.g., "https://example.com" -> ""
func extractBasePath(baseURL string) string {
	if baseURL == "" {
		return ""
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	// Remove trailing slash
	path := strings.TrimSuffix(parsed.Path, "/")
	return path
}
