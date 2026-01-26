---
title: "Custom Deployment"
date: 2025-12-21
---

Deploy your leafpress site to any static hosting platform.

## Build Your Site

First, generate the static files:

```bash
leafpress build
```

This creates your site in `_site/` (or your configured `outputDir`). Upload this folder to any web host.

## Netlify

For the simplest setup, see [[guide/deploy-netlify|Deploy to Netlify]] for one-command deployment with `leafpress deploy --provider netlify`.

For Git-based continuous deployment:

### Git Integration

1. Push your site to GitHub/GitLab
2. Connect repo in Netlify dashboard
3. Set build settings:
   - **Build command**: `go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build`
   - **Publish directory**: `_site`

### netlify.toml

```toml
[build]
  command = "go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build"
  publish = "_site"

[[redirects]]
  from = "/*"
  to = "/404.html"
  status = 404
```

## Docker/Nginx

### Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder
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

### Build and Run

```bash
docker build -t my-garden .
docker run -p 8080:80 my-garden
```

## Any Static Host

leafpress generates plain HTML, CSS, and JavaScript. The `_site/` folder can be uploaded to any static hosting service:

1. Run `leafpress build`
2. Upload the contents of `_site/` to your host
3. Ensure your host serves `404.html` for missing pages

## Custom Domain

Set `baseURL` in `leafpress.json` for correct sitemap and canonical URLs:

```json
{
  "baseURL": "https://yourdomain.com"
}
```

This ensures:
- Sitemap URLs are correct
- Canonical links point to your domain
- Open Graph URLs are accurate
