package content

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Renderer converts markdown to HTML
type Renderer struct {
	md       goldmark.Markdown
	resolver *LinkResolver
}

// NewRenderer creates a new markdown renderer
func NewRenderer(resolver *LinkResolver) *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // Allow raw HTML in markdown
		),
	)

	return &Renderer{
		md:       md,
		resolver: resolver,
	}
}

// Render converts markdown to HTML, processing wiki-links
func (r *Renderer) Render(content string) (string, []string) {
	var warnings []string

	// First, replace wiki-links with HTML anchors
	processed := r.processWikiLinks(content, &warnings)

	// Then render markdown to HTML
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(processed), &buf); err != nil {
		warnings = append(warnings, "markdown conversion error: "+err.Error())
		return content, warnings
	}

	// Process external links
	html := r.processExternalLinks(buf.String())

	return html, warnings
}

// processWikiLinks replaces [[links]] with HTML anchors
func (r *Renderer) processWikiLinks(content string, warnings *[]string) string {
	links := ExtractWikiLinks(content)

	result := content
	for _, link := range links {
		var replacement string

		if r.resolver != nil {
			resolved := r.resolver.Resolve(link.Target)

			if resolved.Broken {
				// Broken link - render as span with class
				replacement = `<span class="lp-broken-link">` + link.Label + `</span>`
				*warnings = append(*warnings, "broken link: [["+link.Target+"]]")
			} else {
				// Valid link
				if resolved.Ambiguous {
					*warnings = append(*warnings, "ambiguous link: [["+link.Target+"]]")
				}
				replacement = `<a class="lp-wikilink" href="` + resolved.Page.Permalink + `">` + link.Label + `</a>`
			}
		} else {
			// No resolver - just render the label
			replacement = link.Label
		}

		result = replaceFirst(result, link.Raw, replacement)
	}

	return result
}

// replaceFirst replaces only the first occurrence
func replaceFirst(s, old, new string) string {
	i := indexOf(s, old)
	if i < 0 {
		return s
	}
	return s[:i] + new + s[i+len(old):]
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// externalLinkRegex matches http/https links
var externalLinkRegex = regexp.MustCompile(`<a\s+href="(https?://[^"]+)"([^>]*)>([^<]+)</a>`)

// processExternalLinks adds target="_blank" and class to external links
func (r *Renderer) processExternalLinks(html string) string {
	return externalLinkRegex.ReplaceAllStringFunc(html, func(match string) string {
		submatches := externalLinkRegex.FindStringSubmatch(match)
		if len(submatches) < 4 {
			return match
		}

		href := submatches[1]
		attrs := submatches[2]
		text := submatches[3]

		return `<a class="lp-external" href="` + href + `" target="_blank" rel="noopener"` + attrs + `>` + text + ` ↗</a>`
	})
}

// RenderPages renders HTML content for all pages
func RenderPages(pages []*Page) []string {
	resolver := NewLinkResolver(pages)
	renderer := NewRenderer(resolver)

	var allWarnings []string
	for _, page := range pages {
		html, warnings := renderer.Render(page.RawContent)
		page.HTMLContent = html
		allWarnings = append(allWarnings, warnings...)
	}

	return allWarnings
}
