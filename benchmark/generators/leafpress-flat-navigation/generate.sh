#!/usr/bin/env bash
# Generate the deliberately pathological flat-navigation Leafpress workload.

set -euo pipefail

COUNT=${1:?page count is required}
DIR=${2:?output directory is required}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../../lib/workload.sh
source "${SCRIPT_DIR}/../../lib/workload.sh"

mkdir -p "$DIR"
cd "$DIR"

cat > leafpress.json << 'EOF'
{
  "site": { "title": "Flat Navigation Stress Test" },
  "features": {
    "graph": false,
    "search": false,
    "toc": false,
    "backlinks": false,
    "wikilinks": false,
    "rss": false
  },
  "navigation": { "mode": "automatic" }
}
EOF

for ((i = 1; i <= COUNT; i++)); do
    cat > "page-${i}.md" << EOF
---
title: Page $i
---

# Page $i

Flat root-level note $i.
EOF
done
