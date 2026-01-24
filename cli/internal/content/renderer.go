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
	basePath        string // Base path for links (e.g., "/repo-name" for GitHub Pages)
}

// Buffer pool for markdown rendering (reduces allocations)
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// NewRenderer creates a new markdown renderer
func NewRenderer(resolver *LinkResolver, enableWikilinks bool, basePath string) *Renderer {
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
		basePath:        basePath,
	}
}

// Render converts markdown to HTML, processing wiki-links
func (r *Renderer) Render(content string) (string, []string) {
	var warnings []string

	// First, process Obsidian image embeds (![[image.png]])
	processed := r.processObsidianImages(content)

	// Process callouts before markdown conversion
	processed = r.processCallouts(processed)

	// Then, replace wiki-links with HTML anchors (if enabled)
	if r.enableWikilinks {
		processed = r.processWikiLinks(processed, &warnings)
	}

	// Get buffer from pool (reduces allocations)
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Render markdown to HTML
	if err := r.md.Convert([]byte(processed), buf); err != nil {
		warnings = append(warnings, "markdown conversion error: "+err.Error())
		return content, warnings
	}

	// Process external links
	html := r.processExternalLinks(buf.String())

	// Add lazy loading to images
	html = processLazyImages(html)

	// Convert blockquote citations (- Author) to <cite> elements
	html = processBlockquoteCitations(html)

	return html, warnings
}

// Pre-compiled regexes (compiled once at startup)
var (
	obsidianImageRegex = regexp.MustCompile(`!\[\[([^\]|]+?)(?:\|([^\]]+))?\]\]`)
	codeBlockRegex     = regexp.MustCompile("(?s)```[^`]*```")
	inlineCodeRegex    = regexp.MustCompile("`[^`]+`")
	externalLinkRegex  = regexp.MustCompile(`<a\s+href="(https?://[^"]+)"([^>]*)>([^<]+)</a>`)
	// Callout regex: matches > [!type] or > [!type] title followed by content lines
	calloutStartRegex = regexp.MustCompile(`(?m)^>\s*\[!(\w+)\](?:\s+(.*))?$`)
	// Image regex for lazy loading (captures attributes, handles self-closing)
	imgTagFullRegex = regexp.MustCompile(`<img\s+([^>]*?)\s*/?\s*>`)
	// Blockquote citation regex: matches <p>- Author</p> or <p>— Author</p> at end of blockquote
	blockquoteCiteRegex = regexp.MustCompile(`(?s)(<blockquote>\s*(?:<p>.*?</p>\s*)*)<p>\s*[-–—]\s*(.+?)\s*</p>\s*(</blockquote>)`)
	// Blockquote citation from list: matches single-item <ul><li>Author</li></ul> at end of blockquote
	// This handles "> - Author" which markdown parses as a list
	blockquoteCiteListRegex = regexp.MustCompile(`(?s)(<blockquote>\s*(?:<p>.*?</p>\s*)*)<ul>\s*<li>(.+?)</li>\s*</ul>\s*(</blockquote>)`)
)

// calloutTypes maps callout type to display title and icon
var calloutTypes = map[string]struct {
	title string
	icon  string
}{
	"note":      {"Note", "📝"},
	"tip":       {"Tip", "💡"},
	"hint":      {"Hint", "💡"},
	"important": {"Important", "❗"},
	"warning":   {"Warning", "⚠️"},
	"caution":   {"Caution", "⚠️"},
	"danger":    {"Danger", "🔴"},
	"error":     {"Error", "🔴"},
	"info":      {"Info", "ℹ️"},
	"todo":      {"Todo", "☑️"},
	"example":   {"Example", "📋"},
	"quote":     {"Quote", "💬"},
	"question":  {"Question", "❓"},
	"faq":       {"FAQ", "❓"},
	"success":   {"Success", "✅"},
	"check":     {"Check", "✅"},
	"done":      {"Done", "✅"},
	"fail":      {"Fail", "❌"},
	"failure":   {"Failure", "❌"},
	"bug":       {"Bug", "🐛"},
	"abstract":  {"Abstract", "📄"},
	"summary":   {"Summary", "📄"},
	"tldr":      {"TL;DR", "📄"},
}

