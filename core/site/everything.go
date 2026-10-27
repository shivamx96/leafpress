package site

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
	"github.com/shivamx96/leafpress/core/templates"
)

// ValidateEverythingRoutes prevents a generated index from overwriting content
// or a section home. Reserve the subtree only when the feature is enabled.
func ValidateEverythingRoutes(pages []*content.Page, nav config.Navigation) error {
	if !nav.Everything {
		return nil
	}
	for _, page := range pages {
		if page.Slug == "everything" || strings.HasPrefix(page.Slug, "everything/") {
			return fmt.Errorf("navigation.everything reserves /everything/; move content %q or disable it", page.Slug)
		}
	}
	return nil
}

// EverythingPages lists non-index pages across all sections, most recent first.
// The caller supplies the same public page set used by the rest of the site.
// Exclusions affect this listing only; they do not make a page private.
func EverythingPages(pages []*content.Page, nav config.Navigation) []*content.Page {
	result := make([]*content.Page, 0, len(pages))
	for _, page := range pages {
		if page.IsIndex {
			continue
		}
		excluded := false
		for _, path := range nav.EverythingExclude {
			prefix := strings.TrimRight(path, "/")
			slug := "/" + page.Slug
			if slug == prefix || strings.HasPrefix(slug, prefix+"/") {
				excluded = true
				break
			}
		}
		if !excluded {
			result = append(result, page)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Date.Equal(result[j].Date) {
			return result[i].Slug < result[j].Slug
		}
		return result[i].Date.After(result[j].Date)
	})
	return result
}

// EverythingIndex is shared by native builds and hosted renderer output.
func EverythingIndex(tmpl *templates.Templates, site templates.SiteData, pages []*content.Page, nav config.Navigation) (string, error) {
	var buf bytes.Buffer
	err := tmpl.RenderIndex(&buf, templates.IndexData{Site: site, Title: "Everything", Pages: EverythingPages(pages, nav), ShowList: true, CurrentPath: "/everything/"})
	return buf.String(), err
}
