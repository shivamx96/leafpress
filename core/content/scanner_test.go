package content

import (
	"path/filepath"
	"testing"
)

func TestScannerPathDerivations(t *testing.T) {
	tests := []struct {
		path      string
		slug      string
		output    string
		permalink string
	}{
		{path: "index.md", slug: "", output: "index.html", permalink: "/"},
		{path: "notes/hello.md", slug: "notes/hello", output: filepath.Join("notes", "hello", "index.html"), permalink: "/notes/hello/"},
		{path: "notes/_index.md", slug: "notes", output: filepath.Join("notes", "index.html"), permalink: "/notes/"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			slug := generateSlug(filepath.FromSlash(tt.path))
			if slug != tt.slug {
				t.Errorf("slug = %q, want %q", slug, tt.slug)
			}
			if got := generateOutputPath(slug, filepath.Base(tt.path) == "_index.md"); got != tt.output {
				t.Errorf("output = %q, want %q", got, tt.output)
			}
			if got := generatePermalink(slug, filepath.Base(tt.path) == "_index.md"); got != tt.permalink {
				t.Errorf("permalink = %q, want %q", got, tt.permalink)
			}
		})
	}
}

func TestGenerateTitleFromUnicodeSlug(t *testing.T) {
	if got := generateTitleFromSlug("éclair-notes"); got != "Éclair Notes" {
		t.Fatalf("title = %q, want Éclair Notes", got)
	}
}
