---
title: "Custom Styles"
date: 2025-12-21
---

Customize your site's appearance beyond the built-in theme options.

## Custom CSS File

leafpress looks for a `style.css` file in your site's root directory. If present, it will be used instead of the default stylesheet.

### Creating a Custom Stylesheet

1. **Copy the default styles** (optional starting point):

```bash
# After running 'leafpress init', style.css is created
# Modify it to customize your site
```

2. **Start from scratch**:

Create `style.css` in your site's root:

```css
/* Your custom styles */
body {
  font-family: Georgia, serif;
  background: #f5f5f5;
}

/* Override specific elements */
.lp-nav {
  background: white;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
```

## CSS Variables

The default theme uses CSS variables that you can override:

```css
:root {
  --lp-accent: #50ac00;
  --lp-bg: #ffffff;
  --lp-text: #1a1a1a;
  --lp-text-muted: #666666;
  --lp-border: #e5e5e5;
  --lp-code-bg: #f7f7f7;
  --lp-max-width: 680px;
  --lp-nav-height: 60px;
  
  --lp-font-heading: "Crimson Pro", Georgia, serif;
  --lp-font-body: "Inter", system-ui, sans-serif;
  --lp-font-mono: "JetBrains Mono", monospace;
}
```

## Common Customizations

### Change Content Width

```css
:root {
  --lp-max-width: 800px; /* Default is 680px */
}
```

### Custom Link Styles

```css
.lp-content a {
  color: #0066cc;
  text-decoration: underline;
}

.lp-content a:hover {
  color: #0052a3;
}
```

### Custom Wiki Link Appearance

```css
.lp-wikilink {
  background-color: rgba(80, 172, 0, 0.1);
  padding: 0.1em 0.3em;
  border-radius: 4px;
  font-weight: 500;
}
```

### Custom Code Block Styling

```css
.lp-content pre {
  background-color: #1e1e1e;
  border-left: 4px solid var(--lp-accent);
}

.lp-content code {
  font-size: 0.95em;
  font-family: "Fira Code", monospace;
}
```

### Custom Table of Contents

```css
.lp-toc {
  background: #f9f9f9;
  padding: 1rem;
  border-radius: 8px;
}

.lp-toc-link {
  font-size: 0.9rem;
  padding: 0.25rem 0;
}
```

### Dark Mode Overrides

```css
[data-theme="dark"] {
  --lp-bg: #0a0a0a;
  --lp-text: #f0f0f0;
  --lp-border: #2a2a2a;
}

/* Custom dark mode styles */
[data-theme="dark"] .lp-nav {
  border-bottom: 1px solid #333;
}
```

## CSS Class Reference

leafpress uses the `lp-` prefix for all CSS classes:

### Layout
- `.lp-body` - Body element
- `.lp-nav` - Navigation bar
- `.lp-main` - Main content container
- `.lp-footer` - Footer

### Navigation
- `.lp-nav-container` - Nav inner wrapper
- `.lp-nav-brand` - Brand/title section
- `.lp-nav-title` - Site title link
- `.lp-nav-links` - Navigation links container
- `.lp-nav-link` - Individual nav link

### Content
- `.lp-article` - Article wrapper
- `.lp-header` - Page header
- `.lp-title` - Page title
- `.lp-content` - Main content area
- `.lp-meta` - Metadata (date, growth)

### Table of Contents
- `.lp-toc` - TOC container
- `.lp-toc-nav` - TOC navigation
- `.lp-toc-list` - TOC list
- `.lp-toc-item` - TOC item
- `.lp-toc-link` - TOC link

### Links
- `.lp-wikilink` - Wiki-style link
- `.lp-broken-link` - Broken wiki link
- `.lp-external` - External link

### Section Pages
- `.lp-section` - Section wrapper
- `.lp-section-title` - Section title
- `.lp-section-intro` - Section intro content
- `.lp-index` - Index list
- `.lp-index-item` - Index item

### Tags
- `.lp-tags` - Tags container
- `.lp-tag` - Individual tag
- `.lp-tag-cloud` - Tag cloud

## Adding Custom Fonts

Use Google Fonts or custom web fonts:

```css
@import url('https://fonts.googleapis.com/css2?family=Merriweather:wght@400;700&display=swap');

:root {
  --lp-font-body: 'Merriweather', Georgia, serif;
}
```

Or load local fonts:

```css
@font-face {
  font-family: 'MyCustomFont';
  src: url('/static/fonts/custom.woff2') format('woff2');
}

:root {
  --lp-font-heading: 'MyCustomFont', sans-serif;
}
```

## Responsive Design

Override mobile styles:

```css
@media (max-width: 768px) {
  .lp-main {
    padding: 1rem;
  }
  
  .lp-title {
    font-size: 1.75rem;
  }
}
```

## Next Steps

- [[guide/configuration|Theme Configuration]]
- [[guide/writing|Writing Content]]
- [[features|Explore Features]]
