#!/usr/bin/env bash

# SSG Benchmark Suite
# Runs clean build benchmarks for multiple static site generators

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE="${SCRIPT_DIR}/results/BENCHMARK_${TIMESTAMP}.md"
RUNS=10
PAGE_COUNTS=(100 1000 2000)
SSGS=(zola hugo leafpress eleventy jekyll)

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Create results directory
mkdir -p "${SCRIPT_DIR}/results"

# Working directory for test sites
WORKDIR=$(mktemp -d)
trap "rm -rf $WORKDIR" EXIT

echo -e "${YELLOW}======================================${NC}"
echo -e "${YELLOW}    SSG Benchmark Suite${NC}"
echo -e "${YELLOW}======================================${NC}"
echo ""
echo "SSGs: ${SSGS[*]}"
echo "Page counts: ${PAGE_COUNTS[*]}"
echo "Runs per test: $RUNS"
echo "Output: $OUTPUT_FILE"
echo ""

# Function to calculate stats (mean, stddev, p50/median, min, max)
calc_stats() {
    local -a times=("$@")
    local -a valid_times=()

    # Filter valid numeric values
    for t in "${times[@]}"; do
        if [[ -n "$t" && "$t" =~ ^[0-9]+$ ]]; then
            valid_times+=($t)
        fi
    done

    local count=${#valid_times[@]}
    if [ $count -eq 0 ]; then
        echo "0 0 0 0 0"
        return
    fi

    # Calculate sum, min, max
    local sum=0
    local min=${valid_times[0]}
    local max=${valid_times[0]}
    for t in "${valid_times[@]}"; do
        sum=$((sum + t))
        if [ "$t" -lt "$min" ]; then min=$t; fi
        if [ "$t" -gt "$max" ]; then max=$t; fi
    done

    # Calculate mean
    local mean=$((sum / count))

    # Calculate standard deviation
    local sum_sq_diff=0
    for t in "${valid_times[@]}"; do
        local diff=$((t - mean))
        sum_sq_diff=$((sum_sq_diff + diff * diff))
    done
    local variance=$((sum_sq_diff / count))
    # Integer square root approximation
    local stddev=0
    if [ $variance -gt 0 ]; then
        stddev=$(awk "BEGIN {printf \"%.0f\", sqrt($variance)}")
    fi

    # Calculate P50 (median) - sort and pick middle value
    IFS=$'\n' sorted=($(sort -n <<< "${valid_times[*]}")); unset IFS
    local mid=$((count / 2))
    local p50
    if [ $((count % 2)) -eq 0 ]; then
        # Even count: average of two middle values
        p50=$(( (sorted[mid-1] + sorted[mid]) / 2 ))
    else
        # Odd count: middle value
        p50=${sorted[mid]}
    fi

    echo "$mean $stddev $p50 $min $max"
}

# Check if SSG is available
check_ssg() {
    local ssg=$1
    case $ssg in
        leafpress) [ -f "${SCRIPT_DIR}/leafpress" ] || [ -f /benchmark/leafpress ] ;;
        hugo) command -v hugo &>/dev/null ;;
        zola) command -v zola &>/dev/null ;;
        eleventy) command -v eleventy &>/dev/null || command -v npx &>/dev/null ;;
        jekyll) command -v jekyll &>/dev/null ;;
    esac
}

# Gather system info
get_cpu_info() {
    if [ "$(uname -s)" == "Darwin" ]; then
        sysctl -n machdep.cpu.brand_string 2>/dev/null || echo "Unknown"
    else
        # Try model name first, then fall back to other methods
        cpu=$(grep -m1 "model name" /proc/cpuinfo 2>/dev/null | cut -d: -f2 | xargs)
        if [ -z "$cpu" ]; then
            # ARM/Docker fallback - try lscpu
            cpu=$(lscpu 2>/dev/null | grep "Model name" | cut -d: -f2 | xargs)
        fi
        if [ -z "$cpu" ]; then
            # Last resort - check if running in Docker
            if [ -f /.dockerenv ]; then
                cpu="Docker container ($(uname -m))"
            else
                cpu="Unknown"
            fi
        fi
        echo "$cpu"
    fi
}

get_memory_info() {
    if [ "$(uname -s)" == "Darwin" ]; then
        mem_bytes=$(sysctl -n hw.memsize 2>/dev/null)
        echo "$((mem_bytes / 1024 / 1024 / 1024))GB"
    else
        mem_kb=$(grep MemTotal /proc/meminfo 2>/dev/null | awk '{print $2}')
        echo "$((mem_kb / 1024 / 1024))GB"
    fi
}

