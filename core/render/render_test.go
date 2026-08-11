package render

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/assets"
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

func clientScriptArtifact(t *testing.T, out *Output) OutputArtifact {
	t.Helper()
	var matches []OutputArtifact
	for _, item := range out.Artifacts {
		if strings.HasPrefix(item.Path, "static/leafpress/app.") && strings.HasSuffix(item.Path, ".js") {
			matches = append(matches, item)
		}
	}
	if len(matches) != 1 {
		t.Fatalf("client script artifacts = %d, want exactly one", len(matches))
	}
	item := matches[0]
	wantPath := "static/leafpress/app." + assets.Sum([]byte(item.Content))[:32] + ".js"
	if item.Path != wantPath {
		t.Fatalf("client script path = %q, want content-addressed path %q", item.Path, wantPath)
	}
	if item.ContentType != "text/javascript; charset=utf-8" || item.Encoding != "utf8" {
		t.Fatalf("client script metadata = (%q, %q), want JavaScript utf8", item.ContentType, item.Encoding)
	}
	return item
}

func TestRunSharesOneClientScriptAcrossRenderedDocuments(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "content": {"pages": [
	    {"slug": "one", "title": "One", "markdown": "one", "tags": ["shared"]},
	    {"slug": "notes/two", "title": "Two", "markdown": "two", "tags": ["shared"]}
	  ]}
	}`)
	client := clientScriptArtifact(t, out)
	scriptTag := `<script src="/` + client.Path + `" defer></script>`
	if len(out.Tags.Pages) != 1 {
		t.Fatalf("tag pages = %d, want 1", len(out.Tags.Pages))
	}
	documents := []struct {
		name string
		html string
	}{
		{name: "home", html: out.Index},
		{name: "page one", html: pageHTML(t, out, "one")},
		{name: "page two", html: pageHTML(t, out, "notes/two")},
		{name: "tag index", html: out.Tags.Index},
		{name: "tag page", html: out.Tags.Pages[0].HTML},
	}
	for _, document := range documents {
		if strings.Count(document.html, scriptTag) != 1 {
			t.Errorf("%s must reference the shared client script exactly once", document.name)
		}
		if scriptAt, headEnd := strings.Index(document.html, scriptTag), strings.Index(document.html, "</head>"); scriptAt < 0 || headEnd < 0 || scriptAt > headEnd {
			t.Errorf("%s must discover the shared client script in <head>", document.name)
		}
		if strings.Contains(document.html, "var LP_BASE_PATH") || strings.Contains(document.html, "lp-copy-button") {
			t.Errorf("%s duplicates client JavaScript inline", document.name)
		}
	}
}

// Unknown or misplaced fields must be rejected as input errors (exit 1),
// never silently ignored — otherwise a v1 payload or a typo renders an empty
// default site. Covers the envelope, the nested config object, and every
// nesting level (including theme, which has a custom unmarshaler).
func TestRun_RejectsUnknownAndMisplacedFields(t *testing.T) {
	cases := map[string]string{
		"v1 garden payload":        `{"garden":{"slug":"x"},"pages":[{"slug":"a","markdown":"hi"}]}`,
		"unknown top-level":        `{"render":{"slug":"x"},"bogus":1}`,
		"unknown render field":     `{"render":{"slug":"x","tagline":"nope"}}`,
		"unknown content field":    `{"content":{"pages":[],"extra":1}}`,
		"unknown options field":    `{"options":{"emitAssets":false,"turbo":true}}`,
		"unknown page field":       `{"content":{"pages":[{"slug":"a","markdown":"hi","foo":1}]}}`,
		"unknown config field":     `{"config":{"site":{"title":"T"},"nope":1},"render":{"slug":"x"}}`,
		"unknown site field":       `{"config":{"site":{"titel":"T"}},"render":{"slug":"x"}}`,
		"unknown features field":   `{"config":{"features":{"grph":true}},"render":{"slug":"x"}}`,
		"unknown navigation field": `{"config":{"navigation":{"modee":"automatic"}},"render":{"slug":"x"}}`,
		"unknown build field":      `{"config":{"build":{"prt":8080}},"render":{"slug":"x"}}`,
		"unknown theme field":      `{"config":{"theme":{"acent":"#fff"}},"render":{"slug":"x"}}`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Run([]byte(in))
			if err == nil {
				t.Fatalf("expected an input error, got nil")
			}
			var ie *InputError
			if !errors.As(err, &ie) {
				t.Fatalf("expected *InputError, got %T: %v", err, err)
			}
		})
	}
}

const twoLinkedPages = `{
  "render": {"slug": "shivam"},
  "config": {"site": {"title": "Shivam's Garden", "baseURL": "https://example.com/g/shivam"}},
  "content": {
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
  }
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

func TestGraphEdgesDoNotDependOnBacklinks(t *testing.T) {
	out := runJSON(t, `{
  "config": {"features": {"graph": true, "backlinks": false}},
  "content": {"pages": [
    {"slug": "alpha", "title": "Alpha", "markdown": "[[beta]] and [[Beta]]"},
    {"slug": "beta", "title": "Beta", "markdown": "Body"}
  ]}
}`)

	graph := artifact(t, out, "graph.json").Content
	if got := strings.Count(graph, `"source": "alpha"`); got != 1 {
		t.Fatalf("alpha edge count = %d, want 1: %s", got, graph)
	}
	if !strings.Contains(graph, `"target": "beta"`) {
		t.Fatalf("graph is missing alpha -> beta: %s", graph)
	}
	if strings.Contains(pageHTML(t, out, "beta"), `class="lp-backlinks"`) {
		t.Fatal("backlinks were rendered while the feature was disabled")
	}
}

func TestBrokenWikilinkDegradesToPlainText(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "content": {"pages": [
	    {"slug": "solo", "title": "Solo", "markdown": "See [[Private Note|my secret]] for more."}
	  ]}
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
	  "render": {"slug": "g"},
	  "config": {
	    "theme": {
	      "accent": "#ff0000",
	      "fontHeading": "Lora",
	      "background": {"light": "#fafafa", "dark": "#101010"}
	    }
	  },
	  "content": {"pages": [{"slug": "p", "title": "P", "markdown": "hello"}]}
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
	if !strings.HasPrefix(out.CSS, templates.DefaultCSS) {
		t.Error("css output should start with the leafpress default stylesheet")
	}
	if !strings.Contains(out.CSS, "@font-face") {
		t.Error("css output should carry self-hosted @font-face rules for bundled families")
	}
}

func TestSiteIdentityAndFooterAttribution(t *testing.T) {
	out := runJSON(t, `{
	  "render": {
	    "slug": "hosted",
	    "footerAttribution": {"name": "Example Host", "url": "https://example.net"}
	  },
	  "config": {"site": {
	    "title": "Field Notes",
	    "description": "A garden about patient ideas.",
	    "author": "Garden Author"
	  }},
	  "content": {"pages": [{"slug": "welcome", "title": "Welcome", "markdown": "hello"}]}
	}`)

	for _, document := range []string{out.Index, pageHTML(t, out, "welcome")} {
		for _, want := range []string{
			"&copy; Garden Author. All rights reserved.",
			`href="https://example.net" target="_blank" rel="noopener noreferrer">Example Host</a>`,
		} {
			if !strings.Contains(document, want) {
				t.Errorf("hosted identity output missing %q", want)
			}
		}
		if strings.Contains(document, "leafpress.in") {
			t.Error("custom hosted attribution should replace the default Leafpress attribution")
		}
	}
	if !strings.Contains(pageHTML(t, out, "welcome"), `href="/welcome/">Welcome</a>`) {
		t.Error("site identity fields should preserve automatic navigation")
	}
	for _, want := range []string{
		`<meta name="description" content="A garden about patient ideas.">`,
		`<meta property="og:description" content="A garden about patient ideas.">`,
		`<meta name="twitter:description" content="A garden about patient ideas.">`,
	} {
		if !strings.Contains(out.Index, want) {
			t.Errorf("garden home metadata missing %q", want)
		}
	}
}

func TestSiteConfigAndStyleMatchLeafpressSemantics(t *testing.T) {
	out := runJSON(t, `{
	  "config": {
	    "site": {
	      "title":"Configured Garden",
	      "description":"Site description",
	      "author":"Garden Author",
	      "baseURL":"https://example.com/notes",
	      "image":"/og-default.png",
	      "headExtra":"<meta name=\"configured\" content=\"yes\">"
	    },
	    "navigation":{"mode":"explicit","items":[{"label":"Start Here","path":"/alpha/"}]},
	    "theme":{
	      "fontHeading":"Fraunces","fontBody":"Atkinson Hyperlegible",
	      "fontMono":"IBM Plex Mono","accent":"#123456",
	      "background":{"light":"#fafafa","dark":"#101010"},
	      "navStyle":"sticky","navActiveStyle":"underlined"
	    },
	    "features":{"graph":true,"search":true,"toc":false,"backlinks":true,"wikilinks":true,"rss":true},
	    "build":{"outputDir":"ignored-by-renderer","port":4444,"ignore":["private/**"]},
	    "deploy":{"provider":"netlify","settings":{"site":"demo"}}
	  },
	  "render": {"slug":"hosted"},
	  "content": {
	    "styleCSS":".custom-rule { color: rebeccapurple; }",
	    "pages":[
	      {"slug":"alpha","title":"Alpha","markdown":"## Heading\n\nSee [[beta]].","createdAt":"2026-01-01T00:00:00Z"},
	      {"slug":"beta","title":"Beta","markdown":"Beta body.","createdAt":"2026-01-02T00:00:00Z"}
	    ]
	  }
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
		t.Error("explicit nav should not be replaced by derived root-note nav")
	}
	if !strings.Contains(out.Index, `<meta name="description" content="Site description">`) {
		t.Error("site description should supply garden-home metadata")
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
		t.Errorf("feed artifact missing site config: %s", feed.Content)
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

func TestConfigDefaultsAndFeatureDisables(t *testing.T) {
	// Empty config uses exactly the CLI defaults, including enabled
	// graph/search/TOC/backlinks/wikilinks/RSS and automatic navigation.
	defaults := runJSON(t, `{
	  "render":{"slug":"g"},
	  "config":{"site":{"baseURL":"https://example.com/g/g"}},
	  "content":{"pages":[
	    {"slug":"one","title":"One","markdown":"## Heading\n\n[[two]]"},
	    {"slug":"two","title":"Two","markdown":"body"}
	  ]}
	}`)
	one := pageHTML(t, defaults, "one")
	for _, want := range []string{
		"My Garden", `class="lp-toc"`, `class="lp-backlink"`,
		`class="lp-wikilink"`, `class="lp-graph-toggle"`, `class="lp-search-toggle"`,
	} {
		combined := one + pageHTML(t, defaults, "two")
		if !strings.Contains(combined, want) {
			t.Errorf("config defaults missing %q", want)
		}
	}
	artifact(t, defaults, "graph.json")
	artifact(t, defaults, "search-index.json")
	artifact(t, defaults, "feed.xml")
	// Automatic navigation (the default mode) lists the garden's root notes.
	if !strings.Contains(one, `class="lp-nav-link"`) {
		t.Error("automatic navigation should list root notes by default")
	}

	disabled := runJSON(t, `{
	  "render":{"slug":"g"},
	  "config":{"features":{"graph":false,"search":false,"toc":false,"backlinks":false,"wikilinks":false,"rss":false}},
	  "content":{"pages":[
	    {"slug":"one","title":"One","markdown":"## Heading\n\n[[two]]"},
	    {"slug":"two","title":"Two","markdown":"body"}
	  ]}
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
		if item.Path == "graph.json" || item.Path == "feed.xml" {
			t.Errorf("disabled artifact still emitted: %s", item.Path)
		}
	}
	// search-index.json is always emitted for link previews even when the
	// search UI is off.
	search := artifact(t, disabled, "search-index.json")
	if !strings.Contains(search.Content, `"url": "/one/"`) ||
		!strings.Contains(search.Content, `"title": "One"`) {
		t.Errorf("search-index should still list pages when search UI is off: %s", search.Content)
	}
	if !strings.Contains(clientScriptArtifact(t, disabled).Content, "search-index.json") {
		t.Error("link preview script should still fetch search-index.json when search UI is off")
	}
}

func TestConfigRejectsInvalidValues(t *testing.T) {
	_, err := Run([]byte(`{
	  "render":{"slug":"g"},
	  "config":{"theme":{"accent":"red"}},
	  "content":{"pages":[]}
	}`))
	if err == nil || !strings.Contains(err.Error(), "accent color") {
		t.Fatalf("expected config validation error, got %v", err)
	}
}

func TestIndexSortOrder(t *testing.T) {
	const pageItems = `
	  {"slug": "b-old", "title": "Zebra", "growth": "evergreen", "markdown": "x", "createdAt": "2026-01-01T00:00:00Z"},
	  {"slug": "a-new", "title": "Apple", "growth": "seedling", "markdown": "x", "createdAt": "2026-03-01T00:00:00Z"},
	  {"slug": "c-mid", "title": "Mango", "growth": "budding", "markdown": "x", "createdAt": "2026-02-01T00:00:00Z"}`

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
			// The garden home defaults to date order. A non-date home ordering
			// is expressed by a root index page carrying the sort key.
			pages := "[" + pageItems + "]"
			if tt.sort != "" && tt.sort != "date" {
				pages = `[{"slug":"","markdown":"","isIndex":true,"sort":"` + tt.sort + `"},` + pageItems + `]`
			}
			out := runJSON(t, `{"render": {"slug": "g"}, "content": {"pages": `+pages+`}}`)
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
		{"malformed JSON", `{"config": {`},
		{"missing page slug", `{"render": {"slug": "g"}, "content": {"pages": [{"title": "T", "markdown": "x"}]}}`},
		{"duplicate page slugs", `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "a", "markdown": "x"}, {"slug": "a", "markdown": "y"}]}}`},
		{"bad createdAt", `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "a", "markdown": "x", "createdAt": "yesterday"}]}}`},
		{"bad page sort", `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "a", "markdown": "x", "sort": "popularity"}]}}`},
		{"bad accent", `{"render": {"slug": "g"}, "config": {"theme": {"accent": "red;}</style>"}}, "content": {"pages": []}}`},
		{"bad font", `{"render": {"slug": "g"}, "config": {"theme": {"fontBody": "Inter\"><script>"}}, "content": {"pages": []}}`},
		{"unsupported contractVersion", `{"contractVersion": 3, "render": {"slug": "g"}, "content": {"pages": []}}`},
		{"dot-segment slug", `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "essays/../secret", "markdown": "x"}]}}`},
		{"attribution without name", `{"render": {"slug": "g", "footerAttribution": {"url": "https://example.net"}}, "content": {"pages": []}}`},
		{"unsafe attribution URL", `{"render": {"slug": "g", "footerAttribution": {"name": "Example Host", "url": "javascript:alert(1)"}}, "content": {"pages": []}}`},
		{"unsafe slug", `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "a\"b", "markdown": "x"}]}}`},
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
	out := runJSON(t, `{"render": {"slug": "shivam"}, "content": {"pages": [{"slug": "my-note", "markdown": "hi"}]}}`)

	// Absent config renders the default site (title "My Garden").
	if !strings.Contains(out.Index, "<title>My Garden | My Garden</title>") {
		t.Error("index title should use the default site title")
	}
	if !strings.Contains(out.Index, `href="https://leafpress.in"`) {
		t.Error("renderer output should retain the default Leafpress attribution")
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
		t.Error("page should link /style.css when baseURL is empty")
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
			input := `{"render": {"slug": "g"}, "config": {"site": {"title": ` + jsonString(gardenTitle) + `}}, "content": {"pages": [` + tt.page + `]}}`
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

func TestXSSTOCHeadingTextRemainsEscaped(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "content": {"pages": [{
	    "slug": "p",
	    "title": "P",
	    "markdown": "## <input autofocus onfocus=alert(1)>"
	  }]}
	}`)
	html := pageHTML(t, out, "p")

	if strings.Contains(html, `<input autofocus onfocus=alert(1)>`) {
		t.Fatal("TOC rendered escaped author HTML as a live input element")
	}
	if !strings.Contains(html, `&lt;input autofocus onfocus=alert(1)&gt;`) {
		t.Fatal("page should preserve the hostile heading as visibly escaped text")
	}
}

func TestXSSThemeBackgroundBreakoutRejected(t *testing.T) {
	_, err := Run([]byte(`{
	  "render": {"slug": "g"},
	  "config": {"theme": {"background": {
	    "light": "rgb(0)</style><input autofocus onfocus=alert(1)>",
	    "dark": "#000"
	  }}}
	}`))
	if err == nil {
		t.Fatal("expected style-breakout background to be rejected")
	}
	var inputErr *InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("expected *InputError, got %T: %v", err, err)
	}
}

func TestXSSBaseURLBreakoutRejected(t *testing.T) {
	_, err := Run([]byte(`{
	  "render": {"slug": "g"},
	  "config": {"site": {
	    "baseURL": "https://example.com/\"/><input autofocus onfocus=alert(1)>"
	  }}
	}`))
	if err == nil {
		t.Fatal("expected markup-bearing baseURL to be rejected")
	}
	var inputErr *InputError
	if !errors.As(err, &inputErr) {
		t.Fatalf("expected *InputError, got %T: %v", err, err)
	}
}

func TestBasePathUsesURLPathEscaping(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "config": {"site": {"baseURL": "https://example.com/%22garden"}},
	  "content": {"pages": [{"slug": "p", "title": "P", "markdown": "body"}]}
	}`)
	html := pageHTML(t, out, "p")
	if !strings.Contains(html, `href="/%22garden/style.css"`) {
		t.Fatalf("base path should retain URL escaping: %s", html)
	}
	if strings.Contains(html, `href="/"garden`) {
		t.Fatal("base path was decoded into an HTML attribute delimiter")
	}
}

func TestSiteIdentityAndAttributionEscaped(t *testing.T) {
	out := runJSON(t, `{
	  "render": {
	    "slug": "g",
	    "footerAttribution": {
	      "name": "<script>alert(3)</script>",
	      "url": "https://example.com/?one=1&two=2"
	    }
	  },
	  "config": {"site": {
	    "description": "say \"><script>alert(1)</script> & leave",
	    "author": "<img src=x onerror=alert(2)>"
	  }},
	  "content": {"pages": [{"slug": "p", "title": "P", "markdown": "body"}]}
	}`)

	for _, document := range []string{out.Index, pageHTML(t, out, "p")} {
		for _, raw := range []string{
			"<script>alert(1)</script>",
			"<img src=x onerror=alert(2)>",
			"<script>alert(3)</script>",
		} {
			if strings.Contains(document, raw) {
				t.Errorf("hosted identity output contains unescaped fragment %q", raw)
			}
		}
		for _, want := range []string{
			"&copy; &lt;img src=x onerror=alert(2)&gt;. All rights reserved.",
			`href="https://example.com/?one=1&amp;two=2"`,
			"&lt;script&gt;alert(3)&lt;/script&gt;",
		} {
			if !strings.Contains(document, want) {
				t.Errorf("hosted identity output missing escaped fragment %q", want)
			}
		}
	}
	if !strings.Contains(
		out.Index,
		`content="say &#34;&gt;&lt;script&gt;alert(1)&lt;/script&gt; &amp; leave"`,
	) {
		t.Error("garden description is not safely escaped in home metadata")
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
	  "render": {"slug": "g"},
	  "content": {"pages": [
	    {"slug": "evil", "title": "<script>alert(1)</script>", "markdown": "links to [[safe]]"},
	    {"slug": "safe", "title": "Safe", "markdown": "plain"}
	  ]}
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
			input := `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "p", "markdown": "x", "tags": ["` + tt.tag + `"]}]}}`
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
	  "render": {"slug": "g"},
	  "config": {"site": {"baseURL": "https://example.com/g/shivam"}},
	  "content": {"pages": [
	    {"slug": "a", "title": "A", "markdown": "x", "tags": ["Systems", "systems", "go-lang"], "createdAt": "2026-01-02T00:00:00Z"},
	    {"slug": "b", "title": "B", "markdown": "y", "tags": ["systems"], "createdAt": "2026-01-01T00:00:00Z"}
	  ]}
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
	if got := strings.Count(pageHTML(t, out, "a"), `/tags/systems/`); got != 1 {
		t.Errorf("case-variant tag rendered %d links, want 1", got)
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

func TestInlineTagsFeedHostedTagOutputs(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "config": {"site": {"baseURL": "https://example.com/g/shivam"}},
	  "content": {"pages": [
	    {"slug": "a", "title": "A", "markdown": "Body with #LeafPress and `+"`#ignored`"+`.", "tags": ["Systems"]},
	    {"slug": "b", "title": "B", "markdown": "Body with #leafpress."}
	  ]}
	}`)

	if len(out.Tags.Pages) != 2 || out.Tags.Pages[0].Tag != "leafpress" || out.Tags.Pages[1].Tag != "systems" {
		t.Fatalf("tag pages = %#v, want leafpress and systems", out.Tags.Pages)
	}
	page := pageHTML(t, out, "a")
	if !strings.Contains(page, `class="lp-tag lp-inline-tag" href="/g/shivam/tags/leafpress/"`) {
		t.Fatalf("page missing linked inline tag:\n%s", page)
	}
	if strings.Contains(out.Tags.Index, "ignored") {
		t.Fatalf("inline-code tag leaked into tag index:\n%s", out.Tags.Index)
	}
	leafpress := out.Tags.Pages[0].HTML
	if !strings.Contains(leafpress, `href="/g/shivam/a/"`) || !strings.Contains(leafpress, `href="/g/shivam/b/"`) {
		t.Fatalf("inline tag page does not list both tagged pages:\n%s", leafpress)
	}
}

func TestNoTagsEmptyTagsOutput(t *testing.T) {
	out := runJSON(t, `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "p", "markdown": "x"}]}}`)

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
	  "render": {"slug": "g"},
	  "config": {"site": {"title": "T"}},
	  "content": {"pages": [
	    {"slug": "a", "title": "<script>x</script>", "markdown": "> [!note] Hi\n> body <script>alert(1)</script>\n\nlink [[b]] and ![[demo.mp4]]", "tags": ["one", "two"]},
	    {"slug": "b", "title": "B", "markdown": "<div class=\"x\">\nraw\n</div>", "tags": ["one"]}
	  ]}
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

// Hosted authors link by display title ([[Beta Note]]), while public
// slugs are hyphenated (beta-note). Titles register as resolver aliases.
func TestWikilinkResolvesByPageTitle(t *testing.T) {
	out := runJSON(t, `{
  "render": {"slug": "shivam"},
  "config": {"site": {"baseURL": "https://example.com/g/shivam"}},
  "content": {"pages": [
    {
      "slug": "alpha-note",
      "title": "Alpha Note",
      "markdown": "See [[Beta Note]] and [[beta   NOTE|the beta one]] and [[Gamma Note]]."
    },
    {"slug": "beta-note", "title": "Beta Note", "markdown": "Beta content."}
  ]}
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
  "render": {"slug": "shivam"},
  "config": {"site": {"title": "Shivam's Garden", "baseURL": "https://example.com/g/shivam"}},
  "content": {"pages": [
    {"slug": "hello", "title": "Hello", "markdown": "Root note.", "createdAt": "2026-01-01T00:00:00Z"},
    {"slug": "essays/first", "title": "First Essay", "markdown": "One.", "createdAt": "2026-02-01T00:00:00Z"},
    {"slug": "essays/second", "title": "Second Essay", "markdown": "Two.", "createdAt": "2026-03-01T00:00:00Z"},
    {"slug": "essays", "title": "Essays", "markdown": "Long-form writing.", "isIndex": true},
    {"slug": "recipes/dal", "title": "Dal", "markdown": "Cook it.", "createdAt": "2026-04-01T00:00:00Z"}
  ]}
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
		Render: RenderOpts{Slug: "garden"},
		Config: json.RawMessage(`{"site":{"title":"Garden"},"navigation":{"includeTags":true}}`),
		Content: Content{Pages: []InputPage{{
			Slug:     "tagged",
			Title:    "Tagged",
			Markdown: "Body",
			Tags:     []string{"ideas"},
		}}},
	}

	out, err := Render(input)
	if err != nil {
		t.Fatal(err)
	}
	html := pageHTML(t, out, "tagged")
	if !strings.Contains(html, `class="lp-nav-link" href="/tags/">Tags</a>`) {
		t.Error("enabled tags navigation should link the generated tags index")
	}

	input.Config = json.RawMessage(`{"site":{"title":"Garden"},"navigation":{"includeTags":false}}`)
	out, err = Render(input)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(pageHTML(t, out, "tagged"), `href="/tags/"`) {
		t.Error("tags navigation should be opt-in")
	}

	input.Config = json.RawMessage(`{"site":{"title":"Garden"},"navigation":{"includeTags":true}}`)
	input.Content.Pages[0].Tags = nil
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
		Render: RenderOpts{Slug: "g"},
		Config: json.RawMessage(`{"site":{"title":"Garden"}}`),
		Content: Content{Pages: []InputPage{{
			Slug:        "configured",
			Title:       "Configured",
			Markdown:    "## Hidden heading\n\nBody.",
			Description: "A custom description",
			Growth:      "evergreen",
			TOC:         &showTOC,
			Image:       "/images/card.png",
			ReadingTime: &readingTime,
		}}},
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
		Render:  RenderOpts{Slug: "g"},
		Content: Content{Pages: []InputPage{{Slug: "bad", ReadingTime: &zero}}},
	})
	if err == nil || !strings.Contains(err.Error(), "readingTime") {
		t.Fatalf("expected readingTime validation error, got %v", err)
	}
}

