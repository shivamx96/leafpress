#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="${SCRIPT_DIR}/.."
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

first="${TEST_ROOT}/first"
second="${TEST_ROOT}/second"
bash "${SCRIPT_DIR}/generators/leafpress/generate.sh" 20 "$first"
bash "${SCRIPT_DIR}/generators/leafpress/generate.sh" 20 "$second"

first_digest=$(tree_digest "$first")
second_digest=$(tree_digest "$second")
[[ $first_digest == "$second_digest" ]] || fail "workload generation is not deterministic"

note_count=$(find "$first/notes" -maxdepth 1 -name 'note-*.md' | wc -l | tr -d ' ')
post_count=$(find "$first/posts" -maxdepth 1 -name 'post-*.md' | wc -l | tr -d ' ')
[[ $note_count == 14 && $post_count == 6 ]] || fail "expected a 14/6 notes/posts split, got $note_count/$post_count"
[[ -f "$first/index.md" && -f "$first/notes/_index.md" && -f "$first/posts/_index.md" ]] || fail "section homes are missing"
grep -Rq '\[\[post-' "$first/notes" || fail "notes do not contain cross-section wikilinks"

bash "${SCRIPT_DIR}/generators/leafpress/build.sh" "$first" >/dev/null
rendered="$first/_site/notes/note-1/index.html"
[[ -f $rendered ]] || fail "expected rendered note"
assert_contains 'href="/notes/">Notes' "$rendered"
assert_contains 'href="/posts/">Posts' "$rendered"
assert_contains 'href="/tags/">Tags' "$rendered"
nav_count=$(grep -c '<a class="lp-nav-link' "$rendered")
[[ $nav_count == 3 ]] || fail "expected exactly three navigation links, got $nav_count"

main_report="${TEST_ROOT}/main-report.md"
BENCHMARK_RESULTS_FILE=$main_report BENCHMARK_RUNS=2 BENCHMARK_WARMUPS=1 \
    BENCHMARK_PAGE_COUNTS=20 BENCHMARK_SSGS="leafpress-minimal leafpress" \
    bash "${SCRIPT_DIR}/run-all.sh" >/dev/null
assert_contains '**Workload**: v2 (hierarchical-notes-posts)' "$main_report"
assert_contains '## Generated Output Size' "$main_report"
assert_contains '| leafpress |' "$main_report"

stress_report="${TEST_ROOT}/stress-report.md"
BENCHMARK_RESULTS_FILE=$stress_report BENCHMARK_RUNS=1 BENCHMARK_WARMUPS=1 \
    BENCHMARK_PAGE_COUNTS=20 bash "${SCRIPT_DIR}/run-navigation-stress.sh" >/dev/null
assert_contains '| 20 |' "$stress_report"
assert_contains 'Nav links/page' "$stress_report"

echo "benchmark smoke tests passed"
