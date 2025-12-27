package content

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Renderer converts markdown to HTML
type Renderer struct {
	md              goldmark.Markdown
	resolver        *LinkResolver
	enableWikilinks bool
}

// NewRenderer creates a new markdown renderer
func NewRenderer(resolver *LinkResolver, enableWikilinks bool) *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown
			extension.Typographer,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithLineNumbers(false),
				),
			),
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
		md:              md,
		resolver:        resolver,
		enableWikilinks: enableWikilinks,
	}
}

// Render converts markdown to HTML, processing wiki-links
func (r *Renderer) Render(content string) (string, []string) {
	var warnings []string

	// First, process Obsidian image embeds (![[image.png]])
	processed := r.processObsidianImages(content)

	// Then, replace wiki-links with HTML anchors (if enabled)
	if r.enableWikilinks {
		processed = r.processWikiLinks(processed, &warnings)
	}

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

// Pre-compiled regexes (compiled once at startup)
var (
	obsidianImageRegex = regexp.MustCompile(`!\[\[([^\]|]+?)(?:\|([^\]]+))?\]\]`)
	codeBlockRegex     = regexp.MustCompile("(?s)```[^`]*```")
	inlineCodeRegex    = regexp.MustCompile("`[^`]+`")
	externalLinkRegex  = regexp.MustCompile(`<a\s+href="(https?://[^"]+)"([^>]*)>([^<]+)</a>`)
)

// processObsidianImages converts Obsidian image embeds to standard markdown
func (r *Renderer) processObsidianImages(content string) string {
	// Extract code blocks to protect them
	codeBlocks := extractCodeBlocks(content)
	protectedContent := content

	// Replace code blocks with placeholders
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		protectedContent = strings.Replace(protectedContent, block, placeholder, 1)
	}

	// Replace ![[image.png]] with ![image.png](/static/images/image.png)
	// Replace ![[image.png|alt]] with ![alt](/static/images/image.png)
	result := obsidianImageRegex.ReplaceAllStringFunc(protectedContent, func(match string) string {
		submatches := obsidianImageRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		filename := strings.TrimSpace(submatches[1])
		alt := filename
		if len(submatches) > 2 && submatches[2] != "" {
			alt = strings.TrimSpace(submatches[2])
		}

		// URL-encode spaces in filename
		encodedFilename := strings.ReplaceAll(filename, " ", "%20")

		return fmt.Sprintf("![%s](/static/images/%s)", alt, encodedFilename)
	})

	// Restore code blocks
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		result = strings.Replace(result, placeholder, block, 1)
	}

	return result
}

// processWikiLinks replaces [[links]] with HTML anchors
func (r *Renderer) processWikiLinks(content string, warnings *[]string) string {
	// Extract code blocks and inline code to protect them
	codeBlocks := extractCodeBlocks(content)
	protectedContent := content

	// Replace code blocks with placeholders
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		protectedContent = strings.Replace(protectedContent, block, placeholder, 1)
	}

	links := ExtractWikiLinks(protectedContent)

	result := protectedContent
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

	// Restore code blocks
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		result = strings.Replace(result, placeholder, block, 1)
	}

	return result
}

// extractCodeBlocks extracts code blocks and inline code from markdown
func extractCodeBlocks(content string) []string {
	var blocks []string

	// Extract fenced code blocks (```...```)
	blocks = append(blocks, codeBlockRegex.FindAllString(content, -1)...)

	// Extract inline code (`...`)
	blocks = append(blocks, inlineCodeRegex.FindAllString(content, -1)...)

	return blocks
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

// RenderPages renders HTML content for all pages in parallel
// If resolver is nil, a new one will be created
func RenderPages(pages []*Page, enableWikilinks bool, resolver *LinkResolver) []string {
	if len(pages) == 0 {
		return nil
	}

	if resolver == nil {
		resolver = NewLinkResolver(pages)
	}
	renderer := NewRenderer(resolver, enableWikilinks)

	numWorkers := runtime.NumCPU()
	if numWorkers > len(pages) {
		numWorkers = len(pages)
	}

	pageChan := make(chan *Page, len(pages))
	var wg sync.WaitGroup
	var warningsMu sync.Mutex
	var allWarnings []string

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for page := range pageChan {
				html, warnings := renderer.Render(page.RawContent)
				page.HTMLContent = html
				if len(warnings) > 0 {
					warningsMu.Lock()
					allWarnings = append(allWarnings, warnings...)
					warningsMu.Unlock()
				}
			}
		}()
	}

	// Send pages to workers
	for _, page := range pages {
		pageChan <- page
	}
	close(pageChan)

	// Wait for all workers
	wg.Wait()

	return allWarnings
}
