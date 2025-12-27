#!/bin/bash
# Build Hugo site
cd "$1"
rm -rf public
# Hugo outputs "Total in X ms" - capture that
output=$(hugo 2>&1)
echo "$output" | grep -oE 'in [0-9]+ ms' | grep -oE '[0-9]+' | head -1
