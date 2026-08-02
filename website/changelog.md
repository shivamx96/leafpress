---
title: "Changelog"
date: 2025-01-06
toc: false
---

## Unreleased

- **Breaking: `leafpress.json` is now grouped into sections.** Settings live under `site` (title, author, baseURL, description, image, headExtra), `theme`, `features` (graph/search/toc/backlinks/wikilinks/rss), `navigation`, `build` (outputDir, port, ignore), and `deploy`. Every field is optional with a default, so an empty `{}` still builds a valid site. See the [Configuration guide](/guide/configuration/).
- **Navigation is now explicit about how it's built.** `navigation.mode` is `"automatic"` (default — derives the nav from your top-level notes and sections; the home is reached via the site title and is no longer duplicated as a link) or `"explicit"` (`navigation.items`). Automatic navigation now works in the CLI too, not just hosted renders. Set `navigation.includeTags: true` to add a Tags item automatically.
- **`baseURL` is a single, unambiguous canonical URL.** `sitemap.xml`, the RSS feed, and the `robots.txt` Sitemap line are generated only when `site.baseURL` is set; without it they're skipped with a warning instead of emitting invalid relative URLs.
- **Renderer contract v2.** `leafpress-render` takes one versioned envelope — `config` (the shared leafpress.json object), `render` (host-only `slug` and `footerAttribution`), `content` (pages, styleCSS, assets), and `options` — replacing the previous `garden`/`config` dual mode. Slugs and output paths now reject `.`/`..` segments. See `docs/05_RENDERER_CONTRACT.md`.

## v1.0.0-beta.13
*August 2, 2026*

- Hosted fallback renders now accept site `description` and `author` without requiring canonical config; the description supplies garden-home SEO metadata and the author supplies the existing copyright footer.
- Renderer consumers can replace “Grown with leafpress” using the structured, escaped `garden.footerAttribution` name and validated HTTP(S) URL. The CLI default is unchanged and raw footer HTML/scripts remain unsupported.

## v1.0.0-beta.12
*August 1, 2026*

- **Mermaid is self-hosted.** Diagram pages load a pinned `mermaid.min.js` (v11.4.1) from `static/leafpress/mermaid/` instead of jsDelivr. The script and its MIT license are only materialized when a page contains a diagram. See `docs/MAINTENANCE.md` for how to bump the vendored copy.
- **Link previews no longer depend on the search UI.** `search-index.json` is always emitted so hover previews keep working when `"search": false`; that flag only toggles the search overlay and ⌘K UI.
- Documentation aligned with runtime behavior: `style.css` is **appended** after defaults (not a full replace); bare video/audio embeds resolve to `static/video/` and `static/audio/`; writing guide lists the full callout type set and Obsidian aliases.
- Added `docs/MAINTENANCE.md` for release smoke checks and built-in asset chores (fonts, favicons, Mermaid).

## v1.0.0-beta.11
*August 1, 2026*

- **Fonts are self-hosted by default.** The default families (Crimson Pro, Inter, JetBrains Mono) now ship bundled with Leafpress and are served from your own site — a default build makes no request to Google Fonts. Other family names produce a build warning and fall back to system fonts; set the deprecated `theme.remoteFonts: true` to temporarily keep the old Google Fonts behavior while you migrate. The bundled set covers latin and latin-ext ranges, and each family's OFL license text ships alongside the font files.
- **Custom local fonts.** Declare your own font files under `static/fonts/` with `theme.fonts` (family, file, weight, style, display); Leafpress validates the declaration, verifies the file exists, and generates portable `@font-face` CSS. Declared families never load remotely.
- `@font-face` rules moved from every page head into the generated `style.css` (cached once, smaller pages).
- **Breaking:** `static/leafpress/` is now reserved for Leafpress built-in assets; builds fail with a clear error if user files are placed there. Move them anywhere else under `static/`.
- **Portable asset contract for hosted consumers.** `leafpress-render` output now carries `assetManifest` (logical path, MIME type, SHA-256, size, site-relative output path for every asset the site requires) and a content-derived `assetRegistryId`; callers declare their own assets — custom fonts, favicon overrides — through the new `assets` input, and can request built-in bytes as base64 artifacts with `emitAssets`. Every artifact now carries an explicit `encoding` field. Synchronization is hash-driven per manifest entry; see `docs/05_RENDERER_CONTRACT.md`.
- Declared asset paths (manifests, `theme.fonts`) are validated against one canonical, portable representation on both interfaces: no traversal, URL syntax, CSS-hostile characters, or Windows-reserved names. Bulk `static/` copying is unaffected.
- Fixed `theme.background` being silently dropped when Leafpress writes `leafpress.json`.
- Title-form wikilinks (`[[Note B]]` for slug `note-b`) now resolve in CLI builds — links, backlinks, and graph edges — matching hosted rendering; alias conflicts resolve deterministically by slug.
- A cross-interface parity suite now proves CLI exports and hosted rendering produce equivalent paths, artifacts, stylesheets, and assets.

## v1.0.0-beta.10
*July 31, 2026*

- Hosted `leafpress-render` consumers can opt the generated Tags index into fallback navigation with `garden.showTagsInNav`; the native `/tags/` item is added only when tag artifacts exist
- Canonical `config.nav` remains authoritative, so CLI projects continue to manage Tags navigation explicitly alongside their other navigation items

## v1.0.0-beta.9
*July 17, 2026*

- Fixed `go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest` for the multi-module repository by publishing scoped core/CLI module versions without local `replace` directives; Go-installed binaries now report their module version instead of `dev`
- `leafpress-render` now accepts the canonical `leafpress.json` configuration shape, custom `style.css` content, and produces graph, search, RSS, sitemap, robots, and 404 artifacts
- Folder sections, `index`/`_index` homes, root navigation, and all page-level frontmatter settings now share Leafpress CLI rendering semantics

## v1.0.0-beta.8
*July 16, 2026*

- New `leafpress-render` binary ships with every release: a stdin/stdout JSON bridge that renders a set of pages into full HTML documents, an index page, tag pages, and theme CSS — embed leafpress as the render engine inside any app, no filesystem or network access
- Rendering core extracted into a standalone Go module (`github.com/shivamx96/leafpress/core`) usable as a library; the CLI now builds on it
- Title-form wikilinks: `[[Page Title]]` resolves to the page's slug, not just filename-form links
- The render bridge escapes author-typed raw HTML (`<script>` renders as literal text) — CLI rendering behavior is unchanged
- Release artifacts now include a sha256 `checksums.txt` for verified downloads

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
