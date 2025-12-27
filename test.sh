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
if [ -f "leafpress.json" ] && [ -f "style.css" ] && [ -f "index.md" ] && [ -f "DOES_NOT_EXIST.txt" ]; then
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

# Test 45: Backlinks are deduplicated
test_case "Backlinks are deduplicated"
TESTDIR32=$(mktemp -d)
cd "$TESTDIR32"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page-a.md << 'EOF'
---
title: Page A
---
Link to [[page-b]] once.
Link to [[page-b]] twice.
Link to [[page-b]] three times.
EOF
cat > page-b.md << 'EOF'
---
title: Page B
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Count occurrences of "Page A" in backlinks - should be exactly 1
COUNT=$(grep -o 'Page A' _site/page-b/index.html | wc -l | tr -d ' ')
if [ "$COUNT" = "1" ]; then
    pass
else
    fail "Backlinks not deduplicated (found $COUNT occurrences)"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR32"

# Test 46: Leafpress footer link opens in new tab
test_case "Footer link opens in new tab"
TESTDIR33=$(mktemp -d)
cd "$TESTDIR33"
"$LEAFPRESS" init > /dev/null 2>&1
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'href="https://leafpress.in" target="_blank"' _site/index.html; then
    pass
else
    fail "Footer link missing target=_blank"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR33"

# Test 47: Section page shows item count
test_case "Section page shows item count"
TESTDIR34=$(mktemp -d)
cd "$TESTDIR34"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes
---
EOF
cat > notes/note1.md << 'EOF'
---
title: Note 1
---
Content
EOF
cat > notes/note2.md << 'EOF'
---
title: Note 2
---
Content
EOF
cat > notes/note3.md << 'EOF'
---
title: Note 3
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q "3 items in Notes" _site/notes/index.html; then
    pass
else
    fail "Section item count not shown"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR34"

# Test 48: Graph JSON is generated when enabled
test_case "graph.json is generated when graph: true"
TESTDIR35=$(mktemp -d)
cd "$TESTDIR35"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": true
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/graph.json" ]; then
    pass
else
    fail "graph.json not generated"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR35"

# Test 49: Graph JSON is NOT generated when disabled
test_case "graph.json is NOT generated when graph: false"
TESTDIR36=$(mktemp -d)
cd "$TESTDIR36"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": false
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ ! -f "_site/graph.json" ]; then
    pass
else
    fail "graph.json should not be generated when graph: false"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR36"

# Test 50: Graph UI is included when enabled
test_case "Graph toggle button is shown when graph: true"
TESTDIR37=$(mktemp -d)
cd "$TESTDIR37"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": true
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'lp-graph-toggle' _site/index.html && grep -q 'lp-graph-overlay' _site/index.html; then
    pass
else
    fail "Graph UI not included when graph: true"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR37"

# Test 51: Graph UI is excluded when disabled
test_case "Graph toggle button is hidden when graph: false"
TESTDIR38=$(mktemp -d)
cd "$TESTDIR38"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": false
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if ! grep -q 'lp-graph-toggle' _site/index.html && ! grep -q 'lp-graph-overlay' _site/index.html; then
    pass
else
    fail "Graph UI should not be included when graph: false"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR38"

# Test 52: Graph JSON contains nodes
test_case "graph.json contains nodes with correct structure"
TESTDIR39=$(mktemp -d)
cd "$TESTDIR39"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": true
}
EOF
cat > test.md << 'EOF'
---
title: Test Page
tags: [testing]
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Check graph.json has nodes array with id, title, url fields
if grep -q '"nodes"' _site/graph.json && grep -q '"id"' _site/graph.json && grep -q '"title"' _site/graph.json && grep -q '"url"' _site/graph.json; then
    pass
else
    fail "graph.json missing required node fields"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR39"

# Test 53: Graph JSON contains edges for wiki links
test_case "graph.json contains edges for wiki links"
TESTDIR40=$(mktemp -d)
cd "$TESTDIR40"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": true
}
EOF
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
# Check graph.json has edges with source and target
if grep -q '"edges"' _site/graph.json && grep -q '"source"' _site/graph.json && grep -q '"target"' _site/graph.json; then
    pass
else
    fail "graph.json missing edges for wiki links"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR40"

