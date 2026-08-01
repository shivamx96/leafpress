// Package render implements the leafpress-render bridge: a pure
// stdin→stdout JSON transform that renders a set of published pages
// (a "garden") into full HTML documents, an index page, and theme CSS.
// It performs no filesystem, network, or database access.
package render

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/shivamx96/leafpress/core/assets"
	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
	sitegen "github.com/shivamx96/leafpress/core/site"
	"github.com/shivamx96/leafpress/core/templates"
)

// Input is the top-level JSON object read from stdin.
type Input struct {
	Garden Garden      `json:"garden"`
	Pages  []InputPage `json:"pages"`
	// Config accepts the canonical leafpress.json object. It overlays the same
	// defaults and uses the same validation as the CLI. Garden remains the
	// bridge envelope (and backwards-compatible fallback) because Slug is a
	// hosted-garden identity rather than a leafpress.json field.
	Config json.RawMessage `json:"config"`
	// StyleCSS is the in-memory counterpart of the CLI project's style.css.
	StyleCSS string `json:"styleCSS"`
	// Assets optionally declares the user assets the caller will serve
	// alongside the rendered site (custom font files under static/fonts/,
	// and in the future other referenced static files). Entries are
	// validated with the shared manifest rules and merged into the output
	// manifest; an entry whose effective output path collides with a
	// built-in's replaces that built-in (the favicon-override rule).
	Assets []assets.Asset `json:"assets"`
	// EmitAssets requests base64 artifacts for the built-in assets the
	// rendered site requires. The asset manifest is always emitted; bytes
	// are opt-in so routine renders stay small. Synchronization is
	// hash-driven per manifest entry — the registry ID alone is never a
	// valid skip signal, because the manifest is a theme-dependent subset.
	EmitAssets bool `json:"emitAssets"`
}

// Garden describes the garden being rendered.
type Garden struct {
	Slug          string          `json:"slug"`
	Title         string          `json:"title"`   // optional; defaults to Slug
	BaseURL       string          `json:"baseUrl"` // optional URL path prefix, e.g. "/g/shivam"
	Sort          string          `json:"sort"`    // optional: date (default) | title | growth
	ShowTagsInNav bool            `json:"showTagsInNav"`
	Theme         json.RawMessage `json:"theme"` // optional; maps onto config.Theme
}

// InputPage is a single published page. Slugs may carry path segments
// ("essays/my-post"); section membership derives from the slug's directory,
// exactly like the CLI build.
type InputPage struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"` // optional; defaults to Slug
	Markdown    string   `json:"markdown"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"createdAt"` // optional RFC3339
	UpdatedAt   string   `json:"updatedAt"` // optional RFC3339
	Description string   `json:"description"`
	Growth      string   `json:"growth"`
	TOC         *bool    `json:"toc"`
	Image       string   `json:"image"`
	ReadingTime *int     `json:"readingTime"`
	// IsIndex marks a section home (the CLI's _index.md): Slug is the
	// section path itself, Markdown becomes the intro above the child
	// listing. An IsIndex page with slug "" is the garden home.
	IsIndex bool `json:"isIndex"`
	// Sort orders the child listing of an index page: date (default) |
	// title | growth. Mirrors the _index.md `sort` frontmatter key.
	Sort string `json:"sort"`
	// ShowList toggles the child listing of an index page (default true).
	// Mirrors the _index.md `showList` frontmatter key.
	ShowList *bool `json:"showList"`
}

// Output is the top-level JSON object written to stdout.
type Output struct {
	Pages    []OutputPage    `json:"pages"`
	Index    string          `json:"index"`
	Sections []OutputSection `json:"sections"`
	Tags     OutputTags      `json:"tags"`
	CSS      string          `json:"css"`
	// AssetManifest is the combined manifest of every asset the rendered
	// site requires: referenced built-ins plus caller-declared assets, with
	// caller entries replacing built-ins on output-path collision. Metadata
	// only, never bytes. Hosted consumers materialize each entry through
	// their own storage using the content hash; built-in entries also
	// appear as base64 artifacts when the input sets emitAssets.
	AssetManifest assets.Manifest `json:"assetManifest"`
	// AssetRegistryID identifies the built-in registry snapshot the manifest
	// came from (content-derived). It is a change signal only — the manifest
	// is a theme-dependent subset, so synchronization stays hash-driven per
	// entry, never keyed on this ID.
	AssetRegistryID string           `json:"assetRegistryId"`
	Artifacts       []OutputArtifact `json:"artifacts"`
	Warnings        []string         `json:"warnings"`
}

