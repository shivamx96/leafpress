---
title: "Changelog"
date: 2025-01-06
toc: false
---

## v1.0.0-beta.7
*April 3, 2026*

- Footnotes support via `[^1]` syntax with auto-numbered references and back-links
- Anchored headings: hover over h2/h3 to reveal a `#` deep-link
- Fixed wiki-links inside inline code being processed as links
- Fixed TOC links not matching heading IDs for headings with special characters (e.g. `&`)
- Combined Google Fonts into a single request — eliminates 2 render-blocking network calls

## v1.0.0-beta.6
*April 2, 2026*

- YouTube auto-embed: paste a YouTube URL on its own line and it renders as a responsive iframe (privacy-enhanced via youtube-nocookie.com)
- Mermaid diagram support: fenced `mermaid` code blocks render client-side with dark mode via CSS invert filter
- Local video/audio embeds: `![[video.mp4]]` and `![[recording.mp3]]` render as native `<video>` and `<audio>` elements
- Dev server now uses `http.ServeFile` for media — enables range requests for video/audio seeking
- Performance: switched to `text/template` — ~40% faster template rendering (1000 pages in ~125ms)
- Benchmark: added `leafpress-minimal` config for fair comparison against Hugo/Zola

## v1.0.0-beta.5
*March 29, 2026*

- Auto-detect system dark/light mode preference on the first visit
- Switching to a newer, optimized and redesigned favicon

## v1.0.0-beta.4
*March 26, 2026*

- RSS feed with configurable toggle (`"rss": true/false` in config)
- RSS icon in nav bar with auto-discovery link in `<head>`
- Obsidian image width syntax support (`![[image.png|500]]`)
- XML content type in dev server for proper feed rendering

## v1.0.0-beta.3
*March 20, 2026*

- Design system: font sizes, border radii, and spacing now use CSS custom properties for easy global overrides
- New theming docs with font scale and proportional scaling examples
- Fixed directory traversal vulnerability in dev server path handling
- Fixed Netlify deploy cancellation silently reporting success
- Fixed `--provider` flag loading credentials for wrong provider
- Deterministic wiki-link resolution for ambiguous page names
- Fixed heading ID generation for 10+ duplicate headings
- Parallel file copying in build for faster site generation
- Footer border now aligns with content width instead of full viewport
- Consistent 1rem spacing between all content blocks

## v1.0.0-beta.2
*January 27, 2026*

- Fixed hot reload not detecting static file changes
- Fixed browser caching during development (pages now refresh properly)
- Vercel deployments now target production environment by default
- Cross-platform path handling for Windows compatibility
- Verbose mode (`-v`) shows detailed rebuild and live reload info

## v1.0.0-beta.1
*January 26, 2026*

- All features enabled by default: graph, toc, search, wikilinks, backlinks
- `leafpress status` now tracks source files instead of build output
- Status command works without building first
- Hidden files (`.env`, `.gitignore`, etc.) excluded from tracking by default
- Improved Netlify deployment reliability with better error handling

## v1.0.0-beta
*January 26, 2026*

- Deployment manifest tracking: stores list of deployed files with hashes
- New `leafpress status` command: show pending changes since last deployment

## v1.0.0-alpha.4
*January 26, 2026*

- Deploy to Netlify with Personal Access Token authentication
- Smart file uploads: only changed files are uploaded in parallel to Netlify

## v1.0.0-alpha.3
*January 25, 2026*

- One-command deploy using `leafpress deploy` for multiple providers
- Deploy to GitHub Pages with browser-based OAuth
- Deploy to Vercel with browser-based authentication (same UX as GitHub)
- Auth codes automatically copied to clipboard on macOS, Linux, and Windows
- CI/CD support via `LEAFPRESS_<provider>_TOKEN` environment variable

## v1.0.0-alpha.2
*January 11, 2026*

- Adding `update` command to update leafpress
- Make blockquotes cleaner and modern

## v1.0.0-alpha.1
*January 8, 2026*

- Callouts restyled to be more modern
- Set height to full dynamic viewport

## v1.0.0-alpha
*January 8, 2026*

Initial release.

- Wiki-links with automatic backlinks
- Full-text search
- Graph visualization
- Table of contents
- Callouts (Obsidian-compatible)
- Tags with auto-generated tag pages
- Growth stages
- Dark mode
- Link previews on hover
- Google Fonts support
- Custom accent colors and gradient backgrounds
- Sitemap, RSS, robots.txt
- Open Graph and Twitter Card meta tags
- Custom 404 page
- Image lazy loading
- Live reload dev server
- Custom HTML into `<head>` for analytics, etc
