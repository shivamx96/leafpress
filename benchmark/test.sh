#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="${SCRIPT_DIR}/.."
# shellcheck source=lib/workload.sh
source "${SCRIPT_DIR}/lib/workload.sh"
TEST_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/leafpress-benchmark-test.XXXXXX")

cleanup() {
    if [[ -n ${TEST_ROOT:-} && -d $TEST_ROOT && $(basename "$TEST_ROOT") == leafpress-benchmark-test.* ]]; then
        rm -rf -- "$TEST_ROOT"
    fi
}
trap cleanup EXIT

fail() {
    echo "benchmark test failed: $*" >&2
    exit 1
}

assert_contains() {
    local expected=$1
    local file=$2
    grep -Fq "$expected" "$file" || fail "expected '$expected' in $file"
}

tree_digest() {
    local root=$1
    (cd "$root" && find . -type f -print | LC_ALL=C sort | while IFS= read -r file; do cksum "$file"; done) | cksum | awk '{print $1 ":" $2}'
}

bash -n "$SCRIPT_DIR"/*.sh "$SCRIPT_DIR"/lib/*.sh "$SCRIPT_DIR"/generators/*/*.sh

binary=${LEAFPRESS_BIN:-"${TEST_ROOT}/leafpress"}
if [[ ! -x $binary ]]; then
    (cd "${REPO_DIR}/cli" && go build -o "$binary" ./cmd/leafpress)
fi
export LEAFPRESS_BIN=$binary

content_path() {
    local generator=$1
    local root=$2
    local section=$3
    local slug=$4
    case $generator in
        astro) echo "$root/src/pages/$section/$slug.md" ;;
        eleventy) echo "$root/src/$section/$slug.md" ;;
        hugo|zola) echo "$root/content/$section/$slug.md" ;;
        jekyll|leafpress|leafpress-minimal) echo "$root/$section/$slug.md" ;;
        *) fail "unknown generator in test: $generator" ;;
    esac
}

section_index_path() {
    local generator=$1
    local root=$2
    local section=$3
    case $generator in
        astro) echo "$root/src/pages/$section/index.md" ;;
        eleventy) echo "$root/src/$section/index.md" ;;
        hugo|zola) echo "$root/content/$section/_index.md" ;;
        jekyll) echo "$root/$section/index.md" ;;
        leafpress|leafpress-minimal) echo "$root/$section/_index.md" ;;
    esac
}

tag_source_path() {
    local generator=$1
    local root=$2
    local slug=$3
    case $generator in
        astro) echo "$root/src/pages/tags/$slug.md" ;;
        eleventy) echo "$root/src/tags/$slug.md" ;;
        hugo|zola) echo "$root/content/tags/${slug/index/_index}.md" ;;
        jekyll) echo "$root/tags/$slug.md" ;;
    esac
}

generators=(astro eleventy hugo jekyll leafpress-minimal leafpress zola)
for generator_name in "${generators[@]}"; do
    first="${TEST_ROOT}/${generator_name}-first"
    second="${TEST_ROOT}/${generator_name}-second"
    BENCHMARK_SKIP_INSTALL=true bash "${SCRIPT_DIR}/generators/${generator_name}/generate.sh" 20 "$first"
    BENCHMARK_SKIP_INSTALL=true bash "${SCRIPT_DIR}/generators/${generator_name}/generate.sh" 20 "$second"

    [[ $(tree_digest "$first") == "$(tree_digest "$second")" ]] || fail "$generator_name generation is not deterministic"

    for ((index = 1; index <= 20; index++)); do
        workload_set_page "$index" 20
        page_file=$(content_path "$generator_name" "$first" "$WORKLOAD_SECTION" "$WORKLOAD_SLUG")
        [[ -f $page_file ]] || fail "$generator_name is missing $WORKLOAD_ROUTE"
        assert_contains "$WORKLOAD_TITLE" "$page_file"
        assert_contains "$WORKLOAD_TAG_ONE" "$page_file"
        assert_contains "$WORKLOAD_TAG_TWO" "$page_file"
        workload_set_paragraph "$index" 1
        assert_contains "$WORKLOAD_PARAGRAPH" "$page_file"
    done

    notes_index=$(section_index_path "$generator_name" "$first" notes)
    posts_index=$(section_index_path "$generator_name" "$first" posts)
    [[ -f $notes_index && -f $posts_index ]] || fail "$generator_name section homes are missing"

    if [[ $generator_name != leafpress && $generator_name != leafpress-minimal ]]; then
        assert_contains '/notes/note-1/' "$notes_index"
        assert_contains '/posts/post-15/' "$posts_index"
        tags_index=$(tag_source_path "$generator_name" "$first" index)
        tag_zero=$(tag_source_path "$generator_name" "$first" tag0)
        [[ -f $tags_index && -f $tag_zero ]] || fail "$generator_name tag support pages are missing"
        assert_contains '/tags/tag0/' "$tags_index"
        assert_contains '/notes/note-13/' "$tag_zero"
    fi
