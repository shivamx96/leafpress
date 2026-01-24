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

The token needs `repo` scope for pushing to repositories. In GitHub Actions, you can use the built-in `GITHUB_TOKEN` secret.

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

## Troubleshooting

**Site not updating?** GitHub Pages can take a few minutes to update. Check the "Actions" tab in your repository for deployment status.

**404 errors on pages?** Make sure GitHub Pages is enabled in your repository settings and set to deploy from the `gh-pages` branch.

**CSS/links broken?** This usually means the `baseURL` isn't set correctly. Run `leafpress deploy --reconfigure` to fix it.
