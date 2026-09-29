package fonts

import (
	"encoding/json"
	"testing"
)

func TestCSSAxes(t *testing.T) {
	cases := []struct {
		name, metadata, want string
	}{
		{
			"variable with italics",
			`{"axes":[{"tag":"wght","min":400,"max":900}],"fonts":{"400":{},"400i":{},"900":{},"900i":{}}}`,
			"ital,wght@0,400..900;1,400..900",
		},
		{
			"variable without italics",
			`{"axes":[{"tag":"opsz","min":8,"max":144},{"tag":"wght","min":100,"max":700}],"fonts":{"400":{}}}`,
			"wght@100..700",
		},
		{
			"static keeps theme weights",
			`{"axes":[],"fonts":{"100":{},"400":{},"400i":{},"700":{},"900":{}}}`,
			"ital,wght@0,400;0,700;1,400",
		},
		{
			"single display weight",
			`{"axes":[],"fonts":{"900":{}}}`,
			"wght@900",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var meta metadata
			if err := json.Unmarshal([]byte(tt.metadata), &meta); err != nil {
				t.Fatal(err)
			}
			if got := cssAxes(&meta); got != tt.want {
				t.Errorf("cssAxes = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSlugAndRepoDir(t *testing.T) {
	for family, want := range map[string][2]string{
		"Source Serif 4":   {"source-serif-4", "sourceserif4"},
		"IBM Plex Mono":    {"ibm-plex-mono", "ibmplexmono"},
		"Playfair Display": {"playfair-display", "playfairdisplay"},
		" Odd  -- Name ":   {"odd-name", "oddname"},
	} {
		if got := Slug(family); got != want[0] {
			t.Errorf("Slug(%q) = %q, want %q", family, got, want[0])
		}
		if got := repoDir(family); got != want[1] {
			t.Errorf("repoDir(%q) = %q, want %q", family, got, want[1])
		}
	}
}
