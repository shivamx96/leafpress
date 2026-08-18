package templates

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"html"
	"io"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template" // text/template used instead of html/template for performance — safe because leafpress is a trusted-content SSG (single author, no user-submitted input)

	"github.com/shivamx96/leafpress/core/assets"
	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
)

// Pre-compiled regexes for ExtractTOC and generateHeadingID (compiled once at startup)
var (
	headingRegex    = regexp.MustCompile(`<h([2-3])([^>]*)>(.*?)</h[2-3]>`)
	htmlTagRegex    = regexp.MustCompile(`<[^>]*>`)
	idAttrRegex     = regexp.MustCompile(`id\s*=`)
	nonASCIIRegex   = regexp.MustCompile(`[^\x00-\x7F]+`)
	nonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)
)

// Cached templates singleton (parsed once at first use)
var (
	cachedTemplates *Templates
	templateFuncs   template.FuncMap
)

func init() {
	templateFuncs = template.FuncMap{
		"growthEmoji":       growthEmoji,
		"growthDescription": growthDescription,
		"lower":             strings.ToLower,
		"safeHTML":          func(s string) string { return s },
		"safeCSS":           func(s string) string { return s },
		"fontURL":           fontURL,
		"fontPreloads":      fontPreloads,
		"remoteFontURL":     remoteFontURL,
		"hasPrefix":         strings.HasPrefix,
	}
}

// Templates holds all parsed templates
type Templates struct {
	base     *template.Template
	page     *template.Template
	index    *template.Template
	tagIndex *template.Template
	tagPage  *template.Template
	notFound *template.Template
}

// PageData is the data passed to page templates
type PageData struct {
	Site        SiteData
	Page        *content.Page
	Content     string
	TOC         []TOCItem
	CurrentPath string // Current page path for nav active state
}

// TOCItem represents a table of contents entry
type TOCItem struct {
	ID    string
	Text  string
	Level int
}

// IndexData is the data passed to index templates
type IndexData struct {
	Site        SiteData
	Title       string
	Pages       []*content.Page
	Intro       string // Optional intro content for section indexes
	ShowList    bool   // Show the page list
	CurrentPath string // Current page path for nav active state
}

// TagIndexData is the data passed to the tags index template
type TagIndexData struct {
	Site        SiteData
	Tags        []TagInfo
	CurrentPath string // Current page path for nav active state
}

// TagPageData is the data passed to individual tag pages
type TagPageData struct {
	Site        SiteData
	Tag         string
	Pages       []*content.Page
	CurrentPath string // Current page path for nav active state
}

// TagInfo holds tag name and count
type TagInfo struct {
	Name  string
	Count int
}

// NotFoundData is the data passed to the 404 template
type NotFoundData struct {
	Site        SiteData
	CurrentPath string
}

// FooterAttribution is optional host branding for renderer consumers. Keeping
// it structured prevents hosted callers from injecting raw footer HTML.
type FooterAttribution struct {
	Name string
	URL  string
}

// SiteData contains site-wide information
type SiteData struct {
	Title             string
	Description       string // Site-wide meta description
	Author            string
	Nav               []config.NavItem
	Theme             config.Theme
	BaseURL           string
	BasePath          string // Path portion of BaseURL (e.g., "/repo-name" for GitHub Pages)
	Image             string // Default OG image path
	TOC               bool
	Graph             bool
	Search            bool
	RSS               bool
	HeadExtra         string // Custom HTML to inject in <head>
	FooterAttribution *FooterAttribution
	ClientScriptPath  string // Content-hashed shared client bundle, relative to the site root
}

// ClientScriptAsset renders the site-wide client code once and returns its
// content-addressed output path. Pages can then share this file instead of
// embedding the same JavaScript in every HTML document.
func (t *Templates) ClientScriptAsset(site SiteData) (string, string, error) {
	data := struct{ Site SiteData }{Site: site}
	scripts := make([]string, 0, 2)
	for _, name := range []string{"clientScriptMain", "clientScriptMermaid"} {
		var rendered bytes.Buffer
		if err := t.base.ExecuteTemplate(&rendered, name, data); err != nil {
			return "", "", err
		}
		if script := strings.TrimSpace(rendered.String()); script != "" {
			scripts = append(scripts, script)
		}
	}

	if len(scripts) == 0 {
		return "", "", fmt.Errorf("Leafpress client script not found")
	}
	content := strings.Join(scripts, "\n\n") + "\n"

	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("static/leafpress/app.%x.js", hash[:16]), content, nil
}

// New returns a cached Templates instance (parsed once, reused on subsequent calls)
func New() (*Templates, error) {
	// Return cached templates if already parsed
	if cachedTemplates != nil {
		return cachedTemplates, nil
	}

	// Parse base template
	base, err := template.New("base").Funcs(templateFuncs).Parse(baseTemplate)
	if err != nil {
		return nil, err
	}

	// Clone base and add page-specific templates
	page, err := template.Must(base.Clone()).Parse(pageTemplate)
	if err != nil {
		return nil, err
	}

	index, err := template.Must(base.Clone()).Parse(indexTemplate)
	if err != nil {
		return nil, err
	}

	tagIndex, err := template.Must(base.Clone()).Parse(tagIndexTemplate)
	if err != nil {
		return nil, err
	}

	tagPage, err := template.Must(base.Clone()).Parse(tagPageTemplate)
	if err != nil {
		return nil, err
	}

	notFound, err := template.Must(base.Clone()).Parse(notFoundTemplate)
	if err != nil {
		return nil, err
	}

	cachedTemplates = &Templates{
		base:     base,
		page:     page,
		index:    index,
		tagIndex: tagIndex,
		tagPage:  tagPage,
		notFound: notFound,
	}

	return cachedTemplates, nil
}

// RenderPage renders a content page
func (t *Templates) RenderPage(w io.Writer, data PageData) error {
	bw := bufio.NewWriterSize(w, 8192)
	if err := t.page.Execute(bw, data); err != nil {
		return err
	}
	return bw.Flush()
}

// RenderIndex renders a section index page
func (t *Templates) RenderIndex(w io.Writer, data IndexData) error {
	bw := bufio.NewWriterSize(w, 8192)
	if err := t.index.Execute(bw, data); err != nil {
		return err
	}
	return bw.Flush()
}

// RenderTagIndex renders the tags index page
func (t *Templates) RenderTagIndex(w io.Writer, data TagIndexData) error {
	bw := bufio.NewWriterSize(w, 4096)
	if err := t.tagIndex.Execute(bw, data); err != nil {
		return err
	}
	return bw.Flush()
}

// RenderTagPage renders an individual tag page
func (t *Templates) RenderTagPage(w io.Writer, data TagPageData) error {
	bw := bufio.NewWriterSize(w, 8192)
	if err := t.tagPage.Execute(bw, data); err != nil {
		return err
	}
	return bw.Flush()
}

// RenderNotFound renders the 404 page
func (t *Templates) RenderNotFound(w io.Writer, data NotFoundData) error {
	bw := bufio.NewWriterSize(w, 4096)
	if err := t.notFound.Execute(bw, data); err != nil {
		return err
	}
	return bw.Flush()
}

func growthEmoji(growth string) string {
	switch growth {
	case "seedling":
		return "🌱"
	case "budding":
		return "🌿"
	case "evergreen":
		return "🌳"
	default:
		return ""
	}
}

func growthDescription(growth string) string {
	switch growth {
	case "seedling":
		return "Seedling: Early idea, still developing"
	case "budding":
		return "Budding: Growing, but not yet complete"
	case "evergreen":
		return "Evergreen: Fully grown and refined"
	default:
		return ""
	}
}

func fontURL(font string) string {
	// Replace spaces with + for Google Fonts URL
	fontParam := strings.ReplaceAll(font, " ", "+")
	return fontParam + ":wght@400;500;600;700"
}

// FontCSS returns every self-hosted @font-face rule for the theme — bundled
// built-in families plus custom static/fonts/ declarations. It is composed
// into the generated stylesheet (site.Styles), not inlined per page, so
// browsers download the rules once.
func FontCSS(theme config.Theme) string {
	return assets.FontFaceCSS(theme.FontHeading, theme.FontBody, theme.FontMono) +
		customFontCSS(theme.Fonts)
}

// fontFormats maps custom font file extensions to CSS src format() names.
var fontFormats = map[string]string{
	".woff2": "woff2",
	".woff":  "woff",
	".ttf":   "truetype",
	".otf":   "opentype",
}

var fontMIMETypes = map[string]string{
	".woff2": "font/woff2",
	".woff":  "font/woff",
	".ttf":   "font/ttf",
	".otf":   "font/otf",
}

// fontPreload describes one self-hosted face linked from every page head.
// Fields are exported because Go templates can only access exported fields.
type fontPreload struct {
	Path        string
	ContentType string
}

