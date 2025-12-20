# LeafPress Implementation Plan

## Overview

This document breaks down the LeafPress implementation into discrete phases with detailed tasks, dependencies, and acceptance criteria.

**Target:** Single Go binary with zero runtime dependencies.

---

## Phase 1: Foundation

**Goal:** Establish project structure, CLI framework, and configuration system.

### 1.1 Project Setup

**Tasks:**
- Initialize Go module (`go mod init github.com/shivam/leafpress`)
- Create directory structure:
  ```
  leafpress/
  ├── cmd/
  │   └── leafpress/
  │       └── main.go          # Entry point
  ├── internal/
  │   ├── cli/                 # Command handlers
  │   ├── config/              # Config parsing
  │   ├── content/             # Content model & parsing
  │   ├── build/               # Build pipeline
  │   ├── server/              # Dev server
  │   └── templates/           # HTML templates
  ├── pkg/                     # Public utilities (if any)
  ├── docs/
  ├── go.mod
  └── go.sum
  ```

**Acceptance Criteria:**
- [ ] `go build ./cmd/leafpress` produces a binary
- [ ] `./leafpress --version` prints version info

### 1.2 CLI Framework

**Dependencies:** `github.com/spf13/cobra`

**Tasks:**
- Implement root command with global flags:
  - `--config, -c` - Path to config file
  - `--verbose, -v` - Verbose logging
- Implement subcommands (stubs initially):
  - `init` - Scaffold project
  - `serve` - Dev server
  - `build` - Static site generation
  - `new <name>` - Create new page
- Add command-specific flags:
  - `serve --port, -p` - Override port
  - `serve --drafts, -d` - Include drafts
  - `build --drafts, -d` - Include drafts

**Acceptance Criteria:**
- [ ] `leafpress --help` shows all commands
- [ ] `leafpress init --help` shows init options
- [ ] Global flags propagate to subcommands

### 1.3 Configuration System

**Dependencies:** Standard library `encoding/json`

**Tasks:**
- Define `Config` struct matching schema:
  ```go
  type Config struct {
      Title     string     `json:"title"`
      BaseURL   string     `json:"baseURL"`
      OutputDir string     `json:"outputDir"`
      Port      int        `json:"port"`
      Nav       []NavItem  `json:"nav"`
      Theme     Theme      `json:"theme"`
      Graph     bool       `json:"graph"`
  }
  ```
- Implement config loading with defaults
- Implement config validation
- Create default config template for `init`

**Acceptance Criteria:**
- [ ] Missing config file returns sensible defaults
- [ ] Invalid JSON reports line number
- [ ] All fields have documented defaults

---

## Phase 2: Content Model

**Goal:** Parse markdown files with frontmatter and wiki-links into structured page objects.

### 2.1 Page Structure

**Tasks:**
- Define core `Page` struct:
  ```go
  type Page struct {
      // Metadata
      Title    string
      Date     time.Time
      Tags     []string
      Draft    bool
      Growth   string    // seedling | budding | evergreen
      
      // Paths
      SourcePath  string  // Relative path to .md file
      Slug        string  // URL slug
      OutputPath  string  // Path in _site/
      
      // Content
      RawContent  string  // Original markdown
      HTMLContent string  // Rendered HTML
      
      // Relationships
      Backlinks   []*Page
      OutLinks    []string  // Wiki-link targets
  }
  ```
- Implement slug generation from file path
- Implement output path generation (clean URLs)

**Acceptance Criteria:**
- [ ] `projects/leafpress.md` → slug: `projects/leafpress`
- [ ] Output: `_site/projects/leafpress/index.html`

### 2.2 Frontmatter Parser

**Dependencies:** `gopkg.in/yaml.v3`

**Tasks:**
- Implement YAML frontmatter extraction (between `---` delimiters)
- Parse into `Frontmatter` struct
- Apply defaults for missing fields:
  - `title` → filename (without extension)
  - `date` → file modification time
  - `tags` → empty slice
  - `draft` → false
  - `growth` → empty string
- Validate `growth` values (seedling|budding|evergreen)

**Acceptance Criteria:**
- [ ] Files without frontmatter use defaults
- [ ] Invalid YAML reports helpful error
- [ ] Growth validation rejects invalid values

### 2.3 Content Scanner

