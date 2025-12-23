package templates

import (
	"html"
	"html/template"
	"io"
	"regexp"
	"strings"

	"github.com/shivamx96/leafpress/internal/config"
	"github.com/shivamx96/leafpress/internal/content"
)

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
	Site    SiteData
	Page    *content.Page
	Content template.HTML
	TOC     []TOCItem
}

// TOCItem represents a table of contents entry
type TOCItem struct {
	ID    string
	Text  string
	Level int
}

// IndexData is the data passed to index templates
type IndexData struct {
	Site     SiteData
	Title    string
	Pages    []*content.Page
	Intro    template.HTML // Optional intro content for section indexes
	ShowList bool          // Show the page list
}

// TagIndexData is the data passed to the tags index template
type TagIndexData struct {
	Site SiteData
	Tags []TagInfo
}

// TagPageData is the data passed to individual tag pages
type TagPageData struct {
	Site  SiteData
	Tag   string
	Pages []*content.Page
}

// TagInfo holds tag name and count
type TagInfo struct {
	Name  string
	Count int
}

// SiteData contains site-wide information
type SiteData struct {
	Title       string
	Nav         []config.NavItem
	Theme       config.Theme
	BaseURL     string
	TOC         bool
	GraphOnHome bool
}

// New creates a new Templates instance
func New() (*Templates, error) {
	funcs := template.FuncMap{
		"growthEmoji": growthEmoji,
		"lower":       strings.ToLower,
		"safeHTML":    func(s string) template.HTML { return template.HTML(s) },
		"safeCSS":     func(s string) template.CSS { return template.CSS(s) },
		"fontURL":     fontURL,
	}

	// Parse base template
	base, err := template.New("base").Funcs(funcs).Parse(baseTemplate)
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

	return &Templates{
		base:     base,
		page:     page,
		index:    index,
		tagIndex: tagIndex,
		tagPage:  tagPage,
	}, nil
}

// RenderPage renders a content page
func (t *Templates) RenderPage(w io.Writer, data PageData) error {
	return t.page.Execute(w, data)
}

// RenderIndex renders a section index page
func (t *Templates) RenderIndex(w io.Writer, data IndexData) error {
	return t.index.Execute(w, data)
}

// RenderTagIndex renders the tags index page
func (t *Templates) RenderTagIndex(w io.Writer, data TagIndexData) error {
	return t.tagIndex.Execute(w, data)
}

