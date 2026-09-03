---
title: "Features"
date: 2025-12-21
---

Everything you need to build a digital garden, nothing you don't.

## Content

**Wiki Links** — Connect pages with `[[page-name]]` or `[[page-name|custom text]]`. Links resolve automatically, broken links get flagged during build.

**Backlinks** — Every page shows what links to it. No configuration needed.

**Tags** — Add `tags: [idea, project]` to frontmatter or write `#idea` directly in your notes. Inline tags become links and feed the same automatically generated tag pages.

**Mermaid Diagrams** — Fenced `mermaid` code blocks render as interactive diagrams. Supports flowcharts, sequence diagrams, Gantt charts, and more. Dark mode compatible. Self-hosted (no CDN).

**Callouts** — Obsidian-compatible admonitions for notes, warnings, tips, and more.

> [!tip] Like this
> Use `> [!note]`, `> [!warning]`, `> [!tip]`, or `> [!danger]`

**Table of Contents** — Auto-generated from your headings. Toggle globally or per-page.

**Anchored Headings** — Hover over any heading to reveal a `#` link for deep linking and sharing.

**Footnotes** — Add references with `[^1]` and definitions with `[^1]: ...`. Renders as superscript links with a footnote section at the bottom.

**Growth Stages** — Track note maturity with `growth: seedling`, `budding`, or `evergreen`.

## Discovery

**Full-Text Search** — Fast client-side search. Works offline, no external services.

**Graph View** — Visualize connections between your notes. Interactive, zoomable.

**Link Previews** — Hover over any wiki-link to preview the target page.

## Media

**YouTube Embeds** — Paste a YouTube URL on its own line and it auto-embeds as a responsive iframe. Privacy-enhanced via youtube-nocookie.com.

**Video & Audio** — Embed local media with Obsidian syntax: `![[video.mp4]]` renders as a native video player, `![[recording.mp3]]` as an audio player.

**Image Width** — Control image size with Obsidian syntax: `![[photo.png|500]]` sets the width to 500px.

## Publishing

**Portable Publishing** — `leafpress build` writes plain static files to `_site/`.
Publish that directory with your hosting provider's own CLI, dashboard, or CI
integration without giving leafpress access to a hosting account.

**SEO Ready** — Automatic robots.txt, Open Graph tags, Twitter cards, plus sitemap.xml when `site.baseURL` is set.

**RSS Feed** — Auto-generated feed.xml with nav icon. Toggle with `features.rss` in config (requires `site.baseURL`).

**Custom 404** — Styled error page, ready for any hosting platform.

**Fast Builds** — Hundreds of pages in milliseconds. Parallel processing, minimal allocations.

**Live Reload** — Changes appear instantly during development.

## Theming

**Typography** — A curated 17-family catalog is self-hosted for headings, body, and code, and you can add your own font files under `static/fonts/`. Only selected families ship with the built site; no third-party font requests.

**Colors** — Set your accent color. Light and dark backgrounds with gradient support.

**Dark Mode** — Built-in toggle with system preference detection.

**Navigation** — Multiple styles (base, sticky, glassmorphic). Box or underlined active link indicators.

**Custom CSS** — Drop in a `style.css` to override anything.

## What's Not Included

leafpress is intentionally minimal. These are out of scope:

- **Comments** — Use Giscus, Utterances, or similar
- **Analytics** — Use Plausible, Fathom, or similar
- **CMS** — Edit markdown files directly
- **Image optimization** — Use a CDN or external tool for advanced needs (lazy loading is built-in)
- **Pagination** — Digital gardens are relational, not chronological
