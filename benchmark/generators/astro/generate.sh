#!/usr/bin/env bash
# Generate the deterministic Astro comparison workload.

set -euo pipefail

COUNT=${1:?page count is required}
DIR=${2:?output directory is required}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../../lib/workload.sh
source "${SCRIPT_DIR}/../../lib/workload.sh"

mkdir -p "$DIR/src/pages/notes" "$DIR/src/pages/posts" "$DIR/src/layouts"
cd "$DIR"

cat > package.json << 'EOF'
{
  "name": "astro-benchmark",
  "private": true,
  "scripts": { "build": "astro build" },
  "dependencies": { "astro": "^4.0.0" }
}
EOF

cat > astro.config.mjs << 'EOF'
import { defineConfig } from 'astro/config';
export default defineConfig({});
EOF

cat > src/layouts/Base.astro << 'EOF'
---
const { title } = Astro.props;
---
<!DOCTYPE html>
<html><head><title>{title}</title></head>
<body><slot /></body></html>
EOF

cat > src/pages/index.astro << 'EOF'
---
import Base from '../layouts/Base.astro';
---
<Base title="Home"><h1>Benchmark Test</h1></Base>
EOF

for section in notes posts; do
    if [[ $section == notes ]]; then title="Notes"; else title="Posts"; fi
    cat > "src/pages/${section}/index.astro" << EOF
---
import Base from '../../layouts/Base.astro';
---
<Base title="$title"><h1>$title</h1></Base>
EOF
done

for ((i = 1; i <= COUNT; i++)); do
    workload_set_page "$i" "$COUNT"
    section=$WORKLOAD_SECTION
    slug=$WORKLOAD_SLUG
    title=$WORKLOAD_TITLE
    tag1=$WORKLOAD_TAG_ONE
    tag2=$WORKLOAD_TAG_TWO
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
- [$WORKLOAD_TARGET_TITLE]($WORKLOAD_TARGET_ROUTE)"
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

    cat > "src/pages/${section}/${slug}.md" << EOF
---
layout: ../../layouts/Base.astro
title: $title
tags: [$tag1, $tag2]
---

# $title
$content
$links$code_block
EOF
done

if [[ ${BENCHMARK_SKIP_INSTALL:-false} != true ]]; then
    npm install --loglevel=error
fi
