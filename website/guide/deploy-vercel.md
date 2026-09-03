---
title: "Deploy to Vercel"
date: 2025-12-21
---

Build your garden with leafpress and publish `_site/` with Vercel's official
CLI. Vercel owns authentication, project selection, uploads, readiness, and
production promotion.

## Local deployment

Install and authenticate the Vercel CLI:

```bash
npm install --global vercel
vercel login
```

Set `site.baseURL` in `leafpress.json` to the production Vercel or custom
domain. Then build and deploy the generated directory, replacing
`<project-name>` with the target Vercel project:

```bash
leafpress build
vercel deploy --prod --cwd _site --project <project-name>
```

If `build.outputDir` is not `_site`, pass that directory to `--cwd` instead.
The Vercel CLI waits for completion by default and writes the deployment URL to
standard output.

## GitHub Actions

Store a Vercel access token as `VERCEL_TOKEN`, the target project ID as
`VERCEL_PROJECT_ID`, and its account or team ID as `VERCEL_ORG_ID` in
repository secrets:

```yaml
name: Deploy to Vercel

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

      - name: Deploy to Vercel
        env:
          VERCEL_TOKEN: ${{ secrets.VERCEL_TOKEN }}
          VERCEL_PROJECT_ID: ${{ secrets.VERCEL_PROJECT_ID }}
          VERCEL_ORG_ID: ${{ secrets.VERCEL_ORG_ID }}
        run: npx --yes vercel deploy --prod --cwd _site --token "$VERCEL_TOKEN"
```

The Vercel deployment page and CLI exit status are the source of truth for
whether the production deployment is ready.

## Git integration

You can instead import the repository in Vercel and configure:

- Build command: `go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build`
- Output directory: `_site`

## Custom domain and security

Configure custom domains in Vercel, then update `site.baseURL`. Keep Vercel
tokens in its supported credential store or CI secrets, grant only the
necessary access, and revoke tokens that are no longer used.

See Vercel's [CLI deployment documentation](https://vercel.com/docs/cli/deploy)
for current options.
