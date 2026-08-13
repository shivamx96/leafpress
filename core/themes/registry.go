// Package themes owns Leafpress's bundled visual themes and their defaults.
// It intentionally has no dependency on config or templates so both layers
// can resolve the same registry without an import cycle.
package themes

import (
	_ "embed"
	"sort"
)

const (
	// Aurora is Leafpress's expressive, layered appearance.
	Aurora = "aurora"
	// Classic is the original Leafpress appearance.
	Classic = "classic"
	// Paper is Leafpress's editorial, print-inspired appearance.
	Paper = "paper"
	// Terminal is Leafpress's compact, command-line-inspired appearance.
	Terminal = "terminal"
	// DefaultPreset is selected when theme.preset is omitted.
	DefaultPreset = Classic
)

// Defaults are the configurable values a preset supplies before explicit
// leafpress.json overrides are applied. Empty backgrounds mean the shared
// light and dark defaults rendered by the page template remain in effect.
type Defaults struct {
	FontHeading     string
	FontBody        string
	FontMono        string
	Accent          string
	BackgroundLight string
	BackgroundDark  string
	NavStyle        string
	NavActiveStyle  string
}

// Definition is one bundled theme and the defaults that belong to it.
type Definition struct {
	Name     string
	CSS      string
	Defaults Defaults
}

// BaseCSS contains the visual invariants shared by every bundled theme.
//
//go:embed styles/base.css
var BaseCSS string

//go:embed styles/classic.css
var classicCSS string

// Aurora builds on the complete component coverage of the original theme and
// then replaces its composition, surfaces, and interaction treatments. Keeping
// the baseline in the bundle means new shared components remain usable while
// Aurora can deliberately diverge wherever its visual language calls for it.
//
//go:embed styles/aurora.css
var auroraCSS string

// Paper keeps the original component coverage as a compatibility foundation,
// then replaces the visual composition with an editorial print language.
//
//go:embed styles/paper.css
var paperCSS string

// Terminal keeps the original component coverage as a compatibility foundation,
// then replaces the visual composition with a compact command-line language.
//
//go:embed styles/terminal.css
var terminalCSS string

var registry = map[string]Definition{
	Aurora: {
		Name: Aurora,
		CSS:  classicCSS + "\n" + auroraCSS,
		Defaults: Defaults{
			FontHeading:     "Space Grotesk",
			FontBody:        "Inter",
			FontMono:        "JetBrains Mono",
			Accent:          "#16813d",
			BackgroundLight: "linear-gradient(135deg, #dcfce7 0%, #f0fdf4 34%, #e0f2fe 72%, #ecfeff 100%)",
			BackgroundDark:  "linear-gradient(135deg, #10261d 0%, #101d23 52%, #10243a 100%)",
			NavStyle:        "glassy",
			NavActiveStyle:  "box",
		},
	},
	Classic: {
		Name: Classic,
		CSS:  classicCSS,
		Defaults: Defaults{
			FontHeading:    "Bricolage Grotesque",
			FontBody:       "Inter",
			FontMono:       "JetBrains Mono",
			Accent:         "#50ac00",
			NavStyle:       "base",
			NavActiveStyle: "base",
		},
	},
	Paper: {
		Name: Paper,
		CSS:  classicCSS + "\n" + paperCSS,
		Defaults: Defaults{
			FontHeading:     "Newsreader",
			FontBody:        "Source Serif 4",
			FontMono:        "IBM Plex Mono",
			Accent:          "#765432",
			BackgroundLight: "#faf8f3",
			BackgroundDark:  "#191714",
			NavStyle:        "sticky",
			NavActiveStyle:  "underlined",
		},
	},
	Terminal: {
		Name: Terminal,
		CSS:  classicCSS + "\n" + terminalCSS,
		Defaults: Defaults{
			FontHeading:     "IBM Plex Mono",
			FontBody:        "IBM Plex Mono",
			FontMono:        "IBM Plex Mono",
			Accent:          "#087f5b",
			BackgroundLight: "#f2f5ef",
			BackgroundDark:  "#0b100e",
			NavStyle:        "sticky",
			NavActiveStyle:  "base",
		},
	},
}

// Lookup returns the exact named bundled theme. Callers decide whether an
// empty or unknown name should fall back or fail validation.
func Lookup(name string) (Definition, bool) {
	definition, ok := registry[name]
	return definition, ok
}

// Names returns bundled preset names in stable order for validation errors,
// documentation surfaces, and future CLI discovery.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
