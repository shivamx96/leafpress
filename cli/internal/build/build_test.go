package build

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/shivamx96/leafpress/core/assets"
	"github.com/shivamx96/leafpress/core/config"
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

	// Default theme uses the three bundled families: every face file must be
	// on disk at its logical path with registry content.
	for _, face := range assets.BuiltinFontFaces() {
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
	for _, family := range []string{"Crimson Pro", "Inter", "JetBrains Mono"} {
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
	if bytes.Contains(page, []byte("family=Crimson+Pro")) || bytes.Contains(page, []byte("family=JetBrains+Mono")) {
		t.Error("bundled families leaked into the remote font URL")
	}
	css, err := os.ReadFile(filepath.Join(dir, "_site", "style.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(css, []byte(`font-family: "Crimson Pro"`)) || !bytes.Contains(css, []byte(`font-family: "JetBrains Mono"`)) {
		t.Error("bundled families missing self-hosted @font-face under remoteFonts")
	}
	if _, err := os.Stat(filepath.Join(dir, "_site", "static", "leafpress", "fonts", "crimson-pro-normal-latin.woff2")); err != nil {
		t.Errorf("bundled font not materialized under remoteFonts: %v", err)
	}
}
