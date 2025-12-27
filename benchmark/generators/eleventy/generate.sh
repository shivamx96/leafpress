#!/bin/bash
# Generate Eleventy test site

COUNT=$1
DIR=$2

mkdir -p "$DIR"
cd "$DIR"

# Create config
cat > .eleventy.js << 'EOF'
module.exports = function(eleventyConfig) {
  return {
    dir: {
      input: "src",
      output: "_site"
    }
  };
};
EOF

# Create package.json
cat > package.json << 'EOF'
{
  "name": "eleventy-benchmark",
  "version": "1.0.0",
  "scripts": {
    "build": "eleventy"
  }
}
EOF

# Create src directory and layouts
mkdir -p src/_includes

cat > src/_includes/base.njk << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ title }}</title></head>
<body>{{ content | safe }}</body></html>
EOF

# Create pages
for i in $(seq 1 $COUNT); do
    tag1="tag$((i % 20))"
    tag2="tag$(((i + 7) % 20))"
    link1=$((((i + 13) % COUNT) + 1))
    link2=$((((i + 137) % COUNT) + 1))

    cat > "src/page-$i.md" << EOF
---
title: Page $i - Topic $((i % 50))
tags:
  - $tag1
  - $tag2
layout: base.njk
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
