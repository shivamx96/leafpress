#!/bin/bash

# Main entry point for running SSG benchmarks
# Usage: ./run.sh [docker|local]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODE=${1:-docker}

echo "SSG Benchmark Suite"
echo "==================="
echo ""

if [ "$MODE" == "docker" ]; then
    echo "Running in Docker (recommended for fair comparison)"
    echo ""

    # Build Leafpress first and copy to benchmark dir
    echo "Building Leafpress..."
    cd "${SCRIPT_DIR}/.."
    go build -o "${SCRIPT_DIR}/leafpress" ./cmd/leafpress

    # Build and run Docker
    echo "Building Docker image (this may take a few minutes)..."
    cd "${SCRIPT_DIR}"
    docker-compose build

    echo ""
    echo "Running benchmarks..."
    docker-compose up --abort-on-container-exit

    echo ""
    echo "Results saved to: ${SCRIPT_DIR}/results/"

elif [ "$MODE" == "local" ]; then
    echo "Running locally (requires all SSGs to be installed)"
    echo ""

    # Build Leafpress
    echo "Building Leafpress..."
    cd "${SCRIPT_DIR}/.."
    go build -o "${SCRIPT_DIR}/leafpress" ./cmd/leafpress

    # Run benchmark
    cd "${SCRIPT_DIR}"
    ./run-all.sh

else
    echo "Usage: ./run.sh [docker|local]"
    echo ""
    echo "  docker  - Run in Docker container (recommended)"
    echo "  local   - Run locally (requires SSGs to be installed)"
    exit 1
fi