done

leafpress_first="${TEST_ROOT}/leafpress-first"
grep -Rq '\[\[post-' "$leafpress_first/notes" || fail "Leafpress notes do not contain cross-section wikilinks"

bash "${SCRIPT_DIR}/generators/leafpress/build.sh" "$leafpress_first" >/dev/null
rendered="$leafpress_first/_site/notes/note-1/index.html"
[[ -f $rendered ]] || fail "expected rendered note"
assert_contains 'href="/notes/">Notes' "$rendered"
assert_contains 'href="/posts/">Posts' "$rendered"
assert_contains 'href="/tags/">Tags' "$rendered"
nav_count=$(grep -c '<a class="lp-nav-link' "$rendered")
[[ $nav_count == 3 ]] || fail "expected exactly three navigation links, got $nav_count"
[[ -f "$leafpress_first/_site/tags/index.html" && -f "$leafpress_first/_site/tags/tag0/index.html" ]] || fail "Leafpress tag pages are missing"

main_report="${TEST_ROOT}/main-report.md"
BENCHMARK_RESULTS_FILE=$main_report BENCHMARK_RUNS=2 BENCHMARK_WARMUPS=1 \
    BENCHMARK_PAGE_COUNTS=20 BENCHMARK_SSGS="leafpress-minimal leafpress" \
    bash "${SCRIPT_DIR}/run-all.sh" >/dev/null
assert_contains '**Workload**: v2 (hierarchical-notes-posts)' "$main_report"
grep -Eq '^\*\*Date\*\*: [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$' "$main_report" || \
    fail "main report timestamp is not UTC RFC 3339"
assert_contains '**Scheduling**: deterministic interleaved rotation' "$main_report"
assert_contains '**Leafpress SHA-256**:' "$main_report"
assert_contains '## Generated Output (logical bytes and file count)' "$main_report"
assert_contains 'files)' "$main_report"
assert_contains '| leafpress |' "$main_report"
if grep -Fq '| 0 B (' "$main_report"; then
    fail "generated output metrics lost the byte count"
fi

strict_report="${TEST_ROOT}/strict-report.md"
printf 'sentinel\n' > "$strict_report"
if BENCHMARK_RESULTS_FILE=$strict_report BENCHMARK_RUNS=1 BENCHMARK_WARMUPS=0 \
    BENCHMARK_PAGE_COUNTS=20 BENCHMARK_SSGS="missing-adapter" \
    bash "${SCRIPT_DIR}/run-all.sh" >/dev/null 2>&1; then
    fail "strict benchmark accepted a missing adapter"
fi
[[ $(cat "$strict_report") == sentinel ]] || fail "failed strict run replaced the existing report"

stress_report="${TEST_ROOT}/stress-report.md"
BENCHMARK_RESULTS_FILE=$stress_report BENCHMARK_RUNS=1 BENCHMARK_WARMUPS=1 \
    BENCHMARK_PAGE_COUNTS=20 bash "${SCRIPT_DIR}/run-navigation-stress.sh" >/dev/null
assert_contains '| 20 |' "$stress_report"
assert_contains 'Nav links/page' "$stress_report"
grep -Eq '^\*\*Date\*\*: [0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$' "$stress_report" || \
    fail "stress report timestamp is not UTC RFC 3339"

echo "benchmark smoke tests passed"
