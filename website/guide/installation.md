---
title: "Installation"
date: 2025-12-21
---

Get leafpress up and running in minutes.

## Download Binary

Download the latest release from [GitHub Releases](https://github.com/shivamx96/leafpress/releases).

```bash
# Extract and move to PATH
tar -xzf leafpress-*.tar.gz
sudo mv leafpress /usr/local/bin/
```

## Build from Source

Requires Go 1.21 or higher.

```bash
# Clone the repository
git clone https://github.com/shivamx96/leafpress.git
cd leafpress

# Build
go build -o leafpress ./cmd/leafpress

# Move to PATH (optional)
sudo mv leafpress /usr/local/bin/
```

## Verify Installation

```bash
leafpress --version
```

## Next Steps

Now that leafpress is installed:

1. [[guide/configuration|Configure your site]]
2. [[guide/writing|Start writing content]]
3. Learn about [[guide/wiki-links|wiki-style linking]]