// fontPreloads resolves one normal Latin/regular face for each selected theme
// family, preserving role order (heading, body, mono). Families and files are
// deduplicated so assigning the same font to multiple roles never creates
// duplicate fetch hints. Remote-only families are deliberately skipped: their
// provider stylesheet owns the final font URLs.
func fontPreloads(theme config.Theme) []fontPreload {
	families := []string{theme.FontHeading, theme.FontBody, theme.FontMono}
	seenFamilies := map[string]bool{}
	seenPaths := map[string]bool{}
	preloads := make([]fontPreload, 0, len(families))

	for _, family := range families {
		if family == "" || seenFamilies[family] {
			continue
		}
		seenFamilies[family] = true

		if face, ok := assets.BuiltinFontPreloadFace(family); ok {
			if !seenPaths[face.LogicalPath] {
				seenPaths[face.LogicalPath] = true
				preloads = append(preloads, fontPreload{
					Path:        assets.EscapedURLPath(face.LogicalPath),
					ContentType: "font/woff2",
				})
			}
			continue
		}

		if face, ok := customFontPreloadFace(theme.Fonts, family); ok {
			fontPath := assets.EscapedURLPath(face.File)
			if !seenPaths[fontPath] {
				seenPaths[fontPath] = true
				preloads = append(preloads, fontPreload{
					Path:        fontPath,
					ContentType: fontMIMETypes[strings.ToLower(path.Ext(face.File))],
				})
			}
		}
	}

	return preloads
}

// customFontPreloadFace prefers the first normal face that can render weight
// 400, then falls back to the first normal face. Config validation guarantees
// the file extension and weight syntax before templates render.
func customFontPreloadFace(faces []config.FontFace, family string) (config.FontFace, bool) {
	var fallback config.FontFace
	for _, face := range faces {
		style := face.Style
		if style == "" {
			style = "normal"
		}
		if face.Family != family || style != "normal" {
			continue
		}
		if fallback.File == "" {
			fallback = face
		}
		weight := face.Weight
		if weight == "" || weight == "400" {
			return face, true
		}
		var low, high int
		if _, err := fmt.Sscanf(weight, "%d %d", &low, &high); err == nil && low <= 400 && high >= 400 {
			return face, true
		}
	}
	return fallback, fallback.File != ""
}

// customFontCSS renders @font-face rules for the theme's custom local font
// declarations. URLs are site-relative like the built-in faces (the rules
// live in the root-served stylesheet) with segments escaped as defense in
// depth. Empty weight/style/display fall back to "400", "normal", and
// "swap" (the same defaults config validation documents).
func customFontCSS(faces []config.FontFace) string {
	var sb strings.Builder
	for _, face := range faces {
		weight := face.Weight
		if weight == "" {
			weight = "400"
		}
		style := face.Style
		if style == "" {
			style = "normal"
		}
		display := face.Display
		if display == "" {
			display = "swap"
		}
		format := fontFormats[strings.ToLower(path.Ext(face.File))]
		if format == "" {
			// Validation rejects unknown extensions; never emit format("")
			// if an unvalidated Theme reaches this point.
			continue
		}
		fmt.Fprintf(&sb, `@font-face {
  font-family: "%s";
  font-style: %s;
  font-weight: %s;
  font-display: %s;
  src: url("%s") format("%s");
}
`, face.Family, style, weight, display, assets.EscapedURLPath(face.File), format)
	}
	return sb.String()
}

// UnhostedFontWarning is the canonical author-facing message for a family
// with no self-hosted source. The CLI and the renderer must emit exactly
// this string so warnings stay greppable and support-friendly.
func UnhostedFontWarning(family string) string {
	return fmt.Sprintf(
		"font family %q is not bundled or declared; falling back to system fonts (declare it under theme.fonts, or set theme.remoteFonts to temporarily keep Google Fonts)",
		family)
}

// UnhostedFamilies returns the configured families that have no self-hosted
// source: neither in the bundled set nor declared as custom local fonts.
// With the deprecated remoteFonts escape hatch off, these fall back to the
// CSS system stacks and callers should warn the author.
func UnhostedFamilies(theme config.Theme) []string {
	var out []string
	seen := map[string]bool{}
	for _, font := range []string{theme.FontHeading, theme.FontBody, theme.FontMono} {
		if font == "" || assets.IsBuiltinFontFamily(font) || theme.DeclaresFamily(font) || seen[font] {
			continue
		}
		seen[font] = true
		out = append(out, font)
	}
	return out
}

// remoteFontURL returns the Google Fonts stylesheet URL for the configured
// families outside the bundled set — but only when the deprecated
// theme.remoteFonts escape hatch is enabled. The default is self-contained
// output: unhosted families fall back to the CSS system stacks (with a
// warning emitted at build/render time).
func remoteFontURL(theme config.Theme) string {
	if !theme.RemoteFonts {
		return ""
	}
	families := []string{}
	seen := map[string]bool{}
	for _, font := range UnhostedFamilies(theme) {
		param := fontURL(font)
		if !seen[param] {
			seen[param] = true
			families = append(families, param)
		}
	}
	if len(families) == 0 {
		return ""
	}
	return "https://fonts.googleapis.com/css2?family=" + strings.Join(families, "&family=") + "&display=swap"
}

// ExtractTOC extracts headings from HTML content and adds IDs to them
func ExtractTOC(htmlContent string) (string, []TOCItem) {
	var toc []TOCItem
	idCounter := make(map[string]int)

	modifiedHTML := headingRegex.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract level, attributes, and text
		matches := headingRegex.FindStringSubmatch(match)
		if len(matches) != 4 {
			return match
		}

		level := matches[1]
		attrs := matches[2]
		text := matches[3]

		// Strip HTML tags from text for TOC display
		plainText := htmlTagRegex.ReplaceAllString(text, "")
		// Unescape HTML entities (e.g., &amp; -> &, &#39; -> ')
		plainText = html.UnescapeString(plainText)

		// Generate ID from text
		id := generateHeadingID(plainText)

		// Handle duplicate IDs
		if count, exists := idCounter[id]; exists {
			idCounter[id] = count + 1
			id = id + "-" + strconv.Itoa(count)
		} else {
			idCounter[id] = 1
		}

		// Add to TOC
		levelInt := 2
		if level == "3" {
			levelInt = 3
		}
		toc = append(toc, TOCItem{
			ID: id,
			// Templates use text/template because rendered markdown is inserted
			// deliberately. Escape TOC labels before they cross that raw boundary:
			// hosted markdown is escaped first, and UnescapeString above must not
			// resurrect author HTML as live markup in the generated TOC.
			Text:  html.EscapeString(plainText),
			Level: levelInt,
		})

		// If heading already has an id from goldmark, replace it with ours for consistency
		if attrs != "" && idAttrRegex.MatchString(attrs) {
			// Replace existing id with our generated one
			existingIDRegex := regexp.MustCompile(`\s*id="[^"]*"`)
			attrs = existingIDRegex.ReplaceAllString(attrs, "")
		}
		anchor := `<a class="lp-heading-anchor" href="#` + id + `" aria-hidden="true">#</a>`
		if attrs != "" {
			return "<h" + level + attrs + " id=\"" + id + "\">" + text + " " + anchor + "</h" + level + ">"
		}
		return "<h" + level + " id=\"" + id + "\">" + text + " " + anchor + "</h" + level + ">"
	})

	return modifiedHTML, toc
}

// generateHeadingID creates a URL-safe ID from heading text
func generateHeadingID(text string) string {
	// Remove emojis and other non-ASCII characters first
	id := nonASCIIRegex.ReplaceAllString(text, "")

	// Trim spaces that may be left after emoji removal
	id = strings.TrimSpace(id)

	// Convert to lowercase
	id = strings.ToLower(id)

	// Replace spaces and special characters with hyphens
	id = nonAlphaNumeric.ReplaceAllString(id, "-")

	// Remove leading/trailing hyphens
	id = strings.Trim(id, "-")

	return id
}

