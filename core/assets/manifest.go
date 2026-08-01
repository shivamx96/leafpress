// Package assets defines the canonical asset-manifest contract shared by the
// CLI and the embedded renderer (docs/07_ASSET_ARCHITECTURE.md). An Asset is
// metadata only — logical path, content type, hash, size — never bytes. The
// CLI materializes assets from disk; hosted consumers resolve them through
// their own storage. This package therefore performs no filesystem or network
// access, preserving the renderer's purity boundary.
package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime"
	"regexp"
	"sort"
	"strings"
	"unicode"
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
	// SHA256 is the lowercase-hex SHA-256 of the asset content.
	SHA256 string `json:"sha256"`
	// Size is the content length in bytes.
	Size int64 `json:"size"`
	// OutputPath optionally overrides where the asset lands relative to the
	// site root when that differs from LogicalPath (e.g. "favicon.ico").
	// It is site-relative, never an absolute URL: the CLI writes it under
	// _site/, templates prefix it with BasePath, and hosts map it inside
	// the garden's own route.
	OutputPath string `json:"outputPath,omitempty"`
}

// Manifest is a set of assets. A valid manifest is in canonical form:
// strictly ascending by logical path with no duplicate logical or effective
// output paths. Use NewManifest to canonicalize arbitrary input.
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

// validatePathShape enforces the single canonical path representation shared
// by logical and output paths: relative, forward-slash, clean file-path
// segments, and no characters that would let one manifest value mean
// different things before and after URL parsing or percent-decoding.
func validatePathShape(kind, p string) error {
	if p == "" {
		return fmt.Errorf("%s path is empty", kind)
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("%s path %q must be site-relative, not absolute", kind, p)
	}
	if strings.Contains(p, "\\") {
		return fmt.Errorf("%s path %q must use forward slashes", kind, p)
	}
	// Paths are literal file paths, not URLs: query/fragment syntax and
	// percent escapes are rejected outright so validation and duplicate
	// detection cannot be bypassed by values that collide after URL
	// parsing or proxy-side percent-decoding (e.g. "a?x", "%2e%2e", "%2f").
	if i := strings.IndexAny(p, "?#%"); i >= 0 {
		return fmt.Errorf("%s path %q must not contain %q: paths are not URLs", kind, p, string(p[i]))
	}
	// Windows drive letters ("C:/...") would escape the site root when the
	// path is joined on that platform.
	if strings.Contains(p, ":") {
		return fmt.Errorf("%s path %q must not contain ':'", kind, p)
	}
	for _, r := range p {
		// unicode.IsControl covers C0, DEL, and the C1 range (U+0080–U+009F).
		if unicode.IsControl(r) {
			return fmt.Errorf("%s path %q contains control characters", kind, p)
		}
	}
	for _, segment := range strings.Split(p, "/") {
		switch segment {
		case "":
			return fmt.Errorf("%s path %q has an empty segment", kind, p)
		case ".", "..":
			return fmt.Errorf("%s path %q contains a traversal segment", kind, p)
		}
	}
	return nil
}

// ValidateLogicalPath enforces the canonical logical-path rules: the shared
// path shape, rooted under "static/". Logical paths become output file paths
// and storage keys, so this is a security boundary.
func ValidateLogicalPath(logicalPath string) error {
	if err := validatePathShape("logical", logicalPath); err != nil {
		return err
	}
	if !strings.HasPrefix(logicalPath, "static/") {
		return fmt.Errorf("logical path %q must be under static/", logicalPath)
	}
	return nil
}

// ValidateOutputPath enforces the shape of an explicit output path: the same
// canonical rules as logical paths, site-relative, but allowed anywhere in
// the site (root favicons live outside static/).
func ValidateOutputPath(outputPath string) error {
	return validatePathShape("output", outputPath)
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
	if a.OutputPath != "" {
		if err := ValidateOutputPath(a.OutputPath); err != nil {
			return fmt.Errorf("asset %q: %w", a.LogicalPath, err)
		}
	}
	return nil
}

// EffectiveOutputPath returns where the asset lands relative to the site
// root: the explicit OutputPath when set, otherwise the logical path.
func (a Asset) EffectiveOutputPath() string {
	if a.OutputPath != "" {
		return a.OutputPath
	}
	return a.LogicalPath
}

// NewManifest canonicalizes and validates a set of assets: the result is
// sorted ascending by logical path and free of duplicates. This is the way
// callers should build manifests; Validate deliberately rejects
// non-canonical order so "validated" always implies "deterministic".
func NewManifest(list ...Asset) (Manifest, error) {
	m := make(Manifest, len(list))
	copy(m, list)
	sort.Slice(m, func(i, j int) bool { return m[i].LogicalPath < m[j].LogicalPath })
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// Validate checks every asset, rejects duplicate logical or effective output
// paths, and requires canonical (strictly ascending) order so a valid
// manifest has exactly one form.
func (m Manifest) Validate() error {
	output := make(map[string]bool, len(m))
	for i, a := range m {
		if err := a.Validate(); err != nil {
			return err
		}
		if i > 0 {
			switch {
			case m[i-1].LogicalPath == a.LogicalPath:
				return fmt.Errorf("duplicate logical path %q", a.LogicalPath)
			case m[i-1].LogicalPath > a.LogicalPath:
				return fmt.Errorf("manifest is not in canonical order: %q after %q (use NewManifest)", a.LogicalPath, m[i-1].LogicalPath)
			}
		}
		out := a.EffectiveOutputPath()
		if output[out] {
			return fmt.Errorf("duplicate output path %q", out)
		}
		output[out] = true
	}
	return nil
}

// Merge overlays overrides onto base: an override whose effective output
// path matches a base entry replaces it (the favicon-override rule from the
// ADR); other overrides are appended. The result is canonical and validated.
func Merge(base Manifest, overrides Manifest) (Manifest, error) {
	overridden := make(map[string]bool, len(overrides))
	for _, o := range overrides {
		overridden[o.EffectiveOutputPath()] = true
	}
	merged := make([]Asset, 0, len(base)+len(overrides))
	for _, a := range base {
		if !overridden[a.EffectiveOutputPath()] {
			merged = append(merged, a)
		}
	}
	merged = append(merged, overrides...)
	return NewManifest(merged...)
}
