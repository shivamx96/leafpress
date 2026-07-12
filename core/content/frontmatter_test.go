package content

import (
	"testing"
)

func TestParseFrontmatter_Basic(t *testing.T) {
	input := "---\ntitle: Hello World\ntags: [a, b]\n---\nBody content."
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Title != "Hello World" {
		t.Errorf("Title = %q, want %q", fm.Title, "Hello World")
	}
	if len(fm.Tags) != 2 || fm.Tags[0] != "a" || fm.Tags[1] != "b" {
		t.Errorf("Tags = %v, want [a, b]", fm.Tags)
	}
	if body != "Body content." {
		t.Errorf("Body = %q, want %q", body, "Body content.")
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	input := "Just regular content."
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Title != "" {
		t.Errorf("Title should be empty, got %q", fm.Title)
	}
	if body != input {
		t.Errorf("Body should be original content")
	}
}

func TestParseFrontmatter_Unclosed(t *testing.T) {
	input := "---\ntitle: Hello\nNo closing delimiter"
	_, _, err := ParseFrontmatter(input)
	if err == nil {
		t.Error("should return error for unclosed frontmatter")
	}
}

func TestParseFrontmatter_Draft(t *testing.T) {
	input := "---\ntitle: Draft Post\ndraft: true\n---\nContent."
	fm, _, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fm.Draft {
		t.Error("Draft should be true")
	}
}

func TestParseFrontmatter_Growth(t *testing.T) {
	valid := []string{"seedling", "budding", "evergreen"}
	for _, g := range valid {
		input := "---\ntitle: Test\ngrowth: " + g + "\n---\nBody."
		fm, _, err := ParseFrontmatter(input)
		if err != nil {
			t.Errorf("growth %q should be valid, got error: %v", g, err)
		}
		if fm.Growth != g {
			t.Errorf("Growth = %q, want %q", fm.Growth, g)
		}
	}
}

func TestParseFrontmatter_InvalidGrowth(t *testing.T) {
	input := "---\ntitle: Test\ngrowth: invalid\n---\nBody."
	_, _, err := ParseFrontmatter(input)
	if err == nil {
		t.Error("invalid growth should return error")
	}
}

func TestParseFrontmatter_TOCOverride(t *testing.T) {
	input := "---\ntitle: Test\ntoc: false\n---\nBody."
	fm, _, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.TOC == nil || *fm.TOC != false {
		t.Error("TOC should be false")
	}
}

func TestParseFrontmatter_TOCNil(t *testing.T) {
	input := "---\ntitle: Test\n---\nBody."
	fm, _, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.TOC != nil {
		t.Error("TOC should be nil when not set")
	}
}

// --- Date helpers ---

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"ISO date", "2026-01-15", false},
		{"ISO datetime UTC", "2026-01-15T10:30:00Z", false},
		{"ISO datetime offset", "2026-01-15T10:30:00+05:30", false},
		{"datetime space", "2026-01-15 10:30:00", false},
		{"long month", "January 15, 2026", false},
		{"short month", "Jan 15, 2026", false},
		{"empty", "", false},
		{"invalid", "not-a-date", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestGetCreatedDate_Priority(t *testing.T) {
	// date takes priority
	fm := &Frontmatter{Date: "2026-01-01", Created: "2026-02-02", CreatedAt: "2026-03-03"}
	if got := fm.GetCreatedDate(); got != "2026-01-01" {
		t.Errorf("got %q, want date field", got)
	}

	// created is fallback
	fm = &Frontmatter{Created: "2026-02-02", CreatedAt: "2026-03-03"}
	if got := fm.GetCreatedDate(); got != "2026-02-02" {
		t.Errorf("got %q, want created field", got)
	}

	// createdAt is last resort
	fm = &Frontmatter{CreatedAt: "2026-03-03"}
	if got := fm.GetCreatedDate(); got != "2026-03-03" {
		t.Errorf("got %q, want createdAt field", got)
	}

	// empty
	fm = &Frontmatter{}
	if got := fm.GetCreatedDate(); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestGetModifiedDate_Priority(t *testing.T) {
	fm := &Frontmatter{Modified: "2026-01-01", Updated: "2026-02-02", UpdatedAt: "2026-03-03"}
	if got := fm.GetModifiedDate(); got != "2026-01-01" {
		t.Errorf("got %q, want modified field", got)
	}

	fm = &Frontmatter{Updated: "2026-02-02", UpdatedAt: "2026-03-03"}
	if got := fm.GetModifiedDate(); got != "2026-02-02" {
		t.Errorf("got %q, want updated field", got)
	}

	fm = &Frontmatter{UpdatedAt: "2026-03-03"}
	if got := fm.GetModifiedDate(); got != "2026-03-03" {
		t.Errorf("got %q, want updatedAt field", got)
	}

	fm = &Frontmatter{}
	if got := fm.GetModifiedDate(); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
