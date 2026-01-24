#!/bin/bash
# Generate Astro test site

COUNT=$1
DIR=$2

mkdir -p "$DIR"
cd "$DIR"

# Create directory structure
mkdir -p src/pages src/layouts

# Create package.json
cat > package.json << 'EOF'
{
  "name": "astro-benchmark",
  "version": "1.0.0",
  "scripts": {
    "build": "astro build"
  },
  "dependencies": {
    "astro": "^4.0.0"
  }
}
EOF

# Create Astro config
cat > astro.config.mjs << 'EOF'
import { defineConfig } from 'astro/config';
export default defineConfig({});
EOF

# Create layout
cat > src/layouts/Base.astro << 'EOF'
---
const { title } = Astro.props;
---
<!DOCTYPE html>
<html><head><title>{title}</title></head>
<body><slot /></body></html>
EOF

# Create index
cat > src/pages/index.astro << 'EOF'
---
import Base from '../layouts/Base.astro';
---
<Base title="Home"><h1>Benchmark Test</h1></Base>
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

# Create pages
for i in $(seq 1 $COUNT); do
    tag1="tag$((i % 20))"
    tag2="tag$(((i + 7) % 20))"

    # Variable number of paragraphs (1-5)
    num_paragraphs=$(( (RANDOM % 5) + 1 ))

    # Variable number of links (2-8), unless orphan (~15% chance)
    is_orphan=$(( RANDOM % 100 ))
    if [ $is_orphan -lt 15 ]; then
        num_links=0
    else
        num_links=$(( (RANDOM % 7) + 2 ))
    fi

    # Build content with variable paragraphs
    content=""
    for p in $(seq 1 $num_paragraphs); do
        para_idx=$(( RANDOM % ${#PARAGRAPHS[@]} ))
        content="$content

${PARAGRAPHS[$para_idx]}"
    done

    # Build links section
    links=""
    if [ $num_links -gt 0 ]; then
        links="

## Related Notes

"
        for l in $(seq 1 $num_links); do
            target=$(( (RANDOM % COUNT) + 1 ))
            # Bias toward "hub" pages (pages 1-10 get more links)
            if [ $(( RANDOM % 100 )) -lt 20 ]; then
                target=$(( (RANDOM % 10) + 1 ))
            fi
            links="$links- [Page $target](/page-$target/)
"
        done
    fi

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

    cat > "src/pages/page-$i.md" << EOF
---
layout: ../layouts/Base.astro
title: Page $i - Topic $((i % 50))
tags: [$tag1, $tag2]
---

# Page $i
$content
$links$code_block
EOF
done

# Install dependencies
npm install --loglevel=error 2>&1 || true
