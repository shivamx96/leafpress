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
1. Open your browser for Vercel authorization
2. Guide you through project configuration
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

Generate a token at [Vercel Settings → Tokens](https://vercel.com/account/tokens) (select "Read/Write" access):

```bash
export LEAFPRESS_VERCEL_TOKEN=your_vercel_token_here
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
          go-version: '1.25'
      
      - name: Install leafpress
        run: go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest
      
      - name: Deploy
        env:
          LEAFPRESS_VERCEL_TOKEN: ${{ secrets.VERCEL_TOKEN }}
        run: leafpress deploy --provider vercel
```

Add your Vercel token as a repository secret named `VERCEL_TOKEN` in your repository settings.

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

## Security Best Practices

**Token Management**:
- Never commit tokens to git or any version control
- Regenerate tokens if they're ever exposed
- Use GitHub secrets for CI/CD tokens, never hardcode
- Consider rotating tokens periodically for added security

**Token Expiration & Rotation**:
- [Vercel Tokens](https://vercel.com/account/tokens) can be revoked anytime
- Regularly review and rotate tokens for security
- When revoking an old token, generate a new one and update your CI/CD secrets immediately

**GitHub Secrets**:
- Store `LEAFPRESS_VERCEL_TOKEN` in [repository secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets), never hardcode
- Restrict secret access to workflows that need them
- Never echo or log tokens in CI/CD output

**Token Scope**:
- Use the minimum necessary permissions for your token
- Vercel tokens should have deploy access only

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
