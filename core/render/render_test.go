package render

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/templates"
)

// runJSON is a test helper that runs the bridge over a JSON string.
func runJSON(t *testing.T, input string) *Output {
	t.Helper()
	out, err := Run([]byte(input))
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	return out
}

func pageHTML(t *testing.T, out *Output, slug string) string {
	t.Helper()
	for _, p := range out.Pages {
		if p.Slug == slug {
			return p.HTML
		}
	}
	t.Fatalf("page %q not found in output", slug)
	return ""
}

const twoLinkedPages = `{
  "garden": {
    "slug": "shivam",
    "title": "Shivam's Garden",
    "baseUrl": "/g/shivam",
    "sort": "date"
  },
  "pages": [
    {
      "slug": "alpha",
      "title": "Alpha",
      "markdown": "Linking to [[beta|the beta note]] here.",
      "tags": ["systems"],
      "createdAt": "2026-05-16T10:00:00Z",
      "updatedAt": "2026-07-11T09:30:00Z"
    },
    {
      "slug": "beta",
      "title": "Beta",
      "markdown": "# Heading\n\nBeta content.",
      "createdAt": "2026-06-01T10:00:00Z"
    }
  ]
}`

func TestWikilinkCarriesBaseURL(t *testing.T) {
	out := runJSON(t, twoLinkedPages)

	alpha := pageHTML(t, out, "alpha")
	want := `<a class="lp-wikilink" href="/g/shivam/beta/">the beta note</a>`
	if !strings.Contains(alpha, want) {
		t.Errorf("alpha HTML missing wikilink anchor %q", want)
	}
	if !strings.HasPrefix(alpha, "<!DOCTYPE html>") {
		t.Errorf("page HTML is not a full document, starts with %q", alpha[:40])
	}
	// style.css href must carry the base path.
	if !strings.Contains(alpha, `href="/g/shivam/style.css"`) {
		t.Error("page HTML missing base-prefixed style.css link")
	}
	// Tag href must carry the base path.
	if !strings.Contains(alpha, `href="/g/shivam/tags/systems/"`) {
		t.Error("page HTML missing base-prefixed tag href")
	}

	// Backlink on beta points to alpha with the base path.
	beta := pageHTML(t, out, "beta")
	if !strings.Contains(beta, `class="lp-backlink" href="/g/shivam/alpha/"`) {
		t.Error("beta HTML missing base-prefixed backlink to alpha")
	}

	// Index lists both pages with base-prefixed links.
	for _, href := range []string{`href="/g/shivam/alpha/"`, `href="/g/shivam/beta/"`} {
		if !strings.Contains(out.Index, href) {
			t.Errorf("index missing %s", href)
		}
	}
	if !strings.Contains(out.Index, `href="/g/shivam/style.css"`) {
		t.Error("index missing base-prefixed style.css link")
	}
}

func TestBrokenWikilinkDegradesToPlainText(t *testing.T) {
	out := runJSON(t, `{
	  "garden": {"slug": "g"},
	  "pages": [
	    {"slug": "solo", "title": "Solo", "markdown": "See [[Private Note|my secret]] for more."}
	  ]
	}`)

	html := pageHTML(t, out, "solo")
	if !strings.Contains(html, "my secret") {
		t.Error("broken wikilink display text missing from output")
	}
	if strings.Contains(html, "lp-broken-link") {
		t.Error("broken wikilink rendered as styled span; want plain text")
	}
	if strings.Contains(html, "Private Note") {
		t.Error("broken wikilink target leaked into output")
	}
	// No anchor around the display text.
	if regexp.MustCompile(`<a[^>]*>my secret</a>`).MatchString(html) {
		t.Error("broken wikilink rendered as an anchor")
	}

	found := false
	for _, w := range out.Warnings {
		if strings.Contains(w, "broken link") && strings.Contains(w, "Private Note") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected broken-link warning, got %v", out.Warnings)
	}
}

func TestThemeReflectedInOutput(t *testing.T) {
	out := runJSON(t, `{
	  "garden": {
	    "slug": "g",
	    "theme": {
	      "accent": "#ff0000",
	      "fontHeading": "Lora",
	      "background": {"light": "#fafafa", "dark": "#101010"}
	    }
	  },
	  "pages": [{"slug": "p", "title": "P", "markdown": "hello"}]
	}`)

	html := pageHTML(t, out, "p")
	for _, want := range []string{
		"--lp-accent: #ff0000",
		`--lp-font-heading: "Lora", Georgia, serif`,
		"--lp-bg: #fafafa",
		"--lp-bg: #101010",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("page HTML missing theme value %q", want)
		}
	}
	if !strings.Contains(out.Index, "--lp-accent: #ff0000") {
		t.Error("index HTML missing theme accent")
	}
	if out.CSS != templates.DefaultCSS {
		t.Error("css output should be the leafpress default stylesheet")
	}
}

