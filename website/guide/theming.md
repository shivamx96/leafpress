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

