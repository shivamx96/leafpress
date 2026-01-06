---
title: "leafpress"
date: 2025-12-21
---

A fast, opinionated static site generator for digital gardens.

```bash
go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest
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

Build times in milliseconds (median of 10 runs):

**Apple M3, 24GB RAM**

| pages | zola | leafpress | hugo | eleventy | jekyll |
|-------|------|-----------|------|----------|--------|
| 100   | 55   | 51        | 142  | 246      | 268    |
| 1000  | 172  | 199       | 307  | 508      | 513    |
| 2000  | 330  | 347       | 494  | 816      | 776    |

leafpress runs with all features enabled (wikilinks, backlinks, graph, TOC). Closer to zola, faster than hugo, eleventy, and jekyll.

## Quick Start

[[guide/installation|Get started in 5 minutes →]]
