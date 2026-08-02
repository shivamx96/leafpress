package assets

import (
	"fmt"
	"sort"
	"strings"
)

// BuiltinFontFace is one @font-face of the curated self-hosted font set:
// a variable-weight woff2 covering one style and unicode subset of a family.
type BuiltinFontFace struct {
	Family       string // CSS font-family name, e.g. "Inter"
	Style        string // "normal" or "italic"
	WeightRange  string // CSS font-weight range, e.g. "400 700"
	Subset       string // "latin" or "latin-ext"
	LogicalPath  string // registry logical path of the woff2
	UnicodeRange string // CSS unicode-range of the subset
}

// builtinFontFaces is the curated set used by the default theme. It is
// deliberately Latin-focused — other scripts fall back to the system stack or
// to declared custom fonts. Faces are sourced from Google Fonts under SIL OFL
// 1.1; full license texts ship as registry assets.
var builtinFontFaces = []BuiltinFontFace{
	{Family: "Bricolage Grotesque", Style: "normal", WeightRange: "200 800", Subset: "latin-ext",
		LogicalPath:  BuiltinPrefix + "fonts/bricolage-grotesque-normal-latin-ext.woff2",
		UnicodeRange: "U+0100-02BA, U+02BD-02C5, U+02C7-02CC, U+02CE-02D7, U+02DD-02FF, U+0304, U+0308, U+0329, U+1D00-1DBF, U+1E00-1E9F, U+1EF2-1EFF, U+2020, U+20A0-20AB, U+20AD-20C0, U+2113, U+2C60-2C7F, U+A720-A7FF"},
	{Family: "Bricolage Grotesque", Style: "normal", WeightRange: "200 800", Subset: "latin",
		LogicalPath:  BuiltinPrefix + "fonts/bricolage-grotesque-normal-latin.woff2",
		UnicodeRange: "U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+0304, U+0308, U+0329, U+2000-206F, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD"},
	{Family: "Inter", Style: "italic", WeightRange: "400 700", Subset: "latin-ext",
		LogicalPath:  BuiltinPrefix + "fonts/inter-italic-latin-ext.woff2",
		UnicodeRange: "U+0100-02BA, U+02BD-02C5, U+02C7-02CC, U+02CE-02D7, U+02DD-02FF, U+0304, U+0308, U+0329, U+1D00-1DBF, U+1E00-1E9F, U+1EF2-1EFF, U+2020, U+20A0-20AB, U+20AD-20C0, U+2113, U+2C60-2C7F, U+A720-A7FF"},
	{Family: "Inter", Style: "italic", WeightRange: "400 700", Subset: "latin",
		LogicalPath:  BuiltinPrefix + "fonts/inter-italic-latin.woff2",
		UnicodeRange: "U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+0304, U+0308, U+0329, U+2000-206F, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD"},
	{Family: "Inter", Style: "normal", WeightRange: "400 700", Subset: "latin-ext",
		LogicalPath:  BuiltinPrefix + "fonts/inter-normal-latin-ext.woff2",
		UnicodeRange: "U+0100-02BA, U+02BD-02C5, U+02C7-02CC, U+02CE-02D7, U+02DD-02FF, U+0304, U+0308, U+0329, U+1D00-1DBF, U+1E00-1E9F, U+1EF2-1EFF, U+2020, U+20A0-20AB, U+20AD-20C0, U+2113, U+2C60-2C7F, U+A720-A7FF"},
	{Family: "Inter", Style: "normal", WeightRange: "400 700", Subset: "latin",
		LogicalPath:  BuiltinPrefix + "fonts/inter-normal-latin.woff2",
		UnicodeRange: "U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+0304, U+0308, U+0329, U+2000-206F, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD"},
	{Family: "JetBrains Mono", Style: "italic", WeightRange: "400 700", Subset: "latin-ext",
		LogicalPath:  BuiltinPrefix + "fonts/jetbrains-mono-italic-latin-ext.woff2",
		UnicodeRange: "U+0100-02BA, U+02BD-02C5, U+02C7-02CC, U+02CE-02D7, U+02DD-02FF, U+0304, U+0308, U+0329, U+1D00-1DBF, U+1E00-1E9F, U+1EF2-1EFF, U+2020, U+20A0-20AB, U+20AD-20C0, U+2113, U+2C60-2C7F, U+A720-A7FF"},
	{Family: "JetBrains Mono", Style: "italic", WeightRange: "400 700", Subset: "latin",
		LogicalPath:  BuiltinPrefix + "fonts/jetbrains-mono-italic-latin.woff2",
		UnicodeRange: "U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+0304, U+0308, U+0329, U+2000-206F, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD"},
	{Family: "JetBrains Mono", Style: "normal", WeightRange: "400 700", Subset: "latin-ext",
		LogicalPath:  BuiltinPrefix + "fonts/jetbrains-mono-normal-latin-ext.woff2",
		UnicodeRange: "U+0100-02BA, U+02BD-02C5, U+02C7-02CC, U+02CE-02D7, U+02DD-02FF, U+0304, U+0308, U+0329, U+1D00-1DBF, U+1E00-1E9F, U+1EF2-1EFF, U+2020, U+20A0-20AB, U+20AD-20C0, U+2113, U+2C60-2C7F, U+A720-A7FF"},
	{Family: "JetBrains Mono", Style: "normal", WeightRange: "400 700", Subset: "latin",
		LogicalPath:  BuiltinPrefix + "fonts/jetbrains-mono-normal-latin.woff2",
		UnicodeRange: "U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+0304, U+0308, U+0329, U+2000-206F, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD"},
}

