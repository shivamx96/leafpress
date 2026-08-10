#!/usr/bin/env bash
# Generate the full-featured Leafpress workload.

set -euo pipefail

COUNT=${1:?page count is required}
DIR=${2:?output directory is required}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../../lib/workload.sh
source "${SCRIPT_DIR}/../../lib/workload.sh"

mkdir -p "$DIR/notes" "$DIR/posts"
cd "$DIR"

cat > leafpress.json << 'EOF'
{
  "site": {
    "title": "Benchmark Test"
  },
  "features": {
    "graph": true,
    "toc": true
  },
  "navigation": {
    "mode": "automatic",
    "includeTags": true
  }
}
EOF

cat > index.md << 'EOF'
---
title: Home
---

# Benchmark Test

A deterministic synthetic garden containing notes and posts.
EOF

cat > notes/_index.md << 'EOF'
---
title: Notes
---

# Notes

Working notes in the benchmark garden.
EOF

cat > posts/_index.md << 'EOF'
---
title: Posts
---

# Posts

Published posts in the benchmark garden.
EOF

for ((i = 1; i <= COUNT; i++)); do
    workload_set_page "$i" "$COUNT"
    section=$WORKLOAD_SECTION
    slug=$WORKLOAD_SLUG
    title=$WORKLOAD_TITLE
    tag1=$WORKLOAD_TAG_ONE
    tag2=$WORKLOAD_TAG_TWO
    growth=$WORKLOAD_GROWTH
    paragraph_count=$WORKLOAD_PARAGRAPH_COUNT
    link_count=$WORKLOAD_LINK_COUNT

    content=""
    for ((p = 1; p <= paragraph_count; p++)); do
        workload_set_paragraph "$i" "$p"
        content="${content}

$WORKLOAD_PARAGRAPH"
    done

    links=""
    if ((link_count > 0)); then
        links="

## Related Notes
"
        for ((l = 1; l <= link_count; l++)); do
            workload_set_target "$i" "$l" "$COUNT"
            links="${links}
- [[$WORKLOAD_TARGET_SLUG]]"
        done
    fi

    code_block=""
    if [[ $WORKLOAD_HAS_CODE_BLOCK == true ]]; then
        code_block="

\`\`\`go
func example$i() {
    fmt.Println(\"Page $i\")
}
\`\`\`"
    fi

    cat > "${section}/${slug}.md" << EOF
---
title: $title
tags: [$tag1, $tag2]
growth: $growth
---

# $title
$content
$links$code_block
EOF
done
