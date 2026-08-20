---
title: "Installation"
date: 2025-12-21
---

Get leafpress running in under a minute.

## Install

### Quick Install (recommended)

```bash
curl -fsSL https://leafpress.in/install.sh | sh
```

Works on macOS (Intel/Apple Silicon), Ubuntu, Fedora, and Arch Linux.

### Using Go

```bash
go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest
```

Requires Go 1.25+. The binary is added to your `$GOPATH/bin`.

### Download Binary

Grab the latest release from [GitHub Releases](https://github.com/shivamx96/leafpress/releases).

Available binaries:
- `leafpress-vX.X.X-darwin-amd64.tar.gz` (macOS Intel)
- `leafpress-vX.X.X-darwin-arm64.tar.gz` (macOS Apple Silicon)
- `leafpress-vX.X.X-linux-amd64.tar.gz` (Linux x86_64)
- `leafpress-vX.X.X-linux-arm64.tar.gz` (Linux ARM64)

### Build from Source

```bash
git clone https://github.com/shivamx96/leafpress.git
cd leafpress/cli
go build -o leafpress ./cmd/leafpress
sudo mv leafpress /usr/local/bin/
```

## Create Your First Site

```bash
mkdir my-garden && cd my-garden
leafpress init
```

`leafpress init` scaffolds into the current directory, so create and enter the folder first. This creates:

```
my-garden/
├── index.md              # Your homepage
├── static/
│   ├── images/           # Put images here
│   ├── video/
│   └── audio/
├── leafpress.json        # Configuration
└── style.css             # Custom styles (optional)
```

## Start Writing

```bash
leafpress serve
```

Open [http://localhost:3000](http://localhost:3000). Edit any markdown file—changes appear instantly.

The dev server listens on `127.0.0.1` only, so your work-in-progress garden is
not reachable from the rest of the network. To preview on a phone or another
machine, opt in explicitly:

```bash
leafpress serve --host 0.0.0.0
```

There is no authentication, so only do that on a network you trust — and
remember `--drafts` would expose unpublished pages too.

## Build for Production

```bash
leafpress build
```

Static files are generated in `_site/`. Upload this folder to any web host.

## Update leafpress

Update to the latest version with a single command:

```bash
leafpress update
```

This checks GitHub for the latest release and replaces your binary automatically. Use `--force` to reinstall even if you're on the latest version.

## Next Steps

- [[guide/writing|Writing Content]] — Learn the markdown features
- [[guide/configuration|Configuration]] — Customize your site

