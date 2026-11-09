package content

// Publishable applies the publication and draft gates after path exclusions.
// Explicit publication is required for authored index pages too. Draft previews
// never override publication restrictions. Unknown modes fail closed.
func Publishable(page *Page, mode string, includeDrafts bool) bool {
	if page.Draft && !includeDrafts {
		return false
	}
	switch mode {
	case "explicit":
		return page.Publish != nil && *page.Publish
	case "all", "":
		return page.Publish == nil || *page.Publish
	default:
		return false
	}
}
