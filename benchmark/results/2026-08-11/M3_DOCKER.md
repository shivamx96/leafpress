# SSG Benchmark Results

**Date**: 2026-08-10T20:54:30Z
**System**: Linux aarch64
**CPU**: Unknown
**Memory**: 11GB
**Host**: Apple M3, 24GB
**Container limits**: 4 CPUs, 8GB
**Source revision**: `0e7bdb91397f`
**Leafpress binary**: `/benchmark/leafpress`
**Leafpress SHA-256**: `faa4e410556ac0ab15228cc8575881045c64d4ea372c45b5fad18acb1973ca94`
**Workload**: v2 (hierarchical-notes-posts)
**Warmups / measured runs**: 2 / 10
**Scheduling**: deterministic interleaved rotation

## Build Times (ms)

*Format: P50 (mean ± stddev)*

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|------------|------------|------------|
| zola | 21 (20±1) | 75 (75±4) | 138 (139±6) |
| hugo | 36 (36±2) | 124 (124±4) | 219 (221±9) |
| leafpress-minimal | 23 (22±1) | 97 (98±7) | 171 (171±6) |
| leafpress | 25 (25±2) | 111 (111±5) | 203 (204±6) |
| eleventy | 258 (258±7) | 523 (525±8) | 805 (808±11) |
| jekyll | 155 (155±1) | 278 (280±5) | 419 (420±4) |

*leafpress-minimal: reader features disabled; standard Markdown links remain.*
*leafpress: default reader features including wikilinks, backlinks, graph, search, and TOC.*

## Generated Output (logical bytes and file count)

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|------------|------------|------------|
| zola | 149.6 KiB (127 files) | 1.4 MiB (1027 files) | 2.8 MiB (2027 files) |
| hugo | 321.8 KiB (129 files) | 3.0 MiB (1029 files) | 6.0 MiB (2029 files) |
| leafpress-minimal | 1.7 MiB (146 files) | 8.9 MiB (1046 files) | 16.9 MiB (2046 files) |
| leafpress | 2.3 MiB (147 files) | 13.9 MiB (1047 files) | 26.9 MiB (2047 files) |
| eleventy | 129.1 KiB (124 files) | 1.2 MiB (1024 files) | 2.5 MiB (2024 files) |
| jekyll | 147.7 KiB (124 files) | 1.4 MiB (1024 files) | 2.8 MiB (2024 files) |

## Methodology

- **Clean builds**: Each output directory is removed before every build.
- **Warmups**: 2 unmeasured clean builds precede 10 measured clean builds.
- **Scheduling**: Page-count and SSG order rotate deterministically between warmups and measured runs.
- **Workload v2**: Deterministic 70/30 split across `notes/` and `posts/`, with section homes.
- **Navigation**: Every adapter renders Notes, Posts, and Tags links; Leafpress derives them automatically.
- **Content**: 1–5 deterministic paragraphs; code blocks on approximately 40% of pages.
- **Links**: Approximately 15% orphan pages; other pages have 2–8 links with deterministic hub bias and cross-section targets.
- **Tags**: Two of 20 deterministic tags per page.
- **Timing**: One Python monotonic-clock wrapper executes each build; interpreter startup is outside the timed interval.
- **Output size**: Sum of logical file bytes plus file count; filesystem allocation is not reported.
- **Support pages**: Home, two section listings, one tag index, and 20 tag pages are additional to the requested note/post count.

The intentionally pathological flat-root automatic-navigation workload is run separately with `./run.sh stress`.

## SSG Versions

| SSG | Version |
|-----|---------|
| Hugo | v0.121.1 |
| Zola | 0.21.0 |
| Eleventy | 3.1.2 |
| Jekyll | 4.4.1 |
| Leafpress | leafpress v1.0.0-beta.17 |
