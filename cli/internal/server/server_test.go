package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

func TestStaticHandlerResolutionAndTraversalRejection(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "site")
	if err := os.MkdirAll(filepath.Join(root, "notes"), 0755); err != nil {
		t.Fatal(err)
	}
	for path, body := range map[string]string{
		filepath.Join(root, "index.html"):          "home",
		filepath.Join(root, "notes", "index.html"): "notes index",
		filepath.Join(root, "404.html"):            "custom missing",
		filepath.Join(parent, "secret.txt"):        "outside secret",
	} {
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}

	handler := (&Server{}).handleStatic(root)
	tests := []struct {
		path       string
		status     int
		want       string
		mustNotSee string
	}{
		{path: "/", status: http.StatusOK, want: "home"},
		{path: "/notes", status: http.StatusOK, want: "notes index"},
		{path: "/missing", status: http.StatusNotFound, want: "custom missing"},
		{path: "/../secret.txt", status: http.StatusNotFound, want: "custom missing", mustNotSee: "outside secret"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://example.test"+tt.path, nil)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			body, err := io.ReadAll(res.Result().Body)
			if err != nil {
				t.Fatal(err)
			}
			if res.Code != tt.status || !strings.Contains(string(body), tt.want) {
				t.Fatalf("status/body = %d, %q; want %d containing %q", res.Code, body, tt.status, tt.want)
			}
			if tt.mustNotSee != "" && strings.Contains(string(body), tt.mustNotSee) {
				t.Fatalf("response exposed %q", tt.mustNotSee)
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
