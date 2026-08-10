#!/usr/bin/env bash
# Build Leafpress site (minimal config)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LEAFPRESS=${LEAFPRESS_BIN:-"${SCRIPT_DIR}/../../leafpress"}
cd "${1:?site directory is required}"
[[ ${2:-} == "warm" ]] || rm -rf _site
if [[ ! -x $LEAFPRESS ]]; then
    LEAFPRESS="/benchmark/leafpress"
fi

python3 "${SCRIPT_DIR}/../../lib/time-command.py" "$LEAFPRESS" build
