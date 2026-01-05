package templates

import (
	"bufio"
	"html"
	"html/template"
	"io"
	"regexp"
	"strings"

	"github.com/shivamx96/leafpress/cli/internal/config"
	"github.com/shivamx96/leafpress/cli/internal/content"
)

// Pre-compiled regexes for ExtractTOC and generateHeadingID (compiled once at startup)
var (
	headingRegex    = regexp.MustCompile(`<h([2-3])([^>]*)>(.*?)</h[2-3]>`)
	htmlTagRegex    = regexp.MustCompile(`<[^>]*>`)
	idAttrRegex     = regexp.MustCompile(`id\s*=`)
	nonASCIIRegex   = regexp.MustCompile(`[^\x00-\x7F]+`)
	nonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)
)

// Cached templates singleton (parsed once at first use)
var (
	cachedTemplates *Templates
	templateFuncs   template.FuncMap
)

func init() {
	templateFuncs = template.FuncMap{
		"growthEmoji":       growthEmoji,
		"growthDescription": growthDescription,
		"lower":             strings.ToLower,
		"safeHTML":          func(s string) template.HTML { return template.HTML(s) },
		"safeCSS":           func(s string) template.CSS { return template.CSS(s) },
		"fontURL":           fontURL,
		"hasPrefix":         strings.HasPrefix,
	}
}

// Templates holds all parsed templates
type Templates struct {
	base     *template.Template
	page     *template.Template
	index    *template.Template
	tagIndex *template.Template
	tagPage  *template.Template
}

// PageData is the data passed to page templates
type PageData struct {
	Site        SiteData
	Page        *content.Page
	Content     template.HTML
	TOC         []TOCItem
	CurrentPath string // Current page path for nav active state
}

// TOCItem represents a table of contents entry
type TOCItem struct {
	ID    string
	Text  string
	Level int
}

// IndexData is the data passed to index templates
type IndexData struct {
	Site        SiteData
	Title       string
	Pages       []*content.Page
	Intro       template.HTML // Optional intro content for section indexes
	ShowList    bool          // Show the page list
	CurrentPath string        // Current page path for nav active state
}

// TagIndexData is the data passed to the tags index template
type TagIndexData struct {
	Site        SiteData
	Tags        []TagInfo
	CurrentPath string // Current page path for nav active state
}

// TagPageData is the data passed to individual tag pages
type TagPageData struct {
	Site        SiteData
	Tag         string
	Pages       []*content.Page
	CurrentPath string // Current page path for nav active state
}

// TagInfo holds tag name and count
type TagInfo struct {
	Name  string
	Count int
}

// SiteData contains site-wide information
type SiteData struct {
	Title   string
	Author  string
	Nav     []config.NavItem
	Theme   config.Theme
	BaseURL string
	TOC     bool
	Graph   bool
	Search  bool
}

// New returns a cached Templates instance (parsed once, reused on subsequent calls)
func New() (*Templates, error) {
	// Return cached templates if already parsed
	if cachedTemplates != nil {
		return cachedTemplates, nil
	}

	// Parse base template
	base, err := template.New("base").Funcs(templateFuncs).Parse(baseTemplate)
	if err != nil {
		return nil, err
	}

	// Clone base and add page-specific templates
	page, err := template.Must(base.Clone()).Parse(pageTemplate)
	if err != nil {
		return nil, err
	}

	index, err := template.Must(base.Clone()).Parse(indexTemplate)
	if err != nil {
		return nil, err
	}

	tagIndex, err := template.Must(base.Clone()).Parse(tagIndexTemplate)
	if err != nil {
		return nil, err
	}

	tagPage, err := template.Must(base.Clone()).Parse(tagPageTemplate)
	if err != nil {
		return nil, err
	}

	cachedTemplates = &Templates{
		base:     base,
		page:     page,
		index:    index,
		tagIndex: tagIndex,
		tagPage:  tagPage,
	}

	return cachedTemplates, nil
}

// RenderPage renders a content page
func (t *Templates) RenderPage(w io.Writer, data PageData) error {
	bw := bufio.NewWriterSize(w, 8192)
	if err := t.page.Execute(bw, data); err != nil {
		return err
	}
	return bw.Flush()
}

// RenderIndex renders a section index page
func (t *Templates) RenderIndex(w io.Writer, data IndexData) error {
	bw := bufio.NewWriterSize(w, 8192)
	if err := t.index.Execute(bw, data); err != nil {
		return err
	}
	return bw.Flush()
}

// RenderTagIndex renders the tags index page
func (t *Templates) RenderTagIndex(w io.Writer, data TagIndexData) error {
	bw := bufio.NewWriterSize(w, 4096)
	if err := t.tagIndex.Execute(bw, data); err != nil {
		return err
	}
	return bw.Flush()
}

// RenderTagPage renders an individual tag page
func (t *Templates) RenderTagPage(w io.Writer, data TagPageData) error {
	bw := bufio.NewWriterSize(w, 8192)
	if err := t.tagPage.Execute(bw, data); err != nil {
		return err
	}
	return bw.Flush()
}

func growthEmoji(growth string) string {
	switch growth {
	case "seedling":
		return "🌱"
	case "budding":
		return "🌿"
	case "evergreen":
		return "🌳"
	default:
		return ""
	}
}

func growthDescription(growth string) string {
	switch growth {
	case "seedling":
		return "Seedling: Early idea, still developing"
	case "budding":
		return "Budding: Growing, but not yet complete"
	case "evergreen":
		return "Evergreen: Fully grown and refined"
	default:
		return ""
	}
}

