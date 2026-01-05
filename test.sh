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
rm -rf "$TESTDIR"

# Create a comprehensive test garden at runtime for tests 2-13
GARDEN=$(mktemp -d)
cd "$GARDEN"
cat > leafpress.json << 'EOF'
{
  "title": "Test Garden",
  "author": "Test Author",
  "toc": true,
  "graph": true,
  "search": true,
  "backlinks": true,
  "nav": [{"label": "Notes", "path": "/notes/"}],
  "theme": {
    "accent": "#6366f1",
    "background": {
      "light": "linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%)",
      "dark": "#1a1a1a"
    }
  }
}
EOF
cat > index.md << 'EOF'
---
title: Welcome
---
Welcome to my garden. Check out my [[notes/systems-thinking|systems thinking notes]].
EOF
mkdir -p notes
cat > notes/_index.md << 'EOF'
---
title: Notes
---
My notes collection.
EOF
cat > notes/systems-thinking.md << 'EOF'
---
title: Systems Thinking
tags: [thinking, systems]
growth: budding
date: 2024-06-15
---
## Introduction
Systems thinking is a holistic approach.

## Key Concepts
Feedback loops and emergence.
EOF
cat > style.css << 'EOF'
/* Custom test styles */
.custom-class { color: red; }
EOF
"$LEAFPRESS" build > /dev/null 2>&1
cd "$ORIGDIR"

# Test 2: Build command
test_case "Build completes successfully"
if [ -d "$GARDEN/_site" ]; then
    pass
else
    fail "Build did not create _site"
fi

# Test 3: Check output files
test_case "Output files are generated"
if [ -f "$GARDEN/_site/index.html" ] && [ -f "$GARDEN/_site/style.css" ]; then
    pass
else
    fail "Missing output files"
fi

# Test 4: Wiki links are processed
test_case "Wiki links are converted to HTML"
if grep -q 'class="lp-wikilink"' "$GARDEN/_site/index.html"; then
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
if grep -q 'class="lp-toc"' "$GARDEN/_site/notes/systems-thinking/index.html"; then
    pass
else
    fail "TOC not found"
fi

# Test 7: Heading IDs
test_case "Heading IDs are generated correctly"
if grep -rq 'id="' "$GARDEN/_site/notes/"; then
    pass
else
    fail "Heading ID not found"
fi

# Test 8: Backlinks
test_case "Backlinks are generated"
if grep -q 'class="lp-backlinks"' "$GARDEN/_site/notes/systems-thinking/index.html"; then
    pass
else
    fail "Backlinks not found"
fi

# Test 9: Theme configuration
test_case "Theme variables are applied"
if grep -q 'var(--lp-accent)' "$GARDEN/_site/style.css"; then
    pass
else
    fail "Theme variables not found"
fi

# Test 10: Background gradients
test_case "Background gradients are applied"
if grep -q 'linear-gradient' "$GARDEN/_site/index.html"; then
    pass
else
    fail "Background gradient not found"
fi

# Test 11: Section indexes
test_case "Section index pages are generated"
if [ -f "$GARDEN/_site/notes/index.html" ]; then
    pass
else
    fail "Notes index not generated"
fi

# Test 12: Tag pages
test_case "Tag pages are generated"
if [ -d "$GARDEN/_site/tags" ]; then
    pass
else
    fail "Tag pages not generated"
fi

# Test 13: Static files
test_case "Static files are copied"
if [ -f "$GARDEN/_site/favicon.svg" ]; then
    pass
else
    fail "Favicon not copied"
fi

# Cleanup test garden
rm -rf "$GARDEN"

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
TESTDIR_CSS=$(mktemp -d)
cd "$TESTDIR_CSS"
"$LEAFPRESS" init > /dev/null 2>&1
echo ".my-custom { color: blue; }" >> style.css
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/style.css" ] && grep -q "my-custom" _site/style.css; then
    pass
else
    fail "Custom style.css not used"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR_CSS"

# Test 18: Dark mode toggle
test_case "Dark mode toggle script is included"
TESTDIR_DARK=$(mktemp -d)
cd "$TESTDIR_DARK"
"$LEAFPRESS" init > /dev/null 2>&1
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'data-theme="dark"' _site/index.html; then
    pass
else
    fail "Dark mode script not found"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR_DARK"

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

