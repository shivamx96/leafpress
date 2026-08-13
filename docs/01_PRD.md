# leafpress

## Overview

leafpress is a CLI-driven static site generator purpose-built for digital gardens. It transforms a folder of Markdown files into a clean, interlinked website with minimal configuration. The aesthetic is terminal-inspired—focused, distraction-free, and typographically sharp.

**Core philosophy:** Your garden folder IS the product. leafpress is invisible infrastructure.

## Goals

1. **Zero-friction setup** – `leafpress init` in your notes folder, start publishing
2. **Wiki-style linking** – `[[page-name]]` syntax with automatic backlink generation
3. **Sensible defaults, full override** – ships beautiful out-of-box, customizable via `style.css`
4. **Fast iteration** – sub-second rebuilds, live reload during development
5. **Single binary** – no Node, no Python, no dependencies

---

## CLI Interface

| Command | Description |
|---------|-------------|
| `leafpress init` | Scaffolds `leafpress.json` and optional `style.css` in current directory |
| `leafpress serve` | Starts dev server with live reload (default: `localhost:3000`) |
| `leafpress build` | Generates static site into `_site/` |
| `leafpress new <name>` | Creates a new page with frontmatter template |
| `leafpress deploy` | Deploys to GitHub Pages, Netlify, or Vercel |
| `leafpress status` | Shows changes since the last deployment |
| `leafpress update` | Updates the CLI from the latest release |
| `leafpress version` | Prints the installed version |

### Flags

```
--config, -c       Config file (global; default: ./leafpress.json)
--verbose, -v      Verbose output (global)
serve --port, -p   Override serve port
build/serve -d     Include draft pages
```

---

## Directory Structure

### After `leafpress init`

User runs command inside their existing notes/garden folder:

```
my-garden/                      # User's garden root
├── leafpress.json              # Config (generated)
├── style.css                   # Optional overrides (generated, empty)
├── static/                     # User-created, for images/fonts/etc
├── index.md                    # Home page
├── now.md
├── projects/
│   ├── _index.md               # Optional section index
│   ├── leafpress.md
│   └── yantra.md
├── notes/
│   ├── go-learning.md
│   └── systems-thinking.md
└── _site/                      # Build output (gitignored)
    └── ...
```

### Reserved Paths (Ignored During Content Scan)

```
leafpress.json
style.css
static/
_site/
.leafpress/         # Reserved internal namespace (not created by init)
.git/
.gitignore
.obsidian/          # Common for Obsidian users migrating
node_modules/
```

These are hardcoded. Any markdown outside these paths is content.

---

## Config (`leafpress.json`)

```json
{
  "site": {
    "title": "My Garden",
    "baseURL": "https://example.com"
  },
  "navigation": {
    "mode": "explicit",
    "items": [
      { "label": "Now", "path": "/now" },
      { "label": "Projects", "path": "/projects" }
    ]
  },
  "theme": {
    "preset": "classic",
    "fontBody": "Inter",
    "accent": "#50ac00"
  },
  "features": { "graph": true },
  "build": { "outputDir": "_site", "port": 3000 }
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `site.title` | `"My Garden"` | Site title, shown in nav and `<title>` |
| `site.baseURL` | `""` | Canonical absolute URL; required for sitemap/RSS |
| `build.outputDir` | `"_site"` | Build output directory |
| `build.port` | `3000` | Dev server port |
| `navigation.mode` | `"automatic"` | `"automatic"` (derive nav) or `"explicit"` (`navigation.items`) |
| `theme.preset` | `"classic"` | Bundled visual theme |
| `theme.fontBody` | `"Inter"` | Primary font family |
| `theme.accent` | `"#50ac00"` | Accent color for links |
| `features.graph` | `true` | Enable graph.json generation |

---

## Content Model

### Frontmatter Schema

```yaml
---
title: "Building leafpress"
date: 2025-01-15
tags: [go, tools, side-projects]
draft: false
growth: "seedling"
---
```

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `title` | No | Filename | Page title |
| `date` | No | File mtime | Publication date |
| `tags` | No | `[]` | Categorization |
| `draft` | No | `false` | Exclude from build unless `--drafts` |
| `growth` | No | `null` | `seedling` \| `budding` \| `evergreen` |

### Wiki-Links

```markdown
Check out my thoughts on [[systems-thinking]].
Related: [[projects/yantra|Yantra VPN]]
```

**Resolution order:**
1. Exact slug match (case-insensitive)
2. Filename match anywhere in tree
3. Warn on ambiguity, pick first alphabetically

### Section Index (`_index.md`)

Optional file in any directory to customize section listing pages:

```yaml
---
title: "Projects"
sort: "date"        # date | title | growth
---

