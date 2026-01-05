---
title: "Installation"
date: 2025-12-21
---

Get leafpress running in under a minute.

## Install

### Using Go (recommended)

```bash
go install github.com/shivamx96/leafpress/cli/cmd/leafpress@latest
```

Requires Go 1.21+. The binary is added to your `$GOPATH/bin`.

### Download Binary

Grab the latest release from [GitHub Releases](https://github.com/shivamx96/leafpress/releases):

```bash
# macOS / Linux
curl -L https://github.com/shivamx96/leafpress/releases/latest/download/leafpress-$(uname -s)-$(uname -m).tar.gz | tar xz
sudo mv leafpress /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/shivamx96/leafpress.git
cd leafpress/cli
go build -o leafpress ./cmd/leafpress
sudo mv leafpress /usr/local/bin/
```

## Create Your First Site

```bash
leafpress init my-garden
cd my-garden
```

This creates:

```
my-garden/
├── content/
│   └── index.md          # Your homepage
├── static/
│   └── images/           # Put images here
├── leafpress.json        # Configuration
└── style.css             # Custom styles (optional)
```

## Start Writing

```bash
leafpress serve
```

Open [http://localhost:3000](http://localhost:3000). Edit any markdown file—changes appear instantly.

## Build for Production

```bash
leafpress build
```

Static files are generated in `_site/`. Upload this folder to any web host.

## Next Steps

- [[guide/writing|Writing Content]] — Learn the markdown features
- [[guide/configuration|Configuration]] — Customize your site