// processCallouts converts Obsidian-style callouts to HTML
// Input: > [!note] Optional title
//
//	> Content here
//
// Output: <div class="lp-callout lp-callout-note">...</div>
func (r *Renderer) processCallouts(content string) string {
	// Extract code blocks to protect them
	codeBlocks := extractCodeBlocks(content)
	protectedContent := content

	// Replace code blocks with placeholders
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		protectedContent = strings.Replace(protectedContent, block, placeholder, 1)
	}

	lines := strings.Split(protectedContent, "\n")
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Check if this line starts a callout
		matches := calloutStartRegex.FindStringSubmatch(line)
		if matches != nil {
			calloutType := strings.ToLower(matches[1])
			customTitle := ""
			if len(matches) > 2 {
				customTitle = strings.TrimSpace(matches[2])
			}

			// Get callout info or use defaults
			info, ok := calloutTypes[calloutType]
			if !ok {
				info = struct {
					title string
					icon  string
				}{strings.Title(calloutType), "📌"}
			}

			// Use custom title if provided
			title := info.title
			if customTitle != "" {
				title = customTitle
			}

			// Collect all content lines (lines starting with >)
			var contentLines []string
			i++
			for i < len(lines) {
				if strings.HasPrefix(lines[i], ">") {
					// Check if this line starts a new callout
					if calloutStartRegex.MatchString(lines[i]) {
						break
					}
					// Remove the leading > and optional space
					contentLine := strings.TrimPrefix(lines[i], ">")
					contentLine = strings.TrimPrefix(contentLine, " ")
					contentLines = append(contentLines, contentLine)
					i++
				} else if strings.TrimSpace(lines[i]) == "" {
					// Empty line might continue the callout if next line has > but is not a new callout
					if i+1 < len(lines) && strings.HasPrefix(lines[i+1], ">") && !calloutStartRegex.MatchString(lines[i+1]) {
						contentLines = append(contentLines, "")
						i++
					} else {
						break
					}
				} else {
					break
				}
			}

			// Build the callout HTML
			calloutContent := strings.Join(contentLines, "\n")
			calloutHTML := fmt.Sprintf(
				"<div class=\"lp-callout lp-callout-%s\">\n<div class=\"lp-callout-title\"><span class=\"lp-callout-icon\">%s</span> %s</div>\n<div class=\"lp-callout-content\">\n\n%s\n\n</div>\n</div>",
				calloutType,
				info.icon,
				title,
				calloutContent,
			)
			result = append(result, calloutHTML)
		} else {
			result = append(result, line)
			i++
		}
	}

	processed := strings.Join(result, "\n")

	// Restore code blocks
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		processed = strings.Replace(processed, placeholder, block, 1)
	}

	return processed
}

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
				replacement = `<a class="lp-wikilink" href="` + r.basePath + resolved.Page.Permalink + `">` + link.Label + `</a>`
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

// processLazyImages adds lazy loading attributes to all images
func processLazyImages(html string) string {
	return imgTagFullRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Don't add if already has loading attribute
		if strings.Contains(match, "loading=") {
			return match
		}
		// Insert loading="lazy" decoding="async" before the closing >
		attrs := imgTagFullRegex.FindStringSubmatch(match)
		if len(attrs) < 2 {
			return match
		}
		return `<img ` + attrs[1] + ` loading="lazy" decoding="async">`
	})
}

// processBlockquoteCitations converts blockquote paragraphs starting with - or — to <cite>
// Input:  <blockquote><p>Quote text</p><p>- Author Name</p></blockquote>
// Output: <blockquote><p>Quote text</p><cite>Author Name</cite></blockquote>
// Also handles: <blockquote><p>Quote text</p><ul><li>Author</li></ul></blockquote>
// (which is what "> - Author" produces in markdown)
func processBlockquoteCitations(html string) string {
	// First, handle explicit dash/em-dash in paragraph
	result := blockquoteCiteRegex.ReplaceAllString(html, `$1<cite>$2</cite>$3`)
	// Then, handle single-item list (from "> - Author" syntax)
	result = blockquoteCiteListRegex.ReplaceAllString(result, `$1<cite>$2</cite>$3`)
	return result
}

// RenderPages renders HTML content for all pages in parallel
// If resolver is nil, a new one will be created
func RenderPages(pages []*Page, enableWikilinks bool, resolver *LinkResolver, basePath string) []string {
	if len(pages) == 0 {
		return nil
	}

	if resolver == nil {
		resolver = NewLinkResolver(pages)
	}
	renderer := NewRenderer(resolver, enableWikilinks, basePath)

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

				// Calculate reading time
				page.WordCount = CountWords(html)
				page.ImageCount = CountImages(html)
				if page.ReadingTimeOverride != nil {
					page.ReadingTime = *page.ReadingTimeOverride
				} else {
					page.ReadingTime = CalculateReadingTime(page.WordCount, page.ImageCount)
				}

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
