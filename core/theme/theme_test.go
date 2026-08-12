package theme

import (
	"regexp"
	"strings"
	"testing"

	"github.com/shivamx96/leafpress/core/assets"
)

var (
	hexColor = regexp.MustCompile(`^#[0-9a-f]{6}$`)
	gradient = regexp.MustCompile(`^(?:repeating-)?(?:linear|radial|conic)-gradient\([A-Za-z0-9#()+.,%/ \t-]+\)$`)
)

func checkPalette(t *testing.T, preset, mode string, p Palette) {
	t.Helper()
	solid := map[string]string{
		"Text":           p.Text,
		"TextMuted":      p.TextMuted,
		"Border":         p.Border,
		"CodeBg":         p.CodeBg,
		"Accent":         p.Accent,
		"AccentContrast": p.AccentContrast,
		"GraphLink":      p.GraphLink,
	}
	for field, value := range solid {
		if !hexColor.MatchString(value) {
			t.Errorf("preset %s %s palette: %s = %q is not a lowercase 6-digit hex color", preset, mode, field, value)
		}
	}
	if !hexColor.MatchString(p.Bg) && !gradient.MatchString(p.Bg) {
		t.Errorf("preset %s %s palette: Bg = %q is neither a hex color nor a gradient", preset, mode, p.Bg)
	}
}

// TestPresetsAreWellFormed pins the registry invariants the config layer and
// templates rely on: resolvable names, self-hosted fonts, valid nav styles,
// and token values that are safe to interpolate into the inline style block.
func TestPresetsAreWellFormed(t *testing.T) {
	all := Presets()
	if len(all) == 0 || all[0].Name != DefaultName {
		t.Fatalf("presets must start with the %q preset", DefaultName)
	}

	seen := map[string]bool{}
	validNav := map[string]bool{"base": true, "sticky": true, "glassy": true}
	validNavActive := map[string]bool{"base": true, "box": true, "underlined": true}

	for _, p := range all {
		if p.Name == "" || p.Description == "" {
			t.Errorf("preset %+v must have a name and description", p)
		}
		if seen[p.Name] {
			t.Errorf("duplicate preset name %q", p.Name)
		}
		seen[p.Name] = true

		for role, family := range map[string]string{
			"fontHeading": p.FontHeading, "fontBody": p.FontBody, "fontMono": p.FontMono,
		} {
			if !assets.IsBuiltinFontFamily(family) {
				t.Errorf("preset %s: %s %q is not in the bundled font catalog", p.Name, role, family)
			}
		}
		if !validNav[p.NavStyle] {
			t.Errorf("preset %s: invalid navStyle %q", p.Name, p.NavStyle)
		}
		if !validNavActive[p.NavActiveStyle] {
			t.Errorf("preset %s: invalid navActiveStyle %q", p.Name, p.NavActiveStyle)
		}
		checkPalette(t, p.Name, "light", p.Light)
		checkPalette(t, p.Name, "dark", p.Dark)
		if strings.Contains(strings.ToLower(p.ExtraCSS), "</style") {
			t.Errorf("preset %s: ExtraCSS must not close a style element", p.Name)
		}
		// Presets are distinct looks, not palette swaps: the default preset
		// IS the base stylesheet, every other preset must restyle it.
		if p.Name == DefaultName && p.ExtraCSS != "" {
			t.Error("the default preset must not alter the base stylesheet")
		}
		if p.Name != DefaultName && p.ExtraCSS == "" {
			t.Errorf("preset %s: non-default presets must carry a structural CSS layer", p.Name)
		}
	}
}

func TestByNameFallsBackToDefault(t *testing.T) {
	if got := ByName("").Name; got != DefaultName {
		t.Errorf("ByName(\"\") = %q, want %q", got, DefaultName)
	}
	if got := ByName("no-such-preset").Name; got != DefaultName {
		t.Errorf("ByName(unknown) = %q, want %q", got, DefaultName)
	}
	if got := ByName("modern").Name; got != "modern" {
		t.Errorf("ByName(modern) = %q, want modern", got)
	}
}

func TestValidAndNames(t *testing.T) {
	if !Valid("") || !Valid("plain") {
		t.Error("empty and known names must be valid")
	}
	if Valid("neon") {
		t.Error("unknown name must be invalid")
	}
	names := Names()
	if len(names) != len(Presets()) || names[0] != DefaultName {
		t.Errorf("Names() = %v, want default first and one entry per preset", names)
	}
}
