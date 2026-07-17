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
    "baseUrl": "/hosted-fallback"
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
  "pages": []
}
```

- `config` is an optional canonical `leafpress.json` object. The renderer uses
  the same default overlay and validation as `config.Load` in the CLI.
- `garden.slug` remains required because hosted-garden identity is not a
  `leafpress.json` concern. Existing `garden.title`, `baseUrl`, `sort`, and
  `theme` inputs remain supported when `config` is absent.
- When `config` is present, its render-relevant fields take precedence. If its
  `baseURL` is empty, `garden.baseUrl` may still provide the hosted path prefix.
- `styleCSS` is the in-memory equivalent of a project's `style.css` and is
  appended after the embedded stylesheet using the same CLI composition.
- Operational fields (`outputDir`, `port`, `ignore`, and `deploy`) are accepted
  and validated for shape parity but do not cause I/O or deployment.

`headExtra` and `styleCSS` are trusted owner configuration, just as they are
in a local Leafpress project. Applications should not expose them to untrusted
garden readers or collaborators without an explicit trust decision.

## Output

The existing page, home, section, tag, CSS, and warning fields remain. The
`artifacts` array adds generated files using their exact CLI filenames:

- `graph.json` when graph is enabled
- `search-index.json` when search is enabled
- `feed.xml` when RSS is enabled
- `robots.txt`
- `sitemap.xml`
- `404.html`

Each artifact has `path`, `content`, and `contentType`, allowing a host to
store and serve files without knowing feature-specific output fields.

The renderer stays a pure transform: no filesystem, network, database, or
deployment access.
