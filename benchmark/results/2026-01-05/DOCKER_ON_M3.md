# SSG Benchmark Results

**Date**: Mon Jan  5 16:39:15 UTC 2026
**System**: Linux aarch64
**CPU**: Docker container (aarch64)
**Memory**: 11GB
**Runs per test**: 10

## Build Times (ms)

*Format: P50 (mean ± stddev)*

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|-----------|------------|------------|
| zola | 25 (25±3) | 76 (76±3) | 131 (130±2) |
| hugo | 39 (40±3) | 125 (125±4) | 222 (220±6) |
| leafpress | 28 (28±2) | 113 (112±3) | 207 (206±6) |
| eleventy | 241 (241±9) | 450 (453±10) | 674 (672±11) |
| jekyll | 159 (176±44) | 309 (348±115) | 481 (546±201) |

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
| Hugo | v0.121.1 |
| Zola | 0.21.0 |
| Eleventy | 3.1.2 |
| Jekyll | 4.4.1 |
| Leafpress | (local build) |
