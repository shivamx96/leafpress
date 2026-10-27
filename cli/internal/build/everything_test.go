package build

import (
	"github.com/shivamx96/leafpress/core/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEverythingBuildAndIncremental(t *testing.T) {
	dir := newTestProject(t)
	cfg := config.Default()
	cfg.Site.BaseURL = "https://example.com/garden"
	cfg.Navigation.Everything = true
	cfg.Navigation.EverythingExclude = []string{"/about/"}
	for name, body := range map[string]string{"note.md": "---\ntitle: Original\n---\nBody\n", "about.md": "# About\n", "draft.md": "---\ndraft: true\n---\n# Private\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	b := New(cfg, Options{})
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "_site/everything/index.html")
	check := func(want string, absent ...string) {
		t.Helper()
		data, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		s := string(data)
		s = s[strings.Index(s, "<main"):strings.Index(s, "</main>")]
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in %s", want, s)
		}
		for _, v := range absent {
			if strings.Contains(s, v) {
				t.Fatalf("unexpected %q in listing", v)
			}
		}
	}
	check("Original", "/garden/about/", "/garden/draft/")
	note := filepath.Join(dir, "note.md")
	if err := os.WriteFile(note, []byte("---\ntitle: Updated\n---\nBody\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(note, ChangeModify); err != nil {
		t.Fatal(err)
	}
	check("Updated", "Original")
	if err := os.Remove(note); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(note, ChangeDelete); err != nil {
		t.Fatal(err)
	}
	check("Everything", "Updated")
	collision := filepath.Join(dir, "everything.md")
	if err := os.WriteFile(collision, []byte("# Collision\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := b.RebuildIncremental(collision, ChangeCreate); err == nil {
		t.Fatal("accepted incremental collision")
	}
	if _, err := b.Build(); err == nil {
		t.Fatal("accepted full-build collision")
	}
}

func TestBottomNavigationBuild(t *testing.T) {
	dir := newTestProject(t)
	cfg := config.Default()
	cfg.Theme.NavStyle = "bottom"
	cfg.Navigation.Everything = true
	b := New(cfg, Options{})
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "_site", "everything", "index.html")
	assertFileContains(t, output, `data-nav-style="bottom"`)
	assertFileContains(t, output, `class="lp-dock-home"`)
}
