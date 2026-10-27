package site

import (
	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
	"testing"
	"time"
)

func TestEverythingSelection(t *testing.T) {
	pages := []*content.Page{
		{Slug: "notes", IsIndex: true}, {Slug: "about"}, {Slug: "notes/old"},
		{Slug: "notes/new", Date: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Slug: "private-list/note"}, {Slug: "about-writing"},
	}
	nav := config.Navigation{Everything: true, EverythingExclude: []string{"/about/", "/private-list/"}}
	got := EverythingPages(pages, nav)
	if len(got) != 3 || got[0].Slug != "notes/new" || got[1].Slug != "about-writing" || got[2].Slug != "notes/old" {
		t.Fatalf("unexpected listing: %+v", got)
	}
	if pages[0].Slug != "notes" {
		t.Fatal("mutated input order")
	}
	for _, slug := range []string{"everything", "everything/child"} {
		if ValidateEverythingRoutes([]*content.Page{{Slug: slug}}, nav) == nil {
			t.Fatalf("accepted collision %s", slug)
		}
		if ValidateEverythingRoutes([]*content.Page{{Slug: slug}}, config.Navigation{}) != nil {
			t.Fatal("reserved disabled route")
		}
	}
}
