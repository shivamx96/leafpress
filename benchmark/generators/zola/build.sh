#!/bin/bash
# Build Zola site
cd "$1"
rm -rf public
# Zola outputs "Done in Xms" or "Done in X.Xs"
output=$(zola build 2>&1)
# Try to extract ms directly first, otherwise convert seconds
ms=$(echo "$output" | grep -oE 'Done in [0-9]+ms' | grep -oE '[0-9]+')
if [ -z "$ms" ]; then
    # Try seconds format (e.g., "Done in 0.5s")
    secs=$(echo "$output" | grep -oE 'Done in [0-9.]+s' | grep -oE '[0-9.]+')
    if [ -n "$secs" ]; then
        ms=$(echo "$secs" | awk '{printf "%.0f\n", $1 * 1000}')
    fi
fi
echo "$ms"