// OutputArtifact is a filesystem-free generated site file. Path uses the
// exact CLI filename so consumers can store/serve artifacts generically.
type OutputArtifact struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	// Encoding says how Content encodes the file bytes and is authoritative
	// (never sniff): generated site artifacts are always "utf8"; asset
	// artifacts emitted under emitAssets are always "base64" regardless of
	// MIME type (OFL license texts included).
	Encoding string `json:"encoding"`
}

// OutputPage is a rendered page document. Index pages appear here too,
// rendered as their section's home.
type OutputPage struct {
	Slug string `json:"slug"`
	HTML string `json:"html"`
}

// OutputSection is an auto-generated home for a section that has no index
// page (the CLI's auto-index), served at {baseUrl}/<slug>/.
type OutputSection struct {
	Slug string `json:"slug"`
	HTML string `json:"html"`
}

// OutputTags holds the rendered tag index and per-tag pages. When no page
// carries any tag, Index is "" and Pages is empty (mirroring the CLI, which
// skips the tags section entirely in that case).
type OutputTags struct {
	Index string          `json:"index"`
	Pages []OutputTagPage `json:"pages"`
}

// OutputTagPage is a rendered page listing everything under one tag,
// served at {baseUrl}/tags/<tag>/.
type OutputTagPage struct {
	Tag  string `json:"tag"`
	HTML string `json:"html"`
}

// InputError marks failures caused by invalid input (exit code 1),
// as opposed to internal render failures (exit code 2).
type InputError struct {
	msg string
}

func (e *InputError) Error() string { return e.msg }

func inputErrorf(format string, args ...any) error {
	return &InputError{msg: fmt.Sprintf(format, args...)}
}

var (
	// unsafeSlugChars are characters that would break hrefs/attributes when a
	// slug is interpolated into template output.
	unsafeSlugChars = "\"'<>\\ \t\r\n"
	// tagRegex restricts tags to letters/digits/underscore/hyphen (the
	// hosted tag shape); tags are interpolated into hrefs and text unescaped.
	tagRegex = regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)
)

// Run decodes raw JSON input and renders it. Errors of type *InputError
// indicate invalid input; any other error is an internal failure.
func Run(raw []byte) (*Output, error) {
	var in Input
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, inputErrorf("invalid input JSON: %v", err)
	}
	return Render(&in)
}

