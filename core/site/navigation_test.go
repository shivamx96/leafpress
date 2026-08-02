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