**Tasks:**
- Implement recursive directory walking
- Filter reserved paths:
  ```
  leafpress.json, style.css, static/, _site/,
  .leafpress/, .git/, .gitignore, .obsidian/, node_modules/
  ```
- Collect all `.md` files as `Page` objects
- Handle `_index.md` files specially (section index)

**Acceptance Criteria:**
- [ ] Reserved paths are excluded
- [ ] Nested directories are traversed
- [ ] `_index.md` marked as section index

### 2.4 Wiki-Link Parser

**Tasks:**
- Implement regex-based wiki-link extraction:
  - `[[page-name]]` → target: page-name, label: page-name
  - `[[path/to/page|Custom Label]]` → target: path/to/page, label: Custom Label
- Store extracted links in `Page.OutLinks`

**Acceptance Criteria:**
- [ ] Both link formats parsed correctly
- [ ] Links with special characters handled
- [ ] Multiple links per file extracted

### 2.5 Link Resolution

**Tasks:**
- Implement 3-step resolution algorithm:
  1. Exact slug match (case-insensitive)
  2. Filename match anywhere in tree
  3. Warn on ambiguity, pick first alphabetically
- Build backlinks map (reverse lookup)
- Handle broken links (warn, add `lp-broken-link` class)

**Acceptance Criteria:**
- [ ] `[[systems-thinking]]` resolves to `notes/systems-thinking.md`
- [ ] Ambiguous links log warning
- [ ] Broken links don't crash build

---

## Phase 3: Rendering Pipeline

**Goal:** Transform parsed content into HTML files with proper styling.

### 3.1 Markdown Renderer

**Dependencies:** `github.com/yuin/goldmark`

**Tasks:**
- Configure goldmark with extensions:
  - Tables
  - Strikethrough
  - Task lists
  - Syntax highlighting (optional)
- Create custom renderer for wiki-links:
  - Replace `[[link]]` with `<a class="lp-wikilink" href="...">...</a>`
  - Broken links: `<span class="lp-broken-link">...</span>`
- Add external link detection:
  - External: `<a class="lp-external" href="..." target="_blank" rel="noopener">... ↗</a>`

**Acceptance Criteria:**
- [ ] Standard markdown renders correctly
- [ ] Wiki-links become clickable anchors
- [ ] External links open in new tab

### 3.2 HTML Template System

**Dependencies:** Standard library `html/template`

**Tasks:**
- Create base layout template:
  ```
  <!DOCTYPE html>
  <html>
    <head>...</head>
    <body class="lp-body">
      <nav class="lp-nav">...</nav>
      <main class="lp-main">{{ block "content" . }}{{ end }}</main>
      <footer class="lp-footer">...</footer>
    </body>
  </html>
  ```
- Create component templates:
  - `page.html` - Single page with content + backlinks
  - `index.html` - Section listing page
  - `tag.html` - Tag page with filtered list
  - `tags-index.html` - All tags listing
- Implement template functions:
  - `growthEmoji(stage)` → emoji
  - `formatDate(date, format)` → formatted string

**Acceptance Criteria:**
- [ ] All page types render without error
- [ ] Template inheritance works
- [ ] Custom functions available

### 3.3 Embedded CSS

**Tasks:**
- Create default CSS with custom properties (embed in binary):
  ```css
  :root {
    --lp-font: "Inter", system-ui, sans-serif;
    --lp-accent: #4a9eff;
    --lp-bg: #ffffff;
    --lp-text: #1a1a1a;
    /* ... */
  }
  ```
- Implement theme injection from config
- Implement CSS merging (embedded + user `style.css`)
- Growth stage indicators via CSS

**Acceptance Criteria:**
- [ ] Site looks good with zero user CSS
- [ ] Theme config values reflected in output
- [ ] User CSS overrides work

---

## Phase 4: Build Pipeline

**Goal:** Orchestrate the full build process from source to `_site/`.

### 4.1 Build Command

**Tasks:**
- Implement build orchestration:
  1. Load config
  2. Scan content
  3. Parse all pages
  4. Resolve links
  5. Render markdown
  6. Generate HTML
  7. Copy static files
  8. Write CSS
  9. Generate tag pages
  10. Generate graph.json (if enabled)
- Clean output directory before build
- Report build statistics (pages, time)

