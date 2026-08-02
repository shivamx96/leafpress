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
		"<description>O&apos;Reilly's digital garden</description>",
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

func TestStylesMatchesCLIComposition(t *testing.T) {
	if got := Styles("", config.Default().Theme); !strings.HasPrefix(got, templates.DefaultCSS) || !strings.Contains(got, "@font-face") {
		t.Error("empty user CSS should return the embedded stylesheet exactly")
	}
	got := Styles("body { outline: none; }", config.Default().Theme)
	if !strings.HasSuffix(got, "\n\n/* User Styles */\nbody { outline: none; }") {
		t.Error("user CSS should use the CLI composition marker and ordering")
	}
}
