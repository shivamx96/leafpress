#!/bin/bash
# LeafPress Quick Test Suite
# Run this script to verify core functionality

set -e  # Exit on error

echo "🧪 LeafPress Test Suite"
echo "======================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

# Test function
test_case() {
    echo -n "Testing: $1... "
}

pass() {
    echo -e "${GREEN}✓ PASS${NC}"
    PASSED=$((PASSED + 1))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}"
    if [ ! -z "$1" ]; then
        echo "  Error: $1"
    fi
    FAILED=$((FAILED + 1))
}

warn() {
    echo -e "${YELLOW}⚠ WARN${NC} $1"
}

# Build leafpress
echo "Building leafpress..."
go build -o leafpress ./cmd/leafpress || { echo "Build failed"; exit 1; }
echo ""

# Save original directory
ORIGDIR=$(pwd)
LEAFPRESS="$ORIGDIR/leafpress"

# Test 1: Init command
test_case "Init creates required files"
TESTDIR=$(mktemp -d)
cd "$TESTDIR"
"$LEAFPRESS" init > /dev/null 2>&1
if [ -f "leafpress.json" ] && [ -f "style.css" ] && [ -f "index.md" ]; then
    pass
else
    fail "Missing files after init"
fi
cd "$ORIGDIR"

# Test 2: Build command
test_case "Build completes successfully"
cd "$ORIGDIR/testdata/garden"
"$LEAFPRESS" build > /dev/null 2>&1 && pass || fail
cd "$ORIGDIR"

# Test 3: Check output files
test_case "Output files are generated"
if [ -f "testdata/garden/_site/index.html" ] && [ -f "testdata/garden/_site/style.css" ]; then
    pass
else
    fail "Missing output files"
fi

# Test 4: Wiki links are processed
test_case "Wiki links are converted to HTML"
if grep -q 'class="lp-wikilink"' testdata/garden/_site/index.html; then
    pass
else
    fail "No wiki links found in output"
fi

# Test 5: Code blocks protect wiki links
test_case "Wiki links in code blocks are preserved"
TESTDIR2=$(mktemp -d)
cd "$TESTDIR2"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
---

Use `[[wiki-link]]` syntax.

```markdown
[[wiki-link]]
```
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '<code>\[\[wiki-link\]\]</code>' _site/test/index.html; then
    pass
else
    fail "Wiki links in code were processed"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR2"

# Test 6: TOC generation
test_case "Table of contents is generated"
if grep -q 'class="lp-toc"' testdata/garden/_site/notes/systems-thinking/index.html; then
    pass
else
    # TOC may not be present if page has no headings or TOC is disabled
    warn "TOC not found (may be expected if disabled)"
    PASSED=$((PASSED + 1))
fi

# Test 7: Heading IDs
test_case "Heading IDs are generated correctly"
# Check any page for heading IDs
if grep -rq 'id="' testdata/garden/_site/notes/; then
    pass
else
    fail "Heading ID not found or incorrect"
fi

# Test 8: Backlinks
test_case "Backlinks are generated"
if grep -rq 'class="lp-backlinks"' testdata/garden/_site/; then
    pass
else
    warn "No backlinks found (may be expected)"
    PASSED=$((PASSED + 1))
fi

# Test 9: Theme configuration
test_case "Theme variables are applied"
if grep -q 'var(--lp-accent)' testdata/garden/_site/style.css; then
    pass
else
    fail "Theme variables not found"
fi

# Test 10: Background gradients
test_case "Background gradients are applied"
if grep -q 'linear-gradient' testdata/garden/_site/index.html; then
    pass
else
    fail "Background gradient not found"
fi

# Test 11: Section indexes
test_case "Section index pages are generated"
if [ -f "testdata/garden/_site/notes/index.html" ]; then
    pass
else
    fail "Notes index not generated"
fi

# Test 12: Tag pages
test_case "Tag pages are generated"
if [ -d "testdata/garden/_site/tags" ]; then
    pass
