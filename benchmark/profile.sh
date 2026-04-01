#!/bin/bash
# Profile leafpress build on a synthetic garden
#
# Usage: ./profile.sh [page_count]
# Output: results/cpu.prof, results/mem.prof

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COUNT=${1:-1000}
RESULTS_DIR="${SCRIPT_DIR}/results"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}LeafPress Build Profiler${NC}"
echo "========================"
echo "Pages: $COUNT"
echo ""

# Build leafpress
echo -e "${YELLOW}1. Building leafpress...${NC}"
cd "${SCRIPT_DIR}/../cli"
go build -o "${SCRIPT_DIR}/leafpress" ./cmd/leafpress
echo -e "${GREEN}   Done${NC}"

# Generate test site
WORKDIR=$(mktemp -d)
trap "rm -rf $WORKDIR" EXIT

echo -e "${YELLOW}2. Generating $COUNT page garden...${NC}"
bash "${SCRIPT_DIR}/generators/leafpress/generate.sh" "$COUNT" "$WORKDIR"
echo -e "${GREEN}   Done${NC}"

cd "$WORKDIR"
mkdir -p "$RESULTS_DIR"

# Warm-up build
echo -e "${YELLOW}3. Warm-up build...${NC}"
"${SCRIPT_DIR}/leafpress" build > /dev/null 2>&1
rm -rf _site
echo -e "${GREEN}   Done${NC}"

# Profiled build
echo -e "${YELLOW}4. Profiled build...${NC}"
"${SCRIPT_DIR}/leafpress" build \
    --cpuprofile="${RESULTS_DIR}/cpu.prof" \
    --memprofile="${RESULTS_DIR}/mem.prof"

echo ""
echo -e "${GREEN}========================${NC}"
echo -e "${GREEN}Profiling Complete!${NC}"
echo -e "${GREEN}========================${NC}"
echo ""
echo "Results:"
echo "  CPU profile:    ${RESULTS_DIR}/cpu.prof"
echo "  Memory profile: ${RESULTS_DIR}/mem.prof"
echo ""
echo "Analyze:"
echo "  go tool pprof -http=:8080 ${RESULTS_DIR}/cpu.prof"
echo "  go tool pprof -http=:8080 ${RESULTS_DIR}/mem.prof"
echo ""
echo "Quick summary:"
echo "  go tool pprof -top ${RESULTS_DIR}/cpu.prof"