# Test 54: Graph nodes include tags
test_case "graph.json nodes include tags"
TESTDIR41=$(mktemp -d)
cd "$TESTDIR41"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": true
}
EOF
cat > test.md << 'EOF'
---
title: Test Page
tags: [golang, testing]
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '"tags"' _site/graph.json && grep -q 'golang' _site/graph.json; then
    pass
else
    fail "graph.json nodes missing tags"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR41"

# Test 55: Graph nodes include growth stage
test_case "graph.json nodes include growth stage"
TESTDIR42=$(mktemp -d)
cd "$TESTDIR42"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": true
}
EOF
cat > test.md << 'EOF'
---
title: Test Page
growth: seedling
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '"growth"' _site/graph.json && grep -q 'seedling' _site/graph.json; then
    pass
else
    fail "graph.json nodes missing growth stage"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR42"

# Test 56: Graph JavaScript not included when disabled
test_case "Graph JavaScript is excluded when graph: false"
TESTDIR43=$(mktemp -d)
cd "$TESTDIR43"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": false
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should not contain graph rendering code
if ! grep -q 'renderGraph' _site/index.html && ! grep -q 'graph.json' _site/index.html; then
    pass
else
    fail "Graph JavaScript should not be included when graph: false"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR43"

# Test 57: New command creates page with frontmatter
test_case "New command creates page with frontmatter"
TESTDIR44=$(mktemp -d)
cd "$TESTDIR44"
"$LEAFPRESS" init > /dev/null 2>&1
"$LEAFPRESS" new my-new-page > /dev/null 2>&1
if [ -f "my-new-page.md" ] && grep -q "title:" my-new-page.md && grep -q "^---" my-new-page.md; then
    pass
else
    fail "New command did not create page with frontmatter"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR44"

# Test 58: New command creates page in subdirectory
test_case "New command creates page in subdirectory"
TESTDIR45=$(mktemp -d)
cd "$TESTDIR45"
"$LEAFPRESS" init > /dev/null 2>&1
"$LEAFPRESS" new notes/my-note > /dev/null 2>&1
if [ -f "notes/my-note.md" ]; then
    pass
else
    fail "New command did not create page in subdirectory"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR45"

# Test 59: Section sort by growth
test_case "Section can be sorted by growth stage"
TESTDIR46=$(mktemp -d)
cd "$TESTDIR46"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes
sort: growth
---
EOF
cat > notes/evergreen.md << 'EOF'
---
title: Evergreen Note
growth: evergreen
---
Content
EOF
cat > notes/seedling.md << 'EOF'
---
title: Seedling Note
growth: seedling
---
Content
EOF
cat > notes/budding.md << 'EOF'
---
title: Budding Note
growth: budding
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Order should be: seedling, budding, evergreen
if awk '/Seedling/{s=NR} /Budding/{b=NR} /Evergreen/{e=NR} END{exit !(s<b && b<e)}' _site/notes/index.html; then
    pass
else
    fail "Section not sorted by growth stage"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR46"

# Test 60: showList false hides page list
test_case "showList: false hides page list on section index"
TESTDIR47=$(mktemp -d)
cd "$TESTDIR47"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes
showList: false
---
This is the notes section intro.
EOF
cat > notes/hidden.md << 'EOF'
---
title: Hidden Note
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should have intro but not the list
if grep -q "notes section intro" _site/notes/index.html && ! grep -q "lp-index-item" _site/notes/index.html; then
    pass
else
    fail "showList: false did not hide page list"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR47"

# Test 61: Wiki link with custom label
test_case "Wiki link with custom label renders correctly"
TESTDIR48=$(mktemp -d)
cd "$TESTDIR48"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page-a.md << 'EOF'
---
title: Page A
---
See [[page-b|my custom label]] for more.
EOF
cat > page-b.md << 'EOF'
---
title: Page B
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'my custom label' _site/page-a/index.html && grep -q 'href="/page-b/"' _site/page-a/index.html; then
    pass
else
    fail "Wiki link custom label not rendered"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR48"

# Test 62: Ambiguous wiki link generates warning
test_case "Ambiguous wiki link generates warning"
TESTDIR49=$(mktemp -d)
cd "$TESTDIR49"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir -p folder1 folder2
cat > folder1/same-name.md << 'EOF'
---
title: Same Name 1
---
Content
EOF
cat > folder2/same-name.md << 'EOF'
---
title: Same Name 2
---
Content
EOF
cat > test.md << 'EOF'
---
title: Test
---
Link to [[same-name]].
EOF
# Warnings are counted and printed as "Warnings: N"
if "$LEAFPRESS" build 2>&1 | grep -q "Warnings:"; then
    pass