Some intro text for the projects section.
```

---

## UI Component System

Semantic HTML with `lp-` prefixed classes. User overrides via `style.css`.

### Layout Shell

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Title }} | {{ .Site.Title }}</title>
  <link rel="stylesheet" href="/style.css">
</head>
<body class="lp-body">
  <nav class="lp-nav">...</nav>
  <main class="lp-main">...</main>
  <footer class="lp-footer">...</footer>
</body>
</html>
```

### Components

#### Navigation (`lp-nav`)

```html
<nav class="lp-nav">
  <a class="lp-nav-title" href="/">{{ .Site.Title }}</a>
  <div class="lp-nav-links">
    {{ range .Site.Nav }}
    <a class="lp-nav-link" href="{{ .Path }}">{{ .Label }}</a>
    {{ end }}
  </div>
</nav>
```

#### Page Header (`lp-header`)

```html
<header class="lp-header">
  <h1 class="lp-title">{{ .Title }}</h1>
  <div class="lp-meta">
    <time class="lp-date" datetime="{{ .Date.Format "2006-01-02" }}">
      {{ .Date.Format "Jan 2, 2006" }}
    </time>
    {{ if .Growth }}
    <span class="lp-growth lp-growth--{{ .Growth }}">{{ growthEmoji .Growth }}</span>
    {{ end }}
  </div>
  {{ if .Tags }}
  <div class="lp-tags">
    {{ range .Tags }}
    <a class="lp-tag" href="/tags/{{ . }}">#{{ . }}</a>
    {{ end }}
  </div>
  {{ end }}
</header>
```

#### Content Body (`lp-content`)

```html
<article class="lp-content">
  <!-- Rendered markdown -->
  <p>...</p>
  <h2 class="lp-h2">...</h2>
  <pre class="lp-codeblock" data-lang="go"><code>...</code></pre>
  <blockquote class="lp-blockquote">...</blockquote>
  <a class="lp-wikilink" href="/notes/systems-thinking">systems thinking</a>
  <a class="lp-external" href="https://..." target="_blank" rel="noopener">External ↗</a>
</article>
```

#### Backlinks (`lp-backlinks`)

```html
<aside class="lp-backlinks">
  <h2 class="lp-backlinks-title">Linked from</h2>
  <ul class="lp-backlinks-list">
    {{ range .Backlinks }}
    <li><a class="lp-backlink" href="{{ .Path }}">{{ .Title }}</a></li>
    {{ end }}
  </ul>
</aside>
```

#### Section Index (`lp-index`)

```html
<ul class="lp-index">
  {{ range .Pages }}
  <li class="lp-index-item">
    <a class="lp-index-link" href="{{ .Path }}">
      <span class="lp-index-title">{{ .Title }}</span>
      {{ if .Growth }}
      <span class="lp-index-growth lp-index-growth--{{ .Growth }}">{{ growthEmoji .Growth }}</span>
      {{ end }}
    </a>
    <time class="lp-index-date">{{ .Date.Format "Jan 2006" }}</time>
  </li>
  {{ end }}
</ul>
```

#### Footer (`lp-footer`)

```html
<footer class="lp-footer">
  {{ if .Site.Author }}<span class="lp-footer-text">&copy; {{ .Site.Author }}. All rights reserved.</span>{{ end }}
  <span class="lp-footer-text">Grown with <a href="https://leafpress.in">leafpress</a></span>
</footer>
```

Renderer hosts may replace only the “Grown with” name and link through the
structured `render.footerAttribution` field. Raw footer HTML and scripts are
not accepted; CLI builds retain the Leafpress attribution.

---

## CSS Architecture

### Embedded Default (CSS Custom Properties)

