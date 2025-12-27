#!/bin/bash
# Build Eleventy site
cd "$1"
rm -rf _site
eleventy --quiet 2>&1 | grep -oE 'in [0-9.]+ seconds' | grep -oE '[0-9.]+' | awk '{printf "%.0f\n", $1 * 1000}'
