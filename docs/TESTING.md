# LeafPress Testing Guide

This document outlines the testing strategy for LeafPress to ensure all features work correctly.

## Manual Testing Checklist

### Core Functionality

#### Build Process
- [ ] `leafpress init` creates all required files
- [ ] `leafpress build` completes without errors
- [ ] `leafpress serve` starts dev server successfully
- [ ] Changes trigger rebuild in serve mode
- [ ] Draft pages are excluded from build
- [ ] Draft pages are included with `--drafts` flag

#### Content Processing
- [ ] Markdown renders correctly (headings, lists, code blocks, etc.)
- [ ] Frontmatter is parsed correctly
- [ ] Invalid frontmatter shows helpful error
- [ ] Pages without frontmatter work
- [ ] Special characters in content render correctly
- [ ] Code syntax highlighting works
- [ ] HTML entities are handled correctly

#### Wiki Links
- [ ] `[[page-name]]` creates working links
- [ ] `[[page-name|Custom Text]]` uses custom text
- [ ] Wiki links in code blocks are NOT processed
- [ ] Wiki links in inline code (backticks) are NOT processed
- [ ] Broken links are styled differently
- [ ] Broken links generate warnings during build
- [ ] Ambiguous links generate warnings
- [ ] Case-insensitive matching works
- [ ] Cross-directory linking works

#### Backlinks
- [ ] Backlinks appear on linked pages
- [ ] Backlinks show correct page titles
- [ ] Backlinks are bidirectional
- [ ] Backlinks section appears only when there are backlinks

#### Table of Contents
- [ ] TOC appears on pages with h2/h3 headings
- [ ] TOC is hidden on mobile (< 1280px)
- [ ] TOC is sticky on desktop
- [ ] TOC links scroll to correct positions (accounting for nav)
- [ ] Heading IDs are generated correctly
- [ ] Duplicate heading IDs are handled
- [ ] Special characters in headings are handled
- [ ] Emojis in headings don't break IDs
- [ ] `toc: false` in frontmatter disables TOC
- [ ] Site-wide `toc: false` disables TOC globally

#### Section Indexes
- [ ] `_index.md` creates section index page
- [ ] Section intro content is displayed
- [ ] Page list is shown by default
- [ ] `showList: false` hides page list
- [ ] Section pages are sorted correctly (date/title)
- [ ] `sort` field in frontmatter works

#### Tags
- [ ] Tag pages are generated
- [ ] Tags index shows all tags
- [ ] Tag cloud shows correct counts
- [ ] Tag pages list correct pages
- [ ] Tags are case-sensitive

#### Navigation
- [ ] Site title links to homepage
- [ ] Nav links work correctly
- [ ] Nav links must start with `/` (validation)
- [ ] Mobile nav is responsive
- [ ] Sticky nav works when enabled
- [ ] Sticky nav background is correct (handles gradients)

#### Theme & Styling
- [ ] Custom fonts load correctly
- [ ] Accent color is applied everywhere
- [ ] Background colors work (solid)
- [ ] Background gradients work (light)
- [ ] Background gradients work (dark)
- [ ] Single background color only applies to light mode
- [ ] Dark mode toggle works
- [ ] Dark mode preference is saved
- [ ] Custom `style.css` is used if present
- [ ] CSS variables work

#### Growth Stages
- [ ] Seedling emoji (🌱) appears correctly
- [ ] Budding emoji (🌿) appears correctly
- [ ] Evergreen emoji (🌳) appears correctly
- [ ] Invalid growth value shows error

#### Static Files
- [ ] Files in `static/` are copied to output
- [ ] Images are accessible
- [ ] Favicon files are included

### Edge Cases

#### File Handling
- [ ] Empty markdown files work
- [ ] Very long content works
- [ ] Files with special characters in names
- [ ] Nested directory structures
- [ ] Files with same names in different directories

#### Content Edge Cases
- [ ] Empty headings
- [ ] Very long headings
- [ ] Headings with HTML entities
- [ ] Multiple consecutive blank lines
- [ ] Markdown inside HTML
- [ ] Tables render correctly
- [ ] Blockquotes render correctly
- [ ] Nested lists work

#### Configuration
- [ ] Missing config file uses defaults
- [ ] Invalid JSON shows error
- [ ] Invalid port number shows error
- [ ] Dangerous output paths are rejected
- [ ] Invalid hex colors show error
- [ ] Invalid background CSS shows error

### Performance

- [ ] Build completes in reasonable time (< 1s for small sites)
- [ ] Large sites (100+ pages) build successfully
- [ ] Memory usage is reasonable
- [ ] No memory leaks in serve mode

### Browser Compatibility

- [ ] Works in Chrome/Edge
- [ ] Works in Firefox
- [ ] Works in Safari
- [ ] Mobile browsers work correctly
- [ ] Dark mode works across browsers

## Automated Test Suite

### Unit Tests (Future)

```
internal/content/
  - wikilink_test.go      # Wiki link parsing and resolution
  - frontmatter_test.go   # Frontmatter parsing
  - renderer_test.go      # Markdown rendering
  - scanner_test.go       # File scanning

internal/config/
  - config_test.go        # Config validation

internal/templates/
  - templates_test.go     # Template rendering
  - toc_test.go          # TOC extraction
```

### Integration Tests (Future)

```
tests/integration/
  - build_test.go         # Full build process
  - wikilinks_test.go     # Wiki link resolution across pages
  - sections_test.go      # Section index generation
  - tags_test.go          # Tag page generation
```

### Test Sites

Create test sites in `testdata/` for various scenarios:

1. **testdata/minimal/** - Bare minimum site
2. **testdata/features/** - All features enabled
3. **testdata/wikilinks/** - Complex wiki-link patterns
4. **testdata/sections/** - Multiple sections with indexes
5. **testdata/edge-cases/** - Edge cases and special characters

## Regression Testing

When fixing bugs, add:
1. Test case that reproduces the bug
2. Verify fix resolves the issue
3. Add to automated test suite to prevent regression

## Pre-Release Checklist

Before releasing a new version:

- [ ] All manual tests pass
- [ ] All automated tests pass
- [ ] Build succeeds on Linux, macOS, Windows
- [ ] Documentation is updated
- [ ] CHANGELOG is updated
- [ ] Example sites build successfully
- [ ] Performance is acceptable

## Reporting Issues

When reporting bugs, include:
1. LeafPress version
2. Operating system
3. Steps to reproduce
4. Expected vs actual behavior
5. Minimal example site (if applicable)
6. Build output/errors
