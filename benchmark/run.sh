#!/usr/bin/env bash

# Main entry point for running SSG benchmarks
# Usage: ./run.sh [docker|local|stress]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE=${1:-docker}

echo "SSG Benchmark Suite"
echo "==================="
echo ""

if [ "$MODE" == "docker" ]; then
    echo "Running in Docker (recommended for fair comparison)"
    echo ""

    if docker compose version >/dev/null 2>&1; then
        COMPOSE=(docker compose)
    elif command -v docker-compose >/dev/null 2>&1; then
        COMPOSE=(docker-compose)
    else
        echo "Docker Compose is required for docker mode" >&2
        exit 1
    fi

    # Build and run Docker
    echo "Building Docker image (this may take a few minutes)..."
    cd "${SCRIPT_DIR}"
    "${COMPOSE[@]}" build

    echo ""
    echo "Running benchmarks..."
    "${COMPOSE[@]}" up --abort-on-container-exit --exit-code-from benchmark

    echo ""
    echo "Results saved to: ${SCRIPT_DIR}/results/"

elif [ "$MODE" == "local" ]; then
    echo "Running locally (requires all SSGs to be installed)"
    echo ""

    # Build Leafpress
    echo "Building Leafpress..."
    cd "${SCRIPT_DIR}/../cli"
    go build -o "${SCRIPT_DIR}/leafpress" ./cmd/leafpress

    # Run benchmark
    cd "${SCRIPT_DIR}"
    ./run-all.sh

elif [ "$MODE" == "stress" ]; then
    echo "Running the local flat-navigation stress benchmark"
    echo ""

    cd "${SCRIPT_DIR}/../cli"
    go build -o "${SCRIPT_DIR}/leafpress" ./cmd/leafpress

    cd "${SCRIPT_DIR}"
    ./run-navigation-stress.sh

else
    echo "Usage: ./run.sh [docker|local|stress]"
    echo ""
    echo "  docker  - Run in Docker container (recommended)"
    echo "  local   - Run locally (requires SSGs to be installed)"
    echo "  stress  - Run Leafpress's flat automatic-navigation stress case"
    exit 1
fi