// Template strings
const baseTemplate = `<!DOCTYPE html>
<html lang="en" data-lp-theme="{{.Site.Theme.ResolvedPreset}}">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="color-scheme" content="light dark">
  <script>
    // Resolve the theme before styles load to avoid flashing the wrong scheme.
    (function() {
      var preference = 'system';
      try {
        var storedPreference = localStorage.getItem('theme');
        if (storedPreference === 'light' || storedPreference === 'dark') {
          preference = storedPreference;
        }
      } catch (error) {
        // Storage can be unavailable in privacy-restricted browsing contexts.
      }
      var theme = preference === 'system'
        ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
        : preference;
      document.documentElement.setAttribute('data-theme-preference', preference);
      document.documentElement.setAttribute('data-theme', theme);
    })();
  </script>
  <title>{{block "title" .}}{{.Site.Title}}{{end}}</title>
  {{block "seo" .}}{{end}}
  <link rel="icon" type="image/svg+xml" href="{{.Site.BasePath}}/favicon.svg">
  <link rel="icon" type="image/png" sizes="96x96" href="{{.Site.BasePath}}/favicon-96x96.png">
  <link rel="icon" type="image/x-icon" href="{{.Site.BasePath}}/favicon.ico">
  {{if .Site.RSS}}<link rel="alternate" type="application/rss+xml" title="{{.Site.Title}}" href="{{.Site.BasePath}}/feed.xml">{{end}}
  {{range fontPreloads .Site.Theme}}<link rel="preload" href="{{$.Site.BasePath}}/{{.Path}}" as="font" type="{{.ContentType}}" crossorigin>
  {{end}}
  <style>
    :root {
      color-scheme: light;
      --lp-font-heading: "{{.Site.Theme.FontHeading}}", Georgia, serif;
      --lp-font-body: "{{.Site.Theme.FontBody}}", system-ui, -apple-system, sans-serif;
      --lp-font-mono: "{{.Site.Theme.FontMono}}", "Fira Code", "Courier New", monospace;
      --lp-accent: {{.Site.Theme.Accent}};
      --lp-font-xs: 0.75rem;
      --lp-font-sm: 0.875rem;
      --lp-font-base: 1rem;
      --lp-font-lg: 1.25rem;
      --lp-font-xl: 1.5rem;
      --lp-font-2xl: 1.75rem;
      --lp-font-3xl: 2rem;
      --lp-font-display: 6rem;
      --lp-radius-sm: 4px;
      --lp-radius-md: 8px;
      --lp-radius-lg: 12px;
      --lp-radius-full: 9999px;
      --lp-bg: #ffffff;
      --lp-text: #1a1a1a;
      --lp-text-muted: #666666;
      --lp-border: #e5e5e5;
      --lp-code-bg: #f7f7f7;
      --lp-max-width: 680px;
      --lp-nav-height: 60px;
    }
    {{if .Site.Theme.Background.Light}}
    :root {
      --lp-bg: {{.Site.Theme.Background.Light | safeCSS}};
    }
    {{end}}
    {{if eq .Site.Theme.NavStyle "sticky"}}
    .lp-nav {
      position: sticky;
      top: 0;
      z-index: 100;
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
    }
    {{end}}
    {{if eq .Site.Theme.NavStyle "glassy"}}
    .lp-nav {
      z-index: 100;
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
    }
    {{end}}

    [data-theme="dark"] {
      color-scheme: dark;
      --lp-bg: #1a1a1a;
      --lp-text: #e5e5e5;
      --lp-text-muted: #a0a0a0;
      --lp-border: #333333;
      --lp-code-bg: #2a2a2a;
    }
    {{if .Site.Theme.Background.Dark}}
    [data-theme="dark"] {
      --lp-bg: {{.Site.Theme.Background.Dark | safeCSS}};
    }
    {{end}}
  </style>
  <link rel="stylesheet" href="{{.Site.BasePath}}/style.css">
  {{if .Site.ClientScriptPath}}<script src="{{.Site.BasePath}}/{{.Site.ClientScriptPath}}" defer></script>{{end}}
  {{$remoteFontURL := remoteFontURL .Site.Theme}}{{if $remoteFontURL}}<link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="{{$remoteFontURL}}" rel="stylesheet">
  {{end}}{{if .Site.HeadExtra}}{{.Site.HeadExtra | safeHTML}}{{end}}
</head>
<body class="lp-body">
  {{if eq .Site.Theme.NavStyle "glassy"}}<div class="lp-nav-placeholder"></div>{{end}}
  <nav class="lp-nav">
    <div class="lp-nav-container">
      <div class="lp-nav-brand">
        <a class="lp-nav-title" href="{{.Site.BasePath}}/">{{.Site.Title}}</a>
        <div class="lp-nav-actions">
          {{if .Site.RSS}}<a href="{{.Site.BasePath}}/feed.xml" class="lp-rss-link" aria-label="RSS feed" title="RSS feed" target="_blank">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M4 11a9 9 0 0 1 9 9"></path>
              <path d="M4 4a16 16 0 0 1 16 16"></path>
              <circle cx="5" cy="19" r="1"></circle>
            </svg>
          </a>{{end}}
          {{if .Site.Graph}}<button class="lp-graph-toggle" aria-label="Open knowledge graph" title="Explore graph">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="6" cy="6" r="3"></circle>
              <circle cx="18" cy="6" r="3"></circle>
              <circle cx="6" cy="18" r="3"></circle>
              <circle cx="18" cy="18" r="3"></circle>
              <line x1="8.5" y1="7.5" x2="15.5" y2="16.5"></line>
              <line x1="8.5" y1="16.5" x2="15.5" y2="7.5"></line>
            </svg>
          </button>{{end}}
          {{if .Site.Search}}<button class="lp-search-toggle" aria-label="Search" title="Search (⌘K)">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="6"></circle>
              <line x1="21" y1="21" x2="15.5" y2="15.5"></line>
            </svg>
          </button>{{end}}
          <button class="lp-theme-toggle" aria-label="Change theme" title="Change theme">
            <svg class="lp-theme-icon lp-theme-icon-system" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="4" width="18" height="13" rx="2"></rect>
              <line x1="8" y1="21" x2="16" y2="21"></line>
              <line x1="12" y1="17" x2="12" y2="21"></line>
            </svg>
            <svg class="lp-theme-icon lp-theme-icon-light" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="5"></circle>
              <line x1="12" y1="1" x2="12" y2="3"></line>
              <line x1="12" y1="21" x2="12" y2="23"></line>
              <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
              <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
              <line x1="1" y1="12" x2="3" y2="12"></line>
              <line x1="21" y1="12" x2="23" y2="12"></line>
              <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
              <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
            </svg>
            <svg class="lp-theme-icon lp-theme-icon-dark" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
            </svg>
          </button>
        </div>
      </div>
      <div class="lp-nav-links">
        {{range .Site.Nav}}
        <a class="lp-nav-link{{if hasPrefix $.CurrentPath .Path}} lp-nav-link--active lp-nav-active-{{$.Site.Theme.NavActiveStyle}}{{end}}" href="{{$.Site.BasePath}}{{.Path}}">{{.Label}}</a>
        {{end}}
      </div>
    </div>
  </nav>
  <main class="lp-main">
    {{block "content" .}}{{end}}
  </main>
  <footer class="lp-footer">
    {{if .Site.Author}}<span class="lp-footer-text">&copy; {{.Site.Author}}. All rights reserved.</span>{{end}}
    {{if .Site.FooterAttribution}}<span class="lp-footer-text">Grown with {{if .Site.FooterAttribution.URL}}<a href="{{.Site.FooterAttribution.URL}}" target="_blank" rel="noopener noreferrer">{{.Site.FooterAttribution.Name}}</a>{{else}}{{.Site.FooterAttribution.Name}}{{end}}</span>
    {{else}}<span class="lp-footer-text">Grown with <a href="https://leafpress.in" target="_blank" rel="noopener noreferrer">leafpress</a></span>{{end}}
  </footer>

  {{if .Site.Graph}}<!-- Graph Overlay -->
  <div class="lp-graph-overlay" id="lp-graph-overlay" aria-hidden="true">
    <div class="lp-graph-backdrop"></div>
    <div class="lp-graph-panel" role="dialog" aria-label="Knowledge Graph" data-current-slug="{{block "currentSlug" .}}{{end}}">
      <button class="lp-graph-close" aria-label="Close graph">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
      <div class="lp-graph-panel-body" id="lp-graph-panel-body"></div>
    </div>
  </div>{{end}}

  {{if .Site.Search}}<!-- Search Overlay -->
  <div class="lp-search-overlay" id="lp-search-overlay" aria-hidden="true">
    <div class="lp-search-backdrop"></div>
    <div class="lp-search-panel" role="dialog" aria-label="Search">
      <div class="lp-search-header">
        <svg class="lp-search-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="6"></circle>
          <line x1="21" y1="21" x2="15.5" y2="15.5"></line>
        </svg>
        <input type="text" class="lp-search-input" id="lp-search-input" placeholder="Search pages..." autocomplete="off" autofocus>
        <kbd class="lp-search-kbd">ESC</kbd>
        <button class="lp-search-close" aria-label="Close search">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>
      <div class="lp-search-results" id="lp-search-results"></div>
    </div>
  </div>{{end}}

  {{if not .Site.ClientScriptPath}}<script>{{template "clientScriptMain" .}}</script>
  <script>{{template "clientScriptMermaid" .}}</script>{{end}}
  {{define "clientScriptMain"}}
    // Base path for asset loading (supports GitHub Pages subdirectory hosting)
    var LP_BASE_PATH = '{{.Site.BasePath}}';

    // Theme switching
    var themeMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    var themePreference = document.documentElement.getAttribute('data-theme-preference') || 'system';

    function getSystemTheme() {
      return themeMediaQuery.matches ? 'dark' : 'light';
    }

    function getEffectiveTheme(preference) {
      return preference === 'system' ? getSystemTheme() : preference;
    }

    function getNextThemePreference() {
      if (themePreference === 'system') {
        return getEffectiveTheme(themePreference) === 'dark' ? 'light' : 'dark';
      }
      if (themePreference !== getSystemTheme()) {
        return getSystemTheme();
      }
      return 'system';
    }

    function updateThemeToggle(theme) {
      var themeToggle = document.querySelector('.lp-theme-toggle');
      if (!themeToggle) return;

      var nextPreference = getNextThemePreference();

      var currentLabel = themePreference === 'system'
        ? 'system (' + theme + ')'
        : themePreference;
      var nextLabel = nextPreference === 'system' ? 'system setting' : nextPreference + ' theme';
      var label = 'Theme: ' + currentLabel + '. Use ' + nextLabel;
      themeToggle.setAttribute('aria-label', label);
      themeToggle.setAttribute('title', label);
    }

    function updateGraphTheme(theme) {
      var graphBody = document.getElementById('lp-graph-panel-body');
      if (!graphBody || !graphBody.querySelector('svg')) return;

      var linkColor = theme === 'dark' ? '#444444' : '#d0d0d0';
      var textColor = getComputedStyle(document.documentElement).getPropertyValue('--lp-text').trim();
      var accentColor = getComputedStyle(document.documentElement).getPropertyValue('--lp-accent').trim();
      graphBody.querySelectorAll('.lp-graph-link').forEach(function(link) {
        if (!link.style.opacity || link.style.opacity === '0.5') {
          link.setAttribute('stroke', linkColor);
        }
      });
      graphBody.querySelectorAll('.lp-graph-label').forEach(function(label) {
        label.style.fill = textColor;
      });
      graphBody.querySelectorAll('.lp-graph-node').forEach(function(node) {
        node.setAttribute('fill', accentColor);
      });
    }

    function applyThemePreference(preference) {
      themePreference = preference;
      var theme = getEffectiveTheme(preference);
      document.documentElement.setAttribute('data-theme-preference', preference);
      document.documentElement.setAttribute('data-theme', theme);
      updateThemeToggle(theme);
      updateGraphTheme(theme);
    }

    function storeThemePreference(preference) {
      try {
        if (preference === 'system') {
          localStorage.removeItem('theme');
        } else {
          localStorage.setItem('theme', preference);
        }
      } catch (error) {
        // Keep theme switching functional even when storage is unavailable.
      }
    }

    function handleSystemThemeChange() {
      if (themePreference === 'system') {
        applyThemePreference('system');
      } else {
        updateThemeToggle(getEffectiveTheme(themePreference));
      }
    }

    if (themeMediaQuery.addEventListener) {
      themeMediaQuery.addEventListener('change', handleSystemThemeChange);
    } else if (themeMediaQuery.addListener) {
      // Safari before 14 uses the legacy MediaQueryList listener API.
      themeMediaQuery.addListener(handleSystemThemeChange);
    }

    window.addEventListener('storage', function(event) {
      if (event.key !== 'theme' && event.key !== null) return;
      var preference = event.newValue === 'light' || event.newValue === 'dark'
        ? event.newValue
        : 'system';
      applyThemePreference(preference);
    });

    // Add copy buttons to code blocks
    document.addEventListener('DOMContentLoaded', function() {
      // Theme toggle
      var themeToggle = document.querySelector('.lp-theme-toggle');
      if (themeToggle) {
        updateThemeToggle(getEffectiveTheme(themePreference));
        themeToggle.addEventListener('click', function() {
          var nextPreference = getNextThemePreference();
          storeThemePreference(nextPreference);
          applyThemePreference(nextPreference);
        });
      }

      {{if eq .Site.Theme.NavStyle "glassy"}}
      // Floating pill navbar on scroll
      var nav = document.querySelector('.lp-nav');
      var navPlaceholder = document.querySelector('.lp-nav-placeholder');
      if (nav && navPlaceholder) {
        var navHeight = nav.offsetHeight;
        navPlaceholder.style.height = navHeight + 'px';

        window.addEventListener('scroll', function() {
          if (window.scrollY > navHeight) {
            nav.classList.add('lp-nav--pill');
            navPlaceholder.classList.add('lp-nav-placeholder--active');
          } else {
            nav.classList.remove('lp-nav--pill');
            navPlaceholder.classList.remove('lp-nav-placeholder--active');
          }
        });
      }
      {{end}}

      // Copy buttons
      document.querySelectorAll('pre.chroma').forEach(function(pre) {
        var button = document.createElement('button');
        button.className = 'lp-copy-button';
        button.textContent = 'Copy';
        button.setAttribute('aria-label', 'Copy code to clipboard');

        button.addEventListener('click', function() {
          var code = pre.querySelector('code').textContent;
          navigator.clipboard.writeText(code).then(function() {
            button.textContent = 'Copied!';
            setTimeout(function() {
              button.textContent = 'Copy';
            }, 2000);
          }).catch(function() {
            button.textContent = 'Failed';
            setTimeout(function() {
              button.textContent = 'Copy';
            }, 2000);
          });
        });

        pre.style.position = 'relative';
        pre.appendChild(button);
      });
      {{if .Site.Graph}}
      // Graph Overlay
      (function() {
        var overlay = document.getElementById('lp-graph-overlay');
        var panel = overlay.querySelector('.lp-graph-panel');
        var graphBody = document.getElementById('lp-graph-panel-body');
        var toggleBtn = document.querySelector('.lp-graph-toggle');
        var closeBtn = overlay.querySelector('.lp-graph-close');
        var backdrop = overlay.querySelector('.lp-graph-backdrop');
        var currentSlug = panel.getAttribute('data-current-slug') || '';
        var graphData = null;
        var graphRendered = false;

        function openGraph() {
          overlay.classList.add('lp-graph-overlay--open');
          overlay.setAttribute('aria-hidden', 'false');
          document.body.style.overflow = 'hidden';

          if (!graphRendered && graphData) {
            renderGraph(graphData);
            graphRendered = true;
          } else if (!graphData) {
            fetch(LP_BASE_PATH + '/graph.json')
              .then(function(r) { return r.json(); })
              .then(function(data) {
                graphData = data;
                renderGraph(data);
                graphRendered = true;
              });
          }
        }

        function closeGraph() {
          overlay.classList.remove('lp-graph-overlay--open');
          overlay.setAttribute('aria-hidden', 'true');
          document.body.style.overflow = '';
        }

        toggleBtn.addEventListener('click', openGraph);
        closeBtn.addEventListener('click', closeGraph);
        backdrop.addEventListener('click', closeGraph);

        document.addEventListener('keydown', function(e) {
          if (e.key === 'Escape' && overlay.classList.contains('lp-graph-overlay--open')) {
            closeGraph();
          }
        });

        function renderGraph(data) {
          var width = graphBody.offsetWidth;
          var height = graphBody.offsetHeight;

          var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
          svg.setAttribute('width', width);
          svg.setAttribute('height', height);
          svg.setAttribute('viewBox', '0 0 ' + width + ' ' + height);
          graphBody.appendChild(svg);

          // Pass 1: Group nodes by primary tag for initial placement
          var tagGroups = new Map();
          var untaggedNodes = [];
          data.nodes.forEach(function(d) {
            var primaryTag = (d.tags && d.tags.length > 0) ? d.tags[0] : null;
            if (primaryTag) {
              if (!tagGroups.has(primaryTag)) tagGroups.set(primaryTag, []);
              tagGroups.get(primaryTag).push(d);
            } else {
              untaggedNodes.push(d);
            }
          });

          // Assign positions by tag group (arrange in sectors around center)
          var tagNames = Array.from(tagGroups.keys());
          var numGroups = tagNames.length;
          var centerX = width / 2;
          var centerY = height / 2;
          var radius = Math.min(width, height) * 0.3;

          var nodes = [];
          tagNames.forEach(function(tag, groupIndex) {
            var angle = (2 * Math.PI * groupIndex) / numGroups;
            var groupCenterX = centerX + radius * Math.cos(angle);
            var groupCenterY = centerY + radius * Math.sin(angle);
            var groupNodes = tagGroups.get(tag);

            groupNodes.forEach(function(d, i) {
              // Spread nodes within group
              var spread = 50;
              var offsetAngle = (2 * Math.PI * i) / groupNodes.length;
              nodes.push({
                id: d.id,
                title: d.title,
                url: d.url,
                tags: d.tags || [],
                x: groupCenterX + spread * Math.cos(offsetAngle) * (0.5 + Math.random() * 0.5),
                y: groupCenterY + spread * Math.sin(offsetAngle) * (0.5 + Math.random() * 0.5),
                vx: 0,
                vy: 0
              });
            });
          });

          // Untagged nodes go near center with some randomness
          untaggedNodes.forEach(function(d) {
            nodes.push({
              id: d.id,
              title: d.title,
              url: d.url,
              tags: d.tags || [],
              x: centerX + (Math.random() - 0.5) * 100,
              y: centerY + (Math.random() - 0.5) * 100,
              vx: 0,
              vy: 0
            });
          });

          var nodeMap = new Map();
          nodes.forEach(function(n) { nodeMap.set(n.id, n); });

          var links = [];
          data.edges.forEach(function(edge) {
            var source = nodeMap.get(edge.source);
            var target = nodeMap.get(edge.target);
            if (source && target) {
              links.push({ source: source, target: target, sourceId: edge.source, targetId: edge.target });
            }
          });

          // Calculate node degrees and build adjacency list for clustering
          nodes.forEach(function(n) {
            n.degree = 0;
            n.neighbors = [];
            n.neighborIds = new Set();
            n.tagSet = new Set(n.tags);
          });
          links.forEach(function(link) {
            link.source.degree++;
            link.target.degree++;
            link.source.neighbors.push(link.target);
            link.target.neighbors.push(link.source);
            link.source.neighborIds.add(link.target.id);
            link.target.neighborIds.add(link.source.id);
          });
          nodes.forEach(function(node) {
            node.clusterIds = new Set();
            node.neighbors.forEach(function(neighbor) {
              neighbor.neighbors.forEach(function(candidate) {
                if (candidate !== node && !node.neighborIds.has(candidate.id)) {
                  node.clusterIds.add(candidate.id);
                }
              });
            });
          });
          var attractionTagGroups = new Map();
          nodes.forEach(function(node) {
            node.tags.forEach(function(tag) {
              if (!attractionTagGroups.has(tag)) attractionTagGroups.set(tag, []);
              attractionTagGroups.get(tag).push(node);
            });
          });
          var maxDegree = Math.max.apply(null, nodes.map(function(n) { return n.degree; })) || 1;

          // Check if two nodes share neighbors (for clustering)
          function shareNeighbors(a, b) {
            return a.clusterIds.has(b.id);
          }

          // Check if two nodes are directly connected
          function areConnected(a, b) {
            return a.neighborIds.has(b.id);
          }

          // Count shared tags between two nodes (for tag-based clustering)
          function sharedTagCount(a, b) {
            var count = 0;
            var smaller = a.tagSet.size < b.tagSet.size ? a.tagSet : b.tagSet;
            var larger = smaller === a.tagSet ? b.tagSet : a.tagSet;
            smaller.forEach(function(tag) {
              if (larger.has(tag)) count++;
            });
            return count;
          }

          // Centrality score: normalized degree (0-1)
          function getCentrality(node) {
            return node.degree / maxDegree;
          }

          var linkGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
          svg.appendChild(linkGroup);

          var nodeGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
          svg.appendChild(nodeGroup);

          var labelGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
          svg.appendChild(labelGroup);

          var isDark = document.documentElement.getAttribute('data-theme') === 'dark';
          var linkColor = isDark ? '#444444' : '#d0d0d0';
          var accentColor = getComputedStyle(document.documentElement).getPropertyValue('--lp-accent').trim();

          links.forEach(function(link) {
            var line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
            line.setAttribute('class', 'lp-graph-link');
            line.setAttribute('stroke', linkColor);
            line.setAttribute('stroke-width', '1.5');
            line.setAttribute('stroke-opacity', '0.5');
            linkGroup.appendChild(line);
            link.element = line;
          });

          var selectedNode = null;

          // Node opacity based on link density (degree)
          function getNodeOpacity(degree) {
            // More connections = more opaque (0.15 to 1.0 for better contrast)
            return 0.15 + (degree / maxDegree) * 0.85;
          }

          nodes.forEach(function(node) {
            var circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
            circle.setAttribute('class', 'lp-graph-node');
            circle.setAttribute('r', '6');
            circle.setAttribute('fill', accentColor);
            circle.setAttribute('fill-opacity', getNodeOpacity(node.degree));
            circle.setAttribute('stroke', '#fff');
            circle.setAttribute('stroke-width', '2');
            circle.style.cursor = 'pointer';

            // Mark current page node
            if (node.id === currentSlug) {
              circle.classList.add('lp-graph-node--current');
            }

            // Hover for preview highlight
            circle.addEventListener('mouseenter', function() {
              if (!selectedNode) {
                highlightConnections(node);
              }
            });

            circle.addEventListener('mouseleave', function() {
              if (!selectedNode) {
                clearHighlight();
              }
            });

            // Click to lock selection, second click to navigate
            circle.addEventListener('click', function(e) {
              e.preventDefault();
              if (selectedNode === node) {
                // Second click - navigate
                window.location.href = node.url || '/';
              } else {
                // First click - lock highlight
                selectedNode = node;
                highlightConnections(node);
              }
            });

            nodeGroup.appendChild(circle);
            node.element = circle;

            var text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
            text.setAttribute('class', 'lp-graph-label');
            text.setAttribute('text-anchor', 'middle');
            text.setAttribute('font-size', '0.5em');
            text.setAttribute('pointer-events', 'none');
            text.style.opacity = '0';
            text.style.fill = getComputedStyle(document.documentElement).getPropertyValue('--lp-text').trim();

            // Split long titles into multiple lines
            var title = node.title || 'Home';
            var maxChars = 18;
            var lines = [];

            if (title.length <= maxChars) {
              lines.push(title);
            } else {
              // Split into words and create lines
              var words = title.split(/[\s-]+/);
              var currentLine = '';

              words.forEach(function(word) {
                if ((currentLine + ' ' + word).trim().length <= maxChars) {
                  currentLine = (currentLine + ' ' + word).trim();
                } else {
                  if (currentLine) lines.push(currentLine);
                  currentLine = word;
                }
              });
              if (currentLine) lines.push(currentLine);

              // Limit to 2 lines max
              if (lines.length > 2) {
                lines = [lines[0], lines[1].substring(0, maxChars - 3) + '...'];
              }
            }

            // Store lines for positioning after simulation
            node.labelLines = lines;

            labelGroup.appendChild(text);
            node.label = text;
          });

          // Click on empty space clears selection
          svg.addEventListener('click', function(e) {
            if (e.target === svg) {
              selectedNode = null;
              clearHighlight();
            }
          });

          function highlightConnections(selected) {
            var currentAccentColor = getComputedStyle(document.documentElement).getPropertyValue('--lp-accent').trim();
            nodes.forEach(function(n) {
              n.element.style.opacity = '0.15';
              if (n.label) n.label.style.opacity = '0';
            });
            links.forEach(function(l) {
              l.element.style.opacity = '0.05';
            });

            selected.element.style.opacity = '1';
            selected.element.setAttribute('r', '8');
            if (selected.label) selected.label.style.opacity = '1';

            links.forEach(function(link) {
              if (link.sourceId === selected.id || link.targetId === selected.id) {
                link.element.style.opacity = '0.8';
                link.element.setAttribute('stroke', currentAccentColor);
                link.element.setAttribute('stroke-width', '2.5');

                var connected = nodeMap.get(link.sourceId === selected.id ? link.targetId : link.sourceId);
                if (connected) {
                  connected.element.style.opacity = '1';
                  connected.element.setAttribute('r', '7');
                  if (connected.label) connected.label.style.opacity = '0.9';
                }
              }
            });
          }

          function clearHighlight() {
            var currentLinkColor = document.documentElement.getAttribute('data-theme') === 'dark' ? '#444444' : '#d0d0d0';
            nodes.forEach(function(n) {
              n.element.style.opacity = '1';
              n.element.setAttribute('r', n.id === currentSlug ? '8' : '6');
              if (n.label) n.label.style.opacity = n.id === currentSlug ? '1' : '0';
            });
            links.forEach(function(l) {
              l.element.style.opacity = '0.5';
              l.element.setAttribute('stroke', currentLinkColor);
              l.element.setAttribute('stroke-width', '1.5');
            });
          }

          // Pass 2: frame-budgeted physics with spatially indexed repulsion.
          // Link and tag attraction are linear in graph size; node repulsion
          // only checks nearby grid cells instead of every pair on every tick.
          function simulate() {
            var n = nodes.length;
            if (n === 0) return;

            var area = width * height;
            var idealSpacing = Math.sqrt(area / n);

            // Link distance: longer for better spread
            var linkRestLength = Math.max(120, Math.min(280, idealSpacing * 0.75));
            var collisionRadius = 25;
            var repulsionStrength = idealSpacing * idealSpacing * 1.2;
            var centerForce = 0.006;
            var iterations = Math.min(240, 80 + n * 2);
            if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
              iterations = Math.min(iterations, 80);
            }
            var padding = 35;
            var alpha = 0.3;
            var alphaDecay = Math.pow(0.01, 1 / iterations);
            var iteration = 0;
            var cellSize = Math.max(collisionRadius * 2, idealSpacing * 1.5);
            var repulsionRange = cellSize * 2;

            function applyPairForce(a, b) {
              var dx = b.x - a.x;
              var dy = b.y - a.y;
              var distSquared = dx * dx + dy * dy;
              if (distSquared > repulsionRange * repulsionRange) return;
              if (distSquared < 1) {
                dx = (Math.random() - 0.5) * 2;
                dy = (Math.random() - 0.5) * 2;
                distSquared = 1;
              }
              var dist = Math.sqrt(distSquared);
              var related = areConnected(a, b) || sharedTagCount(a, b) > 0 || shareNeighbors(a, b);
              var force = -repulsionStrength * (related ? 0.2 : 1) / distSquared;
              if (dist < collisionRadius * 2) {
                force -= (collisionRadius * 2 - dist) * 3;
              }
              var fx = (force * dx) / dist;
              var fy = (force * dy) / dist;
              a.vx += fx;
              a.vy += fy;
              b.vx -= fx;
              b.vy -= fy;
            }

            function tick() {
              nodes.forEach(function(node) { node.vx = 0; node.vy = 0; });

              // Spatial hash: typical work is proportional to nearby nodes,
              // while dense cells retain exact pairwise collision handling.
              var grid = new Map();
              nodes.forEach(function(node, index) {
                var gx = Math.floor(node.x / cellSize);
                var gy = Math.floor(node.y / cellSize);
                var key = gx + ',' + gy;
                if (!grid.has(key)) grid.set(key, []);
                grid.get(key).push(index);
              });
              nodes.forEach(function(a, i) {
                var gx = Math.floor(a.x / cellSize);
                var gy = Math.floor(a.y / cellSize);
                for (var ox = -2; ox <= 2; ox++) {
                  for (var oy = -2; oy <= 2; oy++) {
                    var nearby = grid.get((gx + ox) + ',' + (gy + oy));
                    if (!nearby) continue;
                    for (var p = 0; p < nearby.length; p++) {
                      var j = nearby[p];
                      if (j > i) applyPairForce(a, nodes[j]);
                    }
                  }
                }
              });

              // Exact spring forces only traverse real graph edges.
              links.forEach(function(link) {
                var a = link.source;
                var b = link.target;
                var dx = b.x - a.x;
                var dy = b.y - a.y;
                var dist = Math.max(1, Math.sqrt(dx * dx + dy * dy));
                var centralityMult = 1 + (getCentrality(a) + getCentrality(b)) * 0.5;
                var force = (dist - linkRestLength) * 0.1 * centralityMult;
                var fx = (force * dx) / dist;
                var fy = (force * dy) / dist;
                a.vx += fx;
                a.vy += fy;
                b.vx -= fx;
                b.vy -= fy;
              });

              // Pull tagged nodes toward their group centroid in O(nodes).
              attractionTagGroups.forEach(function(groupNodes) {
                if (!groupNodes || groupNodes.length < 2) return;
                var sumX = 0;
                var sumY = 0;
                groupNodes.forEach(function(node) {
                  sumX += node.x;
                  sumY += node.y;
                });
                var groupX = sumX / groupNodes.length;
                var groupY = sumY / groupNodes.length;
                groupNodes.forEach(function(node) {
                  node.vx += (groupX - node.x) * 0.025;
                  node.vy += (groupY - node.y) * 0.025;
                });
              });

              var cx = width / 2;
              var cy = height / 2;
              nodes.forEach(function(node) {
                var dx = cx - node.x;
                var dy = cy - node.y;
                node.vx += dx * centerForce;
                node.vy += dy * centerForce;
              });

              nodes.forEach(function(node) {
                node.vx *= 0.85;
                node.vy *= 0.85;
                node.x += node.vx * alpha;
                node.y += node.vy * alpha;
                node.x = Math.max(padding, Math.min(width - padding, node.x));
                node.y = Math.max(padding, Math.min(height - padding, node.y));
              });
              alpha *= alphaDecay;
              iteration++;
            }

            function draw() {
              var layoutCenterY = height / 2;
              nodes.forEach(function(node) {
                node.element.setAttribute('cx', node.x);
                node.element.setAttribute('cy', node.y);
                if (node.label && node.labelLines) {
                  if (!node.labelSpans) {
                    node.labelSpans = [];
                    node.labelLines.forEach(function(line) {
                      var tspan = document.createElementNS('http://www.w3.org/2000/svg', 'tspan');
                      tspan.textContent = line;
                      node.label.appendChild(tspan);
                      node.labelSpans.push(tspan);
                    });
                  }
                  var labelBelow = node.y < layoutCenterY;
                  var lineHeight = 12;
                  var offset = labelBelow ? 16 : -(8 + (node.labelLines.length - 1) * lineHeight);
                  node.label.setAttribute('x', node.x);
                  node.label.setAttribute('y', node.y);
                  node.labelSpans.forEach(function(tspan, idx) {
                    tspan.setAttribute('x', node.x);
                    tspan.setAttribute('dy', idx === 0 ? offset : lineHeight);
                  });
                }
              });
              links.forEach(function(link) {
                link.element.setAttribute('x1', link.source.x);
                link.element.setAttribute('y1', link.source.y);
                link.element.setAttribute('x2', link.target.x);
                link.element.setAttribute('y2', link.target.y);
              });
            }

            function runFrame() {
              var started = performance.now();
              do {
                tick();
              } while (iteration < iterations && alpha >= 0.005 && performance.now() - started < 10);
              draw();
              if (iteration < iterations && alpha >= 0.005) {
                requestAnimationFrame(runFrame);
                return;
              }
              if (currentSlug) {
                var current = nodeMap.get(currentSlug);
                if (current) {
                  current.element.setAttribute('r', '8');
                  if (current.label) current.label.style.opacity = '1';
                }
              }
            }

            draw();
            requestAnimationFrame(runFrame);
          }

          simulate();
        }
      })();
      {{end}}
      {{if .Site.Search}}
      // Search functionality
      (function() {
        var overlay = document.getElementById('lp-search-overlay');
        var input = document.getElementById('lp-search-input');
        var results = document.getElementById('lp-search-results');
        var toggleBtn = document.querySelector('.lp-search-toggle');
        var backdrop = overlay.querySelector('.lp-search-backdrop');
        var searchIndex = null;
        var selectedIndex = -1;

        function openSearch() {
          overlay.classList.add('lp-search-overlay--open');
          overlay.setAttribute('aria-hidden', 'false');
          document.body.style.overflow = 'hidden';
          input.value = '';
          results.textContent = '';
          selectedIndex = -1;

          // Focus input - immediate focus for mobile touch events
          input.focus();
          // Backup focus after transition completes
          setTimeout(function() { input.focus(); }, 200);

          if (!searchIndex) {
            fetch(LP_BASE_PATH + '/search-index.json')
              .then(function(r) { return r.json(); })
              .then(function(data) { searchIndex = data; });
          }
        }

        function closeSearch() {
          overlay.classList.remove('lp-search-overlay--open');
          overlay.setAttribute('aria-hidden', 'true');
          document.body.style.overflow = '';
        }

        function search(query) {
          if (!searchIndex || !query.trim()) {
            results.textContent = '';
            selectedIndex = -1;
            return;
          }

          var q = query.toLowerCase();
          var scored = [];
          searchIndex.forEach(function(item) {
            var titleLower = item.title.toLowerCase();
            var contentLower = item.content.toLowerCase();
            var score = 0;

            // Title matches (highest priority)
            if (titleLower === q) {
              score = 100; // Exact title match
            } else if (titleLower.indexOf(q) === 0) {
              score = 80; // Title starts with query
            } else if (titleLower.indexOf(q) !== -1) {
              score = 60; // Title contains query
            }

            // Tag matches
            if (item.tags && item.tags.some(function(t) { return t.toLowerCase().indexOf(q) !== -1; })) {
              score = Math.max(score, 40);
            }

            // Content matches (lowest priority)
            if (contentLower.indexOf(q) !== -1) {
              score = Math.max(score, 20);
            }

            if (score > 0) {
              scored.push({ item: item, score: score });
            }
          });

          // Sort by score descending
          scored.sort(function(a, b) { return b.score - a.score; });
          var matches = scored.slice(0, 10).map(function(s) { return s.item; });

          if (matches.length === 0) {
            results.textContent = '';
            var empty = document.createElement('div');
            empty.className = 'lp-search-empty';
            empty.textContent = 'No results found';
            results.appendChild(empty);
            selectedIndex = -1;
            return;
          }

          results.textContent = '';
          matches.forEach(function(item, i) {
            var snippet = getSnippet(item.content, q);
            var link = document.createElement('a');
            link.className = 'lp-search-result';
            link.setAttribute('href', item.url);
            link.dataset.index = String(i);

            var title = document.createElement('span');
            title.className = 'lp-search-result-title';
            appendHighlightedText(title, item.title, q);
            link.appendChild(title);

            if (snippet) {
              var snippetEl = document.createElement('span');
              snippetEl.className = 'lp-search-result-snippet';
              appendHighlightedText(snippetEl, snippet, q);
              link.appendChild(snippetEl);
            }

            results.appendChild(link);
          });
          selectedIndex = -1;
        }

        function getSnippet(content, query) {
          var idx = content.toLowerCase().indexOf(query);
          if (idx === -1) return '';
          var start = Math.max(0, idx - 40);
          var end = Math.min(content.length, idx + query.length + 60);
          var snippet = content.substring(start, end);
          if (start > 0) snippet = '...' + snippet;
          if (end < content.length) snippet = snippet + '...';
          return snippet;
        }

        function appendHighlightedText(parent, text, query) {
          text = String(text || '');
          var lowerText = text.toLowerCase();
          var lowerQuery = query.toLowerCase();
          var offset = 0;
          var match;

          while ((match = lowerText.indexOf(lowerQuery, offset)) !== -1) {
            parent.appendChild(document.createTextNode(text.substring(offset, match)));
            var mark = document.createElement('mark');
            mark.textContent = text.substring(match, match + query.length);
            parent.appendChild(mark);
            offset = match + query.length;
          }
          parent.appendChild(document.createTextNode(text.substring(offset)));
        }

        function updateSelection() {
          var items = results.querySelectorAll('.lp-search-result');
          items.forEach(function(item, i) {
            item.classList.toggle('lp-search-result--selected', i === selectedIndex);
          });
          if (selectedIndex >= 0 && items[selectedIndex]) {
            items[selectedIndex].scrollIntoView({ block: 'nearest' });
          }
        }

        var closeBtn = overlay.querySelector('.lp-search-close');

        if (toggleBtn) toggleBtn.addEventListener('click', openSearch);
        backdrop.addEventListener('click', closeSearch);
        if (closeBtn) closeBtn.addEventListener('click', closeSearch);

        input.addEventListener('input', function() {
          search(input.value);
        });

        input.addEventListener('keydown', function(e) {
          var items = results.querySelectorAll('.lp-search-result');
          if (e.key === 'ArrowDown') {
            e.preventDefault();
            selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
            updateSelection();
          } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            selectedIndex = Math.max(selectedIndex - 1, -1);
            updateSelection();
          } else if (e.key === 'Enter' && selectedIndex >= 0 && items[selectedIndex]) {
            e.preventDefault();
            window.location.href = items[selectedIndex].getAttribute('href');
          }
        });

        document.addEventListener('keydown', function(e) {
          if (e.key === 'Escape' && overlay.classList.contains('lp-search-overlay--open')) {
            closeSearch();
          }
          if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
            e.preventDefault();
            if (overlay.classList.contains('lp-search-overlay--open')) {
              closeSearch();
            } else {
              openSearch();
            }
          }
        });
      })();
      {{end}}

      // Link preview on hover. Uses search-index.json (always emitted) so
      // previews keep working when the search UI is disabled.
      (function() {
        var previewEl = null;
        var previewIndex = null;
        var hideTimeout = null;
        var currentLink = null;

        function createPreview() {
          if (previewEl) return;
          previewEl = document.createElement('div');
          previewEl.className = 'lp-link-preview';
          previewEl.innerHTML = '<div class="lp-link-preview-title"></div><div class="lp-link-preview-content"></div>';
          document.body.appendChild(previewEl);

          previewEl.addEventListener('mouseenter', function() {
            clearTimeout(hideTimeout);
          });
          previewEl.addEventListener('mouseleave', function() {
            hidePreview();
          });
        }

        function showPreview(link, item) {
          createPreview();
          clearTimeout(hideTimeout);
          currentLink = link;

          var title = previewEl.querySelector('.lp-link-preview-title');
          var content = previewEl.querySelector('.lp-link-preview-content');
          title.textContent = item.title;
          content.textContent = item.content.substring(0, 200) + (item.content.length > 200 ? '...' : '');

          var rect = link.getBoundingClientRect();
          var scrollTop = window.pageYOffset || document.documentElement.scrollTop;
          var scrollLeft = window.pageXOffset || document.documentElement.scrollLeft;

          previewEl.style.display = 'block';
          previewEl.style.opacity = '0';

          // Position below link by default
          var top = rect.bottom + scrollTop + 8;
          var left = rect.left + scrollLeft;

          // Check if preview would go off-screen bottom
          var previewHeight = previewEl.offsetHeight;
          if (rect.bottom + previewHeight + 20 > window.innerHeight) {
            top = rect.top + scrollTop - previewHeight - 8;
          }

          // Check if preview would go off-screen right
          var previewWidth = previewEl.offsetWidth;
          if (left + previewWidth > window.innerWidth - 20) {
            left = window.innerWidth - previewWidth - 20;
          }

          previewEl.style.top = top + 'px';
          previewEl.style.left = left + 'px';
          previewEl.style.opacity = '1';
        }

        function hidePreview() {
          hideTimeout = setTimeout(function() {
            if (previewEl) {
              previewEl.style.display = 'none';
            }
            currentLink = null;
          }, 100);
        }

        function loadPreviewIndex(callback) {
          if (previewIndex) {
            callback(previewIndex);
            return;
          }
          fetch(LP_BASE_PATH + '/search-index.json')
            .then(function(r) { return r.json(); })
            .then(function(data) {
              previewIndex = {};
              data.forEach(function(item) {
                previewIndex[item.url] = item;
              });
              callback(previewIndex);
            })
            .catch(function() {
              previewIndex = {};
              callback(previewIndex);
            });
        }

        // Attach to all wikilinks and backlinks
        document.querySelectorAll('.lp-wikilink, .lp-backlink').forEach(function(link) {
          var url = link.getAttribute('href');

          link.addEventListener('mouseenter', function() {
            loadPreviewIndex(function(index) {
              var item = index[url];
              if (item) {
                showPreview(link, item);
              }
            });
          });

          link.addEventListener('mouseleave', function() {
            hidePreview();
          });
        });
      })();
    });

  {{end}}
  {{define "clientScriptMermaid"}}if (document.querySelector('.mermaid')) {
      var s = document.createElement('script');
      s.src = LP_BASE_PATH + '/static/leafpress/mermaid/mermaid.min.js';
      s.onload = function() {
        mermaid.initialize({ startOnLoad: false, securityLevel: 'strict', theme: 'default', htmlLabels: false, flowchart: { htmlLabels: false, useHtmlLabels: false }, sequence: { useHtmlLabels: false } });
        mermaid.run();
      };
      document.body.appendChild(s);
    }
  {{end}}
</body>
</html>
`

