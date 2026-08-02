package build

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/assets"
	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
)

// newTestProject creates a minimal project in a temp dir and chdirs into it
// (Builder resolves the project root from the working directory).
func newTestProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("# Note\n\nhello\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	return dir
}

func TestBuildWritesDefaultFaviconsFromRegistry(t *testing.T) {
	dir := newTestProject(t)

	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Literal public paths, independent of registry field plumbing: the base
	// template links {BasePath}/favicon.*, so these exact root locations are
	// the historical URL contract.
	for name, logicalPath := range map[string]string{
		"favicon.ico":       assets.BuiltinFaviconICO,
		"favicon.svg":       assets.BuiltinFaviconSVG,
		"favicon-96x96.png": assets.BuiltinFaviconPNG,
	} {
		builtin, ok := assets.BuiltinByLogicalPath(logicalPath)
		if !ok {
			t.Fatalf("registry missing %s", logicalPath)
		}
		got, err := os.ReadFile(filepath.Join(dir, "_site", name))
		if err != nil {
			t.Fatalf("favicon %s not written at site root: %v", name, err)
		}
		if !bytes.Equal(got, builtin.Content()) {
			t.Errorf("favicon %s does not match registry content", name)
		}
		if _, err := os.Stat(filepath.Join(dir, "_site", "static", "leafpress", name)); !os.IsNotExist(err) {
			t.Errorf("favicon %s must not be materialized under static/leafpress", name)
		}
	}
}

func TestBuildPrefersUserFavicons(t *testing.T) {
	dir := newTestProject(t)
	custom := []byte("user icon bytes")
	if err := os.WriteFile(filepath.Join(dir, "favicon.ico"), custom, 0644); err != nil {
		t.Fatal(err)
	}

	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "_site", "favicon.ico"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, custom) {
		t.Error("user favicon.ico was not preferred over the built-in")
	}

	// The other favicons still fall back to the registry.
	svg, err := os.ReadFile(filepath.Join(dir, "_site", "favicon.svg"))
	if err != nil {
		t.Fatal(err)
	}
	builtin, _ := assets.BuiltinByLogicalPath(assets.BuiltinFaviconSVG)
	if !bytes.Equal(svg, builtin.Content()) {
		t.Error("favicon.svg does not match registry content")
	}
}