func TestIndexSortOrder(t *testing.T) {
	const pagesJSON = `[
	  {"slug": "b-old", "title": "Zebra", "growth": "evergreen", "markdown": "x", "createdAt": "2026-01-01T00:00:00Z"},
	  {"slug": "a-new", "title": "Apple", "growth": "seedling", "markdown": "x", "createdAt": "2026-03-01T00:00:00Z"},
	  {"slug": "c-mid", "title": "Mango", "growth": "budding", "markdown": "x", "createdAt": "2026-02-01T00:00:00Z"}
	]`

	tests := []struct {
		sort string
		want []string
	}{
		{"date", []string{"/a-new/", "/c-mid/", "/b-old/"}},   // newest first
		{"title", []string{"/a-new/", "/c-mid/", "/b-old/"}},  // Apple, Mango, Zebra
		{"growth", []string{"/a-new/", "/c-mid/", "/b-old/"}}, // seedling, budding, evergreen
		{"", []string{"/a-new/", "/c-mid/", "/b-old/"}},       // default = date
	}

	linkRegex := regexp.MustCompile(`class="lp-index-link" href="([^"]+)"`)
	for _, tt := range tests {
		t.Run("sort="+tt.sort, func(t *testing.T) {
			out := runJSON(t, `{"garden": {"slug": "g", "sort": "`+tt.sort+`"}, "pages": `+pagesJSON+`}`)
			var got []string
			for _, m := range linkRegex.FindAllStringSubmatch(out.Index, -1) {
				got = append(got, m[1])
			}
			if len(got) != len(tt.want) {
				t.Fatalf("index has %d links, want %d: %v", len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("index order[%d] = %q, want %q (full order %v)", i, got[i], tt.want[i], got)
				}
			}
		})
	}
}

func TestDeterministicOutput(t *testing.T) {
	out1, err := Run([]byte(twoLinkedPages))
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	out2, err := Run([]byte(twoLinkedPages))
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	b1, err := json.Marshal(out1)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	b2, err := json.Marshal(out2)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if string(b1) != string(b2) {
		t.Error("identical input did not produce byte-identical output")
	}
}

func TestInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"malformed JSON", `{"garden": {`},
		{"missing garden slug", `{"garden": {}, "pages": []}`},
		{"missing page slug", `{"garden": {"slug": "g"}, "pages": [{"title": "T", "markdown": "x"}]}`},
		{"duplicate page slugs", `{"garden": {"slug": "g"}, "pages": [{"slug": "a", "markdown": "x"}, {"slug": "a", "markdown": "y"}]}`},
		{"bad createdAt", `{"garden": {"slug": "g"}, "pages": [{"slug": "a", "markdown": "x", "createdAt": "yesterday"}]}`},
		{"bad sort", `{"garden": {"slug": "g", "sort": "popularity"}, "pages": []}`},
		{"bad accent", `{"garden": {"slug": "g", "theme": {"accent": "red;}</style>"}}, "pages": []}`},
		{"bad font", `{"garden": {"slug": "g", "theme": {"fontBody": "Inter\"><script>"}}, "pages": []}`},
		{"bad baseUrl", `{"garden": {"slug": "g", "baseUrl": "g/shivam"}, "pages": []}`},
		{"unsafe slug", `{"garden": {"slug": "g"}, "pages": [{"slug": "a\"b", "markdown": "x"}]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Run([]byte(tt.input))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var inputErr *InputError
			if !errors.As(err, &inputErr) {
				t.Errorf("expected *InputError, got %T: %v", err, err)
			}
		})
	}
}

func TestOptionalFieldsDefaulted(t *testing.T) {
	out := runJSON(t, `{"garden": {"slug": "shivam"}, "pages": [{"slug": "my-note", "markdown": "hi"}]}`)

	// Garden title defaults to slug.
	if !strings.Contains(out.Index, "<title>shivam | shivam</title>") {
		t.Error("index title should default to garden slug")
	}
	// Page title defaults to slug; reading time computed.
	html := pageHTML(t, out, "my-note")
	if !strings.Contains(html, `<h1 class="lp-title">my-note</h1>`) {
		t.Error("page title should default to page slug")
	}
	if !strings.Contains(html, "1 min read") {
		t.Error("page should include computed reading time")
	}
	// No base path: root-relative asset link.
	if !strings.Contains(html, `href="/style.css"`) {
		t.Error("page should link /style.css when baseUrl is empty")
	}
	if out.Warnings == nil {
		t.Error("warnings should be an empty slice, not nil")
	}
}
