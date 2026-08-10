#!/usr/bin/env bash

# Reproducible clean-build benchmark suite for static site generators.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/workload.sh
source "${SCRIPT_DIR}/lib/workload.sh"

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE=${BENCHMARK_RESULTS_FILE:-"${SCRIPT_DIR}/results/BENCHMARK_${TIMESTAMP}.md"}
RUNS=${BENCHMARK_RUNS:-10}
WARMUPS=${BENCHMARK_WARMUPS:-2}
read -r -a PAGE_COUNTS <<< "${BENCHMARK_PAGE_COUNTS:-100 1000 2000}"
read -r -a SSGS <<< "${BENCHMARK_SSGS:-zola hugo leafpress-minimal leafpress eleventy jekyll}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

mkdir -p "$(dirname "$OUTPUT_FILE")"
WORKDIR=$(mktemp -d "${TMPDIR:-/tmp}/leafpress-benchmark.XXXXXX")
SIZE_ROWS="${WORKDIR}/output-sizes.md"

cleanup() {
    if [[ -n ${WORKDIR:-} && -d $WORKDIR && $(basename "$WORKDIR") == leafpress-benchmark.* ]]; then
        rm -rf -- "$WORKDIR"
    fi
}
trap cleanup EXIT

calc_stats() {
    local -a times=("$@")
    local -a valid_times=()
    local t

    for t in "${times[@]}"; do
        if [[ $t =~ ^[0-9]+$ ]]; then
            valid_times+=("$t")
        fi
    done

    local count=${#valid_times[@]}
    if ((count == 0)); then
        echo "0 0 0 0 0"
        return
    fi

    local sum=0
    local min=${valid_times[0]}
    local max=${valid_times[0]}
    for t in "${valid_times[@]}"; do
        sum=$((sum + t))
        ((t < min)) && min=$t
        ((t > max)) && max=$t
    done

    local mean=$((sum / count))
    local sum_sq_diff=0
    local diff
    for t in "${valid_times[@]}"; do
        diff=$((t - mean))
        sum_sq_diff=$((sum_sq_diff + diff * diff))
    done

    local variance=$((sum_sq_diff / count))
    local stddev=0
    if ((variance > 0)); then
        stddev=$(awk "BEGIN {printf \"%.0f\", sqrt($variance)}")
    fi

    local -a sorted
    IFS=$'\n' sorted=($(printf '%s\n' "${valid_times[@]}" | sort -n)); unset IFS
    local mid=$((count / 2))
    local p50
    if ((count % 2 == 0)); then
        p50=$(((sorted[mid - 1] + sorted[mid]) / 2))
    else
        p50=${sorted[mid]}
    fi

    echo "$mean $stddev $p50 $min $max"
}

check_ssg() {
    case $1 in
        leafpress|leafpress-minimal) [[ -x ${LEAFPRESS_BIN:-} || -x "${SCRIPT_DIR}/leafpress" || -x /benchmark/leafpress ]] ;;
        hugo) command -v hugo &>/dev/null ;;
        zola) command -v zola &>/dev/null ;;
        eleventy) command -v eleventy &>/dev/null ;;
        jekyll) command -v jekyll &>/dev/null ;;
        astro) command -v npm &>/dev/null ;;
        *) return 1 ;;
    esac
}

output_dir_for() {
    case $1 in
        hugo) echo "public" ;;
        zola) echo "public" ;;
        astro) echo "dist" ;;
        *) echo "_site" ;;
    esac
}

output_size() {
    local directory=$1
    if [[ ! -d $directory ]]; then
        echo "N/A"
        return
    fi
    local kib
    kib=$(du -sk "$directory" | awk '{print $1}')
    awk -v kib="$kib" 'BEGIN { if (kib >= 1048576) printf "%.2f GiB", kib / 1048576; else printf "%.1f MiB", kib / 1024 }'
}

get_cpu_info() {
    if [[ $(uname -s) == Darwin ]]; then
        sysctl -n machdep.cpu.brand_string 2>/dev/null || echo "Unknown"
    else
        local cpu
        cpu=$(awk -F: '/model name/ {gsub(/^[ \t]+/, "", $2); print $2; exit}' /proc/cpuinfo 2>/dev/null || true)
        [[ -n $cpu ]] || cpu=$(lscpu 2>/dev/null | awk -F: '/Model name/ {gsub(/^[ \t]+/, "", $2); print $2; exit}')
        echo "${cpu:-Unknown}"
    fi
}

