package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckSymlinkEscape reports an error when path is a symlink that resolves
// outside root.
//
// A garden is a publishing boundary: everything under it is intended to be
// world-readable, and everything outside it is not. A link such as
// notes/leak.md -> ~/.ssh/id_rsa is read and published like any other note,
// so the escape is refused rather than silently followed. Links that stay
// inside the garden keep working — they resolve to content that was already
// going to be published.
//
// path must be the leaf being considered. Both tree walkers use Lstat
// semantics and therefore never descend through a directory symlink, so the
// leaf is the only place a link can be introduced.
func CheckSymlinkEscape(root, path string, mode os.FileMode) error {
	if mode&os.ModeSymlink == 0 {
		return nil
	}

	// Resolve the root too: on macOS a temp dir under /var is itself reached
	// through the /var -> /private/var link, and an un-resolved comparison
	// would reject every file in it.
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("resolve project directory %s: %w", root, err)
	}
	target, err := filepath.EvalSymlinks(path)
	if err != nil {
		// A dangling link points at nothing publishable either way.
		return fmt.Errorf("%s is a symlink that cannot be resolved: %w", displayPath(root, path), err)
	}

	rel, err := filepath.Rel(resolvedRoot, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf(
			"%s is a symlink pointing outside the project (%s); refusing to publish it",
			displayPath(root, path), target,
		)
	}
	return nil
}

// displayPath prefers the project-relative spelling so errors name the file
// the author recognises rather than an absolute temp path.
func displayPath(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return rel
	}
	return path
}