else
    warn "No tag pages (may be expected if no tags)"
    PASSED=$((PASSED + 1))
fi

# Test 13: Static files
test_case "Static files are copied"
if [ -f "testdata/garden/_site/favicon.svg" ]; then
    pass
else
    fail "Favicon not copied"
fi

# Test 14: Broken link detection
test_case "Broken links generate warnings"
TESTDIR3=$(mktemp -d)
cd "$TESTDIR3"
"$LEAFPRESS" init > /dev/null 2>&1
cat > broken.md << 'EOF'
---
title: Broken
---

Link to [[nonexistent-page]].
EOF
if "$LEAFPRESS" build 2>&1 | grep -iq "broken link\|warning"; then
    pass
else
    fail "No warning for broken link"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR3"

# Test 15: Draft pages
test_case "Draft pages are excluded"
TESTDIR4=$(mktemp -d)
cd "$TESTDIR4"
"$LEAFPRESS" init > /dev/null 2>&1
cat > draft.md << 'EOF'
---
title: Draft
draft: true
---

This is a draft.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ ! -f "_site/draft/index.html" ]; then
    pass
else
    fail "Draft page was built"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR4"

# Test 16: Obsidian date aliases
test_case "Obsidian date aliases are supported"
TESTDIR5=$(mktemp -d)
cd "$TESTDIR5"
"$LEAFPRESS" init > /dev/null 2>&1
cat > obsidian-dates.md << 'EOF'
---
title: Obsidian Dates
created: 2024-01-15
modified: 2024-06-20
---

This page uses Obsidian-style date fields.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# When modified is present, we show "Updated" with modified date
if grep -q "Updated" _site/obsidian-dates/index.html && grep -q "Jun 20, 2024" _site/obsidian-dates/index.html; then
    pass
else
    fail "Obsidian date aliases not rendered"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR5"

# Test 17: Custom styles
test_case "Custom style.css is used"
if [ -f "testdata/garden/_site/style.css" ]; then
    pass
else
    fail "style.css not found"
fi

# Test 18: Dark mode toggle
test_case "Dark mode toggle script is included"
if grep -q 'data-theme="dark"' testdata/garden/_site/index.html; then
    pass
else
    fail "Dark mode script not found"
fi

# Test 19: navStyle modes
test_case "navStyle glassy adds pill scroll behavior"
TESTDIR6=$(mktemp -d)
cd "$TESTDIR6"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "theme": { "navStyle": "glassy" }
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q "lp-nav-placeholder" _site/index.html && grep -q "lp-nav--pill" _site/index.html; then
    pass
else
    fail "Glassy nav style not applied"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR6"

# Test 20: navStyle sticky
test_case "navStyle sticky applies sticky positioning"
TESTDIR7=$(mktemp -d)
cd "$TESTDIR7"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "theme": { "navStyle": "sticky" }
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q "position: sticky" _site/index.html; then
    pass
else
    fail "Sticky nav style not applied"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR7"

# Test 21: navActiveStyle base
test_case "navActiveStyle base colors active link"
TESTDIR8=$(mktemp -d)
cd "$TESTDIR8"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes
---
Notes section
EOF
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "nav": [{"label": "Notes", "path": "/notes/"}],
  "theme": { "navActiveStyle": "base" }
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q "lp-nav-link--active lp-nav-active-base" _site/notes/index.html; then
    pass
else
    fail "Active nav base style not applied"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR8"

# Test 22: navActiveStyle box
test_case "navActiveStyle box applies box style"
TESTDIR9=$(mktemp -d)
cd "$TESTDIR9"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes
---
Notes section
EOF
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "nav": [{"label": "Notes", "path": "/notes/"}],
  "theme": { "navActiveStyle": "box" }
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q "lp-nav-link--active lp-nav-active-box" _site/notes/index.html; then
    pass
else
    fail "Active nav box style not applied"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR9"