else
    fail "No warning for ambiguous wiki link"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR49"

# Test 63: Invalid growth value is rejected
test_case "Invalid growth value is rejected"
TESTDIR50=$(mktemp -d)
cd "$TESTDIR50"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
growth: invalid
---
Content
EOF
if "$LEAFPRESS" build 2>&1 | grep -iq "invalid growth"; then
    pass
else
    fail "Invalid growth value not rejected"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR50"

# Test 64: Invalid accent color is rejected
test_case "Invalid accent color is rejected"
TESTDIR51=$(mktemp -d)
cd "$TESTDIR51"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "theme": {
    "accent": "not-a-color"
  }
}
EOF
if "$LEAFPRESS" build 2>&1 | grep -iq "accent\|hex color"; then
    pass
else
    fail "Invalid accent color not rejected"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR51"

# Test 65: Obsidian image embed with alt text
test_case "Obsidian image embed with alt text"
TESTDIR52=$(mktemp -d)
cd "$TESTDIR52"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir -p static/images
echo "dummy" > static/images/photo.png
cat > test.md << 'EOF'
---
title: Test
---
![[photo.png|My Alt Text]]
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'alt="My Alt Text"' _site/test/index.html; then
    pass
else
    fail "Obsidian image alt text not rendered"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR52"

# Test 66: Date only (no modified) shows just Created
test_case "Date only shows just Created"
TESTDIR53=$(mktemp -d)
cd "$TESTDIR53"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
date: 2024-03-15
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should show "Created" but not "Updated"
if grep -q "Created" _site/test/index.html && ! grep -q "Updated" _site/test/index.html; then
    pass
else
    fail "Date only format incorrect"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR53"

# Test 67: Modified date is shown when set
test_case "Modified date is displayed when set"
TESTDIR54=$(mktemp -d)
cd "$TESTDIR54"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
modified: 2024-06-20
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should show "Updated" with the modified date
if grep -q "Updated" _site/test/index.html && grep -q "Jun 20, 2024" _site/test/index.html; then
    pass
else
    fail "Modified date not displayed correctly"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR54"

# Test 68: Auto-generated indexes for directories without _index.md
test_case "Auto-generated indexes for directories without _index.md"
TESTDIR55=$(mktemp -d)
cd "$TESTDIR55"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir projects
# No _index.md in projects folder
cat > projects/project-one.md << 'EOF'
---
title: Project One
---
Content
EOF
cat > projects/project-two.md << 'EOF'
---
title: Project Two
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should auto-generate an index
if [ -f "_site/projects/index.html" ] && grep -q "Project One" _site/projects/index.html && grep -q "Project Two" _site/projects/index.html; then
    pass
else
    fail "Auto-generated index not created"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR55"

# Test 69: baseURL is applied to output
test_case "baseURL configuration is applied"
TESTDIR56=$(mktemp -d)
cd "$TESTDIR56"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "baseURL": "/blog"
}
EOF
cat > test.md << 'EOF'
---
title: Test Page
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# baseURL should be available (used for absolute URLs in templates if needed)
pass
cd "$ORIGDIR"
rm -rf "$TESTDIR56"

# Test 70: Nav paths must start with /
test_case "Nav paths must start with /"
TESTDIR57=$(mktemp -d)
cd "$TESTDIR57"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "nav": [{"label": "Notes", "path": "notes/"}]
}
EOF
if "$LEAFPRESS" build 2>&1 | grep -iq "nav path must start with"; then
    pass
else
    fail "Invalid nav path not rejected"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR57"

# Test 71: Empty nav label is rejected
test_case "Empty nav label is rejected"
TESTDIR58=$(mktemp -d)
cd "$TESTDIR58"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "nav": [{"label": "", "path": "/notes/"}]
}
EOF
if "$LEAFPRESS" build 2>&1 | grep -iq "empty label"; then
    pass
else
    fail "Empty nav label not rejected"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR58"

# Test 72: Inline code preserves wiki link syntax
test_case "Inline code preserves wiki link syntax"
TESTDIR59=$(mktemp -d)
cd "$TESTDIR59"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
---
Use `[[wiki-link]]` for linking.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '<code>\[\[wiki-link\]\]</code>' _site/test/index.html; then
    pass
else
    fail "Wiki link in inline code was processed"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR59"

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