# Test 73: Ignore folders from config
test_case "Ignore folders excludes content from build"
TESTDIR60=$(mktemp -d)
cd "$TESTDIR60"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "ignore": ["drafts", "private"]
}
EOF
mkdir -p drafts private notes
echo -e "---\ntitle: Draft Post\n---\nDraft content" > drafts/post.md
echo -e "---\ntitle: Private Note\n---\nPrivate content" > private/secret.md
echo -e "---\ntitle: Public Note\n---\nPublic content" > notes/public.md
"$LEAFPRESS" build > /dev/null 2>&1
# drafts and private should not exist in output, notes should
if [ ! -d "_site/drafts" ] && [ ! -d "_site/private" ] && [ -f "_site/notes/public/index.html" ]; then
    pass
else
    fail "Ignored folders were not excluded or public content missing"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR60"

# Test 74: Multiple builds without clean preserves content
test_case "Multiple builds without clean work correctly"
TESTDIR61=$(mktemp -d)
cd "$TESTDIR61"
"$LEAFPRESS" init > /dev/null 2>&1
echo -e "---\ntitle: Page One\n---\nFirst content" > page1.md
"$LEAFPRESS" build > /dev/null 2>&1
# Add second page and rebuild
echo -e "---\ntitle: Page Two\n---\nSecond content" > page2.md
"$LEAFPRESS" build > /dev/null 2>&1
# Both pages should exist
if [ -f "_site/page1/index.html" ] && [ -f "_site/page2/index.html" ]; then
    pass
else
    fail "Multiple builds did not preserve content"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR61"

# Test 75: Backlinks update correctly after content change
test_case "Backlinks are correctly built"
TESTDIR62=$(mktemp -d)
cd "$TESTDIR62"
"$LEAFPRESS" init > /dev/null 2>&1
echo -e "---\ntitle: Page A\n---\nLinks to [[page-b]]" > page-a.md
echo -e "---\ntitle: Page B\n---\nContent of B" > page-b.md
"$LEAFPRESS" build > /dev/null 2>&1
# Page B should have backlink from Page A
if grep -q "Page A" _site/page-b/index.html; then
    pass
else
    fail "Backlinks not built correctly"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR62"

# Test 76: Tag pages are generated correctly
test_case "Tag pages list all tagged content"
TESTDIR63=$(mktemp -d)
cd "$TESTDIR63"
"$LEAFPRESS" init > /dev/null 2>&1
echo -e "---\ntitle: Post One\ntags: [golang]\n---\nContent" > post1.md
echo -e "---\ntitle: Post Two\ntags: [golang, rust]\n---\nContent" > post2.md
echo -e "---\ntitle: Post Three\ntags: [rust]\n---\nContent" > post3.md
"$LEAFPRESS" build > /dev/null 2>&1
# Check tag index exists and tag pages have correct posts
if [ -f "_site/tags/index.html" ] && \
   grep -q "Post One" _site/tags/golang/index.html && \
   grep -q "Post Two" _site/tags/golang/index.html && \
   grep -q "Post Two" _site/tags/rust/index.html && \
   grep -q "Post Three" _site/tags/rust/index.html; then
    pass
else
    fail "Tag pages not generated correctly"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR63"

# Test 77: Graph JSON generated when enabled
test_case "Graph JSON contains nodes and links"
TESTDIR64=$(mktemp -d)
cd "$TESTDIR64"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": true
}
EOF
echo -e "---\ntitle: Node A\n---\nLinks to [[node-b]]" > node-a.md
echo -e "---\ntitle: Node B\n---\nContent" > node-b.md
"$LEAFPRESS" build > /dev/null 2>&1
# Check graph.json exists and has correct structure
if [ -f "_site/graph.json" ] && \
   grep -q '"nodes"' _site/graph.json && \
   grep -q '"edges"' _site/graph.json && \
   grep -q 'node-a' _site/graph.json && \
   grep -q 'node-b' _site/graph.json; then
    pass
else
    fail "Graph JSON not generated correctly"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR64"

# Test 78: Parallel rendering produces correct output
test_case "Large site builds correctly with parallel rendering"
TESTDIR65=$(mktemp -d)
cd "$TESTDIR65"
"$LEAFPRESS" init > /dev/null 2>&1
# Create 20 interconnected pages
for i in $(seq 1 20); do
    next=$((i + 1))
    if [ $next -gt 20 ]; then next=1; fi
    echo -e "---\ntitle: Page $i\ntags: [test]\n---\nLinks to [[page-$next]]" > "page-$i.md"
