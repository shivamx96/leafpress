package site

import (
	"reflect"
	"testing"
	"time"

	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
)

func TestBuildNavigationModes(t *testing.T) {
	pages := []*content.Page{
		{Slug: "", Permalink: "/", Title: "Home", IsIndex: true},
		{Slug: "now", Permalink: "/now/", Title: "Now", Tags: []string{"notes"}},
		{Slug: "field-notes/one", Permalink: "/field-notes/one/", Title: "One"},
		{Slug: "essays", Permalink: "/essays/", Title: "Writing", IsIndex: true},
		{Slug: "essays/two", Permalink: "/essays/two/", Title: "Two"},
	}
	wantAutomatic := []config.NavItem{
		{Label: "Now", Path: "/now/"},
		{Label: "Writing", Path: "/essays/"},
		{Label: "Field Notes", Path: "/field-notes/"},
		{Label: "Tags", Path: "/tags/"},
	}
	got := BuildNavigation(pages, config.Navigation{Mode: config.NavAutomatic, IncludeTags: true})
	if !reflect.DeepEqual(got, wantAutomatic) {
		t.Fatalf("automatic nav = %#v, want %#v", got, wantAutomatic)
	}

	explicit := []config.NavItem{{Label: "Only", Path: "/only/"}}
	got = BuildNavigation(pages, config.Navigation{Mode: config.NavExplicit, Items: explicit, IncludeTags: true})
	if !reflect.DeepEqual(got, explicit) {
		t.Fatalf("explicit nav = %#v, want %#v", got, explicit)
	}
}

func TestAutoSectionEntryUsesNewestChildDate(t *testing.T) {
	pages := []*content.Page{
		{Slug: "notes/old", Permalink: "/notes/old/", Date: mustDate(t, "2026-01-01")},
		{Slug: "notes/new", Permalink: "/notes/new/", Date: mustDate(t, "2026-02-01")},
	}
	entries := TopLevelEntries(pages)
	if len(entries) != 1 || entries[0].Slug != "notes" || entries[0].Title != "Notes" {
		t.Fatalf("entries = %#v", entries)
	}
	if !entries[0].Date.Equal(pages[1].Date) {
		t.Fatalf("section date = %v, want %v", entries[0].Date, pages[1].Date)
	}
}

func mustDate(t *testing.T, value string) time.Time {
	t.Helper()
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatal(err)
	}
	return date
}

func TestSectionTitleUsesFolderNameAsWritten(t *testing.T) {
	members := []*content.Page{{SourcePath: "Q&A/first.md", Slug: "Q-A/first"}}
	if got := SectionTitle("Q-A", members); got != "Q&amp;A" {
		t.Errorf("SectionTitle(Q-A) = %q, want escaped folder name", got)
	}
	if got := SectionTitle("b", []*content.Page{{SourcePath: "<b>/x.md", Slug: "b/x"}}); got != "&lt;B&gt;" {
		t.Errorf("folder name reached markup unescaped: %q", got)
	}
	// Renderer pages carry no source path: fall back to the slug.
	if got := SectionTitle("essays/field-notes", []*content.Page{{Slug: "essays/field-notes/x"}}); got != "Field Notes" {
		t.Errorf("slug fallback = %q, want Field Notes", got)
	}

	nav := BuildNavigation([]*content.Page{
		{SourcePath: "Q&A/first.md", Slug: "Q-A/first", Permalink: "/Q-A/first/"},
	}, config.Navigation{Mode: config.NavAutomatic})
	if len(nav) != 1 || nav[0].Label != "Q&amp;A" || nav[0].Path != "/Q-A/" {
		t.Fatalf("automatic nav = %+v", nav)
	}
}

func TestGroupSectionsDeepHierarchy(t *testing.T) {
	pages := []*content.Page{
		{Slug: "clients/acme/projects/plan", SourcePath: "Clients & Co/acme/projects/plan.md", Date: mustDate(t, "2026-01-01"), Modified: mustDate(t, "2026-03-01")},
		{Slug: "clients/birch/projects/plan", Date: mustDate(t, "2026-02-01")},
		{Slug: "now"},
		{Slug: "archive/empty/nested", Title: "Empty", IsIndex: true},
	}
	children, indexes := GroupSections(pages)
	want := map[string][]string{
		"":                       {"now", "archive", "clients"},
		"clients":                {"clients/acme", "clients/birch"},
		"clients/acme":           {"clients/acme/projects"},
		"clients/acme/projects":  {"clients/acme/projects/plan"},
		"clients/birch":          {"clients/birch/projects"},
		"clients/birch/projects": {"clients/birch/projects/plan"},
		"archive":                {"archive/empty"},
		"archive/empty":          {"archive/empty/nested"},
		"archive/empty/nested":   nil,
	}
	got := make(map[string][]string)
	for section, entries := range children {
		got[section] = nil
		for _, p := range entries {
			got[section] = append(got[section], p.Slug)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("children = %#v, want %#v", got, want)
	}
	if len(indexes) != 1 || indexes["archive/empty/nested"] != pages[3] {
		t.Fatalf("authored indexes = %v", indexes)
	}
	clients := children[""][2]
	if clients.Title != "Clients &amp; Co" || !clients.Date.Equal(pages[0].Modified) {
		t.Fatalf("ancestor metadata = %+v", clients)
	}
	if len(pages) != 4 {
		t.Fatal("grouping mutated authored pages")
	}
}
