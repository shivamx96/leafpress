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