const pageTemplate = `
{{define "title"}}{{if eq .Page.Slug ""}}{{.Site.Title}}{{else}}{{.Page.Title}} | {{.Site.Title}}{{end}}{{end}}
{{define "currentSlug"}}{{.Page.Slug}}{{end}}
{{define "seo"}}
  <meta name="description" content="{{.Page.SEODescription}}">
  {{if .Site.BaseURL}}<link rel="canonical" href="{{.Site.BaseURL}}{{.Page.Permalink}}">{{end}}
  <meta property="og:title" content="{{.Page.Title}}">
  <meta property="og:description" content="{{.Page.SEODescription}}">
  <meta property="og:type" content="article">
  <meta property="og:site_name" content="{{.Site.Title}}">
  {{if .Site.BaseURL}}<meta property="og:url" content="{{.Site.BaseURL}}{{.Page.Permalink}}">{{end}}
  {{if .Page.Image}}<meta property="og:image" content="{{if .Site.BaseURL}}{{.Site.BaseURL}}{{end}}{{.Page.Image}}">
  {{else if .Site.Image}}<meta property="og:image" content="{{if .Site.BaseURL}}{{.Site.BaseURL}}{{end}}{{.Site.Image}}">{{end}}
  <meta name="twitter:card" content="summary">
  <meta name="twitter:title" content="{{.Page.Title}}">
  <meta name="twitter:description" content="{{.Page.SEODescription}}">
{{end}}
{{define "content"}}
<div class="lp-page-container">
  {{if .TOC}}
  <aside class="lp-toc">
    <nav class="lp-toc-nav">
      <ul class="lp-toc-list">
        {{range .TOC}}
        <li class="lp-toc-item lp-toc-level-{{.Level}}">
          <a href="#{{.ID}}" class="lp-toc-link">{{.Text}}</a>
        </li>
        {{end}}
      </ul>
    </nav>
  </aside>
  {{end}}

  <article class="lp-article">
    <header class="lp-header">
      <h1 class="lp-title">{{.Page.Title}}</h1>
      <div class="lp-meta">
        {{if .Page.Growth}}
        <span class="lp-growth lp-growth--{{.Page.Growth}}" data-growth="{{.Page.Growth}}">{{growthEmoji .Page.Growth}}</span>
        {{end}}
        {{if .Page.ReadingTime}}
        <span class="lp-reading-time">{{.Page.ReadingTimeDisplay}}</span>
        {{end}}
        {{if and .Page.HasModified (not .Page.Date.IsZero)}}
        <span class="lp-date-info">Updated <time class="lp-modified" datetime="{{.Page.ISOModified}}">{{.Page.FormattedModified}}</time> · Created <time class="lp-date" datetime="{{.Page.ISODate}}">{{.Page.FormattedDate}}</time></span>
        {{else if .Page.HasModified}}
        <span class="lp-date-info">Updated <time class="lp-modified" datetime="{{.Page.ISOModified}}">{{.Page.FormattedModified}}</time></span>
        {{else if not .Page.Date.IsZero}}
        <span class="lp-date-info">Created <time class="lp-date" datetime="{{.Page.ISODate}}">{{.Page.FormattedDate}}</time></span>
        {{end}}
      </div>
      {{if .Page.Tags}}
      <div class="lp-tags">
        {{range .Page.Tags}}
        <a class="lp-tag" href="{{$.Site.BasePath}}/tags/{{. | lower}}/">#{{.}}</a>
        {{end}}
      </div>
      {{end}}
    </header>

    <div class="lp-content">
      {{.Content}}
    </div>

    {{if .Page.Backlinks}}
    <aside class="lp-backlinks">
      <h2 class="lp-backlinks-title">Referenced from</h2>
      <ul class="lp-backlinks-list">
        {{range .Page.Backlinks}}
        <li><a class="lp-backlink" href="{{$.Site.BasePath}}{{.Permalink}}">{{.Title}}</a></li>
        {{end}}
      </ul>
    </aside>
    {{end}}
  </article>
</div>
{{end}}
`

