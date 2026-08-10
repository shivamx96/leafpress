#!/usr/bin/env bash
# Generate the deterministic Zola comparison workload.

set -euo pipefail

COUNT=${1:?page count is required}
DIR=${2:?output directory is required}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../../lib/workload.sh
source "${SCRIPT_DIR}/../../lib/workload.sh"

mkdir -p "$DIR/content/notes" "$DIR/content/posts" "$DIR/templates" "$DIR/static"
cd "$DIR"

cat > config.toml << 'EOF'
base_url = "http://example.org"
title = "Benchmark Test"
compile_sass = false
build_search_index = false

[[taxonomies]]
name = "tags"
EOF

cat > templates/index.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ config.title }}</title></head>
<body><h1>{{ config.title }}</h1>
{% for subsection in section.subsections %}{% set item = get_section(path=subsection) %}<a href="{{ item.permalink }}">{{ item.title }}</a>{% endfor %}
</body></html>
EOF

cat > templates/section.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ section.title }}</title></head>
<body><h1>{{ section.title }}</h1>{% for page in section.pages %}<a href="{{ page.permalink }}">{{ page.title }}</a>{% endfor %}</body></html>
EOF

cat > templates/page.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ page.title }}</title></head>
<body><article><h1>{{ page.title }}</h1>{{ page.content | safe }}</article></body></html>
EOF

cat > templates/taxonomy_list.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ taxonomy.name }}</title></head>
<body><h1>{{ taxonomy.name }}</h1>{% for term in terms %}<a href="{{ term.permalink }}">{{ term.name }}</a>{% endfor %}</body></html>
EOF

cat > templates/taxonomy_single.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ term.name }}</title></head>
<body><h1>{{ term.name }}</h1>{% for page in term.pages %}<a href="{{ page.permalink }}">{{ page.title }}</a>{% endfor %}</body></html>
EOF

cat > content/_index.md << 'EOF'
+++
template = "index.html"
+++
EOF

for section in notes posts; do
    if [[ $section == notes ]]; then title="Notes"; else title="Posts"; fi
    cat > "content/${section}/_index.md" << EOF
+++
title = "$title"
sort_by = "title"
template = "section.html"
page_template = "page.html"
+++
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

    cat > "content/${section}/${slug}.md" << EOF
+++
title = "$title"
[taxonomies]
tags = ["$tag1", "$tag2"]
+++

# $title
$content
$links$code_block
EOF
done
