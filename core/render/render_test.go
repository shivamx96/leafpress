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

func artifact(t *testing.T, out *Output, path string) OutputArtifact {
	t.Helper()
	for _, item := range out.Artifacts {
		if item.Path == path {
			return item
		}
	}
	t.Fatalf("artifact %q not found in output", path)
	return OutputArtifact{}
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

func TestCanonicalSiteConfigAndStyleMatchLeafpressSemantics(t *testing.T) {
	out := runJSON(t, `{
	  "garden": {"slug":"hosted", "baseUrl":"/legacy-hosted"},
	  "config": {
	    "title":"Configured Garden",
	    "description":"Site description",
	    "author":"Garden Author",
	    "baseURL":"https://example.com/notes",
	    "image":"/og-default.png",
	    "outputDir":"ignored-by-renderer",
	    "port":4444,
	    "nav":[{"label":"Start Here","path":"/alpha/"}],
	    "theme":{
	      "fontHeading":"Fraunces","fontBody":"Atkinson Hyperlegible",
	      "fontMono":"IBM Plex Mono","accent":"#123456",
	      "background":{"light":"#fafafa","dark":"#101010"},
	      "navStyle":"sticky","navActiveStyle":"underlined"
	    },
	    "graph":true,"search":true,"toc":false,"backlinks":true,
	    "wikilinks":true,"rss":true,
	    "ignore":["private/**"],
	    "headExtra":"<meta name=\"configured\" content=\"yes\">",
	    "deploy":{"provider":"netlify","settings":{"site":"demo"}}
	  },
	  "styleCSS":".custom-rule { color: rebeccapurple; }",
	  "pages":[
	    {"slug":"alpha","title":"Alpha","markdown":"## Heading\n\nSee [[beta]].","createdAt":"2026-01-01T00:00:00Z"},
	    {"slug":"beta","title":"Beta","markdown":"Beta body.","createdAt":"2026-01-02T00:00:00Z"}
	  ]
	}`)

	alpha := pageHTML(t, out, "alpha")
	for _, want := range []string{
		"Configured Garden",
		"Garden Author",
		`href="/notes/alpha/">Start Here</a>`,
		`href="https://example.com/notes/alpha/"`,
		`content="https://example.com/notes/og-default.png"`,
		`--lp-font-heading: "Fraunces"`,
		`--lp-accent: #123456`,
		`--lp-bg: #fafafa`,
		`lp-nav-active-underlined`,
		`name="configured" content="yes"`,
		`class="lp-graph-toggle"`,
		`class="lp-search-toggle"`,
		`href="/notes/feed.xml"`,
		`class="lp-wikilink" href="/notes/beta/"`,
	} {
		if !strings.Contains(alpha, want) {
			t.Errorf("configured HTML missing %q", want)
		}
	}
	if strings.Contains(alpha, `class="lp-toc"`) {
		t.Error("site toc=false should suppress a page TOC without an override")
	}
	if strings.Contains(alpha, `href="/notes/beta/">Beta</a></div>`) {
		t.Error("canonical nav should not be replaced by derived root-note nav")
	}
	if !strings.Contains(out.CSS, "/* User Styles */") ||
		!strings.Contains(out.CSS, ".custom-rule { color: rebeccapurple; }") {
		t.Error("styleCSS should append exactly like the CLI's style.css")
	}

	graph := artifact(t, out, "graph.json")
	if graph.ContentType != "application/json" ||
		!strings.Contains(graph.Content, `"source": "alpha"`) ||
		!strings.Contains(graph.Content, `"target": "beta"`) {
		t.Errorf("graph artifact missing resolved edge: %s", graph.Content)
	}
	search := artifact(t, out, "search-index.json")
	if !strings.Contains(search.Content, `"url": "/notes/alpha/"`) ||
		!strings.Contains(search.Content, `"content": "Heading See beta ."`) {
		t.Errorf("search artifact has unexpected content: %s", search.Content)
	}
	if feed := artifact(t, out, "feed.xml"); !strings.Contains(feed.Content, "<title>Configured Garden</title>") ||
		!strings.Contains(feed.Content, "https://example.com/notes/feed.xml") {
		t.Errorf("feed artifact missing canonical site config: %s", feed.Content)
	}
	if sitemap := artifact(t, out, "sitemap.xml"); !strings.Contains(sitemap.Content, "https://example.com/notes/alpha/") {
		t.Errorf("sitemap artifact missing canonical URL: %s", sitemap.Content)
	}
	if robots := artifact(t, out, "robots.txt"); !strings.Contains(robots.Content, "https://example.com/notes/sitemap.xml") {
		t.Errorf("robots artifact missing canonical sitemap: %s", robots.Content)
	}
	if notFound := artifact(t, out, "404.html"); !strings.Contains(notFound.Content, "Configured Garden") {
		t.Error("404 artifact should use configured site templates")
	}
}

func TestCanonicalConfigDefaultsAndFeatureDisables(t *testing.T) {
	// Empty canonical config uses exactly the CLI defaults, including enabled
	// graph/search/TOC/backlinks/wikilinks/RSS and an intentionally empty nav.
	defaults := runJSON(t, `{
	  "garden":{"slug":"g","baseUrl":"/g/g"},
	  "config":{},
	  "pages":[
	    {"slug":"one","title":"One","markdown":"## Heading\n\n[[two]]"},
	    {"slug":"two","title":"Two","markdown":"body"}
	  ]
	}`)
	one := pageHTML(t, defaults, "one")
	for _, want := range []string{
		"My Garden", `class="lp-toc"`, `class="lp-backlink"`,
		`class="lp-wikilink"`, `class="lp-graph-toggle"`, `class="lp-search-toggle"`,
	} {
		combined := one + pageHTML(t, defaults, "two")
		if !strings.Contains(combined, want) {
			t.Errorf("canonical defaults missing %q", want)
		}
	}
	artifact(t, defaults, "graph.json")
	artifact(t, defaults, "search-index.json")
	artifact(t, defaults, "feed.xml")
	if strings.Contains(one, `class="lp-nav-link"`) {
		t.Error("an empty canonical nav must stay empty")
	}

	disabled := runJSON(t, `{
	  "garden":{"slug":"g"},
	  "config":{"graph":false,"search":false,"toc":false,"backlinks":false,"wikilinks":false,"rss":false},
	  "pages":[
	    {"slug":"one","title":"One","markdown":"## Heading\n\n[[two]]"},
	    {"slug":"two","title":"Two","markdown":"body"}
	  ]
	}`)
	combined := pageHTML(t, disabled, "one") + pageHTML(t, disabled, "two")
	for _, absent := range []string{
		`class="lp-toc"`, `class="lp-backlink"`, `class="lp-wikilink"`,
		`class="lp-graph-toggle"`, `class="lp-search-toggle"`, `feed.xml`,
	} {
		if strings.Contains(combined, absent) {
			t.Errorf("disabled feature still emitted %q", absent)
		}
	}
	for _, item := range disabled.Artifacts {
		if item.Path == "graph.json" || item.Path == "search-index.json" || item.Path == "feed.xml" {
			t.Errorf("disabled artifact still emitted: %s", item.Path)
		}
	}
}

func TestCanonicalConfigRejectsInvalidCLIValues(t *testing.T) {
	_, err := Run([]byte(`{
	  "garden":{"slug":"g"},
	  "config":{"theme":{"accent":"red"}},
	  "pages":[]
	}`))
	if err == nil || !strings.Contains(err.Error(), "accent color") {
		t.Fatalf("expected canonical config validation error, got %v", err)
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

// leafpad markdown links by display title ([[Beta Note]]), while public
// slugs are hyphenated (beta-note). Titles register as resolver aliases.
func TestWikilinkResolvesByPageTitle(t *testing.T) {
	out := runJSON(t, `{
  "garden": {"slug": "shivam", "baseUrl": "/g/shivam"},
  "pages": [
    {
      "slug": "alpha-note",
      "title": "Alpha Note",
      "markdown": "See [[Beta Note]] and [[beta   NOTE|the beta one]] and [[Gamma Note]]."
    },
    {"slug": "beta-note", "title": "Beta Note", "markdown": "Beta content."}
  ]
}`)

	alpha := pageHTML(t, out, "alpha-note")
	if !strings.Contains(alpha, `<a class="lp-wikilink" href="/g/shivam/beta-note/">Beta Note</a>`) {
		t.Errorf("title-form wikilink did not resolve to the page slug:\n%s", alpha)
	}
	// Case- and whitespace-insensitive, alias label preserved.
	if !strings.Contains(alpha, `<a class="lp-wikilink" href="/g/shivam/beta-note/">the beta one</a>`) {
		t.Error("normalized title-form wikilink with label did not resolve")
	}
	// Unpublished title degrades to plain text with no href leak.
	if strings.Contains(alpha, "gamma") || strings.Contains(alpha, "<a class=\"lp-wikilink\" href=\"/g/shivam/gamma") {
		t.Error("unresolved title leaked a URL")
	}
	if !strings.Contains(alpha, "Gamma Note") {
		t.Error("unresolved title should render as plain display text")
	}
	// Backlink from title-form link lands on the target page.
	beta := pageHTML(t, out, "beta-note")
	if !strings.Contains(beta, `class="lp-backlink" href="/g/shivam/alpha-note/"`) {
		t.Error("title-form wikilink did not produce a backlink")
	}
}

// ---- Sections (folder-path slugs, index pages, auto-indexes) ----

const sectionedGarden = `{
  "garden": {"slug": "shivam", "title": "Shivam's Garden", "baseUrl": "/g/shivam"},
  "pages": [
    {"slug": "hello", "title": "Hello", "markdown": "Root note.", "createdAt": "2026-01-01T00:00:00Z"},
    {"slug": "essays/first", "title": "First Essay", "markdown": "One.", "createdAt": "2026-02-01T00:00:00Z"},
    {"slug": "essays/second", "title": "Second Essay", "markdown": "Two.", "createdAt": "2026-03-01T00:00:00Z"},
    {"slug": "essays", "title": "Essays", "markdown": "Long-form writing.", "isIndex": true},
    {"slug": "recipes/dal", "title": "Dal", "markdown": "Cook it.", "createdAt": "2026-04-01T00:00:00Z"}
  ]
}`

func TestSectionHomeRendersIntroAndChildren(t *testing.T) {
	out := runJSON(t, sectionedGarden)

	home := pageHTML(t, out, "essays")
	if !strings.Contains(home, "Long-form writing.") {
		t.Error("section home should contain the index page's markdown as intro")
	}
	for _, href := range []string{`href="/g/shivam/essays/first/"`, `href="/g/shivam/essays/second/"`} {
		if !strings.Contains(home, href) {
			t.Errorf("section home should list child %s", href)
		}
	}
	if strings.Contains(home, `class="lp-index-link" href="/g/shivam/hello/"`) {
		t.Error("section home should not list pages outside the section")
	}
	// Newest first (date sort default).
	if strings.Index(home, "essays/second") > strings.Index(home, "essays/first") {
		t.Error("section children should sort newest first by default")
	}
}

func TestSectionWithoutIndexGetsAutoHome(t *testing.T) {
	out := runJSON(t, sectionedGarden)

	if len(out.Sections) != 1 {
		t.Fatalf("expected 1 auto section, got %d", len(out.Sections))
	}
	auto := out.Sections[0]
	if auto.Slug != "recipes" {
		t.Errorf("auto section slug = %q, want recipes", auto.Slug)
	}
	if !strings.Contains(auto.HTML, "Recipes") {
		t.Error("auto section home should be titled with the title-cased folder name")
	}
	if !strings.Contains(auto.HTML, `href="/g/shivam/recipes/dal/"`) {
		t.Error("auto section home should list its child pages")
	}
}

func TestNavigationContainsRootNotesAndSectionsOnly(t *testing.T) {
	out := runJSON(t, sectionedGarden)
	nested := pageHTML(t, out, "essays/first")

	for _, want := range []string{
		`class="lp-nav-link" href="/g/shivam/hello/">Hello</a>`,
		`class="lp-nav-link lp-nav-link--active lp-nav-active-base" href="/g/shivam/essays/">Essays</a>`,
		`class="lp-nav-link" href="/g/shivam/recipes/">Recipes</a>`,
	} {
		if !strings.Contains(nested, want) {
			t.Errorf("navigation missing %s", want)
		}
	}
	for _, nestedSlug := range []string{"essays/first", "essays/second", "recipes/dal"} {
		if strings.Contains(nested, `class="lp-nav-link" href="/g/shivam/`+nestedSlug+`/"`) {
			t.Errorf("nested note %q should not be a nav item", nestedSlug)
		}
	}
}

func TestHostedTagsNavigationIsOptInAndRequiresTags(t *testing.T) {
	input := &Input{
		Garden: Garden{
			Slug:          "garden",
			Title:         "Garden",
			ShowTagsInNav: true,
		},
		Pages: []InputPage{{
			Slug:     "tagged",
			Title:    "Tagged",
			Markdown: "Body",
			Tags:     []string{"ideas"},
		}},
	}

	out, err := Render(input)
	if err != nil {
		t.Fatal(err)
	}
	html := pageHTML(t, out, "tagged")
	if !strings.Contains(html, `class="lp-nav-link" href="/tags/">Tags</a>`) {
		t.Error("enabled tags navigation should link the generated tags index")
	}

	input.Garden.ShowTagsInNav = false
	out, err = Render(input)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(pageHTML(t, out, "tagged"), `href="/tags/"`) {
		t.Error("tags navigation should be opt-in")
	}

	input.Garden.ShowTagsInNav = true
	input.Pages[0].Tags = nil
	out, err = Render(input)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(pageHTML(t, out, "tagged"), `href="/tags/"`) {
		t.Error("tags navigation should stay absent when no tag index exists")
	}
}

func TestPageFrontmatterConfigMatchesNativeRendering(t *testing.T) {
	showTOC := false
	readingTime := 7
	out, err := Render(&Input{
		Garden: Garden{Slug: "g", Title: "Garden"},
		Pages: []InputPage{{
			Slug:        "configured",
			Title:       "Configured",
			Markdown:    "## Hidden heading\n\nBody.",
			Description: "A custom description",
			Growth:      "evergreen",
			TOC:         &showTOC,
			Image:       "/images/card.png",
			ReadingTime: &readingTime,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	html := pageHTML(t, out, "configured")
	for _, want := range []string{
		`content="A custom description"`,
		`content="/images/card.png"`,
		`lp-growth--evergreen`,
		`7 min read`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("configured page missing %q", want)
		}
	}
	if strings.Contains(html, `class="lp-toc"`) {
		t.Error("toc=false should suppress the table of contents")
	}
}

func TestPageReadingTimeRejectsNonPositiveOverride(t *testing.T) {
	zero := 0
	_, err := Render(&Input{
		Garden: Garden{Slug: "g"},
		Pages:  []InputPage{{Slug: "bad", ReadingTime: &zero}},
	})
	if err == nil || !strings.Contains(err.Error(), "readingTime") {
		t.Fatalf("expected readingTime validation error, got %v", err)
	}
}

func TestRootIndexReplacesGardenHome(t *testing.T) {
	out := runJSON(t, `{
	  "garden": {"slug": "g", "title": "My Garden"},
	  "pages": [
	    {"slug": "", "title": "", "markdown": "Welcome to my garden.", "isIndex": true},
	    {"slug": "note", "title": "Note", "markdown": "n."},
	    {"slug": "essays/one", "title": "One", "markdown": "o."}
	  ]
	}`)

	if !strings.Contains(out.Index, "Welcome to my garden.") {
		t.Error("home should contain the root index page's intro")
	}
	// Root-index home lists root-level pages and one entry for each section.
	if !strings.Contains(out.Index, `href="/note/"`) {
		t.Error("home should list root-level pages")
	}
	if strings.Contains(out.Index, `href="/essays/one/"`) {
		t.Error("root-index home should not flatten nested pages into the listing")
	}
	if !strings.Contains(out.Index, `href="/essays/"`) {
		t.Error("home should link to the section containing nested pages")
	}
	// Empty root-index title falls back to the garden title.
	if !strings.Contains(out.Index, "My Garden") {
		t.Error("home title should fall back to the garden title")
	}
	// The root index page renders as the home, not as a page artifact.
	for _, p := range out.Pages {
		if p.Slug == "" {
			t.Error("root index page should not appear in pages output")
		}
	}
}

func TestSyntheticHomeListsSectionsInsteadOfNestedPages(t *testing.T) {
	out := runJSON(t, sectionedGarden)

	if !strings.Contains(out.Index, `href="/g/shivam/hello/"`) {
		t.Error("synthetic home should list root-level pages")
	}
	for _, href := range []string{`href="/g/shivam/essays/"`, `href="/g/shivam/recipes/"`} {
		if !strings.Contains(out.Index, href) {
			t.Errorf("synthetic home should link to section %s", href)
		}
	}
	for _, href := range []string{`href="/g/shivam/essays/first/"`, `href="/g/shivam/essays/second/"`, `href="/g/shivam/recipes/dal/"`} {
		if strings.Contains(out.Index, href) {
			t.Errorf("synthetic home should not flatten nested page %s", href)
		}
	}
}

func TestSectionSortAndShowList(t *testing.T) {
	out := runJSON(t, `{
	  "garden": {"slug": "g"},
	  "pages": [
	    {"slug": "s/b", "title": "Banana", "markdown": "x", "createdAt": "2026-03-01T00:00:00Z"},
	    {"slug": "s/a", "title": "Apple", "markdown": "x", "createdAt": "2026-01-01T00:00:00Z"},
	    {"slug": "s", "title": "S", "markdown": "intro", "isIndex": true, "sort": "title"},
	    {"slug": "hidden/x", "title": "X", "markdown": "x"},
	    {"slug": "hidden", "title": "Hidden", "markdown": "no list here", "isIndex": true, "showList": false}
	  ]
	}`)

	s := pageHTML(t, out, "s")
	if strings.Index(s, "/s/a/") > strings.Index(s, "/s/b/") {
		t.Error("section with sort=title should list Apple before Banana")
	}
	hidden := pageHTML(t, out, "hidden")
	if strings.Contains(hidden, `href="/hidden/x/"`) {
		t.Error("showList=false should suppress the child listing")
	}
	if !strings.Contains(hidden, "no list here") {
		t.Error("showList=false should still render the intro")
	}
}

func TestWikilinkResolvesToNestedSlug(t *testing.T) {
	out := runJSON(t, `{
	  "garden": {"slug": "g", "baseUrl": "/g/me"},
	  "pages": [
	    {"slug": "essays/deep-note", "title": "Deep Note", "markdown": "x"},
	    {"slug": "top", "title": "Top", "markdown": "See [[Deep Note]]."}
	  ]
	}`)

	top := pageHTML(t, out, "top")
	if !strings.Contains(top, `href="/g/me/essays/deep-note/"`) {
		t.Error("wikilink should resolve to the nested permalink")
	}
}

func TestSectionInputValidation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty slug without isIndex", `{"garden": {"slug": "g"}, "pages": [{"slug": "", "markdown": "x"}]}`},
		{"duplicate root index", `{"garden": {"slug": "g"}, "pages": [{"slug": "", "markdown": "x", "isIndex": true}, {"slug": "/", "markdown": "y", "isIndex": true}]}`},
		{"bad section sort", `{"garden": {"slug": "g"}, "pages": [{"slug": "s", "markdown": "x", "isIndex": true, "sort": "random"}]}`},
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

func TestUntitledIndexPageTitledFromSlug(t *testing.T) {
	out := runJSON(t, `{
	  "garden": {"slug": "g"},
	  "pages": [
	    {"slug": "long-form-essays/one", "title": "One", "markdown": "x"},
	    {"slug": "long-form-essays", "title": "", "markdown": "intro", "isIndex": true}
	  ]
	}`)

	home := pageHTML(t, out, "long-form-essays")
	if !strings.Contains(home, "Long Form Essays") {
		t.Error("untitled index page should be titled from its slug segment")
	}
}

func TestAutoSectionTitleReadsHyphensAsSpaces(t *testing.T) {
	out := runJSON(t, `{
	  "garden": {"slug": "g"},
	  "pages": [{"slug": "field-notes/one", "title": "One", "markdown": "x"}]
	}`)
	if len(out.Sections) != 1 || !strings.Contains(out.Sections[0].HTML, "Field Notes") {
		t.Errorf("auto section should be titled 'Field Notes', got: %v", out.Sections)
	}
}