func TestRootIndexReplacesGardenHome(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "config": {"site": {"title": "My Garden"}},
	  "content": {"pages": [
	    {"slug": "", "title": "", "markdown": "Welcome to my garden.", "isIndex": true},
	    {"slug": "note", "title": "Note", "markdown": "n."},
	    {"slug": "essays/one", "title": "One", "markdown": "o."}
	  ]}
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
	  "render": {"slug": "g"},
	  "content": {"pages": [
	    {"slug": "s/b", "title": "Banana", "markdown": "x", "createdAt": "2026-03-01T00:00:00Z"},
	    {"slug": "s/a", "title": "Apple", "markdown": "x", "createdAt": "2026-01-01T00:00:00Z"},
	    {"slug": "s", "title": "S", "markdown": "intro", "isIndex": true, "sort": "title"},
	    {"slug": "hidden/x", "title": "X", "markdown": "x"},
	    {"slug": "hidden", "title": "Hidden", "markdown": "no list here", "isIndex": true, "showList": false}
	  ]}
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
	  "render": {"slug": "g"},
	  "config": {"site": {"baseURL": "https://example.com/g/me"}},
	  "content": {"pages": [
	    {"slug": "essays/deep-note", "title": "Deep Note", "markdown": "x"},
	    {"slug": "top", "title": "Top", "markdown": "See [[Deep Note]]."}
	  ]}
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
		{"empty slug without isIndex", `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "", "markdown": "x"}]}}`},
		{"duplicate root index", `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "", "markdown": "x", "isIndex": true}, {"slug": "/", "markdown": "y", "isIndex": true}]}}`},
		{"bad section sort", `{"render": {"slug": "g"}, "content": {"pages": [{"slug": "s", "markdown": "x", "isIndex": true, "sort": "random"}]}}`},
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
	  "render": {"slug": "g"},
	  "content": {"pages": [
	    {"slug": "long-form-essays/one", "title": "One", "markdown": "x"},
	    {"slug": "long-form-essays", "title": "", "markdown": "intro", "isIndex": true}
	  ]}
	}`)

	home := pageHTML(t, out, "long-form-essays")
	if !strings.Contains(home, "Long Form Essays") {
		t.Error("untitled index page should be titled from its slug segment")
	}
}

func TestAutoSectionTitleReadsHyphensAsSpaces(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "content": {"pages": [{"slug": "field-notes/one", "title": "One", "markdown": "x"}]}
	}`)
	if len(out.Sections) != 1 || !strings.Contains(out.Sections[0].HTML, "Field Notes") {
		t.Errorf("auto section should be titled 'Field Notes', got: %v", out.Sections)
	}
}

