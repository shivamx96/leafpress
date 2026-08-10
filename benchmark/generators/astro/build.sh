#!/usr/bin/env bash
# Build Astro site
cd "$1"
[ "$2" != "warm" ] && rm -rf dist

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
python3 "${SCRIPT_DIR}/../../lib/time-command.py" npm run build
