---
title: "Deploy to GitHub Pages"
date: 2025-12-21
---

Build your garden with leafpress and publish the generated `_site/` directory
with GitHub's official Pages actions.

## Configure the URL

For a user or organization site named `<owner>.github.io`, set:

```json
{
  "site": {
    "baseURL": "https://<owner>.github.io"
  }
}
```

For a project site, include the repository path:

```json
{
  "site": {
    "baseURL": "https://<owner>.github.io/<repository>"
  }
}
```

The project-site path must be present when leafpress builds so generated links
and assets resolve below `/<repository>/`.

## GitHub Actions workflow

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to GitHub Pages

on:
  push:
    branches: [main]
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - name: Check out repository
        uses: actions/checkout@v7

      - name: Set up Go
        uses: actions/setup-go@v7
        with:
          go-version: "1.25.5"

      - name: Install leafpress
        run: go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest

      - name: Build site
        run: leafpress build

      - name: Configure Pages
        uses: actions/configure-pages@v5

      - name: Upload site
        uses: actions/upload-pages-artifact@v4
        with:
          path: _site

      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

In the repository's **Settings → Pages** screen, select **GitHub Actions** as
the publishing source. The workflow uses GitHub's short-lived token, so no
personal access token or Leafpress credential file is needed.

If `build.outputDir` is not `_site`, update the workflow's artifact path.

## Custom domain

Configure the domain in **Settings → Pages**, update its DNS records, and set
`site.baseURL` to the custom production URL before the next build.

## Troubleshooting

- Inspect the workflow run and its deployment environment for the provider's
  authoritative status.
- Broken CSS or links on a project site usually mean `site.baseURL` is missing
  the repository path.
- A 404 for the whole site usually means Pages is not configured to use GitHub
  Actions or the workflow lacks `pages: write` and `id-token: write`.

See GitHub's [custom workflow documentation](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages)
for the current Pages requirements.
