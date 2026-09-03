---
title: "Changelog"
date: 2025-01-06
toc: false
---

## v1.0.0-beta.20
*September 3, 2026*

- **Breaking: removed `leafpress deploy` and its built-in provider authentication.**
  Build with leafpress, then publish `_site/` with the hosting provider's
  supported CLI or CI integration. Existing `deploy` configuration remains
  accepted but is ignored so gardens can migrate without a config parsing
  failure.
- **Breaking: removed `leafpress status` and its local deployment manifest.**
  Provider-native dashboards and CLIs are now the source of truth for
  deployment state.
- Added `theme.listColumns` to select one, two, or three columns for section and
  tag-page lists on desktop; mobile layouts remain single-column.
- Fixed Aurora and Paper page titles wrapping earlier than their containers
  required.

## v1.0.0-beta.19
*August 20, 2026*

- **`leafpress serve` now binds 127.0.0.1.** The dev server has no authentication and can publish drafts, so it no longer listens on every interface. Preview from another device with `leafpress serve --host 0.0.0.0`, which says so explicitly; the startup URL follows the bind address.
- **`build.ignore` honors the documented glob patterns.** `drafts/**`, `*.draft.md`, `notes/*.wip.md`, and `a/**/tmp.md` now match — previously every documented example silently matched nothing — and a malformed pattern is a config error instead of a no-op. Note the widening: a bare name such as `drafts` hides that folder at any depth, per the gitignore convention. `docs/` is no longer reserved, so a garden can publish its own documentation; hide it with `build.ignore` if you would rather not.
- **Content that points outside the garden is refused.** Symlinks under content or `static/` that resolve outside the project are rejected on scan, single-file parse, and static copy; `leafpress new` writes through `os.Root` so existing links cannot redirect it; and tag names that would escape the generated tags directory are rejected.
- **Site and page metadata is escaped in CLI builds.** A double quote in a description closed the `content="..."` attribute early and truncated `og:description`. The CLI and the hosted renderer now share one set of escaping helpers, and wikilinks resolve from both the raw and escaped spelling of a title, so `[[Q&A]]` still works. Goldmark 1.8.5 also strips entity-encoded `javascript:` autolinks.
- **Wikilinks are parsed as Markdown syntax nodes.** Links inside code spans and blocks, escapes, raw HTML attributes and comments, Obsidian embeds, and ordinary Markdown link labels no longer create backlinks or graph edges, and labels and hrefs are escaped on output.
- **The graph opens without monopolizing the page.** Layout runs in 10 ms animation-frame slices, repulsion uses a spatial hash instead of all-pairs comparison, adjacency and tag data are pre-indexed, labels are built once, and `prefers-reduced-motion` does less work.
- Mermaid is vendored at 11.16.1 with refreshed SHA-256 pins, and the hardening keys are locked so a diagram's own `%%{init: ...}%%` directive cannot re-enable `securityLevel: loose` or HTML labels. With HTML labels off, `$$...$$` math stays literal text.
- `--config`/`-c` is honored by `deploy`, `status`, and `init`, which previously read or wrote `leafpress.json` regardless of the flag.
- The serve watcher shares one exclusion predicate with the content scan, so reserved names and `build.ignore` can no longer disagree about what is content, and incremental rebuilds refresh navigation, wikilink aliases, backlinks, tag listings, RSS, sitemap, and graph/search data.
- Conflicting output routes are rejected before a build instead of quietly overwriting, and ordinary files such as `migration_index.md` keep their own route.
- Frontmatter values and Markdown body lines longer than 64 KiB are preserved instead of being truncated.
- Documentation corrections: the GitHub Actions deploy example needs `permissions: contents: write` and a non-reserved secret name; the theming link scaffolded into `style.css` points at `/guide/theming/`; the deploy providers are exactly GitHub Pages, Vercel, and Netlify, with `leafpress build` as the escape hatch for any other host; and the Obsidian plugin is marked as not yet available.

## v1.0.0-beta.18
*August 13, 2026*

- **Added four bundled visual themes.** Select `classic`, `aurora`, `paper`, or `terminal` with `theme.preset`. Classic remains the default for existing gardens, while explicit font, accent, background, and navigation settings continue to override preset defaults.
- **Added automatic inline tags.** Tags written as `#notes` in ordinary Markdown now become links and feed tag pages, navigation, graph/search metadata, and incremental rebuilds alongside frontmatter tags. Code, links, URLs, raw HTML tags, escaped hashes, and unsupported nested tags remain literal.
- Color schemes now apply before styles load and stay synchronized with saved or operating-system preferences.
- Added a representative theme garden and a 148-case browser conformance suite covering every preset, navigation and active-link style, desktop and mobile layouts, light and dark appearances, and reader tools including search and graph.

