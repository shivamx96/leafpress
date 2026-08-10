#!/usr/bin/env bash
# Generate the deterministic Hugo comparison workload.

set -euo pipefail

COUNT=${1:?page count is required}
DIR=${2:?output directory is required}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../../lib/workload.sh
source "${SCRIPT_DIR}/../../lib/workload.sh"

mkdir -p "$DIR/content/notes" "$DIR/content/posts" \
    "$DIR/themes/minimal/layouts/_default" "$DIR/themes/minimal/layouts/partials" "$DIR/static"
cd "$DIR"

cat > hugo.toml << 'EOF'
baseURL = 'http://example.org/'
languageCode = 'en-us'
title = 'Benchmark Test'
theme = 'minimal'
EOF

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
{{ define "main" }}<h1>{{ .Site.Title }}</h1>{{ range .Sections }}<a href="{{ .Permalink }}">{{ .Title }}</a>{{ end }}{{ end }}
EOF

for section in notes posts; do
    if [[ $section == notes ]]; then title="Notes"; else title="Posts"; fi
    cat > "content/${section}/_index.md" << EOF
---
title: "$title"
---
EOF
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
---
title: "$title"
tags: ["$tag1", "$tag2"]
---

# $title
$content
$links$code_block
EOF
done
