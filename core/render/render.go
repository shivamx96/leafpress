// Package render implements the leafpress-render bridge: a pure
// stdin→stdout JSON transform that renders a set of published pages
// (a "garden") into full HTML documents, an index page, and theme CSS.
// It performs no filesystem, network, or database access.
package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
	"github.com/shivamx96/leafpress/core/templates"
)

// Input is the top-level JSON object read from stdin.
type Input struct {
	Garden Garden      `json:"garden"`
	Pages  []InputPage `json:"pages"`
}

// Garden describes the garden being rendered.
type Garden struct {
	Slug    string          `json:"slug"`
	Title   string          `json:"title"`   // optional; defaults to Slug
	BaseURL string          `json:"baseUrl"` // optional URL path prefix, e.g. "/g/shivam"
	Sort    string          `json:"sort"`    // optional: date (default) | title | growth
	Theme   json.RawMessage `json:"theme"`   // optional; maps onto config.Theme
}

// InputPage is a single published page.
type InputPage struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"` // optional; defaults to Slug
	Markdown    string   `json:"markdown"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"createdAt"` // optional RFC3339
	UpdatedAt   string   `json:"updatedAt"` // optional RFC3339
	Description string   `json:"description"`
	Growth      string   `json:"growth"`
}

// Output is the top-level JSON object written to stdout.
type Output struct {
	Pages    []OutputPage `json:"pages"`
	Index    string       `json:"index"`
	Tags     OutputTags   `json:"tags"`
	CSS      string       `json:"css"`
	Warnings []string     `json:"warnings"`
}

// OutputPage is a rendered page document.
type OutputPage struct {
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
	// fontNameRegex restricts font names to characters safe for direct
	// interpolation into the inline <style> block and Google Fonts URL.
	fontNameRegex = regexp.MustCompile(`^[A-Za-z0-9 _-]+$`)
	// unsafeSlugChars are characters that would break hrefs/attributes when a
	// slug is interpolated into template output.
	unsafeSlugChars = "\"'<>\\ \t\r\n"
	// tagRegex restricts tags to letters/digits/underscore/hyphen (leafpad's
	// tag shape); tags are interpolated into hrefs and text unescaped.
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

	// Templates are text/template, so anything author-controlled must be
	// made HTML-safe at this input boundary.
	title := html.EscapeString(in.Garden.Title)
	if title == "" {
		title = html.EscapeString(in.Garden.Slug)
	}

	basePath, err := normalizeBasePath(in.Garden.BaseURL)
	if err != nil {
		return nil, err
	}

	sortMode := in.Garden.Sort
	if sortMode == "" {
		sortMode = "date"
	}
	switch sortMode {
	case "date", "title", "growth":
	default:
		return nil, inputErrorf("garden.sort must be one of date, title, growth; got %q", sortMode)
	}

	theme, err := resolveTheme(in.Garden.Theme)
	if err != nil {
		return nil, err
	}

	pages, err := buildPages(in.Pages)
	if err != nil {
		return nil, err
	}

	// Resolve wikilinks over exactly these pages; unresolved links degrade
	// to plain text (anything unresolved is private by design).
	resolver := content.NewLinkResolver(pages)
	// Register each page's raw input title as an alias: authors link by
	// display title ([[Beta Note]]), while page slugs are hyphenated
	// (beta-note). buildPages preserves input order, so zip by index.
	for i, ip := range in.Pages {
		resolver.AddAlias(ip.Title, pages[i])
	}
	content.BuildBacklinks(pages, resolver)
	renderer := content.NewRenderer(resolver, true, basePath)
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

	site := templates.SiteData{
		Title:    title,
		Theme:    theme,
		BasePath: basePath,
		TOC:      true,
		// Graph, Search, and RSS stay off: this bridge does not generate
		// graph.json, search-index.json, or feed.xml (future flags).
	}

	outPages := make([]OutputPage, 0, len(pages))
	for _, p := range pages {
		htmlContent, toc := templates.ExtractTOC(p.HTMLContent)
		var buf bytes.Buffer
		if err := tmpl.RenderPage(&buf, templates.PageData{
			Site:        site,
			Page:        p,
			Content:     htmlContent,
			TOC:         toc,
			CurrentPath: p.Permalink,
		}); err != nil {
			return nil, fmt.Errorf("failed to render page %q: %w", p.Slug, err)
		}
		outPages = append(outPages, OutputPage{Slug: p.Slug, HTML: buf.String()})
	}

	sorted := make([]*content.Page, len(pages))
	copy(sorted, pages)
	sortPages(sorted, sortMode)

	var idx bytes.Buffer
	if err := tmpl.RenderIndex(&idx, templates.IndexData{
		Site:        site,
		Title:       title,
		Pages:       sorted,
		ShowList:    true,
		CurrentPath: "/",
	}); err != nil {
		return nil, fmt.Errorf("failed to render index: %w", err)
	}

	tags, err := renderTagPages(tmpl, site, pages)
	if err != nil {
		return nil, err
	}

	return &Output{
		Pages:    outPages,
		Index:    idx.String(),
		Tags:     tags,
		CSS:      templates.DefaultCSS,
		Warnings: warnings,
	}, nil
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

	// Reuse config validation (accent hex, backgrounds, nav styles).
	cfg := config.Default()
	cfg.Theme = theme
	if err := cfg.Validate(); err != nil {
		return theme, inputErrorf("invalid garden.theme: %v", err)
	}
	// Fonts are not covered by config.Validate; restrict to safe characters.
	for _, f := range []string{theme.FontHeading, theme.FontBody, theme.FontMono} {
		if !fontNameRegex.MatchString(f) {
			return theme, inputErrorf("invalid garden.theme font name: %q", f)
		}
	}
	return theme, nil
}

// buildPages converts input pages into content pages.
func buildPages(in []InputPage) ([]*content.Page, error) {
	pages := make([]*content.Page, 0, len(in))
	seen := make(map[string]bool, len(in))
	for i, ip := range in {
		if ip.Slug == "" {
			return nil, inputErrorf("pages[%d].slug is required", i)
		}
		slug := strings.Trim(ip.Slug, "/")
		if slug == "" || strings.ContainsAny(slug, unsafeSlugChars) {
			return nil, inputErrorf("pages[%d].slug is invalid: %q", i, ip.Slug)
		}
		if seen[slug] {
			return nil, inputErrorf("duplicate page slug: %q", slug)
		}
		seen[slug] = true

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

		// Title and description reach <title>, <h1>, and og:/twitter: meta
		// attributes through text/template — escape at the input boundary.
		title := html.EscapeString(ip.Title)
		if title == "" {
			title = html.EscapeString(slug)
		}

		pages = append(pages, &content.Page{
			Title:       title,
			Description: html.EscapeString(ip.Description),
			Date:        created,
			Created:     created,
			Modified:    updated,
			Tags:        ip.Tags,
			Growth:      ip.Growth,
			Slug:        slug,
			Permalink:   "/" + slug + "/",
			RawContent:  ip.Markdown,
		})
	}
	return pages, nil
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
