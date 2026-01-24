---
title: "Deploy to Vercel"
date: 2025-12-21
---

Deploy your leafpress site to Vercel with automatic SSL and edge network distribution.

## One-Command Deploy

The simplest way to deploy to Vercel:

```bash
leafpress deploy --provider vercel
```

First-time setup will:
1. Prompt you for a Vercel access token
2. Let you select or create a project
3. Save configuration for future deploys

After setup, subsequent deploys are just:

```bash
leafpress deploy
```

## Setup

Run the deploy command:

```bash
leafpress deploy --provider vercel
```

This opens your browser for Vercel authentication. After authorizing, the wizard guides you through project selection.

## CI/CD Usage

For automated deployments, use the `LEAFPRESS_VERCEL_TOKEN` environment variable:

```bash
export LEAFPRESS_VERCEL_TOKEN=your_token_here
leafpress deploy
```

### GitHub Actions Example

```yaml
name: Deploy to Vercel

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      
      - name: Install leafpress
        run: go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest
      
      - name: Deploy
        env:
          LEAFPRESS_VERCEL_TOKEN: ${{ secrets.VERCEL_TOKEN }}
        run: leafpress deploy --provider vercel
```

Add your Vercel token as a repository secret named `VERCEL_TOKEN`.

## Dry Run

Validate your setup without deploying:

```bash
leafpress deploy --dry-run
```

## Custom Domain

1. Deploy your site first
2. Go to your project in Vercel dashboard
3. Navigate to Settings > Domains
4. Add your custom domain
5. Update `baseURL` in `leafpress.json`:

```json
{
  "baseURL": "https://yourdomain.com"
}
```

## Reconfigure

To change projects or re-authenticate:

```bash
leafpress deploy --reconfigure
```

## Alternative: Git-Based Deploy

You can also deploy via Vercel's Git integration:

1. Push your site to GitHub/GitLab
2. Import the repository at [vercel.com](https://vercel.com)
3. Configure build settings:
   - **Build Command**: `go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build`
   - **Output Directory**: `_site`

Add a `vercel.json` for consistent settings:

```json
{
  "buildCommand": "go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build",
  "outputDirectory": "_site"
}
```
