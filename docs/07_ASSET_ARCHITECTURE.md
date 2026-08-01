# ADR 07 — Asset architecture

Status: accepted
Date: 2026-08-01

## Context

Leafpress is the canonical static site generator that other product lines build
on. It is consumed through two interfaces that must stay in lockstep:

- **The CLI** (`cli/`), which owns the filesystem: it scans a project directory,
  copies `static/`, and writes a complete `_site/` folder.
- **The embedded renderer** (`core/cmd/leafpress-render`, `core/render`), a pure
  stdin→stdout JSON transform used by hosted consumers. It performs no
  filesystem, network, database, or deployment access, and it never receives
  storage credentials.

Today assets are handled ad hoc:

- User files under `static/` are copied by the CLI (`copyStatic`) but have no
  representation in the renderer contract at all — a hosted garden that needs a
  static file has no shared vocabulary for it.
- Favicons are embedded in the CLI module (`cli/internal/assets`) and written to
  the site root; the renderer emits pages that link to favicons but nothing
  guarantees the hosted consumer serves one.
- Fonts are loaded from `fonts.googleapis.com` at browser time, so every
  exported and hosted site depends on a third-party origin.

Hosted consumers of the renderer store its output however suits them — a
relational database row keyed by path, an object store, a CDN, a plain
filesystem — and those choices evolve independently of Leafpress. **This ADR
must therefore not depend on any particular hosted storage backend.** The
contract below is storage-neutral: it describes assets with enough metadata
(stable identity, content hash, size, output location) that any of those
backends can materialize them. Content hashes make a content-addressed store
a natural fit, but nothing in the contract assumes one.

## Decision

### 1. `static/` is the canonical portable asset namespace

A **logical asset path** identifies an asset independently of any storage
backend. There is exactly one canonical representation:

- Relative, forward-slash separated, rooted at the site: `static/...`.
- Segments must be clean: no `..` or `.` segments, no empty segments, no
  leading `/`.
- Segment characters are restricted to the URL *unreserved* set —
  `A–Z a–z 0–9 - . _ ~` — so a path spells itself identically in a
  filesystem, a URL, and a quoted CSS `url("...")` context. This implicitly
  rejects `?`/`#`/`%` (URL syntax and percent-escapes), quotes and
  parentheses (CSS delimiters), backslashes, `:` (drive letters), spaces,
  and all control characters. Asset paths never need escaping; renderers
  escape segments anyway, defense in depth.
- Segments must be portable to every filesystem Leafpress ships on: Windows
  device names (`CON`, `NUL`, `COM1`, … — with or without an extension) and
  segments ending in a dot are rejected.
- Case-sensitive, unique within a site.

Validation of these rules lives in shared code (`core/assets`) used identically
by the CLI and the renderer. Logical paths become output file paths and
storage keys, so this is a security boundary, not a lint.

Scope of enforcement: every **declared** asset path — manifest entries,
`theme.fonts[].file`, registry built-ins — is validated with these rules on
both interfaces, always. The CLI's bulk `static/` copy is deliberately *not*
subject to them: it keeps copying arbitrary user filenames opaquely (the
reserved namespace being the one exception), so existing projects with
spaces or non-ASCII names in static files keep building. Such files simply
cannot participate in the portable asset contract until renamed.

Two sub-namespaces exist:

- **`static/leafpress/**` — reserved for Leafpress-owned built-in assets**
  (fonts, favicons, future icons). User projects must not place files there;
  the CLI treats a user file under `static/leafpress/` as a build error.
- **Everything else under `static/`** belongs to the user. `static/fonts/` is
  the conventional home for custom font files (§6).

### 2. The asset-manifest contract

A shared `core/assets` package defines the canonical types. An **asset** is:

| Field         | Meaning                                                        |
|---------------|----------------------------------------------------------------|
| `logicalPath` | Canonical `static/...` path (rules above). Identity.           |
| `contentType` | MIME type served with the asset (`font/woff2`, `image/svg+xml`).|
| `sha256`      | Lowercase-hex SHA-256 of the content.                          |
| `size`        | Content length in bytes.                                       |
| `outputPath`  | Optional **site-relative** output location (`favicon.ico`) when it differs from the logical path. Same path rules; never a URL. |

`outputPath` is deliberately *site-relative*, not an absolute browser URL,
because every consumer places assets under its own site root:

- the CLI writes `_site/<outputPath>`;
- templates render links as `{BasePath}/<outputPath>`;
- a host maps `<outputPath>` inside whatever route the garden lives under
  (`/g/name/<outputPath>`, a subdomain root, …).

When `outputPath` is empty it defaults to the logical path. Root-level
well-known files (`favicon.ico`, `favicon.svg`, `favicon-96x96.png`) keep
their historical locations via explicit `outputPath` values.

An **asset manifest** is a validated list of assets. Validation requires
strictly ascending order by logical path and rejects duplicate logical paths
*and* duplicate effective output paths — so any manifest that validates is
already in its one canonical, deterministic form. The package provides a
constructor that canonicalizes (sorts, then validates) so callers cannot
produce a "validated but unsorted" value. A manifest is *metadata only* — it
never carries bytes.

