---
title: "Deployment Status"
date: 2025-01-26
---

Track what files have changed since your last deployment.

## Check Deployment Status

```bash
leafpress status
```

This command shows:
- When your site was last deployed
- How many files are pending deployment
- Which files have changed since the last deploy

## Example Output

```
Deployment Status
=================

Provider:     netlify
Last Deploy:  2 hours ago
Live URL:     https://my-site.netlify.app
Deployed:     47 files

⚠ 3 file(s) pending deployment:

  + posts/new-post.md
  ~ style.css
  - old-page.md

Run 'leafpress deploy' to deploy these changes.
```

## Understanding File Status

Each file shows a status indicator:

- **`+` (new)** — File was added since last deployment
- **`~` (modified)** — File was changed since last deployment
- **`-` (deleted)** — File was removed since last deployment

## Deployment Manifest

Leafpress stores a `.leafpress-deploy-state.json` file in your project root (added to `.gitignore`). This file tracks:

- Last deployment timestamp
- Deployed provider and URL
- SHA1 hash of each deployed file
- Deployment history (last 10 deployments)

This allows the status command to know exactly which files have changed without rebuilding or analyzing git history.

## Use Cases

### Before Deploying

Check what's ready to deploy:
```bash
leafpress status
```

If everything looks good:
```bash
leafpress deploy
```

### Troubleshooting

If a file isn't showing as pending even though you changed it, the hashes might be out of sync. Rebuild and check again:

```bash
leafpress build
leafpress status
```

### CI/CD Integration

In your CI/CD pipeline, you can check if there are pending changes:

```bash
leafpress status
leafpress deploy
```

The deploy will only upload files that have changed, making deployments faster for large sites.

## Obsidian Plugin Integration

The deployment manifest enables the Obsidian plugin to show you pending changes directly in Obsidian:

- See "3 files pending" badge in the sidebar
- Know exactly which files need deployment before clicking "Deploy"
- One-click deployment of only changed files

## Files to Ignore

The `.leafpress-deploy-state.json` file should be added to your `.gitignore`:

```
_site/
node_modules/
.leafpress-deploy-state.json
```

This file is project-specific and shouldn't be shared with collaborators. Each deployment location tracks its own state independently.
