#!/bin/bash
# Generate Leafpress test site (minimal features — comparable to Hugo/Zola)

COUNT=$1
DIR=$2

mkdir -p "$DIR"
cd "$DIR"

# Disable all extra features to match Hugo/Zola baseline
cat > leafpress.json << 'EOF'
{
  "site": {
    "title": "Benchmark Test"
  },
  "features": {
    "graph": false,
    "search": false,
    "toc": false,
    "backlinks": false,
    "wikilinks": false,
    "rss": false
  }
}
EOF

# Lorem ipsum paragraphs for variable content
PARAGRAPHS=(
    "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."
    "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
    "Curabitur pretium tincidunt lacus. Nulla gravida orci a odio. Nullam varius, turpis et commodo pharetra, est eros bibendum elit, nec luctus magna felis sollicitudin mauris."
    "Integer in mauris eu nibh euismod gravida. Duis ac tellus et risus vulputate vehicula. Donec lobortis risus a elit. Etiam tempor ultrices nisi. Praesent interdum mollis neque."
    "Suspendisse potenti. Sed eget dolor. Sed nec libero non leo volutpat consequat. Nullam vel sem. Pellentesque libero tortor, tincidunt et, tincidunt eget, semper nec, quam."
    "Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae. Morbi lacinia molestie dui. Praesent blandit dolor. Sed non quam. In vel mi sit amet augue congue elementum."
    "Fusce commodo aliquam arcu. Nam commodo suscipit quam. Quisque id odio. Praesent venenatis metus at tortor pulvinar varius. Aenean ultricies mi vitae est."
    "Mauris placerat eleifend leo. Quisque sit amet est et sapien ullamcorper pharetra. Vestibulum erat wisi, condimentum sed, commodo vitae, ornare sit amet, wisi."
)

# Create pages (no wikilinks, simpler content — matches Hugo/Zola test pages)
for i in $(seq 1 $COUNT); do
    tag1="tag$((i % 20))"
    tag2="tag$(((i + 7) % 20))"

    # Variable number of paragraphs (1-5)
    num_paragraphs=$(( (RANDOM % 5) + 1 ))

    # Build content with variable paragraphs
    content=""
    for p in $(seq 1 $num_paragraphs); do
        para_idx=$(( RANDOM % ${#PARAGRAPHS[@]} ))
        content="$content

${PARAGRAPHS[$para_idx]}"
    done

    # Randomly add code block (~40% of pages)
    code_block=""
    if [ $(( RANDOM % 100 )) -lt 40 ]; then
        code_block="

\`\`\`go
func example$i() {
    fmt.Println(\"Page $i\")
}
\`\`\`"
    fi

    cat > "page-$i.md" << EOF
---
title: Page $i - Topic $((i % 50))
tags: [$tag1, $tag2]
---

# Page $i
$content
$code_block
EOF
done
