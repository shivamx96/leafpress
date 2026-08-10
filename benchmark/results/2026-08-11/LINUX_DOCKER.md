# SSG Benchmark Results

**Date**: Mon Aug 10 19:45:51 UTC 2026
**System**: Linux x86_64
**CPU**: AMD Ryzen 7 9800X3D 8-Core Processor
**Memory**: 62GB
**Source revision**: `4f6527a9c464-dirty`
**Leafpress binary**: `/benchmark/leafpress`
**Leafpress SHA-256**: `4d900324dc434e7e4849ee08c408e046e0a827ef61d7f1746ae766a1ba3be4df`
**Workload**: v2 (hierarchical-notes-posts)
**Warmups / measured runs**: 2 / 10
**Scheduling**: deterministic interleaved rotation

## Build Times (ms)

*Format: P50 (mean ± stddev)*

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|------------|------------|------------|
| zola | 18 (17±1) | 69 (69±2) | 131 (131±1) |
| hugo | 27 (27±0) | 105 (105±2) | 195 (195±2) |
| leafpress-minimal | 18 (18±0) | 71 (71±1) | 132 (131±2) |
| leafpress | 19 (19±0) | 87 (86±1) | 162 (162±1) |
| eleventy | 228 (227±2) | 437 (437±6) | 643 (643±8) |
| jekyll | 157 (157±1) | 286 (286±1) | 429 (428±2) |

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
| Leafpress | leafpress dev |
