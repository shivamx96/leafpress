package assets

import (
	"fmt"
	"sort"
	"strings"
)

// BuiltinFontFace is one @font-face of the curated self-hosted font set:
// a woff2 covering one weight/style and unicode subset of a family.
type BuiltinFontFace struct {
	Family       string // CSS font-family name, e.g. "Inter"
	Style        string // "normal" or "italic"
	WeightRange  string // CSS font-weight or range, e.g. "400" or "400 700"
	Subset       string // "latin" or "latin-ext"
	LogicalPath  string // registry logical path of the woff2
	UnicodeRange string // CSS unicode-range of the subset
}

// builtinFontFamily describes one curated family. Variable families emit one
// face per style/subset; static families emit one face for each listed weight.
// Files follow a deterministic naming convention so the face, license, and
// registry tables cannot silently drift apart.
type builtinFontFamily struct {
	Family        string
	Slug          string
	WeightRange   string
	Italic        bool
	StaticWeights []string
}

// builtinFontCatalog is the self-hosted catalog available to themes. It is
// deliberately Latin-focused — other scripts fall back to the system stack or
// to declared custom fonts. Faces are sourced from Google Fonts under SIL OFL
// 1.1; full license texts ship as registry assets.
var builtinFontCatalog = []builtinFontFamily{
	{Family: "Atkinson Hyperlegible Mono", Slug: "atkinson-hyperlegible-mono", WeightRange: "200 800", Italic: true},
	{Family: "Atkinson Hyperlegible Next", Slug: "atkinson-hyperlegible-next", WeightRange: "200 800", Italic: true},
	{Family: "Bricolage Grotesque", Slug: "bricolage-grotesque", WeightRange: "200 800"},
	{Family: "Crimson Pro", Slug: "crimson-pro", WeightRange: "200 900", Italic: true},
	{Family: "Fira Code", Slug: "fira-code", WeightRange: "300 700"},
	{Family: "Fraunces", Slug: "fraunces", WeightRange: "100 900", Italic: true},
	{Family: "Geist", Slug: "geist", WeightRange: "100 900", Italic: true},
	{Family: "Geist Mono", Slug: "geist-mono", WeightRange: "100 900", Italic: true},
	{Family: "IBM Plex Mono", Slug: "ibm-plex-mono", StaticWeights: []string{"400", "700"}, Italic: true},
	{Family: "IBM Plex Sans", Slug: "ibm-plex-sans", WeightRange: "100 700", Italic: true},
	{Family: "Inter", Slug: "inter", WeightRange: "400 700", Italic: true},
	{Family: "JetBrains Mono", Slug: "jetbrains-mono", WeightRange: "400 700", Italic: true},
	{Family: "Lora", Slug: "lora", WeightRange: "400 700", Italic: true},
	{Family: "Source Code Pro", Slug: "source-code-pro", WeightRange: "200 900", Italic: true},
	{Family: "Source Serif 4", Slug: "source-serif-4", WeightRange: "200 900", Italic: true},
	{Family: "Space Grotesk", Slug: "space-grotesk", WeightRange: "300 700"},
}

var builtinFontSubsets = []struct {
	Name         string
	UnicodeRange string
}{
	{Name: "latin-ext", UnicodeRange: "U+0100-02BA, U+02BD-02C5, U+02C7-02CC, U+02CE-02D7, U+02DD-02FF, U+0304, U+0308, U+0329, U+1D00-1DBF, U+1E00-1E9F, U+1EF2-1EFF, U+2020, U+20A0-20AB, U+20AD-20C0, U+2113, U+2C60-2C7F, U+A720-A7FF"},
	{Name: "latin", UnicodeRange: "U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+0304, U+0308, U+0329, U+2000-206F, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD"},
}

func buildBuiltinFontFaces() []BuiltinFontFace {
	var faces []BuiltinFontFace
	for _, family := range builtinFontCatalog {
		styles := []string{"normal"}
		if family.Italic {
			styles = append(styles, "italic")
		}
		weights := family.StaticWeights
		if len(weights) == 0 {
			weights = []string{family.WeightRange}
		}
		for _, style := range styles {
			for _, subset := range builtinFontSubsets {
				for _, weight := range weights {
					filename := fmt.Sprintf("%s-%s-%s.woff2", family.Slug, style, subset.Name)
					if len(family.StaticWeights) > 0 {
						filename = fmt.Sprintf("%s-%s-%s-%s.woff2", family.Slug, style, subset.Name, weight)
					}
					faces = append(faces, BuiltinFontFace{
						Family:       family.Family,
						Style:        style,
						WeightRange:  weight,
						Subset:       subset.Name,
						LogicalPath:  BuiltinPrefix + "fonts/" + filename,
						UnicodeRange: subset.UnicodeRange,
					})
				}
			}
		}
	}
	return faces
}

var builtinFontFaces = buildBuiltinFontFaces()

// BuiltinFontFaces returns the curated face set in deterministic order.
func BuiltinFontFaces() []BuiltinFontFace {
	out := make([]BuiltinFontFace, len(builtinFontFaces))
	copy(out, builtinFontFaces)
	return out
}

// BuiltinFontFamilies returns the family names of the curated set, sorted.
func BuiltinFontFamilies() []string {
	out := make([]string, 0, len(builtinFontCatalog))
	for _, family := range builtinFontCatalog {
		out = append(out, family.Family)
	}
	sort.Strings(out)
	return out
}

// IsBuiltinFontFamily reports whether a family is covered by the curated
// self-hosted set (exact, case-sensitive CSS family name).
func IsBuiltinFontFamily(family string) bool {
	for _, candidate := range builtinFontCatalog {
		if candidate.Family == family {
			return true
		}
	}
	return false
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
	for _, family := range builtinFontCatalog {
		licensePath := BuiltinPrefix + "fonts/OFL-" + family.Slug + ".txt"
		fontOwned[licensePath] = true
		if want[family.Family] {
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
	for _, candidate := range builtinFontCatalog {
		if candidate.Family == family {
			return BuiltinPrefix + "fonts/OFL-" + candidate.Slug + ".txt", true
		}
	}
	return "", false
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
