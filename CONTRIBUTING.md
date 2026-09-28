# Contributing to leafpress

Thanks for your interest in leafpress. This document explains how the
repository is organised, how to run the checks that CI runs, and what a
pull request needs before it can merge.

leafpress is a small project maintained by one person. Please open an issue
before starting anything larger than a bug fix so the design can be agreed on
first. Documentation fixes and test improvements are always welcome without
prior discussion.

## Repository layout

leafpress is a Go workspace with two independently versioned modules and a
few supporting directories:

| Path | Purpose |
|------|---------|
| `core/` | `github.com/shivamx96/leafpress/core`. Markdown rendering, configuration, templates, themes, embedded assets, and the `leafpress-render` bridge. Published as a library. |
| `cli/` | `github.com/shivamx96/leafpress/cli`. The `leafpress` binary: `init`, `new`, `build`, `serve`, `update`, `version`. Depends on a released Core version. |
| `website/` | Source for [leafpress.in](https://leafpress.in), built with leafpress itself. Includes the changelog and the install script. |
| `docs/` | Design documents and maintainer runbooks. Start with `docs/01_PRD.md` and `docs/06_RELEASE_PROCESS.md`. |
| `tests/theme-conformance/` | Playwright browser suite covering every theme, navigation style, viewport, and colour scheme. |
| `benchmark/` | Reproducible build-time comparison against other static site generators. |
| `testdata/` | Fixture gardens used by Go tests and the browser suite. |

The root `go.work` wires `cli/` to the local `core/` for development. The
published CLI module must never depend on a local path, so `cli/go.mod` pins
a released or pseudo-versioned Core and contains no `replace` directive.

## Prerequisites

- Go at the version in `.go-version`. Automatic toolchain upgrades are
  disabled in CI, so match the pinned version locally.
- Node.js 24 and npm, only if you are running or changing the browser suite.
- Docker, only if you are running the comparative benchmark.

## Getting started

```bash
git clone git@github.com:shivamx96/leafpress.git
cd leafpress

# Build the CLI from the workspace
cd cli && go build -o leafpress ./cmd/leafpress && cd ..

# Try it against a fixture garden
cd testdata/theme-garden && ../../cli/leafpress serve
```

## Running the checks

CI runs everything below. Run the parts relevant to your change before
opening a pull request.

### Go formatting, vet, and tests

```bash
gofmt -l core cli            # must print nothing
(cd core && go vet ./... && go test -race ./...)
(cd cli  && go vet ./... && go test -race ./...)
```

### CLI as an external consumer

The workspace can hide a stale Core version in `cli/go.mod`. Always run the
CLI tests once with the workspace disabled:

```bash
cd cli && GOWORK=off go test -mod=readonly ./...
```

### End-to-end CLI checks

```bash
cd cli && ./test.sh
```

### Browser conformance suite

```bash
npm ci
npx playwright install chromium
npm run test:themes
```

See `tests/theme-conformance/README.md` for the WebKit and Firefox smoke runs
and the mobile navigation regression tests.

### Benchmark harness and website checks

```bash
LEAFPRESS_BIN="$PWD/cli/leafpress" ./benchmark/test.sh
bash tests/website-cache.sh
```

### Vulnerability scan

```bash
GOTOOLCHAIN=auto go install golang.org/x/vuln/cmd/govulncheck@latest
(cd core && govulncheck ./...)
(cd cli  && GOWORK=off govulncheck ./...)
```

## Making changes

### Cross-module changes

When a CLI change depends on a Core change, the two cannot land in one commit
because the CLI must consume Core through a fetchable module version.

1. Commit the Core change first.
2. In a following commit, pin `github.com/shivamx96/leafpress/core` in
   `cli/go.mod` to that commit's pseudo-version with `GOWORK=off go get`, and
   update `cli/go.sum`.
3. Run `GOWORK=off go test -mod=readonly ./...` from `cli/`.

Do not work around a standalone-module failure by adding a `replace` to
`cli/go.mod`, enabling the workspace in CI, or skipping tests. Those hide the
exact failure external consumers would hit. Full details are in `AGENTS.md`
and `docs/06_RELEASE_PROCESS.md`.

### Documentation parity

If a change affects user-visible behaviour, update the matching page under
`website/` in the same pull request. Configuration fields belong in
`website/guide/configuration.md` or `website/guide/theming.md`, frontmatter
fields in `website/guide/writing.md`, and every user-facing change gets a
line in `website/changelog.md` under the upcoming version.

If a change affects the renderer bridge, update `docs/05_RENDERER_CONTRACT.md`.

### Themes and templates

All bundled themes must pass the conformance suite. Adding a theme means
adding it to `core/themes/registry.go`, adding its stylesheet, and extending
the matrix in `tests/theme-conformance/themes.spec.js`. Template changes in
`core/templates/` should be checked in every theme, both colour schemes, and
both viewports.

### Performance

Performance changes must include a reproducible measurement in the pull
request: the command, fixture size, commit, and before/after numbers. Use
`leafpress build --cpuprofile` and `--memprofile` for local profiling. Do not
update the committed benchmark reports under `benchmark/results/` as part of
a feature change; those are refreshed at release time.

## Coding conventions

- Format with `gofmt`. Keep `go vet` clean.
- Wrap errors with `fmt.Errorf("...: %w", err)` and name the file or input
  involved so the user can act on the message.
- Prefer table-driven tests. Put fixtures under `testdata/` rather than in
  string literals when they are more than a few lines.
- Core is a public library. Exported identifiers need doc comments, and
  breaking changes to exported API or to the renderer contract must be called
  out in the pull request and the changelog.
- Core must not write to stdout or stderr or call `os.Exit`. User-facing
  output belongs in the CLI.
- Do not add dependencies to Core without discussion. The single binary and
  fast build are core product goals.

## Commit messages and pull requests

Commit subjects follow a light Conventional Commits style:

```
fix(cli): keep folder names in generated section titles
feat(core): configure list page columns
docs(website): normalize changelog wrapping
ci: block reachable Go vulnerabilities
```

Use `feat`, `fix`, `refactor`, `docs`, `test`, `ci`, or `chore`, with an
optional `core`, `cli`, `website`, or `release` scope. Write the body for the
reviewer: what changed, why, and how it was verified.

A pull request should:

- target `main` from a feature branch;
- describe the user-visible effect and link the issue it resolves;
- include tests for new behaviour and for the bug it fixes;
- update the website and changelog where applicable;
- pass every CI job. Do not skip or weaken a job to get green.

Squash merges are fine. Keep the pull request title usable as the squashed
commit subject.

## Releases

Releases are cut by the maintainer following `docs/06_RELEASE_PROCESS.md` and
the checklist in `docs/MAINTENANCE.md`. Contributors do not need to bump
versions or tags.

## Reporting security issues

Please do not open public issues for vulnerabilities. See `SECURITY.md`.

## License

By contributing you agree that your contributions are licensed under the MIT
License in `LICENSE`.