// RenderTagPage renders an individual tag page
func (t *Templates) RenderTagPage(w io.Writer, data TagPageData) error {
	return t.tagPage.Execute(w, data)
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

func fontURL(font string) template.URL {
	// Replace spaces with + for Google Fonts URL
	fontParam := strings.ReplaceAll(font, " ", "+")
	return template.URL("https://fonts.googleapis.com/css2?family=" + fontParam + ":wght@400;500;600;700&display=swap")
}

// ExtractTOC extracts headings from HTML content and adds IDs to them
func ExtractTOC(htmlContent string) (string, []TOCItem) {
	headingRegex := regexp.MustCompile(`<h([2-3])([^>]*)>(.*?)</h[2-3]>`)
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
		plainText := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(text, "")
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
		if attrs != "" && !regexp.MustCompile(`id\s*=`).MatchString(attrs) {
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
	id := regexp.MustCompile(`[^\x00-\x7F]+`).ReplaceAllString(text, "")

	// Trim spaces that may be left after emoji removal
	id = strings.TrimSpace(id)

	// Convert to lowercase
	id = strings.ToLower(id)

	// Replace spaces and special characters with hyphens
	id = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(id, "-")

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
    {{if .Site.Theme.StickyNav}}
    .lp-nav {
      position: sticky;
      top: 0;
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
  <div class="lp-nav-placeholder"></div>
  <nav class="lp-nav">
    <div class="lp-nav-container">
      <div class="lp-nav-brand">
        <a class="lp-nav-title" href="/">{{.Site.Title}}</a>
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
      <div class="lp-nav-links">
        {{range .Site.Nav}}
        <a class="lp-nav-link" href="{{.Path}}">{{.Label}}</a>
        {{end}}
      </div>
    </div>
  </nav>
  <main class="lp-main">
    {{block "content" .}}{{end}}
  </main>
  <footer class="lp-footer">
    <span class="lp-footer-text">Grown with <a href="https://leafpress.in">leafpress</a></span>
  </footer>
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

      // Knowledge Graph Visualization
      var graphContainer = document.getElementById('lp-graph');
      if (graphContainer) {
        fetch('/graph.json')
          .then(function(response) { return response.json(); })
          .then(function(data) {
            renderGraph(data);
          });
      }

      function renderGraph(data) {
        var width = graphContainer.offsetWidth;
        var height = 500;

        var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svg.setAttribute('width', width);
        svg.setAttribute('height', height);
        svg.setAttribute('viewBox', '0 0 ' + width + ' ' + height);
        graphContainer.appendChild(svg);

        // Simple force simulation without D3 - create nodes with positions first
        var nodes = data.nodes.map(function(d) {
          return {
            id: d.id,
            title: d.title,
            growth: d.growth,
            x: Math.random() * width,
            y: Math.random() * height,
            vx: 0,
            vy: 0
          };
        });

        // Create node lookup from nodes with x/y coordinates
        var nodeMap = {};
        nodes.forEach(function(n) {
          nodeMap[n.id] = n;
        });

        // Create links with proper node references
        var links = [];
        data.edges.forEach(function(edge) {
          var source = nodeMap[edge.source];
          var target = nodeMap[edge.target];
          if (source && target) {
            links.push({
              source: source,
              target: target,
              sourceId: edge.source,
              targetId: edge.target
            });
          }
        });

        console.log('Graph data:', nodes.length, 'nodes,', links.length, 'links');

        // Create groups for proper layering
        var linkGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        linkGroup.setAttribute('class', 'lp-graph-links');
        svg.appendChild(linkGroup);

        var nodeGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        nodeGroup.setAttribute('class', 'lp-graph-nodes');
        svg.appendChild(nodeGroup);

        // Get theme colors
        var isDark = document.documentElement.getAttribute('data-theme') === 'dark';
        var linkColor = isDark ? '#444444' : '#d0d0d0';

        // Draw links with initial positions
        links.forEach(function(link) {
          var line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
          line.setAttribute('class', 'lp-graph-link');
          line.setAttribute('stroke', linkColor);
          line.setAttribute('stroke-width', '1.5');
          line.setAttribute('stroke-opacity', '0.5');
          line.setAttribute('x1', link.source.x);
          line.setAttribute('y1', link.source.y);
          line.setAttribute('x2', link.target.x);
          line.setAttribute('y2', link.target.y);
          linkGroup.appendChild(line);
          link.element = line;
        });

        console.log('Drew', links.length, 'links');

        // Text labels group
        var labelGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        labelGroup.setAttribute('class', 'lp-graph-labels');
        svg.appendChild(labelGroup);

        // Draw nodes
        nodes.forEach(function(node) {
          var circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
          circle.setAttribute('class', 'lp-graph-node');
          circle.setAttribute('r', '8');
          circle.setAttribute('fill', getNodeColor(node.growth));
          circle.setAttribute('stroke', '#fff');
          circle.setAttribute('stroke-width', '2');
          circle.style.cursor = 'pointer';

          // Hover to highlight connections
          circle.addEventListener('mouseenter', function() {
            highlightConnections(node);
          });

          circle.addEventListener('mouseleave', function() {
            clearHighlight();
          });

          circle.addEventListener('click', function(e) {
            e.preventDefault();
            if (node.id) {
              window.location.href = '/' + node.id + '/';
            } else {
              window.location.href = '/';
            }
          });

          nodeGroup.appendChild(circle);
          node.element = circle;

          // Add text label
          var text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
          text.setAttribute('class', 'lp-graph-label');
          text.setAttribute('text-anchor', 'middle');
          text.setAttribute('dy', '28');
          text.setAttribute('font-size', '12');
          text.setAttribute('font-weight', '500');
          text.setAttribute('pointer-events', 'none');
          text.textContent = node.title || 'Home';
          text.style.opacity = '0';
          text.style.fill = getComputedStyle(document.documentElement).getPropertyValue('--lp-text').trim();
          labelGroup.appendChild(text);
          node.label = text;
        });

        function highlightConnections(selectedNode) {
          var accentColor = getComputedStyle(document.documentElement).getPropertyValue('--lp-accent').trim();

          // Dim all
          nodes.forEach(function(n) {
            n.element.style.opacity = '0.15';
            if (n.label) n.label.style.opacity = '0';
          });
          links.forEach(function(l) {
            l.element.style.opacity = '0.05';
          });

          // Highlight selected
          selectedNode.element.style.opacity = '1';
          selectedNode.element.setAttribute('r', '10');
          if (selectedNode.label) selectedNode.label.style.opacity = '1';

          // Highlight connected
          links.forEach(function(link) {
            if (link.sourceId === selectedNode.id || link.targetId === selectedNode.id) {
              link.element.style.opacity = '0.8';
              link.element.setAttribute('stroke', accentColor);
              link.element.setAttribute('stroke-width', '2.5');

              var connectedNode = link.sourceId === selectedNode.id ?
                nodes.find(function(n) { return n.id === link.targetId; }) :
                nodes.find(function(n) { return n.id === link.sourceId; });

              if (connectedNode) {
                connectedNode.element.style.opacity = '1';
                connectedNode.element.setAttribute('r', '9');
                if (connectedNode.label) connectedNode.label.style.opacity = '0.9';
              }
            }
          });
        }

        function clearHighlight() {
          nodes.forEach(function(n) {
            n.element.style.opacity = '1';
            n.element.setAttribute('r', '8');
            if (n.label) n.label.style.opacity = '0';
          });
          links.forEach(function(l) {
            l.element.style.opacity = '0.5';
            l.element.setAttribute('stroke', linkColor);
            l.element.setAttribute('stroke-width', '1.5');
          });
        }

        function getNodeColor(growth) {
          var accent = getComputedStyle(document.documentElement).getPropertyValue('--lp-accent').trim();
          if (growth === 'seedling') return '#a8e6a1';
          if (growth === 'budding') return accent;
          if (growth === 'evergreen') return '#2d8659';
          return accent;
        }

        // Simple physics simulation
        function simulate() {
          var alpha = 0.3;
          var iterations = 300;

          for (var k = 0; k < iterations; k++) {
            // Apply forces
            nodes.forEach(function(node) {
              node.vx = 0;
              node.vy = 0;
            });

            // Repulsion between nodes
            for (var i = 0; i < nodes.length; i++) {
              for (var j = i + 1; j < nodes.length; j++) {
                var dx = nodes[j].x - nodes[i].x;
                var dy = nodes[j].y - nodes[i].y;
                var dist = Math.sqrt(dx * dx + dy * dy) || 1;
                var force = 100 / (dist * dist);

                nodes[i].vx -= force * dx / dist;
                nodes[i].vy -= force * dy / dist;
                nodes[j].vx += force * dx / dist;
                nodes[j].vy += force * dy / dist;
              }
            }

            // Link attraction
            links.forEach(function(link) {
              var dx = link.target.x - link.source.x;
              var dy = link.target.y - link.source.y;
              var dist = Math.sqrt(dx * dx + dy * dy) || 1;
              var force = (dist - 50) * 0.1;

              link.source.vx += force * dx / dist;
              link.source.vy += force * dy / dist;
              link.target.vx -= force * dx / dist;
              link.target.vy -= force * dy / dist;
            });

            // Center attraction
            var centerX = width / 2;
            var centerY = height / 2;
            nodes.forEach(function(node) {
              node.vx += (centerX - node.x) * 0.01;
              node.vy += (centerY - node.y) * 0.01;
            });

            // Update positions
            nodes.forEach(function(node) {
              node.x += node.vx * alpha;
              node.y += node.vy * alpha;

              // Keep in bounds
              node.x = Math.max(20, Math.min(width - 20, node.x));
              node.y = Math.max(20, Math.min(height - 20, node.y));
            });

            alpha *= 0.99;
          }

          // Update DOM
          nodes.forEach(function(node) {
            node.element.setAttribute('cx', node.x);
            node.element.setAttribute('cy', node.y);
            if (node.label) {
              node.label.setAttribute('x', node.x);
              node.label.setAttribute('y', node.y);
            }
          });

          links.forEach(function(link) {
            link.element.setAttribute('x1', link.source.x);
            link.element.setAttribute('y1', link.source.y);
            link.element.setAttribute('x2', link.target.x);
            link.element.setAttribute('y2', link.target.y);
          });
        }

        simulate();
      }
    });
  </script>
</body>
</html>
`

const pageTemplate = `
{{define "title"}}{{if eq .Page.Slug ""}}{{.Site.Title}}{{else}}{{.Page.Title}} | {{.Site.Title}}{{end}}{{end}}
{{define "content"}}
<div class="lp-page-container">
  {{if and .Site.TOC .TOC}}
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
        <span class="lp-growth lp-growth--{{.Page.Growth}}">{{growthEmoji .Page.Growth}}</span>
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

    {{if and .Site.GraphOnHome (eq .Page.Slug "")}}
    <div class="lp-graph-container">
      <h2 class="lp-graph-title">Knowledge Graph</h2>
      <div id="lp-graph"></div>
    </div>
    {{end}}

    {{if .Page.Backlinks}}
    <aside class="lp-backlinks">
      <h2 class="lp-backlinks-title">Linked from</h2>
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
{{define "content"}}
<div class="lp-section">
  <h1 class="lp-section-title">{{.Title}}</h1>

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
