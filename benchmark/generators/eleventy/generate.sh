#!/usr/bin/env bash
# Generate the deterministic Eleventy comparison workload.

set -euo pipefail

COUNT=${1:?page count is required}
DIR=${2:?output directory is required}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../../lib/workload.sh
source "${SCRIPT_DIR}/../../lib/workload.sh"

mkdir -p "$DIR/src/_includes" "$DIR/src/notes" "$DIR/src/posts" "$DIR/src/tags"
cd "$DIR"

cat > .eleventy.js << 'EOF'
module.exports = function() {
  return { dir: { input: "src", output: "_site" } };
};
EOF

cat > package.json << 'EOF'
{
  "name": "eleventy-benchmark",
  "private": true,
  "scripts": { "build": "eleventy" }
}
EOF

cat > src/_includes/base.njk << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ title }}</title></head>
<body><nav><a href="/notes/">Notes</a><a href="/posts/">Posts</a><a href="/tags/">Tags</a></nav>{{ content | safe }}</body></html>
EOF

cat > src/index.md << 'EOF'
---
layout: base.njk
title: Home
---
# Benchmark Test
EOF

for section in notes posts; do
    if [[ $section == notes ]]; then title="Notes"; else title="Posts"; fi
    cat > "src/${section}/index.md" << EOF
---
layout: base.njk
title: $title
---
# $title
EOF
    workload_write_section_links "$section" "$COUNT" >> "src/${section}/index.md"
done

cat > src/tags/index.md << 'EOF'
---
layout: base.njk
title: Tags
---
# Tags
EOF
workload_write_tag_index_links >> src/tags/index.md

for ((tag = 0; tag < WORKLOAD_TAG_COUNT; tag++)); do
    cat > "src/tags/tag${tag}.md" << EOF
---
layout: base.njk
title: tag${tag}
permalink: /tags/tag${tag}/index.html
---
# tag${tag}
EOF
    workload_write_tag_links "tag${tag}" "$COUNT" >> "src/tags/tag${tag}.md"
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

    cat > "src/${section}/${slug}.md" << EOF
---
title: $title
tags:
  - $tag1
  - $tag2
layout: base.njk
---

# $title
$content
$links$code_block
EOF
done
