#!/usr/bin/env bash
# Build Eleventy site
cd "$1"
[ "$2" != "warm" ] && rm -rf _site

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
python3 "${SCRIPT_DIR}/../../lib/time-command.py" eleventy --quiet
