# leafpress-render contract (v2)

`leafpress-render` is the filesystem-free Leafpress build stack. It reads one
JSON object from stdin and writes one JSON object to stdout. The CLI and this
renderer share one configuration schema, templates, CSS, and generated-artifact
code; the transport is the only difference.

The v2 contract has a single guiding principle: **there is one configuration
schema.** What the CLI reads from `leafpress.json` on disk is exactly what the
renderer receives in its `config` field — same fields, same defaults, same
validation. The renderer envelope adds only what a filesystem build gets for
free (the pages, the stylesheet, the declared assets, and the hosting
identity).

## Contract versioning

Both `leafpress.json` and the renderer input carry `contractVersion`. It is
optional and defaults to the latest version (`2`). An unknown version is
rejected rather than guessed at. v1 (the `garden`/`config` dual-mode envelope)
is no longer supported.

## Configuration schema

This is the shared object: `leafpress.json` on disk, and the renderer's
`config` field on the wire. **Every field is optional and has a default**, so
`{}` is a valid configuration that renders the default site. Defaults reproduce
the historical `config.Default()` values.

```json
{
  "contractVersion": 2,
  "site": {
    "title": "My Garden",
    "description": "Notes grown in public",
    "author": "Garden Author",
    "baseURL": "https://example.com/notes",
    "image": "/og-image.png",
    "headExtra": ""
  },
  "theme": {
    "preset": "classic",
    "fontHeading": "Bricolage Grotesque",
    "fontBody": "Inter",
    "fontMono": "JetBrains Mono",
    "accent": "#50ac00",
    "navStyle": "base",
    "navActiveStyle": "base",
    "listColumns": 2
  },
  "features": {
    "graph": true,
    "search": true,
    "toc": true,
    "backlinks": true,
    "wikilinks": true,
    "rss": true
  },
  "navigation": {
    "mode": "automatic",
    "includeTags": false
  },
  "build": {
    "outputDir": "_site",
    "port": 3000,
    "ignore": []
  },
  "deploy": {
    "provider": "",
    "settings": {}
  }
}
```

### Defaults

| Section | Field | Default |
| --- | --- | --- |
| `site` | `title` | `"My Garden"` |
| | `description`, `author`, `baseURL`, `image`, `headExtra` | `""` |
| `theme` | `preset` | `"classic"` |
| | `fontHeading` / `fontBody` / `fontMono` | `"Bricolage Grotesque"` / `"Inter"` / `"JetBrains Mono"` |
| | `accent` | `"#50ac00"` |
| | `navStyle` / `navActiveStyle` | `"base"` / `"base"` |
| | `listColumns` | `2` |
| | `background` | light/dark theme defaults |
| `features` | `graph` `search` `toc` `backlinks` `wikilinks` `rss` | `true` |
| `navigation` | `mode` | `"automatic"` |
| | `includeTags` | `false` |
| | `items` | `[]` |
| `build` | `outputDir` / `port` | `"_site"` / `3000` |
| | `ignore` | `[]` |
| `deploy` | `provider` / `settings` | `""` / `{}` |
| top | `contractVersion` | `2` |

### `site`

Site identity and SEO. `baseURL` is the site's canonical **absolute** URL
(scheme + host + optional path). The internal base path used for links is
derived from its path component — there is no separate path-prefix field, and
no second URL-shaped key with different meaning. `headExtra` is a trusted
owner escape hatch injected verbatim into `<head>` (see Trust boundary).

### `theme`

`preset` selects a bundled visual theme (`"classic"`, `"aurora"`, `"paper"`, or `"terminal"`) and
defaults to `"classic"`. Unknown names are rejected with the available preset
list. The selected preset supplies
the initial fonts, accent, backgrounds, and navigation defaults; explicitly
provided config fields override them. Custom `fonts[]` declarations and all
values interpolated into the inline `<style>` block retain their existing
validation.

`listColumns` accepts `1`, `2`, or `3` and controls list-page columns on
desktop. List pages always collapse to one column at the mobile breakpoint.

The shared base stylesheet owns the semantic type scale for page titles,
article headings, section headings, body copy, and component text. Presets may
change typefaces, weights, tracking, color, and composition, but should not
change those sizes or invert the hierarchy between heading levels.

### `features`

The six reader-feature toggles, grouped. Defaults are all `true` in both the
CLI and the renderer — introducing a config never flips a feature on or off
relative to omitting it. `search-index.json` is emitted regardless of the
`search` toggle (it also powers hover link previews); `search` only controls
the search UI.

### `navigation`

```json
{ "mode": "automatic", "includeTags": false }
// or
{ "mode": "explicit", "items": [{ "label": "Home", "path": "/" }] }
```

