package templates

import "github.com/shivamx96/leafpress/core/themes"

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
	return themes.BaseCSS + "\n" + definition.CSS
}
