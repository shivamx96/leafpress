package render

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEverythingRenderer(t *testing.T) {
	out := runJSON(t, `{"config":{"site":{"baseURL":"https://example.com/garden"},"navigation":{"everything":true,"everythingExclude":["/about/"]}},"content":{"pages":[{"slug":"notes/one","title":"One & Two","markdown":"Body"},{"slug":"about","title":"About","markdown":"About"}]}}`)
	var doc string
	for _, s := range out.Sections {
		if s.Slug == "everything" {
			doc = s.HTML
		}
	}
	if doc == "" {
		t.Fatal("missing Everything section")
	}
	start := strings.Index(doc, "<main")
	main := doc[start:strings.Index(doc, "</main>")]
	if !strings.Contains(main, `href="/garden/notes/one/"`) || !strings.Contains(main, "One &amp; Two") || strings.Contains(main, `href="/garden/about/"`) {
		t.Fatalf("bad listing: %s", main)
	}
	if !strings.Contains(doc, `href="/garden/everything/" aria-current="page"`) {
		t.Fatal("missing active index destination")
	}
	for _, slug := range []string{"everything", "everything/child"} {
		_, err := Render(&Input{Config: json.RawMessage(`{"navigation":{"everything":true}}`), Content: Content{Pages: []InputPage{{Slug: slug, Title: "Collision"}}}})
		if err == nil {
			t.Fatalf("accepted reserved route %s", slug)
		}
	}
}

func TestBottomNavigationRendering(t *testing.T) {
	for _, everything := range []string{"true", "false"} {
		out := runJSON(t, `{"config":{"site":{"title":"Garden & friends","baseURL":"https://example.com/garden"},"theme":{"navStyle":"bottom"},"navigation":{"everything":`+everything+`}}}`)
		if !strings.Contains(out.Index, `data-nav-style="bottom"`) || !strings.Contains(out.Index, `class="lp-site-title" href="/garden/">Garden &amp; friends</a>`) {
			t.Fatal("missing dock or top identity")
		}
		if strings.Contains(out.Index, `class="lp-nav-title"`) || strings.Contains(out.Index, `class="lp-nav-placeholder"`) {
			t.Fatal("bottom dock retains top navigation spacing")
		}
	}
}
