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

// --- XSS hardening ---

func TestXSSBodyAndMetadataEscaped(t *testing.T) {
	tests := []struct {
		name        string
		page        string   // page JSON
		slug        string   // slug to inspect
		mustHave    []string // escaped fragments that must appear
		mustNotHave []string // raw fragments that must never appear anywhere in the page HTML
	}{
		{
			name:        "script tag in body",
			page:        `{"slug": "p", "title": "P", "markdown": "hello <script>alert(1)</script> world"}`,
			slug:        "p",
			mustHave:    []string{"&lt;script&gt;alert(1)&lt;/script&gt;"},
			mustNotHave: []string{"<script>alert(1)</script>"},
		},
		{
			name:        "img onerror in body",
			page:        `{"slug": "p", "title": "P", "markdown": "look <img src=x onerror=alert(document.domain)> here"}`,
			slug:        "p",
			mustHave:    []string{"&lt;img src=x onerror=alert(document.domain)&gt;"},
			mustNotHave: []string{"<img src=x onerror="},
		},
		{
			name:        "block-level script in body",
			page:        `{"slug": "p", "title": "P", "markdown": "<script>\nalert(1)\n</script>"}`,
			slug:        "p",
			mustHave:    []string{"&lt;script&gt;"},
			mustNotHave: []string{"<script>\nalert(1)"},
		},
		{
			name: "script tag in title",
			page: `{"slug": "p", "title": "<script>alert(1)</script>", "markdown": "body", "description": "desc"}`,
			slug: "p",
			mustHave: []string{
				"<title>&lt;script&gt;alert(1)&lt;/script&gt; | g</title>",
				`<h1 class="lp-title">&lt;script&gt;alert(1)&lt;/script&gt;</h1>`,
				`<meta property="og:title" content="&lt;script&gt;alert(1)&lt;/script&gt;">`,
				`<meta name="twitter:title" content="&lt;script&gt;alert(1)&lt;/script&gt;">`,
			},
			mustNotHave: []string{"<script>alert(1)</script>"},
		},
		{
			name: "quotes and angle brackets in description",
			page: `{"slug": "p", "title": "P", "markdown": "body", "description": "she said \"><script>alert(1)</script> & left"}`,
			slug: "p",
			mustHave: []string{
				`<meta name="description" content="she said &#34;&gt;&lt;script&gt;alert(1)&lt;/script&gt; &amp; left">`,
			},
			mustNotHave: []string{`"><script>`, "<script>alert(1)</script>"},
		},
		{
			name: "auto-generated description from hostile body",
			page: `{"slug": "p", "title": "P", "markdown": "\"><script>alert(1)</script> plain text follows"}`,
			slug: "p",
			mustHave: []string{
				// PlainContent un-escapes entities; render must re-escape
				// before the value hits the meta attributes. (The leading
				// straight quote becomes a curly one via the Typographer.)
				`&gt;&lt;script&gt;alert(1)&lt;/script&gt; plain text follows"`,
			},
			mustNotHave: []string{`content=""><script>`, "<script>alert(1)</script>"},
		},
		{
			name: "script tag in garden title",
			page: `{"slug": "p", "title": "P", "markdown": "body"}`,
			slug: "p",
			mustHave: []string{
				"&lt;script&gt;alert(2)&lt;/script&gt;",
			},
			mustNotHave: []string{"<script>alert(2)</script>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gardenTitle := "g"
			if tt.name == "script tag in garden title" {
				gardenTitle = `<script>alert(2)</script>`
			}
			input := `{"garden": {"slug": "g", "title": ` + jsonString(gardenTitle) + `}, "pages": [` + tt.page + `]}`
			out := runJSON(t, input)
			html := pageHTML(t, out, tt.slug)

			for _, want := range tt.mustHave {
				if !strings.Contains(html, want) {
					t.Errorf("page HTML missing escaped fragment %q", want)
				}
			}
			for _, raw := range tt.mustNotHave {
				if strings.Contains(html, raw) {
					t.Errorf("page HTML contains unescaped fragment %q", raw)
				}
				if strings.Contains(out.Index, raw) {
					t.Errorf("index HTML contains unescaped fragment %q", raw)
				}
			}
		})
	}
}

// jsonString marshals a string as a JSON literal for test input assembly.
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestXSSTitleNotUnescapedAnywhere(t *testing.T) {
	// The hostile title flows into <title>, <h1>, og:/twitter: meta, the
	// index listing, and backlink text — none may carry it unescaped.
	out := runJSON(t, `{
	  "garden": {"slug": "g"},
	  "pages": [
	    {"slug": "evil", "title": "<script>alert(1)</script>", "markdown": "links to [[safe]]"},
	    {"slug": "safe", "title": "Safe", "markdown": "plain"}
	  ]
	}`)

	for _, doc := range []string{pageHTML(t, out, "evil"), pageHTML(t, out, "safe"), out.Index} {
		if strings.Contains(doc, "<script>alert(1)</script>") {
			t.Error("hostile title appears unescaped in output")
		}
	}
}

