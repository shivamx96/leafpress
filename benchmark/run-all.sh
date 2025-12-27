#!/usr/bin/env bash

# SSG Benchmark Suite
# Runs clean build benchmarks for multiple static site generators

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE="${SCRIPT_DIR}/results/BENCHMARK_${TIMESTAMP}.md"
RUNS=5
PAGE_COUNTS=(100 1000 2000)
SSGS=(zola hugo leafpress-minimal leafpress leafpress-full eleventy jekyll astro)

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

# Function to calculate stats (avg min max)
calc_stats() {
    local -a times=("$@")
    local sum=0
    local min=${times[0]}
    local max=${times[0]}

    for t in "${times[@]}"; do
        # Skip empty or non-numeric values
        if [[ -z "$t" || ! "$t" =~ ^[0-9]+$ ]]; then
            continue
        fi
        sum=$((sum + t))
        if [ "$t" -lt "$min" ]; then min=$t; fi
        if [ "$t" -gt "$max" ]; then max=$t; fi
    done

    local count=${#times[@]}
    if [ $count -eq 0 ]; then
        echo "0 0 0"
    else
        local avg=$((sum / count))
        echo "$avg $min $max"
    fi
}

# Check if SSG is available
check_ssg() {
    local ssg=$1
    case $ssg in
        leafpress) [ -f "${SCRIPT_DIR}/leafpress" ] || [ -f /benchmark/leafpress ] ;;
        leafpress-minimal) [ -f "${SCRIPT_DIR}/leafpress" ] || [ -f /benchmark/leafpress ] ;;
        leafpress-full) [ -f "${SCRIPT_DIR}/leafpress" ] || [ -f /benchmark/leafpress ] ;;
        hugo) command -v hugo &>/dev/null ;;
        zola) command -v zola &>/dev/null ;;
        eleventy) command -v eleventy &>/dev/null || command -v npx &>/dev/null ;;
        jekyll) command -v jekyll &>/dev/null ;;
        astro) command -v npm &>/dev/null ;;
    esac
}

# Initialize results file
cat > "$OUTPUT_FILE" << EOF
# SSG Benchmark Results

**Date**: $(date)
**System**: $(uname -s) $(uname -m)
**Runs per test**: $RUNS

## Results (Clean Build Times in ms)

| SSG | 100 pages | 1000 pages | 2000 pages |
|-----|-----------|------------|------------|
EOF

# Run benchmarks and write results directly
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
            read avg min max <<< $(calc_stats "${times[@]}")
            ssg_results="${ssg_results} ${avg}ms |"
            echo -e "${GREEN}    Average: ${avg}ms (${min}-${max}ms)${NC}"
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
echo "*leafpress-minimal: basic rendering only. leafpress: +wikilinks +backlinks. leafpress-full: +wikilinks +backlinks +graph +TOC.*" >> "$OUTPUT_FILE"

# Add methodology section
cat >> "$OUTPUT_FILE" << 'EOF'

## Methodology

- **Clean Build**: Removes output directory before each build
- **Content**: Each page has frontmatter, markdown content, code block, and internal links
- **Tags**: 20 unique tags distributed across pages
- **Links**: Each page links to 2 other pages

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