// Render validates the input and produces rendered output.
func Render(in *Input) (*Output, error) {
	if in.Garden.Slug == "" {
		return nil, inputErrorf("garden.slug is required")
	}

	cfg, site, canonicalConfig, err := resolveSiteInput(in)
	if err != nil {
		return nil, err
	}
	title := site.Title
	basePath := site.BasePath

	sortMode := in.Garden.Sort
	if sortMode == "" {
		sortMode = "date"
	}
	switch sortMode {
	case "date", "title", "growth":
	default:
		return nil, inputErrorf("garden.sort must be one of date, title, growth; got %q", sortMode)
	}

	pages, err := buildPages(in.Pages)
	if err != nil {
		return nil, err
	}

	// Resolve wikilinks over exactly these pages; unresolved links degrade
	// to plain text (anything unresolved is private by design).
	resolver := content.NewLinkResolver(pages)
	// NewLinkResolver already registered each page's stored title — which
	// buildPages HTML-escaped at the input boundary. Authors link by the
	// raw display title ([[Foo & Bar]], not [[Foo &amp; Bar]]), so register
	// the raw input titles too; for titles without special characters this
	// second registration is a no-op. buildPages preserves input order, so
	// zip by index.
	for i, ip := range in.Pages {
		resolver.AddAlias(ip.Title, pages[i])
	}
	if cfg.Backlinks {
		content.BuildBacklinks(pages, resolver)
	}
	renderer := content.NewRenderer(resolver, cfg.Wikilinks, basePath)
	renderer.SetPlainBrokenLinks(true)
	// Raw HTML in author markdown renders as visibly escaped text; this
	// bridge serves multi-tenant user content to third parties.
	renderer.SetEscapeRawHTML(true)

	warnings := []string{}
	for _, p := range pages {
		rendered, warns := renderer.Render(p.RawContent)
		p.HTMLContent = rendered
		p.WordCount = content.CountWords(rendered)
		p.ImageCount = content.CountImages(rendered)
		p.ReadingTime = content.CalculateReadingTime(p.WordCount, p.ImageCount)
		if p.ReadingTimeOverride != nil {
			p.ReadingTime = *p.ReadingTimeOverride
		}
		// Pin the auto-generated SEO description to an escaped value.
		// Page.SEODescription derives it from HTMLContent via PlainContent,
		// which runs html.UnescapeString — that would resurrect quotes and
		// angle brackets from body text right before the value is placed in
		// meta attributes by text/template.
		if p.Description == "" {
			p.Description = html.EscapeString(p.SEODescription())
		}
		for _, w := range warns {
			warnings = append(warnings, fmt.Sprintf("page %q: %s", p.Slug, w))
		}
	}

	tmpl, err := templates.New()
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	// Section membership derives from slug paths, exactly like the CLI:
	// a page "essays/my-post" is a direct child of section "essays", and
	// an index page's slug *is* its section path ("" = the garden home).
	children := make(map[string][]*content.Page)
	indexBySection := make(map[string]*content.Page)
	for _, p := range pages {
		if p.IsIndex {
			indexBySection[p.Slug] = p
		} else {
			dir := sectionOf(p.Slug)
			children[dir] = append(children[dir], p)
		}
	}

	// Root notes and sections are the garden's public top level. Use that
	// same projection for both the home listing and site-wide navigation;
	// pages nested inside a section remain ordinary notes.
	homeEntries := rootEntries(children, indexBySection)
	if !canonicalConfig {
		site.Nav = navItems(homeEntries)
		if in.Garden.ShowTagsInNav && pagesHaveTags(pages) {
			site.Nav = append(site.Nav, config.NavItem{Label: "Tags", Path: "/tags/"})
		}
	}

	outPages := make([]OutputPage, 0, len(pages))
	for _, p := range pages {
		var buf bytes.Buffer
		if p.IsIndex {
			if p.Slug == "" {
				continue // the garden home renders into Output.Index below
			}
			if err := renderSectionHome(&buf, tmpl, site, p, children[p.Slug]); err != nil {
				return nil, fmt.Errorf("failed to render section %q: %w", p.Slug, err)
			}
		} else {
			var toc []templates.TOCItem
			htmlContent := p.HTMLContent
			showTOC := site.TOC
			if p.TOC != nil {
				showTOC = *p.TOC
			}
			if showTOC {
				htmlContent, toc = templates.ExtractTOC(p.HTMLContent)
			}
			if err := tmpl.RenderPage(&buf, templates.PageData{
				Site:        site,
				Page:        p,
				Content:     htmlContent,
				TOC:         toc,
				CurrentPath: p.Permalink,
			}); err != nil {
				return nil, fmt.Errorf("failed to render page %q: %w", p.Slug, err)
			}
		}
		outPages = append(outPages, OutputPage{Slug: p.Slug, HTML: buf.String()})
	}

	// Sections without an index page get an auto-generated home, mirroring
	// the CLI's auto-indexes: title-cased folder name, date sort, list on.
	sections := []OutputSection{}
	autoDirs := make([]string, 0, len(children))
	for dir := range children {
		if dir != "" && indexBySection[dir] == nil {
			autoDirs = append(autoDirs, dir)
		}
	}
	sort.Strings(autoDirs)
	for _, dir := range autoDirs {
		sorted := make([]*content.Page, len(children[dir]))
		copy(sorted, children[dir])
		sortPages(sorted, "date")
		var buf bytes.Buffer
		if err := tmpl.RenderIndex(&buf, templates.IndexData{
			Site:        site,
			Title:       titleCase(path.Base(dir)),
			Pages:       sorted,
			ShowList:    true,
			CurrentPath: "/" + dir + "/",
		}); err != nil {
			return nil, fmt.Errorf("failed to render section %q: %w", dir, err)
		}
		sections = append(sections, OutputSection{Slug: dir, HTML: buf.String()})
	}

	// The garden home: a root index page replaces the generated index and
	// renders its intro over the listing. Without one, the bridge keeps a
	// synthetic home (hosted gardens always have a home, unlike native
	// sites, which omit / entirely in that case).
	var idx bytes.Buffer
	if root := indexBySection[""]; root != nil {
		if root.Title == "" {
			root.Title = title
		}
		if err := renderSectionHome(&idx, tmpl, site, root, homeEntries); err != nil {
			return nil, fmt.Errorf("failed to render garden home: %w", err)
		}
	} else {
		sorted := make([]*content.Page, len(homeEntries))
		copy(sorted, homeEntries)
		sortPages(sorted, sortMode)
		if err := tmpl.RenderIndex(&idx, templates.IndexData{
			Site:        site,
			Title:       title,
			Pages:       sorted,
			ShowList:    true,
			CurrentPath: "/",
		}); err != nil {
			return nil, fmt.Errorf("failed to render index: %w", err)
		}
	}

	tags, err := renderTagPages(tmpl, site, pages)
	if err != nil {
		return nil, err
	}

	artifacts, err := renderArtifacts(tmpl, site, cfg, pages, resolver)
	if err != nil {
		return nil, err
	}

	// Self-contained output is the default: families with no self-hosted
	// source fall back to the CSS system stacks, and the author is told.
	if !cfg.Theme.RemoteFonts {
		for _, family := range templates.UnhostedFamilies(cfg.Theme) {
			warnings = append(warnings, templates.UnhostedFontWarning(family))
		}
	}

	manifest, assetWarnings, err := buildAssetManifest(in, cfg)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, assetWarnings...)
	if in.EmitAssets {
		artifacts = append(artifacts, builtinAssetArtifacts(manifest)...)
	}

	return &Output{
		Pages:           outPages,
		Index:           idx.String(),
		Sections:        sections,
		Tags:            tags,
		CSS:             sitegen.Styles(in.StyleCSS, cfg.Theme),
		AssetManifest:   manifest,
		AssetRegistryID: assets.RegistryID(),
		Artifacts:       artifacts,
		Warnings:        warnings,
	}, nil
}