### 3. Materialization is the consumer's job

- **CLI**: reads bytes from disk (`static/`, plus project-root well-known
  files like a user `favicon.ico` overriding a built-in — §7) or from the
  embedded built-in registry, computes hashes, and writes files into
  `_site/` at each asset's effective output path.
- **Hosted**: resolves each manifest entry through its own storage — today a
  database row keyed by path, tomorrow perhaps a content-addressed object
  store; the `sha256` supports either without contract changes.
- **The renderer**: emits manifests, and emits built-in bytes only on explicit
  request (§5). It never reads or writes storage and never receives
  credentials. This is an invariant, not an optimization.

### 4. Built-in asset registry

Leafpress owns a small set of assets it ships inside the `core` module via
`go:embed`: default favicons, the curated font set, and content-optional
scripts such as Mermaid (`static/leafpress/mermaid/…`). The registry:

- assigns each built-in a **stable logical path** under `static/leafpress/`
  (e.g. `static/leafpress/fonts/inter-normal-latin.woff2`), a content type,
  and a hash computed from the embedded bytes and pinned by tests;
- exposes a **content-derived registry ID**: the SHA-256 of the JSON
  encoding of the canonical (sorted, validated) built-in manifest, exactly
  as `core/assets` serializes it — the same field names and omit-empty
  semantics as the wire manifest, pinned by tests. It changes exactly when
  any built-in is added, removed, or altered — there is no manually-bumped
  version to forget, and no room for two implementations to derive
  different IDs for the same assets;
- returns content only as defensive copies, so no caller can mutate the
  embedded bytes out from under the recorded hashes;
- is the single source of truth — the CLI's favicon fallback and the
  renderer's manifest both read from it.

### 5. The renderer asset flow

A pure renderer cannot discover user-uploaded files, so the flow is explicit
in the JSON contract:

**Input.** The caller may supply `assets`: a manifest of user assets it will
serve (custom font files, and in the future other referenced static files).
Entries are validated with the shared rules. Logical paths in the reserved
`static/leafpress/` namespace are always invalid — built-in content has one
source of truth, the registry. An explicit `outputPath` on a caller entry is
restricted to exactly a built-in's output path (the supported overrides: the
three root favicons); ordinary assets, custom fonts included, leave it empty
so they serve at their logical path. This keeps caller entries from
colliding with generated files (`style.css`, `index.html`, `404.html`,
feeds) and from diverging from the paths generated CSS references.
Duplicate logical or output paths among caller entries are an input error.

**Combination.** The output manifest is:

1. the **referenced built-ins** — favicons (linked from every page head),
   the bundled font faces **and OFL license texts** of configured families
   (license entries follow the same materialization and `emitAssets` rules
   as the faces), and **content-optional** built-ins only when used (Mermaid
   script + MIT license when any page contains a diagram); then
2. **caller entries merged over them**: a caller entry whose effective output
   path collides with a built-in's *replaces* that built-in entry (this is the
   favicon-override rule, identical in spirit to the CLI preferring a user
   `favicon.ico` on disk). Caller entries whose effective output path does
   not collide are appended — which, given the Input restriction above,
   means entries serving at their own logical path.

The output manifest therefore lists exactly what the rendered site requires —
not the whole registry, and not undeclared user files. If configuration
references a `static/fonts/` file that no caller entry declares, the renderer
emits a warning: the site will 404 that font unless the host serves it by
other means.

**Bytes.** Setting `emitAssets: true` adds base64 artifacts for the
**post-merge** output manifest entries that are still identical to a
registry built-in (same logical path *and* `sha256`), each at the entry's
**effective output path** — the same place the CLI writes the file and pages
reference it, so a consumer that stores artifacts by path serves them
correctly without joining against the manifest. An overridden built-in is
never emitted, and caller entries never produce byte artifacts — the
renderer does not have their bytes, only the caller does. Every artifact
carries an explicit `encoding`; **asset** artifacts are always `base64`
regardless of MIME type (OFL license texts included), while **generated**
site artifacts are always `utf8`. The `encoding` field is authoritative;
hosts never sniff.

**Synchronization is hash-driven, per entry.** The manifest lists only what
the current configuration references, so no single identifier can stand in
for "I have everything this render needs" — a theme change can require
built-ins an earlier render never mentioned. The rule consumers must follow,
per entry whose `sha256` is not already stored: for **registry-backed**
entries, re-render with `emitAssets` and take the base64 artifacts; for
**caller-declared** entries the renderer never has the bytes — the caller
must supply content matching the hash from its own storage, and `emitAssets`
cannot fill that gap. A consumer that already has every entry's hash skips
bytes entirely; routine renders stay metadata-only. `assetRegistryId` (§4) identifies the full
built-in snapshot for observability and prefetching, but it is **not** a
valid skip signal on its own.

### 6. Fonts

Font sources are a closed set; this is an explicit product decision:

1. **Bundled built-in family** (the default theme families — initially
   Crimson Pro, Inter, and JetBrains Mono — shipped in the registry as
   woff2): self-hosted via generated `@font-face` rules. The registry is
   the sole membership list; both interfaces must consult it, never a
   private copy.
