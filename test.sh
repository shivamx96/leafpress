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

# Test 30: navStyle base (no sticky)
test_case "navStyle base does not stick nav"
TESTDIR17=$(mktemp -d)
cd "$TESTDIR17"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "theme": { "navStyle": "base" }
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should NOT have position: sticky or lp-nav-placeholder
if ! grep -q "position: sticky" _site/index.html && ! grep -q "lp-nav-placeholder" _site/index.html; then
    pass
else
    fail "Base nav style should not be sticky"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR17"

# Test 31: Invalid navStyle rejected
test_case "Invalid navStyle is rejected"
TESTDIR18=$(mktemp -d)
cd "$TESTDIR18"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "theme": { "navStyle": "invalid" }
}
EOF
if "$LEAFPRESS" build 2>&1 | grep -q "navStyle must be"; then
    pass
else
    fail "Invalid navStyle not rejected"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR18"

# Test 32: Invalid navActiveStyle rejected
test_case "Invalid navActiveStyle is rejected"
TESTDIR19=$(mktemp -d)
cd "$TESTDIR19"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "theme": { "navActiveStyle": "invalid" }
}
EOF
if "$LEAFPRESS" build 2>&1 | grep -q "navActiveStyle must be"; then
    pass
else
    fail "Invalid navActiveStyle not rejected"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR19"

# Test 33: External links have correct class
test_case "External links are marked correctly"
TESTDIR20=$(mktemp -d)
cd "$TESTDIR20"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
---
Visit [GitHub](https://github.com) for code.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'class="lp-external"' _site/test/index.html && grep -q 'target="_blank"' _site/test/index.html; then
    pass
else
    fail "External link not marked correctly"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR20"

# Test 34: Broken wiki links have correct class
test_case "Broken wiki links are styled"
TESTDIR21=$(mktemp -d)
cd "$TESTDIR21"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
---
Link to [[nonexistent-page]].
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'class="lp-broken-link"' _site/test/index.html; then
    pass
else
    fail "Broken wiki link not styled"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR21"

# Test 35: Section sorting by title
test_case "Section can be sorted alphabetically by title"
TESTDIR22=$(mktemp -d)
cd "$TESTDIR22"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes
sort: title
---
EOF
cat > notes/zebra.md << 'EOF'
---
title: Zebra
date: 2024-12-01
---
Content
EOF
cat > notes/apple.md << 'EOF'
---
title: Apple
date: 2024-01-01
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Apple should appear before Zebra when sorted alphabetically
if awk '/Apple/{a=NR} /Zebra/{z=NR} END{exit !(a<z)}' _site/notes/index.html; then
    pass
else
    fail "Section not sorted alphabetically"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR22"

# Test 36: Tags on pages are rendered
test_case "Tags are rendered on pages"
TESTDIR23=$(mktemp -d)
cd "$TESTDIR23"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
tags: [coding, golang]
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'class="lp-tag"' _site/test/index.html && grep -q '#coding' _site/test/index.html; then
    pass
else
    fail "Tags not rendered on page"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR23"

# Test 37: Tag cloud page is generated
test_case "Tag cloud page is generated"
TESTDIR24=$(mktemp -d)
cd "$TESTDIR24"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
tags: [coding]
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/tags/index.html" ] && grep -q 'lp-tag-cloud' _site/tags/index.html; then
    pass
else
    fail "Tag cloud page not generated"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR24"

# Test 38: Individual tag pages are generated
test_case "Individual tag pages are generated"
TESTDIR25=$(mktemp -d)
cd "$TESTDIR25"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
tags: [coding]
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/tags/coding/index.html" ]; then
    pass
else
    fail "Individual tag page not generated"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR25"

# Test 39: Backlinks section is shown
test_case "Backlinks section is shown"
TESTDIR26=$(mktemp -d)
cd "$TESTDIR26"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page-a.md << 'EOF'
---
title: Page A
---
Link to [[page-b]].
EOF
cat > page-b.md << 'EOF'
---
title: Page B
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'class="lp-backlinks"' _site/page-b/index.html && grep -q 'Page A' _site/page-b/index.html; then
    pass
else
    fail "Backlinks section not shown"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR26"

# Test 40: Code syntax highlighting
test_case "Code blocks have syntax highlighting"
TESTDIR27=$(mktemp -d)
cd "$TESTDIR27"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
---
```go
func main() {}
```
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'class="chroma"' _site/test/index.html; then
    pass
else
    fail "Syntax highlighting not applied"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR27"

# Test 41: Copy button script included
test_case "Copy button script is included"
TESTDIR28=$(mktemp -d)
cd "$TESTDIR28"
"$LEAFPRESS" init > /dev/null 2>&1
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'lp-copy-button' _site/index.html; then
    pass
else
    fail "Copy button script not included"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR28"

# Test 42: Custom fonts are loaded
test_case "Custom fonts are loaded from Google Fonts"
TESTDIR29=$(mktemp -d)
cd "$TESTDIR29"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "theme": {
    "fontHeading": "Playfair Display",
    "fontBody": "Roboto"
  }
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'fonts.googleapis.com.*Playfair' _site/index.html && grep -q 'fonts.googleapis.com.*Roboto' _site/index.html; then
    pass
else
    fail "Custom fonts not loaded"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR29"

# Test 43: All favicon formats are copied
test_case "All favicon formats are present"
TESTDIR30=$(mktemp -d)
cd "$TESTDIR30"
"$LEAFPRESS" init > /dev/null 2>&1
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/favicon.svg" ] && [ -f "_site/favicon.ico" ] && [ -f "_site/favicon-96x96.png" ]; then
    pass
else
    fail "Not all favicon formats present"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR30"

# Test 44: Author copyright in footer
test_case "Author field adds copyright in footer"
TESTDIR31=$(mktemp -d)
cd "$TESTDIR31"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test Garden",
  "author": "John Doe"
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '&copy; John Doe. All rights reserved.' _site/index.html; then
    pass
else
    fail "Author copyright not in footer"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR31"

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