func TestSelfHostedFontsInStylesheet(t *testing.T) {
	out := runJSON(t, twoLinkedPages)
	html := pageHTML(t, out, "alpha")

	// Default families are all bundled: @font-face rules live in the
	// generated stylesheet with site-relative URLs (the stylesheet is
	// served from the garden root, so they resolve under the base path).
	if !strings.Contains(out.CSS, "@font-face") {
		t.Error("stylesheet missing @font-face rules")
	}
	if !strings.Contains(out.CSS, `url("static/leafpress/fonts/inter-normal-latin.woff2")`) {
		t.Error("font URLs must be site-relative in the stylesheet")
	}
	if strings.Contains(html, "@font-face") {
		t.Error("@font-face must not be inlined into every page head")
	}
	if strings.Contains(html, "fonts.googleapis.com") || strings.Contains(html, "fonts.gstatic.com") {
		t.Error("bundled default fonts must not reference Google Fonts")
	}
	preloadPaths := []string{
		"/g/shivam/static/leafpress/fonts/bricolage-grotesque-normal-latin.woff2",
		"/g/shivam/static/leafpress/fonts/inter-normal-latin.woff2",
		"/g/shivam/static/leafpress/fonts/jetbrains-mono-normal-latin.woff2",
	}
	last := -1
	for _, fontPath := range preloadPaths {
		link := `rel="preload" href="` + fontPath + `" as="font" type="font/woff2" crossorigin`
		index := strings.Index(html, link)
		if index < 0 {
			t.Errorf("page head missing font preload %q", link)
		}
		if index <= last {
			t.Errorf("font preloads are not in theme role order: %v", preloadPaths)
		}
		last = index
	}
	// No warnings: every family is self-hosted.
	for _, w := range out.Warnings {
		if strings.Contains(w, "font family") {
			t.Errorf("unexpected font warning: %s", w)
		}
	}
}

