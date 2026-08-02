package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/config"
)

func TestIncrementalSectionListingsTrackAddsAndDeletes(t *testing.T) {
	dir := newTestProject(t)
	sectionDir := filepath.Join(dir, "field-notes")
	if err := os.MkdirAll(sectionDir, 0755); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(sectionDir, "_index.md")
	onePath := filepath.Join(sectionDir, "one.md")
	if err := os.WriteFile(indexPath, []byte("# Field Notes\n\nManual introduction.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(onePath, []byte("# One\n\nFirst note.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}
	twoPath := filepath.Join(sectionDir, "two.md")
	if err := os.WriteFile(twoPath, []byte("# Two\n\nSecond note.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(twoPath, ChangeCreate); err != nil {
		t.Fatal(err)
	}

	outputIndex := filepath.Join(dir, "_site", "field-notes", "index.html")
	assertFileContains(t, outputIndex, "Manual introduction.", ">Two<", ">One<")

	if err := os.Remove(onePath); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(onePath, ChangeDelete); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, outputIndex, "Manual introduction.", ">Two<")
	data, err := os.ReadFile(outputIndex)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), ">One<") {
		t.Fatal("deleted page remains in the hand-authored section listing")
	}

	// Removing the manual index should replace it with an auto-index rather
	// than deleting the section listing while pages remain.
	if err := os.Remove(indexPath); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(indexPath, ChangeDelete); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, outputIndex, ">Two<")
	data, err = os.ReadFile(outputIndex)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Manual introduction.") {
		t.Fatal("deleted manual index content remains in the auto-index")
	}
}

func TestIncrementalStaticDeletionRemovesOutput(t *testing.T) {
	dir := newTestProject(t)
	staticDir := filepath.Join(dir, "static", "images")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(staticDir, "logo.txt")
	if err := os.WriteFile(source, []byte("logo"), 0644); err != nil {
		t.Fatal(err)
	}

	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "_site", "static", "images", "logo.txt")
	if _, err := os.Stat(output); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(source, ChangeDelete); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("deleted static output still exists: %v", err)
	}
}

func assertFileContains(t *testing.T, path string, wants ...string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range wants {
		if !strings.Contains(string(data), want) {
			t.Errorf("%s does not contain %q", path, want)
		}
	}
}