func TestBuildMaterializesBuiltinFonts(t *testing.T) {
	dir := newTestProject(t)

	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Every face used by the default theme must be on disk at its logical path
	// with registry content. Retained built-ins that are not selected are not
	// materialized.
	defaultFamilies := map[string]bool{
		config.Default().Theme.FontHeading: true,
		config.Default().Theme.FontBody:    true,
		config.Default().Theme.FontMono:    true,
	}
	for _, face := range assets.BuiltinFontFaces() {
		if !defaultFamilies[face.Family] {
			continue
		}
		builtin, _ := assets.BuiltinByLogicalPath(face.LogicalPath)
		got, err := os.ReadFile(filepath.Join(dir, "_site", filepath.FromSlash(face.LogicalPath)))
		if err != nil {
			t.Fatalf("font %s not materialized: %v", face.LogicalPath, err)
		}
		if !bytes.Equal(got, builtin.Content()) {
			t.Errorf("font %s does not match registry content", face.LogicalPath)
		}
	}

	// Each used family's OFL license text is exported alongside the fonts.
	for _, family := range []string{"Bricolage Grotesque", "Inter", "JetBrains Mono"} {
		licensePath, ok := assets.BuiltinFontLicense(family)
		if !ok {
			t.Fatalf("no license asset for %s", family)
		}
		if _, err := os.Stat(filepath.Join(dir, "_site", filepath.FromSlash(licensePath))); err != nil {
			t.Errorf("license %s not materialized: %v", licensePath, err)
		}
	}

	// @font-face lives in the shared stylesheet, not in every page head.
	css, err := os.ReadFile(filepath.Join(dir, "_site", "style.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(css, []byte("@font-face")) {
		t.Error("style.css missing @font-face")
	}
	if !bytes.Contains(css, []byte(`url("static/leafpress/fonts/inter-normal-latin.woff2")`)) {
		t.Error("style.css font URLs must be site-relative")
	}
	page, err := os.ReadFile(filepath.Join(dir, "_site", "note", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(page, []byte("@font-face")) {
		t.Error("@font-face must not be inlined into page heads")
	}
	if bytes.Contains(page, []byte("fonts.googleapis.com")) {
		t.Error("default build must not reference Google Fonts")
	}
}

func TestBuildUnbundledFamiliesWarnAndStayLocal(t *testing.T) {
	dir := newTestProject(t)

	cfg := config.Default()
	cfg.Theme.FontHeading = "Lobster"
	cfg.Theme.FontBody = "Lobster"
	cfg.Theme.FontMono = "Fira Code"
	b := New(cfg, Options{})
	stats, err := b.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "_site", "static", "leafpress", "fonts")); !os.IsNotExist(err) {
		t.Error("no built-in fonts should be materialized for unbundled families")
	}
	page, err := os.ReadFile(filepath.Join(dir, "_site", "note", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	// Self-contained by default: warning + system-stack fallback, no remote.
	if bytes.Contains(page, []byte("fonts.googleapis.com")) {
		t.Error("unbundled families must not load remotely without the opt-in")
	}
	if stats.WarningCount < 2 {
		t.Errorf("expected fallback warnings for Lobster and Fira Code, got %d", stats.WarningCount)
	}
}

func TestBuildRemoteFontsOptIn(t *testing.T) {
	dir := newTestProject(t)

	cfg := config.Default()
	cfg.Theme.FontBody = "Lobster"
	cfg.Theme.RemoteFonts = true
	b := New(cfg, Options{})
	if _, err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}
	page, err := os.ReadFile(filepath.Join(dir, "_site", "note", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(page, []byte("fonts.googleapis.com/css2?family=Lobster")) {
		t.Error("deprecated remoteFonts opt-in should keep the Google Fonts link")
	}
	// Bundled heading/mono stay self-hosted even under the opt-in: never in
	// the remote URL, still present as @font-face with files on disk.
	if bytes.Contains(page, []byte("family=Bricolage+Grotesque")) || bytes.Contains(page, []byte("family=JetBrains+Mono")) {
		t.Error("bundled families leaked into the remote font URL")
	}
	css, err := os.ReadFile(filepath.Join(dir, "_site", "style.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(css, []byte(`font-family: "Bricolage Grotesque"`)) || !bytes.Contains(css, []byte(`font-family: "JetBrains Mono"`)) {
		t.Error("bundled families missing self-hosted @font-face under remoteFonts")
	}
	if _, err := os.Stat(filepath.Join(dir, "_site", "static", "leafpress", "fonts", "bricolage-grotesque-normal-latin.woff2")); err != nil {
		t.Errorf("bundled font not materialized under remoteFonts: %v", err)
	}
}

func TestBuildWithCustomLocalFont(t *testing.T) {
	dir := newTestProject(t)
	fontDir := filepath.Join(dir, "static", "fonts")
	if err := os.MkdirAll(fontDir, 0755); err != nil {
		t.Fatal(err)
	}
	fontBytes := []byte("wOF2fakefont")
	if err := os.WriteFile(filepath.Join(fontDir, "my.woff2"), fontBytes, 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Theme.FontBody = "My Serif"
	cfg.Theme.Fonts = []config.FontFace{{Family: "My Serif", File: "static/fonts/my.woff2", Weight: "400 700"}}
	b := New(cfg, Options{})
	if _, err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// copyStatic ships the font file; the page references it locally.
	copied, err := os.ReadFile(filepath.Join(dir, "_site", "static", "fonts", "my.woff2"))
	if err != nil {
		t.Fatalf("custom font not copied: %v", err)
	}
	if !bytes.Equal(copied, fontBytes) {
		t.Error("copied font differs")
	}
	// @font-face rules live in the shared stylesheet, not page heads.
	css, err := os.ReadFile(filepath.Join(dir, "_site", "style.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(css, []byte(`font-family: "My Serif"`)) {
		t.Error("custom @font-face missing from style.css")
	}
	if !bytes.Contains(css, []byte(`url("static/fonts/my.woff2")`)) {
		t.Error("custom font URL must be stylesheet-relative")
	}
	page, err := os.ReadFile(filepath.Join(dir, "_site", "note", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(page, []byte("My+Serif")) || bytes.Contains(css, []byte("My+Serif")) {
		t.Error("declared family leaked into a remote font URL")
	}
}

func TestBuildFailsWhenCustomFontFileMissing(t *testing.T) {
	newTestProject(t)

	cfg := config.Default()
	cfg.Theme.Fonts = []config.FontFace{{Family: "Ghost", File: "static/fonts/ghost.woff2"}}
	b := New(cfg, Options{})
	if _, err := b.Build(); err == nil {
		t.Fatal("Build must fail when a declared font file does not exist")
	}
}

func TestBuildRejectsUserFilesInReservedNamespace(t *testing.T) {
	dir := newTestProject(t)
	reserved := filepath.Join(dir, "static", "leafpress")
	if err := os.MkdirAll(reserved, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(reserved, "mine.css"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	b := New(config.Default(), Options{})
	_, err := b.Build()
	if err == nil {
		t.Fatal("Build must reject user files under static/leafpress")
	}
	if !strings.Contains(err.Error(), "static/leafpress is reserved") {
		t.Fatalf("error must name the reserved namespace, got: %v", err)
	}
}

func TestBuildSelfHostsMermaidWhenUsed(t *testing.T) {
	dir := newTestProject(t)
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte(""+
		"# Note\n\n```mermaid\ngraph TD\n  A-->B\n```\n"), 0644); err != nil {
		t.Fatal(err)
	}

	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	jsPath := filepath.Join(dir, "_site", "static", "leafpress", "mermaid", "mermaid.min.js")
	got, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("mermaid must materialize when diagrams are present: %v", err)
	}
	builtin, ok := assets.BuiltinByLogicalPath(assets.BuiltinMermaidJS)
	if !ok {
		t.Fatal("registry missing mermaid")
	}
	if !bytes.Equal(got, builtin.Content()) {
		t.Error("materialized mermaid does not match registry bytes")
	}
	lic := filepath.Join(dir, "_site", "static", "leafpress", "mermaid", "LICENSE.txt")
	if _, err := os.Stat(lic); err != nil {
		t.Fatalf("mermaid license must materialize: %v", err)
	}

	page, err := os.ReadFile(filepath.Join(dir, "_site", "note", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(page)
	if strings.Contains(html, "cdn.jsdelivr") || strings.Contains(html, "cdnjs") {
		t.Error("page must not load Mermaid from a third-party CDN")
	}
	if !strings.Contains(html, "/static/leafpress/mermaid/mermaid.min.js") {
		t.Error("page must load self-hosted mermaid path")
	}
}

func TestBuildSkipsMermaidWhenUnused(t *testing.T) {
	dir := newTestProject(t)
	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}
	jsPath := filepath.Join(dir, "_site", "static", "leafpress", "mermaid", "mermaid.min.js")
	if _, err := os.Stat(jsPath); !os.IsNotExist(err) {
		t.Error("mermaid must not materialize when no diagrams are present")
	}
}

func TestBuildEmitsSearchIndexWhenSearchUIDisabled(t *testing.T) {
	dir := newTestProject(t)
	// Link target so the page has a wikilink for preview script attachment.
	if err := os.WriteFile(filepath.Join(dir, "other.md"), []byte("# Other\n\nbody\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("# Note\n\nSee [[other]].\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Features.Search = false
	cfg.Features.Graph = false
	b := New(cfg, Options{})
	if _, err := b.Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}

	indexPath := filepath.Join(dir, "_site", "search-index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("search-index.json must be emitted when search UI is off: %v", err)
	}
	if !strings.Contains(string(data), `"title": "Note"`) || !strings.Contains(string(data), `/note/`) {
		t.Errorf("search-index.json missing expected page entries: %s", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "_site", "graph.json")); !os.IsNotExist(err) {
		t.Error("graph.json must not be emitted when graph is disabled")
	}

	page, err := os.ReadFile(filepath.Join(dir, "_site", "note", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(page)
	if strings.Contains(html, `class="lp-search-toggle"`) {
		t.Error("search UI toggle must stay off when search is false")
	}
	if !strings.Contains(html, "search-index.json") {
		t.Error("link preview script must still reference search-index.json")
	}
}

func TestBuildGraphEdgesDoNotDependOnBacklinks(t *testing.T) {
	dir := newTestProject(t)
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte("# Note\n\n[[other]] and [[Other]]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "other.md"), []byte("# Other\n\nbody\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Default()
	cfg.Features.Graph = true
	cfg.Features.Backlinks = false
	if _, err := New(cfg, Options{}).Build(); err != nil {
		t.Fatalf("Build: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "_site", "graph.json"))
	if err != nil {
		t.Fatal(err)
	}
	graph := string(data)
	if got := strings.Count(graph, `"source": "note"`); got != 1 {
		t.Fatalf("note edge count = %d, want 1: %s", got, graph)
	}
	if !strings.Contains(graph, `"target": "other"`) {
		t.Fatalf("graph is missing note -> other: %s", graph)
	}
}

func TestBuildTagIndexDeduplicatesCaseVariantsPerPage(t *testing.T) {
	page := &content.Page{Tags: []string{"Go", "go"}}
	index := buildTagIndex([]*content.Page{page})
	if got := len(index["go"]); got != 1 {
		t.Fatalf("go tag page count = %d, want 1", got)
	}
}
