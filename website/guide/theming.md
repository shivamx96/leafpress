---
title: "Theming"
date: 2025-12-21
---

Customize your site's appearance with fonts, colors, and custom CSS.

## Quick Theming

Set options in `leafpress.json`:

```json
{
  "theme": {
    "fontHeading": "Bricolage Grotesque",
    "fontBody": "Inter",
    "accent": "#e11d48"
  }
}
```

## Fonts

Leafpress sites are **self-hosted by default**: no request ever leaves your
site for a font. Three families ship bundled with Leafpress — **Bricolage
Grotesque** (headings), **Inter** (body), and **JetBrains Mono** (code) — and are served
from your own site as `@font-face` rules in the generated stylesheet:

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

Any other family name produces a build warning and falls back to the CSS
system stacks — Leafpress no longer loads arbitrary fonts from Google.

### Migrating from Google Fonts

If your existing config names a Google font (say `"Playfair Display"`), you
have two options:

1. **Recommended**: download the font's woff2 file and declare it as a
   custom local font (see below) — your site stays self-contained.
2. **Temporary bridge**: set `"remoteFonts": true` in `theme` to keep the
   old Google Fonts link for unbundled families. This option is deprecated
   and will be removed; names are matched exactly, and bundled families are
   always self-hosted regardless.

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
- Files must live under `static/fonts/` with a `.woff2`, `.woff`, `.ttf`, or
  `.otf` extension, and file names may only use letters, digits, `-`, `.`,
  `_`, and `~`. The build fails if a declared file is missing.
- Family names are matched **exactly** (case-sensitive) against
  `fontHeading`/`fontBody`/`fontMono`. Declared families are self-hosted:
  they never load remotely, even under `remoteFonts`.

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

## Custom CSS

For deeper customization, create `style.css` in your site root. Leafpress **appends** it after the default stylesheet (and any self-hosted `@font-face` rules), so you can override variables and classes without rewriting the base theme.

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
