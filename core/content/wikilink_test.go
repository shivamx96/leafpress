package content

import (
	"testing"
)

func TestExtractWikiLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []WikiLink
	}{
		{
			name:  "simple link",
			input: "Check out [[my-page]]",
			expected: []WikiLink{
				{Target: "my-page", Label: "my-page", Raw: "[[my-page]]"},
			},
		},
		{
			name:  "link with label",
			input: "See [[my-page|custom text]]",
			expected: []WikiLink{
				{Target: "my-page", Label: "custom text", Raw: "[[my-page|custom text]]"},
			},
		},
		{
			name:  "nested path",
			input: "Read [[guide/installation]]",
			expected: []WikiLink{
				{Target: "guide/installation", Label: "guide/installation", Raw: "[[guide/installation]]"},
			},
		},
		{
			name:  "multiple links",
			input: "See [[page-a]] and [[page-b|B]]",
			expected: []WikiLink{
				{Target: "page-a", Label: "page-a", Raw: "[[page-a]]"},
				{Target: "page-b", Label: "B", Raw: "[[page-b|B]]"},
			},
		},
		{
			name:     "no links",
			input:    "Just regular text",
			expected: nil,
		},
		{
			name:  "whitespace in target",
			input: "[[  spaced  ]]",
			expected: []WikiLink{
				{Target: "spaced", Label: "spaced", Raw: "[[  spaced  ]]"},
			},
		},
		{
			name:  "whitespace in label",
			input: "[[page |  label with spaces  ]]",
			expected: []WikiLink{
				{Target: "page", Label: "label with spaces", Raw: "[[page |  label with spaces  ]]"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractWikiLinks(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d links, want %d", len(got), len(tt.expected))
			}
			for i, link := range got {
				if link.Target != tt.expected[i].Target {
					t.Errorf("link[%d].Target = %q, want %q", i, link.Target, tt.expected[i].Target)
				}
				if link.Label != tt.expected[i].Label {
					t.Errorf("link[%d].Label = %q, want %q", i, link.Label, tt.expected[i].Label)
				}
				if link.Raw != tt.expected[i].Raw {
					t.Errorf("link[%d].Raw = %q, want %q", i, link.Raw, tt.expected[i].Raw)
				}
			}
		})
	}
}

func TestLinkResolver_Resolve(t *testing.T) {
	pages := []*Page{
		{Slug: "getting-started", Title: "Getting Started"},
		{Slug: "guide/installation", Title: "Installation"},
		{Slug: "notes/installation", Title: "Install Notes"},
	}
	resolver := NewLinkResolver(pages)

	tests := []struct {
		name       string
		target     string
		wantSlug   string
		wantAmb    bool
		wantBroken bool
	}{
		{
			name:     "exact slug match",
			target:   "getting-started",
			wantSlug: "getting-started",
		},
		{
			name:     "case insensitive",
			target:   "Getting-Started",
			wantSlug: "getting-started",
		},
		{
			name:     "exact nested slug",
			target:   "guide/installation",
			wantSlug: "guide/installation",
		},
		{
			name:     "filename match",
			target:   "installation",
			wantSlug: "guide/installation",
			wantAmb:  true,
		},
		{
			name:       "broken link",
			target:     "nonexistent",
			wantBroken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.Resolve(tt.target)
			if result.Broken != tt.wantBroken {
				t.Errorf("Broken = %v, want %v", result.Broken, tt.wantBroken)
			}
			if result.Ambiguous != tt.wantAmb {
				t.Errorf("Ambiguous = %v, want %v", result.Ambiguous, tt.wantAmb)
			}
			if !tt.wantBroken && result.Page.Slug != tt.wantSlug {
				t.Errorf("Page.Slug = %q, want %q", result.Page.Slug, tt.wantSlug)
			}
		})
	}
}

