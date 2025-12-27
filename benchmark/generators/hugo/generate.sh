#!/bin/bash
# Generate Hugo test site

COUNT=$1
DIR=$2

mkdir -p "$DIR"
cd "$DIR"

# Initialize Hugo site structure
mkdir -p content themes/minimal/layouts/_default themes/minimal/layouts/partials static

# Create config
cat > hugo.toml << 'EOF'
baseURL = 'http://example.org/'
languageCode = 'en-us'
title = 'Benchmark Test'
theme = 'minimal'
EOF

# Create minimal theme
cat > themes/minimal/layouts/_default/baseof.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ .Title }}</title></head>
<body>{{ block "main" . }}{{ end }}</body></html>
EOF

cat > themes/minimal/layouts/_default/single.html << 'EOF'
{{ define "main" }}<article><h1>{{ .Title }}</h1>{{ .Content }}</article>{{ end }}
EOF

cat > themes/minimal/layouts/_default/list.html << 'EOF'
{{ define "main" }}<h1>{{ .Title }}</h1>{{ range .Pages }}<a href="{{ .Permalink }}">{{ .Title }}</a>{{ end }}{{ end }}
EOF

cat > themes/minimal/layouts/index.html << 'EOF'
{{ define "main" }}<h1>{{ .Site.Title }}</h1>{{ range .Site.RegularPages }}<a href="{{ .Permalink }}">{{ .Title }}</a>{{ end }}{{ end }}
EOF

# Create pages
for i in $(seq 1 $COUNT); do
    tag1="tag$((i % 20))"
    tag2="tag$(((i + 7) % 20))"
    link1=$((((i + 13) % COUNT) + 1))
    link2=$((((i + 137) % COUNT) + 1))

    cat > "content/page-$i.md" << EOF
---
title: "Page $i - Topic $((i % 50))"
tags: ["$tag1", "$tag2"]
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