// BuiltinFontFaces returns the curated face set in deterministic order.
func BuiltinFontFaces() []BuiltinFontFace {
	out := make([]BuiltinFontFace, len(builtinFontFaces))
	copy(out, builtinFontFaces)
	return out
}

// BuiltinFontFamilies returns the family names of the curated set, sorted.
func BuiltinFontFamilies() []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range builtinFontFaces {
		if !seen[f.Family] {
			seen[f.Family] = true
			out = append(out, f.Family)
		}
	}
	sort.Strings(out)
	return out
}

// IsBuiltinFontFamily reports whether a family is covered by the curated
// self-hosted set (exact, case-sensitive CSS family name).
func IsBuiltinFontFamily(family string) bool {
	for _, f := range builtinFontFaces {
		if f.Family == family {
			return true
		}
	}
	return false
}

// builtinFontLicenses maps each curated family to the registry asset holding
// its full SIL OFL 1.1 license text. The OFL requires redistributed copies to
// carry the copyright notice and license, so these materialize into exported
// sites alongside the woff2 files.
var builtinFontLicenses = map[string]string{
	"Bricolage Grotesque": BuiltinPrefix + "fonts/OFL-bricolage-grotesque.txt",
	"Inter":               BuiltinPrefix + "fonts/OFL-inter.txt",
	"JetBrains Mono":      BuiltinPrefix + "fonts/OFL-jetbrains-mono.txt",
}

// RequiredBuiltins returns every built-in a rendered site references for the
// given theme families: all root built-ins (favicons, linked from every page
// head) plus the faces and OFL license texts of families covered by the
// bundled set. Content-optional built-ins such as Mermaid are omitted — use
// RequiredBuiltinsFor when diagram presence is known. This is the single
// selection list — CLI materialization and the renderer's asset manifest must
// both use it (via RequiredBuiltinsFor), never private copies.
func RequiredBuiltins(families ...string) []Builtin {
	return RequiredBuiltinsFor(false, families...)
}

// RequiredBuiltinsFor is RequiredBuiltins plus content-optional built-ins.
// Pass includeMermaid when any page's rendered HTML contains a Mermaid diagram
// so the script (and its MIT license text) land in the site manifest.
func RequiredBuiltinsFor(includeMermaid bool, families ...string) []Builtin {
	want := map[string]bool{}
	for _, family := range families {
		want[family] = true
	}

	fontOwned := map[string]bool{}
	include := map[string]bool{}
	for _, face := range builtinFontFaces {
		fontOwned[face.LogicalPath] = true
		if want[face.Family] {
			include[face.LogicalPath] = true
		}
	}
	for family, licensePath := range builtinFontLicenses {
		fontOwned[licensePath] = true
		if want[family] {
			include[licensePath] = true
		}
	}

	var out []Builtin
	for _, b := range builtins {
		path := b.Asset.LogicalPath
		if isContentOptional(path) && !includeMermaid {
			continue
		}
		if fontOwned[path] && !include[path] {
			continue
		}
		out = append(out, b)
	}
	return out
}

// BuiltinFontLicense returns the logical path of the OFL license asset for a
// curated family.
func BuiltinFontLicense(family string) (string, bool) {
	p, ok := builtinFontLicenses[family]
	return p, ok
}

// FontFaceCSS returns @font-face rules for every requested family that is in
// the curated set. Font URLs are site-relative ("static/leafpress/..."): the
// rules live in the generated stylesheet, which is itself served from the
// site root, so relative URLs resolve correctly under any base path or
// hosted garden route. Families outside the set are skipped; duplicates are
// emitted once.
func FontFaceCSS(families ...string) string {
	want := map[string]bool{}
	for _, family := range families {
		if IsBuiltinFontFamily(family) {
			want[family] = true
		}
	}
	if len(want) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, face := range builtinFontFaces {
		if !want[face.Family] {
			continue
		}
		fmt.Fprintf(&sb, `@font-face {
  font-family: "%s";
  font-style: %s;
  font-weight: %s;
  font-display: swap;
  src: url("%s") format("woff2");
  unicode-range: %s;
}
`, face.Family, face.Style, face.WeightRange, EscapedURLPath(face.LogicalPath), face.UnicodeRange)
	}
	return sb.String()
}
