---
title: "Writing Content"
date: 2025-12-21
---

Create pages using Markdown files in your site's root directory.

## Creating Pages

### Using the CLI

```bash
leafpress new "My First Note"
```

This creates `my-first-note.md` with frontmatter.

### Manual Creation

Create a `.md` file with YAML frontmatter:

```markdown
---
title: "My Note"
date: 2025-12-21
tags: [personal, ideas]
growth: "budding"
---

# My Note

Your content here...
```

## Frontmatter

### Required Fields

- **title**: Page title

### Optional Fields

- **date**: Publication date (YYYY-MM-DD)
- **tags**: Array of tags for categorization
- **growth**: Page maturity (`"seedling"`, `"budding"`, `"evergreen"`)
- **draft**: Set to `true` to exclude from build
- **toc**: Override site-wide TOC setting (`true` or `false`)

## Markdown Features

leafpress supports standard Markdown plus:

### Wiki Links

Link to other pages using double brackets:

```markdown
[[page-name|Display Text]]
[[page-name]]  // Uses page title as text
```

Learn more in [[guide/wiki-links|Wiki Links Guide]].

### Headings

```markdown
# Heading 1
## Heading 2
### Heading 3
```

Headings automatically get anchor IDs for linking.

### Code Blocks

````markdown
```javascript
console.log("Hello, world!");
```
````

Supports syntax highlighting with automatic language detection.

### Lists

```markdown
- Unordered list
- Another item
  - Nested item

1. Ordered list
2. Second item
```

### Blockquotes

```markdown
> This is a quote
> Multiple lines
```

### Images

```markdown
![Alt text](/static/images/photo.jpg)
```

Place images in `static/images/` directory.

## Organization

### Sections

Create directories for different content sections:

```
content/
├── notes/
│   ├── index.md
│   └── my-note.md
├── projects/
│   ├── index.md
│   └── project-1.md
└── index.md
```

Each directory should have an `index.md` that lists its pages.

### Tags

Use tags to categorize content across sections:

```yaml
---
tags: [go, programming, web]
---
```

Tags automatically create listing pages at `/tags/`.

## Next Steps

- [[guide/wiki-links|Master Wiki Links]]
- [[guide/deployment|Deploy Your Site]]
- [[guide/custom-styles|Customize Your Site]]