## v1.0.0-beta.17
*August 11, 2026*

- **Reduced repeated JavaScript in generated sites.** Leafpress now emits one content-addressed client bundle shared by every page instead of embedding the same client code into every HTML document. Bundle paths change with relevant configuration, stale hashes disappear on full rebuilds, and native CLI and hosted-renderer output remain in parity.
- **Preloaded selected self-hosted fonts.** Pages preload one normal face for each configured heading, body, and monospace family in role order, while deduplicating reused families and leaving italic and extended-Latin faces demand-loaded.
- **Made full builds transactional and output cleanup safer.** Builds render into staging and replace the published site only after every artifact succeeds; failures preserve the previous output. Output directories must remain inside the project, symlink escapes are blocked, and non-empty custom directories require Leafpress ownership before cleanup.
- Section index introductions now use the same rich-content styling as ordinary pages.
- Rebuilt the SSG benchmark around a deterministic notes/posts/tags workload, normalized routes and support pages across generators, rotated run order, logical output sizes, strict output contracts, pinned Docker inputs, and independently visible standalone CLI validation.

## v1.0.0-beta.16
*August 3, 2026*

- **Hardened the multi-tenant renderer against stored XSS.** TOC headings are escaped, search results are built without unsafe `innerHTML`, theme backgrounds are validated end to end, renderer failures stay inside the escaping boundary, and sitemap/RSS URLs are XML-escaped.
- **Secured `leafpress update`.** Release downloads are verified against `checksums.txt`, replacement uses a same-directory atomic rename, and accidental downgrades are rejected.
- Fixed graph edges disappearing when backlinks were disabled, deduplicated repeated wikilinks, normalized case-variant tags, made RSS ordering deterministic, and made title/description truncation Unicode-safe.
- Reworked incremental serving so rebuilds are serialized, bulk file changes are retained, new directories are watched, section listings refresh, removed static files disappear from `_site`, and authored section indexes are preserved.
- Hardened deploy credentials and output: credential files use private permissions, GitHub commits supply an explicit CI identity, Netlify tokens are no longer echoed, and deploy status/change detection now has direct coverage.
- Added integration coverage for the renderer bridge, scanner, navigation, server routing, deploy flows, and incremental parity; CI now enforces formatting, vet, coverage, and the full CLI suite with Go 1.25.5.
- The default heading font is now self-hosted Bricolage Grotesque, and the built-in catalog now includes 17 curated proportional and monospace families with their OFL license texts.

## v1.0.0-beta.15
*August 2, 2026*

- **Unknown or misplaced config fields are now rejected.** A `leafpress.json` typo (or a wrongly-nested key), and any unknown field in the renderer input, now fail with a clear error instead of being silently ignored — closing a gap where a v1-shaped payload rendered an empty site.
- Fixed `leafpress deploy` writing `baseURL` at the top level for GitHub Pages; it now writes `site.baseURL`, so project-site (subdirectory) hosting gets the correct base path.
- Documentation corrected for the v2 config schema: `leafpress init` takes no directory argument and scaffolds at the project root (there is no `content/` directory); sitemap.xml and RSS require `site.baseURL`; deploy and status guides updated.

## v1.0.0-beta.14
*August 2, 2026*

- **Breaking: `leafpress.json` is now grouped into sections.** Settings live under `site` (title, author, baseURL, description, image, headExtra), `theme`, `features` (graph/search/toc/backlinks/wikilinks/rss), `navigation`, `build` (outputDir, port, ignore), and `deploy`. Every field is optional with a default, so an empty `{}` still builds a valid site. See the [Configuration guide](/guide/configuration/).
- **Navigation is now explicit about how it's built.** `navigation.mode` is `"automatic"` (default — derives the nav from your top-level notes and sections; the home is reached via the site title and is no longer duplicated as a link) or `"explicit"` (`navigation.items`). Automatic navigation now works in the CLI too, not just hosted renders. Set `navigation.includeTags: true` to add a Tags item automatically.
- **`baseURL` is a single, unambiguous canonical URL.** `sitemap.xml`, the RSS feed, and the `robots.txt` Sitemap line are generated only when `site.baseURL` is set; without it they're skipped with a warning instead of emitting invalid relative URLs.
- **Renderer contract v2.** `leafpress-render` takes one versioned envelope — `config` (the shared leafpress.json object), `render` (host-only `slug` and `footerAttribution`), `content` (pages, styleCSS, assets), and `options` — replacing the previous `garden`/`config` dual mode. Slugs and output paths now reject `.`/`..` segments. See `docs/05_RENDERER_CONTRACT.md`.
- Fixed wiki-links inside inline code leaking into processing on pages that use 4-backtick or nested code fences (e.g. docs that show code-fence examples); code protection now follows CommonMark fence rules.

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
