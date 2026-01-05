# Changelog

All notable changes to leafpress will be documented in this file.

## [1.0.0] - 2026-01-05

**leafpress is ready for production.** A fast, opinionated static site generator purpose-built for digital gardens.

### Core Philosophy

> Your garden folder IS the product. leafpress is invisible infrastructure.

Transform a folder of Markdown files into a beautiful, interlinked website. Single binary, zero runtime dependencies, minimal configuration.

### Features

#### Content & Linking
- **Wiki-style links** — Use `[[Page Name]]` syntax to link between notes
- **Automatic backlinks** — Every page shows which other pages link to it
- **Smart link resolution** — Case-insensitive matching, cross-directory links, custom display text with `[[slug|Display Text]]`
- **Broken link detection** — Build warnings for links that don't resolve
- **Code-aware parsing** — Wiki-links inside code blocks are left untouched

#### Knowledge Organization
- **Growth stages** — Mark pages as seedling 🌱, budding 🌿, or evergreen 🌳
- **Tags** — Full tag support with auto-generated `/tags/` pages
- **Section indexes** — Auto-generated or customizable with `_index.md`
- **Flexible sorting** — Sort sections by date, title, or growth stage
- **Draft mode** — Exclude work-in-progress with `draft: true`

#### Graph & Search
- **Knowledge graph** — Exports `graph.json` with all pages and their connections
- **Search index** — Generates `search-index.json` for client-side search
- **Visual exploration** — Graph data includes growth stages and tags for filtering

#### Theming & Customization
- **Google Fonts** — Any font family for headings, body, and code
- **Accent colors** — Custom hex color for links and highlights
- **Background gradients** — Solid colors or CSS gradients, separate light/dark
- **Navigation styles** — Base, sticky, or floating glass pill
- **Active link styles** — Base, boxed, or underlined
- **Dark mode** — Built-in light/dark theme support
- **Custom CSS** — Override anything with `style.css`
- **Custom favicons** — Drop in your own `.ico`, `.svg`, or `.png`

#### Developer Experience
- **Live reload** — WebSocket-powered instant refresh on file changes
- **Incremental builds** — Only rebuilds what changed during development
- **Parallel processing** — Multi-core rendering for fast builds
- **Verbose mode** — Detailed build timing and diagnostics
- **Auto port detection** — Finds available port if default is in use

#### Obsidian Compatibility
- **Wiki-link syntax** — Same `[[note]]` format
- **Image embeds** — Supports `![[image.png]]` notation
- **Frontmatter aliases** — Recognizes `created`, `createdAt`, `modified`, `updated`, `updatedAt`
- **Vault-friendly** — Ignores `.obsidian/` directory automatically

#### Table of Contents
- **Auto-generated** — Extracts H2 and H3 headings
- **Sticky sidebar** — Fixed position on desktop (1280px+)
- **Scroll tracking** — Highlights current section
- **Per-page control** — Override with `toc: false` in frontmatter

### CLI Commands

```
leafpress init                    Create a new garden
leafpress new <path>              Create a new page with frontmatter
leafpress build [--drafts]        Generate static site
leafpress serve [--port N]        Start development server
```

### Configuration

Minimal `leafpress.json`:

```json
{
  "title": "My Digital Garden",
  "baseURL": "https://example.com"
}
```

Full options:

```json
{
  "title": "My Digital Garden",
  "author": "Your Name",
  "baseURL": "https://example.com",
  "outputDir": "_site",
  "port": 3000,
  "nav": [
    {"label": "Notes", "path": "/notes/"},
    {"label": "Projects", "path": "/projects/"}
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
    "navStyle": "glassy",
    "navActiveStyle": "underlined"
  },
  "ignore": ["drafts", "private"],
  "toc": true,
  "graph": true,
  "search": true,
  "backlinks": true,
  "wikilinks": true
}
```

### Frontmatter

```yaml
---
title: "Page Title"
date: 2026-01-05
tags: [digital-garden, tools]
draft: false
growth: "evergreen"
toc: true
---
```

### Performance

Benchmarked against Hugo (the fastest mainstream SSG):

| Pages | leafpress | Hugo | Difference |
|-------|-----------|------|------------|
| 1,000 | 98ms | 154ms | 36% faster |
| 2,000 | 171ms | 271ms | 37% faster |

Single binary. No Node.js. No Ruby. No dependencies.

### Deployment

Works with any static host:
- Netlify
- Vercel
- GitHub Pages
- Cloudflare Pages
- Any server that serves HTML

### Getting Started

```bash
# Install
go install github.com/shivamx96/leafpress/cmd/leafpress@latest

# Create your garden
mkdir my-garden && cd my-garden
leafpress init

# Start writing
leafpress new "notes/My First Note"

# Preview locally
leafpress serve --drafts

# Build for production
leafpress build
```

### Built With

- [Go](https://go.dev/) — Fast, compiled, single binary
- [Goldmark](https://github.com/yuin/goldmark) — CommonMark-compliant Markdown
- [Chroma](https://github.com/alecthomas/chroma) — Syntax highlighting
- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [fsnotify](https://github.com/fsnotify/fsnotify) — File watching

---

This is v1.0.0. The foundation is complete. Future releases will add RSS feeds, sitemaps, and an Obsidian plugin for one-click publishing.