const indexTemplate = `
{{define "title"}}{{.Title}} | {{.Site.Title}}{{end}}
{{define "currentSlug"}}{{end}}
{{define "seo"}}
  <meta name="description" content="{{if and .Site.Description (eq .CurrentPath "/")}}{{.Site.Description}}{{else}}{{.Title}} - {{.Site.Title}}{{end}}">
  {{if .Site.BaseURL}}<link rel="canonical" href="{{.Site.BaseURL}}{{.CurrentPath}}">{{end}}
  <meta property="og:title" content="{{.Title}}">
  <meta property="og:description" content="{{if and .Site.Description (eq .CurrentPath "/")}}{{.Site.Description}}{{else}}{{.Title}} - {{.Site.Title}}{{end}}">
  <meta property="og:type" content="website">
  <meta property="og:site_name" content="{{.Site.Title}}">
  {{if .Site.BaseURL}}<meta property="og:url" content="{{.Site.BaseURL}}{{.CurrentPath}}">{{end}}
  {{if .Site.Image}}<meta property="og:image" content="{{if .Site.BaseURL}}{{.Site.BaseURL}}{{end}}{{.Site.Image}}">{{end}}
  <meta name="twitter:card" content="summary">
  <meta name="twitter:title" content="{{.Title}}">
  <meta name="twitter:description" content="{{if and .Site.Description (eq .CurrentPath "/")}}{{.Site.Description}}{{else}}{{.Title}} - {{.Site.Title}}{{end}}">
{{end}}
{{define "content"}}
<div class="lp-section">
  <h1 class="lp-section-title">{{.Title}}</h1>
  {{if .ShowList}}<p class="lp-section-count">{{len .Pages}} items in {{.Title}}</p>{{end}}

  {{if .Intro}}
  <div class="lp-section-intro lp-content">
    {{.Intro}}
  </div>
  {{end}}

  {{if .ShowList}}
  <ul class="lp-index">
    {{range .Pages}}
    <li class="lp-index-item">
      <a class="lp-index-link" href="{{$.Site.BasePath}}{{.Permalink}}">
        {{if .Growth}}
        <span class="lp-index-growth lp-index-growth--{{.Growth}}">{{growthEmoji .Growth}}</span>
        {{end}}
        <span class="lp-index-title">{{.Title}}</span>
      </a>
      {{if .DisplayDate}}
      <time class="lp-index-date" datetime="{{.DisplayDateISO}}">{{.DisplayDate}}</time>
      {{end}}
    </li>
    {{end}}
  </ul>
  {{end}}
</div>
{{end}}
`

