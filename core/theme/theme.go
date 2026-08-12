// Package theme defines Leafpress's built-in theme presets.
//
// A preset is a complete look layered between the embedded base stylesheet
// and the user's own overrides: a token pack (colors per light/dark mode,
// font roles, nav defaults) plus a structural CSS layer that restyles
// components. Token resolution happens in Go, so explicit leafpress.json
// theme fields and user style.css always win over the preset, and every
// preset automatically covers dark mode. The cascade is:
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
	// Layout variants, the preset's structural arrangement. Each maps to a
	// body class implemented by the base stylesheet and can be overridden
	// individually via theme.layout in leafpress.json.
	NavLayout    string // "top", "sidebar", or "minimal"
	IndexLayout  string // "list" or "cards"
	ContentWidth string // "narrow", "normal", or "wide"
	Light        Palette
	Dark         Palette
	// ExtraCSS is the preset's structural identity: rules appended after
	// the base stylesheet (and before user CSS) that restyle components —
	// link treatments, cards, header layout, radii and type-scale token
	// overrides. This is what makes presets distinct looks rather than
	// palette swaps. Values a preset varies per mode still belong in the
	// palettes; ExtraCSS should express colors through the --lp-* tokens so
	// it works in both modes and under user overrides.
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
		NavLayout:      "top",
		IndexLayout:    "list",
		ContentWidth:   "normal",
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
		Name:           "plain",
		Description:    "Text-first minimalism: square corners, underlined links, no decoration.",
		FontHeading:    "Atkinson Hyperlegible Next",
		FontBody:       "Atkinson Hyperlegible Next",
		FontMono:       "Atkinson Hyperlegible Mono",
		NavStyle:       "base",
		NavActiveStyle: "base",
		NavLayout:      "minimal",
		IndexLayout:    "list",
		ContentWidth:   "narrow",
		Light: Palette{
			Bg:             "#ffffff",
			Text:           "#222222",
			TextMuted:      "#6b6b6b",
			Border:         "#dddddd",
			CodeBg:         "#f4f4f4",
			Accent:         "#205ea6",
			AccentContrast: "#ffffff",
			GraphLink:      "#cccccc",
		},
		Dark: Palette{
			Bg:             "#171717",
			Text:           "#dddddd",
			TextMuted:      "#9a9a9a",
			Border:         "#3a3a3a",
			CodeBg:         "#262626",
			Accent:         "#7cb2e8",
			AccentContrast: "#171717",
			GraphLink:      "#454545",
		},
		// Strip the decoration the base stylesheet applies: no rounding
		// anywhere, a quieter type scale, underlined links instead of tinted
		// wiki-link chips, and unadorned quotes.
		ExtraCSS: `:root {
  --lp-radius-sm: 0;
  --lp-radius-md: 0;
  --lp-radius-lg: 0;
  --lp-radius-full: 0;
  --lp-font-lg: 1.05rem;
  --lp-font-xl: 1.2rem;
  --lp-font-2xl: 1.4rem;
  --lp-font-3xl: 1.6rem;
  --lp-font-display: 3rem;
}

.lp-content a,
.lp-tag,
.lp-tag-cloud-item,
.lp-backlink {
  text-decoration: underline;
  text-underline-offset: 2px;
}

.lp-wikilink,
.lp-wikilink:hover,
[data-theme="dark"] .lp-wikilink,
[data-theme="dark"] .lp-wikilink:hover {
  background: none;
  padding: 0;
}

.lp-content blockquote {
  font-family: var(--lp-font-body);
  font-size: var(--lp-font-base);
  font-style: normal;
  border-left: 3px solid var(--lp-border);
  padding: 0.25rem 1.25rem;
}

.lp-content blockquote::before {
  content: none;
}

.lp-title {
  font-weight: 600;
}`,
	},
	{
		Name:           "paper",
		Description:    "An editorial serif look on warm cream: centered headers, small caps, dotted wiki-links.",
		FontHeading:    "Fraunces",
		FontBody:       "Source Serif 4",
		FontMono:       "IBM Plex Mono",
		NavStyle:       "base",
		NavActiveStyle: "base",
		NavLayout:      "top",
		IndexLayout:    "list",
		ContentWidth:   "narrow",
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
		// A print-inspired layout: centered page headers under a double
		// rule, small-caps metadata, dotted underlines for wiki-links, and a
		// fleuron in place of bare horizontal rules.
		ExtraCSS: `.lp-title,
.lp-section-title {
  letter-spacing: -0.02em;
}

.lp-header {
  text-align: center;
  border-bottom: 4px double var(--lp-border);
  padding-bottom: 1.25rem;
}

.lp-meta,
.lp-tags {
  justify-content: center;
}

.lp-meta {
  font-variant-caps: small-caps;
  letter-spacing: 0.08em;
}

.lp-section-title,
.lp-section-count {
  text-align: center;
}

.lp-content h2,
.lp-content h3 {
  font-weight: 500;
}

.lp-wikilink,
[data-theme="dark"] .lp-wikilink {
  background: none;
  padding: 0;
  text-decoration: underline;
  text-decoration-style: dotted;
  text-underline-offset: 3px;
}

.lp-wikilink:hover,
[data-theme="dark"] .lp-wikilink:hover {
  background: none;
  text-decoration-style: solid;
}

.lp-content pre {
  border: 1px solid var(--lp-border);
}

.lp-content hr {
  border: none;
  margin: 1.5rem 0;
  text-align: center;
}

.lp-content hr::after {
  content: "\2766";
  display: block;
  color: var(--lp-text-muted);
  font-size: var(--lp-font-lg);
}`,
	},
	{
		Name:           "modern",
		Description:    "A contemporary look: glassy nav, rounded cards, pill tags, soft shadows, gradients.",
		FontHeading:    "Space Grotesk",
		FontBody:       "Geist",
		FontMono:       "Geist Mono",
		NavStyle:       "glassy",
		NavActiveStyle: "box",
		NavLayout:      "top",
		IndexLayout:    "cards",
		ContentWidth:   "normal",
		Light: Palette{
			Bg:             "linear-gradient(180deg, #f8f7fc 0%, #f1eff9 100%)",
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
		// Layered, rounded surfaces on top of the gradient background:
		// index entries become hover-lifting cards, tags become pills, the
		// page header carries an accent bar, and panels get soft shadows.
		ExtraCSS: `:root {
  --lp-radius-sm: 6px;
  --lp-radius-md: 12px;
  --lp-radius-lg: 20px;
}

.lp-title {
  letter-spacing: -0.03em;
}

.lp-header {
  border-bottom: none;
  padding-bottom: 0;
}

.lp-header::after {
  content: "";
  display: block;
  width: 3.5rem;
  height: 4px;
  margin-top: 1rem;
  border-radius: 9999px;
  background: linear-gradient(90deg, var(--lp-accent), color-mix(in srgb, var(--lp-accent) 35%, transparent));
}

.lp-index-item {
  background: color-mix(in srgb, var(--lp-code-bg) 55%, transparent);
  transition: border-color 0.2s, box-shadow 0.2s, transform 0.2s;
}

.lp-index-item:hover {
  border-color: color-mix(in srgb, var(--lp-accent) 45%, var(--lp-border));
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.06);
  transform: translateY(-2px);
}

[data-theme="dark"] .lp-index-item:hover {
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.35);
}

.lp-tag,
.lp-tag-cloud-item {
  background: color-mix(in srgb, var(--lp-accent) 12%, transparent);
  padding: 0.2rem 0.65rem;
  border-radius: 9999px;
}

.lp-tag:hover,
.lp-tag-cloud-item:hover {
  text-decoration: none;
  background: color-mix(in srgb, var(--lp-accent) 20%, transparent);
}

.lp-nav-link.lp-nav-link--active.lp-nav-active-box {
  border-radius: 9999px;
}

.lp-content pre {
  border: 1px solid var(--lp-border);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.05);
}

[data-theme="dark"] .lp-content pre {
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.3);
}

.lp-backlinks {
  border-top: none;
  border: 1px solid var(--lp-border);
  border-radius: var(--lp-radius-md);
  padding: 1.25rem;
  background: color-mix(in srgb, var(--lp-code-bg) 55%, transparent);
}`,
	},
	{
		Name:           "studio",
		Description:    "A workspace feel: fixed sidebar navigation, wide content, cool charcoal with teal.",
		FontHeading:    "Space Grotesk",
		FontBody:       "IBM Plex Sans",
		FontMono:       "IBM Plex Mono",
		NavStyle:       "base",
		NavActiveStyle: "base",
		NavLayout:      "sidebar",
		IndexLayout:    "list",
		ContentWidth:   "wide",
		Light: Palette{
			Bg:             "#fafbfb",
			Text:           "#1b2426",
			TextMuted:      "#5d6b6e",
			Border:         "#dde4e5",
			CodeBg:         "#eef2f2",
			Accent:         "#0f766e",
			AccentContrast: "#ffffff",
			GraphLink:      "#d0dadb",
		},
		Dark: Palette{
			Bg:             "#12181a",
			Text:           "#dbe4e5",
			TextMuted:      "#8a9a9d",
			Border:         "#283336",
			CodeBg:         "#1b2426",
			Accent:         "#4fd1c5",
			AccentContrast: "#12181a",
			GraphLink:      "#324043",
		},
		// The sidebar layout does the structural heavy lifting; this layer
		// gives the rail app-like affordances — accent-barred active links,
		// uppercase brand — and squares things off slightly.
		ExtraCSS: `:root {
  --lp-radius-lg: 8px;
}

.lp-nav-title {
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-size: var(--lp-font-base);
}

.lp-layout-nav--sidebar .lp-nav-link {
  display: block;
  width: 100%;
  padding: 0.3rem 0.75rem;
  margin-left: -0.75rem;
  border-left: 2px solid transparent;
}

.lp-layout-nav--sidebar .lp-nav-link:hover {
  border-left-color: var(--lp-border);
}

.lp-layout-nav--sidebar .lp-nav-link.lp-nav-link--active {
  color: var(--lp-accent);
  border-left-color: var(--lp-accent);
}

.lp-content h2 {
  padding-bottom: 0.35rem;
  border-bottom: 1px solid var(--lp-border);
}

.lp-index-date {
  font-family: var(--lp-font-mono);
  font-size: var(--lp-font-xs);
}`,
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
