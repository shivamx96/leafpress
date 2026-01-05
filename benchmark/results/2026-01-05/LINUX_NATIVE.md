# SSG Benchmark Results

**Date**: Mon Jan  5 10:11:18 PM IST 2026
**System**: Linux x86_64
**CPU**: AMD Ryzen 7 9800X3D 8-Core Processor
**Memory**: 62GB
**Runs per test**: 10

## Build Times (ms)

*Format: P50 (mean ± stddev)*

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|-----------|------------|------------|
| zola | 22 (24±8) | 50 (49±1) | 81 (83±5) |
| hugo | 38 (46±22) | 116 (122±13) | 203 (206±8) |
| leafpress | 22 (21±1) | 62 (62±2) | 97 (97±1) |
| eleventy | 206 (212±21) | 337 (337±5) | 487 (487±3) |
| jekyll | 157 (177±61) | 306 (344±117) | 469 (539±212) |

*leafpress: full features including wikilinks, backlinks, graph, and TOC.*

## Methodology

- **Clean Build**: Output directory removed before each build
- **Runs**: 10 iterations, reporting P50 (median), mean, and standard deviation
- **Content**: Each page has frontmatter, markdown content, code block, and internal links
- **Tags**: 20 unique tags distributed across pages
- **Links**: Each page links to 2 other pages
- **Timing**: External timing via `date +%s%3N` for consistency across all SSGs

## SSG Versions

| SSG | Version |
|-----|---------|
| Hugo | v0.153.3 |
| Zola | 0.21.0 |
| Eleventy | 3.1.2 |
| Jekyll | 4.4.1 |
