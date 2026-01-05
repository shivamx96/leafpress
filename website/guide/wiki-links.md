---
title: "Wiki Links"
date: 2025-12-21
---

Wiki-style links are the heart of a digital garden. Connect ideas, build a web of knowledge.

## Syntax

Link to any page using double brackets:

```markdown
[[page-slug]]
```

The slug is the filename without `.md`:
- `my-note.md` → `[[my-note]]`
- `projects/website.md` → `[[projects/website]]`

### Custom Display Text

```markdown
[[page-slug|Display text]]
```

Example: `[[installation|Get started]]` renders as "Get started" but links to the installation page.

### Case Insensitive

Links are case-insensitive:
- `[[My Note]]` finds `my-note.md`
- `[[MY-NOTE]]` finds `my-note.md`

### Ambiguous Links

If multiple pages match (e.g., `note.md` and `folder/note.md`), leafpress warns during build but links to the first match. Use the full path to be explicit: `[[folder/note]]`.

## Backlinks

Every page automatically shows "Referenced from" at the bottom—a list of all pages that link to it. No configuration needed.

Backlinks make your garden bidirectional. When you link A → B, readers of B can discover A.

## Broken Links

During build, leafpress warns about broken links:

```
Warning: broken link: [[nonexistent-page]]
```

Broken links render as plain text with a visual indicator, so readers know something's missing.

## Link Previews

Hover over any wiki-link to see a preview card with the target page's title and excerpt. Works on backlinks too.

## Graph View

Enable `"graph": true` in config to visualize all connections between pages. The current page is highlighted, and you can click any node to navigate.