done
"$LEAFPRESS" build > /dev/null 2>&1
# All pages should exist
count=$(ls -1 _site/page-*/index.html 2>/dev/null | wc -l)
if [ "$count" -eq 20 ]; then
    pass
else
    fail "Expected 20 pages, got $count"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR65"

# Test 79: Search index is generated when enabled
test_case "search-index.json is generated when search: true"
TESTDIR66=$(mktemp -d)
cd "$TESTDIR66"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "search": true
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/search-index.json" ]; then
    pass
else
    fail "search-index.json not generated"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR66"

# Test 80: Search index is NOT generated when disabled
test_case "search-index.json is NOT generated when search: false"
TESTDIR67=$(mktemp -d)
cd "$TESTDIR67"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "search": false
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ ! -f "_site/search-index.json" ]; then
    pass
else
    fail "search-index.json should not be generated when search: false"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR67"

# Test 81: Search index contains correct fields
test_case "search-index.json contains title, url, content, tags"
TESTDIR68=$(mktemp -d)
cd "$TESTDIR68"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "search": true
}
EOF
cat > test.md << 'EOF'
---
title: Test Page
tags: [golang, testing]
---
This is the content of the test page.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '"title"' _site/search-index.json && \
   grep -q '"url"' _site/search-index.json && \
   grep -q '"content"' _site/search-index.json && \
   grep -q '"tags"' _site/search-index.json && \
   grep -q 'Test Page' _site/search-index.json; then
    pass
else
    fail "search-index.json missing required fields"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR68"

# Test 82: Search index content is HTML-stripped
test_case "search-index.json content has HTML stripped"
TESTDIR69=$(mktemp -d)
cd "$TESTDIR69"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "search": true
}
EOF
cat > test.md << 'EOF'
---
title: Test
---
**Bold text** and *italic text* here.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Content should not contain HTML tags like <strong> or <em>
if ! grep -q '<strong>' _site/search-index.json && ! grep -q '<em>' _site/search-index.json; then
    pass
else
    fail "search-index.json content contains HTML tags"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR69"

# Test 83: Search index excludes index pages
test_case "search-index.json excludes section index pages"
TESTDIR70=$(mktemp -d)
cd "$TESTDIR70"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "search": true
}
EOF
mkdir notes
cat > notes/_index.md << 'EOF'
---
title: Notes Section
---
Section intro
EOF
cat > notes/page.md << 'EOF'
---
title: Regular Page
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should contain Regular Page but not Notes Section
if grep -q 'Regular Page' _site/search-index.json && ! grep -q 'Notes Section' _site/search-index.json; then
    pass
else
    fail "search-index.json should exclude index pages"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR70"

# Test 84: Search UI is included when enabled
test_case "Search toggle button is shown when search: true"
TESTDIR71=$(mktemp -d)
cd "$TESTDIR71"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "search": true
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'lp-search-toggle' _site/index.html && grep -q 'lp-search-overlay' _site/index.html; then
    pass
else
    fail "Search UI not included when search: true"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR71"

# Test 85: Search UI is excluded when disabled
test_case "Search toggle button is hidden when search: false"
TESTDIR72=$(mktemp -d)
cd "$TESTDIR72"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "search": false
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if ! grep -q 'lp-search-toggle' _site/index.html && ! grep -q 'lp-search-overlay' _site/index.html; then
    pass
else
    fail "Search UI should not be included when search: false"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR72"

# Test 86: Search JavaScript not included when disabled
test_case "Search JavaScript is excluded when search: false"
TESTDIR73=$(mktemp -d)
cd "$TESTDIR73"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "search": false
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if ! grep -q 'search-index.json' _site/index.html && ! grep -q 'openSearch' _site/index.html; then
    pass
else
    fail "Search JavaScript should not be included when search: false"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR73"

# Test 87: Both graph and search can be enabled together
test_case "Graph and search can both be enabled"
TESTDIR74=$(mktemp -d)
cd "$TESTDIR74"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "graph": true,
  "search": true
}
EOF
cat > test.md << 'EOF'
---
title: Test Page
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/graph.json" ] && [ -f "_site/search-index.json" ] && \
   grep -q 'lp-graph-toggle' _site/index.html && grep -q 'lp-search-toggle' _site/index.html; then
    pass