CPU_INFO=$(get_cpu_info)
MEM_INFO=$(get_memory_info)

# Initialize results file
cat > "$OUTPUT_FILE" << EOF
# SSG Benchmark Results

**Date**: $(date)
**System**: $(uname -s) $(uname -m)
**CPU**: ${CPU_INFO}
**Memory**: ${MEM_INFO}
**Runs per test**: $RUNS

## Build Times (ms)

*Format: P50 (mean ± stddev)*

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|-----------|------------|------------|
EOF

for ssg in "${SSGS[@]}"; do
    echo -e "${YELLOW}Testing $ssg...${NC}"

    if ! check_ssg "$ssg"; then
        echo -e "${RED}  $ssg not found, skipping${NC}"
        echo "| $ssg | N/A | N/A | N/A |" >> "$OUTPUT_FILE"
        continue
    fi

    ssg_results="| $ssg |"

    for count in "${PAGE_COUNTS[@]}"; do
        echo -e "  ${count} pages..."

        # Create test directory
        TEST_DIR="$WORKDIR/${ssg}_${count}"
        mkdir -p "$TEST_DIR"

        # Generate content
        if [ -f "${SCRIPT_DIR}/generators/${ssg}/generate.sh" ]; then
            bash "${SCRIPT_DIR}/generators/${ssg}/generate.sh" "$count" "$TEST_DIR" 2>/dev/null
        else
            echo -e "${RED}    Generator not found${NC}"
            ssg_results="${ssg_results} N/A |"
            continue
        fi

        # Run builds
        times=()
        for run in $(seq 1 $RUNS); do
            if [ -f "${SCRIPT_DIR}/generators/${ssg}/build.sh" ]; then
                time_ms=$(bash "${SCRIPT_DIR}/generators/${ssg}/build.sh" "$TEST_DIR" 2>/dev/null)
                if [[ "$time_ms" =~ ^[0-9]+$ ]]; then
                    times+=($time_ms)
                    echo "    Run $run: ${time_ms}ms"
                else
                    echo "    Run $run: failed"
                fi
            fi
        done

        # Calculate stats
        if [ ${#times[@]} -gt 0 ]; then
            read mean stddev p50 min max <<< $(calc_stats "${times[@]}")
            ssg_results="${ssg_results} ${p50} (${mean}±${stddev}) |"
            echo -e "${GREEN}    P50: ${p50}ms, Mean: ${mean}ms ± ${stddev}ms (range: ${min}-${max}ms)${NC}"
        else
            ssg_results="${ssg_results} N/A |"
        fi

        # Cleanup test dir to save space
        rm -rf "$TEST_DIR"
    done

    echo "$ssg_results" >> "$OUTPUT_FILE"
    echo ""
done

# Add note about leafpress features
echo "" >> "$OUTPUT_FILE"
echo "*leafpress: full features including wikilinks, backlinks, graph, and TOC.*" >> "$OUTPUT_FILE"

# Add methodology section
cat >> "$OUTPUT_FILE" << 'EOF'

## Methodology

- **Clean Build**: Output directory removed before each build
- **Runs**: 10 iterations, reporting P50 (median), mean, and standard deviation
- **Content**: Each page has frontmatter, markdown content, code block, and internal links
- **Tags**: 20 unique tags distributed across pages
- **Links**: Each page links to 2 other pages
- **Timing**: External timing via `date +%s%3N` for consistency across all SSGs

## SSG Versions

EOF

# Add version info
echo "| SSG | Version |" >> "$OUTPUT_FILE"
echo "|-----|---------|" >> "$OUTPUT_FILE"

if command -v hugo &>/dev/null; then
    echo "| Hugo | $(hugo version 2>/dev/null | grep -oE 'v[0-9.]+' | head -1) |" >> "$OUTPUT_FILE"
fi
if command -v zola &>/dev/null; then
    echo "| Zola | $(zola --version 2>/dev/null | grep -oE '[0-9.]+') |" >> "$OUTPUT_FILE"
fi
if command -v eleventy &>/dev/null; then
    echo "| Eleventy | $(eleventy --version 2>/dev/null) |" >> "$OUTPUT_FILE"
fi
if command -v jekyll &>/dev/null; then
    echo "| Jekyll | $(jekyll --version 2>/dev/null | grep -oE '[0-9.]+') |" >> "$OUTPUT_FILE"
fi
if [ -f /benchmark/leafpress ]; then
    echo "| Leafpress | (local build) |" >> "$OUTPUT_FILE"
fi

echo ""
echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}    Benchmark Complete!${NC}"
echo -e "${GREEN}======================================${NC}"
echo ""
cat "$OUTPUT_FILE"
