#!/usr/bin/env bash
# Build Hugo site
cd "$1"
[ "$2" != "warm" ] && rm -rf public

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
python3 "${SCRIPT_DIR}/../../lib/time-command.py" hugo
