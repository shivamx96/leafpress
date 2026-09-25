package site

import "github.com/shivamx96/leafpress/core/content"

// ApplyHomeTitle gives an untitled home page the site title, in place. The
// scanner cannot derive a title for the site root from a filename, and the
// embedded renderer already falls back to the garden title. Call this before
// EscapePageMeta, with the raw site title.
func ApplyHomeTitle(pages []*content.Page, siteTitle string) {
	for _, page := range pages {
		if page != nil && page.Slug == "" && page.Title == "" {
			page.Title = siteTitle
		}
	}
}
