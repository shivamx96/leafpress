package content

import (
	"bytes"
	"html"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// WikiLink represents a parsed wiki-link
type WikiLink struct {
	Target string // The link target (slug or path)
	Label  string // Display label (defaults to target)
	Raw    string // Original raw text including brackets
}

type wikiLinkNode struct {
	ast.BaseInline
	Link WikiLink
}

var kindWikiLink = ast.NewNodeKind("WikiLink")

func (n *wikiLinkNode) Kind() ast.NodeKind { return kindWikiLink }

func (n *wikiLinkNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Target": n.Link.Target,
		"Label":  n.Link.Label,
	}, nil)
}

type wikiLinkParser struct{}

func (p *wikiLinkParser) Trigger() []byte { return []byte{'['} }

func (p *wikiLinkParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	// A generated wikilink anchor cannot be nested inside an ordinary
	// Markdown link or image label.
	if pc.IsInLinkLabel() {
		return nil
	}
	line, _ := block.PeekLine()
	if len(line) < 4 || line[0] != '[' || line[1] != '[' {
		return nil
	}
	closeAt := bytes.Index(line[2:], []byte("]]"))
	if closeAt < 0 {
		return nil
	}
	closeAt += 2
	inner := line[2:closeAt]
	// Match the established syntax: the target may not contain ] or |, and a
	// label may contain | but not ].
	if len(inner) == 0 || bytes.ContainsRune(inner, ']') {
		return nil
	}
	parts := bytes.SplitN(inner, []byte{'|'}, 2)
	if len(parts[0]) == 0 {
		return nil
	}
	target := strings.TrimSpace(string(parts[0]))
	label := target
	if len(parts) == 2 {
		if len(parts[1]) == 0 {
			return nil
		}
		label = strings.TrimSpace(string(parts[1]))
	}
	raw := string(line[:closeAt+2])
	block.Advance(closeAt + 2)
	return &wikiLinkNode{Link: WikiLink{Target: target, Label: label, Raw: raw}}
}

type wikiLinkHTMLRenderer struct {
	resolver         *LinkResolver
	basePath         string
	plainBrokenLinks bool
}

func (r *wikiLinkHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindWikiLink, r.renderWikiLink)
}

func (r *wikiLinkHTMLRenderer) renderWikiLink(
	w util.BufWriter, _ []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	link := node.(*wikiLinkNode).Link
	label := util.EscapeHTML([]byte(link.Label))
	if r.resolver == nil {
		_, _ = w.Write(label)
		return ast.WalkSkipChildren, nil
	}
	resolved := r.resolver.Resolve(link.Target)
	if resolved.Broken {
		if r.plainBrokenLinks {
			_, _ = w.Write(label)
		} else {
			_, _ = w.WriteString(`<span class="lp-broken-link">`)
			_, _ = w.Write(label)
			_, _ = w.WriteString(`</span>`)
		}
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString(`<a class="lp-wikilink" href="`)
	_, _ = w.Write(util.EscapeHTML([]byte(r.basePath + resolved.Page.Permalink)))
	_, _ = w.WriteString(`">`)
	_, _ = w.Write(label)
	_, _ = w.WriteString(`</a>`)
	return ast.WalkSkipChildren, nil
}

type wikiLinkExtension struct {
	renderer *wikiLinkHTMLRenderer
}

func (e *wikiLinkExtension) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&wikiLinkParser{}, 150),
	))
	if e.renderer != nil {
		md.Renderer().AddOptions(renderer.WithNodeRenderers(
			util.Prioritized(e.renderer, 500),
		))
	}
}

var wikiLinkExtractor = goldmark.New(
	goldmark.WithExtensions(&wikiLinkExtension{}),
)

// ExtractWikiLinks extracts wikilinks from Markdown syntax nodes. Code spans,
// fenced/indented code, escaped brackets, raw HTML attributes, and ordinary
// Markdown link/image labels are therefore excluded by the same parser rules
// used for rendering.
func ExtractWikiLinks(content string) []WikiLink {
	var links []WikiLink
	document := wikiLinkExtractor.Parser().Parse(text.NewReader([]byte(content)))
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && node.Kind() == kindWikiLink {
			links = append(links, node.(*wikiLinkNode).Link)
		}
		return ast.WalkContinue, nil
	})
	return links
}

func wikiLinkWarnings(document ast.Node, resolver *LinkResolver) []string {
	if resolver == nil {
		return nil
	}
	var warnings []string
	_ = ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || node.Kind() != kindWikiLink {
			return ast.WalkContinue, nil
		}
		link := node.(*wikiLinkNode).Link
		resolved := resolver.Resolve(link.Target)
		switch {
		case resolved.Broken:
			warnings = append(warnings, "broken link: [["+link.Target+"]]")
		case resolved.Ambiguous:
			warnings = append(warnings, "ambiguous link: [["+link.Target+"]]")
		}
		return ast.WalkContinue, nil
	})
	return warnings
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
	//
	// Titles arrive HTML-escaped (see site.EscapePageMeta) but link targets
	// are raw Markdown text, so register both spellings: a page titled "Q&A"
	// must still resolve from [[Q&A]]. The second call is a no-op when the
	// title contains nothing to escape.
	for _, page := range pages {
		resolver.AddAlias(page.Title, page)
		resolver.AddAlias(html.UnescapeString(page.Title), page)
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

// PopulateOutLinks extracts each page's wiki-link targets independently of
// whether backlinks are enabled. Duplicate targets are retained here because
// distinct aliases may resolve differently; artifact builders deduplicate
// after resolution.
func PopulateOutLinks(pages []*Page) {
	for _, page := range pages {
		page.OutLinks = nil
		for _, link := range ExtractWikiLinks(page.RawContent) {
			page.OutLinks = append(page.OutLinks, link.Target)
		}
	}
}

// BuildBacklinks populates the Backlinks and OutLinks fields on all pages.
// If resolver is nil, a new one will be created.
func BuildBacklinks(pages []*Page, resolver ...*LinkResolver) {
	var r *LinkResolver
	if len(resolver) > 0 && resolver[0] != nil {
		r = resolver[0]
	} else {
		r = NewLinkResolver(pages)
	}

	// Clear existing backlinks to avoid duplicates on rebuild.
	for _, page := range pages {
		page.Backlinks = nil
	}
	PopulateOutLinks(pages)

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
