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
