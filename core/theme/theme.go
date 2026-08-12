// Package theme defines Leafpress's built-in theme presets.
//
// A preset is a token pack: one named set of design-token values (colors per
// light/dark mode, font roles, nav defaults) layered between the embedded
// base stylesheet and the user's own overrides. Presets are data, not CSS —
// resolution happens in Go, so explicit leafpress.json theme fields and user
// style.css always win over the preset, and every preset automatically covers
// dark mode and future components. The cascade is:
//
//	layer 0: templates.DefaultCSS (structure, components)
//	layer 1: the preset selected by theme.name (this package)
//	layer 2: explicit theme.* config fields and user style.css
package theme

// Palette is one mode's (light or dark) set of design-token values. Every
// field maps 1:1 to a --lp-* CSS custom property emitted in each page head.
type Palette struct {
	Bg             string // --lp-bg: page background; a CSS color or gradient
	Text           string // --lp-text: primary text color
	TextMuted      string // --lp-text-muted: secondary text (meta, nav, TOC)
	Border         string // --lp-border: hairline borders and rules
	CodeBg         string // --lp-code-bg: code blocks, inline code, panels
	Accent         string // --lp-accent: links, active states, highlights
	AccentContrast string // --lp-accent-contrast: text on accent-filled surfaces
	GraphLink      string // --lp-graph-link: knowledge-graph edge strokes
}

// Preset is one built-in theme: font-role and nav-style defaults that fill
// unset config fields, plus the light and dark palettes emitted as tokens.
// Font families must come from the bundled self-hosted catalog so a preset
// never introduces a remote font request.
type Preset struct {
	Name           string
	Description    string
	FontHeading    string
	FontBody       string
	FontMono       string
	NavStyle       string // "base", "sticky", or "glassy"
	NavActiveStyle string // "base", "box", or "underlined"
	Light          Palette
	Dark           Palette
	// ExtraCSS is an optional structural garnish appended after the base
	// stylesheet and before user CSS. Presets should stay token-first: a
	// preset that needs substantial CSS is a sign the base needs a new token.
	ExtraCSS string
}

// DefaultName is the preset applied when theme.name is unset.
const DefaultName = "default"

