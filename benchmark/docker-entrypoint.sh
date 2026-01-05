#!/bin/bash
# Docker entrypoint - builds Leafpress from mounted source

set -e

# Build Leafpress if source is mounted
if [ -d /leafpress-src/cli ]; then
    echo "Building Leafpress from source..."
    cd /leafpress-src/cli
    go build -o /benchmark/leafpress ./cmd/leafpress
    echo "Leafpress built successfully"
fi

# Run the command
exec "$@"
