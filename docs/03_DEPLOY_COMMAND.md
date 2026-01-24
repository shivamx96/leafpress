# Leafpress Deploy

**Date:** January 2026

---

## Overview

A single `leafpress deploy` command that enables one-command publishing to multiple hosting providers, with browser-based OAuth authentication for a frictionless setup experience.

---

## Problem Statement

Static site authors currently face a fragmented deployment experience: each hosting provider has its own CLI, authentication flow, and configuration. Non-technical users (especially Obsidian users wanting to publish digital gardens) find the terminal-to-dashboard context switching overwhelming.

---

## Goals

1. **Zero-config first deploy** - New users run `leafpress deploy` and get a live site in under 2 minutes
2. **No CLI dependencies** - Authentication via browser OAuth, not `gh`/`netlify`/`vercel` CLIs
3. **Provider flexibility** - Support 3 providers at launch; architecture supports adding more
4. **Future-ready** - Clean abstraction layer for eventual Leafpress Cloud integration

---

## User Experience

### First-Time Deploy Flow

```
$ leafpress deploy

No deploy configuration found. Let's set one up!

? Select a deploy provider:
  > GitHub Pages (free, works with any GitHub repo)
    Netlify (free tier, deploy previews, easy custom domains)  
    Vercel (free tier, fast edge network)

? GitHub Pages selected. Let's authenticate.

→ Opening browser to authorize Leafpress...
  If browser doesn't open, visit: https://github.com/login/device
  And enter code: ABCD-1234

Waiting for authorization... ✓

✓ Authenticated as shivam

? Which repository?
  > shivam/my-garden
    shivam/notes
    shivam/blog

? Deploy branch: [gh-pages]

✓ Configuration saved to leafpress.toml

Building site...
✓ Built 47 pages in 120ms

Deploying to GitHub Pages...
✓ Deployed! Live at https://shivam.github.io/my-garden
```

### Subsequent Deploys

```
$ leafpress deploy

Building site...
✓ Built 47 pages in 118ms

Deploying to GitHub Pages...
✓ Deployed! Live at https://shivam.github.io/my-garden
```

---

## Supported Providers

### Launch Providers

| Provider | Auth Method | Free Tier | Best For |
|----------|-------------|-----------|----------|
| GitHub Pages | Device OAuth | Unlimited (public repos) | Developers already on GitHub |
| Netlify | Device OAuth | 100GB bandwidth, 300 build min | Custom domains, deploy previews |
| Vercel | Device OAuth | 100GB bandwidth | Speed-focused users |

### Provider Selection Criteria

- Established player with stable API
- Generous free tier for hobbyist/personal sites
- OAuth or device flow support (no manual token pasting)
- Simple "upload a folder" deployment model

---

## Technical Design

### Configuration

**Project config** (`leafpress.toml`):
```toml
[deploy]
provider = "github-pages"
repo = "shivam/my-garden"
branch = "gh-pages"
```

**Credentials** (`~/.config/leafpress/credentials.toml`):
```toml
[github]
access_token = "gho_xxxx"
username = "shivam"

[netlify]
access_token = "xxxx"
```

Credentials are stored outside project directory to prevent accidental commits.

### Authentication: OAuth Device Flow

All providers use device flow where supported:

```
1. CLI requests device code from provider
2. CLI displays URL + user code, opens browser
3. User authorizes in browser
4. CLI polls for access token
5. Token stored in ~/.config/leafpress/credentials.toml
```

**Fallback:** If device flow unavailable (some Vercel configurations), prompt user to create token in dashboard and paste it.

### Provider Interface

```go
type DeployProvider interface {
    // Provider identity
    Name() string
    Description() string
    
    // Authentication
    Authenticate(ctx context.Context) (*Credentials, error)
    ValidateCredentials(creds *Credentials) error
    
    // Deployment
    Deploy(ctx context.Context, buildDir string, opts DeployOpts) (*DeployResult, error)
    
    // Interactive setup
    Configure(ctx context.Context, creds *Credentials) (*ProviderConfig, error)
}

type DeployResult struct {
    URL        string
    DeployID   string
    DeployedAt time.Time
}
```

### CLI Flags

```
leafpress deploy [flags]

Flags:
  --provider string      Deploy provider (github-pages|netlify|vercel)
  --skip-build          Deploy existing build without rebuilding
  --reconfigure         Re-run provider setup wizard
  --dry-run             Build and validate without deploying
```

### CI/CD Support

Detect non-interactive environments (`!isatty(stdout)`):
- Require explicit `--provider` flag
- Read tokens from environment variables: `LEAFPRESS_GITHUB_TOKEN`, `LEAFPRESS_NETLIFY_TOKEN`, etc.
- Fail with clear error message if config missing

---

## Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| Browser doesn't open | Display manual URL prominently with copy-paste code |
| OAuth timeout (5 min) | Cancel with message, suggest retry |
| Token expired | Auto-detect on deploy failure, re-trigger auth flow |
| Repo not found | List available repos, let user select |
| Rate limited | Exponential backoff with user-friendly wait message |
| Build fails | Stop before deploy, show build errors |
| No git remote | Prompt for manual repo input |

---

## Future: Obsidian Plugin

Phase 2 will introduce an Obsidian plugin that wraps this functionality:

- Plugin bundles platform-specific Leafpress binary (downloaded on first run)
- Settings UI replaces interactive prompts
- Command palette: "Leafpress: Publish" / "Leafpress: Preview"
- OAuth opens in system browser, returns to Obsidian

### Obsidian-Specific Considerations

- Handle `[[wikilinks]]` → standard markdown links
- Resolve Obsidian's flexible asset paths
- Gracefully degrade callouts, Dataview blocks
- Desktop-only (mobile Obsidian can't run binaries)

---

## Future: Leafpress Cloud

Phase 3 introduces Leafpress Cloud as a provider option:

```
? Select a deploy provider:
    GitHub Pages
    Netlify
    Vercel
  > Leafpress Cloud (one-click setup, custom domains included)
```

**Value proposition over free providers:**
- Single OAuth (no provider account needed)
- Custom domains without DNS configuration complexity
- Integrated analytics
- Collaboration features (multiple editors)

**Pricing:** ~$5-10/month for hobbyist tier

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Time to first deploy (new user) | < 2 minutes |
| Deploy command success rate | > 95% |
| Users completing OAuth flow | > 90% |
| Repeat deploys (7-day retention) | > 60% |

---

## Open Questions

1. **Custom domains at launch?** Each provider handles this differently. Consider punting to Phase 2.
2. **Deploy previews?** Netlify/Vercel support this. Expose via `leafpress deploy --preview`?
3. **Monorepo support?** Auto-detect leafpress.toml location or require flag?

---

## Implementation Phases

**Phase 1 (MVP):**
- GitHub Pages provider with device OAuth
- Interactive setup wizard
- Basic deploy command

**Phase 2:**
- Netlify + Vercel providers
- `--reconfigure` and `--dry-run` flags
- CI/CD environment detection

**Phase 3:**
- Obsidian plugin (beta)
- Token refresh handling

**Phase 4:**
- Leafpress Cloud provider
- Obsidian plugin (stable)
