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
REPORT_TMP="${WORKDIR}/report.md"
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
    echo "| Root notes | P50 build | Logical output | Nav links/page | Average HTML/page |"
    echo "|-----------:|----------:|------------:|---------------:|------------------:|"
} > "$REPORT_TMP"

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
    sample="${test_dir}/_site/page-1/index.html"
    nav_links=$(grep -c '<a class="lp-nav-link' "$sample" || true)
    if [[ $(uname -s) == Darwin ]]; then
        read -r html_count html_bytes < <(find "${test_dir}/_site" -name '*.html' -type f -exec stat -f '%z' {} + | awk '{sum+=$1} END {print NR, sum}')
        output_bytes=$(find "${test_dir}/_site" -type f -exec stat -f '%z' {} + | awk '{sum+=$1} END {print sum + 0}')
    else
        read -r html_count html_bytes < <(find "${test_dir}/_site" -name '*.html' -type f -printf '%s\n' | awk '{sum+=$1} END {print NR, sum}')
        output_bytes=$(find "${test_dir}/_site" -type f -printf '%s\n' | awk '{sum+=$1} END {print sum + 0}')
    fi
    output_size=$(awk -v bytes="$output_bytes" 'BEGIN {if (bytes >= 1073741824) printf "%.2f GiB", bytes/1073741824; else if (bytes >= 1048576) printf "%.1f MiB", bytes/1048576; else printf "%.1f KiB", bytes/1024}')
    average_html=$((html_bytes / html_count))

    echo "| $count | ${p50} ms | $output_size | $nav_links | ${average_html} B |" >> "$REPORT_TMP"
    rm -rf -- "$test_dir"
done

{
    echo
    echo "This is an intentional worst case, not the headline SSG workload: automatic navigation renders every root note into every page."
} >> "$REPORT_TMP"

mv "$REPORT_TMP" "$OUTPUT_FILE"

echo "Stress benchmark complete: $OUTPUT_FILE"
cat "$OUTPUT_FILE"
