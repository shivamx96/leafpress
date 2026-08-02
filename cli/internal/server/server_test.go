package server

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/shivamx96/leafpress/cli/internal/build"
)

func TestMergeChangeType(t *testing.T) {
	tests := []struct {
		name     string
		previous build.ChangeType
		next     build.ChangeType
		want     build.ChangeType
	}{
		{name: "create then modify stays create", previous: build.ChangeCreate, next: build.ChangeModify, want: build.ChangeCreate},
		{name: "delete then create becomes create", previous: build.ChangeDelete, next: build.ChangeCreate, want: build.ChangeCreate},
		{name: "modify then delete becomes delete", previous: build.ChangeModify, next: build.ChangeDelete, want: build.ChangeDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeChangeType(tt.previous, tt.next); got != tt.want {
				t.Errorf("mergeChangeType(%v, %v) = %v, want %v", tt.previous, tt.next, got, tt.want)
			}
		})
	}
}

func TestAddWatchDirsIncludesNestedDirectories(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "new", "nested")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	ignored := filepath.Join(root, "_site", "ignored")
	if err := os.MkdirAll(ignored, 0755); err != nil {
		t.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Close()
	s := &Server{watcher: watcher}
	if err := s.addWatchDirs(root); err != nil {
		t.Fatal(err)
	}
	watched := watcher.WatchList()
	for _, dir := range []string{root, filepath.Join(root, "new"), nested} {
		if !slices.Contains(watched, dir) {
			t.Errorf("watch list does not include %s: %v", dir, watched)
		}
	}
	if slices.Contains(watched, ignored) {
		t.Errorf("watch list includes output directory: %v", watched)
	}
}
