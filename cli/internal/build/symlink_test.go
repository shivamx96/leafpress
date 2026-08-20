package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/config"
)

// linkOrSkip creates a symlink, skipping the test on platforms that refuse.
func linkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
}

// secretOutside writes a file in a sibling directory that no build should
// ever be able to read into the published site.
func secretOutside(t *testing.T) string {
	t.Helper()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.md")
	if err := os.WriteFile(secret, []byte("# Secret\n\nprivate key material\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return secret
}

func TestBuildRefusesContentSymlinkOutsideProject(t *testing.T) {
	secret := secretOutside(t)
	dir := newTestProject(t)
	linkOrSkip(t, secret, filepath.Join(dir, "leaked.md"))

	_, err := New(config.Default(), Options{}).Build()
	if err == nil {
		t.Fatal("Build should refuse a content symlink that leaves the project")
	}
	if !strings.Contains(err.Error(), "outside the project") {
		t.Fatalf("error should name the escape, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "_site", "leaked", "index.html")); statErr == nil {
		t.Fatal("the escaping note was published anyway")
	}
}

func TestBuildRefusesStaticSymlinkOutsideProject(t *testing.T) {
	secret := secretOutside(t)
	dir := newTestProject(t)
	staticDir := filepath.Join(dir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatal(err)
	}
	linkOrSkip(t, secret, filepath.Join(staticDir, "leaked.txt"))

	_, err := New(config.Default(), Options{}).Build()
	if err == nil {
		t.Fatal("Build should refuse a static symlink that leaves the project")
	}
	if !strings.Contains(err.Error(), "outside the project") {
		t.Fatalf("error should name the escape, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "_site", "static", "leaked.txt")); statErr == nil {
		t.Fatal("the escaping static file was published anyway")
	}
}

// Links that stay inside the garden point at content that was already going
// to be published, so they must keep working.
func TestBuildFollowsSymlinksInsideProject(t *testing.T) {
	dir := newTestProject(t)
	real := filepath.Join(dir, "real")
	if err := os.MkdirAll(real, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(real, "target.md"), []byte("---\ntitle: Target\n---\n\nbody\n"), 0644); err != nil {
		t.Fatal(err)
	}
	linkOrSkip(t, filepath.Join(real, "target.md"), filepath.Join(dir, "alias.md"))

	staticDir := filepath.Join(dir, "static")
	if err := os.MkdirAll(filepath.Join(staticDir, "real"), 0755); err != nil {
		t.Fatal(err)
	}
	asset := filepath.Join(staticDir, "real", "a.txt")
	if err := os.WriteFile(asset, []byte("asset\n"), 0644); err != nil {
		t.Fatal(err)
	}
	linkOrSkip(t, asset, filepath.Join(staticDir, "alias.txt"))

	if _, err := New(config.Default(), Options{}).Build(); err != nil {
		t.Fatalf("in-project symlinks must still build: %v", err)
	}
	for _, rel := range []string{"alias/index.html", "static/alias.txt"} {
		if _, err := os.Stat(filepath.Join(dir, "_site", filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s should have been published: %v", rel, err)
		}
	}
}

// A serve session re-parses single files; it must apply the same boundary a
// full build does, or watch mode becomes the way around the check.
func TestIncrementalRebuildRefusesSymlinkOutsideProject(t *testing.T) {
	secret := secretOutside(t)
	dir := newTestProject(t)

	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}

	linkOrSkip(t, secret, filepath.Join(dir, "leaked.md"))
	_, err := b.RebuildIncremental("leaked.md", ChangeCreate)
	if err == nil {
		t.Fatal("incremental rebuild should refuse a symlink that leaves the project")
	}
	if !strings.Contains(err.Error(), "outside the project") {
		t.Fatalf("error should name the escape, got: %v", err)
	}
}
