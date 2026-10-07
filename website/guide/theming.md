---
title: "Theming"
date: 2025-12-21
---

Customize your site's appearance with fonts, colors, and custom CSS.

## Bundled Themes

Select a bundled visual theme with `theme.preset`:

```json
{
  "theme": {
    "preset": "aurora"
  }
}
```

leafpress ships five presets:

- `classic` is the default and preserves leafpress's original, focused reading
  experience.
- `aurora` is an expressive composition with a gradient canvas, floating glass
  navigation, layered reading surfaces, card-based indexes and backlinks, and
  elevated search and graph panels. It includes coordinated light and dark
  appearances.
- `paper` is an editorial, print-inspired composition with serif typography,
  ruled reading sheets, sharp geometry, marginalia-like callouts, tabular
  indexes, and document-style search and graph panels. It includes coordinated
  light and dark appearances.
- `quiet` is a minimal personal-garden theme with Inter typography, a narrow
  centered reading column, plain lists with titles on the left and muted dates
  on the right, understated links, and compact section spacing with comfortable
  line height. Light and dark appearances keep callouts, backlinks, search, and
  graph controls subdued. Use `"listColumns": 1` for a simple writing index.
- `terminal` is a compact, command-line-inspired workspace with prompt-marked
  headings, path-like navigation, file-list indexes, structured log callouts,
  session-style code blocks, and diagnostic search and graph panels. Its light
  appearance resembles printed terminal output; its dark appearance uses a
  restrained phosphor palette.

Presets are complete component themes rather than color-only variations.
They share leafpress's semantic type scale, so switching presets preserves
heading hierarchy, body size, and component text sizing. Presets express their
typographic identity through font family, weight, tracking, and composition.

Each preset supplies defaults for fonts, accent, backgrounds, and navigation.
Explicit values in `leafpress.json` override those preset defaults, and a
project's `style.css` remains the final, authoritative layer.

## Quick Theming

Set options in `leafpress.json`:

```json
{
  "theme": {
    "preset": "aurora",
    "fontHeading": "Space Grotesk",
    "fontBody": "Inter",
    "accent": "#e11d48"
  }
}
```

## Fonts

leafpress sites are **self-hosted by default**: no request ever leaves your
site for a font. leafpress bundles the seven families its themes use, so a
new garden renders offline, and serves only the families your theme selects
as `@font-face` rules in the generated stylesheet:

- Headings: **Bricolage Grotesque**, **Space Grotesk**, **Newsreader**
- Long-form text: **Inter**, **Source Serif 4**
- Code: **JetBrains Mono**, **IBM Plex Mono**

