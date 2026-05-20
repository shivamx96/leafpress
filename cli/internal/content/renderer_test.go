package content

import (
	"strings"
	"testing"
)

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
