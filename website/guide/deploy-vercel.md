---
title: "Deploy to Vercel"
date: 2025-12-21
---

Deploy your leafpress site to Vercel with Git-based continuous deployment.

## Quick Start

1. Push your site to GitHub/GitLab
2. Import the repository in Vercel dashboard
3. Configure build settings
4. Deploy

## Setup

### 1. Connect Repository

1. Go to [vercel.com](https://vercel.com) and sign in
2. Click "Add New Project"
3. Import your Git repository

### 2. Configure Build Settings

Set the following in the Vercel project settings:

- **Framework Preset**: Other
- **Build Command**: `go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build`
- **Output Directory**: `_site`
- **Install Command**: (leave empty)

### 3. Deploy

Click "Deploy" and Vercel will build and deploy your site. Future pushes to your main branch will trigger automatic deployments.

## Configuration File

Add a `vercel.json` to your project root for consistent settings:

```json
{
  "buildCommand": "go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest && leafpress build",
  "outputDirectory": "_site"
}
```

## Custom Domain

1. Go to your project settings in Vercel
2. Navigate to "Domains"
3. Add your custom domain
4. Update `baseURL` in `leafpress.json`:

```json
{
  "baseURL": "https://yourdomain.com"
}
```

## Environment Variables

If you need to set environment variables for your build:

1. Go to Project Settings > Environment Variables
2. Add variables as needed

## Preview Deployments

Vercel automatically creates preview deployments for pull requests. Each PR gets a unique URL to preview changes before merging.

## Troubleshooting

**Build failing?** Check that Go is available in Vercel's build environment. The `go install` command should handle this automatically.

**404 on routes?** Vercel handles client-side routing differently. For a static site like leafpress, this shouldn't be an issue as all pages are pre-rendered.

**Slow builds?** The first build downloads and installs leafpress. Subsequent builds use Vercel's cache and should be faster.
