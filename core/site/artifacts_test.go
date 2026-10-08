package site

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
	"github.com/shivamx96/leafpress/core/templates"
)

func TestArtifactShapesAndOrdering(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	modified := created.Add(24 * time.Hour)
	pages := []*content.Page{
		{
			Slug: "alpha", Title: "Alpha & One", Permalink: "/alpha/",
			Date: created, Modified: modified, HTMLContent: "<p>Alpha body</p>",
			Tags: []string{"systems"}, Growth: "evergreen", OutLinks: []string{"beta"},
		},
		{
			Slug: "beta", Title: "Beta", Permalink: "/beta/",
			Date: created, HTMLContent: "<p>Beta body</p>",
		},
		{Slug: "section", Title: "Section", Permalink: "/section/", IsIndex: true},
	}
	resolver := content.NewLinkResolver(pages)
	graph, search, err := GraphSearch(pages, resolver, "/garden", true, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"id": "alpha"`, `"url": "/garden/alpha/"`,
		`"source": "alpha"`, `"target": "beta"`,
	} {
		if !strings.Contains(graph, want) {
			t.Errorf("graph missing %q: %s", want, graph)
		}
	}
	if !strings.HasSuffix(graph, "\n") || !strings.HasSuffix(search, "\n") {
		t.Error("JSON artifacts should retain the CLI encoder's trailing newline")
	}
	if strings.Contains(search, `"title": "Section"`) {
		t.Error("search index should exclude index pages")
	}

	sitemap := Sitemap(pages, "https://example.com/garden/")
	if !strings.Contains(sitemap, "https://example.com/garden/alpha/") ||
		!strings.Contains(sitemap, "<lastmod>2026-01-03</lastmod>") {
		t.Errorf("unexpected sitemap: %s", sitemap)
	}
	robots := Robots("https://example.com/garden/")
	if !strings.Contains(robots, "https://example.com/garden/sitemap.xml") {
		t.Errorf("unexpected robots.txt: %s", robots)
	}

	feed := RSS(
		pages,
		templates.SiteData{Title: "A & B", Author: "O'Reilly"},
		"https://example.com/garden/",
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	for _, want := range []string{
		"<title>A &amp; B</title>",
		"<description>O&apos;Reilly&apos;s digital garden</description>",
		"<title>Alpha &amp; One</title>",
		"https://example.com/garden/feed.xml",
	} {
		if !strings.Contains(feed, want) {
			t.Errorf("feed missing %q: %s", want, feed)
		}
	}
	if strings.Contains(feed, "<title>Section</title>") {
		t.Error("RSS should exclude index pages")
	}
}

func TestGraphDeduplicatesResolvedEdges(t *testing.T) {
	pages := []*content.Page{
		{Slug: "alpha", Title: "Alpha", OutLinks: []string{"beta", "Beta", "alpha"}},
		{Slug: "beta", Title: "Beta"},
	}
	graph, _, err := GraphSearch(pages, content.NewLinkResolver(pages), "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(graph, `"source": "alpha"`); got != 1 {
		t.Fatalf("alpha edge count = %d, want 1: %s", got, graph)
	}
	if strings.Contains(graph, `"target": "alpha"`) {
		t.Fatalf("graph contains a self edge: %s", graph)
	}
}

func TestXMLArtifactsEscapePageURLs(t *testing.T) {
	pages := []*content.Page{{
		Slug:        "r&d",
		Title:       "R&D",
		Permalink:   "/r&d/",
		HTMLContent: "<p>Research</p>",
	}}

	for name, document := range map[string]string{
		"sitemap": Sitemap(pages, "https://example.com"),
		"feed": RSS(
			pages,
			templates.SiteData{Title: "Garden"},
			"https://example.com",
			time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		),
	} {
		t.Run(name, func(t *testing.T) {
			var root struct {
				XMLName xml.Name
			}
			if err := xml.Unmarshal([]byte(document), &root); err != nil {
				t.Fatalf("artifact is invalid XML: %v\n%s", err, document)
			}
			if !strings.Contains(document, "/r&amp;d/") {
				t.Fatalf("artifact did not XML-escape page URL: %s", document)
			}
		})
	}
}

func TestRSSOrderingAndTruncationAreDeterministicAndUTF8Safe(t *testing.T) {
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	pages := []*content.Page{
		{Slug: "beta", Title: "Beta", Date: date, Permalink: "/beta/", HTMLContent: "<p>beta</p>"},
		{Slug: "alpha", Title: "Alpha", Date: date, Permalink: "/alpha/", HTMLContent: "<p>" + strings.Repeat("界", 301) + "</p>"},
	}
	feed := RSS(pages, templates.SiteData{Title: "Garden"}, "https://example.com", date)
	if strings.Index(feed, "<title>Alpha</title>") > strings.Index(feed, "<title>Beta</title>") {
		t.Fatalf("equal-date feed entries are not ordered by slug: %s", feed)
	}
	if strings.Contains(feed, "�") || !strings.Contains(feed, strings.Repeat("界", 300)+"...") {
		t.Fatal("RSS description was not truncated at a rune boundary")
	}
}

func TestFeedsEmitSectionAndTagFeeds(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	older := now.Add(-48 * time.Hour)
	pages := []*content.Page{
		{Slug: "about", Title: "About", Permalink: "/about/", Date: now, Tags: []string{"Meta"}},
		{Slug: "posts", Title: "Posts & Essays", Permalink: "/posts/", IsIndex: true},
		{Slug: "posts/hello", Title: "Hello", Permalink: "/posts/hello/", Date: now, Tags: []string{"meta", "go"}},
		{Slug: "posts/2026/deep", Title: "Deep", Permalink: "/posts/2026/deep/", Date: older},
		{Slug: "recipes/soup", Title: "Soup", Permalink: "/recipes/soup/", Date: older},
	}
	feeds := Feeds(pages, templates.SiteData{Title: "Garden"}, "https://example.com/g/", now)

	var paths []string
	byPath := make(map[string]string)
	for _, feed := range feeds {
		paths = append(paths, feed.Path)
		byPath[feed.Path] = feed.Content
	}
	wantPaths := []string{
		"feed.xml",
		"posts/feed.xml", "posts/2026/feed.xml", "recipes/feed.xml",
		"tags/go/feed.xml", "tags/meta/feed.xml",
	}
	if strings.Join(paths, "|") != strings.Join(wantPaths, "|") {
		t.Fatalf("feed paths = %v, want %v", paths, wantPaths)
	}

	for path, document := range byPath {
		var root struct{ XMLName xml.Name }
		if err := xml.Unmarshal([]byte(document), &root); err != nil {
			t.Errorf("%s is invalid XML: %v\n%s", path, err, document)
		}
	}

	posts := byPath["posts/feed.xml"]
	for _, want := range []string{
		"<title>Posts &amp; Essays | Garden</title>",
		"<link>https://example.com/g/posts/</link>",
		`<atom:link href="https://example.com/g/posts/feed.xml" rel="self"`,
		"<description>Posts &amp; Essays - Garden</description>",
		"<title>Hello</title>",
		"<title>Deep</title>",
	} {
		if !strings.Contains(posts, want) {
			t.Errorf("posts feed missing %q: %s", want, posts)
		}
	}
	for _, absent := range []string{"<title>About</title>", "<title>Soup</title>", "<title>Posts &amp; Essays</title>\n"} {
		if strings.Contains(posts, absent) {
			t.Errorf("posts feed should not contain %q: %s", absent, posts)
		}
	}
	if strings.Index(posts, "<title>Hello</title>") > strings.Index(posts, "<title>Deep</title>") {
		t.Error("section feed should list newest pages first")
	}

	recipes := byPath["recipes/feed.xml"]
	if !strings.Contains(recipes, "<title>Recipes | Garden</title>") || !strings.Contains(recipes, "<title>Soup</title>") {
		t.Errorf("auto-titled section feed is wrong: %s", recipes)
	}

	meta := byPath["tags/meta/feed.xml"]
	for _, want := range []string{
		"<title>#meta | Garden</title>",
		"<link>https://example.com/g/tags/meta/</link>",
		`<atom:link href="https://example.com/g/tags/meta/feed.xml" rel="self"`,
		"<description>Pages tagged with #meta - Garden</description>",
		"<title>About</title>",
		"<title>Hello</title>",
	} {
		if !strings.Contains(meta, want) {
			t.Errorf("tag feed missing %q: %s", want, meta)
		}
	}
	if strings.Contains(meta, "<title>Deep</title>") {
		t.Errorf("tag feed should only list tagged pages: %s", meta)
	}
	if strings.Contains(byPath["feed.xml"], "<title>Posts &amp; Essays</title>") {
		t.Error("global feed should still exclude index pages")
	}
}

func TestStylesMatchesCLIComposition(t *testing.T) {
	got := Styles("", config.Default().Theme)
	baseAt := strings.Index(got, "/* leafpress Base Styles */")
	classicAt := strings.Index(got, "/* leafpress Classic Theme */")
	fontsAt := strings.Index(got, "/* Self-hosted fonts */")
	if baseAt != 0 || classicAt <= baseAt || fontsAt <= classicAt {
		t.Errorf("stylesheet order = base:%d classic:%d fonts:%d", baseAt, classicAt, fontsAt)
	}
	got = Styles("body { outline: none; }", config.Default().Theme)
	if !strings.HasSuffix(got, "\n\n/* User Styles */\nbody { outline: none; }") {
		t.Error("user CSS should use the CLI composition marker and ordering")
	}
}

func TestStylesDefaultsEmptyPresetToClassic(t *testing.T) {
	theme := config.Default().Theme
	want := Styles("", theme)
	theme.Preset = ""
	if got := Styles("", theme); got != want {
		t.Error("empty preset did not preserve classic stylesheet compatibility")
	}
}

func TestStylesComposesAuroraBeforeFontsAndUserCSS(t *testing.T) {
	theme, err := config.Parse([]byte(`{"theme":{"preset":"aurora"}}`))
	if err != nil {
		t.Fatalf("parse aurora config: %v", err)
	}
	got := Styles(".custom { color: hotpink; }", theme.Theme)
	baseAt := strings.Index(got, "/* leafpress Base Styles */")
	classicAt := strings.Index(got, "/* leafpress Classic Theme */")
	auroraAt := strings.Index(got, "/* leafpress Aurora Theme */")
	fontsAt := strings.Index(got, "/* Self-hosted fonts */")
	userAt := strings.Index(got, "/* User Styles */")
	if baseAt != 0 || classicAt <= baseAt || auroraAt <= classicAt ||
		fontsAt <= auroraAt || userAt <= fontsAt {
		t.Errorf("stylesheet order = base:%d classic:%d aurora:%d fonts:%d user:%d",
			baseAt, classicAt, auroraAt, fontsAt, userAt)
	}
}

func TestStylesComposesPaperBeforeFontsAndUserCSS(t *testing.T) {
	theme, err := config.Parse([]byte(`{"theme":{"preset":"paper"}}`))
	if err != nil {
		t.Fatalf("parse paper config: %v", err)
	}
	got := Styles(".custom { color: hotpink; }", theme.Theme)
	baseAt := strings.Index(got, "/* leafpress Base Styles */")
	classicAt := strings.Index(got, "/* leafpress Classic Theme */")
	paperAt := strings.Index(got, "/* leafpress Paper Theme */")
	fontsAt := strings.Index(got, "/* Self-hosted fonts */")
	userAt := strings.Index(got, "/* User Styles */")
	if baseAt != 0 || classicAt <= baseAt || paperAt <= classicAt ||
		fontsAt <= paperAt || userAt <= fontsAt {
		t.Errorf("stylesheet order = base:%d classic:%d paper:%d fonts:%d user:%d",
			baseAt, classicAt, paperAt, fontsAt, userAt)
	}
}

func TestStylesComposesTerminalBeforeFontsAndUserCSS(t *testing.T) {
	theme, err := config.Parse([]byte(`{"theme":{"preset":"terminal"}}`))
	if err != nil {
		t.Fatalf("parse terminal config: %v", err)
	}
	got := Styles(".custom { color: hotpink; }", theme.Theme)
	baseAt := strings.Index(got, "/* leafpress Base Styles */")
	classicAt := strings.Index(got, "/* leafpress Classic Theme */")
	terminalAt := strings.Index(got, "/* leafpress Terminal Theme */")
	fontsAt := strings.Index(got, "/* Self-hosted fonts */")
	userAt := strings.Index(got, "/* User Styles */")
	if baseAt != 0 || classicAt <= baseAt || terminalAt <= classicAt ||
		fontsAt <= terminalAt || userAt <= fontsAt {
		t.Errorf("stylesheet order = base:%d classic:%d terminal:%d fonts:%d user:%d",
			baseAt, classicAt, terminalAt, fontsAt, userAt)
	}
}
