// Package assets defines the canonical asset-manifest contract shared by the
// CLI and the embedded renderer (docs/07_ASSET_ARCHITECTURE.md). An Asset is
// metadata only — logical path, content type, hash, size — never bytes. The
// CLI materializes assets from disk; hosted consumers resolve them through
// their own storage manifest. This package therefore performs no filesystem
// or network access, preserving the renderer's purity boundary.
package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime"
	"regexp"
	"sort"
	"strings"
)

// BuiltinPrefix is the logical-path namespace reserved for assets Leafpress
// itself ships (fonts, favicons). User projects must not place files there.
const BuiltinPrefix = "static/leafpress/"

// Asset describes one portable site asset.
type Asset struct {
	// LogicalPath is the canonical identity of the asset: a clean,
	// forward-slash, relative path under "static/".
	LogicalPath string `json:"logicalPath"`
	// ContentType is the MIME type the asset must be served with.
	ContentType string `json:"contentType"`
	// SHA256 is the lowercase-hex SHA-256 of the asset content. It doubles
	// as the content-addressed storage key for hosted consumers.
	SHA256 string `json:"sha256"`
	// Size is the content length in bytes.
	Size int64 `json:"size"`
	// PublicPath optionally overrides where the asset is served when that
	// differs from "/" + LogicalPath (e.g. root-level favicons).
	PublicPath string `json:"publicPath,omitempty"`
}

// Manifest is a set of assets. A valid manifest has no duplicate logical or
// public paths; Sort makes its order deterministic.
type Manifest []Asset

var sha256Regex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Sum returns the canonical lowercase-hex SHA-256 of data.
func Sum(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

// IsBuiltinPath reports whether a logical path lies in the namespace reserved
// for Leafpress-owned built-in assets.
func IsBuiltinPath(logicalPath string) bool {
	return strings.HasPrefix(logicalPath, BuiltinPrefix)
}

// ValidateLogicalPath enforces the canonical logical-path rules: relative,
// forward-slash, rooted under "static/", with clean segments. Logical paths
// become output file paths (CLI) and storage keys (hosted), so absolute and
// traversal shapes are rejected as a security boundary.
func ValidateLogicalPath(logicalPath string) error {
	if logicalPath == "" {
		return fmt.Errorf("logical path is empty")
	}
	if strings.HasPrefix(logicalPath, "/") {
		return fmt.Errorf("logical path %q must be relative, not absolute", logicalPath)
	}
	if strings.Contains(logicalPath, "\\") {
		return fmt.Errorf("logical path %q must use forward slashes", logicalPath)
	}
	for _, r := range logicalPath {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("logical path %q contains control characters", logicalPath)
		}
	}
	// Windows drive letters ("C:/...") would escape the site root when the
	// path is joined on that platform.
	if strings.Contains(logicalPath, ":") {
		return fmt.Errorf("logical path %q must not contain ':'", logicalPath)
	}
	if !strings.HasPrefix(logicalPath, "static/") {
		return fmt.Errorf("logical path %q must be under static/", logicalPath)
	}
	segments := strings.Split(logicalPath, "/")
	for _, segment := range segments {
		switch segment {
		case "":
			return fmt.Errorf("logical path %q has an empty segment", logicalPath)
		case ".", "..":
			return fmt.Errorf("logical path %q contains a traversal segment", logicalPath)
		}
	}
	if segments[len(segments)-1] == "static" || len(segments) < 2 {
		return fmt.Errorf("logical path %q must name a file under static/", logicalPath)
	}
	return nil
}

// ValidatePublicPath enforces the shape of an explicit public path: a clean
// absolute URL path.
func ValidatePublicPath(publicPath string) error {
	if !strings.HasPrefix(publicPath, "/") {
		return fmt.Errorf("public path %q must start with /", publicPath)
	}
	if strings.Contains(publicPath, "\\") {
		return fmt.Errorf("public path %q must use forward slashes", publicPath)
	}
	for _, r := range publicPath {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("public path %q contains control characters", publicPath)
		}
	}
	segments := strings.Split(publicPath[1:], "/")
	for _, segment := range segments {
		switch segment {
		case "":
			return fmt.Errorf("public path %q has an empty segment", publicPath)
		case ".", "..":
			return fmt.Errorf("public path %q contains a traversal segment", publicPath)
		}
	}
	return nil
}

// Validate checks a single asset against the manifest contract.
func (a Asset) Validate() error {
	if err := ValidateLogicalPath(a.LogicalPath); err != nil {
		return err
	}
	if a.ContentType == "" {
		return fmt.Errorf("asset %q has no content type", a.LogicalPath)
	}
	mediaType, _, err := mime.ParseMediaType(a.ContentType)
	if err != nil {
		return fmt.Errorf("asset %q has invalid content type %q: %v", a.LogicalPath, a.ContentType, err)
	}
	if !strings.Contains(mediaType, "/") {
		return fmt.Errorf("asset %q content type %q must be type/subtype", a.LogicalPath, a.ContentType)
	}
	if !sha256Regex.MatchString(a.SHA256) {
		return fmt.Errorf("asset %q sha256 must be 64 lowercase hex characters, got %q", a.LogicalPath, a.SHA256)
	}
	if a.Size < 0 {
		return fmt.Errorf("asset %q has negative size %d", a.LogicalPath, a.Size)
	}
	if a.PublicPath != "" {
		if err := ValidatePublicPath(a.PublicPath); err != nil {
			return fmt.Errorf("asset %q: %w", a.LogicalPath, err)
		}
	}
	return nil
}

// EffectivePublicPath returns where the asset is served: the explicit
// PublicPath when set, otherwise "/" + LogicalPath.
func (a Asset) EffectivePublicPath() string {
	if a.PublicPath != "" {
		return a.PublicPath
	}
	return "/" + a.LogicalPath
}

// Validate checks every asset and rejects duplicate logical or public paths.
func (m Manifest) Validate() error {
	logical := make(map[string]bool, len(m))
	public := make(map[string]bool, len(m))
	for _, a := range m {
		if err := a.Validate(); err != nil {
			return err
		}
		if logical[a.LogicalPath] {
			return fmt.Errorf("duplicate logical path %q", a.LogicalPath)
		}
		logical[a.LogicalPath] = true
		pub := a.EffectivePublicPath()
		if public[pub] {
			return fmt.Errorf("duplicate public path %q", pub)
		}
		public[pub] = true
	}
	return nil
}

// Sort orders the manifest deterministically by logical path.
func (m Manifest) Sort() {
	sort.Slice(m, func(i, j int) bool { return m[i].LogicalPath < m[j].LogicalPath })
}
