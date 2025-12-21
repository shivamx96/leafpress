package templates

import (
	"html/template"
	"io"
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
}

// IndexData is the data passed to index templates
type IndexData struct {
	Site  SiteData
	Title string
	Pages []*content.Page
	Intro template.HTML // Optional intro content for section indexes
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
	Title   string
	Nav     []config.NavItem
	Theme   config.Theme
	BaseURL string
}

// New creates a new Templates instance
func New() (*Templates, error) {
	funcs := template.FuncMap{
		"growthEmoji": growthEmoji,
		"lower":       strings.ToLower,
		"safeHTML":    func(s string) template.HTML { return template.HTML(s) },
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

    [data-theme="dark"] {
      --lp-bg: #1a1a1a;
      --lp-text: #e5e5e5;
      --lp-text-muted: #a0a0a0;
      --lp-border: #333333;
      --lp-code-bg: #2a2a2a;
    }
  </style>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="{{.Site.Theme.FontHeading | fontURL}}" rel="stylesheet">
  <link href="{{.Site.Theme.FontBody | fontURL}}" rel="stylesheet">
  <link href="{{.Site.Theme.FontMono | fontURL}}" rel="stylesheet">
</head>
<body class="lp-body">
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
    <span class="lp-footer-text">Grown with <a href="https://github.com/shivamx96/leafpress">LeafPress</a></span>
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
    });
  </script>
</body>
</html>
`

const pageTemplate = `
{{define "title"}}{{.Page.Title}} | {{.Site.Title}}{{end}}
{{define "content"}}
<article class="lp-article">
  <header class="lp-header">
    <h1 class="lp-title">{{.Page.Title}}</h1>
    <div class="lp-meta">
      {{if not .Page.Date.IsZero}}
      <time class="lp-date" datetime="{{.Page.ISODate}}">{{.Page.FormattedDate}}</time>
      {{end}}
      {{if .Page.Growth}}
      <span class="lp-growth lp-growth--{{.Page.Growth}}">{{growthEmoji .Page.Growth}}</span>
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
    <h2 class="lp-backlinks-title">Linked from</h2>
    <ul class="lp-backlinks-list">
      {{range .Page.Backlinks}}
      <li><a class="lp-backlink" href="{{.Permalink}}">{{.Title}}</a></li>
      {{end}}
    </ul>
  </aside>
  {{end}}
</article>
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

  <ul class="lp-index">
    {{range .Pages}}
    <li class="lp-index-item">
      <a class="lp-index-link" href="{{.Permalink}}">
        <span class="lp-index-title">{{.Title}}</span>
        {{if .Growth}}
        <span class="lp-index-growth lp-index-growth--{{.Growth}}">{{growthEmoji .Growth}}</span>
        {{end}}
      </a>
      {{if not .Date.IsZero}}
      <time class="lp-index-date">{{.ShortDate}}</time>
      {{end}}
    </li>
    {{end}}
  </ul>
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
        <span class="lp-index-title">{{.Title}}</span>
        {{if .Growth}}
        <span class="lp-index-growth lp-index-growth--{{.Growth}}">{{growthEmoji .Growth}}</span>
        {{end}}
      </a>
      {{if not .Date.IsZero}}
      <time class="lp-index-date">{{.ShortDate}}</time>
      {{end}}
    </li>
    {{end}}
  </ul>
</div>
{{end}}
`
