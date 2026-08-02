package content

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Renderer converts markdown to HTML
type Renderer struct {
	md               goldmark.Markdown
	mdEscaped        goldmark.Markdown // Lazily built variant that escapes raw HTML (see SetEscapeRawHTML)
	resolver         *LinkResolver
	enableWikilinks  bool
	basePath         string // Base path for links (e.g., "/repo-name" for GitHub Pages)
	plainBrokenLinks bool   // Render unresolved wikilinks as plain text instead of a styled span
	escapeRawHTML    bool   // Render raw HTML in markdown as visibly escaped text instead of passing it through
}

// Buffer pool for markdown rendering (reduces allocations)
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// newGoldmark builds the goldmark instance shared by both render modes.
// The default mode uses html.WithUnsafe() (raw HTML passes through, as
// always). The escaping mode instead registers rawHTMLEscaper for the raw
// HTML node kinds — and, since WithUnsafe is dropped, goldmark also filters
// dangerous link/image URLs (javascript:, data:, ...) in that mode.
func newGoldmark(escapeRawHTML bool) goldmark.Markdown {
	rendererOpts := []renderer.Option{
		html.WithHardWraps(),
		html.WithXHTML(),
	}
	if escapeRawHTML {
		rendererOpts = append(rendererOpts,
			renderer.WithNodeRenderers(util.Prioritized(&rawHTMLEscaper{}, 100)))
	} else {
		rendererOpts = append(rendererOpts, html.WithUnsafe()) // Allow raw HTML in markdown
	}

	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown
			extension.Typographer,
			extension.NewFootnote(),
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
		goldmark.WithRendererOptions(rendererOpts...),
	)
}

// NewRenderer creates a new markdown renderer
func NewRenderer(resolver *LinkResolver, enableWikilinks bool, basePath string) *Renderer {
	return &Renderer{
		md:              newGoldmark(false),
		resolver:        resolver,
		enableWikilinks: enableWikilinks,
		basePath:        basePath,
	}
}

// SetPlainBrokenLinks controls how unresolved wikilinks are rendered.
// By default (false), a broken wikilink renders as a styled span
// (<span class="lp-broken-link">…</span>). When enabled, broken wikilinks
// render as plain display text instead — no anchor, no class.
func (r *Renderer) SetPlainBrokenLinks(plain bool) {
	r.plainBrokenLinks = plain
}

// SetEscapeRawHTML controls how raw HTML in markdown is rendered.
// By default (false), raw HTML passes through unchanged (goldmark's unsafe
// mode) — appropriate for trusted single-author content. When enabled, raw
// HTML renders as visibly escaped text (e.g. <script> becomes &lt;script&gt;
// in the output, so the reader sees the literal characters the author
// typed), and renderer-generated HTML (wikilinks, callouts, media embeds)
// is still emitted live. Call this before Render; it is not safe to call
// concurrently with Render.
func (r *Renderer) SetEscapeRawHTML(escape bool) {
	r.escapeRawHTML = escape
	if escape && r.mdEscaped == nil {
		r.mdEscaped = newGoldmark(true)
	}
}

// Render converts markdown to HTML, processing wiki-links
func (r *Renderer) Render(content string) (string, []string) {
	var warnings []string

	// Extract fenced code blocks first, then inline code from the remainder
	fencedBlocks := findFencedBlocks(content)
	protected := content
	for i, block := range fencedBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		protected = strings.Replace(protected, block, placeholder, 1)
	}
	// Now extract inline code from content with fenced blocks already removed
	inlineBlocks := inlineCodeRegex.FindAllString(protected, -1)
	codeBlocks := append(fencedBlocks, inlineBlocks...)
	for i, block := range inlineBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", len(fencedBlocks)+i)
		protected = strings.Replace(protected, block, placeholder, 1)
	}

	// In escape mode, renderer-generated HTML is swapped for placeholder
	// tokens so the raw-HTML escaper only sees author-typed HTML.
	var trusted *trustedChunks
	if r.escapeRawHTML {
		trusted = newTrustedChunks()
	}

	// Pre-markdown processing (all on protected content)
	processed := r.processObsidianImagesProtected(protected, trusted)
	processed = r.processCalloutsProtected(processed, trusted)
	if r.enableWikilinks {
		processed = r.processWikiLinksProtected(processed, &warnings, trusted)
	}

	// Restore code blocks ONCE before markdown conversion
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		processed = strings.Replace(processed, placeholder, block, 1)
	}

	// Get buffer from pool (reduces allocations)
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Render markdown to HTML
	md := r.md
	if r.escapeRawHTML {
		md = r.mdEscaped
	}
	if err := md.Convert([]byte(processed), buf); err != nil {
		warnings = append(warnings, "markdown conversion error: "+err.Error())
		if r.escapeRawHTML {
			return stdhtml.EscapeString(content), warnings
		}
		return content, warnings
	}

	rendered := buf.String()
	if trusted != nil {
		rendered = trusted.restore(rendered)
	}

	// Post-markdown processing (single pass)
	html := r.processPostMarkdown(rendered)

	return html, warnings
}

