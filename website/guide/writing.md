---
title: "Writing Content"
date: 2025-12-21
---

Create pages using Markdown files with YAML frontmatter.

## Creating Pages

### Using the CLI

```bash
leafpress new "My First Note"
```

Creates `content/my-first-note.md` with frontmatter.

### Manually

Create any `.md` file in your content directory:

```markdown
---
title: "My First Note"
date: 2025-01-06
tags: [ideas, projects]
---

Your content here.
```

## Frontmatter

Required:
- `title` — Page title

Optional:
- `date` — Publication date (YYYY-MM-DD)
- `modified` — Last modified date
- `tags` — List of tags: `[tag1, tag2]`
- `growth` — Note maturity: `seedling`, `budding`, or `evergreen`
- `toc` — Override global TOC setting: `true` or `false`
- `description` — SEO meta description (auto-generated if omitted)
- `image` — OG image path for social sharing
- `draft` — Set `true` to exclude from build
- `readingTime` — Override calculated reading time (minutes)

## Markdown Features

### Standard Markdown

All CommonMark syntax works: headings, bold, italic, lists, links, images, code blocks.

### Wiki Links

Connect pages with double brackets:

```markdown
[[other-page]]
[[other-page|Custom text]]
[[folder/nested-page]]
```

See [[guide/wiki-links|Wiki Links]] for details.

### Callouts

Obsidian-compatible admonitions:

```markdown
> [!note]
> This is a note callout.

> [!warning] Custom Title
> Warning with a custom title.
```

Available types: `note`, `tip`, `warning`, `danger`, `info`, `example`, `quote`, `question`, `bug`, `success`, `failure`, `abstract`, `todo`

### Images

Standard markdown images:

```markdown
![Alt text](/static/images/photo.jpg)
```

Obsidian-style embeds also work:

```markdown
![[photo.jpg]]
![[photo.jpg|Alt text]]
```

Images in `static/images/` are copied to the output.

### Code Blocks

Fenced code blocks with syntax highlighting:

````markdown
```javascript
function hello() {
  console.log("Hello, world!");
}
```
````

Copy button appears on hover.

## Folders

Organize content in folders. Create `folder/_index.md` for section pages:

```
content/
├── index.md
├── projects/
│   ├── _index.md      # /projects/
│   ├── website.md     # /projects/website/
│   └── cli.md         # /projects/cli/
└── notes/
    └── ideas.md       # /notes/ideas/
```

Link to nested pages: `[[projects/website]]`