func fontURL(font string) template.URL {
	// Replace spaces with + for Google Fonts URL
	fontParam := strings.ReplaceAll(font, " ", "+")
	return template.URL("https://fonts.googleapis.com/css2?family=" + fontParam + ":wght@400;500;600;700&display=swap")
}

// ExtractTOC extracts headings from HTML content and adds IDs to them
func ExtractTOC(htmlContent string) (string, []TOCItem) {
	var toc []TOCItem
	idCounter := make(map[string]int)

	modifiedHTML := headingRegex.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract level, attributes, and text
		matches := headingRegex.FindStringSubmatch(match)
		if len(matches) != 4 {
			return match
		}

		level := matches[1]
		attrs := matches[2]
		text := matches[3]

		// Strip HTML tags from text for TOC display
		plainText := htmlTagRegex.ReplaceAllString(text, "")
		// Unescape HTML entities (e.g., &amp; -> &, &#39; -> ')
		plainText = html.UnescapeString(plainText)

		// Generate ID from text
		id := generateHeadingID(plainText)

		// Handle duplicate IDs
		if count, exists := idCounter[id]; exists {
			idCounter[id] = count + 1
			id = id + "-" + string(rune('0'+count))
		} else {
			idCounter[id] = 1
		}

		// Add to TOC
		levelInt := 2
		if level == "3" {
			levelInt = 3
		}
		toc = append(toc, TOCItem{
			ID:    id,
			Text:  plainText,
			Level: levelInt,
		})

		// Return heading with ID (preserve existing attributes if any)
		if attrs != "" && !idAttrRegex.MatchString(attrs) {
			return "<h" + level + attrs + " id=\"" + id + "\">" + text + "</h" + level + ">"
		} else if attrs == "" {
			return "<h" + level + " id=\"" + id + "\">" + text + "</h" + level + ">"
		}
		// If it already has an id, skip
		return match
	})

	return modifiedHTML, toc
}

// generateHeadingID creates a URL-safe ID from heading text
func generateHeadingID(text string) string {
	// Remove emojis and other non-ASCII characters first
	id := nonASCIIRegex.ReplaceAllString(text, "")

	// Trim spaces that may be left after emoji removal
	id = strings.TrimSpace(id)

	// Convert to lowercase
	id = strings.ToLower(id)

	// Replace spaces and special characters with hyphens
	id = nonAlphaNumeric.ReplaceAllString(id, "-")

	// Remove leading/trailing hyphens
	id = strings.Trim(id, "-")

	return id
}

// Template strings
const baseTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{block "title" .}}{{.Site.Title}}{{end}}</title>
  <link rel="icon" type="image/svg+xml" href="/favicon.svg">
  <link rel="icon" type="image/png" sizes="96x96" href="/favicon-96x96.png">
  <link rel="icon" type="image/x-icon" href="/favicon.ico">
  <link rel="stylesheet" href="/style.css">
  <style>
    :root {
      --lp-font-heading: "{{.Site.Theme.FontHeading}}", Georgia, serif;
      --lp-font-body: "{{.Site.Theme.FontBody}}", system-ui, -apple-system, sans-serif;
      --lp-font-mono: "{{.Site.Theme.FontMono}}", "Fira Code", "Courier New", monospace;
      --lp-accent: {{.Site.Theme.Accent}};
      --lp-bg: #ffffff;
      --lp-text: #1a1a1a;
      --lp-text-muted: #666666;
      --lp-border: #e5e5e5;
      --lp-code-bg: #f7f7f7;
      --lp-max-width: 680px;
      --lp-nav-height: 60px;
    }
    {{if .Site.Theme.Background.Light}}
    :root {
      --lp-bg: {{.Site.Theme.Background.Light | safeCSS}};
    }
    {{end}}
    {{if eq .Site.Theme.NavStyle "sticky"}}
    .lp-nav {
      position: sticky;
      top: 0;
      z-index: 100;
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
    }
    {{end}}
    {{if eq .Site.Theme.NavStyle "glassy"}}
    .lp-nav {
      z-index: 100;
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
    }
    {{end}}

    [data-theme="dark"] {
      --lp-bg: #1a1a1a;
      --lp-text: #e5e5e5;
      --lp-text-muted: #a0a0a0;
      --lp-border: #333333;
      --lp-code-bg: #2a2a2a;
    }
    {{if .Site.Theme.Background.Dark}}
    [data-theme="dark"] {
      --lp-bg: {{.Site.Theme.Background.Dark | safeCSS}};
    }
    {{end}}
  </style>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="{{.Site.Theme.FontHeading | fontURL}}" rel="stylesheet">
  <link href="{{.Site.Theme.FontBody | fontURL}}" rel="stylesheet">
  <link href="{{.Site.Theme.FontMono | fontURL}}" rel="stylesheet">
