package site

import (
	"html"

	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
	"github.com/shivamx96/leafpress/core/templates"
)

// Templates render with text/template, so every author-supplied string that
// lands in an attribute or element must arrive pre-escaped. Both interfaces
// escape at their trust boundary and un-escape again for the artifacts that
// carry their own encoding (RSS/XML, JSON). The helpers live here so the CLI
// and the embedded renderer cannot drift: a quote in site.description is a
// broken og:description tag, not only an XSS vector.

// SafeSiteData escapes the site fields that reach markup. HeadExtra
// deliberately stays raw: it is the documented escape hatch for trusted
// operator HTML. Nav is escaped separately by EscapeNavItems, because it is
// assembled after this call from config items and page titles.
func SafeSiteData(site templates.SiteData) templates.SiteData {
	site.Title = html.EscapeString(site.Title)
	site.Description = html.EscapeString(site.Description)
	site.Author = html.EscapeString(site.Author)
	site.BaseURL = html.EscapeString(site.BaseURL)
	site.Image = html.EscapeString(site.Image)
	if site.FooterAttribution != nil {
		attribution := *site.FooterAttribution
		attribution.Name = html.EscapeString(attribution.Name)
		attribution.URL = html.EscapeString(attribution.URL)
		site.FooterAttribution = &attribution
	}
	return site
}

// RawSiteData reverses SafeSiteData for artifacts that do their own encoding.
// baseURL is restored from config rather than un-escaped, so the canonical
// origin is used verbatim in feed and sitemap URLs.
func RawSiteData(site templates.SiteData, baseURL string) templates.SiteData {
	site.Title = html.UnescapeString(site.Title)
	site.Description = html.UnescapeString(site.Description)
	site.Author = html.UnescapeString(site.Author)
	site.Image = html.UnescapeString(site.Image)
	site.BaseURL = baseURL
	return site
}

// EscapeNavItems escapes explicit navigation entries in place, at the config
// trust boundary. Automatic navigation needs no pass here: it is assembled
// from page titles that EscapePageMeta has already escaped.
func EscapeNavItems(nav *config.Navigation) {
	if nav == nil {
		return
	}
	for i := range nav.Items {
		nav.Items[i].Label = html.EscapeString(nav.Items[i].Label)
		nav.Items[i].Path = html.EscapeString(nav.Items[i].Path)
	}
}

// EscapePageMeta escapes the page metadata that reaches markup, in place.
// Body HTML is not touched: it is already-rendered Markdown output.
//
// Call this before building a LinkResolver. NewLinkResolver registers both
// the escaped and the raw spelling of a title as aliases, so a page titled
// "Q&A" still resolves from [[Q&A]].
func EscapePageMeta(pages []*content.Page) {
	for _, page := range pages {
		if page == nil {
			continue
		}
		page.Title = html.EscapeString(page.Title)
		page.Description = html.EscapeString(page.Description)
		page.Image = html.EscapeString(page.Image)
	}
}

// RawPages returns shallow copies with page metadata un-escaped, for feeds,
// sitemaps and JSON indexes. The originals keep their escaped values for
// template rendering.
func RawPages(pages []*content.Page) []*content.Page {
	raw := make([]*content.Page, 0, len(pages))
	for _, page := range pages {
		if page == nil {
			continue
		}
		copied := *page
		copied.Title = html.UnescapeString(copied.Title)
		copied.Description = html.UnescapeString(copied.Description)
		copied.Image = html.UnescapeString(copied.Image)
		raw = append(raw, &copied)
	}
	return raw
}