// Pre-compiled regexes (compiled once at startup)
var (
	obsidianImageRegex = regexp.MustCompile(`!\[\[([^\]|]+?)(?:\|([^\]]+))?\]\]`)
	inlineCodeRegex    = regexp.MustCompile("`[^`]+`")
	externalLinkRegex  = regexp.MustCompile(`<a\s+href="(https?://[^"]+)"([^>]*)>([^<]+)</a>`)
	// YouTube URL patterns for auto-embed
	youtubeRegex = regexp.MustCompile(`<p>\s*<a[^>]+href="https?://(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/)([\w-]+)[^"]*"[^>]*>[^<]*</a>\s*</p>`)
	// Mermaid code block detection (Chroma-highlighted or plain)
	mermaidRegex = regexp.MustCompile(`(?s)<pre[^>]*><code[^>]*class="[^"]*language-mermaid[^"]*"[^>]*>(.*?)</code></pre>|<div class="highlight"><pre[^>]*class="chroma"><code class="language-mermaid"[^>]*>(.*?)</code></pre></div>`)
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

// processCalloutsProtected converts Obsidian-style callouts (assumes code blocks already protected).
// When trusted is non-nil (escape mode), the generated wrapper HTML is emitted
// as trusted placeholder tokens and the custom title is HTML-escaped.
func (r *Renderer) processCalloutsProtected(content string, trusted *trustedChunks) string {
	lines := strings.Split(content, "\n")
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

			// Build the callout HTML (calloutType is regex-restricted to \w+;
			// the title is author text, so escape it in escape mode)
			if trusted != nil {
				title = stdhtml.EscapeString(title)
			}
			calloutContent := strings.Join(contentLines, "\n")
			open := fmt.Sprintf(
				"<div class=\"lp-callout lp-callout-%s\">\n<div class=\"lp-callout-title\"><span class=\"lp-callout-icon\">%s</span> %s</div>\n<div class=\"lp-callout-content\">",
				calloutType,
				info.icon,
				title,
			)
			const closing = "</div>\n</div>"
			var calloutHTML string
			if trusted != nil {
				calloutHTML = trusted.wrap(open) + "\n\n" + calloutContent + "\n\n" + trusted.wrap(closing)
			} else {
				calloutHTML = open + "\n\n" + calloutContent + "\n\n" + closing
			}
			result = append(result, calloutHTML)
		} else {
			result = append(result, line)
			i++
		}
	}

	return strings.Join(result, "\n")
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

	processed := r.processCalloutsProtected(protectedContent, nil)

	// Restore code blocks
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		processed = strings.Replace(processed, placeholder, block, 1)
	}

	return processed
}

// isVideoFile checks if a filename has a video extension
func isVideoFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".mp4", ".webm", ".ogv", ".mov":
		return true
	}
	return false
}

// isAudioFile checks if a filename has an audio extension
func isAudioFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".mp3", ".wav", ".ogg", ".m4a", ".flac":
		return true
	}
	return false
}

// resolveMediaSrc resolves an Obsidian embed filename to a URL path.
// Bare filenames get routed to a media-specific default under /static/.
// Paths (e.g. "static/video.mp4") get a leading slash only.
func resolveMediaSrc(filename, basePath string) string {
	cleaned := filepath.ToSlash(filepath.Clean(filename))
	cleaned = strings.TrimPrefix(cleaned, "/")

	var src string
	if strings.Contains(cleaned, "/") {
		src = "/" + cleaned
	} else {
		switch {
		case isVideoFile(cleaned):
			src = "/static/video/" + cleaned
		case isAudioFile(cleaned):
			src = "/static/audio/" + cleaned
		default:
			src = "/static/images/" + cleaned
		}
	}

	// Encode path segments (preserve /)
	parts := strings.Split(src, "/")
	for i, p := range parts {
		parts[i] = strings.ReplaceAll(p, " ", "%20")
	}
	src = strings.Join(parts, "/")

	if basePath != "" && basePath != "/" {
		src = basePath + src
	}
	return src
}

