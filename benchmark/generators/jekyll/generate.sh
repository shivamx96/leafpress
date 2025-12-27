#!/bin/bash
# Generate Jekyll test site

COUNT=$1
DIR=$2

mkdir -p "$DIR"
cd "$DIR"

# Create directory structure
mkdir -p _layouts _posts

# Create config
cat > _config.yml << 'EOF'
title: Benchmark Test
baseurl: ""
url: "http://example.org"
markdown: kramdown
EOF

# Create layout
cat > _layouts/default.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ page.title }}</title></head>
<body>{{ content }}</body></html>
EOF

cat > _layouts/post.html << 'EOF'
---
layout: default
---
<article><h1>{{ page.title }}</h1>{{ content }}</article>
EOF

# Create index
cat > index.html << 'EOF'
---
layout: default
title: Home
---
<h1>{{ site.title }}</h1>
{% for post in site.posts %}
<a href="{{ post.url }}">{{ post.title }}</a>
{% endfor %}
EOF

# Create pages (Jekyll uses _posts with date prefix)
TODAY=$(date +%Y-%m-%d)
for i in $(seq 1 $COUNT); do
    tag1="tag$((i % 20))"
    tag2="tag$(((i + 7) % 20))"
    link1=$((((i + 13) % COUNT) + 1))
    link2=$((((i + 137) % COUNT) + 1))

    cat > "_posts/${TODAY}-page-$i.md" << EOF
---
layout: post
title: "Page $i - Topic $((i % 50))"
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
