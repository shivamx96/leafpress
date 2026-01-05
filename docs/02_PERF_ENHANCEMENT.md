# Performance enhancements
Date: Dec 27, 2025

- **Parallel page rendering** to render pages in parallel using worker threads
- **Parallel markdown rendering** to render MD files using goldmark via worker threads
- **Parallel tag page, auto-index generation** to generate pages in parallel using worker threads
- **Buffered I/O** for batching writes to disk instead of individually writing files
- **Unified LinkResolver instance** to avoid rebuilding slug <> name maps repeatedly
- **Pre-indexed tag data and section data** to reuse them when creating pages
- **Regex caching** to avoid recompiling regex patterns repeatedly
- **Template caching** to avoid recompiling templates repeatedly
- **WalkDir optimization** to avoid unnecessary syscalls by reading all files

## Impact:
- 1000 pages build time: 154ms -> 98ms (36%) vs 107ms for Hugo
- 2000 pages build time: 272ms -> 171ms (37%) vs 206ms for Hugo
