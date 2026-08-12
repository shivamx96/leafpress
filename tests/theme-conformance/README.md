# Theme conformance suite

This suite builds the representative garden in `testdata/theme-garden` with
the repository's local Core and CLI, then checks every supported theme against
the same browser-level contract.

The primary matrix covers:

- `classic` and `aurora`
- `base`, `sticky`, and `glassy` navigation
- `base`, `underlined`, and `box` active navigation treatments
- desktop and mobile viewports
- light and dark color schemes

Those dimensions produce 72 conformance states. Two additional smoke tests
exercise the fixture's article, index, tags, table, code, quote, callouts,
backlinks, table of contents, theme toggle, search, and graph in both themes.

Run the suite from the repository root:

```sh
npm ci
npx playwright install chromium
npm run test:themes
```

Generated gardens live in a temporary directory and are removed when the test
server stops. Playwright retains a trace and screenshot for failures and writes
its HTML report to `playwright-report/`.