</head>
<body class="lp-body">
  {{if eq .Site.Theme.NavStyle "glassy"}}<div class="lp-nav-placeholder"></div>{{end}}
  <nav class="lp-nav">
    <div class="lp-nav-container">
      <div class="lp-nav-brand">
        <a class="lp-nav-title" href="/">{{.Site.Title}}</a>
        <div class="lp-nav-actions">
          {{if .Site.Graph}}<button class="lp-graph-toggle" aria-label="Open knowledge graph" title="Explore graph">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="6" cy="6" r="3"></circle>
              <circle cx="18" cy="6" r="3"></circle>
              <circle cx="6" cy="18" r="3"></circle>
              <circle cx="18" cy="18" r="3"></circle>
              <line x1="8.5" y1="7.5" x2="15.5" y2="16.5"></line>
              <line x1="8.5" y1="16.5" x2="15.5" y2="7.5"></line>
            </svg>
          </button>{{end}}
          {{if .Site.Search}}<button class="lp-search-toggle" aria-label="Search" title="Search (⌘K)">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="6"></circle>
              <line x1="21" y1="21" x2="15.5" y2="15.5"></line>
            </svg>
          </button>{{end}}
          <button class="lp-theme-toggle" aria-label="Toggle dark mode" title="Toggle theme">
            <svg class="lp-theme-icon lp-theme-icon-light" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="5"></circle>
              <line x1="12" y1="1" x2="12" y2="3"></line>
              <line x1="12" y1="21" x2="12" y2="23"></line>
              <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
              <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
              <line x1="1" y1="12" x2="3" y2="12"></line>
              <line x1="21" y1="12" x2="23" y2="12"></line>
              <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
              <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
            </svg>
            <svg class="lp-theme-icon lp-theme-icon-dark" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
            </svg>
          </button>
        </div>
      </div>
      <div class="lp-nav-links">
        {{range .Site.Nav}}
        <a class="lp-nav-link{{if hasPrefix $.CurrentPath .Path}} lp-nav-link--active lp-nav-active-{{$.Site.Theme.NavActiveStyle}}{{end}}" href="{{.Path}}">{{.Label}}</a>
        {{end}}
      </div>
    </div>
  </nav>
  <main class="lp-main">
    {{block "content" .}}{{end}}
  </main>
  <footer class="lp-footer">
    {{if .Site.Author}}<span class="lp-footer-text">&copy; {{.Site.Author}}. All rights reserved.</span>{{end}}
    <span class="lp-footer-text">Grown with <a href="https://leafpress.in" target="_blank">leafpress</a></span>
  </footer>

  {{if .Site.Graph}}<!-- Graph Overlay -->
  <div class="lp-graph-overlay" id="lp-graph-overlay" aria-hidden="true">
    <div class="lp-graph-backdrop"></div>
    <div class="lp-graph-panel" role="dialog" aria-label="Knowledge Graph" data-current-slug="{{block "currentSlug" .}}{{end}}">
      <button class="lp-graph-close" aria-label="Close graph">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
      <div class="lp-graph-panel-body" id="lp-graph-panel-body"></div>
    </div>
  </div>{{end}}

  {{if .Site.Search}}<!-- Search Overlay -->
  <div class="lp-search-overlay" id="lp-search-overlay" aria-hidden="true">
    <div class="lp-search-backdrop"></div>
    <div class="lp-search-panel" role="dialog" aria-label="Search">
      <div class="lp-search-header">
        <svg class="lp-search-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="6"></circle>
          <line x1="21" y1="21" x2="15.5" y2="15.5"></line>
        </svg>
        <input type="text" class="lp-search-input" id="lp-search-input" placeholder="Search pages..." autocomplete="off" autofocus>
        <kbd class="lp-search-kbd">ESC</kbd>
        <button class="lp-search-close" aria-label="Close search">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>
      <div class="lp-search-results" id="lp-search-results"></div>
    </div>
  </div>{{end}}

  <script>
    // Theme switching
    (function() {
      var theme = localStorage.getItem('theme') || 'light';
      document.documentElement.setAttribute('data-theme', theme);
    })();

    // Add copy buttons to code blocks
    document.addEventListener('DOMContentLoaded', function() {
      // Theme toggle
      var themeToggle = document.querySelector('.lp-theme-toggle');
      if (themeToggle) {
        themeToggle.addEventListener('click', function() {
          var currentTheme = document.documentElement.getAttribute('data-theme') || 'light';
          var newTheme = currentTheme === 'light' ? 'dark' : 'light';
          document.documentElement.setAttribute('data-theme', newTheme);
          localStorage.setItem('theme', newTheme);
        });
      }

      {{if eq .Site.Theme.NavStyle "glassy"}}
      // Floating pill navbar on scroll
      var nav = document.querySelector('.lp-nav');
      var navPlaceholder = document.querySelector('.lp-nav-placeholder');
      if (nav && navPlaceholder) {
        var navHeight = nav.offsetHeight;
        navPlaceholder.style.height = navHeight + 'px';

        window.addEventListener('scroll', function() {
          if (window.scrollY > navHeight) {
            nav.classList.add('lp-nav--pill');
            navPlaceholder.classList.add('lp-nav-placeholder--active');
          } else {
            nav.classList.remove('lp-nav--pill');
            navPlaceholder.classList.remove('lp-nav-placeholder--active');
          }
        });
      }
      {{end}}

      // Copy buttons
      document.querySelectorAll('pre.chroma').forEach(function(pre) {
        var button = document.createElement('button');
        button.className = 'lp-copy-button';
        button.textContent = 'Copy';
        button.setAttribute('aria-label', 'Copy code to clipboard');

        button.addEventListener('click', function() {
          var code = pre.querySelector('code').textContent;
          navigator.clipboard.writeText(code).then(function() {
            button.textContent = 'Copied!';
            setTimeout(function() {
              button.textContent = 'Copy';
            }, 2000);
          }).catch(function() {
            button.textContent = 'Failed';
            setTimeout(function() {
              button.textContent = 'Copy';
            }, 2000);
          });
        });

        pre.style.position = 'relative';
        pre.appendChild(button);
      });
      {{if .Site.Graph}}
      // Graph Overlay
      (function() {
        var overlay = document.getElementById('lp-graph-overlay');
        var panel = overlay.querySelector('.lp-graph-panel');
        var graphBody = document.getElementById('lp-graph-panel-body');
        var toggleBtn = document.querySelector('.lp-graph-toggle');
        var closeBtn = overlay.querySelector('.lp-graph-close');
        var backdrop = overlay.querySelector('.lp-graph-backdrop');
        var currentSlug = panel.getAttribute('data-current-slug') || '';
        var graphData = null;
        var graphRendered = false;

        function openGraph() {
          overlay.classList.add('lp-graph-overlay--open');
          overlay.setAttribute('aria-hidden', 'false');
          document.body.style.overflow = 'hidden';

          if (!graphRendered && graphData) {
            renderGraph(graphData);
            graphRendered = true;
          } else if (!graphData) {
            fetch('/graph.json')
              .then(function(r) { return r.json(); })
              .then(function(data) {
                graphData = data;
                renderGraph(data);
                graphRendered = true;
              });
          }
        }

        function closeGraph() {
          overlay.classList.remove('lp-graph-overlay--open');
          overlay.setAttribute('aria-hidden', 'true');
          document.body.style.overflow = '';
        }

        toggleBtn.addEventListener('click', openGraph);
        closeBtn.addEventListener('click', closeGraph);
        backdrop.addEventListener('click', closeGraph);

        document.addEventListener('keydown', function(e) {
          if (e.key === 'Escape' && overlay.classList.contains('lp-graph-overlay--open')) {
            closeGraph();
          }
        });

        function renderGraph(data) {
          var width = graphBody.offsetWidth;
          var height = graphBody.offsetHeight;

          var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
          svg.setAttribute('width', width);
          svg.setAttribute('height', height);
          svg.setAttribute('viewBox', '0 0 ' + width + ' ' + height);
          graphBody.appendChild(svg);

          // Pass 1: Group nodes by primary tag for initial placement
          var tagGroups = {};
          var untaggedNodes = [];
          data.nodes.forEach(function(d) {
            var primaryTag = (d.tags && d.tags.length > 0) ? d.tags[0] : null;
            if (primaryTag) {
              if (!tagGroups[primaryTag]) tagGroups[primaryTag] = [];
              tagGroups[primaryTag].push(d);
            } else {
              untaggedNodes.push(d);
            }
          });

          // Assign positions by tag group (arrange in sectors around center)
          var tagNames = Object.keys(tagGroups);
          var numGroups = tagNames.length;
          var centerX = width / 2;
          var centerY = height / 2;
          var radius = Math.min(width, height) * 0.3;

          var nodes = [];
          tagNames.forEach(function(tag, groupIndex) {
            var angle = (2 * Math.PI * groupIndex) / numGroups;
            var groupCenterX = centerX + radius * Math.cos(angle);
            var groupCenterY = centerY + radius * Math.sin(angle);
            var groupNodes = tagGroups[tag];

            groupNodes.forEach(function(d, i) {
              // Spread nodes within group
              var spread = 50;
              var offsetAngle = (2 * Math.PI * i) / groupNodes.length;
              nodes.push({
                id: d.id,
                title: d.title,
                url: d.url,
                tags: d.tags || [],
                x: groupCenterX + spread * Math.cos(offsetAngle) * (0.5 + Math.random() * 0.5),
                y: groupCenterY + spread * Math.sin(offsetAngle) * (0.5 + Math.random() * 0.5),
                vx: 0,
                vy: 0
              });
            });
          });

          // Untagged nodes go near center with some randomness
          untaggedNodes.forEach(function(d) {
            nodes.push({
              id: d.id,
              title: d.title,
              url: d.url,
              tags: d.tags || [],
              x: centerX + (Math.random() - 0.5) * 100,
              y: centerY + (Math.random() - 0.5) * 100,
              vx: 0,
              vy: 0
            });
          });

          var nodeMap = {};
          nodes.forEach(function(n) { nodeMap[n.id] = n; });

          var links = [];
          data.edges.forEach(function(edge) {
            var source = nodeMap[edge.source];
            var target = nodeMap[edge.target];
            if (source && target) {
              links.push({ source: source, target: target, sourceId: edge.source, targetId: edge.target });
            }
          });

          // Calculate node degrees and build adjacency list for clustering
          nodes.forEach(function(n) {
            n.degree = 0;
            n.neighbors = [];
          });
          links.forEach(function(link) {
            link.source.degree++;
            link.target.degree++;
            link.source.neighbors.push(link.target);
            link.target.neighbors.push(link.source);
          });
          var maxDegree = Math.max.apply(null, nodes.map(function(n) { return n.degree; })) || 1;

          // Check if two nodes share neighbors (for clustering)
          function shareNeighbors(a, b) {
            for (var i = 0; i < a.neighbors.length; i++) {
              if (b.neighbors.indexOf(a.neighbors[i]) !== -1) return true;
            }
            return false;
          }

          // Check if two nodes are directly connected
          function areConnected(a, b) {
            return a.neighbors.indexOf(b) !== -1;
          }

          // Count shared tags between two nodes (for tag-based clustering)
          function sharedTagCount(a, b) {
            var count = 0;
            for (var i = 0; i < a.tags.length; i++) {
              if (b.tags.indexOf(a.tags[i]) !== -1) count++;
            }
            return count;
          }

          // Centrality score: normalized degree (0-1)
          function getCentrality(node) {
            return node.degree / maxDegree;
          }

          var linkGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
          svg.appendChild(linkGroup);

          var nodeGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
          svg.appendChild(nodeGroup);

          var labelGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
          svg.appendChild(labelGroup);

          var isDark = document.documentElement.getAttribute('data-theme') === 'dark';
          var linkColor = isDark ? '#444444' : '#d0d0d0';
          var accentColor = getComputedStyle(document.documentElement).getPropertyValue('--lp-accent').trim();

          links.forEach(function(link) {
            var line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
            line.setAttribute('class', 'lp-graph-link');
            line.setAttribute('stroke', linkColor);
            line.setAttribute('stroke-width', '1.5');
            line.setAttribute('stroke-opacity', '0.5');
            linkGroup.appendChild(line);
            link.element = line;
          });

          var selectedNode = null;

          // Node opacity based on link density (degree)
          function getNodeOpacity(degree) {
            // More connections = more opaque (0.15 to 1.0 for better contrast)
            return 0.15 + (degree / maxDegree) * 0.85;
          }

          nodes.forEach(function(node) {
            var circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
            circle.setAttribute('class', 'lp-graph-node');
            circle.setAttribute('r', '6');
            circle.setAttribute('fill', accentColor);
            circle.setAttribute('fill-opacity', getNodeOpacity(node.degree));
            circle.setAttribute('stroke', '#fff');
            circle.setAttribute('stroke-width', '2');
            circle.style.cursor = 'pointer';

            // Mark current page node
            if (node.id === currentSlug) {
              circle.classList.add('lp-graph-node--current');
            }

            // Hover for preview highlight
            circle.addEventListener('mouseenter', function() {
              if (!selectedNode) {
                highlightConnections(node);
              }
            });

            circle.addEventListener('mouseleave', function() {
              if (!selectedNode) {
                clearHighlight();
              }
            });

            // Click to lock selection, second click to navigate
            circle.addEventListener('click', function(e) {
              e.preventDefault();
              if (selectedNode === node) {
                // Second click - navigate
                window.location.href = node.url || '/';
              } else {
                // First click - lock highlight
                selectedNode = node;
                highlightConnections(node);
              }
            });

            nodeGroup.appendChild(circle);
            node.element = circle;

            var text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
            text.setAttribute('class', 'lp-graph-label');
            text.setAttribute('text-anchor', 'middle');
            text.setAttribute('font-size', '0.5em');
            text.setAttribute('pointer-events', 'none');
            text.style.opacity = '0';
            text.style.fill = getComputedStyle(document.documentElement).getPropertyValue('--lp-text').trim();

            // Split long titles into multiple lines
            var title = node.title || 'Home';
            var maxChars = 18;
            var lines = [];

            if (title.length <= maxChars) {
              lines.push(title);
            } else {
              // Split into words and create lines
              var words = title.split(/[\s-]+/);
              var currentLine = '';

              words.forEach(function(word) {
                if ((currentLine + ' ' + word).trim().length <= maxChars) {
                  currentLine = (currentLine + ' ' + word).trim();
                } else {
                  if (currentLine) lines.push(currentLine);
                  currentLine = word;
                }
              });
              if (currentLine) lines.push(currentLine);

              // Limit to 2 lines max
              if (lines.length > 2) {
                lines = [lines[0], lines[1].substring(0, maxChars - 3) + '...'];
              }
            }

            // Store lines for positioning after simulation
            node.labelLines = lines;

            labelGroup.appendChild(text);
            node.label = text;
          });

          // Click on empty space clears selection
          svg.addEventListener('click', function(e) {
            if (e.target === svg) {
              selectedNode = null;
              clearHighlight();
            }
          });

          function highlightConnections(selected) {
            nodes.forEach(function(n) {
              n.element.style.opacity = '0.15';
              if (n.label) n.label.style.opacity = '0';
            });
            links.forEach(function(l) {
              l.element.style.opacity = '0.05';
            });

            selected.element.style.opacity = '1';
            selected.element.setAttribute('r', '8');
            if (selected.label) selected.label.style.opacity = '1';

            links.forEach(function(link) {
              if (link.sourceId === selected.id || link.targetId === selected.id) {
                link.element.style.opacity = '0.8';
                link.element.setAttribute('stroke', accentColor);
                link.element.setAttribute('stroke-width', '2.5');

                var connected = link.sourceId === selected.id ? nodeMap[link.targetId] : nodeMap[link.sourceId];
                if (connected) {
                  connected.element.style.opacity = '1';
                  connected.element.setAttribute('r', '7');
                  if (connected.label) connected.label.style.opacity = '0.9';
                }
              }
            });
          }

          function clearHighlight() {
            nodes.forEach(function(n) {
              n.element.style.opacity = '1';
              n.element.setAttribute('r', n.id === currentSlug ? '8' : '6');
              if (n.label) n.label.style.opacity = n.id === currentSlug ? '1' : '0';
            });
            links.forEach(function(l) {
              l.element.style.opacity = '0.5';
              l.element.setAttribute('stroke', linkColor);
              l.element.setAttribute('stroke-width', '1.5');
            });
          }

          // Pass 2: Physics simulation with tag-based clustering and centrality
          function simulate() {
            var n = nodes.length;
            if (n === 0) return;

            var area = width * height;
            var idealSpacing = Math.sqrt(area / n);

            // Link distance: longer for better spread
            var linkRestLength = Math.max(120, Math.min(280, idealSpacing * 0.75));
            var tagRestLength = linkRestLength * 1.1;
            var clusterRestLength = linkRestLength * 1.3;
            var collisionRadius = 25;

            // Stronger repulsion for better spread
            var repulsionStrength = idealSpacing * idealSpacing * 1.2;

            // Much weaker center force - let nodes spread naturally
            var centerForce = 0.006;

            var iterations = Math.min(350, 120 + n * 6);
            var padding = 35;

            var alpha = 0.3;
            var alphaDecay = 0.995;

            for (var k = 0; k < iterations; k++) {
              // Reset velocities
              nodes.forEach(function(node) { node.vx = 0; node.vy = 0; });

              // Node-node forces
              for (var i = 0; i < n; i++) {
                for (var j = i + 1; j < n; j++) {
                  var a = nodes[i];
                  var b = nodes[j];
                  var dx = b.x - a.x;
                  var dy = b.y - a.y;
                  var dist = Math.sqrt(dx * dx + dy * dy);

                  // Prevent division by zero
                  if (dist < 1) {
                    dx = (Math.random() - 0.5) * 2;
                    dy = (Math.random() - 0.5) * 2;
                    dist = 1;
                  }

                  var force = 0;
                  var connected = areConnected(a, b);
                  var sharedTags = sharedTagCount(a, b);
                  var clustered = !connected && shareNeighbors(a, b);

                  // Centrality weighting: high-degree nodes exert more influence
                  var centralityMult = 1 + (getCentrality(a) + getCentrality(b)) * 0.5;

                  if (connected) {
                    // Connected nodes: strong spring attraction (link force = 1.0 in Obsidian)
                    // Higher centrality = stronger pull
                    var displacement = dist - linkRestLength;
                    force = displacement * 0.1 * centralityMult;
                  } else if (sharedTags > 0) {
                    // Nodes with shared tags: attraction based on tag overlap
                    var displacement = dist - tagRestLength;
                    var tagStrength = 0.08 * Math.min(sharedTags, 3); // Cap at 3 shared tags
                    if (displacement > 0) {
                      force = displacement * tagStrength;
                    } else {
                      // Still repel if too close
                      force = -repulsionStrength * 0.2 / (dist * dist);
                    }
                  } else if (clustered) {
                    // Nodes sharing neighbors: weaker attraction
                    var displacement = dist - clusterRestLength;
                    if (displacement > 0) {
                      force = displacement * 0.04;
                    } else {
                      force = -repulsionStrength * 0.3 / (dist * dist);
                    }
                  } else {
                    // Unrelated nodes: repulsion with distance falloff
                    force = -repulsionStrength / (dist * dist);

                    // Reduced repulsion at large distances (allows clusters)
                    if (dist > idealSpacing * 2) {
                      force *= 0.25;
                    }
                  }

                  // Collision avoidance
                  if (dist < collisionRadius * 2) {
                    force -= (collisionRadius * 2 - dist) * 3;
                  }

                  var fx = (force * dx) / dist;
                  var fy = (force * dy) / dist;
                  a.vx += fx;
                  a.vy += fy;
                  b.vx -= fx;
                  b.vy -= fy;
                }
              }

              // Center gravity (0.52 in Obsidian = strong pull toward center)
              var cx = width / 2;
              var cy = height / 2;
              nodes.forEach(function(node) {
                var dx = cx - node.x;
                var dy = cy - node.y;
                node.vx += dx * centerForce;
                node.vy += dy * centerForce;
              });

              // Apply velocities with damping
              nodes.forEach(function(node) {
                // Velocity damping
                node.vx *= 0.85;
                node.vy *= 0.85;

                node.x += node.vx * alpha;
                node.y += node.vy * alpha;

                // Keep within bounds with padding
                node.x = Math.max(padding, Math.min(width - padding, node.x));
                node.y = Math.max(padding, Math.min(height - padding, node.y));
              });

              alpha *= alphaDecay;

              // Early termination if simulation has settled
              if (alpha < 0.005) break;
            }

            // Update DOM positions
            var centerY = height / 2;
            nodes.forEach(function(node) {
              node.element.setAttribute('cx', node.x);
              node.element.setAttribute('cy', node.y);
              if (node.label && node.labelLines) {
                // Clear existing tspans
                while (node.label.firstChild) {
                  node.label.removeChild(node.label.firstChild);
                }

                // Position label above or below based on node position
                // Nodes in top half -> label below, nodes in bottom half -> label above
                var labelBelow = node.y < centerY;
                var lineHeight = 12;
                var offset = labelBelow ? 16 : -(8 + (node.labelLines.length - 1) * lineHeight);

                node.label.setAttribute('x', node.x);
                node.label.setAttribute('y', node.y);

                node.labelLines.forEach(function(line, idx) {
                  var tspan = document.createElementNS('http://www.w3.org/2000/svg', 'tspan');
                  tspan.setAttribute('x', node.x);
                  tspan.setAttribute('dy', idx === 0 ? offset : lineHeight);
                  tspan.textContent = line;
                  node.label.appendChild(tspan);
                });
              }
            });

            links.forEach(function(link) {
              link.element.setAttribute('x1', link.source.x);
              link.element.setAttribute('y1', link.source.y);
              link.element.setAttribute('x2', link.target.x);
              link.element.setAttribute('y2', link.target.y);
            });

            // Highlight current node after simulation
            if (currentSlug) {
              var current = nodeMap[currentSlug];
              if (current) {
                current.element.setAttribute('r', '8');
                if (current.label) current.label.style.opacity = '1';
              }
            }
          }

          simulate();
        }
      })();
      {{end}}
      {{if .Site.Search}}
      // Search functionality
      (function() {
        var overlay = document.getElementById('lp-search-overlay');
        var input = document.getElementById('lp-search-input');
        var results = document.getElementById('lp-search-results');
        var toggleBtn = document.querySelector('.lp-search-toggle');
        var backdrop = overlay.querySelector('.lp-search-backdrop');
        var searchIndex = null;
        var selectedIndex = -1;

        function openSearch() {
          overlay.classList.add('lp-search-overlay--open');
          overlay.setAttribute('aria-hidden', 'false');
          document.body.style.overflow = 'hidden';
          input.value = '';
          results.innerHTML = '';
          selectedIndex = -1;

          // Focus input - immediate focus for mobile touch events
          input.focus();
          // Backup focus after transition completes
          setTimeout(function() { input.focus(); }, 200);

          if (!searchIndex) {
            fetch('/search-index.json')
              .then(function(r) { return r.json(); })
              .then(function(data) { searchIndex = data; });
          }
        }

        function closeSearch() {
          overlay.classList.remove('lp-search-overlay--open');
          overlay.setAttribute('aria-hidden', 'true');
          document.body.style.overflow = '';
        }

        function search(query) {
          if (!searchIndex || !query.trim()) {
            results.innerHTML = '';
            selectedIndex = -1;
            return;
          }

          var q = query.toLowerCase();
          var scored = [];
          searchIndex.forEach(function(item) {
            var titleLower = item.title.toLowerCase();
            var contentLower = item.content.toLowerCase();
            var score = 0;

            // Title matches (highest priority)
            if (titleLower === q) {
              score = 100; // Exact title match
            } else if (titleLower.indexOf(q) === 0) {
              score = 80; // Title starts with query
            } else if (titleLower.indexOf(q) !== -1) {
              score = 60; // Title contains query
            }

            // Tag matches
            if (item.tags && item.tags.some(function(t) { return t.toLowerCase().indexOf(q) !== -1; })) {
              score = Math.max(score, 40);
            }

            // Content matches (lowest priority)
            if (contentLower.indexOf(q) !== -1) {
              score = Math.max(score, 20);
            }

            if (score > 0) {
              scored.push({ item: item, score: score });
            }
          });

          // Sort by score descending
          scored.sort(function(a, b) { return b.score - a.score; });
          var matches = scored.slice(0, 10).map(function(s) { return s.item; });

          if (matches.length === 0) {
            results.innerHTML = '<div class="lp-search-empty">No results found</div>';
            selectedIndex = -1;
            return;
          }

          results.innerHTML = matches.map(function(item, i) {
            var snippet = getSnippet(item.content, q);
            return '<a class="lp-search-result" href="' + item.url + '" data-index="' + i + '">' +
              '<span class="lp-search-result-title">' + highlightMatch(item.title, q) + '</span>' +
              (snippet ? '<span class="lp-search-result-snippet">' + highlightMatch(snippet, q) + '</span>' : '') +
              '</a>';
          }).join('');
          selectedIndex = -1;
        }

        function getSnippet(content, query) {
          var idx = content.toLowerCase().indexOf(query);
          if (idx === -1) return '';
          var start = Math.max(0, idx - 40);
          var end = Math.min(content.length, idx + query.length + 60);
          var snippet = content.substring(start, end);
          if (start > 0) snippet = '...' + snippet;
          if (end < content.length) snippet = snippet + '...';
          return snippet;
        }

        function highlightMatch(text, query) {
          var regex = new RegExp('(' + query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + ')', 'gi');
          return text.replace(regex, '<mark>$1</mark>');
        }

        function updateSelection() {
          var items = results.querySelectorAll('.lp-search-result');
          items.forEach(function(item, i) {
            item.classList.toggle('lp-search-result--selected', i === selectedIndex);
          });
          if (selectedIndex >= 0 && items[selectedIndex]) {
            items[selectedIndex].scrollIntoView({ block: 'nearest' });
          }
        }

        var closeBtn = overlay.querySelector('.lp-search-close');

        if (toggleBtn) toggleBtn.addEventListener('click', openSearch);
        backdrop.addEventListener('click', closeSearch);
        if (closeBtn) closeBtn.addEventListener('click', closeSearch);

        input.addEventListener('input', function() {
          search(input.value);
        });

        input.addEventListener('keydown', function(e) {
          var items = results.querySelectorAll('.lp-search-result');
          if (e.key === 'ArrowDown') {
            e.preventDefault();
            selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
            updateSelection();
          } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            selectedIndex = Math.max(selectedIndex - 1, -1);
            updateSelection();
          } else if (e.key === 'Enter' && selectedIndex >= 0 && items[selectedIndex]) {
            e.preventDefault();
            window.location.href = items[selectedIndex].getAttribute('href');
          }
        });

        document.addEventListener('keydown', function(e) {
          if (e.key === 'Escape' && overlay.classList.contains('lp-search-overlay--open')) {
            closeSearch();
          }
          if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
            e.preventDefault();
            if (overlay.classList.contains('lp-search-overlay--open')) {
              closeSearch();
            } else {
              openSearch();
            }
          }
        });
      })();
      {{end}}
    });
  </script>