get_memory_info() {
    if [[ $(uname -s) == Darwin ]]; then
        local mem_bytes
        mem_bytes=$(sysctl -n hw.memsize 2>/dev/null)
        echo "$((mem_bytes / 1024 / 1024 / 1024))GB"
    else
        local mem_kb
        mem_kb=$(awk '/MemTotal/ {print $2}' /proc/meminfo 2>/dev/null)
        echo "$((mem_kb / 1024 / 1024))GB"
    fi
}

source_revision() {
    local repo="${SCRIPT_DIR}/.."
    [[ -d /leafpress-src/.git ]] && repo=/leafpress-src
    if ! git -C "$repo" rev-parse --git-dir &>/dev/null; then
        echo "unknown"
        return
    fi
    local revision
    revision=$(git -C "$repo" rev-parse --short=12 HEAD)
    if [[ -n $(git -C "$repo" status --porcelain --untracked-files=no) ]]; then
        revision="${revision}-dirty"
    fi
    echo "$revision"
}

CPU_INFO=$(get_cpu_info)
MEM_INFO=$(get_memory_info)
SOURCE_REVISION=$(source_revision)

{
    echo "# SSG Benchmark Results"
    echo
    echo "**Date**: $(date)"
    echo "**System**: $(uname -s) $(uname -m)"
    echo "**CPU**: ${CPU_INFO}"
    echo "**Memory**: ${MEM_INFO}"
    echo "**Source revision**: \`${SOURCE_REVISION}\`"
    echo "**Workload**: v${WORKLOAD_VERSION} (hierarchical-notes-posts)"
    echo "**Warmups / measured runs**: ${WARMUPS} / ${RUNS}"
    echo
    echo "## Build Times (ms)"
    echo
    echo "*Format: P50 (mean ± stddev)*"
    echo
    printf '| SSG |'
    for count in "${PAGE_COUNTS[@]}"; do printf ' %s pages |' "$count"; done
    echo
    printf '|-----|'
    for _ in "${PAGE_COUNTS[@]}"; do printf '%s' '------------|'; done
    echo
} > "$OUTPUT_FILE"

{
    printf '| SSG |'
    for count in "${PAGE_COUNTS[@]}"; do printf ' %s pages |' "$count"; done
    echo
    printf '|-----|'
    for _ in "${PAGE_COUNTS[@]}"; do printf '%s' '------------|'; done
    echo
} > "$SIZE_ROWS"

echo -e "${YELLOW}SSG Benchmark Suite — workload v${WORKLOAD_VERSION}${NC}"
echo "SSGs: ${SSGS[*]}"
echo "Page counts: ${PAGE_COUNTS[*]}"
echo "Warmups / measured runs: $WARMUPS / $RUNS"
echo "Source revision: $SOURCE_REVISION"
echo "Output: $OUTPUT_FILE"
echo

