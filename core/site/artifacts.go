// Package site generates the non-page artifacts of a Leafpress site without
// touching the filesystem. The CLI writes these bytes to disk; embedders such
// as leafpress-render return the same bytes through their own transport.
package site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shivamx96/leafpress/core/config"
	"github.com/shivamx96/leafpress/core/content"
	"github.com/shivamx96/leafpress/core/templates"
)

// GraphNode is the public graph.json node shape.
type GraphNode struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	URL    string   `json:"url"`
	Growth string   `json:"growth,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

// GraphEdge is the public graph.json edge shape.
type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// Graph is the public graph.json shape.
type Graph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// SearchEntry is one public search-index.json record.
type SearchEntry struct {
	Title   string   `json:"title"`
	URL     string   `json:"url"`
	Content string   `json:"content"`
	Tags    []string `json:"tags,omitempty"`
}

// GraphSearch returns the exact JSON artifacts used by the CLI. A disabled
// artifact is returned as an empty string.
func GraphSearch(
	pages []*content.Page,
	resolver *content.LinkResolver,
	basePath string,
	genGraph bool,
	genSearch bool,
) (graphJSON string, searchJSON string, err error) {
	if !genGraph && !genSearch {
		return "", "", nil
	}

	var graph Graph
	var searchIndex []SearchEntry
	for _, page := range pages {
		if genGraph {
			graph.Nodes = append(graph.Nodes, GraphNode{
				ID:     page.Slug,
				Title:  page.Title,
				URL:    basePath + page.Permalink,
				Growth: page.Growth,
				Tags:   page.Tags,
			})
			for _, target := range page.OutLinks {
				if result := resolver.Resolve(target); result.Page != nil {
					graph.Edges = append(graph.Edges, GraphEdge{
						Source: page.Slug,
						Target: result.Page.Slug,
					})
				}
			}
		}

		if genSearch && !page.IsIndex {
			searchIndex = append(searchIndex, SearchEntry{
				Title:   page.Title,
				URL:     basePath + page.Permalink,
				Content: page.PlainContent(),
				Tags:    page.Tags,
			})
		}
	}

	if genGraph {
		graphJSON, err = encodeJSON(graph)
		if err != nil {
			return "", "", err
		}
	}
	if genSearch {
		searchJSON, err = encodeJSON(searchIndex)
		if err != nil {
			return "", "", err
		}
	}
	return graphJSON, searchJSON, nil
}

// Robots returns robots.txt with the canonical sitemap URL when available.
func Robots(baseURL string) string {
	if baseURL != "" {
		return fmt.Sprintf(
			"User-agent: *\nAllow: /\n\nSitemap: %s/sitemap.xml\n",
			strings.TrimSuffix(baseURL, "/"),
		)
	}
	return "User-agent: *\nAllow: /\n"
}

// Sitemap returns sitemap.xml using the CLI's page and last-modified rules.
func Sitemap(pages []*content.Page, baseURL string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	sb.WriteString("\n")
	for _, page := range pages {
		loc := page.Permalink
		if baseURL != "" {
			loc = baseURL + page.Permalink
		}
		var lastmod string
		if page.HasModified() {
			lastmod = page.Modified.Format("2006-01-02")
		} else if !page.Date.IsZero() {
			lastmod = page.Date.Format("2006-01-02")
		}
		sb.WriteString("  <url>\n")
		sb.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", loc))
		if lastmod != "" {
			sb.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", lastmod))
		}
		sb.WriteString("  </url>\n")
	}
	sb.WriteString("</urlset>\n")
	return sb.String()
}

// RSS returns feed.xml. now is used only when no page supplies a date; pass a
// zero value for normal wall-clock behavior or a fixed time in tests.
func RSS(pages []*content.Page, siteData templates.SiteData, baseURL string, now time.Time) string {
	baseURL = strings.TrimSuffix(baseURL, "/")
	feedPages := make([]*content.Page, 0, len(pages))
	for _, page := range pages {
		if !page.IsIndex {
			feedPages = append(feedPages, page)
		}
	}
	sort.Slice(feedPages, func(i, j int) bool {
		return effectiveDate(feedPages[i]).After(effectiveDate(feedPages[j]))
	})
	if len(feedPages) > 20 {
		feedPages = feedPages[:20]
	}
	if now.IsZero() {
		now = time.Now()
	}
	lastBuild := now.Format(time.RFC1123Z)
	if len(feedPages) > 0 && !effectiveDate(feedPages[0]).IsZero() {
		lastBuild = effectiveDate(feedPages[0]).Format(time.RFC1123Z)
	}

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(`<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">`)
	sb.WriteString("\n  <channel>\n")
	sb.WriteString(fmt.Sprintf("    <title>%s</title>\n", escapeXML(siteData.Title)))
	if baseURL != "" {
		sb.WriteString(fmt.Sprintf("    <link>%s</link>\n", baseURL))
		sb.WriteString(fmt.Sprintf(
			"    <atom:link href=\"%s/feed.xml\" rel=\"self\" type=\"application/rss+xml\"/>\n",
			baseURL,
		))
	}
	if siteData.Author != "" {
		sb.WriteString(fmt.Sprintf(
			"    <description>%s's digital garden</description>\n",
			escapeXML(siteData.Author),
		))
	} else {
		sb.WriteString(fmt.Sprintf("    <description>%s</description>\n", escapeXML(siteData.Title)))
	}
	sb.WriteString(fmt.Sprintf("    <lastBuildDate>%s</lastBuildDate>\n", lastBuild))
	sb.WriteString("    <generator>leafpress</generator>\n")
	for _, page := range feedPages {
		link := page.Permalink
		if baseURL != "" {
			link = baseURL + page.Permalink
		}
		sb.WriteString("    <item>\n")
		sb.WriteString(fmt.Sprintf("      <title>%s</title>\n", escapeXML(page.Title)))
		sb.WriteString(fmt.Sprintf("      <link>%s</link>\n", link))
		sb.WriteString(fmt.Sprintf("      <guid>%s</guid>\n", link))
		if date := effectiveDate(page); !date.IsZero() {
			sb.WriteString(fmt.Sprintf("      <pubDate>%s</pubDate>\n", date.Format(time.RFC1123Z)))
		}
		description := page.PlainContent()
		if len(description) > 300 {
			description = description[:300] + "..."
		}
		if description != "" {
			sb.WriteString(fmt.Sprintf(
				"      <description>%s</description>\n",
				escapeXML(description),
			))
		}
		sb.WriteString("    </item>\n")
	}
	sb.WriteString("  </channel>\n</rss>\n")
	return sb.String()
}

// NotFound renders the shared 404 page template.
func NotFound(tmpl *templates.Templates, siteData templates.SiteData) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.RenderNotFound(&buf, templates.NotFoundData{Site: siteData}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Styles combines Leafpress's embedded stylesheet, the theme's self-hosted
// @font-face rules, and the optional contents of the CLI's style.css (or the
// renderer's styleCSS input). Font rules live here rather than in each page
// head so they are downloaded once and cached; their URLs are site-relative,
// which resolves correctly because this stylesheet is served from the site
// root.
func Styles(userCSS string, theme config.Theme) string {
	css := templates.DefaultCSS
	if fontCSS := templates.FontCSS(theme); fontCSS != "" {
		css += "\n\n/* Self-hosted fonts */\n" + fontCSS
	}
	if userCSS != "" {
		css += "\n\n/* User Styles */\n" + userCSS
	}
	return css
}

func effectiveDate(page *content.Page) time.Time {
	if page.HasModified() {
		return page.Modified
	}
	return page.Date
}

func encodeJSON(value any) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
