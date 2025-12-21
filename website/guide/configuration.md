---
title: "Configuration"
date: 2025-12-21
---

leafpress is configured through `leafpress.json` in your site's root directory.

## Basic Configuration

```json
{
  "title": "My Digital Garden",
  "baseURL": "https://example.com",
  "outputDir": "_site",
  "port": 3000
}
```

### Options

- **title**: Site title shown in navigation and page titles
- **baseURL**: Your site's URL (used for canonical links)
- **outputDir**: Where to output built files (default: `_site`)
- **port**: Dev server port (default: `3000`)

## Navigation

Add links to your site's navigation bar:

```json
{
  "nav": [
    { "label": "Home", "path": "/" },
    { "label": "Notes", "path": "/notes/" },
    { "label": "About", "path": "/about/" }
  ]
}
```

## Theme Configuration

### Fonts

Choose from Google Fonts:

```json
{
  "theme": {
    "fontHeading": "Crimson Pro",
    "fontBody": "Inter",
    "fontMono": "JetBrains Mono"
  }
}
```

### Colors

Set your accent color:

```json
{
  "theme": {
    "accent": "#50ac00"
  }
}
```

### Background

Use solid colors or gradients:

```json
{
  "theme": {
    "background": "#ffffff"
  }
}
```

Or separate backgrounds for light and dark mode:

```json
{
  "theme": {
    "background": {
      "light": "linear-gradient(135deg, #e8f5e9 0%, #ffffff 100%)",
      "dark": "linear-gradient(135deg, #0d1f12 0%, #1a1a1a 100%)"
    }
  }
}
```

### Sticky Navigation

Enable sticky navigation bar:

```json
{
  "theme": {
    "stickyNav": true
  }
}
```

## Features

### Table of Contents

Enable/disable automatic table of contents:

```json
{
  "toc": true
}
```

You can also disable TOC for specific pages in frontmatter:

```yaml
---
title: "My Page"
toc: false
---
```

### Section Index Lists

For section index pages (`_index.md`), control whether to show the automatic list of pages:

```yaml
---
title: "My Section"
showList: false  # Hide automatic page list (default: true)
---
```

This is useful when you want full control over your section index content.

## Next Steps

- [[guide/writing|Writing Content]]
- [[guide/custom-styles|Custom Styles]]
- [[guide/wiki-links|Using Wiki Links]]