// callerAssets validates the caller-supplied asset manifest: the declaration
// of user files (custom fonts today) the caller will serve alongside the
// rendered site. A pure renderer cannot discover uploads, so this input is
// how they enter the asset contract.
func callerAssets(in *Input) (assets.Manifest, error) {
	if len(in.Assets) == 0 {
		return nil, nil
	}
	// Policy (reserved namespace, outputPath restricted to built-in
	// overrides) is shared with the CLI via assets.ValidateUserAsset, so
	// the two interfaces cannot drift on what a legal user asset is.
	for i, a := range in.Assets {
		if err := assets.ValidateUserAsset(a); err != nil {
			return nil, inputErrorf("assets[%d]: %v", i, err)
		}
	}
	m, err := assets.NewManifest(in.Assets...)
	if err != nil {
		return nil, inputErrorf("assets: %v", err)
	}
	return m, nil
}

// buildAssetManifest produces the combined referenced manifest: required
// built-ins with caller entries merged over them (a caller entry replaces a
// built-in on effective-output-path collision — the favicon-override rule),
// plus warnings for custom font files the configuration references but the
// caller never declared.
func buildAssetManifest(in *Input, cfg *config.Config) (assets.Manifest, []string, error) {
	caller, err := callerAssets(in)
	if err != nil {
		return nil, nil, err
	}

	required := assets.RequiredBuiltins(cfg.Theme.FontHeading, cfg.Theme.FontBody, cfg.Theme.FontMono)
	builtinAssets := make([]assets.Asset, 0, len(required))
	for _, b := range required {
		builtinAssets = append(builtinAssets, b.Asset)
	}
	base, err := assets.NewManifest(builtinAssets...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build built-in manifest: %w", err)
	}

	merged, err := assets.Merge(base, caller)
	if err != nil {
		return nil, nil, inputErrorf("assets: %v", err)
	}

	var warnings []string
	declared := map[string]bool{}
	for _, a := range caller {
		declared[a.LogicalPath] = true
	}
	for _, face := range cfg.Theme.Fonts {
		if !declared[face.File] {
			warnings = append(warnings, fmt.Sprintf(
				"custom font %q references %s, which is not declared in the caller asset manifest; the site will 404 it unless the host serves it another way",
				face.Family, face.File))
		}
	}
	return merged, warnings, nil
}

// builtinAssetArtifacts returns base64 artifacts for the manifest entries
// whose bytes the renderer actually has: built-ins that survived the merge.
// Caller assets never produce byte artifacts (only the caller has them), and
// an overridden built-in is not emitted. Artifact paths use the effective
// output path — the exact filename a CLI export serves — so generic hosts
// can store artifacts by path without remapping.
func builtinAssetArtifacts(manifest assets.Manifest) []OutputArtifact {
	var out []OutputArtifact
	for _, entry := range manifest {
		b, ok := assets.BuiltinByLogicalPath(entry.LogicalPath)
		if !ok || b.Asset.SHA256 != entry.SHA256 {
			continue
		}
		out = append(out, OutputArtifact{
			Path:        entry.EffectiveOutputPath(),
			Content:     base64.StdEncoding.EncodeToString(b.Content()),
			ContentType: entry.ContentType,
			Encoding:    "base64",
		})
	}
	return out
}

// sectionOf returns the section path a slug belongs to ("" for root),
// mirroring the CLI's filepath.Dir-based grouping.
func sectionOf(slug string) string {
	dir := path.Dir(slug)
	if dir == "." {
		return ""
	}
	return dir
}