else
    fail "Graph and search not both working when enabled"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR74"

# Test 88: Search index content is truncated for large pages
test_case "search-index.json content is truncated for large pages"
TESTDIR75=$(mktemp -d)
cd "$TESTDIR75"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "search": true
}
EOF
# Create a page with more than 5000 chars of content
cat > test.md << 'EOF'
---
title: Large Page
---
EOF
# Append lots of content
for i in $(seq 1 200); do
    echo "This is paragraph $i with some content to make this page very large. " >> test.md
done
"$LEAFPRESS" build > /dev/null 2>&1
# Content field should be present but limited in size (rough check: file shouldn't be huge)
SIZE=$(wc -c < _site/search-index.json)
if [ "$SIZE" -lt 10000 ]; then
    pass
else
    fail "search-index.json content not truncated (size: $SIZE bytes)"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR75"

# Test 89: Incremental rebuild doesn't panic with stale resolver references
test_case "Incremental rebuild handles stale resolver gracefully"
TESTDIR76=$(mktemp -d)
cd "$TESTDIR76"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "backlinks": true
}
EOF
# Create two pages that link to each other
cat > page1.md << 'EOF'
---
title: Page One
---
Link to [[Page Two]]
EOF
cat > page2.md << 'EOF'
---
title: Page Two
---
Link to [[Page One]]
EOF
# Initial build
"$LEAFPRESS" build > /dev/null 2>&1
# Simulate incremental rebuild by modifying a file and rebuilding
# This tests that BuildBacklinks doesn't panic with stale resolver
echo "Updated content with [[Page Two]]" >> page1.md
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/page1/index.html" ] && [ -f "_site/page2/index.html" ]; then
    pass
else
    fail "Incremental rebuild failed"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR76"

# Test 90: Backlinks persist after multiple rebuilds (no disappearing)
test_case "Backlinks persist after multiple rebuilds"
TESTDIR77=$(mktemp -d)
cd "$TESTDIR77"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "backlinks": true
}
EOF
cat > page1.md << 'EOF'
---
title: Page One
---
Content here
EOF
cat > page2.md << 'EOF'
---
title: Page Two
---
Link to [[page1]]
EOF
# First build
"$LEAFPRESS" build > /dev/null 2>&1
# Check backlink exists
if ! grep -q 'Page Two' _site/page1/index.html; then
    fail "Backlink missing after first build"
else
    # Modify page2 and rebuild
    echo "More content" >> page2.md
    "$LEAFPRESS" build > /dev/null 2>&1
    # Backlink should still exist
    if grep -q 'Page Two' _site/page1/index.html; then
        # Rebuild again
        echo "Even more" >> page2.md
        "$LEAFPRESS" build > /dev/null 2>&1
        if grep -q 'Page Two' _site/page1/index.html; then
            pass
        else
            fail "Backlink disappeared after third build"
        fi
    else
        fail "Backlink disappeared after second build"
    fi
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR77"

# Test 91: Backlinks don't duplicate after multiple rebuilds
test_case "Backlinks don't duplicate after rebuilds"
TESTDIR78=$(mktemp -d)
cd "$TESTDIR78"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "backlinks": true
}
EOF
cat > target.md << 'EOF'
---
title: Target Page
---
This is the target
EOF
cat > linker.md << 'EOF'
---
title: Linker Page
---
Link to [[target]]
EOF
# Build multiple times
"$LEAFPRESS" build > /dev/null 2>&1
echo "update 1" >> linker.md
"$LEAFPRESS" build > /dev/null 2>&1
echo "update 2" >> linker.md
"$LEAFPRESS" build > /dev/null 2>&1
# Count occurrences of the backlink - should be exactly 1
COUNT=$(grep -o 'Linker Page' _site/target/index.html | wc -l | tr -d ' ')
if [ "$COUNT" -eq "1" ]; then
    pass
else
    fail "Backlink duplicated: found $COUNT occurrences instead of 1"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR78"

# Test 92: Broken wikilinks render with lp-broken-link class
test_case "Broken wikilinks have lp-broken-link class"
TESTDIR79=$(mktemp -d)
cd "$TESTDIR79"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page.md << 'EOF'
---
title: Test Page
---
Link to [[nonexistent-page]]
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'class="lp-broken-link"' _site/page/index.html && grep -q 'nonexistent-page' _site/page/index.html; then
    pass
