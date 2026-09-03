# leafpress

A fast, opinionated static site generator for digital gardens.

## Install

```bash
curl -fsSL https://leafpress.in/install.sh | sh
```

Or with Go:
```bash
go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest
```

## Quick Start

```bash
mkdir my-garden
cd my-garden
leafpress init
leafpress serve
```

Visit `http://localhost:3000` to see your site live.

## Publish

Build the portable static site, then publish `_site/` with your hosting
provider's CLI or CI integration:

```bash
leafpress build
```

See the guides for [GitHub Pages](https://leafpress.in/guide/deploy-github),
[Vercel](https://leafpress.in/guide/deploy-vercel), and
[Netlify](https://leafpress.in/guide/deploy-netlify).

## Features

- Wiki-links with automatic backlinks
- Full-text search
- Graph visualization
- Table of contents
- Footnotes
- Callouts (Obsidian-compatible)
- Mermaid diagrams
- YouTube auto-embeds
- Local video and audio embeds
- Obsidian image width syntax (`![[image.png|500]]`)
- RSS feed with nav icon
- Dark mode with system preference detection
- Link previews on hover
- Design system with CSS custom properties (font scale, border radius)
- SEO ready (sitemap, RSS, Open Graph)
- Fast parallel builds with incremental rebuilds during `serve`
- Portable static output for any hosting provider

## Documentation

Full docs at [leafpress.in](https://leafpress.in)

Embedding the pure renderer: [leafpress-render contract](docs/05_RENDERER_CONTRACT.md)

Maintainers: [multi-module release process](docs/06_RELEASE_PROCESS.md)

## License

MIT
