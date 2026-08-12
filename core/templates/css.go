package templates

import "github.com/shivamx96/leafpress/core/themes"

var (
	// BaseCSS and ClassicCSS remain exported from templates for compatibility
	// with callers that inspected the original embedded stylesheet directly.
	BaseCSS    = themes.BaseCSS
	ClassicCSS = mustThemeCSS(themes.Classic)

	// DefaultCSS is the complete classic stylesheet used by existing callers.
	DefaultCSS = BaseCSS + "\n" + ClassicCSS
)

// CSSForPreset composes the shared base with the selected bundled theme. An
// empty name preserves compatibility by selecting the default. Unknown names
// also fall back defensively; config validation rejects them before builds.
func CSSForPreset(name string) string {
	if name == "" {
		name = themes.DefaultPreset
	}
	definition, ok := themes.Lookup(name)
	if !ok {
		definition, _ = themes.Lookup(themes.DefaultPreset)
	}
	return BaseCSS + "\n" + definition.CSS
}

func mustThemeCSS(name string) string {
	definition, ok := themes.Lookup(name)
	if !ok {
		panic("leafpress: bundled theme is not registered: " + name)
	}
	return definition.CSS
}
