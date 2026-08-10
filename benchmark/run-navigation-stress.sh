#!/usr/bin/env bash

# Measure the quadratic-risk case where every note is a root navigation item.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUNS=${BENCHMARK_RUNS:-5}
WARMUPS=${BENCHMARK_WARMUPS:-1}
read -r -a PAGE_COUNTS <<< "${BENCHMARK_PAGE_COUNTS:-100 1000 2000}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE=${BENCHMARK_RESULTS_FILE:-"${SCRIPT_DIR}/results/NAVIGATION_STRESS_${TIMESTAMP}.md"}
GENERATOR="${SCRIPT_DIR}/generators/leafpress-flat-navigation/generate.sh"
BUILDER="${SCRIPT_DIR}/generators/leafpress-flat-navigation/build.sh"

if [[ ! -x ${LEAFPRESS_BIN:-} && ! -x "${SCRIPT_DIR}/leafpress" && ! -x /benchmark/leafpress ]]; then
    echo "Leafpress benchmark binary not found; run this through ./run.sh stress." >&2
    exit 1
fi

mkdir -p "$(dirname "$OUTPUT_FILE")"
WORKDIR=$(mktemp -d "${TMPDIR:-/tmp}/leafpress-navigation-stress.XXXXXX")
cleanup() {
    if [[ -n ${WORKDIR:-} && -d $WORKDIR && $(basename "$WORKDIR") == leafpress-navigation-stress.* ]]; then
        rm -rf -- "$WORKDIR"
    fi
}
trap cleanup EXIT

median() {
    printf '%s\n' "$@" | sort -n | awk '{values[NR]=$1} END {if (NR % 2) print values[(NR+1)/2]; else printf "%.0f\n", (values[NR/2]+values[NR/2+1])/2}'
}

{
    echo "# Leafpress Flat-Navigation Stress Results"
    echo
    echo "**Date**: $(date)"
    echo "**System**: $(uname -s) $(uname -m)"
    echo "**Warmups / measured runs**: ${WARMUPS} / ${RUNS}"
    echo
    echo "| Root notes | P50 build | Output size | Nav links/page | Average HTML/page |"
    echo "|-----------:|----------:|------------:|---------------:|------------------:|"
} > "$OUTPUT_FILE"

for count in "${PAGE_COUNTS[@]}"; do
    test_dir="${WORKDIR}/${count}"
    bash "$GENERATOR" "$count" "$test_dir"

    for _ in $(seq 1 "$WARMUPS"); do bash "$BUILDER" "$test_dir" >/dev/null; done
    times=()
    for run in $(seq 1 "$RUNS"); do
        time_ms=$(bash "$BUILDER" "$test_dir")
        times+=("$time_ms")
        echo "${count} pages, run ${run}: ${time_ms}ms"
    done

    p50=$(median "${times[@]}")
    output_kib=$(du -sk "${test_dir}/_site" | awk '{print $1}')
    output_size=$(awk -v kib="$output_kib" 'BEGIN {if (kib >= 1048576) printf "%.2f GiB", kib/1048576; else printf "%.1f MiB", kib/1024}')
    sample="${test_dir}/_site/page-1/index.html"
    nav_links=$(grep -c '<a class="lp-nav-link' "$sample" || true)
    if [[ $(uname -s) == Darwin ]]; then
        read -r html_count html_bytes < <(find "${test_dir}/_site" -name '*.html' -type f -exec stat -f '%z' {} + | awk '{sum+=$1} END {print NR, sum}')
    else
        read -r html_count html_bytes < <(find "${test_dir}/_site" -name '*.html' -type f -printf '%s\n' | awk '{sum+=$1} END {print NR, sum}')
    fi
    average_html=$((html_bytes / html_count))

    echo "| $count | ${p50} ms | $output_size | $nav_links | ${average_html} B |" >> "$OUTPUT_FILE"
    rm -rf -- "$test_dir"
done

{
    echo
    echo "This is an intentional worst case, not the headline SSG workload: automatic navigation renders every root note into every page."
} >> "$OUTPUT_FILE"

echo "Stress benchmark complete: $OUTPUT_FILE"
cat "$OUTPUT_FILE"