const tagIndexTemplate = `
{{define "title"}}Tags | {{.Site.Title}}{{end}}
{{define "currentSlug"}}tags{{end}}
{{define "seo"}}
  <meta name="description" content="Browse all tags - {{.Site.Title}}">
  {{if .Site.BaseURL}}<link rel="canonical" href="{{.Site.BaseURL}}/tags/">{{end}}
  <meta property="og:title" content="Tags">
  <meta property="og:description" content="Browse all tags - {{.Site.Title}}">
  <meta property="og:type" content="website">
  <meta property="og:site_name" content="{{.Site.Title}}">
  {{if .Site.BaseURL}}<meta property="og:url" content="{{.Site.BaseURL}}/tags/">{{end}}
  <meta name="twitter:card" content="summary">
  <meta name="twitter:title" content="Tags">
  <meta name="twitter:description" content="Browse all tags - {{.Site.Title}}">
{{end}}
{{define "content"}}
<div class="lp-section">
  <h1 class="lp-section-title">Tags</h1>

  <div class="lp-tag-cloud">
    {{range .Tags}}
    <a class="lp-tag-cloud-item" href="{{$.Site.BasePath}}/tags/{{.Name | lower}}/">
      #{{.Name}} <span class="lp-tag-count">({{.Count}})</span>
    </a>
    {{end}}
  </div>
</div>
{{end}}
`

