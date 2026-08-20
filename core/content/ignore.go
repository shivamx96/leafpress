package content

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// IgnoreMatcher applies the glob patterns from build.ignore.
//
// The documented shapes are gitignore-flavoured:
//
//	drafts          a name with no slash matches at any depth
//	drafts/**       everything under drafts/, and drafts/ itself
//	*.draft.md      matches the file name at any depth
//	private/**      as above
//	notes/*.wip.md  a pattern with a slash is anchored at the garden root
//
// Within a path segment the usual path.Match metacharacters apply (* ? [x-z]);
// ** additionally spans separators.
type IgnoreMatcher struct {
	patterns [][]string // each pattern pre-split into segments
	anchored []bool     // pattern contained a slash, so it matches from the root
}

// NewIgnoreMatcher compiles patterns, rejecting malformed ones so a typo
// surfaces as a config error instead of silently ignoring nothing.
func NewIgnoreMatcher(patterns []string) (*IgnoreMatcher, error) {
	m := &IgnoreMatcher{}
	for _, pattern := range patterns {
		cleaned := strings.Trim(strings.TrimPrefix(strings.TrimSpace(pattern), "./"), "/")
		if cleaned == "" {
			continue
		}
		segments := strings.Split(cleaned, "/")
		for _, segment := range segments {
			if segment == "**" {
				continue
			}
			if strings.Contains(segment, "**") {
				return nil, fmt.Errorf("ignore pattern %q: ** must be a whole path segment", pattern)
			}
			// path.Match only reports a bad pattern when it reaches the
			// offending syntax, so probe it against a non-empty name.
			if _, err := path.Match(segment, "x"); err != nil {
				return nil, fmt.Errorf("ignore pattern %q: %w", pattern, err)
			}
		}
		m.patterns = append(m.patterns, segments)
		m.anchored = append(m.anchored, strings.Contains(cleaned, "/"))
	}
	return m, nil
}

// Match reports whether relPath is ignored. relPath is relative to the garden
// root, in OS form. Callers walking a tree should prune on a directory match:
// because ** also matches zero segments, "drafts/**" matches the drafts
// directory itself, not only its contents.
func (m *IgnoreMatcher) Match(relPath string) bool {
	if m == nil || len(m.patterns) == 0 {
		return false
	}
	segments := strings.Split(filepath.ToSlash(relPath), "/")

	for i, pattern := range m.patterns {
		// Test the path and every ancestor of it, so a pattern naming a
		// directory also hides what is inside it. Callers that walk a tree
		// prune on the directory; callers holding a single path (the
		// incremental rebuild) still need "drafts" to hide drafts/a.md.
		for end := 1; end <= len(segments); end++ {
			prefix := segments[:end]
			// An unanchored pattern is about the name itself, wherever it
			// sits: "drafts" hides notes/drafts/ as well as drafts/, and
			// "*.draft.md" hides that suffix at any depth.
			if !m.anchored[i] {
				if matchSegment(pattern[0], prefix[end-1]) {
					return true
				}
				continue
			}
			if matchSegments(pattern, prefix) {
				return true
			}
		}
	}
	return false
}

// matchSegments matches pattern segments against path segments, with ** able
// to consume any number of segments including none.
func matchSegments(pattern, segments []string) bool {
	if len(pattern) == 0 {
		return len(segments) == 0
	}
	if pattern[0] == "**" {
		// Try every split point: ** consumes 0..len(segments) segments.
		for i := 0; i <= len(segments); i++ {
			if matchSegments(pattern[1:], segments[i:]) {
				return true
			}
		}
		return false
	}
	if len(segments) == 0 {
		return false
	}
	if !matchSegment(pattern[0], segments[0]) {
		return false
	}
	return matchSegments(pattern[1:], segments[1:])
}

func matchSegment(pattern, name string) bool {
	ok, err := path.Match(pattern, name)
	return err == nil && ok
}
