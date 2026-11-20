package site

import (
	"html"
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

// GroupSections builds immediate-child listings for every section, including
// ancestors with no direct notes. indexBySection contains only authored indexes;
// synthetic entries live exclusively in listings, never in the authored pages.
func GroupSections(pages []*content.Page) (children map[string][]*content.Page, indexBySection map[string]*content.Page) {
	children = make(map[string][]*content.Page)
	indexBySection = make(map[string]*content.Page)
	// Keep one authored descendant for each folder's original display name.
	representatives := make(map[string]*content.Page)
	for _, p := range pages {
		dir := p.Slug
		if p.IsIndex {
			indexBySection[dir] = p
		} else {
			dir = sectionOf(p.Slug)
			children[dir] = append(children[dir], p)
		}
		for ; dir != ""; dir = sectionOf(dir) {
			if _, exists := representatives[dir]; !exists {
				representatives[dir] = p
			}
			if _, exists := children[dir]; !exists {
				children[dir] = nil
			}
		}
	}
	// Descendants precede parents so synthetic dates propagate all the way up.
	dirs := make([]string, 0, len(representatives))
	for dir := range representatives {
		dirs = append(dirs, dir)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	for _, dir := range dirs {
		entry := indexBySection[dir]
		if entry == nil {
			entry = autoSectionEntry(dir, children[dir])
			entry.Title = SectionTitle(dir, []*content.Page{representatives[dir]})
		}
		parent := sectionOf(dir)
		children[parent] = append(children[parent], entry)
	}
	// Retain note input order and the existing alphabetical folder ordering.
	for _, entries := range children {
		sort.SliceStable(entries, func(i, j int) bool {
			if entries[i].IsIndex != entries[j].IsIndex {
				return !entries[i].IsIndex
			}
			return entries[i].IsIndex && entries[i].Slug < entries[j].Slug
		})
	}
	return children, indexBySection
}

// TopLevelEntries returns root notes and immediate top-level folders.
func TopLevelEntries(pages []*content.Page) []*content.Page {
	return RootEntries(GroupSections(pages))
}

// RootEntries returns the home listing from a GroupSections projection.
// The returned slice can be sorted without changing the grouping.
func RootEntries(children map[string][]*content.Page, _ map[string]*content.Page) []*content.Page {
	return append([]*content.Page(nil), children[""]...)
}

// TitleCase title-cases an auto-index section name, matching the CLI's
// generateAutoIndexes. Hyphens read as word separators ("field-notes" →
// "Field Notes").
func TitleCase(s string) string {
	return cases.Title(language.English).String(strings.ReplaceAll(s, "-", " "))
}

// SectionTitle is the display title of a section without an explicit index
// page: its folder name as the author wrote it ("Q&A", not the "Q-A" URL),
// title-cased by TitleCase. members are pages inside the section; renderer
// pages carry no source path, so the slug's last segment is the fallback. The
// result is HTML-escaped like page titles, because folder names reach markup.
func SectionTitle(section string, members []*content.Page) string {
	return html.EscapeString(sectionTitle(section, members))
}

// sectionTitle is SectionTitle before HTML escaping, for artifacts such as
// RSS that apply their own encoding.
func sectionTitle(section string, members []*content.Page) string {
	name := path.Base(section)
	for _, member := range members {
		if folder, ok := content.FolderName(section, member); ok {
			name = folder
			break
		}
	}
	return TitleCase(name)
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

// autoSectionEntry represents a generated section in its parent listing.
func autoSectionEntry(slug string, children []*content.Page) *content.Page {
	entry := &content.Page{
		Title:     SectionTitle(slug, children),
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
