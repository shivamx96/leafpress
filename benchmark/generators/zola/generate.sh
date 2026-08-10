#!/usr/bin/env bash
# Generate the deterministic Zola comparison workload.

set -euo pipefail

COUNT=${1:?page count is required}
DIR=${2:?output directory is required}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../../lib/workload.sh
source "${SCRIPT_DIR}/../../lib/workload.sh"

mkdir -p "$DIR/content/notes" "$DIR/content/posts" "$DIR/content/tags" "$DIR/templates/partials" "$DIR/static"
cd "$DIR"

cat > config.toml << 'EOF'
base_url = "http://example.org"
title = "Benchmark Test"
compile_sass = false
build_search_index = false
EOF

cat > templates/partials/nav.html << 'EOF'
<nav><a href="/notes/">Notes</a><a href="/posts/">Posts</a><a href="/tags/">Tags</a></nav>
EOF

cat > templates/index.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ config.title }}</title></head>
<body>{% include "partials/nav.html" %}<h1>{{ config.title }}</h1>
{% for subsection in section.subsections %}{% set item = get_section(path=subsection) %}<a href="{{ item.permalink }}">{{ item.title }}</a>{% endfor %}
</body></html>
EOF

cat > templates/section.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ section.title }}</title></head>
<body>{% include "partials/nav.html" %}<h1>{{ section.title }}</h1>{{ section.content | safe }}</body></html>
EOF

cat > templates/page.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ page.title }}</title></head>
<body>{% include "partials/nav.html" %}<article><h1>{{ page.title }}</h1>{{ page.content | safe }}</article></body></html>
EOF

cat > templates/taxonomy_list.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ taxonomy.name }}</title></head>
<body>{% include "partials/nav.html" %}<h1>{{ taxonomy.name }}</h1>{% for term in terms %}<a href="{{ term.permalink }}">{{ term.name }}</a>{% endfor %}</body></html>
EOF

cat > templates/taxonomy_single.html << 'EOF'
<!DOCTYPE html>
<html><head><title>{{ term.name }}</title></head>
<body>{% include "partials/nav.html" %}<h1>{{ term.name }}</h1>{% for page in term.pages %}<a href="{{ page.permalink }}">{{ page.title }}</a>{% endfor %}</body></html>
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

# $title
EOF
    workload_write_section_links "$section" "$COUNT" >> "content/${section}/_index.md"
done

cat > content/tags/_index.md << 'EOF'
+++
title = "Tags"
template = "section.html"
page_template = "page.html"
+++

# Tags
EOF
workload_write_tag_index_links >> content/tags/_index.md

for ((tag = 0; tag < WORKLOAD_TAG_COUNT; tag++)); do
    cat > "content/tags/tag${tag}.md" << EOF
+++
title = "tag${tag}"
path = "tags/tag${tag}"
+++

# tag${tag}
EOF
    workload_write_tag_links "tag${tag}" "$COUNT" >> "content/tags/tag${tag}.md"
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

    cat > "content/${section}/${slug}.md" << EOF
+++
title = "$title"
[extra]
tags = ["$tag1", "$tag2"]
+++

# $title
$content
$links$code_block
EOF
done
