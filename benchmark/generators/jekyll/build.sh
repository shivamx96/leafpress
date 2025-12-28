#!/bin/bash
# Build Jekyll site
cd "$1"
[ "$2" != "warm" ] && rm -rf _site

# Cross-platform milliseconds (macOS date doesn't support %N)
now_ms() { python3 -c "import time; print(int(time.time() * 1000))"; }

start=$(now_ms)
jekyll build --quiet 2>/dev/null
end=$(now_ms)
echo $((end - start))