// renderSectionHome renders an index page as its section's home, mirroring
// the CLI's renderSectionIndex: the page body becomes the intro above the
// child listing, sorted by the page's sort key (default date), with the
// listing toggled by showList (default true).
func renderSectionHome(buf *bytes.Buffer, tmpl *templates.Templates, site templates.SiteData, indexPage *content.Page, sectionPages []*content.Page) error {
	sorted := make([]*content.Page, len(sectionPages))
	copy(sorted, sectionPages)
	sortPages(sorted, indexPage.SectionSort)

	showList := true
	if indexPage.ShowList != nil {
		showList = *indexPage.ShowList
	}
	currentPath := "/" + indexPage.Slug
	if currentPath != "/" {
		currentPath += "/"
	}
	return tmpl.RenderIndex(buf, templates.IndexData{
		Site:        site,
		Title:       indexPage.Title,
		Pages:       sorted,
		Intro:       indexPage.HTMLContent,
		ShowList:    showList,
		CurrentPath: currentPath,
	})
}

// rootEntries returns the garden's public top level: root notes plus one
// entry per section. Nested notes belong to their section home and never
// appear directly at the root.
func rootEntries(children map[string][]*content.Page, indexBySection map[string]*content.Page) []*content.Page {
	sectionDirs := make([]string, 0, len(children)+len(indexBySection))
	for dir := range children {
		if dir != "" {
			sectionDirs = append(sectionDirs, dir)
		}
	}
	for dir := range indexBySection {
		if dir != "" && children[dir] == nil {
			sectionDirs = append(sectionDirs, dir) // index page of an empty section
		}
	}
	sort.Strings(sectionDirs)

	entries := make([]*content.Page, 0, len(children[""])+len(sectionDirs))
	entries = append(entries, children[""]...)
	for _, dir := range sectionDirs {
		if p := indexBySection[dir]; p != nil {
			entries = append(entries, p)
		} else {
			entries = append(entries, autoSectionEntry(dir, children[dir]))
		}
	}
	return entries
}

func navItems(entries []*content.Page) []config.NavItem {
	nav := make([]config.NavItem, 0, len(entries))
	for _, entry := range entries {
		nav = append(nav, config.NavItem{Label: entry.Title, Path: entry.Permalink})
	}
	return nav
}

func pagesHaveTags(pages []*content.Page) bool {
	for _, page := range pages {
		if len(page.Tags) > 0 {
			return true
		}
	}
	return false
}

// renderTagPages renders the tags index and one page per tag, matching the
// CLI's tag output (tags are grouped case-insensitively; pages within a tag
// are sorted by date; tags are listed alphabetically). Page headers link to
// {baseUrl}/tags/<tag>/, so these pages are what keeps those links alive.
func renderTagPages(tmpl *templates.Templates, site templates.SiteData, pages []*content.Page) (OutputTags, error) {
	out := OutputTags{Pages: []OutputTagPage{}}

	// Group pages by lowercased tag (mirrors the CLI's buildTagIndex).
	pagesByTag := make(map[string][]*content.Page)
	for _, p := range pages {
		for _, tag := range p.Tags {
			key := strings.ToLower(tag)
			pagesByTag[key] = append(pagesByTag[key], p)
		}
	}
	if len(pagesByTag) == 0 {
		// No tags anywhere: no tag links are rendered on pages (the tag
		// block is per-page), so emit nothing — like the CLI, which skips
		// the tags/ output directory entirely.
		return out, nil
	}

	tagNames := make([]string, 0, len(pagesByTag))
	for tag := range pagesByTag {
		tagNames = append(tagNames, tag)
	}
	sort.Strings(tagNames)

	tagInfos := make([]templates.TagInfo, 0, len(tagNames))
	for _, tag := range tagNames {
		tagInfos = append(tagInfos, templates.TagInfo{Name: tag, Count: len(pagesByTag[tag])})
	}

	var idx bytes.Buffer
	if err := tmpl.RenderTagIndex(&idx, templates.TagIndexData{
		Site:        site,
		Tags:        tagInfos,
		CurrentPath: "/tags/",
	}); err != nil {
		return out, fmt.Errorf("failed to render tag index: %w", err)
	}
	out.Index = idx.String()

	for _, tag := range tagNames {
		tagged := make([]*content.Page, len(pagesByTag[tag]))
		copy(tagged, pagesByTag[tag])
		sortPages(tagged, "date")

		var buf bytes.Buffer
		if err := tmpl.RenderTagPage(&buf, templates.TagPageData{
			Site:        site,
			Tag:         tag,
			Pages:       tagged,
			CurrentPath: "/tags/" + tag + "/",
		}); err != nil {
			return out, fmt.Errorf("failed to render tag page %q: %w", tag, err)
		}
		out.Pages = append(out.Pages, OutputTagPage{Tag: tag, HTML: buf.String()})
	}

	return out, nil
}

