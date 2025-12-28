#!/bin/bash
# Build Leafpress site (full features)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LEAFPRESS="${SCRIPT_DIR}/../../leafpress"
cd "$1"
[ "$2" != "warm" ] && rm -rf _site
if [ ! -f "$LEAFPRESS" ]; then
    LEAFPRESS="/benchmark/leafpress"
fi

# Cross-platform milliseconds (macOS date doesn't support %N)
now_ms() { python3 -c "import time; print(int(time.time() * 1000))"; }

start=$(now_ms)
"$LEAFPRESS" build 2>&1 >/dev/null
end=$(now_ms)
echo $((end - start))
