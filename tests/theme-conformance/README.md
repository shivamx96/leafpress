# Theme conformance suite

This suite builds the representative garden in `testdata/theme-garden` with
the repository's local Core and CLI, then checks every supported theme against
the same browser-level contract.

The primary matrix covers:

- `classic`, `aurora`, `paper`, and `terminal`
- `base`, `sticky`, and `glassy` navigation
- `base`, `underlined`, and `box` active navigation treatments
- desktop and mobile viewports
- light and dark color schemes

Those dimensions produce 144 conformance states. Four additional smoke tests
exercise the fixture's article, index, tags, table, code, quote, footnotes,
callouts, backlinks, table of contents, theme toggle, search, and graph in every
theme. Footnote references, endnotes, and return links are also checked in all
144 matrix states.

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

`navigation.spec.js` adds 16 regression tests for long branding, overflow links,
heading fragments, changing header dimensions, viewport changes, and the mobile
Site menu disclosure, including section links in the compact floating state. Run them against WebKit with:

```sh
npx playwright install webkit
LEAFPRESS_CONFORMANCE_BROWSER=webkit npm run test:themes -- navigation.spec.js
```

WebKit on macOS uses Option-Tab in the keyboard test to include links in focus
navigation. Browser-engine checks do not replace a real-device Safari review.
