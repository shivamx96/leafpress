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
    "fontHeading": "Playfair Display",
    "fontBody": "Source Sans Pro",
    "accent": "#e11d48"
  }
}
```

## Fonts

Choose any [Google Font](https://fonts.google.com/):

```json
{
  "theme": {
    "fontHeading": "Crimson Pro",
    "fontBody": "Inter", 
    "fontMono": "Fira Code"
  }
}
```

Popular combinations:
- **Classic**: Crimson Pro + Inter
- **Modern**: Geist + Geist
- **Technical**: IBM Plex Sans + IBM Plex Mono
- **Elegant**: Playfair Display + Lora

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

For deeper customization, create `style.css` in your site root. It completely replaces the default stylesheet.

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