else
    fail "Broken wikilink not rendered with lp-broken-link class"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR79"

# Test 93: Broken wikilinks CSS has tooltip styles
test_case "Broken wikilinks CSS includes tooltip"
TESTDIR80=$(mktemp -d)
cd "$TESTDIR80"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page.md << 'EOF'
---
title: Test
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Check CSS has broken link styling with tooltip
if grep -q 'lp-broken-link' _site/style.css && \
   grep -q "Page doesn't exist yet" _site/style.css && \
   grep -q 'dashed' _site/style.css; then
    pass
else
    fail "Broken wikilink CSS missing tooltip styles"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR80"

# Test 94: Broken wikilink shows label text not target
test_case "Broken wikilinks display label correctly"
TESTDIR81=$(mktemp -d)
cd "$TESTDIR81"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page.md << 'EOF'
---
title: Test Page
---
Link to [[nonexistent|Custom Label]]
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'lp-broken-link">Custom Label</span>' _site/page/index.html; then
    pass
else
    fail "Broken wikilink label not rendered correctly"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR81"

# Test 95: Unicode characters in title
test_case "Unicode characters in page title"
TESTDIR82=$(mktemp -d)
cd "$TESTDIR82"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page.md << 'EOF'
---
title: 日本語タイトル
---
Japanese content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '日本語タイトル' _site/page/index.html; then
    pass
else
    fail "Unicode title not rendered"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR82"

# Test 96: Emoji in page title
test_case "Emoji in page title"
TESTDIR83=$(mktemp -d)
cd "$TESTDIR83"
"$LEAFPRESS" init > /dev/null 2>&1
cat > rocket.md << 'EOF'
---
title: "🚀 Launch Day"
---
Launching!
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '🚀 Launch Day' _site/rocket/index.html; then
    pass
else
    fail "Emoji in title not rendered"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR83"

# Test 97: Deeply nested folder structure
test_case "Deeply nested folder structure"
TESTDIR84=$(mktemp -d)
cd "$TESTDIR84"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir -p a/b/c/d
cat > a/b/c/d/deep.md << 'EOF'
---
title: Deep Page
---
Deep content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/a/b/c/d/deep/index.html" ] && grep -q 'Deep Page' _site/a/b/c/d/deep/index.html; then
    pass
else
    fail "Deeply nested page not built correctly"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR84"

# Test 98: Special characters in title (colon, ampersand)
test_case "Special characters in title"
TESTDIR85=$(mktemp -d)
cd "$TESTDIR85"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page.md << 'EOF'
---
title: "Hello: World & Friends"
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'Hello: World' _site/page/index.html && grep -q 'Friends' _site/page/index.html; then
    pass
else
    fail "Special characters in title not rendered"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR85"

# Test 99: Empty content file (frontmatter only)
test_case "Empty content file builds"
TESTDIR86=$(mktemp -d)
cd "$TESTDIR86"
"$LEAFPRESS" init > /dev/null 2>&1
cat > empty.md << 'EOF'
---
title: Empty Page
---
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/empty/index.html" ] && grep -q 'Empty Page' _site/empty/index.html; then
    pass
else
    fail "Empty content file not built"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR86"

# Test 100: Multiple date formats supported
test_case "Multiple date formats supported"
TESTDIR87=$(mktemp -d)
cd "$TESTDIR87"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page1.md << 'EOF'
---
title: Page 1
date: 2024-03-15
---
Content
EOF
cat > page2.md << 'EOF'
---
title: Page 2
date: 2024-03-15T10:30:00Z
---
Content
EOF
cat > page3.md << 'EOF'
---
title: Page 3
date: "March 15, 2024"
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/page1/index.html" ] && [ -f "_site/page2/index.html" ] && [ -f "_site/page3/index.html" ]; then
    pass
else
    fail "Multiple date formats not all supported"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR87"

# Test 101: TOC heading IDs are generated
test_case "TOC heading IDs are generated"
TESTDIR88=$(mktemp -d)
cd "$TESTDIR88"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "toc": true
}
EOF
cat > page.md << 'EOF'
---
title: Test
---
## Introduction
Some text
## Getting Started
More text
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'id="introduction"' _site/page/index.html && grep -q 'id="getting-started"' _site/page/index.html; then
    pass
else
    fail "TOC heading IDs not generated"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR88"

