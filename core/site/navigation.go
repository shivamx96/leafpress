package site

import (
	"path"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
)

// BuildNavigation resolves the site nav bar from the configured mode. It is the
// single source of navigation for both the CLI build and the renderer, so the
// two surfaces produce identical nav from identical config.
//
//   - explicit: the configured items verbatim.
//   - automatic: derived from the garden's public top level (root notes and
//     section homes), optionally appending a Tags item.
func BuildNavigation(pages []*content.Page, nav config.Navigation) []config.NavItem {
	if nav.Mode == config.NavExplicit {
		return nav.Items
	}
	items := navItems(TopLevelEntries(pages))
	if nav.IncludeTags && pagesHaveTags(pages) {
		items = append(items, config.NavItem{Label: "Tags", Path: "/tags/"})
	}
	return items
}

// GroupSections partitions pages into per-section children and per-section
// index pages, mirroring the CLI's filepath.Dir-based grouping. An index
// page's slug is its section path ("" is the garden home).
func GroupSections(pages []*content.Page) (children map[string][]*content.Page, indexBySection map[string]*content.Page) {
	children = make(map[string][]*content.Page)
	indexBySection = make(map[string]*content.Page)
	for _, p := range pages {
		if p.IsIndex {
			indexBySection[p.Slug] = p
		} else {
			children[sectionOf(p.Slug)] = append(children[sectionOf(p.Slug)], p)
		}
	}
	return children, indexBySection
}

// TopLevelEntries returns the garden's public top level: root notes plus one
// entry per section. Nested notes belong to their section home and never appear
// directly at the root. This drives both the home listing and automatic nav.
func TopLevelEntries(pages []*content.Page) []*content.Page {
	return RootEntries(GroupSections(pages))
}

// RootEntries returns root notes plus one entry per section, given a section
// grouping (see GroupSections).
func RootEntries(children map[string][]*content.Page, indexBySection map[string]*content.Page) []*content.Page {
	sectionDirs := make([]string, 0, len(children)+len(indexBySection))
	for dir := range children {
		if dir != "" {
			sectionDirs = append(sectionDirs, dir)
		}
	}
	for dir := range indexBySection {
		if dir != "" && children[dir] == nil {
			sectionDirs = append(sectionDirs, dir) // index page of an empty section
		}
	}
	sort.Strings(sectionDirs)

	entries := make([]*content.Page, 0, len(children[""])+len(sectionDirs))
	entries = append(entries, children[""]...)
	for _, dir := range sectionDirs {
		if p := indexBySection[dir]; p != nil {
			entries = append(entries, p)
		} else {
			entries = append(entries, autoSectionEntry(dir, children[dir]))
		}
	}
	return entries
}

// TitleCase title-cases an auto-index section name, matching the CLI's
// generateAutoIndexes. Hyphens read as word separators ("field-notes" →
// "Field Notes").
func TitleCase(s string) string {
	return cases.Title(language.English).String(strings.ReplaceAll(s, "-", " "))
}

// HasOrigin reports whether baseURL supplies an absolute origin. Artifacts that
// require absolute URLs (sitemap <loc>, the robots Sitemap directive, RSS
// links) must be skipped when this is false, rather than emitted with invalid
// relative values.
func HasOrigin(baseURL string) bool {
	return strings.TrimSpace(baseURL) != ""
}

func navItems(entries []*content.Page) []config.NavItem {
	nav := make([]config.NavItem, 0, len(entries))
	for _, entry := range entries {
		// The garden home is reached via the brand title, not a nav link.
		if entry.Slug == "" || entry.Permalink == "/" {
			continue
		}
		nav = append(nav, config.NavItem{Label: entry.Title, Path: entry.Permalink})
	}
	return nav
}

func pagesHaveTags(pages []*content.Page) bool {
	for _, page := range pages {
		if len(page.Tags) > 0 {
			return true
		}
	}
	return false
}

// sectionOf returns the section path a slug belongs to ("" for root),
// mirroring the CLI's filepath.Dir-based grouping.
func sectionOf(slug string) string {
	dir := path.Dir(slug)
	if dir == "." {
		return ""
	}
	return dir
}

// autoSectionEntry is the single homepage/nav link for a section that has no
// explicit index page. The section's real children remain on its generated
// section home instead of being flattened into the garden home.
func autoSectionEntry(slug string, children []*content.Page) *content.Page {
	entry := &content.Page{
		Title:     TitleCase(path.Base(slug)),
		Slug:      slug,
		Permalink: "/" + slug + "/",
		IsIndex:   true,
	}

	// Give the section the date of its most recently changed child so a
	// date-sorted homepage places the folder alongside current root notes.
	for _, child := range children {
		date := child.Date
		if child.HasModified() {
			date = child.Modified
		}
		if date.After(entry.Date) {
			entry.Date = date
		}
	}
	return entry
}
