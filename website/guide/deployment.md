---
title: "Deployment"
date: 2025-12-21
---

Deploy your leafpress site to various hosting platforms.

## Build for Production

First, build your site:

```bash
leafpress build
```

This generates static HTML files in the `_site` directory (or your configured `outputDir`).

## Netlify

### Deploy via Git

1. Push your site to a Git repository
2. Connect your repo to Netlify
3. Configure build settings:
   - **Build command**: `leafpress build`
   - **Publish directory**: `_site`

### Deploy via CLI

```bash
# Install Netlify CLI
npm install -g netlify-cli

# Build your site
leafpress build

# Deploy
netlify deploy --prod --dir=_site
```

## Vercel

### Deploy via Git

1. Import your Git repository in Vercel
2. Configure build settings:
   - **Build command**: `leafpress build`
   - **Output directory**: `_site`

### Deploy via CLI

```bash
# Install Vercel CLI
npm install -g vercel

# Build and deploy
leafpress build
vercel --prod
```

## GitHub Pages

Add a GitHub Actions workflow (`.github/workflows/deploy.yml`):

```yaml
name: Deploy to GitHub Pages

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Download leafpress
        run: |
          curl -L https://github.com/shivamx96/leafpress/releases/latest/download/leafpress-linux-amd64.tar.gz | tar xz
          chmod +x leafpress
      
      - name: Build
        run: ./leafpress build
      
      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./_site
```

Configure GitHub Pages to use the `gh-pages` branch.

## Cloudflare Pages

1. Connect your Git repository to Cloudflare Pages
2. Configure build settings:
   - **Build command**: `leafpress build`
   - **Build output directory**: `_site`

## Custom Server

Deploy to any static file hosting:

```bash
# Build the site
leafpress build

# Copy _site directory to your web server
rsync -avz _site/ user@server:/var/www/html/
```

Or use any web server (nginx, Apache, etc.) to serve the `_site` directory.

### 404 Page Configuration

Leafpress generates a `404.html` file automatically. Most platforms serve this for missing routes out of the box. For custom servers:

**nginx:**
```nginx
error_page 404 /404.html;
```

**Apache (.htaccess):**
```apache
ErrorDocument 404 /404.html
```

**AWS S3/CloudFront:** Set the error document to `404.html` in bucket settings.

## Custom Domain

After deploying, configure your custom domain in your hosting platform's settings. Update the `baseURL` in `leafpress.json`:

```json
{
  "baseURL": "https://yourdomain.com"
}
```

## Next Steps

- [[guide/configuration|Configure your site]]
- [[guide/writing|Write content]]
- [[features|Explore features]]
