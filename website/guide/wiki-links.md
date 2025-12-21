---
title: "Wiki Links"
date: 2025-12-21
---

Wiki-style links are the heart of leafpress, enabling you to create interconnected digital gardens.

## Basic Syntax

Link to other pages using double brackets:

```markdown
[[page-slug]]
```

The page slug is the filename without `.md`:
- `my-note.md` → `[[my-note]]`
- `projects/website.md` → `[[projects/website]]`

## Custom Display Text

Override the link text:

```markdown
[[page-slug|Custom Text]]
```

Example:
```markdown
Learn more about [[guide/configuration|configuration options]].
```

## How It Works

When you use `[[page-name]]`:

1. leafpress finds the matching `.md` file
2. Creates an internal link to that page
3. Uses the page's title as link text (if not specified)
4. Tracks the connection for backlinks

## Broken Links

If a linked page doesn't exist, leafpress will:
- Show a warning during build
- Style the link differently (strikethrough)
- List broken links in build output

## Backlinks

Every page automatically shows backlinks - other pages that link to it.

This creates a bidirectional network of your content, making it easy to discover related pages.

## Best Practices

### Use Descriptive Slugs

Good:
```markdown
[[react-hooks-guide|React Hooks Guide]]
```

Less good:
```markdown
[[page1|Guide]]
```

### Link Often

Don't be afraid to link frequently. Over-linking is better than under-linking in a digital garden.

### Create Index Pages

For sections, create index pages that link to related content:

```markdown
---
title: "JavaScript Notes"
---

# JavaScript Notes

- [[js/async-await]]
- [[js/promises]]
- [[js/event-loop]]
```

## External Links

Use standard Markdown for external links:

```markdown
[GitHub](https://github.com)
```

## Next Steps

- [[guide/writing|Writing Content]]
- [[guide/custom-styles|Customize Your Site]]
- [[features|Explore Features]]
