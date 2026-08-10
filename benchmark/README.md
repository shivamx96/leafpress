# Leafpress benchmarks

The headline benchmark uses deterministic workload **v2**. It models a small
digital garden rather than placing every note at the site root:

```text
index.md
notes/
  _index.md
  note-*.md
posts/
  _index.md
  post-*.md
tags/
  _index.md
  tag*.md
```

Seventy percent of generated pages are notes and thirty percent are posts.
Leafpress automatic navigation therefore renders **Notes**, **Posts**, and
**Tags** instead of one navigation item per generated page. Content length,
tags, code blocks, orphan pages, and cross-section links are deterministic and
shared by every SSG generator. Each adapter also renders the same two section
listings, tag index, twenty tag pages, and Notes/Posts/Tags navigation. File
counts remain visible because generator-specific global assets can still vary.

Results from workload v2 must not be compared directly with the older flat,
random workload. Each report records its workload version and source revision.

## Run

For the comparable container environment:

```sh
./benchmark/run.sh docker
```

Docker runs persist files written below `/benchmark/results/`. To give a
canonical report its final repository path, pass the path as seen inside the
container:

```sh
BENCHMARK_RESULTS_FILE=/benchmark/results/YYYY-MM-DD/MACHINE_DOCKER.md \
  ./benchmark/run.sh docker
```

To use locally installed SSGs:

```sh
./benchmark/run.sh local
```

The flat-root automatic-navigation scenario remains available as an explicit
Leafpress stress benchmark:

```sh
./benchmark/run.sh stress
```

The main report records build-time distributions, logical output bytes, file
counts, the exact Leafpress binary hash, and the deterministic rotated run
order. Fixture generation completes before warmups so it is outside the timed
interval and does not give the first measured generator a thermal advantage.
The stress report additionally records navigation links and average HTML bytes
per page so quadratic growth is visible.

Full runs are strict by default: a missing adapter or any failed warmup/build
aborts without replacing the requested result file. Set
`BENCHMARK_STRICT=false` only for exploratory partial runs.

For quick local checks, the harness accepts these environment variables:

| Variable | Default | Example |
|---|---|---|
| `BENCHMARK_RUNS` | `10` (`5` for stress) | `BENCHMARK_RUNS=3` |
| `BENCHMARK_WARMUPS` | `2` (`1` for stress) | `BENCHMARK_WARMUPS=1` |
| `BENCHMARK_PAGE_COUNTS` | `100 1000 2000` | `BENCHMARK_PAGE_COUNTS="20 100"` |
| `BENCHMARK_SSGS` | all comparison SSGs | `BENCHMARK_SSGS="leafpress-minimal leafpress"` |
| `BENCHMARK_STRICT` | `true` | `BENCHMARK_STRICT=false` |
| `BENCHMARK_RESULTS_FILE` | timestamped file in `results/` | `/tmp/leafpress-benchmark.md` |
| `LEAFPRESS_BIN` | `benchmark/leafpress` | `/path/to/leafpress` |

All `BENCHMARK_*` controls in this table are forwarded to Docker. `LEAFPRESS_BIN`
is local-only because Docker always builds `/benchmark/leafpress` from the
mounted checkout.

Run the fast workload and reporting checks with:

```sh
./benchmark/test.sh
```
