#!/bin/bash
# Build Astro site
cd "$1"
rm -rf dist
start=$(date +%s%3N)
npm run build 2>&1 >/dev/null
end=$(date +%s%3N)
echo $((end - start))
