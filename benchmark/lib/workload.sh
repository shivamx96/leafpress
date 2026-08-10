#!/usr/bin/env bash

# Shared deterministic workload contract for every benchmark generator.
# Increment this when content shape, routing, or feature distribution changes.
WORKLOAD_VERSION=2
WORKLOAD_NOTES_PERCENT=70

WORKLOAD_PARAGRAPHS=(
    "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat."
    "Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
    "Curabitur pretium tincidunt lacus. Nulla gravida orci a odio. Nullam varius, turpis et commodo pharetra, est eros bibendum elit, nec luctus magna felis sollicitudin mauris."
    "Integer in mauris eu nibh euismod gravida. Duis ac tellus et risus vulputate vehicula. Donec lobortis risus a elit. Etiam tempor ultrices nisi. Praesent interdum mollis neque."
    "Suspendisse potenti. Sed eget dolor. Sed nec libero non leo volutpat consequat. Nullam vel sem. Pellentesque libero tortor, tincidunt et, tincidunt eget, semper nec, quam."
    "Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae. Morbi lacinia molestie dui. Praesent blandit dolor. Sed non quam. In vel mi sit amet augue congue elementum."
    "Fusce commodo aliquam arcu. Nam commodo suscipit quam. Quisque id odio. Praesent venenatis metus at tortor pulvinar varius. Aenean ultricies mi vitae est."
    "Mauris placerat eleifend leo. Quisque sit amet est et sapien ullamcorper pharetra. Vestibulum erat wisi, condimentum sed, commodo vitae, ornare sit amet, wisi."
)

workload_notes_count() {
    local count=$1
    echo $(((count * WORKLOAD_NOTES_PERCENT + 99) / 100))
}

workload_section() {
    local index=$1
    local count=$2
    if ((index <= $(workload_notes_count "$count"))); then
        echo "notes"
    else
        echo "posts"
    fi
}

workload_kind() {
    if [[ $(workload_section "$1" "$2") == "notes" ]]; then
        echo "note"
    else
        echo "post"
    fi
}

workload_slug() {
    echo "$(workload_kind "$1" "$2")-$1"
}

workload_route() {
    echo "/$(workload_section "$1" "$2")/$(workload_slug "$1" "$2")/"
}

workload_title() {
    local index=$1
    local count=$2
    local kind
    local label
    kind=$(workload_kind "$index" "$count")
    if [[ $kind == "note" ]]; then label="Note"; else label="Post"; fi
    printf '%s %d - Topic %d\n' "$label" "$index" "$((index % 50))"
}

workload_tag_one() {
    local index=$1
    echo "tag$((index % 20))"
}

workload_tag_two() {
    local index=$1
    echo "tag$(((index + 7) % 20))"
}

workload_growth() {
    local index=$1
    case $((index % 3)) in
        0) echo "seedling" ;;
        1) echo "budding" ;;
        2) echo "evergreen" ;;
    esac
}

workload_paragraph_count() {
    local index=$1
    echo "$((((index * 17 + 3) % 5) + 1))"
}

workload_paragraph() {
    local index=$1
    local paragraph=$2
    local paragraph_index=$(((index * 23 + paragraph * 19) % ${#WORKLOAD_PARAGRAPHS[@]}))
    echo "${WORKLOAD_PARAGRAPHS[$paragraph_index]}"
}

workload_link_count() {
    local index=$1
    if (((index * 37 + 11) % 100 < 15)); then
        echo 0
    else
        echo "$((((index * 13 + 5) % 7) + 2))"
    fi
}

workload_link_target() {
    local index=$1
    local link=$2
    local count=$3
    local target=$(((index * 97 + link * 53) % count + 1))
    local hub_count=10

    if ((count < hub_count)); then
        hub_count=$count
    fi
    if (((index * 29 + link * 31) % 100 < 20)); then
        target=$(((index + link * 7) % hub_count + 1))
    fi
    echo "$target"
}

workload_has_code_block() {
    local index=$1
    (((index * 41 + 7) % 100 < 40))
}
