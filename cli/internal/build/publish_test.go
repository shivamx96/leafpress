package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/config"
)

func TestExplicitPublishingAndIncrementalWithdrawal(t *testing.T) {
	dir := newTestProject(t)
	write := func(name, body string) string {
		t.Helper()
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	note := write("note.md", "---\npublish: true\n---\nPublic body\n")
	write("private.md", "Private body")
	write("_index.md", "Private home body")
	write("ignored.md", "---\npublish: true\n---\nIgnored body")
	write(".hidden.md", "---\npublish: true\n---\nHidden body")
	write("draft.md", "---\npublish: false\ndraft: true\n---\nPrivate draft")
	cfg := config.Default()
	cfg.Build.Publish = "explicit"
	cfg.Build.Ignore = []string{"ignored.md"}
	b := New(cfg, Options{IncludeDrafts: true})
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "_site")
	assertFileContains(t, filepath.Join(out, "note", "index.html"), "Public body")
	if body, err := os.ReadFile(filepath.Join(out, "index.html")); err == nil {
		if strings.Contains(string(body), "Private home body") {
			t.Fatal("private home content published")
		}
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, name := range []string{"private", "draft", "ignored", ".hidden"} {
		if _, err := os.Stat(filepath.Join(out, name, "index.html")); !os.IsNotExist(err) {
			t.Fatalf("private output %s exists: %v", name, err)
		}
	}
	write("note.md", "Withdrawn body")
	if _, err := b.RebuildIncremental(note, ChangeModify); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "note", "index.html")); !os.IsNotExist(err) {
		t.Fatalf("withdrawn output exists: %v", err)
	}
	assertFileNotContains(t, filepath.Join(out, "search-index.json"), `"/note/"`)
	write("note.md", "---\npublish: true\n---\nRestored body")
	if _, err := b.RebuildIncremental(note, ChangeModify); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(out, "note", "index.html"), "Restored body")
}