Any other [Google Fonts](https://fonts.google.com) family works too: leafpress
downloads it once and self-hosts it (see [Any Google Font](#any-google-font)).

leafpress also preloads one normal face for each selected family in theme role
order (`fontHeading`, `fontBody`, then `fontMono`). Reused families and files
are deduplicated; italic and extended-Latin faces remain demand-loaded.

These groups are recommendations, not validation restrictions; any family
can be assigned to any theme role:

```json
{
  "theme": {
    "fontHeading": "Bricolage Grotesque",
    "fontBody": "Inter",
    "fontMono": "JetBrains Mono"
  }
}
```

The bundled files cover the latin and latin-ext character ranges. Text in
other scripts (Cyrillic, Greek, Vietnamese, …) falls back to your readers'
system fonts; use a custom local font if you need full coverage for another
script.

### Any Google Font

Name any [Google Fonts](https://fonts.google.com) family and leafpress
self-hosts it:

```json
{
  "theme": {
    "fontHeading": "Playfair Display"
  }
}
```

The first `leafpress build` or `leafpress serve` that needs the family
downloads it once:

```
Downloading "Playfair Display" from Google Fonts...
  saved static/fonts/playfair-display/ (5 files, 123 KB, SIL Open Font License 1.1)
  Commit static/fonts/ so later builds work offline.
```

- leafpress saves the family's Latin and Latin Extended files, every weight
  and italic the themes use, and its license under `static/fonts/<family>/`.
  `static/fonts/fonts.lock.json` records each file with a SHA-256 checksum.
- Later builds read those files and never contact Google. Commit
  `static/fonts/` with your garden so CI and other machines build offline
  and produce identical output.
- Your readers' browsers load the fonts from your site, not from Google.
- Family names are case-sensitive, as on Google Fonts. A misspelled name
  keeps the system fonts and suggests the closest match:
  `font family "Playfiar Display" is not a Google Fonts family (did you mean
  "Playfair Display"?)`.
- During `leafpress serve`, changing a font in `leafpress.json` downloads it
  and reloads the preview.
- If the download fails, for example offline, the build warns, uses system
  fonts, and tries again next time.
- `leafpress build --strict` never downloads. A font that is not downloaded
  yet is a warning, so strict CI builds fail rather than fetching or falling
  back.
- `--offline` on `build` and `serve`, or `LEAFPRESS_OFFLINE=1`, turns
  downloads off.
- If a downloaded file changes, the build stops rather than overwrite it.
  Delete the family's folder and build again to download it afresh.

### Migrating from `remoteFonts`

The `remoteFonts` option, which linked unbundled families from Google Fonts
at read time, has been removed, and a config that still sets it fails to
load with an error saying so. Delete it and build once: the same families
are downloaded and self-hosted instead.

### Custom local fonts

Ship your own font files under `static/fonts/` and declare them in `theme`:

```json
{
  "theme": {
    "fontBody": "My Serif",
    "fonts": [
      {
        "family": "My Serif",
        "file": "static/fonts/my-serif.woff2",
        "weight": "400 700",
        "style": "normal",
        "display": "swap"
      }
    ]
  }
}
```

- `family` and `file` are required; `weight` (a number or a variable range),
  `style` (`normal`/`italic`/`oblique`), and `display` default to `400`,
  `normal`, and `swap`.
- `unicodeRange` (optional) limits a file to a character subset, as a CSS
  `unicode-range` list such as `"U+0000-00FF, U+0131"`. Use it to declare a
  family split into several subset files; without it, each later file of the
  same family, weight, and style replaces the earlier one.
- Files must live under `static/fonts/` with a `.woff2`, `.woff`, `.ttf`, or
  `.otf` extension, and file names may only use letters, digits, `-`, `.`,
  `_`, and `~`. The build fails if a declared file is missing.
- Family names are matched **exactly** (case-sensitive) against
  `fontHeading`/`fontBody`/`fontMono`. Declared families are used as
  declared and never downloaded.

## Colors

### Accent Color

Used for links, active states, and highlights:

```json
{
  "theme": {
    "accent": "#50ac00"
  }
}
```

### Backgrounds

Solid colors or gradients:

```json
{
  "theme": {
    "background": {
      "light": "#ffffff",
      "dark": "#0a0a0a"
    }
  }
}
```

```json
{
  "theme": {
    "background": {
      "light": "linear-gradient(180deg, #fefefe 0%, #f0f0f0 100%)",
      "dark": "linear-gradient(180deg, #0a0a0a 0%, #1a1a1a 100%)"
    }
  }
}
```

## Navigation Style

### Nav Position

```json
{
  "theme": {
    "navStyle": "base"
  }
}
```

- `"base"` — Standard navigation bar (default)
- `"sticky"` — Fixed bar at top
- `"glassy"` — Glassmorphic blur effect (appears as floating pill on scroll)

On mobile, the site title stays on one line and section links scroll horizontally.
Search remains beside the title; the **Site menu** contains theme switching,
the knowledge graph, and RSS when enabled. Long titles are shortened visually
with an ellipsis. Sticky and glassy bars keep heading links clear of the header
and adjust their spacing when the viewport or fonts change.
While scrolling on mobile, glassy navigation becomes a single-row floating bar;
section links move into the Site menu until you return to the top.
The floating bar sizes to its contents, with a width cap and extra side clearance,
so it remains visually separate from the article.

### Active Link Style

```json
{
  "theme": {
    "navActiveStyle": "base"
  }
}
```

- `"base"` — No special styling (default)
- `"underlined"` — Underline on active link
- `"box"` — Background box on active link

## List Columns

Choose one, two, or three columns for section and tag-page lists on desktop:

```json
{
  "theme": {
    "listColumns": 3
  }
}
```

The default is `2`. On mobile, leafpress always uses a single column.

## Custom CSS

For deeper customization, create `style.css` in your site root. leafpress
composes the shared base, selected bundled theme, self-hosted `@font-face`
rules, and finally your custom stylesheet. Because `style.css` is last, it can
override variables and classes without rewriting the selected theme.

### Starting Point

Run `leafpress init` to get a `style.css` you can modify. Or start from scratch using CSS variables:

```css
:root {
  --lp-font-heading: "Your Font", serif;
  --lp-font-body: "Your Font", sans-serif;
  --lp-font-mono: "Your Font", monospace;
  --lp-accent: #50ac00;
  --lp-bg: #ffffff;
  --lp-text: #1a1a1a;
  --lp-text-muted: #666666;
  --lp-border: #e5e5e5;
  --lp-code-bg: #f7f7f7;
  --lp-max-width: 680px;
}

[data-theme="dark"] {
  --lp-bg: #1a1a1a;
  --lp-text: #e5e5e5;
  --lp-text-muted: #a0a0a0;
  --lp-border: #333333;
  --lp-code-bg: #2a2a2a;
}
```

### Font Size Scale

All font sizes use a consistent type scale via CSS variables. Override any of these to adjust sizing globally:

```css
:root {
  --lp-font-xs: 0.75rem;      /* badges, tooltips, kbd shortcuts */
  --lp-font-sm: 0.875rem;     /* nav, tags, meta, footer, code blocks, TOC */
  --lp-font-base: 1rem;       /* body text, search input */
  --lp-font-lg: 1.25rem;      /* h3, h4, blockquotes */
  --lp-font-xl: 1.5rem;       /* h2 */
  --lp-font-2xl: 1.75rem;     /* h1 */
  --lp-font-3xl: 2rem;        /* page titles */
  --lp-font-display: 6rem;    /* decorative (404 page) */
}
```

To scale all text up or down proportionally, override the variables with `calc()`:

```css
:root {
  --lp-font-xs: calc(0.75rem * 1.1);
  --lp-font-sm: calc(0.875rem * 1.1);
  --lp-font-base: calc(1rem * 1.1);
  --lp-font-lg: calc(1.25rem * 1.1);
  --lp-font-xl: calc(1.5rem * 1.1);
  --lp-font-2xl: calc(1.75rem * 1.1);
  --lp-font-3xl: calc(2rem * 1.1);
}
```

### Border Radius

All border radii use a consistent scale:

```css
:root {
  --lp-radius-sm: 4px;       /* inline code, buttons, tooltips, badges */
  --lp-radius-md: 8px;       /* code blocks, callouts, cards, link previews */
  --lp-radius-lg: 12px;      /* overlay panels (graph, search) */
  --lp-radius-full: 9999px;  /* pill-shaped elements */
}
```

### CSS Classes

Key classes you might want to customize:

- `.lp-nav` — Navigation bar
- `.lp-content` — Main content area
- `.lp-article` — Article container
- `.lp-wikilink` — Wiki links
- `.lp-backlinks` — Backlinks section
- `.lp-toc` — Table of contents
- `.lp-callout` — Callout boxes
- `.lp-graph` — Graph container
- `.lp-search` — Search component
