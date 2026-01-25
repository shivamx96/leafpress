---
title: "Changelog"
date: 2025-01-06
toc: false
---

## v1.1.0-alpha.1
*January 25, 2025*

- **Vercel deployment provider** - Deploy to Vercel with `leafpress deploy --provider vercel`
- **OAuth device flow for Vercel** - Browser-based authentication (same UX as GitHub)
- **Clipboard support for auth codes** - Auth codes automatically copied to clipboard on macOS, Linux, and Windows
- **Better error messages** - Improved error reporting for deployment failures
- Fixed Vercel username display after authentication and docs for navStyle/navActiveStyle options

## v1.1.0-alpha
*January 25, 2025*

- **One-command deploy**: `leafpress deploy` for GitHub Pages with browser-based OAuth
- GitHub Pages subdirectory hosting support (auto-sets baseURL)
- Secure token handling (tokens not exposed in process arguments)
- CI/CD support via `LEAFPRESS_GITHUB_TOKEN` environment variable

## v1.0.0-alpha.2
*January 11, 2025*

- Adding `update` command to CLI
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