// resolveSiteInput turns either the canonical leafpress.json object or the
// legacy garden bridge fields into template data. Canonical config is parsed
// by core/config, so the CLI and renderer share defaults and validation.
func resolveSiteInput(in *Input) (*config.Config, templates.SiteData, bool, error) {
	canonical := len(bytes.TrimSpace(in.Config)) > 0 && string(bytes.TrimSpace(in.Config)) != "null"
	var cfg *config.Config
	var basePath string
	if canonical {
		var err error
		cfg, err = config.Parse(in.Config)
		if err != nil {
			return nil, templates.SiteData{}, true, inputErrorf("invalid config: %v", err)
		}
		if cfg.BaseURL != "" {
			parsed, err := url.Parse(cfg.BaseURL)
			if err != nil {
				return nil, templates.SiteData{}, true, inputErrorf("invalid config.baseURL: %v", err)
			}
			basePath = strings.TrimSuffix(parsed.Path, "/")
		} else {
			var err error
			basePath, err = normalizeBasePath(in.Garden.BaseURL)
			if err != nil {
				return nil, templates.SiteData{}, true, err
			}
		}
	} else {
		theme, err := resolveTheme(in.Garden.Theme)
		if err != nil {
			return nil, templates.SiteData{}, false, err
		}
		basePath, err = normalizeBasePath(in.Garden.BaseURL)
		if err != nil {
			return nil, templates.SiteData{}, false, err
		}
		cfg = config.Default()
		cfg.Title = in.Garden.Title
		if cfg.Title == "" {
			cfg.Title = in.Garden.Slug
		}
		cfg.Theme = theme
		// Preserve the original bridge contract: rich site artifacts and their
		// controls activate only through canonical config.
		cfg.Graph = false
		cfg.Search = false
		cfg.RSS = false
	}

	rawSite := templates.SiteData{
		Title:       cfg.Title,
		Description: cfg.Description,
		Author:      cfg.Author,
		Nav:         cfg.Nav,
		Theme:       cfg.Theme,
		BaseURL:     cfg.BaseURL,
		BasePath:    basePath,
		Image:       cfg.Image,
		TOC:         cfg.TOC,
		Graph:       cfg.Graph,
		Search:      cfg.Search,
		RSS:         cfg.RSS,
		HeadExtra:   cfg.HeadExtra,
	}
	return cfg, safeSiteData(rawSite), canonical, nil
}

// safeSiteData protects text/template interpolation at the renderer trust
// boundary. headExtra deliberately remains raw because it is an explicit
// trusted Leafpress configuration escape hatch, matching the CLI.
func safeSiteData(site templates.SiteData) templates.SiteData {
	site.Title = html.EscapeString(site.Title)
	site.Description = html.EscapeString(site.Description)
	site.Author = html.EscapeString(site.Author)
	site.BaseURL = html.EscapeString(site.BaseURL)
	site.Image = html.EscapeString(site.Image)
	for i := range site.Nav {
		site.Nav[i].Label = html.EscapeString(site.Nav[i].Label)
		site.Nav[i].Path = html.EscapeString(site.Nav[i].Path)
	}
	return site
}

