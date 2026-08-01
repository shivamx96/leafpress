package content

import (
	"regexp"
	"sort"
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
	pages    []*Page
	slugMap  map[string]*Page   // Exact slug -> page
	nameMap  map[string][]*Page // Filename -> pages (may have duplicates)
	aliasMap map[string]*Page   // Normalized alias (e.g. page title) -> page
}

// normalizeAlias is the matching key for aliases: lowercase with interior
// whitespace collapsed, so "Beta  Note" and "beta note" resolve alike.
func normalizeAlias(name string) string {
	return strings.Join(strings.Fields(strings.ToLower(name)), " ")
}

// NewLinkResolver creates a new link resolver
func NewLinkResolver(pages []*Page) *LinkResolver {
	resolver := &LinkResolver{
		pages:    pages,
		slugMap:  make(map[string]*Page, len(pages)),     // Pre-size to avoid rehashing
		nameMap:  make(map[string][]*Page, len(pages)/2), // Estimate ~2 pages per name on average
		aliasMap: make(map[string]*Page),
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

	// Sort nameMap slices by slug for deterministic ambiguous link resolution
	for _, pages := range resolver.nameMap {
		if len(pages) > 1 {
			sort.Slice(pages, func(i, j int) bool {
				return pages[i].Slug < pages[j].Slug
			})
		}
	}

	// Register display titles as aliases so title-form links ([[Note B]])
	// resolve to hyphenated slugs (note-b). The embedded renderer has always
	// done this for hosted gardens; registering here gives the CLI the same
	// semantics, keeping wikilinks, backlinks, and graph edges in parity.
	for _, page := range pages {
		resolver.AddAlias(page.Title, page)
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

	// 3. Registered alias (e.g. a page title distinct from its slug, so
	// title-form links like [[Beta Note]] reach slug "beta-note")
	if page, ok := r.aliasMap[normalizeAlias(target)]; ok {
		return ResolveResult{Page: page}
	}

	// 4. Broken link
	return ResolveResult{Broken: true}
}

// AddAlias registers an additional name (e.g. a page's display title) that
// resolves to [page]. Matching is case- and interior-whitespace-insensitive.
// Slug and filename matches take precedence. When two pages claim the same
// alias, the lexicographically smaller slug wins — like nameMap's ambiguous
// resolution — so the outcome does not depend on registration order (CLI
// scan order vs renderer input order).
func (r *LinkResolver) AddAlias(name string, page *Page) {
	key := normalizeAlias(name)
	if key == "" {
		return
	}
	if existing, ok := r.aliasMap[key]; ok && existing.Slug <= page.Slug {
		return
	}
	r.aliasMap[key] = page
}

// BuildBacklinks populates the Backlinks field on all pages
// If resolver is nil, a new one will be created
func BuildBacklinks(pages []*Page, resolver ...*LinkResolver) {
	var r *LinkResolver
	if len(resolver) > 0 && resolver[0] != nil {
		r = resolver[0]
	} else {
		r = NewLinkResolver(pages)
	}

	// Clear existing backlinks and outlinks to avoid duplicates on rebuild
	for _, page := range pages {
		page.Backlinks = nil
		page.OutLinks = nil
	}

	// First, extract outlinks for all pages
	for _, page := range pages {
		links := ExtractWikiLinks(page.RawContent)
		for _, link := range links {
			page.OutLinks = append(page.OutLinks, link.Target)
		}
	}

	// Track which pages have already been added as backlinks to avoid duplicates
	backlinkSeen := make(map[*Page]map[*Page]bool)
	for _, page := range pages {
		backlinkSeen[page] = make(map[*Page]bool)
	}

	// Build reverse lookup (backlinks)
	for _, page := range pages {
		for _, target := range page.OutLinks {
			result := r.Resolve(target)
			if result.Page != nil && result.Page != page {
				// Skip if target page isn't in our pages slice (resolver may have stale references)
				if backlinkSeen[result.Page] == nil {
					continue
				}
				// Only add if not already a backlink (deduplicate)
				if !backlinkSeen[result.Page][page] {
					backlinkSeen[result.Page][page] = true
					result.Page.Backlinks = append(result.Page.Backlinks, page)
				}
			}
		}
	}
}
