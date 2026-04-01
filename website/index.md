---
title: "leafpress"
date: 2025-12-21
---

A fast, opinionated static site generator for digital gardens. [GitHub](https://github.com/shivamx96/leafpress)

```bash
curl -fsSL https://leafpress.in/install.sh | sh
leafpress init my-garden
cd my-garden && leafpress serve
```

## Why leafpress?

Most static site generators make you choose: simple but limited, or powerful but complex. leafpress is different.

**Write in Obsidian, publish anywhere.** Your `[[wiki-links]]` just work. Backlinks are automatic. No plugins, no configuration, no friction.

**Fast by default.** Single binary, no dependencies. Builds hundreds of pages in milliseconds. Live reload that actually feels instant.

**Looks good out of the box.** Beautiful typography, dark mode, responsive design. Customize with a few lines of JSON, or go deeper with CSS.

## Features

- **[[guide/wiki-links|Wiki-style linking]]** with automatic backlinks
- **Full-text search** built-in, no external services
- **Graph visualization** of your knowledge connections
- **Table of contents** auto-generated from headings
- **SEO ready** with sitemap, RSS, Open Graph, and meta tags
- **Callouts** for notes, warnings, tips (`> [!note]`)
- **Syntax highlighting** for code blocks
- **Dark mode** with system preference detection
- **Link previews** on hover

## Performance

Build times in milliseconds (median of 10 runs) as of Apr 1, 2026 on **v1.0.0-beta.6**:

**Docker on Apple M3, 24GB RAM**

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|-----|------------|------------|
| zola | 25 | 76 | 135  |
| hugo | 40| 128 | 224  |
| leafpress-minimal | 24| 89 | 145  |
| leafpress | 30| 125 | 233  |
| eleventy | 266| 555 | 836  |
| jekyll | 175| 332 | 499  |

*leafpress-minimal: all extra features disabled (comparable to Hugo/Zola).*
*leafpress: full features including wikilinks, backlinks, graph, search, and TOC.*

On apples-to-apples comparison, we're literally the **2nd fastest just behind zola**. The cost of digital gardens is real and it adds up taking us to **hugo's** level once we include everything that matters.

## Quick Start

[[guide/installation|Get started in 5 minutes →]]
