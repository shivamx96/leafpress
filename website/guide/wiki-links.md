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

Link with the filename without `.md`, the page's URL slug, or its title:
- `my-note.md` → `[[my-note]]`
- `projects/website.md` → `[[projects/website]]`
- `My Note.md` (published at `/My-Note/`) → `[[My Note]]` or `[[My-Note]]`
- A page with `slug: first-note` in its frontmatter → `[[first-note]]` or its filename

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

Broken links render with a subtle broken-link style, so readers know something's missing.

## Link Previews

Hover over any wiki-link to see a preview card with the target page's title and excerpt. Works on backlinks too.

## Graph View

Enable `features.graph` in config to visualize all connections between pages. The current page is highlighted, and you can click any node to navigate.