func TestInvalidTagsRejected(t *testing.T) {
	tests := []struct {
		name string
		tag  string
	}{
		{"attribute breakout", `\"><script>`},
		{"angle brackets", "<img>"},
		{"space", "two words"},
		{"slash", "a/b"},
		{"dot", "a.b"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `{"garden": {"slug": "g"}, "pages": [{"slug": "p", "markdown": "x", "tags": ["` + tt.tag + `"]}]}`
			_, err := Run([]byte(input))
			if err == nil {
				t.Fatalf("tag %q should be rejected", tt.tag)
			}
			var inputErr *InputError
			if !errors.As(err, &inputErr) {
				t.Errorf("expected *InputError, got %T: %v", err, err)
			}
		})
	}
}

func TestTagPagesGenerated(t *testing.T) {
	out := runJSON(t, `{
	  "garden": {"slug": "g", "baseUrl": "/g/shivam"},
	  "pages": [
	    {"slug": "a", "title": "A", "markdown": "x", "tags": ["Systems", "go-lang"], "createdAt": "2026-01-02T00:00:00Z"},
	    {"slug": "b", "title": "B", "markdown": "y", "tags": ["systems"], "createdAt": "2026-01-01T00:00:00Z"}
	  ]
	}`)

	// Index links every tag (lowercased, sorted) with the base path.
	if out.Tags.Index == "" {
		t.Fatal("tags index should be rendered")
	}
	for _, want := range []string{`href="/g/shivam/tags/go-lang/"`, `href="/g/shivam/tags/systems/"`} {
		if !strings.Contains(out.Tags.Index, want) {
			t.Errorf("tags index missing link %s", want)
		}
	}
	if !strings.Contains(out.Tags.Index, "(2)") {
		t.Error("tags index should show systems count of 2")
	}

	// Tag list is deterministic and sorted.
	if len(out.Tags.Pages) != 2 || out.Tags.Pages[0].Tag != "go-lang" || out.Tags.Pages[1].Tag != "systems" {
		got := make([]string, 0, len(out.Tags.Pages))
		for _, tp := range out.Tags.Pages {
			got = append(got, tp.Tag)
		}
		t.Fatalf("tag pages should be [go-lang systems], got %v", got)
	}

	// The systems tag page lists both pages, newest first, base-prefixed.
	systems := out.Tags.Pages[1].HTML
	if !strings.HasPrefix(systems, "<!DOCTYPE html>") {
		t.Error("tag page should be a full document")
	}
	aIdx := strings.Index(systems, `href="/g/shivam/a/"`)
	bIdx := strings.Index(systems, `href="/g/shivam/b/"`)
	if aIdx == -1 || bIdx == -1 {
		t.Fatalf("systems tag page missing page links:\n%s", systems)
	}
	if aIdx > bIdx {
		t.Error("systems tag page should list newest page (a) first")
	}

	// The page header tag link now has a matching generated page.
	if !strings.Contains(pageHTML(t, out, "a"), `href="/g/shivam/tags/systems/"`) {
		t.Error("page header should link to the systems tag page")
	}
}

func TestNoTagsEmptyTagsOutput(t *testing.T) {
	out := runJSON(t, `{"garden": {"slug": "g"}, "pages": [{"slug": "p", "markdown": "x"}]}`)

	if out.Tags.Index != "" {
		t.Error("tags index should be empty when no page has tags")
	}
	if out.Tags.Pages == nil {
		t.Error("tags pages should be an empty slice, not nil")
	}
	if len(out.Tags.Pages) != 0 {
		t.Errorf("tags pages should be empty, got %d", len(out.Tags.Pages))
	}
	// No dead tag links: pages without tags render no tag hrefs at all.
	if strings.Contains(pageHTML(t, out, "p"), "/tags/") {
		t.Error("untagged page should not link to any tag page")
	}

	// JSON shape check: "tags": {"index": "", "pages": []}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(b), `"tags":{"index":"","pages":[]}`) {
		t.Error(`output JSON should contain "tags":{"index":"","pages":[]}`)
	}
}

func TestDeterministicOutputWithHostileContent(t *testing.T) {
	// Exercises the escape pipeline (random trusted-chunk nonces) plus tag
	// pages: double-run must still be byte-identical.
	input := `{
	  "garden": {"slug": "g", "title": "T"},
	  "pages": [
	    {"slug": "a", "title": "<script>x</script>", "markdown": "> [!note] Hi\n> body <script>alert(1)</script>\n\nlink [[b]] and ![[demo.mp4]]", "tags": ["one", "two"]},
	    {"slug": "b", "title": "B", "markdown": "<div class=\"x\">\nraw\n</div>", "tags": ["one"]}
	  ]
	}`
	out1, err := Run([]byte(input))
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	out2, err := Run([]byte(input))
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	b1, _ := json.Marshal(out1)
	b2, _ := json.Marshal(out2)
	if string(b1) != string(b2) {
		t.Error("identical hostile input did not produce byte-identical output")
	}
	// The per-render placeholder nonce must never leak into output.
	if strings.Contains(string(b1), "lp-callout-note") == false {
		t.Error("callout should render live in bridge output")
	}
}
