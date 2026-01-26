---
title: "Deploy to Netlify"
date: 2025-12-21
---

Deploy your leafpress site to Netlify with automatic SSL and CDN distribution.

## Quick Start

```bash
leafpress deploy --provider netlify
```

First-time setup will:
1. Prompt you for a Personal Access Token
2. Guide you through site selection or creation
3. Save configuration for future deploys

After setup, subsequent deploys are just:

```bash
leafpress deploy
```

## Setup

### 1. Generate a Personal Access Token

1. Go to [Netlify Settings → Applications](https://app.netlify.com/user/applications)
2. Click "New access token"
3. Give it a name like "leafpress CLI" and create it
4. Copy the token (it won't be shown again!)

**Note**: Keep the token safe. If it's exposed, you can revoke it from Netlify Settings and generate a new one.

### 2. Run the Deploy Wizard

```bash
leafpress deploy --provider netlify
```

When prompted, paste your Personal Access Token. The wizard will:
- Authenticate with your Netlify account
- Show your existing sites
- Let you create a new site if needed
- Save the site configuration

## CI/CD Usage

For automated deployments, set the `LEAFPRESS_NETLIFY_TOKEN` environment variable:

```bash
export LEAFPRESS_NETLIFY_TOKEN=nf_xxxxxxxxxxxxxxxxxxxxxxxxxxxx
leafpress deploy
```

### GitHub Actions Example

```yaml
name: Deploy to Netlify

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
          LEAFPRESS_NETLIFY_TOKEN: ${{ secrets.NETLIFY_TOKEN }}
        run: leafpress deploy --provider netlify
```

Add your Netlify token as a repository secret named `NETLIFY_TOKEN` in your repository settings.

## Configuration

Deploy settings are stored in `leafpress.json`:

```json
{
  "deploy": {
    "provider": "netlify",
    "settings": {
      "site_id": "abc123def456"
    }
  }
}
```

## Dry Run

Validate your setup without deploying:

```bash
leafpress deploy --dry-run
```

This will show what would be deployed without actually uploading files.

## Site Management

### Select an Existing Site

When running the wizard, choose from your existing sites in the interactive menu.

### Create a New Site

Select "Create new site" in the wizard and enter a site name. Netlify will:
- Generate a unique subdomain: `sitename.netlify.app`
- Create the site in your account
- Begin deployment

### Switch Sites

To deploy to a different Netlify site:

```bash
leafpress deploy --reconfigure
```

This will re-run the setup wizard and let you choose a different site.

## Custom Domain

To use a custom domain with your Netlify site:

1. Deploy your site first
2. Go to your site in the [Netlify dashboard](https://app.netlify.com)
3. Navigate to **Domain settings → Custom domains**
4. Add your custom domain
5. Update your DNS records (Netlify will provide instructions)
6. Update `baseURL` in `leafpress.json`:

```json
{
  "baseURL": "https://yourdomain.com"
}
```

Netlify automatically provides free SSL certificates for custom domains via Let's Encrypt.

## Features

- **Smart Uploads**: Only changed files are uploaded, saving time and bandwidth
- **Instant Rollbacks**: Each deploy is individually versioned
- **Edge CDN**: Content delivered globally from edge servers
- **Automatic SSL**: HTTPS included automatically
- **Preview URLs**: Each deploy gets a unique preview URL
- **Deploy Previews**: Optional branch deploys for testing
- **Deployment Tracking**: `leafpress status` shows pending changes since last deploy

## Check What's Pending

Before deploying, see what files have changed:

```bash
leafpress status
```

This shows which files are new, modified, or deleted since your last deployment. See [[guide/deployment-status|Deployment Status]] for details.

## Reconfigure

To change sites, re-authenticate, or update settings:

```bash
leafpress deploy --reconfigure
```

## Security Best Practices

**Token Management**:
- Never commit tokens to git or any version control
- Regenerate tokens if they're ever exposed
- Use GitHub secrets for CI/CD, never hardcode tokens
- Consider rotating tokens periodically for added security

**Token Expiration & Rotation**:
- [Netlify Personal Access Tokens](https://app.netlify.com/user/applications) can be revoked anytime
- Regularly review and rotate tokens for security
- When revoking an old token, generate a new one and update your CI/CD secrets immediately

**GitHub Secrets**:
- Use [repository secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets) for CI/CD tokens
- Restrict secret access to workflows that need them
- Never echo or log tokens in CI/CD output

**Token Scope**:
- Use the minimum necessary permissions for your token
- Netlify tokens should have deploy permissions only

## Troubleshooting

**"Invalid token" error?**
- Check that you copied the full token from Netlify
- Tokens can only be viewed once; generate a new one if needed
- Make sure the token has not expired

**Site not updating?**
- Netlify deploys are usually instant
- Check the Deploy Log in your Netlify site settings for any errors
- Verify the site ID is correct in `leafpress.json`

**Files not being uploaded?**
- Check that your `_site` directory contains the built files
- Run `leafpress build` first to generate the site
- Use `--dry-run` to see which files would be uploaded