- `mode` (default `automatic`) chooses how the nav bar is built:
  - `automatic` derives nav from the garden's public top level (root notes and
    section homes). `includeTags` appends a Tags item when tagged pages produce
    a Tags index.
  - `explicit` uses `items` verbatim; `includeTags` is ignored (add `/tags/`
    yourself).
- The mode is chosen by the `mode` value alone. The presence of any other
  field never changes the mode.
- Automatic navigation is generated by shared core code, so the CLI and
  renderer produce identical navigation from identical config.

### `build` / `deploy`

Operational fields. The CLI honors them; the renderer accepts and validates
them for shape parity but performs no I/O or deployment.

## Renderer input

The full object read from stdin:

```json
{
  "contractVersion": 2,
  "config": { "…the configuration schema above…" },
  "render": {
    "slug": "hosted-garden",
    "footerAttribution": { "name": "Acme", "url": "https://acme.example" }
  },
  "content": {
    "pages": [],
    "styleCSS": "",
    "assets": []
  },
  "options": {
    "emitAssets": false
  }
}
```

- `config` is the shared configuration schema. Omit it entirely for the default
  site.
- `render` holds the only genuinely host-only concerns:
  - `slug` is the hosted garden's identity/routing key. It has no natural
    default; if omitted it defaults to `"garden"` and the renderer emits a
    warning, so a render never fails on it but the omission is visible. Hosts
    should always supply it.
  - `footerAttribution` is renderer-only white-label branding. Its required
    `name` renders after "Grown with"; its optional `url` must be an absolute
    HTTP(S) URL. Omitting the block yields the default "Grown with leafpress"
    footer. It is structured, never raw HTML or script.
- `content` carries the in-memory equivalents of the filesystem inputs a CLI
  build reads:
  - `pages` (default `[]`) — the published pages (schema below).
  - `styleCSS` (default `""`) — the in-memory `style.css`, appended after the
    shared base, selected bundled theme, and self-hosted fonts using the same
    CLI composition. Trusted owner configuration.
  - `assets` (default `[]`) — declares the user assets the caller will serve
    alongside the site (custom fonts under `static/fonts/` today). Each entry
    carries `logicalPath`, `contentType`, `sha256`, `size`, and an optional
    site-relative `outputPath`, validated with the shared manifest rules.
    Entries in the reserved `static/leafpress/` namespace are rejected. An
    explicit `outputPath` is restricted to exactly a built-in's output path
    (the three root favicons); ordinary assets, custom fonts included, leave it
    empty and serve at their logical path, so a caller entry can never claim a
    generated file's location (`style.css`, `index.html`, `404.html`, feeds).
    A caller favicon replaces the bundled manifest entry. Custom fonts
    referenced by the config but not declared here produce a warning: the site
    will 404 them unless the host serves them another way.
- `options.emitAssets` (default `false`) requests base64 artifacts carrying the
  bytes of the built-in manifest entries. Leave it off for routine renders: the
  asset manifest is always present, and synchronization is hash-driven (below).

### Page schema

Each entry in `content.pages`:

```json
{
  "slug": "essays/my-post",
  "title": "My Post",
  "markdown": "…",
  "tags": ["notes"],
  "createdAt": "2026-01-01T00:00:00Z",
  "updatedAt": "2026-01-02T00:00:00Z",
  "description": "",
  "growth": "seedling",
  "image": "",
  "toc": null,
  "readingTime": null,
  "isIndex": false,
  "sort": "date",
  "showList": null
}
```

- `slug` is required for ordinary pages and may carry path segments; section
  membership derives from the slug's directory, exactly like the CLI build.
  Slugs and derived output paths reject unsafe characters and `.`/`..` path
  segments.
