#!/usr/bin/env sh
# Print the GitHub release notes for a product version: the matching
# "## vX.Y.Z" section of the website changelog with site-relative links made
# absolute, followed by a compare link against the previous product release.
#
# Usage: .github/scripts/release-notes.sh v1.0.0-rc.1
set -eu

version="$1"
changelog="website/changelog.md"
site="https://leafpress.in"
repo="https://github.com/shivamx96/leafpress"

section=$(awk -v heading="## ${version}" '
  $0 == heading { found = 1 }
  found && /^## / && $0 != heading { exit }
  found { print }
' "$changelog")

if [ -z "$section" ]; then
  echo "release-notes: no '## ${version}' section in ${changelog}" >&2
  exit 1
fi

# The previous product tag is the nearest v* tag reachable from the parent of
# the release commit. Scoped core/* and cli/* tags do not match the glob.
previous=$(git describe --tags --abbrev=0 --match 'v*' "${version}^")

printf '%s\n' "$section" | sed -e "s#](/#](${site}/#g"
printf '\n**Full changelog:** [%s...%s](%s/compare/%s...%s)\n' \
  "$previous" "$version" "$repo" "$previous" "$version"
