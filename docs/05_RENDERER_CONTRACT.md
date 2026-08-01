# leafpress-render contract

`leafpress-render` is the filesystem-free Leafpress build stack. It reads one
JSON object from stdin and writes one JSON object to stdout. The CLI and this
renderer share configuration parsing, templates, CSS, and generated-artifact
code; the transport is the only difference.

## Input

```json
{
  "garden": {
    "slug": "hosted-garden",
    "baseUrl": "/hosted-fallback",
    "showTagsInNav": true
  },
  "config": {
    "title": "My Garden",
    "baseURL": "https://example.com/notes",
    "nav": [{ "label": "Home", "path": "/" }],
    "theme": { "accent": "#50ac00" },
    "graph": true,
    "search": true,
    "toc": true,
    "backlinks": true,
    "wikilinks": true,
    "rss": true
  },
  "styleCSS": ".my-rule { color: rebeccapurple; }",
  "assets": [
    {
      "logicalPath": "static/fonts/my-serif.woff2",
      "contentType": "font/woff2",
      "sha256": "…64 lowercase hex…",
      "size": 18244
    },
    {
      "logicalPath": "static/my-favicon.ico",
      "contentType": "image/x-icon",
      "sha256": "…64 lowercase hex…",
      "size": 15086,
      "outputPath": "favicon.ico"
    }
  ],
  "emitAssets": false,
  "pages": []
}
```

- `config` is an optional canonical `leafpress.json` object. The renderer uses
  the same default overlay and validation as `config.Load` in the CLI.
- `garden.slug` remains required because hosted-garden identity is not a
  `leafpress.json` concern. Existing `garden.title`, `baseUrl`, `sort`, and
  `theme` inputs remain supported when `config` is absent.
- `garden.showTagsInNav` is a hosted fallback convenience when `config` is
  absent. It appends a native Tags item only when tagged pages generate a Tags
  index. Canonical `config.nav` remains authoritative; CLI projects add
  `/tags/` to that list explicitly.
- When `config` is present, its render-relevant fields take precedence. If its
  `baseURL` is empty, `garden.baseUrl` may still provide the hosted path prefix.
- `styleCSS` is the in-memory equivalent of a project's `style.css` and is
  appended after the embedded stylesheet using the same CLI composition.
- Operational fields (`outputDir`, `port`, `ignore`, and `deploy`) are accepted
  and validated for shape parity but do not cause I/O or deployment.

`headExtra` and `styleCSS` are trusted owner configuration, just as they are
in a local Leafpress project. Applications should not expose them to untrusted
garden readers or collaborators without an explicit trust decision.

- `assets` (optional) declares the user assets the caller will serve
  alongside the rendered site — custom font files under `static/fonts/`
  today. Each entry carries `logicalPath`, `contentType`, `sha256`, `size`,
  and optional site-relative `outputPath`, validated with the shared manifest
  rules; entries in the reserved `static/leafpress/` namespace are rejected.
  An explicit `outputPath` is restricted to exactly a built-in's output path
  (the supported overrides: the three root favicons) — ordinary assets,
  custom fonts included, leave it empty and serve at their logical path, so
  caller entries can never claim a generated file's location (`style.css`,
  `index.html`, `404.html`, feeds) or diverge from the paths generated CSS
  references. A caller favicon replaces the bundled manifest entry. Custom
  fonts referenced by the config but not declared here produce a warning:
  the site will 404 them unless the host serves them by other means.
- `emitAssets` (optional, default false) requests base64 artifacts carrying
  the bytes of the built-in manifest entries. Leave it off for routine
  renders: the asset manifest is always present, and synchronization is
  hash-driven (below).

## Output

The existing page, home, section, tag, CSS, and warning fields remain. The
`artifacts` array adds generated files using their exact CLI filenames:

- `graph.json` when graph is enabled
- `search-index.json` when search is enabled
- `feed.xml` when RSS is enabled
- `robots.txt`
- `sitemap.xml`
- `404.html`

Each artifact has `path`, `content`, `contentType`, and `encoding`; the
`encoding` field is present on every artifact and is authoritative — hosts
must never sniff. Generated site artifacts are always `utf8`. Asset
artifacts — one per built-in manifest entry, appearing only when
`emitAssets` is set — are always `base64`, regardless of MIME type (the OFL
license texts are base64 too).

`assetManifest` is the combined manifest of every asset the rendered site
requires: the referenced built-ins (favicons, the bundled font faces and OFL
license texts of configured families) plus every caller-declared asset, with
caller entries replacing built-ins on output-path collision. Each entry is
metadata only: logical path, content type, lowercase-hex SHA-256, size, and
the site-relative output path when it differs from the logical path. Hosts
materialize each entry through their own storage and serve it at its output
path inside the garden's route.

**Synchronization is hash-driven, per entry.** The manifest lists only what
the current configuration references, so no single identifier can stand in
for "I have everything this render needs" — a theme change can require
built-ins an earlier render never mentioned. For each entry: if content with
that `sha256` is not already stored, bytes are needed (re-render with
`emitAssets` for built-ins; caller assets are the caller's own files).
`assetRegistryId` is the content-derived identity of the full built-in set —
useful for observability and prefetching, but it is **not** a valid skip
signal on its own (see docs/07_ASSET_ARCHITECTURE.md).

When `emitAssets` is set, each *built-in* manifest entry also appears in
`artifacts` with `encoding: "base64"`, at its **effective output path** —
the exact filename a CLI export serves (`favicon.ico`, not its registry
logical path) — so a host that stores artifacts by `path` puts every file
where the site links expect it. Overridden built-ins and caller assets are
never emitted: the renderer does not have their bytes.

The renderer stays a pure transform: no filesystem, network, database, or
deployment access. Asset bytes it emits come exclusively from its own
embedded registry; it never receives storage credentials.
