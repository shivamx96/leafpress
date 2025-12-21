package content

import (
	"regexp"
	"strings"
)

// WikiLink represents a parsed wiki-link
type WikiLink struct {
	Target string // The link target (slug or path)
	Label  string // Display label (defaults to target)
	Raw    string // Original raw text including brackets
}

// wikiLinkRegex matches [[target]] or [[target|label]]
var wikiLinkRegex = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)

// ExtractWikiLinks extracts all wiki-links from content
func ExtractWikiLinks(content string) []WikiLink {
	matches := wikiLinkRegex.FindAllStringSubmatch(content, -1)
	var links []WikiLink

	for _, match := range matches {
		target := strings.TrimSpace(match[1])
		label := target
		if len(match) > 2 && match[2] != "" {
			label = strings.TrimSpace(match[2])
		}

		links = append(links, WikiLink{
			Target: target,
			Label:  label,
			Raw:    match[0],
		})
	}

	return links
}

// LinkResolver resolves wiki-links to actual pages
type LinkResolver struct {
	pages   []*Page
	slugMap map[string]*Page   // Exact slug -> page
	nameMap map[string][]*Page // Filename -> pages (may have duplicates)
}

// NewLinkResolver creates a new link resolver
func NewLinkResolver(pages []*Page) *LinkResolver {
	resolver := &LinkResolver{
		pages:   pages,
		slugMap: make(map[string]*Page),
		nameMap: make(map[string][]*Page),
	}

	for _, page := range pages {
		// Map by exact slug (lowercase)
		slugLower := strings.ToLower(page.Slug)
		resolver.slugMap[slugLower] = page

		// Map by filename (lowercase)
		parts := strings.Split(page.Slug, "/")
		name := strings.ToLower(parts[len(parts)-1])
		resolver.nameMap[name] = append(resolver.nameMap[name], page)
	}

	return resolver
}

// ResolveResult represents the result of resolving a wiki-link
type ResolveResult struct {
	Page      *Page
	Ambiguous bool
	Broken    bool
}

// Resolve resolves a wiki-link target to a page
func (r *LinkResolver) Resolve(target string) ResolveResult {
	targetLower := strings.ToLower(target)

	// 1. Exact slug match
	if page, ok := r.slugMap[targetLower]; ok {
		return ResolveResult{Page: page}
	}

	// 2. Filename match anywhere
	if pages, ok := r.nameMap[targetLower]; ok {
		if len(pages) == 1 {
			return ResolveResult{Page: pages[0]}
		}
		if len(pages) > 1 {
			// Ambiguous - return first alphabetically (already sorted by slug)
			return ResolveResult{Page: pages[0], Ambiguous: true}
		}
	}

	// 3. Broken link
	return ResolveResult{Broken: true}
}

// BuildBacklinks populates the Backlinks field on all pages
func BuildBacklinks(pages []*Page) {
	resolver := NewLinkResolver(pages)

	// First, extract outlinks for all pages
	for _, page := range pages {
		links := ExtractWikiLinks(page.RawContent)
		for _, link := range links {
			page.OutLinks = append(page.OutLinks, link.Target)
		}
	}

	// Build reverse lookup (backlinks)
	for _, page := range pages {
		for _, target := range page.OutLinks {
			result := resolver.Resolve(target)
			if result.Page != nil && result.Page != page {
				// Add this page as a backlink to the target
				result.Page.Backlinks = append(result.Page.Backlinks, page)
			}
		}
	}
}
