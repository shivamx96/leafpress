#!/bin/bash
# Generate Zola test site

COUNT=$1
DIR=$2

mkdir -p "$DIR"
cd "$DIR"

# Create directory structure
mkdir -p content templates static

# Create config
cat > config.toml << 'EOF'
base_url = "http://example.org"
title = "Benchmark Test"
compile_sass = false
build_search_index = false

[[taxonomies]]
name = "tags"
EOF

# Create templates
cat > templates/index.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ config.title }}</title></head>
<body><h1>{{ config.title }}</h1>
{% for page in section.pages %}<a href="{{ page.permalink }}">{{ page.title }}</a>{% endfor %}
</body></html>
EOF

cat > templates/page.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ page.title }}</title></head>
<body><article><h1>{{ page.title }}</h1>{{ page.content | safe }}</article></body></html>
EOF

cat > templates/taxonomy_list.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ taxonomy.name }}</title></head>
<body><h1>{{ taxonomy.name }}</h1>
{% for term in terms %}<a href="{{ term.permalink }}">{{ term.name }}</a>{% endfor %}
</body></html>
EOF

cat > templates/taxonomy_single.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ term.name }}</title></head>
<body><h1>{{ term.name }}</h1>
{% for page in term.pages %}<a href="{{ page.permalink }}">{{ page.title }}</a>{% endfor %}
</body></html>
EOF

# Create section
cat > content/_index.md << 'EOF'
+++
template = "index.html"
+++
EOF

# Create pages
for i in $(seq 1 $COUNT); do
    tag1="tag$((i % 20))"
    tag2="tag$(((i + 7) % 20))"
    link1=$((((i + 13) % COUNT) + 1))
    link2=$((((i + 137) % COUNT) + 1))

    cat > "content/page-$i.md" << EOF
+++
title = "Page $i - Topic $((i % 50))"
[taxonomies]
tags = ["$tag1", "$tag2"]
+++

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
