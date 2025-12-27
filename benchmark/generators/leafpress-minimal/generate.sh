#!/bin/bash
# Generate Leafpress test site (minimal: no graph, TOC, backlinks, wikilinks)

COUNT=$1
DIR=$2

mkdir -p "$DIR"
cd "$DIR"

# Create config with all extra features disabled
cat > leafpress.json << 'EOF'
{
  "title": "Benchmark Test",
  "graph": false,
  "toc": false,
  "backlinks": false,
  "wikilinks": false
}
EOF

# Create pages (use standard markdown links, not wikilinks)
for i in $(seq 1 $COUNT); do
    tag1="tag$((i % 20))"
    tag2="tag$(((i + 7) % 20))"
    link1=$((((i + 13) % COUNT) + 1))
    link2=$((((i + 137) % COUNT) + 1))

    case $((i % 3)) in
        0) growth="seedling" ;;
        1) growth="budding" ;;
        2) growth="evergreen" ;;
    esac

    cat > "page-$i.md" << EOF
---
title: Page $i - Topic $((i % 50))
tags: [$tag1, $tag2]
growth: $growth
---

# Page $i

Content for page $i about topic $((i % 50)).

Related: [page-$link1](/page-$link1/) and [page-$link2](/page-$link2/)

\`\`\`go
func example$i() {
    fmt.Println("Page $i")
}
\`\`\`
EOF
done
