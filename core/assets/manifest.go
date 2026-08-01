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
	"net/url"
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
	// ContentType is the MIME type the asset must be served with. Callers
	// must pass one canonical spelling (lowercase type/subtype, single
	// space before parameters): manifests are compared and hashed as JSON,
	// so equivalent-but-differently-spelled MIME strings would break
	// equality and registry-ID determinism.
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

// IsBuiltinPath reports whether a logical path lies in (or is exactly) the
// namespace reserved for Leafpress-owned built-in assets. The bare directory
// path matches too: a user file named "static/leafpress" would shadow the
// reserved directory. It assumes a path that already passed
// ValidateLogicalPath; it is a namespace classifier, not a standalone
// security boundary (non-canonical shapes like "static/leafpress/../x" must
// be rejected by validation first).
func IsBuiltinPath(logicalPath string) bool {
	return strings.HasPrefix(logicalPath, BuiltinPrefix) ||
		logicalPath == strings.TrimSuffix(BuiltinPrefix, "/")
}

// validatePathShape enforces the single canonical path representation shared
// by logical and output paths: relative, forward-slash, clean file-path
// segments, with segment characters restricted to the URL unreserved set so
// a path spells itself identically in a filesystem, a URL, and a quoted CSS
// url() context. The whitelist implicitly rejects query/fragment syntax and
// percent escapes (URL-parse/decode ambiguity), quotes and parentheses (CSS
// delimiters), backslashes, drive-letter colons, spaces, and all control
// characters (C0, DEL, and C1 alike).
func validatePathShape(kind, p string) error {
	if p == "" {
		return fmt.Errorf("%s path is empty", kind)
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("%s path %q must be site-relative, not absolute", kind, p)
	}
	// Targeted message for the URL-shaped mistakes people actually make;
	// the whitelist below would catch these too.
	if i := strings.IndexAny(p, "?#%"); i >= 0 {
		return fmt.Errorf("%s path %q must not contain %q: paths are not URLs", kind, p, string(p[i]))
	}
	for _, segment := range strings.Split(p, "/") {
		switch segment {
		case "":
			return fmt.Errorf("%s path %q has an empty segment", kind, p)
		case ".", "..":
			return fmt.Errorf("%s path %q contains a traversal segment", kind, p)
		}
		for _, r := range segment {
			if !isUnreservedChar(r) {
				return fmt.Errorf("%s path %q contains %q: segments may only use A-Z a-z 0-9 - . _ ~", kind, p, r)
			}
		}
		if err := validateWindowsPortableSegment(segment); err != nil {
			return fmt.Errorf("%s path %q: %w", kind, p, err)
		}
	}
	return nil
}

// windowsDeviceNames are segment base names Windows reserves regardless of
// extension ("CON", "con.txt", and "COM1.woff2" all name devices). Leafpress
// publishes Windows binaries, so canonical paths must stay writable there.
var windowsDeviceNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

func validateWindowsPortableSegment(segment string) error {
	// Windows silently strips trailing dots, so "a." and "a" collide.
	if strings.HasSuffix(segment, ".") {
		return fmt.Errorf("segment %q ends with a dot", segment)
	}
	base := segment
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	if windowsDeviceNames[strings.ToUpper(base)] {
		return fmt.Errorf("segment %q is a reserved Windows device name", segment)
	}
	return nil
}

// isUnreservedChar reports whether r is in RFC 3986's unreserved set.
func isUnreservedChar(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	case r == '-' || r == '.' || r == '_' || r == '~':
		return true
	}
	return false
}

// EscapedURLPath returns the path with each segment URL-escaped, for
// interpolation into URLs or CSS url() values. Valid canonical paths are
// escape-free, so this is normally the identity — it exists as defense in
// depth for renderers, never as a substitute for validation.
func EscapedURLPath(p string) string {
	segments := strings.Split(p, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
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
// manifest has exactly one form. Duplicate detection is case-folded: paths
// are case-sensitive identifiers, but the default macOS and Windows
// filesystems are case-insensitive, so entries differing only by case would
// materialize to one file on the CLI while a case-sensitive hosted store
// keeps both — a guaranteed cross-interface divergence.
func (m Manifest) Validate() error {
	logical := make(map[string]string, len(m))
	output := make(map[string]string, len(m))
	for i, a := range m {
		if err := a.Validate(); err != nil {
			return err
		}
		if i > 0 && m[i-1].LogicalPath > a.LogicalPath {
			return fmt.Errorf("manifest is not in canonical order: %q after %q (use NewManifest)", a.LogicalPath, m[i-1].LogicalPath)
		}
		logicalFold := strings.ToLower(a.LogicalPath)
		if prev, ok := logical[logicalFold]; ok {
			return fmt.Errorf("logical paths %q and %q collide on case-insensitive filesystems", prev, a.LogicalPath)
		}
		logical[logicalFold] = a.LogicalPath
		out := a.EffectiveOutputPath()
		outFold := strings.ToLower(out)
		if prev, ok := output[outFold]; ok {
			return fmt.Errorf("output paths %q and %q collide on case-insensitive filesystems", prev, out)
		}
		output[outFold] = out
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