const tagPageTemplate = `
{{define "title"}}#{{.Tag}} | {{.Site.Title}}{{end}}
{{define "currentSlug"}}tags/{{.Tag}}{{end}}
{{define "seo"}}
  <meta name="description" content="Pages tagged with #{{.Tag}} - {{.Site.Title}}">
  {{if .Site.BaseURL}}<link rel="canonical" href="{{.Site.BaseURL}}/tags/{{.Tag}}/">{{end}}
  <meta property="og:title" content="#{{.Tag}}">
  <meta property="og:description" content="Pages tagged with #{{.Tag}} - {{.Site.Title}}">
  <meta property="og:type" content="website">
  <meta property="og:site_name" content="{{.Site.Title}}">
  {{if .Site.BaseURL}}<meta property="og:url" content="{{.Site.BaseURL}}/tags/{{.Tag}}/">{{end}}
  <meta name="twitter:card" content="summary">
  <meta name="twitter:title" content="#{{.Tag}}">
  <meta name="twitter:description" content="Pages tagged with #{{.Tag}} - {{.Site.Title}}">
{{end}}
{{define "content"}}
<div class="lp-section">
  <h1 class="lp-section-title">#{{.Tag}}</h1>

  <ul class="lp-index">
    {{range .Pages}}
    <li class="lp-index-item">
      <a class="lp-index-link" href="{{$.Site.BasePath}}{{.Permalink}}">
        {{if .Growth}}
        <span class="lp-index-growth lp-index-growth--{{.Growth}}">{{growthEmoji .Growth}}</span>
        {{end}}
        <span class="lp-index-title">{{.Title}}</span>
      </a>
      {{if .DisplayDate}}
      <time class="lp-index-date" datetime="{{.DisplayDateISO}}">{{.DisplayDate}}</time>
      {{end}}
    </li>
    {{end}}
  </ul>
</div>
{{end}}
`

const notFoundTemplate = `
{{define "title"}}Page Not Found | {{.Site.Title}}{{end}}
{{define "currentSlug"}}{{end}}
{{define "seo"}}
  <meta name="robots" content="noindex">
{{end}}
{{define "content"}}
<div class="lp-not-found">
  <h1 class="lp-not-found-title">404</h1>
  <p class="lp-not-found-message">This page doesn't exist yet.</p>
  <a class="lp-not-found-link" href="{{.Site.BasePath}}/">Return home</a>
</div>
{{end}}
`
