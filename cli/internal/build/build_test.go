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

	for _, logicalPath := range []string{
		assets.BuiltinFaviconICO,
		assets.BuiltinFaviconSVG,
		assets.BuiltinFaviconPNG,
	} {
		builtin, ok := assets.BuiltinByLogicalPath(logicalPath)
		if !ok {
			t.Fatalf("registry missing %s", logicalPath)
		}
		name := builtin.Asset.EffectivePublicPath()
		got, err := os.ReadFile(filepath.Join(dir, "_site", name))
		if err != nil {
			t.Fatalf("favicon %s not written: %v", name, err)
		}
		if !bytes.Equal(got, builtin.Content) {
			t.Errorf("favicon %s does not match registry content", name)
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
	if !bytes.Equal(svg, builtin.Content) {
		t.Error("favicon.svg does not match registry content")
	}
}
