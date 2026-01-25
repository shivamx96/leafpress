---
title: "Configuration"
date: 2025-12-21
---

Configure leafpress through `leafpress.json` in your site root.

## Minimal Config

```json
{
  "title": "My Garden"
}
```

That's it. Everything else has sensible defaults.

## Full Reference

```json
{
  "title": "My Digital Garden",
  "author": "Your Name",
  "baseURL": "https://example.com",
  "description": "A collection of thoughts and ideas",
  "image": "/static/images/og-image.png",
  "outputDir": "_site",
  "port": 3000,
  
  "nav": [
    { "label": "About", "path": "/about" },
    { "label": "Projects", "path": "/projects/" }
  ],
  
  "theme": {
    "fontHeading": "Crimson Pro",
    "fontBody": "Inter",
    "fontMono": "JetBrains Mono",
    "accent": "#50ac00",
    "background": {
      "light": "#ffffff",
      "dark": "#1a1a1a"
    },
    "navStyle": "base",
    "navActiveStyle": "base"
  },
  
  "graph": true,
  "toc": true,
  "search": true,
  "wikilinks": true,
  
  "headExtra": "<script defer data-domain=\"example.com\" src=\"https://plausible.io/js/script.js\"></script>"
}
```

## Options

### Site Metadata

| Option | Default | Description |
|--------|---------|-------------|
| `title` | `"My Garden"` | Site title, shown in nav and browser tab |
| `author` | `""` | Author name for RSS feed |
| `baseURL` | `""` | Production URL for sitemap and canonical links |
| `description` | `""` | Site description for SEO |
| `image` | `""` | Default OG image for social sharing |

### Build Options

| Option | Default | Description |
|--------|---------|-------------|
| `outputDir` | `"_site"` | Build output directory |
| `port` | `3000` | Dev server port |
| `headExtra` | `""` | Custom HTML to inject in `<head>` |

### Navigation

```json
{
  "nav": [
    { "label": "Home", "path": "/" },
    { "label": "Docs", "path": "/docs/" }
  ]
}
```

### Theme

| Option | Default | Description |
|--------|---------|-------------|
| `fontHeading` | `"Crimson Pro"` | Google Font for headings |
| `fontBody` | `"Inter"` | Google Font for body text |
| `fontMono` | `"JetBrains Mono"` | Google Font for code |
| `accent` | `"#50ac00"` | Accent color for links and highlights |
| `background.light` | `"#ffffff"` | Light mode background (color or gradient) |
| `background.dark` | `"#1a1a1a"` | Dark mode background (color or gradient) |
| `navStyle` | `"base"` | `"base"`, `"sticky"`, or `"glassy"` |
| `navActiveStyle` | `"base"` | `"base"`, `"box"`, or `"underlined"` |

Gradients work too:
```json
{
  "theme": {
    "background": {
      "light": "linear-gradient(180deg, #ffffff 0%, #f5f5f5 100%)",
      "dark": "linear-gradient(180deg, #0a0a0a 0%, #171717 100%)"
    }
  }
}
```

### Features

| Option | Default | Description |
|--------|---------|-------------|
| `graph` | `false` | Show interactive graph visualization |
| `toc` | `true` | Show table of contents on pages |
| `search` | `true` | Enable full-text search |
| `wikilinks` | `true` | Enable wiki-link processing |
| `backlinks` | `true` | Show backlinks section on pages |

### Ignore Patterns

Exclude files from builds using glob patterns:

```json
{
  "ignore": ["drafts/**", "*.draft.md", "private/**"]
}
```

## Custom Head Content

Use `headExtra` to inject custom HTML into `<head>`. Useful for analytics, verification tags, or additional scripts.

```json
{
  "headExtra": "<script defer data-domain=\"example.com\" src=\"https://plausible.io/js/script.js\"></script>"
}
```

**Examples:**

Plausible Analytics:
```json
{
  "headExtra": "<script defer data-domain=\"example.com\" src=\"https://plausible.io/js/script.js\"></script>"
}
```

Umami Analytics:
```json
{
  "headExtra": "<script defer src=\"https://analytics.example.com/script.js\" data-website-id=\"xxx\"></script>"
}
```

Google Site Verification:
```json
{
  "headExtra": "<meta name=\"google-site-verification\" content=\"xxx\" />"
}
```

## Per-Page Overrides

Override global settings in frontmatter:

```yaml
---
title: "Long Article"
toc: true
---
```

```yaml
---
title: "Short Note"  
toc: false
---
```