for ssg in "${SSGS[@]}"; do
    echo -e "${YELLOW}Testing $ssg...${NC}"
    timing_row="| $ssg |"
    size_row="| $ssg |"

    if ! check_ssg "$ssg"; then
        echo -e "${RED}  $ssg not found, skipping${NC}"
        for _ in "${PAGE_COUNTS[@]}"; do
            timing_row="${timing_row} N/A |"
            size_row="${size_row} N/A |"
        done
        echo "$timing_row" >> "$OUTPUT_FILE"
        echo "$size_row" >> "$SIZE_ROWS"
        continue
    fi

    for count in "${PAGE_COUNTS[@]}"; do
        echo "  ${count} pages..."
        TEST_DIR="${WORKDIR}/${ssg}_${count}"
        mkdir -p "$TEST_DIR"

        generator="${SCRIPT_DIR}/generators/${ssg}/generate.sh"
        builder="${SCRIPT_DIR}/generators/${ssg}/build.sh"
        if [[ ! -f $generator || ! -f $builder ]] || ! bash "$generator" "$count" "$TEST_DIR" >/dev/null; then
            echo -e "${RED}    generation failed${NC}"
            timing_row="${timing_row} N/A |"
            size_row="${size_row} N/A |"
            rm -rf -- "$TEST_DIR"
            continue
        fi

        build_failed=false
        for warmup in $(seq 1 "$WARMUPS"); do
            if ! bash "$builder" "$TEST_DIR" >/dev/null; then
                echo -e "${RED}    warmup $warmup failed${NC}"
                build_failed=true
                break
            fi
        done

        times=()
        if [[ $build_failed == false ]]; then
            for run in $(seq 1 "$RUNS"); do
                if time_ms=$(bash "$builder" "$TEST_DIR"); then
                    times+=("$time_ms")
                    echo "    Run $run: ${time_ms}ms"
                else
                    echo -e "${RED}    Run $run: failed${NC}"
                fi
            done
        fi

        if ((${#times[@]} == RUNS)); then
            read -r mean stddev p50 min max <<< "$(calc_stats "${times[@]}")"
            timing_row="${timing_row} ${p50} (${mean}±${stddev}) |"
            echo -e "${GREEN}    P50: ${p50}ms, mean: ${mean}ms ± ${stddev}ms (range: ${min}-${max}ms)${NC}"
        else
            timing_row="${timing_row} N/A |"
        fi

        generated_output="${TEST_DIR}/$(output_dir_for "$ssg")"
        size_row="${size_row} $(output_size "$generated_output") |"
        rm -rf -- "$TEST_DIR"
    done

    echo "$timing_row" >> "$OUTPUT_FILE"
    echo "$size_row" >> "$SIZE_ROWS"
    echo
done

{
    echo
    echo "*leafpress-minimal: reader features disabled; standard Markdown links remain.*"
    echo "*leafpress: default reader features including wikilinks, backlinks, graph, search, and TOC.*"
    echo
    echo "## Generated Output Size"
    echo
    cat "$SIZE_ROWS"
    echo
    echo "## Methodology"
    echo
    echo "- **Clean builds**: Each output directory is removed before every build."
    echo "- **Warmups**: ${WARMUPS} unmeasured clean builds precede ${RUNS} measured clean builds."
    echo "- **Workload v${WORKLOAD_VERSION}**: Deterministic 70/30 split across \`notes/\` and \`posts/\`, with section homes."
    echo "- **Navigation**: Leafpress automatic navigation contains Notes, Posts, and Tags."
    echo "- **Content**: 1–5 deterministic paragraphs; code blocks on approximately 40% of pages."
    echo "- **Links**: Approximately 15% orphan pages; other pages have 2–8 links with deterministic hub bias and cross-section targets."
    echo "- **Tags**: Two of 20 deterministic tags per page."
    echo "- **Timing**: One Python monotonic-clock wrapper executes each build; interpreter startup is outside the timed interval."
    echo "- **Page count**: Counts benchmark notes/posts; home and two section pages are additional."
    echo
    echo "The intentionally pathological flat-root automatic-navigation workload is run separately with \`./run.sh stress\`."
    echo
    echo "## SSG Versions"
    echo
    echo "| SSG | Version |"
    echo "|-----|---------|"
} >> "$OUTPUT_FILE"

command -v hugo &>/dev/null && echo "| Hugo | $(hugo version 2>/dev/null | grep -oE 'v[0-9.]+' | head -1) |" >> "$OUTPUT_FILE"
command -v zola &>/dev/null && echo "| Zola | $(zola --version 2>/dev/null | grep -oE '[0-9.]+') |" >> "$OUTPUT_FILE"
command -v eleventy &>/dev/null && echo "| Eleventy | $(eleventy --version 2>/dev/null) |" >> "$OUTPUT_FILE"
command -v jekyll &>/dev/null && echo "| Jekyll | $(jekyll --version 2>/dev/null | grep -oE '[0-9.]+' | head -1) |" >> "$OUTPUT_FILE"
if [[ -x "${SCRIPT_DIR}/leafpress" ]]; then
    leafpress_version=$("${SCRIPT_DIR}/leafpress" version 2>/dev/null | head -1 || true)
    echo "| Leafpress | ${leafpress_version:-local build} |" >> "$OUTPUT_FILE"
elif [[ -x /benchmark/leafpress ]]; then
    echo "| Leafpress | local build |" >> "$OUTPUT_FILE"
fi

echo
echo -e "${GREEN}Benchmark complete: ${OUTPUT_FILE}${NC}"
cat "$OUTPUT_FILE"
