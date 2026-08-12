package templates

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/theme"
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
	if toc[0].Text != "Video &amp; Audio" {
		t.Errorf("toc[0].Text = %q, want \"Video &amp; Audio\"", toc[0].Text)
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

func TestExtractTOC_EscapesDecodedAuthorHTML(t *testing.T) {
	input := `<h2>&lt;input autofocus onfocus=alert(1)&gt;</h2>`
	_, toc := ExtractTOC(input)

	if len(toc) != 1 {
		t.Fatalf("got %d TOC items, want 1", len(toc))
	}
	if toc[0].Text != "&lt;input autofocus onfocus=alert(1)&gt;" {
		t.Errorf("toc text resurrected author HTML: %q", toc[0].Text)
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

func TestSearchResultsDoNotInterpolateIndexHTML(t *testing.T) {
	if strings.Contains(baseTemplate, "results.innerHTML") {
		t.Fatal("search results must not interpolate index fields through innerHTML")
	}
	for _, want := range []string{
		"appendHighlightedText(title, item.title, q)",
		"appendHighlightedText(snippetEl, snippet, q)",
		"mark.textContent = text.substring",
	} {
		if !strings.Contains(baseTemplate, want) {
			t.Errorf("search template missing DOM-safe rendering fragment %q", want)
		}
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
	if url := remoteFontURL(testTheme("Bricolage Grotesque", "Inter", "JetBrains Mono")); url != "" {
		t.Errorf("bundled families produced remote URL %q", url)
	}
	if url := remoteFontURL(testTheme("Lobster", "Inter", "Roboto Mono")); url != "" {
		t.Errorf("unbundled families produced remote URL %q without opt-in", url)
	}
}

func TestRemoteFontURL_OptInCoversOnlyUnbundled(t *testing.T) {
	theme := testTheme("Bricolage Grotesque", "Inter", "Roboto Mono")
	theme.RemoteFonts = true
	url := remoteFontURL(theme)
	if strings.Contains(url, "Bricolage+Grotesque") || strings.Contains(url, "family=Inter") {
		t.Errorf("bundled families leaked into remote URL %q", url)
	}
	if !strings.Contains(url, "family=Roboto+Mono") {
		t.Error("should contain unbundled mono font")
	}
	if !strings.Contains(url, "display=swap") {
		t.Error("should include display=swap")
	}
	if strings.Count(url, "fonts.googleapis.com") != 1 {
		t.Error("should be a single combined URL")
	}

	// Fully bundled theme yields no URL even when opted in.
	allBundled := testTheme("Bricolage Grotesque", "Inter", "JetBrains Mono")
	allBundled.RemoteFonts = true
	if url := remoteFontURL(allBundled); url != "" {
		t.Errorf("fully bundled theme produced remote URL %q", url)
	}
}

func TestRemoteFontURL_Dedup(t *testing.T) {
	theme := testTheme("Lobster", "Lobster", "Roboto Mono")
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
	if got := UnhostedFamilies(testTheme("Bricolage Grotesque", "Inter", "JetBrains Mono")); len(got) != 0 {
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
	css := FontCSS(testTheme("Bricolage Grotesque", "Inter", "JetBrains Mono"))
	if !strings.Contains(css, `font-family: "Inter"`) || !strings.Contains(css, "@font-face") {
		t.Error("FontCSS missing bundled @font-face rules")
	}
}

func TestFontPreloadsPreserveRoleOrder(t *testing.T) {
	preloads := fontPreloads(testTheme("Bricolage Grotesque", "Inter", "JetBrains Mono"))
	want := []fontPreload{
		{Path: "static/leafpress/fonts/bricolage-grotesque-normal-latin.woff2", ContentType: "font/woff2"},
		{Path: "static/leafpress/fonts/inter-normal-latin.woff2", ContentType: "font/woff2"},
		{Path: "static/leafpress/fonts/jetbrains-mono-normal-latin.woff2", ContentType: "font/woff2"},
	}
	if !reflect.DeepEqual(preloads, want) {
		t.Fatalf("fontPreloads() = %#v, want %#v", preloads, want)
	}
}

func TestFontPreloadsDeduplicateAndSkipUnhosted(t *testing.T) {
	preloads := fontPreloads(testTheme("Inter", "Inter", "Lobster"))
	if len(preloads) != 1 || preloads[0].Path != "static/leafpress/fonts/inter-normal-latin.woff2" {
		t.Fatalf("fontPreloads() = %#v, want one Inter preload", preloads)
	}
}

func TestFontPreloadsCustomFace(t *testing.T) {
	theme := testTheme("My Serif", "Inter", "My Mono")
	theme.Fonts = []config.FontFace{
		{Family: "My Serif", File: "static/fonts/my-serif-italic.woff2", Style: "italic", Weight: "400"},
		{Family: "My Serif", File: "static/fonts/my-serif-bold.woff2", Weight: "700"},
		{Family: "My Serif", File: "static/fonts/my-serif-variable.woff2", Weight: "300 600"},
		{Family: "My Mono", File: "static/fonts/my-mono.ttf"},
	}
	preloads := fontPreloads(theme)
	want := []fontPreload{
		{Path: "static/fonts/my-serif-variable.woff2", ContentType: "font/woff2"},
		{Path: "static/leafpress/fonts/inter-normal-latin.woff2", ContentType: "font/woff2"},
		{Path: "static/fonts/my-mono.ttf", ContentType: "font/ttf"},
	}
	if !reflect.DeepEqual(preloads, want) {
		t.Fatalf("fontPreloads() = %#v, want %#v", preloads, want)
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
	theme := testTheme("Bricolage Grotesque", "My Serif", "JetBrains Mono")
	theme.Fonts = []config.FontFace{{Family: "My Serif", File: "static/fonts/my.woff2"}}
	css := FontCSS(theme)
	if !strings.Contains(css, `font-family: "My Serif"`) {
		t.Error("custom @font-face missing from FontCSS")
	}
	if !strings.Contains(css, `font-family: "Bricolage Grotesque"`) {
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

func TestMermaidUsesStrictSecurityLevel(t *testing.T) {
	if !strings.Contains(baseTemplate, "securityLevel: 'strict'") {
		t.Fatal("Mermaid initialization must explicitly use strict security")
	}
}

func TestClientScriptAssetIsContentAddressed(t *testing.T) {
	tmpl, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	site := SiteData{
		Title:    "Test Garden",
		BasePath: "/garden",
		Theme:    config.Default().Theme,
		Graph:    true,
		Search:   true,
	}
	assetPath, content, err := tmpl.ClientScriptAsset(site)
	if err != nil {
		t.Fatalf("ClientScriptAsset() error: %v", err)
	}

	hash := sha256.Sum256([]byte(content))
	wantPath := fmt.Sprintf("static/leafpress/app.%x.js", hash[:16])
	if assetPath != wantPath {
		t.Fatalf("ClientScriptAsset() path = %q, want %q", assetPath, wantPath)
	}
	for _, want := range []string{
		`var LP_BASE_PATH = '/garden'`,
		`themeMediaQuery.addEventListener('change', handleSystemThemeChange)`,
		`localStorage.removeItem('theme')`,
		`window.addEventListener('storage'`,
		"lp-graph-panel-body",
		"lp-search-input",
		"static/leafpress/mermaid/mermaid.min.js",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("client script missing %q", want)
		}
	}
	if strings.Contains(content, "<script") || strings.Contains(content, "</script>") {
		t.Fatal("client asset must contain JavaScript only, without script elements")
	}

	site.Search = false
	withoutSearchPath, withoutSearch, err := tmpl.ClientScriptAsset(site)
	if err != nil {
		t.Fatalf("ClientScriptAsset(search disabled) error: %v", err)
	}
	if withoutSearchPath == assetPath || withoutSearch == content {
		t.Fatal("feature-dependent client code must produce a different asset")
	}
	if strings.Contains(withoutSearch, "lp-search-input") {
		t.Fatal("search-disabled client asset contains search UI code")
	}
}

func TestClientScriptAssetLoadsFromHead(t *testing.T) {
	tmpl, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	var out bytes.Buffer
	err = tmpl.RenderIndex(&out, IndexData{
		Site: SiteData{
			Title:            "Test Garden",
			BasePath:         "/garden",
			Theme:            config.Default().Theme,
			ClientScriptPath: "static/leafpress/app.0123456789abcdef.js",
		},
		Title: "Home",
	})
	if err != nil {
		t.Fatalf("RenderIndex() error: %v", err)
	}

	html := out.String()
	bootstrap := `localStorage.getItem('theme')`
	script := `<script src="/garden/static/leafpress/app.0123456789abcdef.js" defer></script>`
	bootstrapAt := strings.Index(html, bootstrap)
	styleAt := strings.Index(html, `<style>`)
	scriptAt := strings.Index(html, script)
	headEnd := strings.Index(html, "</head>")
	if bootstrapAt < 0 || styleAt < 0 || bootstrapAt > styleAt {
		t.Fatalf("theme bootstrap must run before styles are loaded:\n%s", html)
	}
	if scriptAt < 0 || headEnd < 0 || scriptAt > headEnd {
		t.Fatalf("shared client script must be discovered in the document head:\n%s", html)
	}
	if strings.Contains(html, "var LP_BASE_PATH") {
		t.Fatal("page with a shared client asset must not duplicate inline client JavaScript")
	}

	out.Reset()
	err = tmpl.RenderIndex(&out, IndexData{
		Site:  SiteData{Title: "Test Garden", BasePath: "/garden", Theme: config.Default().Theme},
		Title: "Home",
	})
	if err != nil {
		t.Fatalf("RenderIndex(inline fallback) error: %v", err)
	}
	inline := out.String()
	if strings.Count(inline, "<script>") != 3 ||
		!strings.Contains(inline, "var LP_BASE_PATH") ||
		!strings.Contains(inline, "static/leafpress/mermaid/mermaid.min.js") {
		t.Fatal("inline fallback must preserve both legacy client script blocks")
	}
}

// An explicit accent replaces the preset accent in both modes and re-derives
// the contrast token: white for mid/dark accents (the historical behavior),
// dark text only for genuinely pale accents.
func TestThemePalettesAccentOverride(t *testing.T) {
	pair := themePalettes(config.Theme{Name: "dusk", Accent: "#e11d48"})
	if pair.Light.Accent != "#e11d48" || pair.Dark.Accent != "#e11d48" {
		t.Errorf("accent override not applied to both modes: %+v", pair)
	}
	if pair.Light.AccentContrast != "#ffffff" || pair.Dark.AccentContrast != "#ffffff" {
		t.Errorf("crimson accent should get white contrast text, got %q/%q", pair.Light.AccentContrast, pair.Dark.AccentContrast)
	}

	if pale := themePalettes(config.Theme{Accent: "#ffe066"}); pale.Light.AccentContrast != "#1a1a1a" {
		t.Errorf("pale accent should get dark contrast text, got %q", pale.Light.AccentContrast)
	}
	// The default green keeps its historical white pairing, including in the
	// 3-digit form.
	if green := themePalettes(config.Theme{Accent: "#5a0"}); green.Light.AccentContrast != "#ffffff" {
		t.Errorf("mid-tone accent should keep white contrast text, got %q", green.Light.AccentContrast)
	}

	dusk := theme.ByName("dusk")
	if pair := themePalettes(config.Theme{Name: "dusk"}); pair.Dark.AccentContrast != dusk.Dark.AccentContrast {
		t.Error("without an accent override the preset contrast pairing must be kept")
	}
}

// A theme preset resolves into the inline token block: light tokens in
// :root, dark tokens in [data-theme="dark"], with explicit config fields
// (accent, background) overriding the preset in both modes.
func TestThemePresetTokensInHead(t *testing.T) {
	tmpl, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	dusk := theme.ByName("dusk")
	render := func(themeCfg config.Theme) string {
		var out bytes.Buffer
		if err := tmpl.RenderIndex(&out, IndexData{
			Site:  SiteData{Title: "Test Garden", Theme: themeCfg},
			Title: "Home",
		}); err != nil {
			t.Fatalf("RenderIndex() error: %v", err)
		}
		return out.String()
	}

	html := render(config.Theme{
		Name:        "dusk",
		FontHeading: dusk.FontHeading, FontBody: dusk.FontBody, FontMono: dusk.FontMono,
		NavStyle: dusk.NavStyle, NavActiveStyle: dusk.NavActiveStyle,
	})
	for _, want := range []string{
		"--lp-bg: " + dusk.Light.Bg,
		"--lp-bg: " + dusk.Dark.Bg,
		"--lp-accent: " + dusk.Light.Accent,
		"--lp-accent: " + dusk.Dark.Accent,
		"--lp-accent-contrast: " + dusk.Light.AccentContrast,
		"--lp-text: " + dusk.Dark.Text,
		"--lp-graph-link: " + dusk.Light.GraphLink,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("dusk page missing token %q", want)
		}
	}

	overridden := render(config.Theme{
		Name:   "dusk",
		Accent: "#e11d48",
		Background: config.Background{
			Light: "#fafafa",
		},
	})
	if strings.Contains(overridden, dusk.Light.Accent) || strings.Contains(overridden, dusk.Dark.Accent) {
		t.Error("explicit accent must replace the preset accent in both modes")
	}
	if got := strings.Count(overridden, "--lp-accent: #e11d48"); got != 2 {
		t.Errorf("explicit accent should be emitted for light and dark, found %d occurrences", got)
	}
	if !strings.Contains(overridden, "--lp-bg: #fafafa") || !strings.Contains(overridden, "--lp-bg: "+dusk.Dark.Bg) {
		t.Error("explicit light background must win while dark keeps the preset value")
	}
}

func TestThemeBootstrapAndControlExposeSystemMode(t *testing.T) {
	tmpl, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	var out bytes.Buffer
	err = tmpl.RenderIndex(&out, IndexData{
		Site:  SiteData{Title: "Test Garden", Theme: config.Default().Theme},
		Title: "Home",
	})
	if err != nil {
		t.Fatalf("RenderIndex() error: %v", err)
	}

	html := out.String()
	for _, want := range []string{
		`<meta name="color-scheme" content="light dark">`,
		`data-theme-preference`,
		`preference = 'system'`,
		`lp-theme-icon-system`,
		`function getNextThemePreference()`,
		`applyThemePreference('system')`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered page missing system theme behavior %q", want)
		}
	}
}

func TestIndexIntroUsesSharedContentStyles(t *testing.T) {
	tmpl, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	var out bytes.Buffer
	err = tmpl.RenderIndex(&out, IndexData{
		Site: SiteData{
			Title: "Test Garden",
			Theme: config.Default().Theme,
		},
		Title: "Notes",
		Intro: `<pre class="chroma"><code>const answer = 42</code></pre>`,
	})
	if err != nil {
		t.Fatalf("RenderIndex() error: %v", err)
	}

	html := out.String()
	if !strings.Contains(html, `class="lp-section-intro lp-content"`) {
		t.Fatalf("index intro does not share the rich-content styling surface:\n%s", html)
	}
	if !strings.Contains(html, `<pre class="chroma">`) {
		t.Fatal("index intro dropped rendered code content")
	}
}
