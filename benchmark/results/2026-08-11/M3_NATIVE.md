# SSG Benchmark Results

**Date**: 2026-08-10T20:45:32Z
**System**: Darwin arm64
**CPU**: Apple M3
**Memory**: 24GB
**Source revision**: `0e7bdb91397f`
**Leafpress binary**: `/tmp/leafpress-m3-benchmark.8c8cpc/leafpress-darwin/leafpress`
**Leafpress SHA-256**: `be009e16b24b323e57d00bc2ee7f3570f3b614e15fe28df8b454fdde0473a4b2`
**Workload**: v2 (hierarchical-notes-posts)
**Warmups / measured runs**: 2 / 10
**Scheduling**: deterministic interleaved rotation

## Build Times (ms)

*Format: P50 (mean ± stddev)*

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|------------|------------|------------|
| zola | 34 (34±0) | 159 (158±5) | 296 (298±7) |
| hugo | 46 (46±1) | 184 (188±11) | 334 (336±10) |
| leafpress-minimal | 35 (34±1) | 164 (165±5) | 312 (311±12) |
| leafpress | 36 (36±3) | 171 (171±3) | 321 (327±11) |
| eleventy | 262 (263±3) | 602 (604±10) | 962 (965±11) |
| jekyll | 207 (207±0) | 470 (472±7) | 774 (776±6) |

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
