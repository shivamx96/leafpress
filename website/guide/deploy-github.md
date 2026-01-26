---
title: "Deploy to GitHub Pages"
date: 2025-12-21
---

The easiest way to deploy your leafpress site is with the built-in deploy command.

## Quick Start

```bash
leafpress deploy
```

That's it. The command handles authentication, building, and deployment in one step.

## First-Time Setup

On first run, `leafpress deploy` will:

1. Open your browser for GitHub authentication
2. Let you select a repository
3. Configure the deployment branch
4. Build your site
5. Push to GitHub Pages

Your credentials are saved locally for future deploys.

## Repository Types

GitHub Pages supports two types of sites:

**User/Organization Site** — Create a repo named `<username>.github.io` (e.g., `shivamx96.github.io`). Your site will be available at `https://<username>.github.io`. This is ideal for your main personal site or portfolio.

**Project Site** — Any other repo name (e.g., `my-garden`). Your site will be at `https://<username>.github.io/<repo-name>`. leafpress automatically handles the subdirectory URL path.

## Command Options

```bash
leafpress deploy [flags]

Flags:
  --dry-run        Validate without deploying
  --skip-build     Deploy without rebuilding
  --reconfigure    Re-run setup wizard
  --provider       Specify provider (default: from config)
```

## How It Works

When you run `leafpress deploy`, it will:

- Create the `gh-pages` branch if it doesn't exist
- Add `.nojekyll` to disable Jekyll processing
- Set the correct `baseURL` for subdirectory hosting (project sites only)
- Push your built site to GitHub

## CI/CD Usage

For automated deployments, set the `LEAFPRESS_GITHUB_TOKEN` environment variable instead of interactive OAuth:

```bash
export LEAFPRESS_GITHUB_TOKEN=ghp_your_token_here
leafpress deploy
```

The token needs `repo` scope for pushing to repositories.

### GitHub Actions Example

The easiest approach is to use GitHub's built-in `GITHUB_TOKEN`, which requires no setup:

```yaml
name: Deploy to GitHub Pages

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
          LEAFPRESS_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: leafpress deploy
```

Alternatively, if you want to use a Personal Access Token instead, create one with `repo` scope and add it as a repository secret named `GITHUB_TOKEN`.

## Check What's Pending

Before deploying, see what files have changed:

```bash
leafpress status
```

This shows which files are new, modified, or deleted since your last deployment. See [[guide/deployment-status|Deployment Status]] for details.

## Configuration

Deploy settings are stored in `leafpress.json`:

```json
{
  "deploy": {
    "provider": "github-pages",
    "settings": {
      "repo": "username/repo-name",
      "branch": "gh-pages"
    }
  }
}
```

To reconfigure deployment settings:

```bash
leafpress deploy --reconfigure
```

## Custom Domain

To use a custom domain with GitHub Pages:

1. Go to your repository **Settings > Pages > Custom domain**
2. Enter your domain and save
3. Configure DNS with your domain registrar (GitHub provides instructions)
4. Update `baseURL` in `leafpress.json`:

```json
{
  "baseURL": "https://yourdomain.com"
}
```

## Security Best Practices

**Token Management**:
- Never commit tokens to git or any version control
- Regenerate tokens if they're ever exposed
- Use GitHub secrets for CI/CD tokens in Actions workflows
- Consider rotating tokens periodically for added security

**Token Expiration & Rotation**:
- [GitHub Personal Access Tokens](https://github.com/settings/tokens) can be set to expire (recommended for security)
- When a token expires, generate a new one and update your CI/CD secrets
- Use the built-in `GITHUB_TOKEN` in GitHub Actions (it's ephemeral and handled by GitHub automatically)

**GitHub Secrets** (for CI/CD):
- Store `LEAFPRESS_GITHUB_TOKEN` in [repository secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets), never hardcode
- Use `${{ secrets.GITHUB_TOKEN }}` for GitHub Actions, which requires no setup
- Restrict secret access to workflows that need them

**Token Scope**:
- Use Personal Access Tokens with only `repo` scope (if using PAT instead of built-in token)
- The built-in `GITHUB_TOKEN` has appropriate scope automatically

## Troubleshooting

**Site not updating?** GitHub Pages can take a few minutes to update. Check the "Actions" tab in your repository for deployment status.

**404 errors on pages?** Make sure GitHub Pages is enabled in your repository settings and set to deploy from the `gh-pages` branch.

**CSS/links broken?** This usually means the `baseURL` isn't set correctly. Run `leafpress deploy --reconfigure` to fix it.
