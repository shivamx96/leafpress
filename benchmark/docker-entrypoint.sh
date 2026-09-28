#!/bin/bash
# Docker entrypoint - builds leafpress from mounted source

set -e

# Build leafpress if source is mounted
if [ -d /leafpress-src/cli ]; then
    echo "Building leafpress from source..."
    git config --global --add safe.directory /leafpress-src
    cd /leafpress-src/cli
    go build -o /benchmark/leafpress ./cmd/leafpress
    echo "leafpress built successfully"
fi

# Run the command
exec "$@"