**Acceptance Criteria:**
- [ ] `leafpress build` produces complete site
- [ ] All pages accessible via clean URLs
- [ ] Build completes in <1s for small sites

### 4.2 Tag System

**Tasks:**
- Collect all unique tags across pages
- Generate `/tags/index.html` with tag cloud
- Generate `/tags/<tag>/index.html` for each tag
- Sort pages within tag by date (descending)

**Acceptance Criteria:**
- [ ] Tag pages list correct posts
- [ ] Tag index shows all tags
- [ ] Empty tags handled gracefully

### 4.3 Section Indexes

**Tasks:**
- Detect `_index.md` in directories
- Parse section frontmatter (`title`, `sort`)
- Generate section listing with configured sort:
  - `date` - By date descending
  - `title` - Alphabetically
  - `growth` - By growth stage
- Render section intro content

**Acceptance Criteria:**
- [ ] Directories without `_index.md` get auto-generated index
- [ ] Custom sort orders work
- [ ] Section content renders above listing

### 4.4 Static File Handling

**Tasks:**
- Copy `static/` directory to `_site/static/`
- Preserve directory structure
- Skip hidden files (`.DS_Store`, etc.)

**Acceptance Criteria:**
- [ ] Images and assets accessible
- [ ] Nested directories preserved

### 4.5 Graph Generation

**Tasks:**
- Generate `graph.json` when `config.graph = true`:
  ```json
  {
    "nodes": [
      { "id": "slug", "title": "Title", "growth": "seedling" }
    ],
    "edges": [
      { "source": "slug1", "target": "slug2" }
    ]
  }
  ```

**Acceptance Criteria:**
- [ ] Graph includes all non-draft pages
- [ ] Edges represent wiki-links
- [ ] File only generated when enabled

---

## Phase 5: Dev Server

**Goal:** Provide fast iteration with live reload.

### 5.1 HTTP Server

**Dependencies:** Standard library `net/http`

**Tasks:**
- Serve `_site/` directory
- Handle clean URLs (try `/path/index.html`)
- Custom 404 page
- Configurable port with auto-increment on conflict

**Acceptance Criteria:**
- [ ] All pages accessible
- [ ] Port conflict handled gracefully
- [ ] 404 returns styled page

### 5.2 File Watcher

**Dependencies:** `github.com/fsnotify/fsnotify`

**Tasks:**
- Watch `.md` files and `style.css`
- Debounce rapid changes (100ms)
- Trigger rebuild on change
- Ignore `_site/` and `.leafpress/`

**Acceptance Criteria:**
- [ ] Changes detected within 200ms
- [ ] Only relevant files trigger rebuild
- [ ] No infinite loops from output changes

### 5.3 Live Reload

**Dependencies:** `github.com/gorilla/websocket`

**Tasks:**
- WebSocket endpoint at `/_lr`
- Inject reload script before `</body>` during serve
- Broadcast reload message after successful rebuild
- Handle client reconnection

**Acceptance Criteria:**
- [ ] Browser reloads automatically on save
- [ ] Multiple browsers sync
- [ ] Graceful handling of disconnects

### 5.4 Serve Command Integration

**Tasks:**
- Initial build before server start
- Start watcher, server, and WebSocket concurrently
- Graceful shutdown on SIGINT
- Print server URL on startup

**Acceptance Criteria:**
- [ ] `leafpress serve` starts everything
- [ ] Ctrl+C stops cleanly
- [ ] Errors during rebuild don't crash server

---

## Phase 6: Init & New Commands

**Goal:** Streamline project setup and content creation.

### 6.1 Init Command

**Tasks:**
- Generate `leafpress.json` with defaults
- Generate empty `style.css` with helpful comments
- Append to `.gitignore` (or create):
  ```
  _site/
  .leafpress/
  ```
- Create sample `index.md` if no markdown exists
- Detect existing config and warn

**Acceptance Criteria:**
- [ ] Running in empty folder creates starter project
- [ ] Running in existing garden just adds config
- [ ] No overwriting of existing files

### 6.2 New Command

**Tasks:**
- `leafpress new <name>` creates page with frontmatter:
  ```markdown
  ---
  title: "Name"
  date: 2025-01-15
  tags: []
  draft: true
  growth: "seedling"
  ---
  
  ```
- Support nested paths: `leafpress new projects/new-idea`
- Slugify name for filename
- Open in `$EDITOR` if set (optional enhancement)

