package content

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

// SlugSegment converts one file or folder name into its published URL form.
// Letters (with combining marks), digits, hyphens, underscores, and dots keep
// their case. Each run of other characters, such as whitespace, URL
// delimiters, quotes, and punctuation, becomes one hyphen, merging with any
// hyphens beside it; runs at either end are dropped, as are trailing dots.
// Names that are already URL-safe are returned unchanged, so existing URLs
// stay stable. An empty result means the name has no usable characters.
func SlugSegment(name string) string {
	out := make([]rune, 0, len(name))
	runStart := -1 // start of the current hyphen/separator run in out
	replaced := false
	flush := func(final bool) {
		if runStart < 0 {
			return
		}
		if replaced {
			// Collapse the run to one hyphen, or drop it at either end.
			out = out[:runStart]
			if runStart > 0 && !final {
				out = append(out, '-')
			}
		}
		runStart, replaced = -1, false
	}
	for _, r := range name {
		if r == '-' || !slugRune(r) {
			if runStart < 0 {
				runStart = len(out)
			}
			if r == '-' {
				out = append(out, r)
			} else {
				replaced = true
			}
			continue
		}
		flush(false)
		out = append(out, r)
	}
	flush(true)

	segment := string(out)
	if trimmed := strings.TrimRight(segment, "."); trimmed != segment {
		// A dropped trailing dot may expose a separator hyphen ("a? ." → "a-").
		segment = strings.TrimRight(trimmed, "-")
	}
	return segment
}

func slugRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsDigit(r) || r == '_' || r == '.'
}

// sourceRoute is the slash-separated route a Markdown path names before URL
// cleanup: "notes/My Note.md" → "notes/My Note". _index.md names its folder
// and the root index.md names the site root.
func sourceRoute(relPath string) string {
	route := filepath.ToSlash(strings.TrimSuffix(relPath, ".md"))

	// Handle the reserved _index.md basename only. A normal filename such as
	// migration_index.md must retain its own route.
	if path.Base(route) == "_index" {
		route = path.Dir(route)
		if route == "." {
			route = ""
		}
	}
	if route == "index" {
		route = ""
	}
	return route
}

// pageSlug resolves a page's published slug. A frontmatter slug replaces the
// page's own name and must already be URL-safe; folders still decide the
// section. Without one, every segment of the source route is cleaned with
// SlugSegment.
func pageSlug(source, frontmatterSlug string, isIndex bool) (string, error) {
	segments := strings.Split(source, "/")
	if source == "" {
		segments = nil
	}
	for i, segment := range segments {
		cleaned := SlugSegment(segment)
		if cleaned == "" {
			if i == len(segments)-1 && !isIndex {
				return "", fmt.Errorf("cannot derive a URL from the filename %q; rename the file or set slug in its frontmatter", segment)
			}
			return "", fmt.Errorf("cannot derive a URL from the folder name %q; rename the folder", segment)
		}
		segments[i] = cleaned
	}

	if frontmatterSlug != "" {
		if isIndex || source == "" {
			return "", fmt.Errorf("frontmatter slug %q is not supported on index pages; rename the folder to change its URL", frontmatterSlug)
		}
		if strings.Contains(frontmatterSlug, "/") {
			return "", fmt.Errorf("frontmatter slug %q must be a single name without slashes; move the file to change its section", frontmatterSlug)
		}
		if cleaned := SlugSegment(frontmatterSlug); cleaned != frontmatterSlug {
			if cleaned == "" {
				return "", fmt.Errorf("frontmatter slug %q has no URL-safe characters; use letters, numbers, hyphens, underscores, or dots", frontmatterSlug)
			}
			return "", fmt.Errorf("frontmatter slug %q is not URL-safe; use %q", frontmatterSlug, cleaned)
		}
		segments[len(segments)-1] = frontmatterSlug
	}
	return strings.Join(segments, "/"), nil
}
