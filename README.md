# LeafPress

A CLI-driven static site generator for digital gardens. Transform a folder of Markdown files into a clean, interlinked website with minimal configuration.

**Your garden folder IS the product. LeafPress is invisible infrastructure.**

## Features

- **Wiki-style linking** - Use `[[Page Name]]` to link between notes
- **Backlinks** - Automatically track which pages link to each other
- **Growth stages** - Mark pages as seedling 🌱, budding 🌿, or evergreen 🌳
- **Tags & organization** - Auto-generated tag pages and section indexes
- **Live reload** - Development server with instant preview
- **Zero config** - Sensible defaults, customize only what you need
- **Graph export** - Visualize your knowledge graph with `graph.json`

## Installation

### From Source

```bash
git clone https://github.com/shivamx96/leafpress.git
cd leafpress
go build -o leafpress ./cmd/leafpress
```

Add the binary to your PATH or move it to a location in your PATH:

```bash
sudo mv leafpress /usr/local/bin/
```

## Quick Start

### 1. Initialize a new garden

```bash
mkdir my-garden
cd my-garden
leafpress init
```

This creates:
- `leafpress.json` - Configuration file
- `index.md` - Your homepage
- `style.css` - Custom styles (optional)
- `.gitignore` - Git ignore rules

### 2. Create your first note

```bash
leafpress new "notes/Hello World"
```

This creates `notes/hello-world.md` with frontmatter:

```markdown
---
title: "Hello World"
date: 2025-12-21
tags: []
draft: true
growth: "seedling"
---

Your content here...
```

### 3. Start the development server

```bash
leafpress serve --drafts
```

Visit http://localhost:3000 to see your garden. The server rebuilds automatically when you edit files.

### 4. Build for production

```bash
leafpress build
```

Your static site is generated in `_site/` directory.

## Usage

### Commands

```bash
leafpress init              # Initialize a new site
leafpress new <path>        # Create a new page
leafpress build             # Build static site
leafpress build --drafts    # Include draft pages
leafpress serve             # Start dev server
leafpress serve --drafts    # Serve with drafts
leafpress serve --port 8080 # Use custom port
```

### Frontmatter

All pages support YAML frontmatter:

```yaml
---
title: "Page Title"
date: 2025-12-21
tags: [programming, go]
draft: false
growth: "budding"  # seedling, budding, or evergreen
---
```

### Wiki Links

Link to other pages using wiki-style syntax:

```markdown
Check out [[My Other Note]] for more details.

You can also use [[Custom Text|notes/my-note]].
```

### Section Indexes

Create `_index.md` in any directory to customize its index page:

```markdown
---
title: "My Projects"
sort: "date"  # date, title, or growth
---

Here are my projects...
```

Without `_index.md`, LeafPress auto-generates index pages.

## Configuration

Edit `leafpress.json` to customize your site:

```json
{
  "title": "My Digital Garden",
  "baseURL": "https://example.com",
  "outputDir": "_site",
  "port": 3000,
  "nav": [
    {"label": "Notes", "path": "/notes/"},
    {"label": "Projects", "path": "/projects/"}
  ],
  "theme": {
    "font": "Inter",
    "accent": "#4a9eff"
  },
  "graph": true
}
```

### Theme Options

**Fonts**: Any Google Fonts family name (e.g., "Inter", "Space Grotesk", "Roboto")

**Accent**: Any CSS color (hex, rgb, hsl)

## Project Structure

```
my-garden/
├── leafpress.json          # Configuration
├── index.md                # Homepage
├── style.css               # Custom styles (optional)
├── notes/
│   ├── _index.md          # Notes section page (optional)
│   ├── first-note.md
│   └── second-note.md
├── projects/
│   └── my-project.md
└── static/                 # Static assets (optional)
    └── images/
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/shivamx96/leafpress.git
cd leafpress

# Install dependencies
go mod download

# Build
go build -o leafpress ./cmd/leafpress
```

### Testing

```bash
# Run the test suite (when available)
go test ./...

# Manual testing
./leafpress init
./leafpress new "test/Sample Page"
./leafpress build --drafts --verbose
./leafpress serve --drafts
```

### Running Tests with Test Garden

```bash
# Build the binary
go build -o leafpress ./cmd/leafpress

# Create a test garden
mkdir -p /tmp/test-garden
cd /tmp/test-garden

# Initialize
../leafpress init

# Create test pages
../leafpress new "notes/First Note"
../leafpress new "projects/Test Project"

# Build (excluding drafts)
../leafpress build

# Build including drafts
../leafpress build --drafts --verbose

# Serve locally
../leafpress serve --drafts --port 3000
```

### Project Structure

- `cmd/leafpress/` - CLI entry point
- `internal/cli/` - CLI command implementations
- `internal/content/` - Markdown parsing, rendering, wiki links
- `internal/build/` - Site generation logic
- `internal/config/` - Configuration management
- `internal/templates/` - HTML templates and CSS
- `internal/server/` - Development server with live reload
- `testdata/garden/` - Example garden for testing

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Goldmark](https://github.com/yuin/goldmark) - Markdown parser
- [fsnotify](https://github.com/fsnotify/fsnotify) - File watcher
