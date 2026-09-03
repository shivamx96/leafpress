# Leafpress Obsidian Plugin - PRD

## Overview
Enable Obsidian users to initialize, preview, and build their vaults as
Leafpress digital gardens without using a terminal. The plugin bundles the
Leafpress CLI and opens the generated static output for use with any hosting
provider.

## Goals
- Make Leafpress accessible to non-technical Obsidian users
- One-click preview and build workflow
- Keep hosting credentials and deployment outside the plugin
- Cross-platform support (macOS, Linux, Windows)

## Key Features

| Feature | Description |
|---------|-------------|
| **Initialize** | Wizard to create `leafpress.json` in vault root |
| **Build** | One-click build with progress indicator |
| **Preview** | Start and stop the local preview server |
| **Open Output** | Reveal the generated static-site directory |
| **Settings** | Select the binary version and build options |

## User Flow

```
1. User installs plugin
2. First command: "Initialize" → wizard creates leafpress.json
3. Edit vault content as normal
4. Click "Preview" while editing
5. Click "Build" to generate the static site
6. Publish the output with provider-native tooling or CI
```

## Technical Approach

- **Binary Bundling**: Download CLI from GitHub Releases on first use, store in vault's `.obsidian/plugins/leafpress/bin/`
- **Execution**: Spawn CLI as subprocess, capture output for UI feedback
- **Platform Detection**: Auto-detect macOS (Intel/ARM), Linux, Windows at runtime

## Out of Scope
- Config file UI editor (users edit `leafpress.json` directly)
- Theme customization UI
- Multi-site support per vault
- Hosting authentication and deployment

## Success Criteria
- Initialize site in < 2 minutes
- Preview or build with ≤ 2 clicks
- Binary downloads correctly on all platforms
- No CLI knowledge required