func renderArtifacts(
	tmpl *templates.Templates,
	safeSite templates.SiteData,
	cfg *config.Config,
	pages []*content.Page,
	resolver *content.LinkResolver,
) ([]OutputArtifact, error) {
	artifactPages := make([]*content.Page, 0, len(pages))
	for _, page := range pages {
		copy := *page
		copy.Title = html.UnescapeString(copy.Title)
		copy.Description = html.UnescapeString(copy.Description)
		copy.Image = html.UnescapeString(copy.Image)
		artifactPages = append(artifactPages, &copy)
	}
	rawSite := safeSite
	rawSite.Title = html.UnescapeString(rawSite.Title)
	rawSite.Description = html.UnescapeString(rawSite.Description)
	rawSite.Author = html.UnescapeString(rawSite.Author)
	rawSite.BaseURL = cfg.BaseURL
	rawSite.Image = html.UnescapeString(rawSite.Image)

	// search-index.json is always emitted: full-text search UI and hover link
	// previews share it. cfg.Search only controls the search UI chrome/JS.
	graphJSON, searchJSON, err := sitegen.GraphSearch(
		artifactPages, resolver, safeSite.BasePath, cfg.Graph, true,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate graph/search artifacts: %w", err)
	}
	artifacts := make([]OutputArtifact, 0, 6)
	if graphJSON != "" {
		artifacts = append(artifacts, OutputArtifact{
			Path: "graph.json", Content: graphJSON, ContentType: "application/json",
		})
	}
	if searchJSON != "" {
		artifacts = append(artifacts, OutputArtifact{
			Path: "search-index.json", Content: searchJSON, ContentType: "application/json",
		})
	}
	artifacts = append(artifacts,
		OutputArtifact{Path: "robots.txt", Content: sitegen.Robots(cfg.BaseURL), ContentType: "text/plain; charset=utf-8"},
		OutputArtifact{Path: "sitemap.xml", Content: sitegen.Sitemap(artifactPages, cfg.BaseURL), ContentType: "application/xml"},
	)
	notFound, err := sitegen.NotFound(tmpl, safeSite)
	if err != nil {
		return nil, fmt.Errorf("failed to render 404 artifact: %w", err)
	}
	artifacts = append(artifacts, OutputArtifact{
		Path: "404.html", Content: notFound, ContentType: "text/html; charset=utf-8",
	})
	if cfg.RSS {
		artifacts = append(artifacts, OutputArtifact{
			Path:        "feed.xml",
			Content:     sitegen.RSS(artifactPages, rawSite, cfg.BaseURL, time.Time{}),
			ContentType: "application/rss+xml",
		})
	}
	// Every generated artifact is text; asset artifacts (base64) are
	// appended by the caller.
	for i := range artifacts {
		artifacts[i].Encoding = "utf8"
	}
	return artifacts, nil
}

// normalizeBasePath validates and normalizes the garden baseUrl into a URL
// path prefix ("" or "/prefix" with no trailing slash).
func normalizeBasePath(baseURL string) (string, error) {
	if baseURL == "" || baseURL == "/" {
		return "", nil
	}
	if strings.ContainsAny(baseURL, unsafeSlugChars) {
		return "", inputErrorf("garden.baseUrl contains invalid characters: %q", baseURL)
	}
	if !strings.HasPrefix(baseURL, "/") {
		return "", inputErrorf("garden.baseUrl must start with /, got %q", baseURL)
	}
	return strings.TrimSuffix(baseURL, "/"), nil
}

// resolveTheme starts from the default theme and overlays any provided
// fields, then validates the result (theme values are interpolated into
// the inline <style> block, so they must be safe).
func resolveTheme(raw json.RawMessage) (config.Theme, error) {
	theme := config.Default().Theme
	if len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &theme); err != nil {
			return theme, inputErrorf("invalid garden.theme: %v", err)
		}
	}
	// Restore defaults for fields explicitly set to empty.
	def := config.Default().Theme
	if theme.FontHeading == "" {
		theme.FontHeading = def.FontHeading
	}
	if theme.FontBody == "" {
		theme.FontBody = def.FontBody
	}
	if theme.FontMono == "" {
		theme.FontMono = def.FontMono
	}
	if theme.Accent == "" {
		theme.Accent = def.Accent
	}
	if theme.NavStyle == "" {
		theme.NavStyle = def.NavStyle
	}
	if theme.NavActiveStyle == "" {
		theme.NavActiveStyle = def.NavActiveStyle
	}

	// Reuse config validation (accent hex, backgrounds, nav styles, font
	// names, custom font declarations).
	cfg := config.Default()
	cfg.Theme = theme
	if err := cfg.Validate(); err != nil {
		return theme, inputErrorf("invalid garden.theme: %v", err)
	}
	return theme, nil
}

