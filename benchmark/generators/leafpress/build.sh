#!/bin/bash
# Build Leafpress site
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LEAFPRESS="${SCRIPT_DIR}/../../leafpress"
cd "$1"
rm -rf _site
if [ ! -f "$LEAFPRESS" ]; then
    LEAFPRESS="/benchmark/leafpress"
fi
"$LEAFPRESS" build 2>&1 | grep -oE '[0-9]+ms' | head -1 | tr -d 'ms'
