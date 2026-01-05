# Changelog

All notable changes to leafpress will be documented in this file.

## [1.0.0] - 2026-01-05

First stable release. A fast, opinionated static site generator for digital gardens.

### Added

- **Wiki-style linking** with `[[Page Name]]` syntax and automatic backlinks
- **Growth stages** — seedling 🌱, budding 🌿, evergreen 🌳
- **Tags** with auto-generated tag index and individual tag pages
- **Section indexes** — auto-generated or customizable via `_index.md`
- **Knowledge graph** export (`graph.json`) with page connections
- **Search index** generation (`search-index.json`) for client-side search
- **Table of contents** — sticky sidebar extracted from H2/H3 headings
- **Theming** — Google Fonts, accent colors, light/dark mode, nav styles
- **Custom CSS** support via `style.css`
- **Live reload** dev server with WebSocket
- **Incremental builds** during development
- **Parallel rendering** for fast builds
- **Obsidian compatibility** — wiki-links, image embeds, frontmatter aliases
- **Draft mode** — exclude pages with `draft: true`
- **Custom favicons** — override defaults with your own

### Performance

- 1,000 pages in 98ms (36% faster than Hugo)
- 2,000 pages in 171ms (37% faster than Hugo)
- Single binary, zero runtime dependencies
