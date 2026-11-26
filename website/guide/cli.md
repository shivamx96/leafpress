---
title: "CLI Reference"
date: 2026-11-26
---

Run commands from your garden directory. Use `leafpress --help` or
`leafpress <command> --help` for help for your installed version.

## Commands

| Command | Behavior |
|---------|----------|
| `leafpress init` | Creates configuration, a commented `style.css`, static media directories, and a `.gitignore` entry for `_site/`. Adds a sample `index.md` if no Markdown content exists. Refuses to overwrite an existing configuration. |
| `leafpress new "My Note"` | Creates `my-note.md` with today's date, `draft: true`, and `growth: "seedling"`. Accepts nested names such as `notes/my-note`; refuses to overwrite existing files. |
| `leafpress build` | Generates the site in `build.outputDir` (default `_site/`). |
| `leafpress serve` | Builds and watches the garden with live reload at `http://127.0.0.1:3000` by default. |
| `leafpress update` | Downloads and verifies the latest release, then replaces the installed binary. |
| `leafpress version` | Prints the installed version. `leafpress --version` also works. |

## Global flags

| Flag | Behavior |
|------|----------|
| `--config <file>`, `-c <file>` | Selects the configuration file. Defaults to `./leafpress.json`; `init` writes to the selected path. An explicitly selected missing file is an error for `build` and `serve`. |
| `--verbose`, `-v` | Prints detailed output, including individual build warnings. |
| `--help`, `-h` | Shows command help. |

## Build and serve flags

| Flag | Commands | Behavior |
|------|----------|----------|
| `--drafts`, `-d` | `build`, `serve` | Includes draft notes. Does not bypass `publish: false`, explicit publishing permission, hidden paths, or ignore rules. |
| `--offline` | `build`, `serve` | Disables font downloads. `LEAFPRESS_OFFLINE=1` also enables this mode. Bundled and declared local fonts still work. |
| `--strict` | `build` | Prints warning details and fails on any warning before replacing existing output. Never downloads fonts. |
| `--cpuprofile <file>` | `build` | Writes a Go CPU profile to the specified file. |
| `--memprofile <file>` | `build` | Writes a Go heap profile to the specified file. |
| `--port <number>`, `-p <number>` | `serve` | Overrides `build.port` (default `3000`). |
| `--host <address>` | `serve` | Binds to this interface (default `127.0.0.1`). Use `0.0.0.0` to expose the preview on your network; the development server has no authentication. |

For example, preview new drafts on a different port:

```bash
leafpress serve --drafts --port 4000
```

Before publishing, remove `draft: true` from ready notes and run:

```bash
leafpress build --strict
```

See [[guide/configuration|Configuration]] for publishing policies and font
handling in strict builds.

## Update flags

`leafpress update --force` (or `-f`) reinstalls even when the installed version
is already the latest release.
