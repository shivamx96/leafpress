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

# Populate page fields without command substitutions. Generators call this in
# their hot loop so fixture preparation does not preheat the benchmark host.
workload_set_page() {
    local index=$1
    local count=$2
    local notes_count=$(((count * WORKLOAD_NOTES_PERCENT + 99) / 100))

    if ((index <= notes_count)); then
        WORKLOAD_SECTION="notes"
        WORKLOAD_KIND="note"
        WORKLOAD_LABEL="Note"
    else
        WORKLOAD_SECTION="posts"
        WORKLOAD_KIND="post"
        WORKLOAD_LABEL="Post"
    fi

    WORKLOAD_SLUG="${WORKLOAD_KIND}-${index}"
    WORKLOAD_ROUTE="/${WORKLOAD_SECTION}/${WORKLOAD_SLUG}/"
    WORKLOAD_TITLE="${WORKLOAD_LABEL} ${index} - Topic $((index % 50))"
    WORKLOAD_TAG_ONE="tag$((index % 20))"
    WORKLOAD_TAG_TWO="tag$(((index + 7) % 20))"

    case $((index % 3)) in
        0) WORKLOAD_GROWTH="seedling" ;;
        1) WORKLOAD_GROWTH="budding" ;;
        2) WORKLOAD_GROWTH="evergreen" ;;
    esac

    WORKLOAD_PARAGRAPH_COUNT=$((((index * 17 + 3) % 5) + 1))
    if (((index * 37 + 11) % 100 < 15)); then
        WORKLOAD_LINK_COUNT=0
    else
        WORKLOAD_LINK_COUNT=$((((index * 13 + 5) % 7) + 2))
    fi

    if (((index * 41 + 7) % 100 < 40)); then
        WORKLOAD_HAS_CODE_BLOCK=true
    else
        WORKLOAD_HAS_CODE_BLOCK=false
    fi
}

workload_set_paragraph() {
    local index=$1
    local paragraph=$2
    local paragraph_index=$(((index * 23 + paragraph * 19) % ${#WORKLOAD_PARAGRAPHS[@]}))
    WORKLOAD_PARAGRAPH=${WORKLOAD_PARAGRAPHS[$paragraph_index]}
}

workload_set_target() {
    local index=$1
    local link=$2
    local count=$3
    local target=$(((index * 97 + link * 53) % count + 1))
    local hub_count=10
    local notes_count=$(((count * WORKLOAD_NOTES_PERCENT + 99) / 100))

    if ((count < hub_count)); then
        hub_count=$count
    fi
    if (((index * 29 + link * 31) % 100 < 20)); then
        target=$(((index + link * 7) % hub_count + 1))
    fi

    if ((target <= notes_count)); then
        WORKLOAD_TARGET_SECTION="notes"
        WORKLOAD_TARGET_KIND="note"
        WORKLOAD_TARGET_LABEL="Note"
    else
        WORKLOAD_TARGET_SECTION="posts"
        WORKLOAD_TARGET_KIND="post"
        WORKLOAD_TARGET_LABEL="Post"
    fi
    WORKLOAD_TARGET=$target
    WORKLOAD_TARGET_SLUG="${WORKLOAD_TARGET_KIND}-${target}"
    WORKLOAD_TARGET_ROUTE="/${WORKLOAD_TARGET_SECTION}/${WORKLOAD_TARGET_SLUG}/"
    WORKLOAD_TARGET_TITLE="${WORKLOAD_TARGET_LABEL} ${target} - Topic $((target % 50))"
}

workload_notes_count() {
    local count=$1
    echo $(((count * WORKLOAD_NOTES_PERCENT + 99) / 100))
}

workload_section() {
    workload_set_page "$1" "$2"
    echo "$WORKLOAD_SECTION"
}

workload_kind() {
    workload_set_page "$1" "$2"
    echo "$WORKLOAD_KIND"
}

workload_slug() {
    workload_set_page "$1" "$2"
    echo "$WORKLOAD_SLUG"
}

workload_route() {
    workload_set_page "$1" "$2"
    echo "$WORKLOAD_ROUTE"
}

workload_title() {
    workload_set_page "$1" "$2"
    echo "$WORKLOAD_TITLE"
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
    workload_set_paragraph "$1" "$2"
    echo "$WORKLOAD_PARAGRAPH"
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
    workload_set_target "$1" "$2" "$3"
    echo "$WORKLOAD_TARGET"
}

workload_has_code_block() {
    local index=$1
    (((index * 41 + 7) % 100 < 40))
}
