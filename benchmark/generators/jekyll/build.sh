#!/bin/bash
# Build Jekyll site
cd "$1"
rm -rf _site
jekyll build --quiet 2>&1
# Jekyll doesn't output time by default, use time command
start=$(date +%s%3N)
jekyll build --quiet 2>/dev/null
end=$(date +%s%3N)
echo $((end - start))
