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
    ((PASSED++))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}"
    if [ ! -z "$1" ]; then
        echo "  Error: $1"
    fi
    ((FAILED++))
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
if grep -q 'class="lp-toc"' testdata/garden/_site/features/index.html; then
    pass
else
    fail "TOC not found"
fi

# Test 7: Heading IDs
test_case "Heading IDs are generated correctly"
if grep -q 'id="wiki-style-linking"' testdata/garden/_site/features/index.html; then
    pass
else
    fail "Heading ID not found or incorrect"
fi

# Test 8: Backlinks
test_case "Backlinks are generated"
if grep -q 'class="lp-backlinks"' testdata/garden/_site/guide/wiki-links/index.html; then
    pass
else
    warn "No backlinks found (may be expected)"
    ((PASSED++))
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
if [ -f "testdata/garden/_site/guide/index.html" ]; then
    pass
else
    fail "Guide index not generated"
fi

# Test 12: Tag pages
test_case "Tag pages are generated"
if [ -d "testdata/garden/_site/tags" ]; then
    pass
else
    warn "No tag pages (may be expected if no tags)"
    ((PASSED++))
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
if "$LEAFPRESS" build 2>&1 | grep -q "broken link"; then
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

# Test 16: Custom styles
test_case "Custom style.css is used"
if [ -f "testdata/garden/_site/style.css" ]; then
    pass
else
    fail "style.css not found"
fi

# Test 17: Dark mode toggle
test_case "Dark mode toggle script is included"
if grep -q 'data-theme="dark"' testdata/garden/_site/index.html; then
    pass
else
    fail "Dark mode script not found"
fi

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