```css
:root {
  --lp-font: "Inter", system-ui, -apple-system, sans-serif;
  --lp-font-mono: "JetBrains Mono", "Fira Code", monospace;
  --lp-accent: #50ac00;
  --lp-bg: #ffffff;
  --lp-text: #1a1a1a;
  --lp-text-muted: #666666;
  --lp-border: #e5e5e5;
  --lp-code-bg: #f7f7f7;
  --lp-max-width: 680px;
  --lp-nav-height: 60px;
}

/* Theme values from config are injected */
:root {
  --lp-font: "{{ .Site.Theme.Font }}", system-ui, sans-serif;
  --lp-accent: {{ .Site.Theme.Accent }};
}
```

### User Override (`style.css`)

Generated as empty file. User adds overrides:

```css
/* Dark mode */
:root {
  --lp-bg: #0d1117;
  --lp-text: #c9d1d9;
  --lp-text-muted: #8b949e;
  --lp-border: #30363d;
  --lp-code-bg: #161b22;
  --lp-accent: #58a6ff;
}

/* Custom heading style */
.lp-title {
  font-family: "Playfair Display", serif;
}
```

### Growth Stage Indicators

```css
.lp-growth--seedling::before { content: "🌱"; }
.lp-growth--budding::before { content: "🌿"; }
.lp-growth--evergreen::before { content: "🌳"; }
```

---

## Build Output

```
_site/
├── index.html
├── now/index.html
├── projects/
│   ├── index.html          # Section listing
│   ├── leafpress/index.html
│   └── yantra/index.html
├── notes/
│   ├── index.html
│   ├── go-learning/index.html
│   └── systems-thinking/index.html
├── tags/
│   ├── index.html          # All tags
│   ├── go/index.html
│   └── tools/index.html
├── static/                 # User files plus Leafpress-owned assets
│   ├── leafpress/app.<hash>.js # Shared content-addressed client bundle
│   └── ...
├── style.css               # Merged: embedded + user overrides
└── graph.json              # If config.graph = true
```

**URL structure:** Clean URLs (`/projects/leafpress/` not `/projects/leafpress.html`)

---

## Generated Files on `leafpress init`

### `leafpress.json`

```json
{
  "site": {
    "title": "My Garden",
    "baseURL": ""
  },
  "navigation": { "mode": "automatic" },
  "theme": {
    "preset": "classic",
    "fontBody": "Inter",
    "accent": "#50ac00"
  },
  "features": {
    "graph": true,
    "search": true,
    "toc": true,
    "backlinks": true,
    "wikilinks": true,
    "rss": true
  },
  "build": { "outputDir": "_site", "port": 3000 }
}
```

### `style.css`

```css
/* leafpress Custom Styles
 * Override CSS variables or add custom rules below.
 * See: https://leafpress.in/docs/theming
 */
```

### `.gitignore` (appended or created)

```
_site/
```

---

## Technical Implementation Notes

### Link Resolution

```go
func resolveWikiLink(link string, allPages []Page) (*Page, error) {
    slug := slugify(link)  // lowercase, trim
    
    // 1. Exact path match
    for _, p := range allPages {
        if p.Slug == slug || p.Path == slug {
            return &p, nil
        }
    }
    
    // 2. Filename match anywhere
    matches := []Page{}
    for _, p := range allPages {
        if filepath.Base(p.Slug) == slug {
            matches = append(matches, p)
        }
    }
    
    if len(matches) == 1 {
        return &matches[0], nil
    }
    if len(matches) > 1 {
        log.Warnf("Ambiguous link [[%s]], matched: %v", link, matches)
        return &matches[0], nil  // First alphabetically
    }
    
    return nil, fmt.Errorf("broken link: [[%s]]", link)
}
```

### Incremental Rebuild State

`leafpress serve` keeps parsed pages, section/tag indexes, and link resolution
state in memory for the lifetime of the process. It does not create a
`.leafpress/cache.json` file. File changes update the affected pages and shared
artifacts; restarting `serve` performs a fresh scan.

### Live Reload

Inject before `</body>` during `serve`:

```html
<script>
  new WebSocket(`ws://${location.host}/_lr`).onmessage = () => location.reload();
</script>
```

Server sends message on any `.md` or `style.css` change.

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Broken wiki-link | Warn in console, render as plain text with `lp-broken-link` class |
| Missing frontmatter | Use defaults (title from filename, date from mtime) |
| Duplicate slugs | Error on build, list conflicts |
| Invalid config JSON | Error with line number |
| Port in use | Auto-increment port, notify user |

---

## Future Considerations

- Custom shortcodes/components
- Multiple output formats (gemini, plain text)