func TestBuildBacklinks(t *testing.T) {
	pageA := &Page{Slug: "page-a", RawContent: "Links to [[page-b]] and [[page-c]]"}
	pageB := &Page{Slug: "page-b", RawContent: "Links to [[page-a]]"}
	pageC := &Page{Slug: "page-c", RawContent: "No links here"}

	pages := []*Page{pageA, pageB, pageC}
	BuildBacklinks(pages)

	// page-a should have backlink from page-b
	if len(pageA.Backlinks) != 1 || pageA.Backlinks[0].Slug != "page-b" {
		t.Errorf("page-a backlinks = %v, want [page-b]", slugs(pageA.Backlinks))
	}

	// page-b should have backlink from page-a
	if len(pageB.Backlinks) != 1 || pageB.Backlinks[0].Slug != "page-a" {
		t.Errorf("page-b backlinks = %v, want [page-a]", slugs(pageB.Backlinks))
	}

	// page-c should have backlink from page-a
	if len(pageC.Backlinks) != 1 || pageC.Backlinks[0].Slug != "page-a" {
		t.Errorf("page-c backlinks = %v, want [page-a]", slugs(pageC.Backlinks))
	}
}

func TestBuildBacklinks_NoDuplicates(t *testing.T) {
	pageA := &Page{Slug: "page-a", RawContent: "[[page-b]] and again [[page-b]]"}
	pageB := &Page{Slug: "page-b", RawContent: ""}

	pages := []*Page{pageA, pageB}
	BuildBacklinks(pages)

	if len(pageB.Backlinks) != 1 {
		t.Errorf("page-b should have 1 backlink, got %d", len(pageB.Backlinks))
	}
}

func TestBuildBacklinks_NoSelfLink(t *testing.T) {
	pageA := &Page{Slug: "page-a", RawContent: "Links to itself [[page-a]]"}

	pages := []*Page{pageA}
	BuildBacklinks(pages)

	if len(pageA.Backlinks) != 0 {
		t.Errorf("page-a should have 0 backlinks (no self-links), got %d", len(pageA.Backlinks))
	}
}

func slugs(pages []*Page) []string {
	var s []string
	for _, p := range pages {
		s = append(s, p.Slug)
	}
	return s
}

func TestResolveByPageTitleAlias(t *testing.T) {
	pages := []*Page{
		{Slug: "essays/note-b", Title: "Note B"},
		{Slug: "note-a", Title: "Note A"},
	}
	r := NewLinkResolver(pages)

	result := r.Resolve("Note B")
	if result.Page == nil || result.Page.Slug != "essays/note-b" {
		t.Fatalf("Resolve(\"Note B\") = %+v, want essays/note-b", result)
	}
	// Case- and whitespace-insensitive, like the embedded renderer.
	if got := r.Resolve("note  b"); got.Page == nil || got.Page.Slug != "essays/note-b" {
		t.Fatalf("normalized title resolution failed: %+v", got)
	}
	// Slug and filename matches still win over aliases.
	pages = append(pages, &Page{Slug: "note-b", Title: "Something Else"})
	r = NewLinkResolver(pages)
	if got := r.Resolve("note-b"); got.Page == nil || got.Page.Slug != "note-b" {
		t.Fatalf("filename match should take precedence: %+v", got)
	}
}

func TestAliasCollisionIsOrderIndependent(t *testing.T) {
	// Two pages with the same display title: the smaller slug must win no
	// matter which order they were registered in (CLI scan order and
	// renderer input order can differ).
	forward := NewLinkResolver([]*Page{
		{Slug: "a-note", Title: "Shared Title"},
		{Slug: "z-note", Title: "Shared Title"},
	})
	reverse := NewLinkResolver([]*Page{
		{Slug: "z-note", Title: "Shared Title"},
		{Slug: "a-note", Title: "Shared Title"},
	})
	for name, r := range map[string]*LinkResolver{"forward": forward, "reverse": reverse} {
		got := r.Resolve("Shared Title")
		if got.Page == nil || got.Page.Slug != "a-note" {
			t.Errorf("%s order: Resolve = %+v, want a-note", name, got)
		}
	}
}