// processObsidianImagesProtected converts Obsidian image/video/audio embeds (assumes code blocks already protected).
// When trusted is non-nil (escape mode), the generated embed HTML is emitted
// as trusted placeholder tokens with author-controlled values escaped.
func (r *Renderer) processObsidianImagesProtected(content string, trusted *trustedChunks) string {
	return obsidianImageRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := obsidianImageRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		filename := strings.TrimSpace(submatches[1])
		alt := filepath.Base(filename) // Use just the filename for alt text
		width := 0

		if len(submatches) > 2 && submatches[2] != "" {
			pipeValue := strings.TrimSpace(submatches[2])
			if w, err := strconv.Atoi(pipeValue); err == nil {
				width = w
			} else {
				alt = pipeValue
			}
		}

		src := resolveMediaSrc(filename, r.basePath)

		// The HTML-emitting branches below interpolate author-controlled
		// values into trusted HTML that bypasses the raw-HTML escaper, so in
		// escape mode escape those values and wrap the result in a trusted
		// placeholder. The markdown-image fallback is left untouched —
		// goldmark escapes it itself.
		emitHTML := func(format, src, alt string, args ...interface{}) string {
			if trusted == nil {
				return fmt.Sprintf(format, append([]interface{}{src, alt}, args...)...)
			}
			src = stdhtml.EscapeString(src)
			alt = stdhtml.EscapeString(alt)
			return trusted.wrap(fmt.Sprintf(format, append([]interface{}{src, alt}, args...)...))
		}

		if isVideoFile(filename) {
			return emitHTML(`<div class="lp-video-local"><video controls playsinline preload="metadata"><source src="%s">%s</video></div>`, src, alt)
		}
		if isAudioFile(filename) {
			return emitHTML(`<audio controls preload="metadata"><source src="%s">%s</audio>`, src, alt)
		}
		if width > 0 {
			return emitHTML(`<img src="%s" alt="%s" width="%d">`, src, alt, width)
		}
		return fmt.Sprintf("![%s](%s)", alt, src)
	})
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

	result := r.processObsidianImagesProtected(protectedContent, nil)

	// Restore code blocks
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		result = strings.Replace(result, placeholder, block, 1)
	}

	return result
}

// processWikiLinksProtected replaces [[links]] with HTML anchors (assumes code blocks already protected).
// When trusted is non-nil (escape mode), the generated open/close tags are
// emitted as trusted placeholder tokens; the label stays inline so it is
// still processed (and escaped) as markdown, exactly like the default path.
func (r *Renderer) processWikiLinksProtected(content string, warnings *[]string, trusted *trustedChunks) string {
	links := ExtractWikiLinks(content)

	// tag wraps generated HTML tags as trusted placeholders in escape mode.
	tag := func(html string) string {
		if trusted != nil {
			return trusted.wrap(html)
		}
		return html
	}

	result := content
	for _, link := range links {
		var replacement string

		if r.resolver != nil {
			resolved := r.resolver.Resolve(link.Target)

			if resolved.Broken {
				if r.plainBrokenLinks {
					// Plain mode - render just the display text
					replacement = link.Label
				} else {
					// Broken link - render as span with class
					replacement = tag(`<span class="lp-broken-link">`) + link.Label + tag(`</span>`)
				}
				*warnings = append(*warnings, "broken link: [["+link.Target+"]]")
			} else {
				// Valid link
				if resolved.Ambiguous {
					*warnings = append(*warnings, "ambiguous link: [["+link.Target+"]]")
				}
				href := r.basePath + resolved.Page.Permalink
				if trusted != nil {
					// href bypasses the raw-HTML escaper via the trusted
					// placeholder, so make it attribute-safe here.
					href = stdhtml.EscapeString(href)
				}
				replacement = tag(`<a class="lp-wikilink" href="`+href+`">`) + link.Label + tag(`</a>`)
			}
		} else {
			// No resolver - just render the label
			replacement = link.Label
		}

		result = replaceFirst(result, link.Raw, replacement)
	}

	return result
}

func (r *Renderer) processWikiLinks(content string, warnings *[]string) string {
	// Extract code blocks and inline code to protect them
	codeBlocks := extractCodeBlocks(content)
	protectedContent := content

	// Replace code blocks with placeholders
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		protectedContent = strings.Replace(protectedContent, block, placeholder, 1)
	}

	result := r.processWikiLinksProtected(protectedContent, warnings, nil)

	// Restore code blocks
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("___CODE_BLOCK_%d___", i)
		result = strings.Replace(result, placeholder, block, 1)
	}

	return result
}

