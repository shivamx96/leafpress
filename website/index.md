---
title: "leafpress"
date: 2025-12-21
---

A fast, opinionated static site generator for digital gardens. [GitHub](https://github.com/shivamx96/leafpress)

```bash
curl -fsSL https://leafpress.in/install.sh | sh
```

To initialize a new garden and see it live:
```bash
mkdir my-garden
cd my-garden
leafpress init
leafpress serve
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
- **Mermaid diagrams** rendered from fenced code blocks
- **YouTube auto-embeds** from pasted URLs
- **Video & audio embeds** via Obsidian syntax (`![[video.mp4]]`)
- **Image width control** with `![[image.png|500]]`
- **RSS feed** with nav icon (toggle in config)
- **SEO ready** with Open Graph and meta tags, plus `sitemap.xml` when `baseURL` is set
- **Callouts** for notes, warnings, tips (`> [!note]`)
- **Syntax highlighting** for code blocks
- **Dark mode** with system preference detection
- **Link previews** on hover
- **Design system** with CSS custom properties for fonts, radii, and spacing

## Performance

Clean-build times in milliseconds on **v1.0.0-beta.17** using the deterministic
workload v2 garden (median of 10 measured runs after 2 warmups):

**Docker on Apple M3, 24GB host — container limited to 4 CPUs and 8GB RAM**

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|-----|------------|------------|
| zola | 21 | 75 | 138 |
| hugo | 36 | 124 | 219 |
| leafpress-minimal | 23 | 97 | 171 |
| leafpress | 25 | 111 | 203 |
| eleventy | 258 | 523 | 805 |
| jekyll | 155 | 278 | 419 |

*leafpress-minimal: all extra features disabled (comparable to Hugo/Zola).*
*leafpress: full features including wikilinks, backlinks, graph, search, and TOC.*

In this run, full-featured leafpress is the **second-fastest generator behind
Zola** and stays ahead of Hugo at every tested size. The fixture models a
hierarchical garden with Notes, Posts, Tags, deterministic links, code blocks,
and orphan pages. [See the full methodology and native/Docker reports.](https://github.com/shivamx96/leafpress/tree/main/benchmark/results/2026-08-11)

## Quick Start

[[guide/installation|Get started in 5 minutes →]]
