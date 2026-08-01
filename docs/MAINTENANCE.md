# Maintenance chores

Operational checklist for releases and vendored assets. Design decisions live
in the numbered docs (`05_RENDERER_CONTRACT.md`, `07_ASSET_ARCHITECTURE.md`);
this file is only *what to do* and *how to verify*.

## Every release

- [ ] Bump version / prepare release notes (`website/changelog.md`, tags).
- [ ] `go test ./...` from both module roots that ship (`core/`, `cli/`) as
      used in CI.
- [ ] Smoke a CLI build of `website/` (or a fixture garden): pages render,
      graph/search/previews work, no unexpected console/network calls.
- [ ] Confirm the cross-interface parity suite still passes
      (`cli/internal/build` parity tests).
- [ ] Grep for residual third-party browser origins that should not ship by
      default (`cdn.`, `fonts.googleapis`, `jsdelivr`, `unpkg`) and treat any
      hit as intentional + documented or as a bug.

## Built-in asset registry (`core/assets`)

Pins live in `registry_test.go` (`pinnedBuiltins`). `RegistryID()` is
content-derived — it changes automatically when any built-in is added,
removed, or altered. There is no manual registry version to bump.

### Fonts (curated families)

- [ ] When adding or updating a bundled family: place woff2 + OFL text under
      `core/assets/builtin/fonts/`, update `fonts.go` face/license tables,
      refresh SHA pins in `registry_test.go`.
- [ ] Verify `@font-face` URLs stay site-relative (`static/leafpress/fonts/…`)
      and that CLI materialization + renderer manifest stay in lockstep
      (`RequiredBuiltins` / `RequiredBuiltinsFor`).
- [ ] Smoke light + dark pages with the default theme.

### Favicons

- [ ] Replacing a root favicon: update `core/assets/builtin/favicon.*`,
      refresh pins, confirm `outputPath` remains the public root name
      (`favicon.ico`, etc.).
- [ ] Confirm user project-root overrides still win over built-ins.

### Mermaid (content-optional)

Vendored at `core/assets/builtin/mermaid/mermaid.min.js` (npm
[`mermaid`](https://www.npmjs.com/package/mermaid) version
`assets.MermaidVersion`, currently **11.4.1**). Logical paths:

- `static/leafpress/mermaid/mermaid.min.js`
- `static/leafpress/mermaid/LICENSE.txt`

Materialized into `_site/` / listed in the renderer asset manifest **only**
when rendered HTML contains a Mermaid diagram.

To bump Mermaid:

1. Download the browser UMD build matching the new version, e.g.  
   `https://cdn.jsdelivr.net/npm/mermaid@VERSION/dist/mermaid.min.js`  
   into `core/assets/builtin/mermaid/mermaid.min.js`.
2. Copy the package `LICENSE` to
   `core/assets/builtin/mermaid/LICENSE.txt`.
3. Set `assets.MermaidVersion` to the new version string.
4. Refresh SHA-256 pins for both paths in `registry_test.go`.
5. Run `go test ./core/assets/ ./core/render/ ./cli/internal/build/`.
6. Manually open a page with a ` ```mermaid ` block (light + dark) and
   confirm the script loads from `/static/leafpress/mermaid/mermaid.min.js`
   with **no** CDN request.

## Deprecations

- [ ] `theme.remoteFonts`: still an escape hatch; do not extend. Track
      removal once migrations are done.
- [ ] Scan config docs and changelog for other deprecated flags before a
      major cut.

## Docs hygiene (as needed)

- User guides under `website/guide/` must match runtime contracts (CSS
  composition, media bare paths, feature flags vs artifacts).
- Renderer contract (`docs/05_RENDERER_CONTRACT.md`) and asset ADR
  (`docs/07_ASSET_ARCHITECTURE.md`) must match CLI + `leafpress-render`
  behavior; prefer fixing code+docs in the same change when they drift.
