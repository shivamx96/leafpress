---
title: "Examples"
date: 2025-12-21
---

See leafpress in action with these example sites.

## Digital Garden

A personal digital garden with interconnected notes and thoughts.

**Features Used:**
- Wiki links with backlinks
- Tags
- Growth stages
- Table of contents
- Custom gradient backgrounds

**Config highlights:**
```json
{
  "toc": true,
  "theme": {
    "background": {
      "light": "linear-gradient(135deg, #e8f5e9 0%, #ffffff 100%)",
      "dark": "linear-gradient(135deg, #0d1f12 0%, #1a1a1a 100%)"
    }
  }
}
```

## Documentation Site

Technical documentation with clear navigation and search.

**Features Used:**
- Sticky navigation
- Table of contents
- Code syntax highlighting
- Section organization

**Config highlights:**
```json
{
  "theme": {
    "navStyle": "sticky"
  },
  "toc": true
}
```

## Personal Blog

A clean blog with posts organized by tags and dates.

**Features Used:**
- Date-based sorting
- Tag organization
- Custom accent color
- Gradient backgrounds

**Config highlights:**
```json
{
  "theme": {
    "accent": "#ff6b6b",
    "background": {
      "light": "linear-gradient(135deg, #fff5f5 0%, #ffffff 100%)",
      "dark": "linear-gradient(135deg, #2d1818 0%, #1a1a1a 100%)"
    }
  }
}
```

## Knowledge Base

An internal knowledge base with extensive cross-linking.

**Features Used:**
- Wiki links everywhere
- Backlinks on every page
- Table of contents
- Section organization

**Config highlights:**
```json
{
  "toc": true,
  "theme": {
    "navStyle": "sticky"
  }
}
```

## Create Your Own

Ready to build your own site?

1. [[guide/installation|Install leafpress]]
2. [[guide/configuration|Configure your site]]
3. [[guide/writing|Start writing]]

Have a site built with leafpress? [Submit it on GitHub](https://github.com/shivamx96/leafpress) to be featured here!