// extractCodeBlocks extracts code blocks and inline code from markdown
// findFencedBlocks returns the fenced code blocks in content, in order,
// following CommonMark fence rules: an opening fence of three or more backticks
// or tildes is closed by a line with at least as many of the same character.
// This correctly spans 4+ backtick fences and nested fences (used to document
// code fences), which a fixed-length regex cannot match — a misparse there
// would leak protection and let wiki-links inside later inline code resolve.
func findFencedBlocks(content string) []string {
	lines := strings.Split(content, "\n")
	var blocks []string
	for i := 0; i < len(lines); {
		char, n, ok := fenceOpen(lines[i])
		if !ok {
			i++
			continue
		}
		start := i
		i++
		for i < len(lines) && !fenceClose(lines[i], char, n) {
			i++
		}
		if i < len(lines) {
			i++ // include the closing fence line
		}
		blocks = append(blocks, strings.Join(lines[start:i], "\n"))
	}
	return blocks
}

// fenceOpen reports whether line opens a fenced code block, returning the fence
// character ('`' or '~') and its run length. Up to three leading spaces are
// allowed; a backtick fence's info string may not contain a backtick.
func fenceOpen(line string) (byte, int, bool) {
	s := strings.TrimLeft(line, " ")
	if len(line)-len(s) > 3 || len(s) < 3 {
		return 0, 0, false
	}
	c := s[0]
	if c != '`' && c != '~' {
		return 0, 0, false
	}
	n := 0
	for n < len(s) && s[n] == c {
		n++
	}
	if n < 3 {
		return 0, 0, false
	}
	if c == '`' && strings.ContainsRune(s[n:], '`') {
		return 0, 0, false
	}
	return c, n, true
}

// fenceClose reports whether line closes a fence of char c opened with n
// characters: at least n of the same character, then only spaces.
func fenceClose(line string, c byte, n int) bool {
	s := strings.TrimLeft(line, " ")
	if len(line)-len(s) > 3 {
		return false
	}
	m := 0
	for m < len(s) && s[m] == c {
		m++
	}
	return m >= n && strings.TrimRight(s[m:], " ") == ""
}

func extractCodeBlocks(content string) []string {
	var blocks []string

	// Extract fenced code blocks (```...```)
	blocks = append(blocks, findFencedBlocks(content)...)

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

// processPostMarkdown combines all post-markdown HTML processing in one function
func (r *Renderer) processPostMarkdown(html string) string {
	// Convert mermaid code blocks to divs for client-side rendering.
	// In escape mode the diagram source stays HTML-escaped inside the div
	// (unescaping it would reintroduce live author HTML).
	result := processMermaidBlocks(html, !r.escapeRawHTML)
	// Embed YouTube links before processing external links
	result = processYouTubeEmbeds(result)
	// Process external links
	result = r.processExternalLinks(result)
	// Add lazy loading to images
	result = processLazyImages(result)
	// Convert blockquote citations
	result = processBlockquoteCitations(result)
	return result
}

// processMermaidBlocks converts mermaid code blocks into divs for client-side rendering
func processMermaidBlocks(html string, unescapeEntities bool) string {
	return mermaidRegex.ReplaceAllStringFunc(html, func(match string) string {
		submatches := mermaidRegex.FindStringSubmatch(match)
		// Content is in group 1 or group 2 depending on which pattern matched
		content := submatches[1]
		if content == "" {
			content = submatches[2]
		}
		if unescapeEntities {
			// Unescape HTML entities back to raw text for Mermaid.js
			content = strings.ReplaceAll(content, "&amp;", "&")
			content = strings.ReplaceAll(content, "&lt;", "<")
			content = strings.ReplaceAll(content, "&gt;", ">")
			content = strings.ReplaceAll(content, "&#34;", `"`)
			content = strings.ReplaceAll(content, "&quot;", `"`)
			content = strings.ReplaceAll(content, "&#39;", `'`)
			content = strings.ReplaceAll(content, "&#x27;", `'`)
			content = strings.ReplaceAll(content, "&apos;", `'`)
		}
		return `<div class="mermaid">` + content + `</div>`
	})
}

// processYouTubeEmbeds converts standalone YouTube links into responsive iframes
func processYouTubeEmbeds(html string) string {
	return youtubeRegex.ReplaceAllStringFunc(html, func(match string) string {
		submatches := youtubeRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		videoID := submatches[1]
		return `<div class="lp-video"><iframe src="https://www.youtube-nocookie.com/embed/` + videoID + `" frameborder="0" allowfullscreen allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"></iframe></div>`
	})
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
