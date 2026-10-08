package content

import (
	"strings"
	"testing"
)

func TestValidateOutputRoutes(t *testing.T) {
	tests := []struct {
		name  string
		pages []*Page
		want  string
	}{
		{
			name: "unique routes",
			pages: []*Page{
				{SourcePath: "notes/_index.md", Slug: "notes", IsIndex: true},
				{SourcePath: "notes/one.md", Slug: "notes/one", Tags: []string{"go"}},
			},
		},
		{
			name: "duplicate pages",
			pages: []*Page{
				{SourcePath: "notes.md", Slug: "notes"},
				{SourcePath: "notes/_index.md", Slug: "notes", IsIndex: true},
			},
			want: `output route "/notes/" is claimed by both page "notes.md" and page "notes/_index.md"`,
		},
		{
			name: "page and generated section",
			pages: []*Page{
				{SourcePath: "notes.md", Slug: "notes"},
				{SourcePath: "notes/one.md", Slug: "notes/one"},
			},
			want: `output route "/notes/" is claimed by both page "notes.md" and generated section "notes"`,
		},
		{
			name: "page and tag index",
			pages: []*Page{
				{SourcePath: "tags.md", Slug: "tags"},
				{SourcePath: "one.md", Slug: "one", Tags: []string{"go"}},
			},
			want: `output route "/tags/"`,
		},
		{
			name: "content below reserved tag route",
			pages: []*Page{
				{SourcePath: "tags/go.md", Slug: "tags/go"},
				{SourcePath: "one.md", Slug: "one", Tags: []string{"Go"}},
			},
			want: `output route "/tags/"`,
		},
		{
			name:  "unsafe tag path",
			pages: []*Page{{SourcePath: "one.md", Slug: "one", Tags: []string{"../outside"}}},
			want:  `invalid tag "../outside"`,
		},
		{
			name:  "page named after the global feed",
			pages: []*Page{{SourcePath: "feed.xml.md", Slug: "feed.xml"}},
			want:  `page "feed.xml.md" publishes at /feed.xml/, which is where leafpress writes the generated file feed.xml; rename the file`,
		},
		{
			name:  "page named after another root artifact",
			pages: []*Page{{SourcePath: "Sitemap.XML.md", Slug: "Sitemap.XML"}},
			want:  `generated file sitemap.xml`,
		},
		{
			name: "page named after a section feed",
			pages: []*Page{
				{SourcePath: "posts/hello.md", Slug: "posts/hello"},
				{SourcePath: "posts/feed.xml.md", Slug: "posts/feed.xml"},
			},
			want: `page "posts/feed.xml.md" publishes at /posts/feed.xml/, which is where leafpress writes the generated file posts/feed.xml`,
		},
		{
			name: "page named after an explicit section's feed",
			pages: []*Page{
				{SourcePath: "posts/_index.md", Slug: "posts", IsIndex: true},
				{SourcePath: "posts/feed.xml.md", Slug: "posts/feed.xml"},
			},
			want: `generated file posts/feed.xml`,
		},
		{
			name: "content below a section feed",
			pages: []*Page{
				{SourcePath: "posts/hello.md", Slug: "posts/hello"},
				{SourcePath: "posts/feed.xml/deep.md", Slug: "posts/feed.xml/deep"},
			},
			want: `generated file posts/feed.xml`,
		},
		{
			name: "page named after a tag feed",
			pages: []*Page{
				{SourcePath: "one.md", Slug: "one", Tags: []string{"Go"}},
				{SourcePath: "tags/go/feed.xml.md", Slug: "tags/go/feed.xml"},
			},
			want: `output route "/tags/go/"`, // the tag route itself is reserved first
		},
		{
			name: "page named after a page's index.html",
			pages: []*Page{
				{SourcePath: "notes.md", Slug: "notes"},
				{SourcePath: "notes/index.html.md", Slug: "notes/index.html"},
			},
			want: `output route "/notes/"`, // notes/ is already a section-vs-page clash
		},
		{
			name: "page below a section's index.html",
			pages: []*Page{
				{SourcePath: "notes/_index.md", Slug: "notes", IsIndex: true},
				{SourcePath: "notes/index.html.md", Slug: "notes/index.html"},
			},
			want: `page "notes/index.html.md" publishes at /notes/index.html/, which is where leafpress writes the generated file notes/index.html`,
		},
		{
			name: "dotted slugs that do not name artifacts stay valid",
			pages: []*Page{
				{SourcePath: "release/_index.md", Slug: "release", IsIndex: true},
				{SourcePath: "release/v1.2.md", Slug: "release/v1.2", Tags: []string{"go"}},
				{SourcePath: "notes/feed.md", Slug: "notes/feed"},
				{SourcePath: "notes/feed.xml.old.md", Slug: "notes/feed.xml.old"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutputRoutes(tt.pages)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("ValidateOutputRoutes: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ValidateOutputRoutes error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestValidateSlug(t *testing.T) {
	for _, slug := range []string{"", "notes/hello-world", "éclair/日本語", "notes/e\u0301clair", "release/v1.2", "com10"} {
		if err := ValidateSlug(slug); err != nil {
			t.Errorf("ValidateSlug(%q): %v", slug, err)
		}
	}
	for _, slug := range []string{"has#fragment", "query?x", "encoded%2fpath", "a&b", "a b", "a\u00a0b", "a\x00b", "a\x7fb", "a\xff", `a\b`, `a"b`, "a'b", "<a>", "a:b", "a*b", "a|b", "../x", "a/./b", "a//b", "/a", "a/", "a.", "CON", "aux.txt", "notes/LPT1", "COM¹", "CONOUT$"} {
		if err := ValidateSlug(slug); err == nil {
			t.Errorf("ValidateSlug(%q) succeeded", slug)
		}
	}
}

func TestOutputRoutesRejectUnsafePathsAndCaseCollisions(t *testing.T) {
	for _, pages := range [][]*Page{
		{{SourcePath: "has#fragment.md", Slug: "has#fragment"}},
		{{SourcePath: "a.md", Slug: "a", Tags: []string{"CON"}}},
		{{SourcePath: "Notes.md", Slug: "Notes"}, {SourcePath: "notes.md", Slug: "notes"}},
		{{SourcePath: "Notes.md", Slug: "Notes"}, {SourcePath: "notes/child.md", Slug: "notes/child"}},
	} {
		if err := ValidateOutputRoutes(pages); err == nil {
			t.Errorf("ValidateOutputRoutes(%+v) succeeded", pages)
		}
	}
}

func TestCleanedSlugsPassRouteValidation(t *testing.T) {
	for _, name := range []string{"My Note", "Q&A", "has#fragment", "a?b", "Wait...", `quote"and'apostrophe`, "tab\tname", "a:b*c|d", "<tag>", "éclair notes"} {
		if err := ValidateSlug(SlugSegment(name)); err != nil {
			t.Errorf("ValidateSlug(SlugSegment(%q) = %q): %v", name, SlugSegment(name), err)
		}
	}
}

func TestReservedFilenameErrorSuggestsSlug(t *testing.T) {
	err := ValidateOutputRoutes([]*Page{{SourcePath: "notes/con.md", Slug: "notes/con"}})
	if err == nil || !strings.Contains(err.Error(), `page "notes/con.md"`) || !strings.Contains(err.Error(), "slug in the page's frontmatter") {
		t.Fatalf("error = %v", err)
	}
}

func TestFoldersCleaningToOneSectionAreRejected(t *testing.T) {
	err := ValidateOutputRoutes([]*Page{
		{SourcePath: "Field Notes/rain.md", Slug: "Field-Notes/rain"},
		{SourcePath: "Field-Notes/sun.md", Slug: "Field-Notes/sun"},
	})
	if err == nil || !strings.Contains(err.Error(), `"Field Notes"`) || !strings.Contains(err.Error(), `"Field-Notes"`) || !strings.Contains(err.Error(), "/Field-Notes/") {
		t.Fatalf("error = %v", err)
	}
	// Pages in the same folder, including its _index.md, are one section.
	if err := ValidateOutputRoutes([]*Page{
		{SourcePath: "Q&A/_index.md", Slug: "Q-A", IsIndex: true},
		{SourcePath: "Q&A/one.md", Slug: "Q-A/one"},
		{SourcePath: "Q&A/Deep Dive/two.md", Slug: "Q-A/Deep-Dive/two"},
	}); err != nil {
		t.Fatalf("one folder rejected: %v", err)
	}
}
