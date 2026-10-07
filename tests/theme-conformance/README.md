# Theme conformance suite

This suite builds the representative garden in `testdata/theme-garden` with
the repository's local Core and CLI, then checks every supported theme against
the same browser-level contract.

The checked-in garden keeps its Terminal preview settings. The CLI fixture test
discovers presets through `themes.Names()` and builds each in a separate temporary
copy, replacing only `theme.preset` before loading the config. Every preset runs
the same component and artifact assertions, including its registered stylesheet.
Generated `_site` output from manual previews is excluded from these copies.

For a local preview of another theme, build the CLI and copy the config outside
the fixture:

```sh
go build -o cli/leafpress ./cli/cmd/leafpress
cd testdata/theme-garden
cp leafpress.json /tmp/leafpress-preview.json
```

Edit `theme.preset` in that copy. Remove `navStyle` and `navActiveStyle` to try the
preset's navigation defaults, and optionally set `listColumns` to `1`. Then run
from the garden directory:

```sh
../../cli/leafpress serve --config /tmp/leafpress-preview.json
```

Keep the checked-in config unchanged when reviewing a theme. The browser matrix
sets its own preset, navigation, and column settings independently.

The primary matrix covers:

- `classic`, `aurora`, `paper`, `quiet`, and `terminal`
- `base`, `sticky`, and `glassy` navigation
- `base`, `underlined`, and `box` active navigation treatments
- desktop and mobile viewports
- light and dark color schemes

Those dimensions produce 180 conformance states. Five additional smoke tests
exercise the fixture's article, index, tags, table, code, quote, footnotes,
callouts, backlinks, table of contents, theme toggle, search, and graph in every
theme. Footnote references, endnotes, and return links are also checked in all
180 matrix states.

Run the suite from the repository root:

```sh
npm ci
npx playwright install chromium
npm run test:themes
```

Generated gardens live in a temporary directory and are removed when the test
server stops. Playwright retains a trace and screenshot for failures and writes
its HTML report to `playwright-report/`.

## Mobile navigation regression checks

`navigation.spec.js` adds 25 regression tests for long branding, overflow links,
heading fragments, changing header dimensions, viewport changes, and the mobile
Site menu disclosure, including section links in the compact floating state. Run them against WebKit with:

```sh
npx playwright install webkit
LEAFPRESS_CONFORMANCE_BROWSER=webkit npm run test:themes -- navigation.spec.js
```

WebKit on macOS uses Option-Tab in the keyboard test to include links in focus
navigation. Browser-engine checks do not replace a real-device Safari review.

## Secondary browser smoke checks

CI runs `browser-smoke.spec.js` on Firefox and WebKit (five tests per engine),
while Chromium retains the complete conformance matrix. To reproduce locally:

```sh
npx playwright install firefox webkit
LEAFPRESS_CONFORMANCE_BROWSER=firefox npm run test:themes -- browser-smoke.spec.js
LEAFPRESS_CONFORMANCE_BROWSER=webkit npm run test:themes -- browser-smoke.spec.js
```
