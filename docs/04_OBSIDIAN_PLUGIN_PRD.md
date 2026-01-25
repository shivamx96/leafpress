# Leafpress Obsidian Plugin - PRD

## Overview
Enable Obsidian users to build and deploy their vaults as Leafpress digital gardens without touching CLI. Plugin bundles the Leafpress CLI and provides a simple UI for initialization, building, and deploying to GitHub Pages or Vercel.

## Goals
- Make Leafpress accessible to non-technical Obsidian users
- One-click build and deploy workflow
- Automatic credential handling via Leafpress's existing auth system
- Cross-platform support (macOS, Linux, Windows)

## Key Features

| Feature | Description |
|---------|-------------|
| **Initialize** | Wizard to create `leafpress.json` in vault root |
| **Build** | One-click build with progress indicator |
| **Deploy** | Single command to deploy; handles auth on first use |
| **Status Panel** | Sidebar widget showing deployment status and history |
| **Settings** | Manage credentials, re-authenticate, check binary version |

## User Flow

```
1. User installs plugin
2. First command: "Initialize" → wizard creates leafpress.json
3. Edit vault content as normal
4. Click "Deploy" → browser opens for OAuth (first time only)
5. Subsequent deploys: one click, done
6. Sidebar shows deployment URL and status
```

## Technical Approach

- **Binary Bundling**: Download CLI from GitHub Releases on first use, store in vault's `.obsidian/plugins/leafpress/bin/`
- **Credential Storage**: Leverage Leafpress's `~/.config/leafpress/credentials.json` (same as CLI)
- **Execution**: Spawn CLI as subprocess, capture output for UI feedback
- **Platform Detection**: Auto-detect macOS (Intel/ARM), Linux, Windows at runtime

## Out of Scope
- Config file UI editor (users edit `leafpress.json` directly)
- Theme customization UI
- Multi-site support per vault
- Custom deployment providers

## Success Criteria
- Initialize site in < 2 minutes
- Deploy with ≤ 3 clicks after first auth
- Binary downloads correctly on all platforms
- No CLI knowledge required