</body>
</html>
`

const pageTemplate = `
{{define "title"}}{{if eq .Page.Slug ""}}{{.Site.Title}}{{else}}{{.Page.Title}} | {{.Site.Title}}{{end}}{{end}}
{{define "currentSlug"}}{{.Page.Slug}}{{end}}
{{define "content"}}
<div class="lp-page-container">
  {{if .TOC}}
  <aside class="lp-toc">
    <nav class="lp-toc-nav">
      <ul class="lp-toc-list">
        {{range .TOC}}
        <li class="lp-toc-item lp-toc-level-{{.Level}}">
          <a href="#{{.ID}}" class="lp-toc-link">{{.Text}}</a>
        </li>
        {{end}}
      </ul>
    </nav>
  </aside>
  {{end}}

  <article class="lp-article">
    <header class="lp-header">
      <h1 class="lp-title">{{.Page.Title}}</h1>
      <div class="lp-meta">
        {{if .Page.Growth}}
        <span class="lp-growth lp-growth--{{.Page.Growth}}" data-growth="{{.Page.Growth}}">{{growthEmoji .Page.Growth}}</span>
        {{end}}
        {{if .Page.ReadingTime}}
        <span class="lp-reading-time">{{.Page.ReadingTimeDisplay}}</span>
        {{end}}
        {{if and .Page.HasModified (not .Page.Date.IsZero)}}
        <span class="lp-date-info">Updated <time class="lp-modified" datetime="{{.Page.ISOModified}}">{{.Page.FormattedModified}}</time> · Created <time class="lp-date" datetime="{{.Page.ISODate}}">{{.Page.FormattedDate}}</time></span>
        {{else if .Page.HasModified}}
        <span class="lp-date-info">Updated <time class="lp-modified" datetime="{{.Page.ISOModified}}">{{.Page.FormattedModified}}</time></span>
        {{else if not .Page.Date.IsZero}}
        <span class="lp-date-info">Created <time class="lp-date" datetime="{{.Page.ISODate}}">{{.Page.FormattedDate}}</time></span>
        {{end}}
      </div>
      {{if .Page.Tags}}
      <div class="lp-tags">
        {{range .Page.Tags}}
        <a class="lp-tag" href="/tags/{{. | lower}}/">#{{.}}</a>
        {{end}}
      </div>
      {{end}}
    </header>

    <div class="lp-content">
      {{.Content}}
    </div>

    {{if .Page.Backlinks}}
    <aside class="lp-backlinks">
      <h2 class="lp-backlinks-title">Referenced from</h2>
      <ul class="lp-backlinks-list">
        {{range .Page.Backlinks}}
        <li><a class="lp-backlink" href="{{.Permalink}}">{{.Title}}</a></li>
        {{end}}
      </ul>
    </aside>
    {{end}}
  </article>
