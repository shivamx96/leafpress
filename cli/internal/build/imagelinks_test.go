package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/config"
)

func TestBuildLinksImagesToTheirSource(t *testing.T) {
	dir := newTestProject(t)
	note := "# Note\n\n![A photo](/static/images/photo.png)\n\n[![Badge](/static/images/badge.svg)](https://example.com/project)\n"
	if err := os.WriteFile(filepath.Join(dir, "note.md"), []byte(note), 0644); err != nil {
		t.Fatal(err)
	}

	b := New(config.Default(), Options{})
	if _, err := b.Build(); err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join(dir, "_site", "note", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(html)
	if !strings.Contains(page, `<a class="lp-image-link" href="/static/images/photo.png" target="_blank" rel="noopener"><img src="/static/images/photo.png"`) {
		t.Errorf("plain image should link to its source, got:\n%s", page)
	}
	if strings.Count(page, "lp-image-link") != 1 {
		t.Errorf("the author-linked badge must keep its own link, got:\n%s", page)
	}
	if !strings.Contains(page, `href="https://example.com/project"`) {
		t.Error("author link was lost")
	}
}
