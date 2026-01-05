---
title: "Deployment"
date: 2025-12-21
---

Deploy your leafpress site to any static hosting platform.

## Build

```bash
leafpress build
```

This generates static files in `_site/` (or your configured `outputDir`). Upload this folder to any web host.

## Netlify

### Git Integration

1. Push your site to GitHub/GitLab
2. Connect repo in Netlify dashboard
3. Set build settings:
   - **Build command**: `leafpress build`
   - **Publish directory**: `_site`

### netlify.toml

```toml
[build]
  command = "leafpress build"
  publish = "_site"

[[redirects]]
  from = "/*"
  to = "/404.html"
  status = 404
```

## Vercel

### Git Integration

1. Push to GitHub/GitLab
2. Import in Vercel dashboard
3. Set framework preset to "Other"
4. Build command: `leafpress build`
5. Output directory: `_site`

### vercel.json

```json
{
  "buildCommand": "leafpress build",
  "outputDirectory": "_site"
}
```

## GitHub Pages

### Using GitHub Actions

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy

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
          go-version: '1.21'
      
      - name: Install leafpress
        run: go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest
      
      - name: Build
        run: leafpress build
      
      - name: Deploy
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./_site
```

Enable GitHub Pages in repo settings, set source to `gh-pages` branch.

## Cloudflare Pages

1. Connect your Git repository
2. Set build configuration:
   - **Build command**: `go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build`
   - **Build output directory**: `_site`

## AWS S3 + CloudFront

### Upload to S3

```bash
leafpress build
aws s3 sync _site/ s3://your-bucket-name --delete
```

### CloudFront 404 Handling

Set custom error response:
- HTTP error code: 404
- Response page path: `/404.html`
- HTTP response code: 404

## Docker / Nginx

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
RUN go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest
WORKDIR /site
COPY . .
RUN leafpress build

FROM nginx:alpine
COPY --from=builder /site/_site /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
```

### nginx.conf

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ =404;
    }

    error_page 404 /404.html;
    location = /404.html {
        internal;
    }
}
```

## Custom Domain

Set `baseURL` in `leafpress.json` for correct sitemap and canonical URLs:

```json
{
  "baseURL": "https://yourdomain.com"
}
```