func TestUnbundledFontWarnsAndFallsBackLocally(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "config": {"theme": {"fontBody": "Lobster"}},
	  "content": {"pages": [{"slug": "note", "title": "Note", "markdown": "hi"}]}
	}`)
	html := pageHTML(t, out, "note")

	// Self-contained by default: no remote link, a warning instead, and
	// the CSS variable fallback stack takes over in the browser.
	if strings.Contains(html, "fonts.googleapis.com") {
		t.Error("unbundled family must not load remotely by default")
	}
	var warned bool
	for _, w := range out.Warnings {
		if strings.Contains(w, `"Lobster"`) {
			warned = true
		}
	}
	if !warned {
		t.Errorf("expected a fallback warning for Lobster, got %v", out.Warnings)
	}
	// Bundled heading/mono fonts still self-host.
	if !strings.Contains(out.CSS, "@font-face") {
		t.Error("bundled families should still emit @font-face rules")
	}
}

func TestRemoteFontsIsDeprecatedOptIn(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "config": {"theme": {"fontBody": "Lobster", "remoteFonts": true}},
	  "content": {"pages": [{"slug": "note", "title": "Note", "markdown": "hi"}]}
	}`)
	html := pageHTML(t, out, "note")

	if !strings.Contains(html, "fonts.googleapis.com/css2?family=Lobster") {
		t.Error("remoteFonts opt-in should keep the Google Fonts link")
	}
	if strings.Contains(html, "family=Crimson+Pro") || strings.Contains(html, "family=JetBrains+Mono") {
		t.Error("bundled families must not appear in the remote URL")
	}
	for _, w := range out.Warnings {
		if strings.Contains(w, `"Lobster"`) {
			t.Errorf("explicit opt-in should not warn: %s", w)
		}
	}
}

