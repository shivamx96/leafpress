# Performance architecture

Leafpress keeps full builds and local rebuilds responsive through:

- parallel Markdown/page rendering and parallel tag/auto-index generation;
- a shared `LinkResolver` for rendering, backlinks, and graph generation;
- pre-indexed tag and section data;
- compiled regular expressions and parsed-template reuse;
- a two-phase `WalkDir` scanner that avoids unnecessary file-info calls; and
- incremental in-memory state during `leafpress serve`.

## Measuring changes

Historical wall-clock numbers are intentionally not kept here: they depend on
hardware, filesystem cache state, content shape, Go version, and the exact
commit. Performance changes should include a reproducible benchmark or command,
the fixture size, the tested commit, and before/after results in the relevant PR.

For local profiling, the CLI exposes:

```bash
leafpress build --cpuprofile cpu.out
leafpress build --memprofile mem.out
go tool pprof cpu.out
```

Run ordinary correctness and race tests alongside performance work; a faster
incremental build is not acceptable if cached navigation, links, tags, or
section listings become stale.
