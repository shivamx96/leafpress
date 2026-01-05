# SSG Benchmark Results

**Date**: Mon Jan  5 22:06:33 IST 2026
**System**: Darwin arm64
**CPU**: Apple M3
**Memory**: 24GB
**Runs per test**: 10

## Build Times (ms)

*Format: P50 (mean ± stddev)*

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|-----------|------------|------------|
| zola | 55 (59±14) | 172 (175±12) | 330 (346±36) |
| hugo | 142 (163±63) | 307 (306±7) | 494 (500±27) |
| leafpress | 51 (143±276) | 199 (197±7) | 347 (347±12) |
| eleventy | 246 (254±23) | 508 (508±7) | 816 (817±8) |
| jekyll | 268 (294±73) | 513 (557±116) | 776 (847±211) |

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