**Acceptance Criteria:**
- [ ] File created with correct frontmatter
- [ ] Nested directories created as needed
- [ ] Existing file not overwritten

---

## Phase 7: Caching & Optimization

**Goal:** Enable fast incremental builds.

### 7.1 Content Hash Cache

**Tasks:**
- Store file hashes in `.leafpress/cache.json`:
  ```json
  {
    "version": 1,
    "files": {
      "path/to/file.md": {
        "hash": "sha256:...",
        "lastBuild": "2025-01-15T10:30:00Z",
        "outLinks": ["target1", "target2"]
      }
    }
  }
  ```
- Skip processing unchanged files
- Invalidate if any inbound link changes
- Clear cache on `--force` flag

**Acceptance Criteria:**
- [ ] Unchanged files skip parsing
- [ ] Link changes trigger re-render of affected pages
- [ ] Sub-second rebuilds for single file changes

### 7.2 Parallel Processing

**Tasks:**
- Parse pages concurrently (worker pool)
- Render templates concurrently
- Write files concurrently
- Respect system CPU count

**Acceptance Criteria:**
- [ ] Build time scales with CPU cores
- [ ] No race conditions
- [ ] Memory usage reasonable

---

## Phase 8: Error Handling & Polish

**Goal:** Production-ready error handling and user experience.

### 8.1 Error Messages

**Tasks:**
- Broken wiki-links: Warning with source location
- Invalid frontmatter: Error with file path and line
- Duplicate slugs: List all conflicts
- Missing config: Use defaults, inform user
- Port in use: Auto-increment, notify

**Acceptance Criteria:**
- [ ] All errors include file path
- [ ] Warnings don't stop build
- [ ] Errors stop build with clear message

### 8.2 Verbose Logging

**Tasks:**
- Implement log levels: error, warn, info, debug
- `--verbose` enables debug output
- Log file processing, timing, link resolution
- Colorized output (detect TTY)

**Acceptance Criteria:**
- [ ] Default output is clean and minimal
- [ ] Verbose shows detailed progress
- [ ] Colors disabled for non-TTY

### 8.3 Validation

**Tasks:**
- Validate config schema on load
- Validate frontmatter values
- Check for circular links (optional warning)
- Validate output directory is writable

**Acceptance Criteria:**
- [ ] Invalid config caught early
- [ ] Helpful suggestions for fixes

---

## Phase 9: Testing

**Goal:** Comprehensive test coverage.

### 9.1 Unit Tests

**Tasks:**
- Config parsing tests
- Frontmatter parsing tests
- Wiki-link extraction tests
- Link resolution tests
- Slug generation tests
- Template rendering tests

**Acceptance Criteria:**
- [ ] >80% code coverage on core packages
- [ ] Edge cases covered

### 9.2 Integration Tests

**Tasks:**
- Full build pipeline test
- Test garden fixture with various content
- Verify output structure
- Verify link integrity

**Acceptance Criteria:**
- [ ] End-to-end build works
- [ ] Output matches expected structure

### 9.3 CLI Tests

**Tasks:**
- Test all commands
- Test flag combinations
- Test error conditions

**Acceptance Criteria:**
- [ ] All commands testable
- [ ] Exit codes correct

---

## Dependency Summary

| Package | Purpose | Phase |
|---------|---------|-------|
| `github.com/spf13/cobra` | CLI framework | 1 |
| `gopkg.in/yaml.v3` | Frontmatter parsing | 2 |
| `github.com/yuin/goldmark` | Markdown rendering | 3 |
| `github.com/fsnotify/fsnotify` | File watching | 5 |
| `github.com/gorilla/websocket` | Live reload | 5 |

---

## Milestones

| Milestone | Phases | Deliverable |
|-----------|--------|-------------|
| **M1: Skeleton** | 1 | CLI with stub commands |
| **M2: Parser** | 2 | Content parsing complete |
| **M3: Builder** | 3, 4 | Static site generation works |
| **M4: Dev Mode** | 5 | Live reload functional |
| **M5: Complete** | 6, 7, 8 | All features, caching, polish |
| **M6: Tested** | 9 | Full test coverage |

---

## Next Steps

1. Review and approve this plan
2. Begin Phase 1 implementation
3. Iterate through milestones with regular check-ins