2. **Declared custom family** (`theme.fonts` entries referencing
   `static/fonts/` files, with family, file, weight, style, display):
   self-hosted via generated `@font-face`. Files are validated as asset
   paths; the CLI additionally checks they exist.
3. **Any other family name** produces a **warning plus the CSS fallback
   stack** (Georgia/system-ui/monospace): the site stays self-contained and
   the author is told why the font is not loading. The old behavior — a
   remote `fonts.googleapis.com` link for the unbundled families — survives
   only behind an explicit, deprecated opt-in (`theme.remoteFonts: true`)
   for configurations that need time to migrate; it must not be extended
   and will be removed.

There are no arbitrary remote URLs in font configuration under any mode.

Mechanics and scope:

- `@font-face` rules are emitted into the generated stylesheet (`style.css` /
  the renderer's `css` output), not inlined into every page head — one
  cached copy instead of kilobytes per page. Font URLs inside the stylesheet
  are relative (`static/...`), which resolves correctly because the
  stylesheet itself is served from the site root under any base path or
  garden route.
- The curated bundled set is **Latin-focused**: latin and latin-ext subsets
  only. Text in other scripts (Cyrillic, Greek, Vietnamese, …) falls back to
  the system stack; authors who need full coverage for another script should
  declare a custom local font. Expanding the curated set is a deliberate
  future decision, not an accident of file size.
- The bundled families are SIL OFL 1.1. The full OFL license text (with each
  family's copyright notice) ships in the registry and is materialized into
  exported sites alongside the font files, satisfying the OFL's
  redistribution requirement.

### 7. Override semantics

A user or caller asset that lands on the same effective output path as a
built-in **replaces the built-in's manifest entry**, everywhere:

- CLI: a project-root `favicon.ico` is written instead of the bundled one, and
  any manifest the CLI describes must carry the user file's hash, not the
  registry's.
- Renderer: a caller `assets` entry whose **effective output path** is
  `favicon.ico` — that is, `logicalPath` under `static/...` with
  `outputPath: "favicon.ico"`, never a root or reserved-namespace logical
  path — replaces the built-in entry in the output manifest (§5).

Manifests always describe what is actually served; a consumer must never see
a built-in hash for a file the user overrode.

### 8. Security rules (summary)

- Asset paths: one canonical representation; absolute paths, traversal, empty
  segments, backslashes, colons, `?`/`#`/`%`, and Unicode control characters
  are rejected at validation time, in shared code, for logical and output
  paths alike.
- Renderer: no filesystem, no network, no credentials, ever. The only bytes it
  can emit come from its own embedded registry.
- Font sources: the closed set of §6. No arbitrary remote fetch targets.
- Hashes are lowercase-hex SHA-256 so consumers can verify content integrity
  end to end.

## Consequences

- The CLI and hosted consumers share one vocabulary for assets; conformance
  can be tested (a parity suite proves `_site/` and the renderer output
  describe the same set of output paths and contents).
- Exported sites become self-contained for every bundled or declared family;
  the deprecated compatibility mode is the one documented exception until it
  is retired.
- The `core` module grows by the embedded built-ins (a curated, bounded set of
  woff2 files) in exchange for removing the third-party origin.
- The artifact protocol change is additive; existing consumers that ignore
  `encoding` and the manifest fields keep working because every current
  artifact remains utf8 text.

## Implementation plan

This ADR is intentionally implementation-free. The work lands as:

1. this ADR;
2. canonical asset manifest types + validation (`core/assets`);
3. built-in asset registry (favicons first);
4. built-in self-hosted fonts;
5. custom local fonts;
6. CLI materialization + the renderer asset flow of §5;
7. config correctness (e.g. `Theme.Background` serialization) and the
   cross-interface parity suite.

Each implementation PR carries its own regression tests; the parity suite is a
contract net over the whole, not a substitute.

## Alternatives considered

- **Emit binary assets inline on every render** — simple, but wasteful: every
  hosted render would carry hundreds of KB of identical base64 fonts. Rejected
  in favor of manifest + opt-in emission with per-entry hash-driven sync.
- **Let the renderer fetch assets from storage itself** — collapses the purity
  boundary, requires credentials in the render path, and couples core to a
  storage vendor. Rejected; materialization stays with the consumer.
- **Absolute browser URLs in the manifest (`/favicon.ico`)** — breaks under
  base paths: the CLI serves `{BasePath}/favicon.ico` and hosts serve
  `/g/name/favicon.ico`, so an absolute URL is wrong for every consumer except
  root-hosted sites. Rejected in favor of site-relative `outputPath`.
- **A manually-bumped integer registry version** — drifts silently the first
  time someone updates an asset without touching it. Rejected in favor of a
  content-derived registry ID.
- **Allow remote URLs in custom font config** — reintroduces third-party
  origins, defeats export portability, and creates an SSRF-shaped
  configuration surface for hosted consumers. Rejected.
- **Keep favicons/fonts as CLI-only embeds** — status quo; leaves hosted sites
  without built-ins and duplicates asset knowledge per consumer. Rejected.
