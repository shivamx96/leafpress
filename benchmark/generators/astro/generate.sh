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

# Create pages
for i in $(seq 1 $COUNT); do
    tag1="tag$((i % 20))"
    tag2="tag$(((i + 7) % 20))"
    link1=$((((i + 13) % COUNT) + 1))
    link2=$((((i + 137) % COUNT) + 1))

    cat > "src/pages/page-$i.md" << EOF
---
layout: ../layouts/Base.astro
title: Page $i - Topic $((i % 50))
tags: [$tag1, $tag2]
---

# Page $i

Content for page $i about topic $((i % 50)).

Related: [Page $link1](/page-$link1/) and [Page $link2](/page-$link2/)

\`\`\`go
func example$i() {
    fmt.Println("Page $i")
}
\`\`\`
EOF
done

# Install dependencies
npm install --loglevel=error 2>&1 || true
