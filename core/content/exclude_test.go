package content

import (
	"os"
	"path/filepath"
	"testing"
)

// docs/ was reserved, so a garden with a docs/ folder silently lost every
// page in it. Authors exclude their own folders with build.ignore instead.
func TestScannerPublishesDocsFolder(t *testing.T) {
	if ReservedPaths["docs"] {
		t.Fatal("docs must not be a reserved path")
	}

	dir := t.TempDir()
	full := filepath.Join(dir, "docs", "guide.md")
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte("# Guide\n\nbody\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pages, err := NewScanner(dir, nil).Scan()
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 || pages[0].Slug != "docs/guide" {
		t.Fatalf("docs/guide.md should be published; got %d pages", len(pages))
	}

	// ...and build.ignore still hides it for authors who want that.
	pages, err = NewScanner(dir, []string{"docs/**"}).Scan()
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 0 {
		t.Fatalf("build.ignore should still exclude docs/; got %d pages", len(pages))
	}
}

func TestIsExcluded(t *testing.T) {
	ignore, err := NewIgnoreMatcher([]string{"drafts/**"})
	if err != nil {
		t.Fatal(err)
	}

	excluded := []string{
		"_site", "_site/index.html", "node_modules/pkg/a.md",
		".git/config", ".obsidian/workspace.json", "static", "leafpress.json",
		"drafts/a.md", "notes/.hidden/a.md",
	}
	for _, p := range excluded {
		if !IsExcluded(filepath.FromSlash(p), ignore) {
			t.Errorf("%q should be excluded", p)
		}
	}

	kept := []string{"note.md", "docs/guide.md", "notes/deep/a.md", "draftsy/a.md"}
	for _, p := range kept {
		if IsExcluded(filepath.FromSlash(p), ignore) {
			t.Errorf("%q should NOT be excluded", p)
		}
	}

	if IsExcluded(".", nil) || IsExcluded("", nil) {
		t.Error("the garden root is not excluded")
	}
}
