package content

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

type failingMarkdown struct {
	goldmark.Markdown
}

func (failingMarkdown) Convert([]byte, io.Writer, ...parser.ParseOption) error {
	return errors.New("forced conversion failure")
}

// --- Media type detection ---

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"video.mp4", true},
		{"video.webm", true},
		{"video.ogv", true},
		{"video.mov", true},
		{"video.MP4", true},
		{"image.png", false},
		{"audio.mp3", false},
		{"file.txt", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isVideoFile(tt.name); got != tt.want {
				t.Errorf("isVideoFile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsAudioFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"audio.mp3", true},
		{"audio.wav", true},
		{"audio.ogg", true},
		{"audio.m4a", true},
		{"audio.flac", true},
		{"audio.FLAC", true},
		{"image.png", false},
		{"video.mp4", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAudioFile(tt.name); got != tt.want {
				t.Errorf("isAudioFile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// --- Media path resolution ---

func TestResolveMediaSrc(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		basePath string
		want     string
	}{
		{"bare image filename", "photo.png", "", "/static/images/photo.png"},
		{"bare video filename", "clip.mp4", "", "/static/video/clip.mp4"},
		{"bare audio filename", "track.mp3", "", "/static/audio/track.mp3"},
		{"path with slash", "static/video.mp4", "", "/static/video.mp4"},
		{"nested path", "images/sub/photo.png", "", "/images/sub/photo.png"},
		{"spaces encoded", "my photo.png", "", "/static/images/my%20photo.png"},
		{"with basePath", "photo.png", "/blog", "/blog/static/images/photo.png"},
		{"video with basePath", "clip.mp4", "/blog", "/blog/static/video/clip.mp4"},
		{"audio with basePath", "track.mp3", "/blog", "/blog/static/audio/track.mp3"},
		{"path with basePath", "static/video.mp4", "/blog", "/blog/static/video.mp4"},
		{"leading slash stripped", "/static/video.mp4", "", "/static/video.mp4"},
		{"dot-slash cleaned", "./photo.png", "", "/static/images/photo.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveMediaSrc(tt.filename, tt.basePath)
			if got != tt.want {
				t.Errorf("resolveMediaSrc(%q, %q) = %q, want %q", tt.filename, tt.basePath, got, tt.want)
			}
		})
	}
}

// --- Obsidian embed rendering ---

func TestRender_ImageEmbed(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("![[photo.png]]")
	if !strings.Contains(html, "/static/images/photo.png") {
		t.Error("image embed should resolve to /static/images/photo.png")
	}
}

func TestRender_ImageEmbedWithAlt(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("![[photo.png|my caption]]")
	if !strings.Contains(html, "my caption") {
		t.Error("image embed should use pipe value as alt text")
	}
}

func TestRender_ImageEmbedWithWidth(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("![[photo.png|500]]")
	if !strings.Contains(html, `width="500"`) {
		t.Error("numeric pipe value should set image width")
	}
}

func TestRender_VideoEmbed(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("![[demo.mp4]]")
	if !strings.Contains(html, "<video") {
		t.Error("mp4 embed should render as <video>")
	}
	if !strings.Contains(html, "playsinline") {
		t.Error("video should have playsinline attribute")
	}
	if !strings.Contains(html, `/static/video/demo.mp4`) {
		t.Error("video embed should resolve to /static/video/demo.mp4")
	}
}

func TestRender_AudioEmbed(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("![[recording.mp3]]")
	if !strings.Contains(html, "<audio") {
		t.Error("mp3 embed should render as <audio>")
	}
	if !strings.Contains(html, `/static/audio/recording.mp3`) {
		t.Error("audio embed should resolve to /static/audio/recording.mp3")
	}
}

func TestRender_PathAwareEmbed(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("![[static/demo.mp4]]")
	if strings.Contains(html, "/static/images/static/") {
		t.Error("path embed should not double-prefix with /static/images/")
	}
	if !strings.Contains(html, "/static/demo.mp4") {
		t.Error("path embed should resolve to /static/demo.mp4")
	}
}

// --- YouTube auto-embed ---

func TestRender_YouTubeFullURL(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if !strings.Contains(html, "youtube-nocookie.com/embed/dQw4w9WgXcQ") {
		t.Error("YouTube full URL should auto-embed as iframe")
	}
	if !strings.Contains(html, `class="lp-video"`) {
		t.Error("YouTube embed should have lp-video wrapper")
	}
}

func TestRender_YouTubeShortURL(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("https://youtu.be/dQw4w9WgXcQ")
	if !strings.Contains(html, "youtube-nocookie.com/embed/dQw4w9WgXcQ") {
		t.Error("YouTube short URL should auto-embed as iframe")
	}
}

func TestRender_YouTubeInlineNotEmbedded(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("Check this out: https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if strings.Contains(html, "lp-video") {
		t.Error("inline YouTube link should not be auto-embedded")
	}
}

// --- Mermaid ---

func TestRender_MermaidBlock(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("```mermaid\ngraph TD\n    A --> B\n```")
	if !strings.Contains(html, `class="mermaid"`) {
		t.Error("mermaid code block should be converted to div.mermaid")
	}
	if strings.Contains(html, "language-mermaid") {
		t.Error("mermaid should not remain as a code block")
	}
}

func TestRender_MermaidContentPreserved(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("```mermaid\ngraph TD\n    A-->B\n```")
	if !strings.Contains(html, "A-->B") {
		t.Error("mermaid diagram source should be preserved with unescaped arrows")
	}
}

// --- Code block protection ---

func TestRender_WikilinkInInlineCode(t *testing.T) {
	r := NewRenderer(nil, true, "")
	html, _ := r.Render("Use `[[wiki-links]]` syntax")
	if strings.Contains(html, "lp-wikilink") || strings.Contains(html, "lp-broken-link") {
		t.Error("wiki-link inside inline code should not be processed")
	}
	if !strings.Contains(html, "<code>") {
		t.Error("inline code should render as <code>")
	}
}

func TestRender_WikilinkInFencedCode(t *testing.T) {
	r := NewRenderer(nil, true, "")
	html, _ := r.Render("```\n[[some-link]]\n```")
	if strings.Contains(html, "lp-wikilink") || strings.Contains(html, "lp-broken-link") {
		t.Error("wiki-link inside fenced code should not be processed")
	}
}

func TestRender_WikilinkInCodeWithFencedBlockOnSamePage(t *testing.T) {
	r := NewRenderer(nil, true, "")
	input := "```markdown\n[[other]]\n```\n\nUse `[[wiki-links]]` syntax"
	html, _ := r.Render(input)
	if strings.Contains(html, "lp-broken-link") && strings.Contains(html, "<code>") {
		// Check the inline code specifically
		if strings.Contains(html, `<code><span class="lp-broken-link"`) || strings.Contains(html, `<code>&lt;span`) {
			t.Error("inline code wiki-link should not be processed even with fenced blocks on same page")
		}
	}
}

// Regression: a 4-backtick fence containing a nested 3-backtick fence (used in
// docs to show code-fence examples) must not desync code protection for the
// inline wiki-link that follows. The old fixed-length fence regex mis-parsed
// this, leaking the inline `[[...]]` into wiki-link processing.
func TestRender_WikilinkInInlineCodeAfterNestedFence(t *testing.T) {
	r := NewRenderer(NewLinkResolver(nil), true, "")
	input := "````markdown\n```mermaid\ngraph TD\n```\n````\n\nLink: `[[projects/website]]`"
	html, warnings := r.Render(input)
	if strings.Contains(html, "lp-wikilink") || strings.Contains(html, "lp-broken-link") {
		t.Errorf("wiki-link inside inline code after a nested fence should not be processed:\n%s", html)
	}
	if !strings.Contains(html, "<code>[[projects/website]]</code>") {
		t.Errorf("inline code wiki-link should render literally, got:\n%s", html)
	}
	for _, w := range warnings {
		if strings.Contains(w, "projects/website") {
			t.Errorf("inline-code wiki-link should not warn as a broken link, got: %q", w)
		}
	}
}

// Regression: a wiki-link inside a 4-backtick fenced block must be preserved.
func TestRender_WikilinkInQuadFencedCode(t *testing.T) {
	r := NewRenderer(NewLinkResolver(nil), true, "")
	html, warnings := r.Render("````\n[[inside-quad]]\n````")
	if strings.Contains(html, "lp-wikilink") || strings.Contains(html, "lp-broken-link") {
		t.Errorf("wiki-link inside a 4-backtick fence should not be processed:\n%s", html)
	}
	if len(warnings) != 0 {
		t.Errorf("no warnings expected for wiki-link inside a fence, got %v", warnings)
	}
}

// --- Broken wikilinks ---

func TestRender_BrokenWikilinkDefault(t *testing.T) {
	r := NewRenderer(NewLinkResolver(nil), true, "")
	html, warnings := r.Render("See [[missing-page]] here")
	if !strings.Contains(html, `<span class="lp-broken-link">missing-page</span>`) {
		t.Errorf("broken wiki-link should render as lp-broken-link span by default, got: %s", html)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "broken link: [[missing-page]]") {
		t.Errorf("broken wiki-link should produce a warning, got: %v", warnings)
	}
}

func TestRender_BrokenWikilinkPlainMode(t *testing.T) {
	r := NewRenderer(NewLinkResolver(nil), true, "")
	r.SetPlainBrokenLinks(true)
	html, warnings := r.Render("See [[missing-page]] here")
	if strings.Contains(html, "lp-broken-link") || strings.Contains(html, "<a ") {
		t.Errorf("plain mode should render broken wiki-link without anchor or class, got: %s", html)
	}
	if !strings.Contains(html, "missing-page") {
		t.Errorf("plain mode should render the display text, got: %s", html)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "broken link: [[missing-page]]") {
		t.Errorf("plain mode should still produce a warning, got: %v", warnings)
	}
}

func TestRender_BrokenWikilinkPlainModeUsesLabel(t *testing.T) {
	r := NewRenderer(NewLinkResolver(nil), true, "")
	r.SetPlainBrokenLinks(true)
	html, _ := r.Render("See [[missing-page|Custom Label]] here")
	if !strings.Contains(html, "Custom Label") {
		t.Errorf("plain mode should render the label of a broken wiki-link, got: %s", html)
	}
	if strings.Contains(html, "lp-broken-link") || strings.Contains(html, "[[") {
		t.Errorf("plain mode should strip wiki-link syntax entirely, got: %s", html)
	}
}

// --- Footnotes ---

func TestRender_Footnotes(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("Text with footnote[^1].\n\n[^1]: The footnote content.")
	if !strings.Contains(html, "footnote-ref") {
		t.Error("footnote reference should render with footnote-ref class")
	}
	if !strings.Contains(html, "footnote-backref") {
		t.Error("footnote section should have backref links")
	}
}

func TestRender_FootnoteMarkdown(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("See[^1].\n\n[^1]: Has **bold** and `code`.")
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Error("footnote content should support bold")
	}
	if !strings.Contains(html, "<code>code</code>") {
		t.Error("footnote content should support inline code")
	}
}

// --- Raw HTML escaping (SetEscapeRawHTML) ---

// escapingRenderer returns a renderer with raw-HTML escaping enabled.
func escapingRenderer(resolver *LinkResolver, enableWikilinks bool, basePath string) *Renderer {
	r := NewRenderer(resolver, enableWikilinks, basePath)
	r.SetEscapeRawHTML(true)
	return r
}

func TestRender_RawHTMLEscaping(t *testing.T) {
	tests := []struct {
		name        string
		markdown    string
		wantDefault string // must appear in default (pass-through) output
		wantEscaped string // must appear in escaped output
		rawFragment string // must NOT appear in escaped output
	}{
		{
			name:        "inline raw HTML",
			markdown:    "before <script>alert(1)</script> after",
			wantDefault: "<script>alert(1)</script>",
			wantEscaped: "&lt;script&gt;alert(1)&lt;/script&gt;",
			rawFragment: "<script>",
		},
		{
			name:        "block-level raw HTML",
			markdown:    "<div class=\"evil\">\nowned\n</div>\n\nparagraph",
			wantDefault: `<div class="evil">`,
			wantEscaped: "&lt;div class=&quot;evil&quot;&gt;",
			rawFragment: `<div class="evil">`,
		},
		{
			name:        "mixed paragraph",
			markdown:    "text with <img src=x onerror=alert(1)> embedded **bold**",
			wantDefault: "<img src=x onerror=alert(1)",
			wantEscaped: "&lt;img src=x onerror=alert(1)&gt;",
			rawFragment: "<img src=x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultOut, _ := NewRenderer(nil, false, "").Render(tt.markdown)
			if !strings.Contains(defaultOut, tt.wantDefault) {
				t.Errorf("default mode should pass raw HTML through, want %q in:\n%s", tt.wantDefault, defaultOut)
			}

			escapedOut, _ := escapingRenderer(nil, false, "").Render(tt.markdown)
			if !strings.Contains(escapedOut, tt.wantEscaped) {
				t.Errorf("escape mode should render raw HTML as escaped text, want %q in:\n%s", tt.wantEscaped, escapedOut)
			}
			if strings.Contains(escapedOut, tt.rawFragment) {
				t.Errorf("escape mode leaked raw HTML %q in:\n%s", tt.rawFragment, escapedOut)
			}
			if strings.Contains(escapedOut, "raw HTML omitted") {
				t.Errorf("escape mode should show escaped text, not goldmark's omission comment:\n%s", escapedOut)
			}
		})
	}
}

func TestRender_ConversionErrorEscapesHostedFallback(t *testing.T) {
	r := escapingRenderer(nil, false, "")
	r.mdEscaped = failingMarkdown{Markdown: r.mdEscaped}
	got, warnings := r.Render(`<img src=x onerror=alert(1)>`)
	if got != `&lt;img src=x onerror=alert(1)&gt;` {
		t.Fatalf("fallback = %q, want escaped source", got)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "forced conversion failure") {
		t.Fatalf("warnings = %v", warnings)
	}
}

func TestProcessMermaidBlocksUnescapesApostrophes(t *testing.T) {
	input := `<pre><code class="language-mermaid">A[&#39;quoted&#39;]</code></pre>`
	if got := processMermaidBlocks(input, true); got != `<div class="mermaid">A['quoted']</div>` {
		t.Fatalf("trusted Mermaid = %q", got)
	}
	if got := processMermaidBlocks(input, false); strings.Contains(got, "A['quoted']") {
		t.Fatalf("hosted Mermaid unexpectedly unescaped entities: %q", got)
	}
}

func TestRender_EscapeModeCodeBlocksUnchanged(t *testing.T) {
	// Fenced and inline code containing HTML must render exactly as before
	// in both modes: no double-escaping via the protect/restore pipeline,
	// no interference from the raw-HTML escaper.
	markdown := "Intro with `<b>inline</b>` code.\n\n```html\n<script>alert(1)</script> & <div>\n```\n\ndone"

	defaultOut, _ := NewRenderer(nil, false, "").Render(markdown)
	escapedOut, _ := escapingRenderer(nil, false, "").Render(markdown)

	if defaultOut != escapedOut {
		t.Errorf("code blocks should render identically in both modes:\ndefault:\n%s\nescaped:\n%s", defaultOut, escapedOut)
	}
	if strings.Contains(escapedOut, "&amp;lt;") || strings.Contains(escapedOut, "&amp;gt;") {
		t.Errorf("code block content was double-escaped:\n%s", escapedOut)
	}
	if !strings.Contains(escapedOut, "&lt;b&gt;inline&lt;/b&gt;") {
		t.Errorf("inline code should render singly-escaped, got:\n%s", escapedOut)
	}
}

func TestRender_EscapeModeCalloutStillLive(t *testing.T) {
	r := escapingRenderer(nil, false, "")
	html, _ := r.Render("> [!note] My Title\n> Callout body with <script>x</script> inside.")

	if !strings.Contains(html, `<div class="lp-callout lp-callout-note">`) {
		t.Errorf("callout wrapper should stay live HTML in escape mode, got:\n%s", html)
	}
	if !strings.Contains(html, "My Title") {
		t.Errorf("callout title missing, got:\n%s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") || strings.Contains(html, "<script>") {
		t.Errorf("raw HTML inside callout body should be escaped, got:\n%s", html)
	}
}

func TestRender_EscapeModeCalloutTitleEscaped(t *testing.T) {
	r := escapingRenderer(nil, false, "")
	html, _ := r.Render("> [!note] <img src=x onerror=alert(1)>\n> body")

	if strings.Contains(html, "<img") {
		t.Errorf("callout title HTML should be escaped in escape mode, got:\n%s", html)
	}
	if !strings.Contains(html, "&lt;img src=x onerror=alert(1)&gt;") {
		t.Errorf("callout title should render as escaped text, got:\n%s", html)
	}
}

func TestRender_EscapeModeWikilinkStillAnchor(t *testing.T) {
	pages := []*Page{{Title: "Beta", Slug: "beta", Permalink: "/beta/"}}
	r := escapingRenderer(NewLinkResolver(pages), true, "/g/s")
	html, _ := r.Render("Linking [[beta|the beta note]] here.")

	if !strings.Contains(html, `<a class="lp-wikilink" href="/g/s/beta/">the beta note</a>`) {
		t.Errorf("wikilink anchor should stay live HTML in escape mode, got:\n%s", html)
	}
}

func TestRender_EscapeModeWikilinkLabelEscaped(t *testing.T) {
	pages := []*Page{{Title: "Beta", Slug: "beta", Permalink: "/beta/"}}
	r := escapingRenderer(NewLinkResolver(pages), true, "")
	html, _ := r.Render("Linking [[beta|</a><script>alert(1)</script>]] here.")

	if strings.Contains(html, "<script>") {
		t.Errorf("wikilink label HTML should be escaped in escape mode, got:\n%s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("wikilink label should render as escaped text, got:\n%s", html)
	}
}

func TestRender_EscapeModeMediaEmbedsStillLive(t *testing.T) {
	r := escapingRenderer(nil, false, "")

	html, _ := r.Render("![[demo.mp4]]")
	if !strings.Contains(html, `<div class="lp-video-local">`) {
		t.Errorf("video embed should stay live HTML in escape mode, got:\n%s", html)
	}

	html, _ = r.Render("![[photo.png|300]]")
	if !strings.Contains(html, `<img src="/static/images/photo.png" alt="photo.png" width="300"`) {
		t.Errorf("sized image embed should stay live HTML in escape mode, got:\n%s", html)
	}
}

func TestRender_EscapeModeMermaidStaysEscaped(t *testing.T) {
	markdown := "```mermaid\ngraph TD\nA[\"<script>alert(1)</script>\"] --> B\n```"

	defaultOut, _ := NewRenderer(nil, false, "").Render(markdown)
	if !strings.Contains(defaultOut, `<div class="mermaid">`) {
		t.Fatalf("mermaid block should convert to a div, got:\n%s", defaultOut)
	}

	escapedOut, _ := escapingRenderer(nil, false, "").Render(markdown)
	if !strings.Contains(escapedOut, `<div class="mermaid">`) {
		t.Fatalf("mermaid block should convert to a div in escape mode, got:\n%s", escapedOut)
	}
	if strings.Contains(escapedOut, "<script>") {
		t.Errorf("mermaid content must not be unescaped into live HTML in escape mode, got:\n%s", escapedOut)
	}
}

func TestRender_EscapeModeDefaultIsOff(t *testing.T) {
	r := NewRenderer(nil, false, "")
	html, _ := r.Render("keep <b>me</b> raw")
	if !strings.Contains(html, "<b>me</b>") {
		t.Errorf("raw HTML should pass through by default, got:\n%s", html)
	}

	r.SetEscapeRawHTML(true)
	html, _ = r.Render("keep <b>me</b> raw")
	if !strings.Contains(html, "&lt;b&gt;me&lt;/b&gt;") {
		t.Errorf("raw HTML should escape after SetEscapeRawHTML(true), got:\n%s", html)
	}

	r.SetEscapeRawHTML(false)
	html, _ = r.Render("keep <b>me</b> raw")
	if !strings.Contains(html, "<b>me</b>") {
		t.Errorf("raw HTML should pass through again after SetEscapeRawHTML(false), got:\n%s", html)
	}
}
