# Leafpress deploy command

`leafpress deploy` builds the current garden and publishes `_site/` to GitHub
Pages, Netlify, or Vercel. First use runs an interactive configuration wizard;
later deploys reuse the project configuration and credential store.

## Commands and flags

```text
leafpress deploy
leafpress deploy --provider github-pages
leafpress deploy --skip-build
leafpress deploy --reconfigure
leafpress deploy --dry-run
leafpress status
```

- `--provider` selects `github-pages`, `netlify`, or `vercel`.
- `--skip-build` deploys the existing output directory.
- `--reconfigure` reruns provider authentication and setup.
- `--dry-run` builds and validates without publishing.
- `status` compares current source hashes with the last deployment manifest.

## Stored configuration

Provider selection and non-secret settings are part of `leafpress.json`:

```json
{
  "deploy": {
    "provider": "github-pages",
    "settings": {
      "repo": "example/my-garden",
      "branch": "gh-pages"
    }
  }
}
```

Credentials are JSON outside the project directory. On macOS and Linux the
default is `~/.config/leafpress/credentials.json`; on Windows it is under the
user's roaming application-data directory. The file is written with owner-only
permissions where the platform supports Unix modes.

```json
{
  "github-pages": {
    "provider": "github-pages",
    "accessToken": "...",
    "username": "example"
  }
}
```

Automation can avoid the credential file by setting one of:

```text
LEAFPRESS_GITHUB_TOKEN
LEAFPRESS_NETLIFY_TOKEN
LEAFPRESS_VERCEL_TOKEN
```

The environment variable takes precedence over stored credentials.

## Authentication

- GitHub Pages uses GitHub's OAuth device flow.
- Vercel uses its device authorization flow.
- Netlify requests a Personal Access Token with hidden terminal input. In a
  non-interactive environment, use `LEAFPRESS_NETLIFY_TOKEN` instead.

Tokens are sent only in provider authorization headers or Git's environment
configuration; they are not embedded in remote URLs.

## Provider behavior

### GitHub Pages

The wizard selects a repository and deploy branch (default `gh-pages`). Each
deployment clones or initializes that branch in a temporary directory, replaces
its contents with the built site, adds `.nojekyll`, commits with a Leafpress
deployment identity, and pushes over authenticated HTTPS.

### Netlify

The wizard selects or creates a site. Deployment hashes files, asks Netlify
which blobs are missing, uploads only required hashes, and finalizes the deploy.

### Vercel

The wizard selects or creates a project. Deployment uploads the build files and
creates a production deployment using the configured project/team identifiers.

## Deployment state

After a successful non-dry-run deployment, Leafpress writes
`.leafpress-deploy-state.json` in the project root. It records the last deploy,
up to ten history entries, deployed-file hashes, and source-file hashes. This is
the data used by `leafpress status` to report added, modified, and deleted files.
It does not contain provider credentials.

## Failure behavior

- A build error stops deployment before any provider mutation.
- Missing or invalid credentials produce a reconfiguration instruction.
- Provider and Git failures are returned without recording a successful deploy.
- `Ctrl+C` cancels the active deployment context.
- Manifest write failures are warnings after a provider deployment succeeds;
  the published site remains live, but `status` may lack the latest state.

## Implementation boundary

Providers implement `internal/deploy.Provider`, which owns authentication,
credential validation, configuration, and deployment. The CLI owns config
loading, optional build execution, provider selection, and manifest recording.
