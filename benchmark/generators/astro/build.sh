#!/usr/bin/env bash
# Build Astro site
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${1:?site directory is required}"
[[ ${2:-} == "warm" ]] || rm -rf dist
python3 "${SCRIPT_DIR}/../../lib/time-command.py" npm run build