</div>
{{end}}
`

const indexTemplate = `
{{define "title"}}{{.Title}} | {{.Site.Title}}{{end}}
{{define "currentSlug"}}{{end}}
{{define "content"}}
<div class="lp-section">
  <h1 class="lp-section-title">{{.Title}}</h1>
  {{if .ShowList}}<p class="lp-section-count">{{len .Pages}} items in {{.Title}}</p>{{end}}

  {{if .Intro}}
  <div class="lp-section-intro">
    {{.Intro}}
  </div>
  {{end}}

  {{if .ShowList}}
  <ul class="lp-index">
    {{range .Pages}}
    <li class="lp-index-item">
      <a class="lp-index-link" href="{{.Permalink}}">
        {{if .Growth}}
        <span class="lp-index-growth lp-index-growth--{{.Growth}}">{{growthEmoji .Growth}}</span>
        {{end}}
        <span class="lp-index-title">{{.Title}}</span>
      </a>
      {{if .DisplayDate}}
      <time class="lp-index-date" datetime="{{.DisplayDateISO}}">{{.DisplayDate}}</time>
      {{end}}
    </li>
    {{end}}
  </ul>
  {{end}}
</div>
{{end}}
`

const tagIndexTemplate = `
{{define "title"}}Tags | {{.Site.Title}}{{end}}
{{define "currentSlug"}}tags{{end}}
{{define "content"}}
<div class="lp-section">
  <h1 class="lp-section-title">Tags</h1>

  <div class="lp-tag-cloud">
    {{range .Tags}}
    <a class="lp-tag-cloud-item" href="/tags/{{.Name | lower}}/">
      #{{.Name}} <span class="lp-tag-count">({{.Count}})</span>
    </a>
    {{end}}
  </div>
</div>
{{end}}
`

const tagPageTemplate = `
{{define "title"}}#{{.Tag}} | {{.Site.Title}}{{end}}
{{define "currentSlug"}}tags/{{.Tag}}{{end}}
{{define "content"}}
<div class="lp-section">
  <h1 class="lp-section-title">#{{.Tag}}</h1>

  <ul class="lp-index">
    {{range .Pages}}
    <li class="lp-index-item">
      <a class="lp-index-link" href="{{.Permalink}}">
        {{if .Growth}}
        <span class="lp-index-growth lp-index-growth--{{.Growth}}">{{growthEmoji .Growth}}</span>
        {{end}}
        <span class="lp-index-title">{{.Title}}</span>
      </a>
      {{if .DisplayDate}}
      <time class="lp-index-date" datetime="{{.DisplayDateISO}}">{{.DisplayDate}}</time>
      {{end}}
    </li>
    {{end}}
  </ul>
</div>
{{end}}
`
