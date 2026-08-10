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
```

Seventy percent of generated pages are notes and thirty percent are posts.
Leafpress automatic navigation therefore renders **Notes**, **Posts**, and
**Tags** instead of one navigation item per generated page. Content length,
tags, code blocks, orphan pages, and cross-section links are deterministic and
shared by every SSG generator.

Results from workload v2 must not be compared directly with the older flat,
random workload. Each report records its workload version and source revision.

## Run

For the comparable container environment:

```sh
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

The main report records build-time distributions and generated output sizes.
The stress report additionally records navigation links and average HTML bytes
per page so quadratic growth is visible.

For quick local checks, the harness accepts these environment variables:

| Variable | Default | Example |
|---|---|---|
| `BENCHMARK_RUNS` | `10` (`5` for stress) | `BENCHMARK_RUNS=3` |
| `BENCHMARK_WARMUPS` | `2` (`1` for stress) | `BENCHMARK_WARMUPS=1` |
| `BENCHMARK_PAGE_COUNTS` | `100 1000 2000` | `BENCHMARK_PAGE_COUNTS="20 100"` |
| `BENCHMARK_SSGS` | all comparison SSGs | `BENCHMARK_SSGS="leafpress-minimal leafpress"` |
| `BENCHMARK_RESULTS_FILE` | timestamped file in `results/` | `/tmp/leafpress-benchmark.md` |
| `LEAFPRESS_BIN` | `benchmark/leafpress` | `/path/to/leafpress` |

Run the fast workload and reporting checks with:

```sh
./benchmark/test.sh
```