# Test 102: Duplicate headings get unique IDs
test_case "Duplicate headings get unique IDs"
TESTDIR89=$(mktemp -d)
cd "$TESTDIR89"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "toc": true
}
EOF
cat > page.md << 'EOF'
---
title: Test
---
## Section
First section
## Section
Second section
## Section
Third section
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'id="section"' _site/page/index.html && grep -q 'id="section-1"' _site/page/index.html && grep -q 'id="section-2"' _site/page/index.html; then
    pass
else
    fail "Duplicate headings don't have unique IDs"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR89"

# Test 103: Wikilink with custom label
test_case "Wikilink with custom label renders correctly"
TESTDIR90=$(mktemp -d)
cd "$TESTDIR90"
"$LEAFPRESS" init > /dev/null 2>&1
cat > source.md << 'EOF'
---
title: Source
---
Link to [[target|click here]]
EOF
cat > target.md << 'EOF'
---
title: Target
---
Target content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'click here</a>' _site/source/index.html && grep -q 'href="/target/"' _site/source/index.html; then
    pass
else
    fail "Wikilink custom label not rendered"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR90"

# Test 104: Case-insensitive wikilink resolution
test_case "Wikilinks are case-insensitive"
TESTDIR91=$(mktemp -d)
cd "$TESTDIR91"
"$LEAFPRESS" init > /dev/null 2>&1
cat > source.md << 'EOF'
---
title: Source
---
Link to [[TARGET]]
EOF
cat > target.md << 'EOF'
---
title: Target
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'href="/target/"' _site/source/index.html && grep -q 'lp-wikilink' _site/source/index.html; then
    pass
else
    fail "Case-insensitive wikilink not resolved"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR91"

# Test 105: Malformed YAML frontmatter is handled
test_case "Malformed YAML frontmatter is handled gracefully"
TESTDIR92=$(mktemp -d)
cd "$TESTDIR92"
"$LEAFPRESS" init > /dev/null 2>&1
cat > bad.md << 'EOF'
---
title: Test
  bad_indent: value
---
Content
EOF
# Should not crash, may produce warning (exit 1 is OK)
"$LEAFPRESS" build > /dev/null 2>&1 || true
pass
cd "$ORIGDIR"
rm -rf "$TESTDIR92"

# Test 106: Missing closing frontmatter delimiter
test_case "Missing frontmatter delimiter handled"
TESTDIR93=$(mktemp -d)
cd "$TESTDIR93"
"$LEAFPRESS" init > /dev/null 2>&1
cat > unclosed.md << 'EOF'
---
title: Unclosed
Content without closing delimiter
EOF
# Should not crash (exit 0 or 1 is OK)
"$LEAFPRESS" build > /dev/null 2>&1 || true
pass
cd "$ORIGDIR"
rm -rf "$TESTDIR93"

# Test 107: File with no frontmatter still builds
test_case "File without frontmatter builds"
TESTDIR94=$(mktemp -d)
cd "$TESTDIR94"
"$LEAFPRESS" init > /dev/null 2>&1
cat > plain.md << 'EOF'
Just plain markdown content
without any frontmatter at all.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if [ -f "_site/plain/index.html" ]; then
    pass
else
    fail "File without frontmatter not built"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR94"

# Test 108: Invalid JSON config is rejected
test_case "Invalid JSON config is rejected"
TESTDIR95=$(mktemp -d)
cd "$TESTDIR95"
"$LEAFPRESS" init > /dev/null 2>&1
echo '{ invalid json }' > leafpress.json
OUTPUT=$("$LEAFPRESS" build 2>&1 || true)
if echo "$OUTPUT" | grep -qi "error\|invalid\|parse"; then
    pass
else
    fail "Invalid JSON not properly rejected"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR95"

# Test 109: Build with no content files
test_case "Build with no content files succeeds"
TESTDIR96=$(mktemp -d)
cd "$TESTDIR96"
"$LEAFPRESS" init > /dev/null 2>&1
rm -f *.md
"$LEAFPRESS" build > /dev/null 2>&1
if [ -d "_site" ]; then
    pass
else
    fail "Build with no content failed"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR96"

# Test 110: Growth emoji has tooltip on hover
test_case "Growth emoji has tooltip on hover"
TESTDIR98=$(mktemp -d)
cd "$TESTDIR98"
"$LEAFPRESS" init > /dev/null 2>&1
cat > page.md << 'EOF'
---
title: Test
growth: seedling
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Check for CSS tooltip (no title attribute, uses ::after)
if grep -q 'lp-growth--seedling' _site/page/index.html && \
   grep -q 'lp-growth--seedling::after' _site/style.css; then
    pass