func TestCustomLocalFontsRenderPortableCSS(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "config": {
	    "site": {"title": "G"},
	    "theme": {
	      "fontBody": "My Serif",
	      "fonts": [{"family": "My Serif", "file": "static/fonts/my.woff2", "weight": "400 700"}]
	    }
	  },
	  "content": {"pages": [{"slug": "note", "title": "Note", "markdown": "hi"}]}
	}`)
	// Font CSS lives in the shared stylesheet with stylesheet-relative
	// URLs: url("static/...") resolves against {basePath}/style.css, so no
	// base path is baked into the CSS itself.
	if !strings.Contains(out.CSS, `font-family: "My Serif"`) {
		t.Error("custom @font-face missing from stylesheet")
	}
	if !strings.Contains(out.CSS, `url("static/fonts/my.woff2")`) {
		t.Error("custom font URL must be stylesheet-relative")
	}
	html := pageHTML(t, out, "note")
	if strings.Contains(html, "fonts.googleapis.com") || strings.Contains(out.CSS, "fonts.googleapis.com") {
		t.Error("declared custom family must not trigger a remote font link")
	}
	if !strings.Contains(html, `rel="preload" href="/static/fonts/my.woff2" as="font" type="font/woff2" crossorigin`) {
		t.Error("declared custom family should emit a matching preload link")
	}
}

func TestInvalidCustomFontConfigRejected(t *testing.T) {
	_, err := Run([]byte(`{
	  "render": {"slug": "g"},
	  "config": {"theme": {"fonts": [{"family": "X", "file": "static/fonts/../../etc/passwd"}]}},
	  "content": {"pages": []}
	}`))
	var inputErr *InputError
	if err == nil || !errors.As(err, &inputErr) {
		t.Fatalf("traversal font file must be an input error, got %v", err)
	}
}

func TestAssetManifestAlwaysEmitted(t *testing.T) {
	out := runJSON(t, twoLinkedPages)

	if out.AssetRegistryID != assets.RegistryID() {
		t.Errorf("assetRegistryId = %q, want %q", out.AssetRegistryID, assets.RegistryID())
	}
	if err := out.AssetManifest.Validate(); err != nil {
		t.Fatalf("asset manifest invalid: %v", err)
	}
	// Default families are all bundled: 3 favicons + 10 font faces + 3 OFL
	// license texts. Mermaid is content-optional and absent here.
	if len(out.AssetManifest) != 16 {
		t.Fatalf("manifest has %d entries, want 16", len(out.AssetManifest))
	}
	byPath := map[string]assets.Asset{}
	for _, a := range out.AssetManifest {
		byPath[a.LogicalPath] = a
	}
	if a, ok := byPath[assets.BuiltinFaviconICO]; !ok || a.OutputPath != "favicon.ico" {
		t.Error("favicon.ico missing or missing root output path")
	}
	if _, ok := byPath["static/leafpress/fonts/inter-normal-latin.woff2"]; !ok {
		t.Error("bundled font face missing from manifest")
	}
	if _, ok := byPath["static/leafpress/fonts/OFL-inter.txt"]; !ok {
		t.Error("OFL license text missing from manifest")
	}

	// Without emitAssets no binary artifacts appear, and every artifact is utf8.
	for _, a := range out.Artifacts {
		if a.Encoding != "utf8" {
			t.Errorf("artifact %s encoding = %q, want utf8", a.Path, a.Encoding)
		}
	}
}

func TestManifestSkipsFontsOfUnbundledFamilies(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "config": {"theme": {"fontHeading": "Lobster", "fontBody": "Lobster", "fontMono": "Roboto Mono"}},
	  "content": {"pages": [{"slug": "note", "title": "Note", "markdown": "hi"}]}
	}`)
	for _, a := range out.AssetManifest {
		if strings.HasPrefix(a.LogicalPath, assets.BuiltinPrefix+"fonts/") {
			t.Errorf("unbundled theme should not require font asset %s", a.LogicalPath)
		}
	}
	// Favicons are still required (every page head links them).
	if len(out.AssetManifest) != 3 {
		t.Errorf("manifest has %d entries, want 3 favicons", len(out.AssetManifest))
	}
}

