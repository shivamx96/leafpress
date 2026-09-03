---
title: "Deploy to Netlify"
date: 2025-12-21
---

Build your garden with leafpress and publish `_site/` with Netlify's official
CLI. Netlify owns authentication, project linking, uploads, and deployment
status.

## Local deployment

Install and authenticate the Netlify CLI, then link the garden to a Netlify
project once:

```bash
npm install --global netlify-cli
netlify login
netlify link
```

Set `site.baseURL` in `leafpress.json` to the production Netlify or custom
domain, then build and publish:

```bash
leafpress build
netlify deploy --prod --dir=_site
```

If `build.outputDir` is not `_site`, pass that directory instead.

## GitHub Actions

Store a Netlify personal access token as `NETLIFY_AUTH_TOKEN` and the target
project ID as `NETLIFY_SITE_ID` in repository secrets:

```yaml
name: Deploy to Netlify

on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7

      - uses: actions/setup-go@v7
        with:
          go-version: "1.25.5"

      - name: Install leafpress
        run: go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest

      - name: Build site
        run: leafpress build

      - name: Deploy to Netlify
        env:
          NETLIFY_AUTH_TOKEN: ${{ secrets.NETLIFY_AUTH_TOKEN }}
          NETLIFY_SITE_ID: ${{ secrets.NETLIFY_SITE_ID }}
        run: npx --yes netlify-cli deploy --prod --dir=_site
```

Netlify's deploy log and CLI exit status are the source of truth for whether
the site is live.

## Git-based builds

You can instead connect the repository in Netlify and use:

- Build command: `go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build`
- Publish directory: `_site`

The same settings can be committed in `netlify.toml`:

```toml
[build]
  command = "go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build"
  publish = "_site"
```

## Custom domain and security

Configure custom domains in Netlify, then update `site.baseURL`. Keep Netlify
tokens in the provider's supported credential store or CI secrets, grant only
the necessary access, and revoke tokens that are no longer used.

See Netlify's [CLI documentation](https://docs.netlify.com/api-and-cli-guides/cli-guides/get-started-with-cli/)
for current authentication and deployment options.
