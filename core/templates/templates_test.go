package templates

import (
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/config"
)

// --- Heading ID generation ---

func TestGenerateHeadingID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Getting Started", "getting-started"},
		{"with numbers", "Step 1 Setup", "step-1-setup"},
		{"special chars", "Video & Audio", "video-audio"},
		{"multiple spaces", "hello   world", "hello-world"},
		{"leading trailing spaces", "  hello  ", "hello"},
		{"with emoji", "Hello 🌱 World", "hello-world"},
		{"only emoji", "🌱🌿", ""},
		{"with punctuation", "What's New?", "what-s-new"},
		{"hyphens already", "my-page-title", "my-page-title"},
		{"mixed case", "CamelCase Title", "camelcase-title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateHeadingID(tt.input)
			if got != tt.want {
				t.Errorf("generateHeadingID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- TOC extraction ---

func TestExtractTOC_Basic(t *testing.T) {
	input := `<h2>Introduction</h2><p>text</p><h3>Details</h3><p>more</p>`
	html, toc := ExtractTOC(input)

	if len(toc) != 2 {
		t.Fatalf("got %d TOC items, want 2", len(toc))
	}

	if toc[0].Text != "Introduction" || toc[0].Level != 2 {
		t.Errorf("toc[0] = {%q, %d}, want {\"Introduction\", 2}", toc[0].Text, toc[0].Level)
	}
	if toc[1].Text != "Details" || toc[1].Level != 3 {
		t.Errorf("toc[1] = {%q, %d}, want {\"Details\", 3}", toc[1].Text, toc[1].Level)
	}

	// Headings should have IDs
	if !strings.Contains(html, `id="introduction"`) {
		t.Error("h2 should have id=\"introduction\"")
	}
	if !strings.Contains(html, `id="details"`) {
		t.Error("h3 should have id=\"details\"")
	}
}

func TestExtractTOC_AnchorLinks(t *testing.T) {
	input := `<h2>Section</h2>`
	html, _ := ExtractTOC(input)

	if !strings.Contains(html, `class="lp-heading-anchor"`) {
		t.Error("heading should contain anchor link")
	}
	if !strings.Contains(html, `href="#section"`) {
		t.Error("anchor should link to heading ID")
	}
}

func TestExtractTOC_DuplicateHeadings(t *testing.T) {
	input := `<h2>Setup</h2><p>text</p><h2>Setup</h2><p>text</p><h2>Setup</h2>`
	_, toc := ExtractTOC(input)

	if len(toc) != 3 {
		t.Fatalf("got %d TOC items, want 3", len(toc))
	}
	if toc[0].ID != "setup" {
		t.Errorf("toc[0].ID = %q, want \"setup\"", toc[0].ID)
	}
	if toc[1].ID != "setup-1" {
		t.Errorf("toc[1].ID = %q, want \"setup-1\"", toc[1].ID)
	}
	if toc[2].ID != "setup-2" {
		t.Errorf("toc[2].ID = %q, want \"setup-2\"", toc[2].ID)
	}
}

func TestExtractTOC_SpecialChars(t *testing.T) {
	input := `<h2>Video &amp; Audio</h2>`
	html, toc := ExtractTOC(input)

	if len(toc) != 1 {
		t.Fatalf("got %d TOC items, want 1", len(toc))
	}
	if toc[0].ID != "video-audio" {
		t.Errorf("toc[0].ID = %q, want \"video-audio\"", toc[0].ID)
	}
	if toc[0].Text != "Video & Audio" {
		t.Errorf("toc[0].Text = %q, want \"Video & Audio\"", toc[0].Text)
	}

	// Heading ID and TOC href must match
	if !strings.Contains(html, `id="video-audio"`) {
		t.Error("heading should have id=\"video-audio\"")
	}
}

func TestExtractTOC_ExistingID(t *testing.T) {
	input := `<h2 id="old-id">Section</h2>`
	html, toc := ExtractTOC(input)

	// Should replace goldmark's ID with ours
	if !strings.Contains(html, `id="section"`) {
		t.Error("should replace existing ID with generated one")
	}
	if strings.Contains(html, `id="old-id"`) {
		t.Error("old ID should be removed")
	}
	if toc[0].ID != "section" {
		t.Errorf("toc ID should be \"section\", got %q", toc[0].ID)
	}
}

func TestExtractTOC_HTMLInHeading(t *testing.T) {
	input := `<h2><code>func</code> main</h2>`
	_, toc := ExtractTOC(input)

	if len(toc) != 1 {
		t.Fatalf("got %d TOC items, want 1", len(toc))
	}
	if toc[0].Text != "func main" {
		t.Errorf("toc text should strip HTML tags, got %q", toc[0].Text)
	}
}

func TestExtractTOC_NoHeadings(t *testing.T) {
	input := `<p>Just a paragraph.</p>`
	html, toc := ExtractTOC(input)

	if len(toc) != 0 {
		t.Errorf("got %d TOC items, want 0", len(toc))
	}
	if html != input {
		t.Error("HTML should be unchanged when no headings")
	}
}

// --- Fonts ---

func testTheme(heading, body, mono string) config.Theme {
	theme := config.Default().Theme
	theme.FontHeading = heading
	theme.FontBody = body
	theme.FontMono = mono
	return theme
}

func TestRemoteFontURL_OffByDefault(t *testing.T) {
	// Without the deprecated remoteFonts opt-in there is never a remote
	// URL, bundled or not.
	if url := remoteFontURL(testTheme("Crimson Pro", "Inter", "JetBrains Mono")); url != "" {
		t.Errorf("bundled families produced remote URL %q", url)
	}
	if url := remoteFontURL(testTheme("Lobster", "Inter", "Fira Code")); url != "" {
		t.Errorf("unbundled families produced remote URL %q without opt-in", url)
	}
}

func TestRemoteFontURL_OptInCoversOnlyUnbundled(t *testing.T) {
	theme := testTheme("Crimson Pro", "Inter", "Fira Code")
	theme.RemoteFonts = true
	url := remoteFontURL(theme)
	if strings.Contains(url, "Crimson+Pro") || strings.Contains(url, "family=Inter") {
		t.Errorf("bundled families leaked into remote URL %q", url)
	}
	if !strings.Contains(url, "family=Fira+Code") {
		t.Error("should contain unbundled mono font")
	}
	if !strings.Contains(url, "display=swap") {
		t.Error("should include display=swap")
	}
	if strings.Count(url, "fonts.googleapis.com") != 1 {
		t.Error("should be a single combined URL")
	}

	// Fully bundled theme yields no URL even when opted in.
	allBundled := testTheme("Crimson Pro", "Inter", "JetBrains Mono")
	allBundled.RemoteFonts = true
	if url := remoteFontURL(allBundled); url != "" {
		t.Errorf("fully bundled theme produced remote URL %q", url)
	}
}

func TestRemoteFontURL_Dedup(t *testing.T) {
	theme := testTheme("Lobster", "Lobster", "Fira Code")
	theme.RemoteFonts = true
	url := remoteFontURL(theme)
	if strings.Count(url, "Lobster") != 1 {
		t.Error("duplicate fonts should be deduplicated")
	}
}

func TestUnhostedFamilies(t *testing.T) {
	got := UnhostedFamilies(testTheme("Lobster", "Lobster", "JetBrains Mono"))
	if len(got) != 1 || got[0] != "Lobster" {
		t.Fatalf("UnhostedFamilies = %v, want [Lobster]", got)
	}
	if got := UnhostedFamilies(testTheme("Crimson Pro", "Inter", "JetBrains Mono")); len(got) != 0 {
		t.Fatalf("fully bundled theme reported unhosted families: %v", got)
	}
}

func TestUnhostedFamiliesExcludesDeclaredCustom(t *testing.T) {
	theme := testTheme("Lobster", "My Serif", "JetBrains Mono")
	theme.Fonts = []config.FontFace{{Family: "My Serif", File: "static/fonts/my.woff2"}}
	got := UnhostedFamilies(theme)
	if len(got) != 1 || got[0] != "Lobster" {
		t.Fatalf("UnhostedFamilies = %v, want [Lobster] (declared custom family is hosted)", got)
	}
}

func TestFontCSSIncludesBundledFamilies(t *testing.T) {
	css := FontCSS(testTheme("Crimson Pro", "Inter", "JetBrains Mono"))
	if !strings.Contains(css, `font-family: "Inter"`) || !strings.Contains(css, "@font-face") {
		t.Error("FontCSS missing bundled @font-face rules")
	}
}

func TestCustomFontCSS(t *testing.T) {
	css := customFontCSS([]config.FontFace{
		{Family: "My Serif", File: "static/fonts/my.woff2", Weight: "400 700", Style: "italic", Display: "optional"},
		{Family: "Old", File: "static/fonts/old.ttf"},
	})
	if !strings.Contains(css, `font-family: "My Serif"`) {
		t.Error("missing family")
	}
	// Stylesheet-relative URL: resolves against the style.css location, so
	// no base path is baked in.
	if !strings.Contains(css, `src: url("static/fonts/my.woff2") format("woff2")`) {
		t.Errorf("missing site-relative src:\n%s", css)
	}
	if !strings.Contains(css, "font-weight: 400 700") || !strings.Contains(css, "font-style: italic") || !strings.Contains(css, "font-display: optional") {
		t.Errorf("explicit fields not honored:\n%s", css)
	}
	// Defaults for the second face.
	if !strings.Contains(css, "font-weight: 400;") || !strings.Contains(css, "font-style: normal") || !strings.Contains(css, "font-display: swap") {
		t.Errorf("defaults not applied:\n%s", css)
	}
	if !strings.Contains(css, `format("truetype")`) {
		t.Errorf("ttf format mapping missing:\n%s", css)
	}
}

func TestFontCSSCombinesBuiltinAndCustom(t *testing.T) {
	theme := testTheme("Crimson Pro", "My Serif", "JetBrains Mono")
	theme.Fonts = []config.FontFace{{Family: "My Serif", File: "static/fonts/my.woff2"}}
	css := FontCSS(theme)
	if !strings.Contains(css, `font-family: "My Serif"`) {
		t.Error("custom @font-face missing from FontCSS")
	}
	if !strings.Contains(css, `font-family: "Crimson Pro"`) {
		t.Error("bundled @font-face missing from FontCSS")
	}
}

func TestRemoteFontURL_ExcludesDeclaredCustomFamilies(t *testing.T) {
	theme := testTheme("Lobster", "My Serif", "JetBrains Mono")
	theme.Fonts = []config.FontFace{{Family: "My Serif", File: "static/fonts/my.woff2"}}
	theme.RemoteFonts = true
	url := remoteFontURL(theme)
	if !strings.Contains(url, "family=Lobster") || strings.Contains(url, "My+Serif") {
		t.Errorf("declared custom family must not appear in remote URL: %q", url)
	}
}