// buildPages converts input pages into content pages.
func buildPages(in []InputPage) ([]*content.Page, error) {
	pages := make([]*content.Page, 0, len(in))
	seen := make(map[string]bool, len(in))
	for i, ip := range in {
		slug := strings.Trim(ip.Slug, "/")
		// Only the garden-home index page may have an empty slug (the
		// CLI's root _index.md).
		if slug == "" && !ip.IsIndex {
			return nil, inputErrorf("pages[%d].slug is required", i)
		}
		if strings.ContainsAny(slug, unsafeSlugChars) {
			return nil, inputErrorf("pages[%d].slug is invalid: %q", i, ip.Slug)
		}
		if seen[slug] {
			if slug == "" {
				return nil, inputErrorf("duplicate garden home index page")
			}
			return nil, inputErrorf("duplicate page slug: %q", slug)
		}
		seen[slug] = true

		switch ip.Sort {
		case "", "date", "title", "growth":
		default:
			return nil, inputErrorf("pages[%d].sort must be one of date, title, growth; got %q", i, ip.Sort)
		}

		for _, tag := range ip.Tags {
			if !tagRegex.MatchString(tag) {
				return nil, inputErrorf("pages[%d] has invalid tag %q: tags may only contain letters, digits, underscores, and hyphens", i, tag)
			}
		}

		// Growth flows into class/data attributes unescaped; restrict it to
		// the known enum.
		switch ip.Growth {
		case "", "seedling", "budding", "evergreen":
		default:
			return nil, inputErrorf("pages[%d].growth must be one of seedling, budding, evergreen; got %q", i, ip.Growth)
		}

		created, err := parseTime(ip.CreatedAt)
		if err != nil {
			return nil, inputErrorf("pages[%d].createdAt: %v", i, err)
		}
		updated, err := parseTime(ip.UpdatedAt)
		if err != nil {
			return nil, inputErrorf("pages[%d].updatedAt: %v", i, err)
		}
		if ip.ReadingTime != nil && *ip.ReadingTime <= 0 {
			return nil, inputErrorf("pages[%d].readingTime must be a positive integer", i)
		}

		// Title and description reach <title>, <h1>, and og:/twitter: meta
		// attributes through text/template — escape at the input boundary.
		title := html.EscapeString(ip.Title)
		if title == "" && slug != "" {
			if ip.IsIndex {
				// Mirror the scanner's generateTitleFromSlug fallback for
				// _index.md files without a frontmatter title. (An untitled
				// root index falls back to the garden title at render time.)
				title = html.EscapeString(titleFromSlug(path.Base(slug)))
			} else {
				title = html.EscapeString(slug)
			}
		}

		permalink := "/" + slug + "/"
		if slug == "" {
			permalink = "/"
		}
		pages = append(pages, &content.Page{
			Title:               title,
			Description:         html.EscapeString(ip.Description),
			Date:                created,
			Created:             created,
			Modified:            updated,
			Tags:                ip.Tags,
			Growth:              ip.Growth,
			TOC:                 ip.TOC,
			Image:               html.EscapeString(ip.Image),
			Slug:                slug,
			Permalink:           permalink,
			RawContent:          ip.Markdown,
			IsIndex:             ip.IsIndex,
			SectionSort:         ip.Sort,
			ShowList:            ip.ShowList,
			ReadingTimeOverride: ip.ReadingTime,
		})
	}
	return pages, nil
}

// titleCase title-cases an auto-index section name, matching the CLI's
// generateAutoIndexes. Hyphens read as word separators ("field-notes" →
// "Field Notes"), consistent with titleFromSlug below.
func titleCase(s string) string {
	return cases.Title(language.English).String(strings.ReplaceAll(s, "-", " "))
}

// autoSectionEntry is the single homepage link for a section that has no
// explicit index page. The section's real children remain on its generated
// section home instead of being flattened into the garden home.
func autoSectionEntry(slug string, children []*content.Page) *content.Page {
	entry := &content.Page{
		Title:     titleCase(path.Base(slug)),
		Slug:      slug,
		Permalink: "/" + slug + "/",
		IsIndex:   true,
	}

	// Give the section the date of its most recently changed child so a
	// date-sorted homepage places the folder alongside current root notes.
	for _, child := range children {
		date := child.Date
		if child.HasModified() {
			date = child.Modified
		}
		if date.After(entry.Date) {
			entry.Date = date
		}
	}
	return entry
}

// titleFromSlug turns a slug segment into a display title, matching the
// scanner's generateTitleFromSlug (hyphens to spaces, each word capitalized).
func titleFromSlug(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "-", " "))
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

// parseTime parses an optional RFC3339 timestamp ("" means unset).
func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("must be RFC3339: %v", err)
	}
	return t, nil
}

// sortPages orders pages for the index using the same semantics as the
// CLI's section sort (cli/internal/build). sort.SliceStable keeps output
// deterministic when keys compare equal.
func sortPages(pages []*content.Page, sortBy string) {
	switch sortBy {
	case "title":
		sort.SliceStable(pages, func(i, j int) bool {
			return pages[i].Title < pages[j].Title
		})
	case "growth":
		growthOrder := map[string]int{"seedling": 0, "budding": 1, "evergreen": 2, "": 3}
		sort.SliceStable(pages, func(i, j int) bool {
			return growthOrder[pages[i].Growth] < growthOrder[pages[j].Growth]
		})
	default: // date - display date logic (modified if present, otherwise created)
		sort.SliceStable(pages, func(i, j int) bool {
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
