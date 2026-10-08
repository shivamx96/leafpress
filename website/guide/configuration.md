---
title: "Configuration"
date: 2025-12-21
---

Configure leafpress through `leafpress.json` in your site root. Settings are
grouped into five active sections — `site`, `theme`, `features`, `navigation`,
and `build`. Every field is optional and has a sensible default, so a tiny
config goes a long way.

Configuration must contain exactly one JSON object. Unknown fields and trailing
content are rejected; errors identify the configuration file.

## Strict builds

For CI publishing, run `leafpress build --strict`. Any build warning, including
broken wiki-links and fonts without a local source, fails the command before
replacing the last successful output. Strict builds never download fonts, so
commit `static/fonts/` after a normal build has downloaded your Google Fonts
families. Warning details are printed without
requiring `--verbose`. An initial failed strict build publishes no site.
Ordinary `leafpress build` still permits warnings. Strict mode applies to full
builds, not development-server incremental rebuilds.

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
    "favicon": "/static/images/icon.png",
    "headExtra": "<script defer data-domain=\"example.com\" src=\"https://plausible.io/js/script.js\"></script>"
  },

  "theme": {
    "preset": "classic",
    "fontHeading": "Bricolage Grotesque",
    "fontBody": "Inter",
    "fontMono": "JetBrains Mono",
    "accent": "#50ac00",
    "background": {
      "light": "#ffffff",
      "dark": "#1a1a1a"
    },
    "navStyle": "base",
    "navActiveStyle": "base",
    "listColumns": 2
  },

  "features": {
    "graph": true,
    "toc": true,
    "search": true,
    "wikilinks": true,
    "backlinks": true,
    "rss": true,
    "sharing": false
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
| `image` | `""` | Default Open Graph and Twitter image for social sharing |
| `favicon` | `""` | Single raster icon under `/static/`; empty uses built-in favicons |
| `headExtra` | `""` | Custom HTML injected into `<head>` (see [Custom Head Content](#custom-head-content)) |

### `theme`

| Option | Default | Description |
|--------|---------|-------------|
| `preset` | `"classic"` | Bundled visual theme: `"classic"`, `"aurora"`, `"paper"`, `"quiet"`, or `"terminal"` |
| `fontHeading` | `"Bricolage Grotesque"` | Heading font: a bundled family, a declared custom font, or any Google Fonts family (downloaded once and self-hosted) |
| `fontBody` | `"Inter"` | Body font, chosen the same way as `fontHeading` |
| `fontMono` | `"JetBrains Mono"` | Code font, chosen the same way as `fontHeading` |
| `fonts` | `[]` | Custom local font declarations (family, file under `static/fonts/`, weight, style, display, unicodeRange) — see [Theming](/guide/theming/) |
| `accent` | `"#50ac00"` | Accent color for links and highlights |
| `background.light` | `"#ffffff"` | Light mode background (color or gradient) |
| `background.dark` | `"#1a1a1a"` | Dark mode background (color or gradient) |
| `navStyle` | `"base"` | `"base"`, `"sticky"`, or `"glassy"` |
| `navActiveStyle` | `"base"` | `"base"`, `"box"`, or `"underlined"` |
| `listColumns` | `2` | List-page columns on desktop: `1`, `2`, or `3`. Mobile always uses one column |

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
| `rss` | `true` | Generate RSS feeds and show the feed icon in nav (requires `site.baseURL`). Writes the global `feed.xml` plus one feed per section and per tag; see [RSS feeds](#rss-feeds). |
| `sharing` | `false` | Show inline native-share (when supported) and copy-link buttons on pages |

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
      { "label": "Tags", "path": "/tags/" },
      { "label": "GitHub", "path": "https://github.com/shivamx96/leafpress" }
    ]
  }
}
```

Each `path` is either a site-relative path starting with `/` or an absolute
`http://` or `https://` URL. Site-relative paths are prefixed with the
`site.baseURL` path and highlight as the active link when the reader is on that
page. External URLs link out of the garden: they are rendered verbatim, open in
a new tab with `rel="noopener"`, and carry the `lp-nav-link--external` class
for styling. Other schemes such as `mailto:` are rejected.

### `build`

| Option | Default | Description |
|--------|---------|-------------|
| `outputDir` | `"_site"` | Build output directory |
| `port` | `3000` | Dev server port |
| `ignore` | `[]` | Glob patterns to exclude from builds, e.g. `["drafts/**", "*.draft.md", "private/**"]` |

### Ignore patterns

Patterns are matched against each path relative to the project root, using
forward slashes on every platform:

| Pattern | Matches |
|---------|---------|
| `drafts` | a file or folder named `drafts` at **any** depth, and everything inside it |
| `drafts/**` | everything under a top-level `drafts/`, and the folder itself |
| `*.draft.md` | that suffix at any depth — `a.draft.md`, `notes/b.draft.md` |
| `notes/*.wip.md` | only direct children of a top-level `notes/` |
| `a/**/tmp.md` | `tmp.md` anywhere beneath `a/` |

A pattern containing a `/` is **anchored** to the project root; one without a
`/` matches by name wherever it appears. Within a path segment, `*`, `?` and
`[a-z]` behave as usual; `**` additionally spans folder boundaries and must be
a whole segment. Naming a folder also excludes its contents, so `drafts` and
`drafts/**` are equivalent.

A malformed pattern (`draft-[0-9.md`, `a/**b/c`) fails the build rather than
silently matching nothing — otherwise a typo would publish the very drafts the
pattern was written to hold back.

`outputDir` must be a relative directory inside the project and cannot contain
`.` or `..` path segments. leafpress places a hidden `.leafpress-output`
ownership marker in generated output; it will not clean an existing non-empty
custom directory unless that directory was previously generated by leafpress.
Never point `outputDir` at a source or asset directory.

Full builds are failure-safe: leafpress generates the replacement in a hidden
sibling directory and publishes it only after every page and asset succeeds.
Generation failures leave the previous output untouched, and a failed final
promotion restores it before returning an error.

### Removed: `deploy`

Earlier versions accepted a `deploy` object. It is now rejected: delete the
block from `leafpress.json` and publish `_site/` with your hosting provider's
tooling. See the deployment guides for examples.

## RSS feeds

With `features.rss` enabled, leafpress writes three kinds of feed:

| Feed | Path | Contents |
|------|------|----------|
| Global | `/feed.xml` | Every page in the garden |
| Section | `/posts/feed.xml` | Every page under that folder, nested folders included |
| Tag | `/tags/idea/feed.xml` | Every page carrying that tag |

Each feed lists the 20 most recent pages, newest first, using the `modified`
date when present and `date` otherwise. Section feeds take their title from
the section's `_index.md`, or from the folder name when there is none. Tag
feeds use the lowercase tag name that the tag page already uses.

The nav icon and the `<link rel="alternate">` in every page's `<head>` point
at the global feed. Section homes and tag pages additionally advertise their
own feed, so a reader app pointed at `https://example.com/posts/` discovers
`https://example.com/posts/feed.xml` on its own.

## A note on `baseURL`, sitemap & RSS

`sitemap.xml`, the RSS feeds (`feed.xml` and the per-section and per-tag
feeds), and the `Sitemap:` line in `robots.txt` all need an absolute origin.
If `site.baseURL` is empty, leafpress **skips** those artifacts (and prints a
warning) rather than emitting invalid relative URLs. Set `site.baseURL` to
your production URL to enable them.

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

## Favicons and social images

Set `site.favicon` to `/static/images/icon.png` (or a JPEG/WebP) to use one
image for both the browser icon and Apple touch icon. Put the file at
`static/images/icon.png`; leafpress copies its original bytes. It skips the
built-in `favicon.svg`, `favicon-96x96.png`, and `favicon.ico` in this mode,
including any project-root overrides. No conversion is performed. Leave
`favicon` empty to retain the built-in set and root-file override behavior.
CLI builds require custom favicons under `/static/`.

`site.image` supplies the social image for home, section and tag pages, and
is the fallback for notes without a frontmatter `image`. Notes with their own
image take precedence. Images produce matching Open Graph and Twitter image
tags and a `summary_large_image` card; pages without an image use `summary`.

Both site fields require a site-relative path starting with `/`, without a
scheme or host, `..` segments, query strings or fragments. The site's base
path is added automatically. For hosted rendering, declare uploaded images in
the caller asset manifest; a missing `/static/` reference produces a warning.