- `isIndex` marks a section home (the CLI's `_index.md`): `slug` is the section
  path, `markdown` becomes the intro above the child listing. An `isIndex` page
  with slug `""` is the garden home. `sort` (`date` | `title` | `growth`) and
  `showList` mirror the `_index.md` frontmatter keys.
- `title` defaults to the slug; `growth` must be one of
  `seedling` | `budding` | `evergreen`; `readingTime`, when set, overrides the
  computed value and must be positive; timestamps are RFC3339.
- `tags` supplies explicit page metadata. Inline tags found in `markdown`
  (for example, `#notes`) are merged into that list case-insensitively; the
  explicit tag's order and spelling win when both forms name the same tag.

## Invariants

- **One schema.** `leafpress.json` and the renderer's `config` are the same
  object with the same defaults and the same validation. No dual mode, no
  fallback branch.
- **Everything has a default.** Any config field may be omitted; omission
  yields the documented default and never changes behavior relative to another
  field being present. `render.slug` is the sole exception (host identity), and
  even it defaults with a warning rather than failing.
- **One URL, one meaning.** `site.baseURL` is the canonical absolute URL; the
  internal base path is derived from it. There is no second URL field.
- **Explicit navigation mode.** Nav is `automatic` or `explicit` by the `mode`
  value alone.
- **No silent loss.** Conflicting or unknown fields are rejected, not ignored.
- **Safe paths.** Every slug and output path is rejected if it contains unsafe
  characters or dot segments.
- **Known origin for absolute URLs.** When `site.baseURL` is empty, artifacts
  that require an absolute origin (sitemap `<loc>`, the robots `Sitemap:`
  directive, RSS links) are skipped with a warning rather than emitted with
  invalid relative values.

## Output

The top-level output object:

```json
{
  "pages": [{ "slug": "…", "html": "…" }],
  "index": "…",
  "sections": [{ "slug": "…", "html": "…" }],
  "tags": { "index": "…", "pages": [{ "tag": "…", "html": "…" }] },
  "css": "…",
  "assetManifest": [ "…" ],
  "assetRegistryId": "…",
  "artifacts": [ "…" ],
  "warnings": [ "…" ]
}
```

### Artifacts

Generated site files, each using the exact CLI filename:

- `static/leafpress/app.<sha256-prefix>.js` always; every rendered HTML
  document references this one site-wide client bundle from its `<head>`. The
  32-hex-character prefix is the first 128 bits of the bundle's SHA-256 and
  changes when feature or base-path configuration changes.
- `graph.json` when the graph feature is enabled
- `search-index.json` always (full-text search and hover link previews share
  it; the `search` feature only toggles the search UI)
- `feed.xml` when RSS is enabled
- `robots.txt`
- `sitemap.xml`
- `404.html`

Each artifact has `path`, `content`, `contentType`, and `encoding`. The
`encoding` field is present on every artifact and is authoritative — hosts must
never sniff. Generated site artifacts are always `utf8`. Asset artifacts — one
per built-in manifest entry, appearing only when `options.emitAssets` is set —
are always `base64`, regardless of MIME type (the OFL license texts are base64
too).

The generated client bundle is an artifact, not an `assetManifest` entry. Its
bytes depend on the current render configuration and are returned on every
render so a host can publish the exact content-addressed path referenced by the
HTML. Hosts should replace the publication snapshot atomically so superseded
bundle hashes do not accumulate.

### Asset manifest

`assetManifest` is the combined manifest of every asset the rendered site
requires: the referenced built-ins (favicons, the bundled font faces and OFL
license texts of configured families) plus every caller-declared asset, with
caller entries replacing built-ins on output-path collision. Each entry is
metadata only: logical path, content type, lowercase-hex SHA-256, size, and the
site-relative output path when it differs from the logical path. Hosts
materialize each entry through their own storage and serve it at its output
path inside the garden's route.

**Synchronization is hash-driven, per entry.** The manifest lists only what the
current configuration references, so no single identifier can stand in for "I
have everything this render needs" — a theme change can require built-ins an
earlier render never mentioned. For each entry: if content with that `sha256`
is not already stored, bytes are needed (re-render with `options.emitAssets`
for built-ins; caller assets are the caller's own files). `assetRegistryId` is
the content-derived identity of the full built-in set — useful for
observability and prefetching, but it is **not** a valid skip signal on its own
(see docs/07_ASSET_ARCHITECTURE.md).

When `options.emitAssets` is set, each *built-in* manifest entry also appears in
`artifacts` with `encoding: "base64"`, at its **effective output path** — the
exact filename a CLI export serves (`favicon.ico`, not its registry logical
path) — so a host that stores artifacts by `path` puts every file where the
site links expect it. Overridden built-ins and caller assets are never emitted:
the renderer does not have their bytes.

## Trust boundary

`site.headExtra` and `content.styleCSS` are trusted owner configuration, just
as they are in a local Leafpress project, and are emitted without escaping.
Applications should not expose them to untrusted garden readers or
collaborators without an explicit trust decision. Everything else that reaches
template output — titles, descriptions, author, nav labels, theme values, tags,
and footer attribution — is escaped or validated at the input boundary. Raw
HTML in author markdown renders as visibly escaped text, because this bridge
serves multi-tenant user content to third parties.

## Purity

The renderer is a pure transform: no filesystem, network, database, or
deployment access. Asset bytes it emits come exclusively from its own embedded
registry; it never receives storage credentials.

Exit codes: `0` success (warnings allowed), `1` invalid input, `2` internal
render failure. Only JSON is ever written to stdout.