// presets holds the built-in registry; the default preset is first and must
// reproduce Leafpress's historical look exactly.
var presets = []Preset{
	{
		Name:           "default",
		Description:    "The classic Leafpress look: clean neutrals with a fresh green accent.",
		FontHeading:    "Bricolage Grotesque",
		FontBody:       "Inter",
		FontMono:       "JetBrains Mono",
		NavStyle:       "base",
		NavActiveStyle: "base",
		Light: Palette{
			Bg:             "#ffffff",
			Text:           "#1a1a1a",
			TextMuted:      "#666666",
			Border:         "#e5e5e5",
			CodeBg:         "#f7f7f7",
			Accent:         "#50ac00",
			AccentContrast: "#ffffff",
			GraphLink:      "#d0d0d0",
		},
		Dark: Palette{
			Bg:             "#1a1a1a",
			Text:           "#e5e5e5",
			TextMuted:      "#a0a0a0",
			Border:         "#333333",
			CodeBg:         "#2a2a2a",
			Accent:         "#50ac00",
			AccentContrast: "#ffffff",
			GraphLink:      "#444444",
		},
	},
	{
		Name:           "paper",
		Description:    "Warm cream and terracotta with serif type, like notes on good paper.",
		FontHeading:    "Fraunces",
		FontBody:       "Source Serif 4",
		FontMono:       "IBM Plex Mono",
		NavStyle:       "base",
		NavActiveStyle: "base",
		Light: Palette{
			Bg:             "#f9f5ec",
			Text:           "#2c2418",
			TextMuted:      "#6e6252",
			Border:         "#e3d9c6",
			CodeBg:         "#f1ead9",
			Accent:         "#a8552a",
			AccentContrast: "#ffffff",
			GraphLink:      "#d6cab2",
		},
		Dark: Palette{
			Bg:             "#211c14",
			Text:           "#ece4d3",
			TextMuted:      "#a89a82",
			Border:         "#3b3324",
			CodeBg:         "#2b2418",
			Accent:         "#d69a66",
			AccentContrast: "#211c14",
			GraphLink:      "#4a4130",
		},
		ExtraCSS: `.lp-title,
.lp-section-title {
  letter-spacing: -0.02em;
}`,
	},
	{
		Name:           "dusk",
		Description:    "Indigo and violet with modern grotesque type and a glassy nav.",
		FontHeading:    "Space Grotesk",
		FontBody:       "Geist",
		FontMono:       "Geist Mono",
		NavStyle:       "glassy",
		NavActiveStyle: "underlined",
		Light: Palette{
			Bg:             "#f7f6fb",
			Text:           "#201d33",
			TextMuted:      "#67638a",
			Border:         "#e1dff0",
			CodeBg:         "#eeecf7",
			Accent:         "#6247d4",
			AccentContrast: "#ffffff",
			GraphLink:      "#d2cfe6",
		},
		Dark: Palette{
			Bg:             "linear-gradient(180deg, #14121d 0%, #1b1726 100%)",
			Text:           "#e6e2f5",
			TextMuted:      "#9c95bd",
			Border:         "#2f2a45",
			CodeBg:         "#221e33",
			Accent:         "#a493ff",
			AccentContrast: "#14121d",
			GraphLink:      "#3b3457",
		},
	},
	{
		Name:           "forest",
		Description:    "Deep pine greens with readable humanist type for a true garden feel.",
		FontHeading:    "Crimson Pro",
		FontBody:       "Atkinson Hyperlegible Next",
		FontMono:       "Source Code Pro",
		NavStyle:       "base",
		NavActiveStyle: "base",
		Light: Palette{
			Bg:             "#f5f8f2",
			Text:           "#243020",
			TextMuted:      "#5f6f58",
			Border:         "#dbe5d2",
			CodeBg:         "#ebf1e4",
			Accent:         "#2e7d43",
			AccentContrast: "#ffffff",
			GraphLink:      "#ccd8c1",
		},
		Dark: Palette{
			Bg:             "#161c12",
			Text:           "#dee7d6",
			TextMuted:      "#91a186",
			Border:         "#2c3824",
			CodeBg:         "#202919",
			Accent:         "#7cc47f",
			AccentContrast: "#161c12",
			GraphLink:      "#364330",
		},
	},
	{
		Name:           "mist",
		Description:    "Cool gray-blues with quiet, minimal type and a sticky nav.",
		FontHeading:    "Geist",
		FontBody:       "Inter",
		FontMono:       "JetBrains Mono",
		NavStyle:       "sticky",
		NavActiveStyle: "base",
		Light: Palette{
			Bg:             "#f6f8fa",
			Text:           "#1d2733",
			TextMuted:      "#5c6b7c",
			Border:         "#dde4eb",
			CodeBg:         "#ecf0f4",
			Accent:         "#2b6cb8",
			AccentContrast: "#ffffff",
			GraphLink:      "#d3dbe3",
		},
		Dark: Palette{
			Bg:             "#13171c",
			Text:           "#dbe3eb",
			TextMuted:      "#8896a6",
			Border:         "#29323d",
			CodeBg:         "#1d232b",
			Accent:         "#6aa9e8",
			AccentContrast: "#13171c",
			GraphLink:      "#333e4a",
		},
	},
}

// ByName returns the preset for name; an empty or unknown name returns the
// default preset. Config validation rejects unknown names before render, so
// the fallback only shields consumers handed an unvalidated Theme.
func ByName(name string) Preset {
	for _, p := range presets {
		if p.Name == name {
			return p
		}
	}
	return presets[0]
}

// Valid reports whether name selects a preset; empty means the default.
func Valid(name string) bool {
	if name == "" {
		return true
	}
	for _, p := range presets {
		if p.Name == name {
			return true
		}
	}
	return false
}

// Names returns the registry's preset names in declaration order.
func Names() []string {
	out := make([]string, len(presets))
	for i, p := range presets {
		out[i] = p.Name
	}
	return out
}

// Presets returns a copy of the built-in registry in declaration order.
func Presets() []Preset {
	out := make([]Preset, len(presets))
	copy(out, presets)
	return out
}
