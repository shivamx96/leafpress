package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugSegment(t *testing.T) {
	tests := map[string]string{
		// URL-safe names keep their published URL.
		"hello-world":     "hello-world",
		"Ideas":           "Ideas",
		"migration_index": "migration_index",
		"v1.2":            "v1.2",
		"a--b":            "a--b",
		"-draft":          "-draft",
		"éclair":          "éclair",
		"e\u0301clair":    "e\u0301clair",
		"日本語":             "日本語",
		// Separators become one hyphen and merge with neighboring hyphens.
		"My Note":            "My-Note",
		"Q&A":                "Q-A",
		"has#fragment":       "has-fragment",
		"encoded%20name":     "encoded-20name",
		"a - b":              "a-b",
		"Meeting (2024)":     "Meeting-2024",
		"Don't Panic":        "Don-t-Panic",
		"what?":              "what",
		"  padded  ":         "padded",
		"🌱 Seeds":            "Seeds",
		"tab\tand\nnewline":  "tab-and-newline",
		`back\slash:colon*|`: "back-slash-colon",
		// Trailing dots are not portable.
		"Wait...": "Wait",
		"a? .":    "a",
		// Nothing usable.
		"???": "",
		"&":   "",
	}
	for input, want := range tests {
		if got := SlugSegment(input); got != want {
			t.Errorf("SlugSegment(%q) = %q, want %q", input, got, want)
		}
		if got := SlugSegment(want); got != want {
			t.Errorf("SlugSegment is not stable for %q: got %q", want, got)
		}
	}
}

func TestPageSlugFrontmatter(t *testing.T) {
	tests := []struct {
		source, frontmatter string
		isIndex             bool
		want, err           string
	}{
		{source: "notes/My Note", frontmatter: "first-note", want: "notes/first-note"},
		{source: "My Projects/Idea", frontmatter: "idea", want: "My-Projects/idea"},
		{source: "notes/My Note", frontmatter: "My Note", err: `use "My-Note"`},
		{source: "notes/My Note", frontmatter: "a/b", err: "without slashes"},
		{source: "notes/My Note", frontmatter: "???", err: "no URL-safe characters"},
		{source: "notes", frontmatter: "renamed", isIndex: true, err: "index pages"},
		{source: "", frontmatter: "home", err: "index pages"},
		{source: "notes/???", err: `filename "???"`},
		{source: "???/note", err: `folder name "???"`},
		{source: "???", isIndex: true, err: `folder name "???"`},
	}
	for _, tt := range tests {
		got, err := pageSlug(tt.source, tt.frontmatter, tt.isIndex)
		if tt.err != "" {
			if err == nil || !strings.Contains(err.Error(), tt.err) {
				t.Errorf("pageSlug(%q, %q) error = %v, want containing %q", tt.source, tt.frontmatter, err, tt.err)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Errorf("pageSlug(%q, %q) = %q, %v; want %q", tt.source, tt.frontmatter, got, err, tt.want)
		}
	}
}

func TestScannedSlugsKeepFilenameTitlesAndLinks(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"index.md":               "Home",
		"Q&A.md":                 "Answers",
		"notes/My Note.md":       "Note",
		"notes/Custom Name.md":   "---\nslug: renamed\n---\nCustom",
		"Field Notes/_index.md":  "Section",
		"Field Notes/Weather.md": "Weather",
	}
	for name, body := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	pages, err := NewScanner(root, nil).Scan()
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateOutputRoutes(pages); err != nil {
		t.Fatal(err)
	}
	bySource := make(map[string]*Page)
	for _, page := range pages {
		bySource[filepath.ToSlash(page.SourcePath)] = page
	}
	for source, want := range map[string][2]string{
		"Q&A.md":                 {"Q-A", "Q&A"},
		"notes/My Note.md":       {"notes/My-Note", "My Note"},
		"notes/Custom Name.md":   {"notes/renamed", "Custom Name"},
		"Field Notes/_index.md":  {"Field-Notes", "Field Notes"},
		"Field Notes/Weather.md": {"Field-Notes/Weather", "Weather"},
	} {
		page := bySource[source]
		if page == nil || page.Slug != want[0] || page.Title != want[1] {
			t.Errorf("%s: got %+v, want slug %q title %q", source, page, want[0], want[1])
		}
	}

	resolver := NewLinkResolver(pages)
	for target, source := range map[string]string{
		"Q&A":                 "Q&A.md",
		"q-a":                 "Q&A.md",
		"My Note":             "notes/My Note.md",
		"notes/My Note":       "notes/My Note.md",
		"notes/my-note":       "notes/My Note.md",
		"Custom Name":         "notes/Custom Name.md",
		"renamed":             "notes/Custom Name.md",
		"notes/Custom Name":   "notes/Custom Name.md",
		"Field Notes":         "Field Notes/_index.md",
		"Field Notes/Weather": "Field Notes/Weather.md",
	} {
		result := resolver.Resolve(target)
		if result.Broken || result.Ambiguous || result.Page != bySource[source] {
			t.Errorf("Resolve(%q) = %+v, want %s", target, result, source)
		}
	}
}

func TestCleanedFilenameCollisionSuggestsFrontmatterSlug(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"My Note.md", "My-Note.md"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	pages, err := NewScanner(root, nil).Scan()
	if err != nil {
		t.Fatal(err)
	}
	err = ValidateOutputRoutes(pages)
	if err == nil || !strings.Contains(err.Error(), "/My-Note/") || !strings.Contains(err.Error(), "slug in its frontmatter") {
		t.Fatalf("collision error = %v", err)
	}
}

func TestScanReportsFileWithInvalidFrontmatterSlug(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "note.md"), []byte("---\nslug: My Note\n---\nx"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, scan := range []func() error{
		func() error { _, err := NewScanner(root, nil).Scan(); return err },
		func() error { _, err := ParseSingleFile(root, "note.md"); return err },
	} {
		if err := scan(); err == nil || !strings.Contains(err.Error(), "note.md") || !strings.Contains(err.Error(), `use "My-Note"`) {
			t.Errorf("error = %v", err)
		}
	}
}
