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

for i in $(seq 1 "$COUNT"); do
    section=$(workload_section "$i" "$COUNT")
    slug=$(workload_slug "$i" "$COUNT")
    title=$(workload_title "$i" "$COUNT")
    tag1=$(workload_tag_one "$i")
    tag2=$(workload_tag_two "$i")
    paragraph_count=$(workload_paragraph_count "$i")
    link_count=$(workload_link_count "$i")
    content=""

    for p in $(seq 1 "$paragraph_count"); do
        content="${content}

$(workload_paragraph "$i" "$p")"
    done

    links=""
    if ((link_count > 0)); then
        links="

## Related Notes
"
        for l in $(seq 1 "$link_count"); do
            target=$(workload_link_target "$i" "$l" "$COUNT")
            links="${links}
- [$(workload_title "$target" "$COUNT")]($(workload_route "$target" "$COUNT"))"
        done
    fi

    code_block=""
    if workload_has_code_block "$i"; then
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