func TestMermaidSelfHostedInManifestAndHTML(t *testing.T) {
	// JSON body built with a quoted string so mermaid fences do not break
	// Go raw-string literals (which also use backticks).
	const body = "```mermaid\ngraph TD\nA-->B\n```"
	payload, err := json.Marshal(map[string]any{
		"render": map[string]any{"slug": "g"},
		"config": map[string]any{"site": map[string]any{"baseURL": "https://example.com/g/g"}},
		"content": map[string]any{"pages": []map[string]any{
			{"slug": "diag", "title": "Diag", "markdown": body},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := runJSON(t, string(payload))
	byPath := map[string]assets.Asset{}
	for _, a := range out.AssetManifest {
		byPath[a.LogicalPath] = a
	}
	if _, ok := byPath[assets.BuiltinMermaidJS]; !ok {
		t.Error("mermaid script missing from asset manifest when diagrams are used")
	}
	if _, ok := byPath[assets.BuiltinMermaidLicense]; !ok {
		t.Error("mermaid license missing from asset manifest when diagrams are used")
	}
	// Default fonts (13) + favicons (3) + mermaid (2) = 18
	if len(out.AssetManifest) != 18 {
		t.Fatalf("manifest has %d entries, want 18 with mermaid", len(out.AssetManifest))
	}

	html := pageHTML(t, out, "diag")
	if strings.Contains(html, "cdn.jsdelivr") || strings.Contains(html, "unpkg.com") {
		t.Error("rendered page must not load Mermaid from a CDN")
	}
	client := clientScriptArtifact(t, out)
	if !strings.Contains(html, `src="/g/g/`+client.Path+`" defer`) {
		t.Error("page must load the base-prefixed shared client script")
	}
	if !strings.Contains(client.Content, `var LP_BASE_PATH = '/g/g'`) {
		t.Error("client script must set LP_BASE_PATH from garden baseURL")
	}
	if !strings.Contains(client.Content, `LP_BASE_PATH + '/static/leafpress/mermaid/mermaid.min.js'`) {
		t.Error("page must load self-hosted mermaid via LP_BASE_PATH")
	}
	if !strings.Contains(html, `class="mermaid"`) {
		t.Error("page body should contain mermaid diagram container")
	}
}

func TestEmitAssetsProducesBase64Artifacts(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "options": {"emitAssets": true},
	  "content": {"pages": [{"slug": "note", "title": "Note", "markdown": "hi"}]}
	}`)

	assetArtifacts := map[string]OutputArtifact{}
	for _, a := range out.Artifacts {
		if a.Encoding == "base64" {
			assetArtifacts[a.Path] = a
		}
	}
	if len(assetArtifacts) != len(out.AssetManifest) {
		t.Fatalf("%d base64 artifacts, want %d (one per manifest entry)", len(assetArtifacts), len(out.AssetManifest))
	}
	for _, entry := range out.AssetManifest {
		// Artifacts are keyed by the effective output path: the exact
		// filename a CLI export serves (favicon.ico at the root, not its
		// registry logical path).
		artifact, ok := assetArtifacts[entry.EffectiveOutputPath()]
		if !ok {
			t.Errorf("no artifact for manifest entry %s at %s", entry.LogicalPath, entry.EffectiveOutputPath())
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(artifact.Content)
		if err != nil {
			t.Errorf("%s: content is not valid base64: %v", entry.LogicalPath, err)
			continue
		}
		if got := assets.Sum(raw); got != entry.SHA256 {
			t.Errorf("%s: decoded hash %s != manifest hash %s", entry.LogicalPath, got, entry.SHA256)
		}
		if int64(len(raw)) != entry.Size {
			t.Errorf("%s: decoded size %d != manifest size %d", entry.LogicalPath, len(raw), entry.Size)
		}
		if artifact.ContentType != entry.ContentType {
			t.Errorf("%s: artifact content type %q != manifest %q", entry.LogicalPath, artifact.ContentType, entry.ContentType)
		}
	}

	// Text artifacts stay utf8 alongside.
	if a := artifact(t, out, "robots.txt"); a.Encoding != "utf8" {
		t.Errorf("robots.txt encoding = %q", a.Encoding)
	}
}

func TestCallerAssetsMergeOverrideAndWarnings(t *testing.T) {
	fontHash := assets.Sum([]byte("custom font bytes"))
	faviconHash := assets.Sum([]byte("user favicon bytes"))
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "config": {
	    "site": {"title": "G"},
	    "theme": {
	      "fontBody": "My Serif",
	      "fonts": [{"family": "My Serif", "file": "static/fonts/my.woff2", "weight": "400 700"}]
	    }
	  },
	  "options": {"emitAssets": true},
	  "content": {
	    "assets": [
	      {"logicalPath": "static/fonts/my.woff2", "contentType": "font/woff2", "sha256": "`+fontHash+`", "size": 17},
	      {"logicalPath": "static/user-favicon.ico", "contentType": "image/x-icon", "sha256": "`+faviconHash+`", "size": 18, "outputPath": "favicon.ico"}
	    ],
	    "pages": [{"slug": "note", "title": "Note", "markdown": "hi"}]
	  }
	}`)

	if err := out.AssetManifest.Validate(); err != nil {
		t.Fatalf("combined manifest invalid: %v", err)
	}
	byOutput := map[string]assets.Asset{}
	for _, a := range out.AssetManifest {
		byOutput[a.EffectiveOutputPath()] = a
	}

	// The declared custom font joins the manifest, so hosts can pin it into
	// the same publication snapshot as the HTML.
	if a, ok := byOutput["static/fonts/my.woff2"]; !ok || a.SHA256 != fontHash {
		t.Error("caller-declared custom font missing from combined manifest")
	}
	// The caller favicon replaces the built-in entry at favicon.ico.
	if a, ok := byOutput["favicon.ico"]; !ok || a.SHA256 != faviconHash {
		t.Errorf("caller favicon did not override the built-in entry: %+v", byOutput["favicon.ico"])
	}
	// A declared custom font produces no undeclared-font warning.
	for _, w := range out.Warnings {
		if strings.Contains(w, "caller asset manifest") {
			t.Errorf("unexpected undeclared-font warning: %q", w)
		}
	}

	// Byte emission: built-ins only, keyed by effective output path. The
	// overridden favicon and the caller font must not be emitted — the
	// renderer does not have those bytes.
	emitted := map[string]bool{}
	for _, a := range out.Artifacts {
		if a.Encoding == "base64" {
			emitted[a.Path] = true
		}
	}
	if emitted["favicon.ico"] {
		t.Error("overridden favicon bytes emitted from the registry")
	}
	if emitted["static/fonts/my.woff2"] {
		t.Error("caller font bytes emitted — the renderer cannot have them")
	}
	if !emitted["favicon.svg"] {
		t.Error("non-overridden built-in favicon.svg not emitted")
	}
}

func TestUndeclaredCustomFontWarns(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "config": {
	    "site": {"title": "G"},
	    "theme": {"fonts": [{"family": "My Serif", "file": "static/fonts/my.woff2"}]}
	  },
	  "content": {"pages": [{"slug": "note", "title": "Note", "markdown": "hi"}]}
	}`)
	found := false
	for _, w := range out.Warnings {
		if strings.Contains(w, "static/fonts/my.woff2") && strings.Contains(w, "caller asset manifest") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected undeclared-custom-font warning, got %v", out.Warnings)
	}
}

func TestCallerAssetsRejectedWhenInvalid(t *testing.T) {
	hash := assets.Sum([]byte("x"))
	cases := map[string]string{
		"reserved namespace":    `{"logicalPath": "static/leafpress/evil.woff2", "contentType": "font/woff2", "sha256": "` + hash + `", "size": 1}`,
		"traversal":             `{"logicalPath": "static/../evil.woff2", "contentType": "font/woff2", "sha256": "` + hash + `", "size": 1}`,
		"bad hash":              `{"logicalPath": "static/fonts/a.woff2", "contentType": "font/woff2", "sha256": "nope", "size": 1}`,
		"style.css collision":   `{"logicalPath": "static/user.css", "contentType": "text/css", "sha256": "` + hash + `", "size": 1, "outputPath": "style.css"}`,
		"index.html collision":  `{"logicalPath": "static/evil.html", "contentType": "text/html", "sha256": "` + hash + `", "size": 1, "outputPath": "index.html"}`,
		"404.html collision":    `{"logicalPath": "static/evil404.html", "contentType": "text/html", "sha256": "` + hash + `", "size": 1, "outputPath": "404.html"}`,
		"feed.xml collision":    `{"logicalPath": "static/evil.xml", "contentType": "application/xml", "sha256": "` + hash + `", "size": 1, "outputPath": "feed.xml"}`,
		"arbitrary outputPath":  `{"logicalPath": "static/a.txt", "contentType": "text/plain", "sha256": "` + hash + `", "size": 1, "outputPath": "elsewhere.txt"}`,
		"font outputPath drift": `{"logicalPath": "static/fonts/my.woff2", "contentType": "font/woff2", "sha256": "` + hash + `", "size": 1, "outputPath": "static/fonts/other.woff2"}`,
		"duplicate output path": `{"logicalPath": "static/a.ico", "contentType": "image/x-icon", "sha256": "` + hash + `", "size": 1, "outputPath": "favicon.ico"},
		{"logicalPath": "static/b.ico", "contentType": "image/x-icon", "sha256": "` + hash + `", "size": 1, "outputPath": "favicon.ico"}`,
		"bare reserved dir": `{"logicalPath": "static/leafpress", "contentType": "text/plain", "sha256": "` + hash + `", "size": 1}`,
	}
	for name, entry := range cases {
		_, err := Run([]byte(`{
		  "render": {"slug": "g"},
		  "content": {"assets": [` + entry + `], "pages": []}
		}`))
		var inputErr *InputError
		if err == nil || !errors.As(err, &inputErr) {
			t.Errorf("%s: want input error, got %v", name, err)
		}
	}
}

func TestWikilinkResolvesRawTitleWithEntities(t *testing.T) {
	out := runJSON(t, `{
	  "render": {"slug": "g"},
	  "content": {"pages": [
	    {"slug": "target", "title": "Foo & Bar", "markdown": "content"},
	    {"slug": "source", "title": "Source", "markdown": "See [[Foo & Bar]]."}
	  ]}
	}`)
	html := pageHTML(t, out, "source")
	if !strings.Contains(html, `href="/target/"`) {
		t.Errorf("raw title with & must resolve as a wikilink:\n%s", html)
	}
}

func TestInvalidThemeFontFailsViaSharedValidation(t *testing.T) {
	_, err := Run([]byte(`{
	  "render": {"slug": "g"},
	  "config": {"theme": {"fontBody": "Evil\"; @import url(x); \""}},
	  "content": {"pages": []}
	}`))
	var inputErr *InputError
	if err == nil || !errors.As(err, &inputErr) {
		t.Fatalf("unsafe font name must fail as input error via config.Validate, got %v", err)
	}
}
