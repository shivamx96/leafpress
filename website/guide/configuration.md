---
title: "Configuration"
date: 2025-12-21
---

Configure leafpress through `leafpress.json` in your site root. Settings are
grouped into sections — `site`, `theme`, `features`, `navigation`, `build`, and
`deploy`. Every field is optional and has a sensible default, so a tiny config
goes a long way.

## Minimal Config

```json
{
  "site": { "title": "My Garden" }
}
```

That's it. Everything else uses defaults — in fact an empty `{}` builds a valid
default site.

## Full Reference

```json
{
  "site": {
    "title": "My Digital Garden",
    "author": "Your Name",
    "baseURL": "https://example.com",
    "description": "A collection of thoughts and ideas",
    "image": "/static/images/og-image.png",
    "headExtra": "<script defer data-domain=\"example.com\" src=\"https://plausible.io/js/script.js\"></script>"
  },

  "theme": {
    "fontHeading": "Bricolage Grotesque",
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

  "features": {
    "graph": true,
    "toc": true,
    "search": true,
    "wikilinks": true,
    "backlinks": true,
    "rss": true
  },

  "navigation": {
    "mode": "automatic",
    "includeTags": false
  },

  "build": {
    "outputDir": "_site",
    "port": 3000,
    "ignore": ["drafts/**", "*.draft.md"]
  }
}
```

## Options

### `site` — identity & SEO

| Option | Default | Description |
|--------|---------|-------------|
| `title` | `"My Garden"` | Site title, shown in nav and browser tab |
| `author` | `""` | Author name for RSS feed and footer copyright |
| `baseURL` | `""` | Canonical **absolute** URL (e.g. `https://example.com` or `https://example.com/notes`). Used for canonical links, and required for `sitemap.xml` and the RSS feed (see below). The internal link path is derived from its path component. |
| `description` | `""` | Site description for SEO |
| `image` | `""` | Default OG image for social sharing |
| `headExtra` | `""` | Custom HTML injected into `<head>` (see [Custom Head Content](#custom-head-content)) |

### `theme`

| Option | Default | Description |
|--------|---------|-------------|
| `fontHeading` | `"Bricolage Grotesque"` | Heading font (bundled families are self-hosted) |
| `fontBody` | `"Inter"` | Body font (bundled families are self-hosted) |
| `fontMono` | `"JetBrains Mono"` | Code font (bundled families are self-hosted) |
| `remoteFonts` | `false` | Deprecated: load unbundled families from Google Fonts |
| `fonts` | `[]` | Custom local font declarations (family, file under `static/fonts/`, weight, style, display) — see [Theming](/guide/theming/) |
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

### `features`

| Option | Default | Description |
|--------|---------|-------------|
| `graph` | `true` | Show interactive graph visualization |
| `toc` | `true` | Show table of contents on pages |
| `search` | `true` | Enable the full-text search UI (⌘K). The page index used by search and link previews is always generated |
| `wikilinks` | `true` | Enable wiki-link processing |
| `backlinks` | `true` | Show backlinks section on pages |
| `rss` | `true` | Generate RSS feed and show feed icon in nav (requires `site.baseURL`) |

### `navigation`

Choose how the top nav bar is built with `mode`:

**Automatic** (default) derives the nav from your top-level content — root notes
and section homes. The home page itself is reached via the site title, so it is
not repeated as a nav link.

```json
{
  "navigation": {
    "mode": "automatic",
    "includeTags": true
  }
}
```

- `includeTags` (default `false`) appends a **Tags** item when tagged pages
  produce a tags index.

**Explicit** uses exactly the items you list:

```json
{
  "navigation": {
    "mode": "explicit",
    "items": [
      { "label": "Home", "path": "/" },
      { "label": "Docs", "path": "/docs/" },
      { "label": "Tags", "path": "/tags/" }
    ]
  }
}
```

Nav paths must start with `/`.

### `build`

| Option | Default | Description |
|--------|---------|-------------|
| `outputDir` | `"_site"` | Build output directory |
| `port` | `3000` | Dev server port |
| `ignore` | `[]` | Glob patterns to exclude from builds, e.g. `["drafts/**", "*.draft.md", "private/**"]` |

### `deploy`

Deployment settings (provider and provider-specific options) are stored under
`deploy` and managed by `leafpress deploy` — see the deploy guides.

## A note on `baseURL`, sitemap & RSS

`sitemap.xml`, the RSS feed (`feed.xml`), and the `Sitemap:` line in
`robots.txt` all need an absolute origin. If `site.baseURL` is empty, leafpress
**skips** those artifacts (and prints a warning) rather than emitting invalid
relative URLs. Set `site.baseURL` to your production URL to enable them.

## Custom Head Content

Use `site.headExtra` to inject custom HTML into `<head>`. Useful for analytics,
verification tags, or additional scripts.

```json
{
  "site": {
    "headExtra": "<script defer data-domain=\"example.com\" src=\"https://plausible.io/js/script.js\"></script>"
  }
}
```

**Examples:**

Plausible Analytics:
```json
{
  "site": {
    "headExtra": "<script defer data-domain=\"example.com\" src=\"https://plausible.io/js/script.js\"></script>"
  }
}
```

Umami Analytics:
```json
{
  "site": {
    "headExtra": "<script defer src=\"https://analytics.example.com/script.js\" data-website-id=\"xxx\"></script>"
  }
}
```

Google Site Verification:
```json
{
  "site": {
    "headExtra": "<meta name=\"google-site-verification\" content=\"xxx\" />"
  }
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
