package templates

import (
	"strings"
	"testing"
)

func TestDefaultCSSComposesBaseAndClassicTheme(t *testing.T) {
	want := BaseCSS + "\n" + ClassicCSS
	if DefaultCSS != want {
		t.Fatal("DefaultCSS does not compose the embedded base and classic theme in order")
	}
	for name, css := range map[string]string{
		"base":    BaseCSS,
		"classic": ClassicCSS,
	} {
		if strings.TrimSpace(css) == "" {
			t.Errorf("embedded %s stylesheet is empty", name)
		}
	}
	if !strings.Contains(BaseCSS, "box-sizing: border-box") {
		t.Error("base stylesheet is missing the shared box model")
	}
	for _, selector := range []string{".lp-body", ".lp-nav", ".lp-content", ".lp-search-overlay", ".lp-callout"} {
		if !strings.Contains(ClassicCSS, selector) {
			t.Errorf("classic theme is missing representative selector %q", selector)
		}
	}
}

func TestBaseCSSOwnsSemanticTypeScale(t *testing.T) {
	for _, want := range []string{
		"Semantic typography is a product invariant",
		".lp-title,\n.lp-section-title",
		".lp-content h1,\n.lp-section-intro h1",
		".lp-content h2,\n.lp-section-intro h2",
		"font-size: var(--lp-font-3xl)",
		"font-size: var(--lp-font-2xl)",
		"font-size: var(--lp-font-xl)",
	} {
		if !strings.Contains(BaseCSS, want) {
			t.Errorf("base stylesheet is missing semantic typography rule %q", want)
		}
	}
}

func TestCSSForPresetSelectsClassicAndDefaultsDefensively(t *testing.T) {
	if got := CSSForPreset("classic"); got != DefaultCSS {
		t.Error("classic preset does not reproduce DefaultCSS")
	}
	if got := CSSForPreset(""); got != DefaultCSS {
		t.Error("empty preset does not select the default stylesheet")
	}
	if got := CSSForPreset("unknown"); got != DefaultCSS {
		t.Error("unknown preset does not fall back defensively")
	}
}

func TestCSSForPresetSelectsAuroraVisualLayer(t *testing.T) {
	got := CSSForPreset("aurora")
	if got == DefaultCSS {
		t.Fatal("aurora unexpectedly reproduced the default stylesheet")
	}
	for _, want := range []string{
		"leafpress Classic Theme",
		"leafpress Aurora Theme",
		".lp-article",
		".lp-index",
		".lp-content table",
		".lp-callout",
		".lp-search-panel",
		".lp-graph-panel",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("aurora stylesheet is missing %q", want)
		}
	}
}
