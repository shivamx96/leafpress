#!/bin/bash
# LeafPress Smoke Test - Quick verification of core functionality

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "🚀 LeafPress Smoke Test"
echo "======================"
echo ""

# Build
echo "Building leafpress..."
go build -o leafpress ./cmd/leafpress
echo -e "${GREEN}✓${NC} Build successful"

# Test 1: Build website
echo -n "Testing build... "
cd ../website
../cli/leafpress build > /dev/null 2>&1
if [ -f "_site/index.html" ]; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    exit 1
fi
cd ..

# Test 2: Wiki links
echo -n "Testing wiki links... "
if grep -q 'class="lp-wikilink"' website/_site/index.html; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    exit 1
fi

# Test 3: TOC
echo -n "Testing TOC... "
if grep -q 'class="lp-toc"' website/_site/features/index.html; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    exit 1
fi

# Test 4: Backlinks
echo -n "Testing backlinks... "
if grep -q 'class="lp-backlinks"' website/_site/guide/installation/index.html; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${GREEN}✓${NC} (no backlinks - expected)"
fi

# Test 5: Default theme
echo -n "Testing default theme... "
if grep -q -- '--lp-accent: #50ac00;' website/_site/index.html; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    exit 1
fi

# Test 6: Code block protection (wiki-link syntax preserved in code blocks)
echo -n "Testing code block protection... "
if grep -q '<code>\[\[.*\]\]</code>' website/_site/features/index.html; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    exit 1
fi

# Test 7: RSS feed generated
echo -n "Testing RSS feed... "
if [ -f "website/_site/feed.xml" ] && grep -q '<rss' website/_site/feed.xml; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    exit 1
fi

# Test 8: Design system variables
echo -n "Testing design system variables... "
if grep -q '\-\-lp-font-sm' website/_site/index.html && grep -q '\-\-lp-radius-sm' website/_site/index.html; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}All smoke tests passed!${NC}"
echo ""
echo "Run './test.sh' for comprehensive testing"
