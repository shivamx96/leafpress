#!/bin/bash
# Build Hugo site
cd "$1"
[ "$2" != "warm" ] && rm -rf public

# Cross-platform milliseconds (macOS date doesn't support %N)
now_ms() { python3 -c "import time; print(int(time.time() * 1000))"; }

start=$(now_ms)
hugo 2>&1 >/dev/null
end=$(now_ms)
echo $((end - start))
