package assets

import (
	"strings"
	"testing"
)

func TestBuiltinFontFamilies(t *testing.T) {
	families := BuiltinFontFamilies()
	want := []string{"Crimson Pro", "Inter", "JetBrains Mono"}
	if len(families) != len(want) {
		t.Fatalf("families = %v, want %v", families, want)
	}
	for i, f := range want {
		if families[i] != f {
			t.Fatalf("families = %v, want %v", families, want)
		}
	}
	for _, f := range want {
		if !IsBuiltinFontFamily(f) {
			t.Errorf("IsBuiltinFontFamily(%q) = false", f)
		}
	}
	if IsBuiltinFontFamily("Lobster") {
		t.Error("Lobster misreported as builtin")
	}
	if IsBuiltinFontFamily("inter") {
		t.Error("family match must be exact (CSS family names are)")
	}
}

func TestBuiltinFontFacesRegistered(t *testing.T) {
	// Every face's woff2 must exist in the registry with matching metadata.
	for _, face := range BuiltinFontFaces() {
		b, ok := BuiltinByLogicalPath(face.LogicalPath)
		if !ok {
			t.Errorf("face %s has no registry entry", face.LogicalPath)
			continue
		}
		if b.Asset.ContentType != "font/woff2" {
			t.Errorf("%s: content type %q", face.LogicalPath, b.Asset.ContentType)
		}
		if data := b.Content(); len(data) < 4 || string(data[:4]) != "wOF2" {
			t.Errorf("%s: content is not woff2", face.LogicalPath)
		}
		if face.Style != "normal" && face.Style != "italic" {
			t.Errorf("%s: style %q", face.LogicalPath, face.Style)
		}
		if face.UnicodeRange == "" || face.WeightRange == "" {
			t.Errorf("%s: missing unicode/weight range", face.LogicalPath)
		}
	}
}

func TestFontFaceCSS(t *testing.T) {
	css := FontFaceCSS("Inter")
	if !strings.Contains(css, `font-family: "Inter"`) {
		t.Error("missing Inter @font-face")
	}
	// Site-relative URL: the rules live in the stylesheet served from the
	// site root, so relative resolution lands under any base path.
	if !strings.Contains(css, `src: url("static/leafpress/fonts/inter-normal-latin.woff2") format("woff2")`) {
		t.Errorf("missing relative woff2 src:\n%s", css)
	}
	if !strings.Contains(css, "font-display: swap") {
		t.Error("missing font-display: swap")
	}
	if !strings.Contains(css, "unicode-range:") {
		t.Error("missing unicode-range")
	}
	if strings.Contains(css, "Crimson Pro") {
		t.Error("unrequested family emitted")
	}
	// 2 styles x 2 subsets for one family.
	if got := strings.Count(css, "@font-face"); got != 4 {
		t.Errorf("Inter faces = %d, want 4", got)
	}
}

func TestRequiredBuiltinsOmitsMermaidByDefault(t *testing.T) {
	for _, b := range RequiredBuiltins("Inter") {
		if isContentOptional(b.Asset.LogicalPath) {
			t.Errorf("RequiredBuiltins must omit content-optional %s", b.Asset.LogicalPath)
		}
	}
	with := RequiredBuiltinsFor(true, "Inter")
	var sawJS, sawLic bool
	for _, b := range with {
		if b.Asset.LogicalPath == BuiltinMermaidJS {
			sawJS = true
		}
		if b.Asset.LogicalPath == BuiltinMermaidLicense {
			sawLic = true
		}
	}
	if !sawJS || !sawLic {
		t.Fatalf("RequiredBuiltinsFor(true) missing mermaid assets (js=%v license=%v)", sawJS, sawLic)
	}
	if MermaidVersion == "" {
		t.Error("MermaidVersion must be pinned for maintenance")
	}
}

func TestFontFaceCSSDedup(t *testing.T) {
	css := FontFaceCSS("Inter", "Inter", "JetBrains Mono")
	if got := strings.Count(css, "@font-face"); got != 8 {
		t.Errorf("faces = %d, want 8 (Inter deduped + JetBrains Mono)", got)
	}
}

func TestFontFaceCSSUnknownFamilies(t *testing.T) {
	if css := FontFaceCSS("Lobster", "Comic Sans MS"); css != "" {
		t.Errorf("unknown families produced CSS:\n%s", css)
	}
	if css := FontFaceCSS(); css != "" {
		t.Error("no families produced CSS")
	}
}

func TestBuiltinFontLicenses(t *testing.T) {
	for _, family := range BuiltinFontFamilies() {
		licensePath, ok := BuiltinFontLicense(family)
		if !ok {
			t.Errorf("family %q has no OFL license asset", family)
			continue
		}
		b, ok := BuiltinByLogicalPath(licensePath)
		if !ok {
			t.Errorf("license %s not in registry", licensePath)
			continue
		}
		text := string(b.Content())
		if !strings.Contains(text, "SIL OPEN FONT LICENSE") || !strings.Contains(text, "Copyright") {
			t.Errorf("%s does not look like a full OFL text with copyright notice", licensePath)
		}
	}
	if _, ok := BuiltinFontLicense("Lobster"); ok {
		t.Error("unknown family reported a license")
	}
}
