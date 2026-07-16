package templates

import (
	"strings"
	"testing"
)

// --- Heading ID generation ---

func TestGenerateHeadingID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Getting Started", "getting-started"},
		{"with numbers", "Step 1 Setup", "step-1-setup"},
		{"special chars", "Video & Audio", "video-audio"},
		{"multiple spaces", "hello   world", "hello-world"},
		{"leading trailing spaces", "  hello  ", "hello"},
		{"with emoji", "Hello 🌱 World", "hello-world"},
		{"only emoji", "🌱🌿", ""},
		{"with punctuation", "What's New?", "what-s-new"},
		{"hyphens already", "my-page-title", "my-page-title"},
		{"mixed case", "CamelCase Title", "camelcase-title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateHeadingID(tt.input)
			if got != tt.want {
				t.Errorf("generateHeadingID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- TOC extraction ---

func TestExtractTOC_Basic(t *testing.T) {
	input := `<h2>Introduction</h2><p>text</p><h3>Details</h3><p>more</p>`
	html, toc := ExtractTOC(input)

	if len(toc) != 2 {
		t.Fatalf("got %d TOC items, want 2", len(toc))
	}

	if toc[0].Text != "Introduction" || toc[0].Level != 2 {
		t.Errorf("toc[0] = {%q, %d}, want {\"Introduction\", 2}", toc[0].Text, toc[0].Level)
	}
	if toc[1].Text != "Details" || toc[1].Level != 3 {
		t.Errorf("toc[1] = {%q, %d}, want {\"Details\", 3}", toc[1].Text, toc[1].Level)
	}

	// Headings should have IDs
	if !strings.Contains(html, `id="introduction"`) {
		t.Error("h2 should have id=\"introduction\"")
	}
	if !strings.Contains(html, `id="details"`) {
		t.Error("h3 should have id=\"details\"")
	}
}

func TestExtractTOC_AnchorLinks(t *testing.T) {
	input := `<h2>Section</h2>`
	html, _ := ExtractTOC(input)

	if !strings.Contains(html, `class="lp-heading-anchor"`) {
		t.Error("heading should contain anchor link")
	}
	if !strings.Contains(html, `href="#section"`) {
		t.Error("anchor should link to heading ID")
	}
}

func TestExtractTOC_DuplicateHeadings(t *testing.T) {
	input := `<h2>Setup</h2><p>text</p><h2>Setup</h2><p>text</p><h2>Setup</h2>`
	_, toc := ExtractTOC(input)

	if len(toc) != 3 {
		t.Fatalf("got %d TOC items, want 3", len(toc))
	}
	if toc[0].ID != "setup" {
		t.Errorf("toc[0].ID = %q, want \"setup\"", toc[0].ID)
	}
	if toc[1].ID != "setup-1" {
		t.Errorf("toc[1].ID = %q, want \"setup-1\"", toc[1].ID)
	}
	if toc[2].ID != "setup-2" {
		t.Errorf("toc[2].ID = %q, want \"setup-2\"", toc[2].ID)
	}
}

func TestExtractTOC_SpecialChars(t *testing.T) {
	input := `<h2>Video &amp; Audio</h2>`
	html, toc := ExtractTOC(input)

	if len(toc) != 1 {
		t.Fatalf("got %d TOC items, want 1", len(toc))
	}
	if toc[0].ID != "video-audio" {
		t.Errorf("toc[0].ID = %q, want \"video-audio\"", toc[0].ID)
	}
	if toc[0].Text != "Video & Audio" {
		t.Errorf("toc[0].Text = %q, want \"Video & Audio\"", toc[0].Text)
	}

	// Heading ID and TOC href must match
	if !strings.Contains(html, `id="video-audio"`) {
		t.Error("heading should have id=\"video-audio\"")
	}
}

func TestExtractTOC_ExistingID(t *testing.T) {
	input := `<h2 id="old-id">Section</h2>`
	html, toc := ExtractTOC(input)

	// Should replace goldmark's ID with ours
	if !strings.Contains(html, `id="section"`) {
		t.Error("should replace existing ID with generated one")
	}
	if strings.Contains(html, `id="old-id"`) {
		t.Error("old ID should be removed")
	}
	if toc[0].ID != "section" {
		t.Errorf("toc ID should be \"section\", got %q", toc[0].ID)
	}
}

func TestExtractTOC_HTMLInHeading(t *testing.T) {
	input := `<h2><code>func</code> main</h2>`
	_, toc := ExtractTOC(input)

	if len(toc) != 1 {
		t.Fatalf("got %d TOC items, want 1", len(toc))
	}
	if toc[0].Text != "func main" {
		t.Errorf("toc text should strip HTML tags, got %q", toc[0].Text)
	}
}

func TestExtractTOC_NoHeadings(t *testing.T) {
	input := `<p>Just a paragraph.</p>`
	html, toc := ExtractTOC(input)

	if len(toc) != 0 {
		t.Errorf("got %d TOC items, want 0", len(toc))
	}
	if html != input {
		t.Error("HTML should be unchanged when no headings")
	}
}

// --- Font URL ---

func TestCombinedFontURL(t *testing.T) {
	url := combinedFontURL("Crimson Pro", "Inter", "Fira Code")
	if !strings.Contains(url, "family=Crimson+Pro") {
		t.Error("should contain heading font")
	}
	if !strings.Contains(url, "family=Inter") {
		t.Error("should contain body font")
	}
	if !strings.Contains(url, "family=Fira+Code") {
		t.Error("should contain mono font")
	}
	if !strings.Contains(url, "display=swap") {
		t.Error("should include display=swap")
	}
	// Should be a single URL
	if strings.Count(url, "fonts.googleapis.com") != 1 {
		t.Error("should be a single combined URL")
	}
}

func TestCombinedFontURL_Dedup(t *testing.T) {
	url := combinedFontURL("Inter", "Inter", "Fira Code")
	if strings.Count(url, "Inter") != 1 {
		t.Error("duplicate fonts should be deduplicated")
	}
}

func TestCombinedFontURL_AllSame(t *testing.T) {
	url := combinedFontURL("Inter", "Inter", "Inter")
	if strings.Count(url, "family=") != 1 {
		t.Error("all-same fonts should produce single family param")
	}
}
