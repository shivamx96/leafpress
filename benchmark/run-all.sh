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
STRICT=${BENCHMARK_STRICT:-true}
read -r -a PAGE_COUNTS <<< "${BENCHMARK_PAGE_COUNTS:-100 1000 2000}"
read -r -a SSGS <<< "${BENCHMARK_SSGS:-zola hugo leafpress-minimal leafpress eleventy jekyll}"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

mkdir -p "$(dirname "$OUTPUT_FILE")"
WORKDIR=$(mktemp -d "${TMPDIR:-/tmp}/leafpress-benchmark.XXXXXX")
REPORT_TMP="${WORKDIR}/report.md"
SIZE_ROWS="${WORKDIR}/output-sizes.md"
TIMES_DIR="${WORKDIR}/times"
mkdir -p "$TIMES_DIR"

cleanup() {
    if [[ -n ${WORKDIR:-} && -d $WORKDIR && $(basename "$WORKDIR") == leafpress-benchmark.* ]]; then
        rm -rf -- "$WORKDIR"
    fi
}
trap cleanup EXIT

fail() {
    echo -e "${RED}benchmark failed: $*${NC}" >&2
    exit 1
}

[[ $RUNS =~ ^[1-9][0-9]*$ ]] || fail "BENCHMARK_RUNS must be a positive integer"
[[ $WARMUPS =~ ^[0-9]+$ ]] || fail "BENCHMARK_WARMUPS must be a non-negative integer"
[[ $STRICT == true || $STRICT == false ]] || fail "BENCHMARK_STRICT must be true or false"
((${#PAGE_COUNTS[@]} > 0)) || fail "at least one page count is required"
((${#SSGS[@]} > 0)) || fail "at least one SSG is required"
for count in "${PAGE_COUNTS[@]}"; do
    [[ $count =~ ^[1-9][0-9]*$ ]] || fail "invalid page count: $count"
done

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

resolve_leafpress_binary() {
    local candidate
    if [[ -n ${LEAFPRESS_BIN:-} ]]; then
        if [[ -x $LEAFPRESS_BIN ]]; then
            LEAFPRESS_BIN="$(cd "$(dirname "$LEAFPRESS_BIN")" && pwd)/$(basename "$LEAFPRESS_BIN")"
            export LEAFPRESS_BIN
        fi
        return
    fi

    for candidate in "${SCRIPT_DIR}/leafpress" /benchmark/leafpress; do
        if [[ -n $candidate && -x $candidate ]]; then
            LEAFPRESS_BIN="$(cd "$(dirname "$candidate")" && pwd)/$(basename "$candidate")"
            export LEAFPRESS_BIN
            return
        fi
    done
}

output_dir_for() {
    case $1 in
        hugo) echo "public" ;;
        zola) echo "public" ;;
        astro) echo "dist" ;;
        *) echo "_site" ;;
    esac
}

output_metrics() {
    local directory=$1
    local IFS=$' \t\n'
    if [[ ! -d $directory ]]; then
        printf 'N/A\tN/A\n'
        return
    fi

    local bytes files
    if [[ $(uname -s) == Darwin ]]; then
        read -r files bytes < <(find "$directory" -type f -exec stat -f '%z' {} + | awk '{sum += $1} END {print NR + 0, sum + 0}')
    else
        read -r files bytes < <(find "$directory" -type f -printf '%s\n' | awk '{sum += $1} END {print NR + 0, sum + 0}')
    fi

    local size
    size=$(awk -v bytes="$bytes" 'BEGIN {
        if (bytes >= 1073741824) printf "%.2f GiB", bytes / 1073741824;
        else if (bytes >= 1048576) printf "%.1f MiB", bytes / 1048576;
        else if (bytes >= 1024) printf "%.1f KiB", bytes / 1024;
        else printf "%d B", bytes;
    }')
    printf '%s\t%s\n' "$size" "$files"
}

validate_generated_output() {
    local directory=$1
    local count=$2
    local expected_html=$((count + 24))
    local sample="${directory}/notes/note-1/index.html"

    [[ -f $sample ]] || return 1
    [[ -f ${directory}/posts/index.html ]] || return 1
    [[ -f ${directory}/tags/index.html ]] || return 1
    [[ -f ${directory}/tags/tag0/index.html ]] || return 1
    grep -Fq 'href="/notes/"' "$sample" || return 1
    grep -Fq 'href="/posts/"' "$sample" || return 1
    grep -Fq 'href="/tags/"' "$sample" || return 1

    local html_count
    html_count=$(find "$directory" -type f -name '*.html' | wc -l | tr -d ' ')
    ((html_count >= expected_html))
}

binary_sha256() {
    local binary=$1
    if command -v sha256sum &>/dev/null; then
        sha256sum "$binary" | awk '{print $1}'
    else
        shasum -a 256 "$binary" | awk '{print $1}'
    fi
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
resolve_leafpress_binary
LEAFPRESS_VERSION="N/A"
LEAFPRESS_SHA256="N/A"
if [[ -x ${LEAFPRESS_BIN:-} ]]; then
    LEAFPRESS_VERSION=$("$LEAFPRESS_BIN" version 2>/dev/null | head -1 || true)
    LEAFPRESS_VERSION=${LEAFPRESS_VERSION:-local-build}
    LEAFPRESS_SHA256=$(binary_sha256 "$LEAFPRESS_BIN")
fi

{
    echo "# SSG Benchmark Results"
    echo
    echo "**Date**: $(date -u +'%Y-%m-%dT%H:%M:%SZ')"
    echo "**System**: $(uname -s) $(uname -m)"
    echo "**CPU**: ${CPU_INFO}"
    echo "**Memory**: ${MEM_INFO}"
    echo "**Source revision**: \`${SOURCE_REVISION}\`"
    echo "**Leafpress binary**: \`${LEAFPRESS_BIN:-N/A}\`"
    echo "**Leafpress SHA-256**: \`${LEAFPRESS_SHA256}\`"
    echo "**Workload**: v${WORKLOAD_VERSION} (hierarchical-notes-posts)"
    echo "**Warmups / measured runs**: ${WARMUPS} / ${RUNS}"
    echo "**Scheduling**: deterministic interleaved rotation"
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
} > "$REPORT_TMP"

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
echo "Strict mode: $STRICT"
echo "Output: $OUTPUT_FILE"
echo

# Generate every fixture before timing. Warmups then absorb fixture-generation
# heat, while the rotated matrix keeps one SSG from always running first or last.
AVAILABLE_SSGS=()
for ssg in "${SSGS[@]}"; do
    generator="${SCRIPT_DIR}/generators/${ssg}/generate.sh"
    builder="${SCRIPT_DIR}/generators/${ssg}/build.sh"
    if ! check_ssg "$ssg" || [[ ! -f $generator || ! -f $builder ]]; then
        if [[ $STRICT == true ]]; then
            fail "$ssg is unavailable or has no benchmark adapter"
        fi
        echo -e "${RED}$ssg unavailable; recording N/A${NC}"
        touch "${WORKDIR}/unavailable_${ssg}"
        continue
    fi

    AVAILABLE_SSGS+=("$ssg")
    for count in "${PAGE_COUNTS[@]}"; do
        test_dir="${WORKDIR}/fixture_${ssg}_${count}"
        echo "Preparing $ssg, $count pages..."
        if ! bash "$generator" "$count" "$test_dir" >/dev/null; then
            if [[ $STRICT == true ]]; then
                fail "$ssg fixture generation failed at $count pages"
            fi
            touch "${WORKDIR}/failed_${ssg}_${count}"
        fi
    done
done

((${#AVAILABLE_SSGS[@]} > 0)) || fail "no benchmark generators are available"

for ((warmup = 0; warmup < WARMUPS; warmup++)); do
    echo -e "${YELLOW}Warmup $((warmup + 1))/$WARMUPS${NC}"
    for ((count_slot = 0; count_slot < ${#PAGE_COUNTS[@]}; count_slot++)); do
        count_index=$(((count_slot + warmup) % ${#PAGE_COUNTS[@]}))
        count=${PAGE_COUNTS[$count_index]}
        for ((ssg_slot = 0; ssg_slot < ${#AVAILABLE_SSGS[@]}; ssg_slot++)); do
            ssg_index=$(((ssg_slot + warmup + count_index) % ${#AVAILABLE_SSGS[@]}))
            ssg=${AVAILABLE_SSGS[$ssg_index]}
            [[ -f ${WORKDIR}/failed_${ssg}_${count} ]] && continue
            builder="${SCRIPT_DIR}/generators/${ssg}/build.sh"
            test_dir="${WORKDIR}/fixture_${ssg}_${count}"
            if ! bash "$builder" "$test_dir" >/dev/null; then
                if [[ $STRICT == true ]]; then
                    fail "$ssg warmup failed at $count pages"
                fi
                touch "${WORKDIR}/failed_${ssg}_${count}"
            fi
        done
    done
done

for ((run = 0; run < RUNS; run++)); do
    echo -e "${YELLOW}Measured rotation $((run + 1))/$RUNS${NC}"
    for ((count_slot = 0; count_slot < ${#PAGE_COUNTS[@]}; count_slot++)); do
        count_index=$(((count_slot + run) % ${#PAGE_COUNTS[@]}))
        count=${PAGE_COUNTS[$count_index]}
        for ((ssg_slot = 0; ssg_slot < ${#AVAILABLE_SSGS[@]}; ssg_slot++)); do
            ssg_index=$(((ssg_slot + run + count_index) % ${#AVAILABLE_SSGS[@]}))
            ssg=${AVAILABLE_SSGS[$ssg_index]}
            [[ -f ${WORKDIR}/failed_${ssg}_${count} ]] && continue
            builder="${SCRIPT_DIR}/generators/${ssg}/build.sh"
            test_dir="${WORKDIR}/fixture_${ssg}_${count}"
            if time_ms=$(bash "$builder" "$test_dir"); then
                echo "$time_ms" >> "${TIMES_DIR}/${ssg}_${count}"
                echo "  $ssg, $count pages: ${time_ms}ms"
            else
                if [[ $STRICT == true ]]; then
                    fail "$ssg measured run $((run + 1)) failed at $count pages"
                fi
                touch "${WORKDIR}/failed_${ssg}_${count}"
            fi
        done
    done
done

for ssg in "${SSGS[@]}"; do
    timing_row="| $ssg |"
    size_row="| $ssg |"
    for count in "${PAGE_COUNTS[@]}"; do
        times_file="${TIMES_DIR}/${ssg}_${count}"
        if [[ -f ${WORKDIR}/unavailable_${ssg} || -f ${WORKDIR}/failed_${ssg}_${count} || ! -f $times_file ]]; then
            timing_row="${timing_row} N/A |"
            size_row="${size_row} N/A |"
            continue
        fi

        times=()
        while IFS= read -r time_ms; do times+=("$time_ms"); done < "$times_file"
        if ((${#times[@]} != RUNS)); then
            [[ $STRICT == true ]] && fail "$ssg recorded ${#times[@]}/$RUNS runs at $count pages"
            timing_row="${timing_row} N/A |"
            size_row="${size_row} N/A |"
            continue
        fi

        read -r mean stddev p50 min max <<< "$(calc_stats "${times[@]}")"
        generated_output="${WORKDIR}/fixture_${ssg}_${count}/$(output_dir_for "$ssg")"
        if ! validate_generated_output "$generated_output" "$count"; then
            if [[ $STRICT == true ]]; then
                fail "$ssg output contract failed at $count pages"
            fi
            timing_row="${timing_row} N/A |"
            size_row="${size_row} N/A |"
            continue
        fi

        timing_row="${timing_row} ${p50} (${mean}±${stddev}) |"
        IFS=$'\t' read -r size file_count <<< "$(output_metrics "$generated_output")"
        size_row="${size_row} ${size} (${file_count} files) |"
        echo -e "${GREEN}$ssg, $count pages — P50 ${p50}ms, ${size} across ${file_count} files${NC}"
    done
    echo "$timing_row" >> "$REPORT_TMP"
    echo "$size_row" >> "$SIZE_ROWS"
done

{
    echo
    echo "*leafpress-minimal: reader features disabled; standard Markdown links remain.*"
    echo "*leafpress: default reader features including wikilinks, backlinks, graph, search, and TOC.*"
    echo
    echo "## Generated Output (logical bytes and file count)"
    echo
    cat "$SIZE_ROWS"
    echo
    echo "## Methodology"
    echo
    echo "- **Clean builds**: Each output directory is removed before every build."
    echo "- **Warmups**: ${WARMUPS} unmeasured clean builds precede ${RUNS} measured clean builds."
    echo "- **Scheduling**: Page-count and SSG order rotate deterministically between warmups and measured runs."
    echo "- **Workload v${WORKLOAD_VERSION}**: Deterministic 70/30 split across \`notes/\` and \`posts/\`, with section homes."
    echo "- **Navigation**: Every adapter renders Notes, Posts, and Tags links; Leafpress derives them automatically."
    echo "- **Content**: 1–5 deterministic paragraphs; code blocks on approximately 40% of pages."
    echo "- **Links**: Approximately 15% orphan pages; other pages have 2–8 links with deterministic hub bias and cross-section targets."
    echo "- **Tags**: Two of 20 deterministic tags per page."
    echo "- **Timing**: One Python monotonic-clock wrapper executes each build; interpreter startup is outside the timed interval."
    echo "- **Output size**: Sum of logical file bytes plus file count; filesystem allocation is not reported."
    echo "- **Support pages**: Home, two section listings, one tag index, and 20 tag pages are additional to the requested note/post count."
    echo
    echo "The intentionally pathological flat-root automatic-navigation workload is run separately with \`./run.sh stress\`."
    echo
    echo "## SSG Versions"
    echo
    echo "| SSG | Version |"
    echo "|-----|---------|"
} >> "$REPORT_TMP"

command -v hugo &>/dev/null && echo "| Hugo | $(hugo version 2>/dev/null | grep -oE 'v[0-9.]+' | head -1) |" >> "$REPORT_TMP"
command -v zola &>/dev/null && echo "| Zola | $(zola --version 2>/dev/null | grep -oE '[0-9.]+') |" >> "$REPORT_TMP"
command -v eleventy &>/dev/null && echo "| Eleventy | $(eleventy --version 2>/dev/null) |" >> "$REPORT_TMP"
command -v jekyll &>/dev/null && echo "| Jekyll | $(jekyll --version 2>/dev/null | grep -oE '[0-9.]+' | head -1) |" >> "$REPORT_TMP"
[[ -x ${LEAFPRESS_BIN:-} ]] && echo "| Leafpress | ${LEAFPRESS_VERSION} |" >> "$REPORT_TMP"

mv "$REPORT_TMP" "$OUTPUT_FILE"

echo
echo -e "${GREEN}Benchmark complete: ${OUTPUT_FILE}${NC}"
cat "$OUTPUT_FILE"