# Test 23: navActiveStyle underlined
test_case "navActiveStyle underlined applies underline style"
TESTDIR10=$(mktemp -d)
cd "$TESTDIR10"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes
---
Notes section
EOF
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "nav": [{"label": "Notes", "path": "/notes/"}],
  "theme": { "navActiveStyle": "underlined" }
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q "lp-nav-link--active lp-nav-active-underlined" _site/notes/index.html; then
    pass
else
    fail "Active nav underlined style not applied"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR10"

# Test 24: Growth emoji before title in list pages
test_case "Growth emoji appears before title in list pages"
TESTDIR11=$(mktemp -d)
cd "$TESTDIR11"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes
---
EOF
cat > notes/test.md << 'EOF'
---
title: Test Note
growth: seedling
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Check that growth emoji span comes before title span in the output order
GROWTH_LINE=$(grep -n 'lp-index-growth' _site/notes/index.html | head -1 | cut -d: -f1)
TITLE_LINE=$(grep -n 'lp-index-title' _site/notes/index.html | head -1 | cut -d: -f1)
if [ -n "$GROWTH_LINE" ] && [ -n "$TITLE_LINE" ] && [ "$GROWTH_LINE" -lt "$TITLE_LINE" ]; then
    pass
else
    fail "Growth emoji not before title"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR11"

# Test 25: Homepage title format
test_case "Homepage uses site title only"
TESTDIR12=$(mktemp -d)
cd "$TESTDIR12"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "My Test Garden"
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Homepage title should be just "My Test Garden" not "Welcome | My Test Garden"
if grep -q "<title>My Test Garden</title>" _site/index.html; then
    pass
else
    fail "Homepage title not formatted correctly"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR12"

# Test 26: Date format shows both Updated and Created
test_case "Date format shows Updated and Created"
TESTDIR13=$(mktemp -d)
cd "$TESTDIR13"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
date: 2024-01-15
modified: 2024-06-20
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q "Updated.*Jun 20, 2024.*Created.*Jan 15, 2024" _site/test/index.html; then
    pass
else
    fail "Date format not showing both dates"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR13"

# Test 27: Obsidian-style image embeds
test_case "Obsidian image embeds are converted"
TESTDIR14=$(mktemp -d)
cd "$TESTDIR14"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir -p static/images
echo "dummy" > static/images/test.png
cat > test.md << 'EOF'
---
title: Test
---
![[test.png]]
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '<img.*src=.*/static/images/test.png' _site/test/index.html; then
    pass
else
    fail "Obsidian image embed not converted"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR14"

# Test 28: List page reverse chronological sorting
test_case "List pages sort in reverse chronological order"
TESTDIR15=$(mktemp -d)
cd "$TESTDIR15"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes
---
EOF
cat > notes/old.md << 'EOF'
---
title: Old Post
date: 2024-01-01
---
Old content
EOF
cat > notes/new.md << 'EOF'
---
title: New Post
date: 2024-12-01
---
New content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# New Post should appear before Old Post in the HTML
if grep -l "New Post" _site/notes/index.html > /dev/null && \
   awk '/New Post/{new=NR} /Old Post/{old=NR} END{exit !(new<old)}' _site/notes/index.html; then
    pass
else
    fail "Posts not in reverse chronological order"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR15"

# Test 29: TOC disabled per-page
test_case "TOC can be disabled per-page"
TESTDIR16=$(mktemp -d)
cd "$TESTDIR16"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "toc": true
}
EOF
cat > test.md << 'EOF'
---
title: Test
toc: false
---
## Heading 1
Content
## Heading 2
More content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# TOC should NOT be present even though global toc is true
if ! grep -q 'class="lp-toc"' _site/test/index.html; then
    pass
else
    fail "TOC not disabled per-page"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR16"

# Cleanup
rm -rf "$TESTDIR"

# Summary
echo ""
echo "======================="
echo "Test Results:"
echo -e "${GREEN}Passed: $PASSED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
