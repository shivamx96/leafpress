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
leafpress init my-garden
cd my-garden
leafpress serve
```

Visit `http://localhost:3000` to see your site live.

## Deploy

One-command deployment to your choice of hosting:

```bash
leafpress deploy
```

Supports:
- **GitHub Pages** – Free hosting for public repos
- **Vercel** – Fast edge CDN with automatic deployments
- **Netlify** – Global CDN with one-click rollbacks

See the [deployment guide](https://leafpress.in/guide/deploy-github) for details.

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
- Fast builds (~125ms for 1000 pages)
- One-command deploy to GitHub Pages, Vercel, or Netlify

## Documentation

Full docs at [leafpress.in](https://leafpress.in)

Embedding the pure renderer: [leafpress-render contract](docs/05_RENDERER_CONTRACT.md)

Maintainers: [multi-module release process](docs/06_RELEASE_PROCESS.md)

## License

MIT
