---
title: "Changelog"
date: 2025-01-06
toc: false
---

## v1.0.0-beta.2
*January 27, 2025*

- Fixed hot reload not detecting static file changes
- Fixed browser caching during development (pages now refresh properly)
- Vercel deployments now target production environment by default
- Cross-platform path handling for Windows compatibility
- Verbose mode (`-v`) shows detailed rebuild and live reload info

## v1.0.0-beta.1
*January 26, 2025*

- All features enabled by default: graph, toc, search, wikilinks, backlinks
- `leafpress status` now tracks source files instead of build output
- Status command works without building first
- Hidden files (`.env`, `.gitignore`, etc.) excluded from tracking by default
- Improved Netlify deployment reliability with better error handling

## v1.0.0-beta
*January 26, 2025*

- Deployment manifest tracking: stores list of deployed files with hashes
- New `leafpress status` command: show pending changes since last deployment

## v1.0.0-alpha.4
*January 26, 2025*

- Deploy to Netlify with Personal Access Token authentication
- Smart file uploads: only changed files are uploaded in parallel to Netlify

## v1.0.0-alpha.3
*January 25, 2025*

- One-command deploy using `leafpress deploy` for multiple providers
- Deploy to GitHub Pages with browser-based OAuth
- Deploy to Vercel with browser-based authentication (same UX as GitHub)
- Auth codes automatically copied to clipboard on macOS, Linux, and Windows
- CI/CD support via `LEAFPRESS_<provider>_TOKEN` environment variable

## v1.0.0-alpha.2
*January 11, 2025*

- Adding `update` command to update leafpress
- Make blockquotes cleaner and modern

## v1.0.0-alpha.1
*January 8, 2025*

- Callouts restyled to be more modern
- Set height to full dynamic viewport

## v1.0.0-alpha
*January 8, 2025*

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
