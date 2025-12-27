package build

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shivamx96/leafpress/internal/assets"
	"github.com/shivamx96/leafpress/internal/config"
	"github.com/shivamx96/leafpress/internal/content"
	"github.com/shivamx96/leafpress/internal/templates"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Options configures the build process
type Options struct {
	IncludeDrafts bool
	Verbose       bool
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

	// Clean output directory
	t0 = time.Now()
	if err := os.RemoveAll(b.outputDir); err != nil {
		return nil, fmt.Errorf("failed to clean output directory: %w", err)
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

	// Build backlinks
	t0 = time.Now()
	content.BuildBacklinks(pages)
	b.logTiming("backlinks", time.Since(t0))

	// Render markdown to HTML
	t0 = time.Now()
	warnings := content.RenderPages(pages)
	b.logTiming("markdown", time.Since(t0))
	stats.WarningCount = len(warnings)

	if b.opts.Verbose {
		for _, w := range warnings {
			fmt.Printf("  warning: %s\n", w)
		}
	}

	// Generate site data
	siteData := templates.SiteData{
		Title:   b.cfg.Title,
		Author:  b.cfg.Author,
		Nav:     b.cfg.Nav,
		Theme:   b.cfg.Theme,
		BaseURL: b.cfg.BaseURL,
		TOC:     b.cfg.TOC,
		Graph:   b.cfg.Graph,
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

	// Generate graph.json if enabled
	if b.cfg.Graph {
		t0 = time.Now()
		if err := b.generateGraph(pages); err != nil {
			return nil, fmt.Errorf("failed to generate graph: %w", err)
		}
		b.logTiming("graph", time.Since(t0))
	}

	return stats, nil
}

// renderPage renders a single content page
func (b *Builder) renderPage(page *content.Page, siteData templates.SiteData) error {
	outPath := filepath.Join(b.outputDir, page.OutputPath)

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
	sectionPages := getSectionPages(indexPage.Slug, allPages)

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

	// Generate indexes for directories without _index.md
	for dir := range dirs {
		if indexedDirs[dir] {
			continue
		}

		sectionPages := getSectionPages(dir, pages)
		sortPages(sectionPages, "date")

		outPath := filepath.Join(b.outputDir, dir, "index.html")
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		f, err := os.Create(outPath)
		if err != nil {
			return err
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
			return err
		}
		f.Close()
	}

	return nil
}

// generateTagPages creates tag index and individual tag pages
func (b *Builder) generateTagPages(pages []*content.Page, siteData templates.SiteData) error {
	// Collect all tags
	tagPages := make(map[string][]*content.Page)
	for _, page := range pages {
		for _, tag := range page.Tags {
			tagLower := strings.ToLower(tag)
			tagPages[tagLower] = append(tagPages[tagLower], page)
		}
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

	// Generate individual tag pages
	for tag, pages := range tagPages {
		sortPages(pages, "date")

		tagDir := filepath.Join(tagsDir, tag)
		if err := os.MkdirAll(tagDir, 0755); err != nil {
			return err
		}

		tagPath := filepath.Join(tagDir, "index.html")
		f, err := os.Create(tagPath)
		if err != nil {
			return err
		}

		if err := b.templates.RenderTagPage(f, templates.TagPageData{
			Site:        siteData,
			Tag:         tag,
			Pages:       pages,
			CurrentPath: "/tags/" + tag + "/",
		}); err != nil {
			f.Close()
			return err
		}
		f.Close()
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

// generateGraph creates graph.json for visualization
func (b *Builder) generateGraph(pages []*content.Page) error {
	// Build nodes and edges
	type Node struct {
		ID     string   `json:"id"`
		Title  string   `json:"title"`
		URL    string   `json:"url"`
		Growth string   `json:"growth,omitempty"`
		Tags   []string `json:"tags,omitempty"`
	}
	type Edge struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}
	type Graph struct {
		Nodes []Node `json:"nodes"`
		Edges []Edge `json:"edges"`
	}

	resolver := content.NewLinkResolver(pages)
	graph := Graph{}

	for _, page := range pages {
		graph.Nodes = append(graph.Nodes, Node{
			ID:     page.Slug,
			Title:  page.Title,
			URL:    page.Permalink,
			Growth: page.Growth,
			Tags:   page.Tags,
		})

		for _, target := range page.OutLinks {
			result := resolver.Resolve(target)
			if result.Page != nil {
				graph.Edges = append(graph.Edges, Edge{
					Source: page.Slug,
					Target: result.Page.Slug,
				})
			}
		}
	}

	// Write JSON
	outPath := filepath.Join(b.outputDir, "graph.json")
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return encodeJSON(f, graph)
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
