package server

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/shivamx96/leafpress/core/content"
)

// The watcher pruned a hardcoded list that had drifted from the scanner's.
// Anything the content scan drops must not be watched, or serve publishes
// pages the next full build removes.
func TestAddWatchDirsMatchesScannerExclusions(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{
		"notes", "docs", "drafts/deep", "static/img",
		"_site/out", "node_modules/pkg", ".obsidian",
	} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(rel)), 0755); err != nil {
			t.Fatal(err)
		}
	}

	ignore, err := content.NewIgnoreMatcher([]string{"drafts/**"})
	if err != nil {
		t.Fatal(err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()

	s := &Server{watcher: watcher, ignore: ignore}
	if err := s.addWatchDirs(root); err != nil {
		t.Fatal(err)
	}
	watched := watcher.WatchList()

	// docs/ is ordinary content now; static/ is reserved for the scan but
	// still watched, because its files are copied into the site.
	for _, rel := range []string{"notes", "docs", "static", "static/img"} {
		dir := filepath.Join(root, filepath.FromSlash(rel))
		if !slices.Contains(watched, dir) {
			t.Errorf("%s should be watched: %v", rel, watched)
		}
	}
	for _, rel := range []string{"drafts", "drafts/deep", "_site", "node_modules", ".obsidian"} {
		dir := filepath.Join(root, filepath.FromSlash(rel))
		if slices.Contains(watched, dir) {
			t.Errorf("%s should not be watched: %v", rel, watched)
		}
	}
}

func TestIsStaticTree(t *testing.T) {
	for _, rel := range []string{"static", filepath.Join("static", "a.css")} {
		if !isStaticTree(rel) {
			t.Errorf("%q is in the static tree", rel)
		}
	}
	for _, rel := range []string{"staticky", filepath.Join("notes", "static"), "note.md"} {
		if isStaticTree(rel) {
			t.Errorf("%q is not in the static tree", rel)
		}
	}
}
