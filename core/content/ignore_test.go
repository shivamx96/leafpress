package content

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestIgnoreMatcher(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		ignored  []string
		kept     []string
	}{
		{
			name:     "documented example",
			patterns: []string{"drafts/**", "*.draft.md"},
			ignored: []string{
				"drafts", "drafts/a.md", "drafts/deep/b.md",
				"a.draft.md", "notes/b.draft.md",
			},
			kept: []string{"notes/a.md", "draftsy/a.md", "a.md", "notes/draft.md"},
		},
		{
			name:     "bare name matches at any depth",
			patterns: []string{"private"},
			ignored:  []string{"private", "private/a.md", "notes/private/a.md"},
			kept:     []string{"privateer/a.md", "notes/a.md"},
		},
		{
			// A tree walk prunes the directory, but the incremental rebuild
			// asks about one file path at a time and must reach the same
			// verdict.
			name:     "directory pattern hides its contents",
			patterns: []string{"notes/private"},
			ignored:  []string{"notes/private", "notes/private/a.md", "notes/private/deep/b.md"},
			kept:     []string{"notes/a.md", "private/a.md"},
		},
		{
			name:     "pattern with a slash is anchored at the root",
			patterns: []string{"notes/*.wip.md"},
			ignored:  []string{"notes/a.wip.md"},
			kept:     []string{"notes", "other/a.wip.md", "notes/deep/a.wip.md"},
		},
		{
			name:     "double star spans segments",
			patterns: []string{"a/**/tmp.md"},
			ignored:  []string{"a/tmp.md", "a/b/tmp.md", "a/b/c/tmp.md"},
			kept:     []string{"tmp.md", "b/tmp.md"},
		},
		{
			name:     "character class",
			patterns: []string{"draft-[0-9].md"},
			ignored:  []string{"draft-3.md", "notes/draft-7.md"},
			kept:     []string{"draft-x.md"},
		},
		{
			name:     "empty pattern list ignores nothing",
			patterns: nil,
			kept:     []string{"a.md", "drafts/a.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewIgnoreMatcher(tt.patterns)
			if err != nil {
				t.Fatalf("NewIgnoreMatcher: %v", err)
			}
			for _, p := range tt.ignored {
				if !m.Match(filepath.FromSlash(p)) {
					t.Errorf("%q should be ignored by %v", p, tt.patterns)
				}
			}
			for _, p := range tt.kept {
				if m.Match(filepath.FromSlash(p)) {
					t.Errorf("%q should NOT be ignored by %v", p, tt.patterns)
				}
			}
		})
	}
}

// A typo that silently ignores nothing publishes the drafts the pattern was
// written to hold back, so malformed globs are an error.
func TestIgnoreMatcherRejectsMalformedPatterns(t *testing.T) {
	for _, pattern := range []string{"draft-[0-9.md", "a/**b/c"} {
		if _, err := NewIgnoreMatcher([]string{pattern}); err == nil {
			t.Errorf("pattern %q should have been rejected", pattern)
		}
	}
}

func TestIgnoreMatcherSkipsBlankPatterns(t *testing.T) {
	m, err := NewIgnoreMatcher([]string{"", "   ", "./drafts/"})
	if err != nil {
		t.Fatalf("NewIgnoreMatcher: %v", err)
	}
	if !m.Match("drafts") {
		t.Error(`"./drafts/" should normalise to "drafts"`)
	}
	if m.Match("notes/a.md") {
		t.Error("blank patterns must not match everything")
	}
}

func TestScannerAppliesIgnoreGlobs(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		"keep.md",
		"notes/keep.md",
		"drafts/hidden.md",
		"drafts/deep/hidden.md",
		"notes/wip.draft.md",
	}
	for _, rel := range files {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("# x\n\nbody\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	pages, err := NewScanner(dir, []string{"drafts/**", "*.draft.md"}).Scan()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range pages {
		got = append(got, filepath.ToSlash(p.SourcePath))
	}
	sort.Strings(got)

	want := "keep.md notes/keep.md"
	if strings.Join(got, " ") != want {
		t.Errorf("scanned %v, want [%s]", got, want)
	}
}

func TestScannerReportsMalformedIgnorePattern(t *testing.T) {
	_, err := NewScanner(t.TempDir(), []string{"draft-[0-9.md"}).Scan()
	if err == nil {
		t.Fatal("Scan should surface a malformed ignore pattern")
	}
	if !strings.Contains(err.Error(), "ignore pattern") {
		t.Fatalf("error should name the pattern, got: %v", err)
	}
}