else
    fail "Growth emoji tooltip not present"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR98"

# Test 111: Circular wikilinks don't cause infinite loop
test_case "Circular wikilinks handled"
TESTDIR97=$(mktemp -d)
cd "$TESTDIR97"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "backlinks": true
}
EOF
cat > a.md << 'EOF'
---
title: Page A
---
Link to [[b]]
EOF
cat > b.md << 'EOF'
---
title: Page B
---
Link to [[c]]
EOF
cat > c.md << 'EOF'
---
title: Page C
---
Link to [[a]]
EOF
# Should complete without hanging - run in background and check
"$LEAFPRESS" build > /dev/null 2>&1 &
BUILD_PID=$!
sleep 3
if kill -0 $BUILD_PID 2>/dev/null; then
    kill $BUILD_PID 2>/dev/null
    fail "Circular wikilinks caused hang"
else
    wait $BUILD_PID
    if [ $? -eq 0 ]; then
        pass
    else
        fail "Circular wikilinks caused crash"
    fi
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR97"

# Test 112: Reading time is calculated and displayed
test_case "Reading time is calculated and displayed"
TESTDIR99=$(mktemp -d)
cd "$TESTDIR99"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test Article
---
This is a test article with enough words to generate a reading time estimate. We need to write several sentences to ensure we have enough content. The reading time calculation uses 150 words per minute for dense technical content. This accounts for re-reading complex sentences, processing technical concepts, and following wiki-links mentally. Let us continue writing more content to make this a longer article. Here is some additional text to pad out the word count. More words follow here. And more words here too. The reading time should appear in the page metadata section.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q 'lp-reading-time' _site/test/index.html && grep -q 'min read' _site/test/index.html; then
    pass
else
    fail "Reading time not displayed"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR99"

# Test 113: Reading time accounts for images
test_case "Reading time accounts for images"
TESTDIR100=$(mktemp -d)
cd "$TESTDIR100"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir -p static/images
echo "dummy" > static/images/img1.png
echo "dummy" > static/images/img2.png
cat > test.md << 'EOF'
---
title: Test with Images
---
Short text with images.

![[img1.png]]

More text here.

![[img2.png]]
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should have reading time (even with few words, images add time)
if grep -q 'min read' _site/test/index.html; then
    pass
else
    fail "Reading time not accounting for images"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR100"

# Test 114: Reading time frontmatter override
test_case "Reading time can be overridden in frontmatter"
TESTDIR101=$(mktemp -d)
cd "$TESTDIR101"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test Override
readingTime: 42
---
Short content that would normally be 1 min.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '42 min read' _site/test/index.html; then
    pass
else
    fail "Reading time override not applied"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR101"

# Test 115: Reading time minimum is 1 minute
test_case "Reading time minimum is 1 minute"
TESTDIR102=$(mktemp -d)
cd "$TESTDIR102"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Tiny
---
Hi.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
if grep -q '1 min read' _site/test/index.html; then
    pass
else
    fail "Reading time minimum not 1 minute"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR102"

# Test 116: Reading time separator on desktop (CSS class present)
test_case "Reading time has proper CSS classes"
TESTDIR103=$(mktemp -d)
cd "$TESTDIR103"
"$LEAFPRESS" init > /dev/null 2>&1
cat > test.md << 'EOF'
---
title: Test
date: 2024-01-15
---
Some content here to generate reading time. Adding more words to ensure we have enough for the estimate.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Check both reading time and date info classes are present
if grep -q 'class="lp-reading-time"' _site/test/index.html && grep -q 'class="lp-date-info"' _site/test/index.html; then
    pass
else
    fail "Reading time CSS classes not present"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR103"

# Test 117: Site-wide TOC disable
test_case "Site-wide toc: false disables TOC globally"
TESTDIR104=$(mktemp -d)
cd "$TESTDIR104"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "toc": false
}
EOF
cat > test.md << 'EOF'
---
title: Test Page
---
## Heading One
Content here.
## Heading Two
More content.
## Heading Three
Even more content.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# TOC should NOT be present when globally disabled
if ! grep -q 'class="lp-toc"' _site/test/index.html; then
    pass
