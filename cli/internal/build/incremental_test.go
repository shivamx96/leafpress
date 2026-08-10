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

func TestIncrementalInlineTagChangesRebuildTagPages(t *testing.T) {
	dir := newTestProject(t)
	notePath := filepath.Join(dir, "note.md")
	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(notePath, []byte("# Note\n\nNow tracking #alpha.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	stats, err := b.RebuildIncremental(notePath, ChangeModify)
	if err != nil {
		t.Fatal(err)
	}
	if stats.TagsRebuilt != 1 {
		t.Fatalf("TagsRebuilt = %d, want 1", stats.TagsRebuilt)
	}
	alphaPage := filepath.Join(dir, "_site", "tags", "alpha", "index.html")
	assertFileContains(t, alphaPage, ">Note<")

	if err := os.WriteFile(notePath, []byte("# Note\n\nNow tracking #beta.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	stats, err = b.RebuildIncremental(notePath, ChangeModify)
	if err != nil {
		t.Fatal(err)
	}
	if stats.TagsRebuilt != 2 {
		t.Fatalf("TagsRebuilt = %d, want 2", stats.TagsRebuilt)
	}
	if _, err := os.Stat(alphaPage); !os.IsNotExist(err) {
		t.Fatalf("removed inline tag page still exists: %v", err)
	}
	assertFileContains(t, filepath.Join(dir, "_site", "tags", "beta", "index.html"), ">Note<")
}

func TestFailedConfigRebuildRestoresPreviousBuilderState(t *testing.T) {
	dir := newTestProject(t)
	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatalf("first Build: %v", err)
	}

	oldCfg := b.cfg
	oldOutputDir := b.outputDir
	oldTemplates := b.templates
	pagePath := filepath.Join(dir, "_site", "note", "index.html")
	published, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "leafpress.json")
	if err := os.WriteFile(configPath, []byte(`{"site":{"title":"Updated"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "static", "leafpress"), 0755); err != nil {
		t.Fatal(err)
	}

	if _, err := b.RebuildIncremental(configPath, ChangeModify); err == nil {
		t.Fatal("config rebuild should fail during static generation")
	}
	if b.cfg != oldCfg || b.outputDir != oldOutputDir || b.templates != oldTemplates {
		t.Error("failed config rebuild did not restore the previous builder configuration")
	}
	got, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(published) {
		t.Error("failed config rebuild replaced the published output")
	}
	assertNoOutputTransactions(t, dir, "_site")
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
