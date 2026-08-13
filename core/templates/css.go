package templates

import _ "embed"

// BaseCSS contains the visual invariants shared by every Leafpress theme.
//
//go:embed styles/base.css
var BaseCSS string

// ClassicCSS preserves the original Leafpress appearance as the first bundled
// theme. Theme selection is introduced separately; until then, every build
// composes this stylesheet after BaseCSS.
//
//go:embed styles/themes/classic.css
var ClassicCSS string

// DefaultCSS is the complete built-in stylesheet used by existing callers.
// Keep this compatibility surface until theme selection moves composition to
// the selected registry entry.
var DefaultCSS = BaseCSS + "\n" + ClassicCSS