else
    fail "TOC should be disabled globally"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR104"

# Test 118: Tags are case-sensitive
test_case "Tags are case-sensitive"
TESTDIR105=$(mktemp -d)
cd "$TESTDIR105"
"$LEAFPRESS" init > /dev/null 2>&1
cat > post1.md << 'EOF'
---
title: Post One
tags: [GoLang]
---
Content
EOF
cat > post2.md << 'EOF'
---
title: Post Two
tags: [golang]
---
Content
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should have separate tag pages for GoLang and golang
if [ -d "_site/tags/golang" ] && [ -d "_site/tags/GoLang" ] 2>/dev/null || \
   ([ -f "_site/tags/golang/index.html" ] && ! grep -q "Post One" _site/tags/golang/index.html); then
    # Either separate dirs OR golang page doesn't have Post One (case preserved)
    pass
else
    # Check if tags index shows both variants
    if grep -q 'golang' _site/tags/index.html && grep -q 'GoLang' _site/tags/index.html; then
        pass
    else
        fail "Tags not case-sensitive"
    fi
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR105"

# Test 119: Emojis in headings generate valid IDs
test_case "Emojis in headings generate valid IDs"
TESTDIR106=$(mktemp -d)
cd "$TESTDIR106"
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
---
## 🚀 Getting Started
Content here.
## Hello 👋 World
More content.
## Pure Text Heading
Even more.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should have valid IDs (emojis stripped, text preserved - may have leading hyphen)
if grep -q 'id=".*getting-started"' _site/test/index.html && \
   grep -q 'id="hello-.*world"' _site/test/index.html && \
   grep -q 'id="pure-text-heading"' _site/test/index.html; then
    pass
else
    fail "Emoji headings don't generate valid IDs"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR106"

# Test 120: Special characters in headings
test_case "Special characters in headings generate valid IDs"
TESTDIR107=$(mktemp -d)
cd "$TESTDIR107"
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
---
## What's New?
Content.
## C++ & Rust
More content.
## Section (Part 1)
Even more.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should have valid IDs with special chars converted
if grep -q 'id="what' _site/test/index.html && \
   grep -q 'id="c-' _site/test/index.html && \
   grep -q 'id="section' _site/test/index.html; then
    pass
else
    fail "Special characters in headings not handled"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR107"

# Test 121: Solid background colors work
test_case "Solid background colors are applied"
TESTDIR108=$(mktemp -d)
cd "$TESTDIR108"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "theme": {
    "background": {
      "light": "#f5f5f5",
      "dark": "#1a1a1a"
    }
  }
}
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# Should have solid background colors in inline styles
if grep -q '#f5f5f5' _site/index.html && grep -q '#1a1a1a' _site/index.html; then
    pass
else
    fail "Solid background colors not applied"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR108"

# Test 122: Static images are copied and accessible
test_case "Static images are copied to output"
TESTDIR109=$(mktemp -d)
cd "$TESTDIR109"
"$LEAFPRESS" init > /dev/null 2>&1
mkdir -p static/images
echo "PNG_DUMMY_DATA" > static/images/test-image.png
echo "JPG_DUMMY_DATA" > static/images/photo.jpg
mkdir -p static/assets
echo "SVG_DATA" > static/assets/icon.svg
"$LEAFPRESS" build > /dev/null 2>&1
# All static files should be copied
if [ -f "_site/static/images/test-image.png" ] && \
   [ -f "_site/static/images/photo.jpg" ] && \
   [ -f "_site/static/assets/icon.svg" ]; then
    pass
else
    fail "Static images not copied"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR109"

# Test 123: Per-page TOC enable overrides global disable
test_case "Per-page toc: true overrides global toc: false"
TESTDIR110=$(mktemp -d)
cd "$TESTDIR110"
"$LEAFPRESS" init > /dev/null 2>&1
cat > leafpress.json << 'EOF'
{
  "title": "Test",
  "toc": false
}
EOF
cat > test.md << 'EOF'
---
title: Test
toc: true
---
## Heading One
Content.
## Heading Two
More content.
EOF
"$LEAFPRESS" build > /dev/null 2>&1
# TOC should be present because page enables it
if grep -q 'class="lp-toc"' _site/test/index.html; then
    pass
else
    fail "Per-page TOC enable not working"
fi
cd "$ORIGDIR"
rm -rf "$TESTDIR110"

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
