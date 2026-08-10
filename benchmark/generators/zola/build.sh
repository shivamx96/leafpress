#!/usr/bin/env bash
# Build Zola site
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${1:?site directory is required}"
[[ ${2:-} == "warm" ]] || rm -rf public
python3 "${SCRIPT_DIR}/../../lib/time-command.py" zola build
