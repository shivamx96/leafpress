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
---
title: "$title"
tags: ["$tag1", "$tag2"]
---

# $title
$content
$links$code_block
EOF
done
