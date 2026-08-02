package assets

import (
	"strings"
	"testing"
)

func validAsset() Asset {
	return Asset{
		LogicalPath: "static/fonts/inter-400.woff2",
		ContentType: "font/woff2",
		SHA256:      Sum([]byte("font bytes")),
		Size:        10,
	}
}

func TestSum(t *testing.T) {
	got := Sum([]byte("hello"))
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Fatalf("Sum = %q, want %q", got, want)
	}
	if len(got) != 64 || got != strings.ToLower(got) {
		t.Fatalf("Sum must be 64 lowercase hex chars, got %q", got)
	}
}

func TestValidateLogicalPath(t *testing.T) {
	valid := []string{
		"static/fonts/inter-400.woff2",
		"static/leafpress/fonts/example-serif-700.woff2",
		"static/images/photo.JPG",
		"static/a/b/c/d.txt",
		"static/tilde~file.txt",
	}
	for _, p := range valid {
		if err := ValidateLogicalPath(p); err != nil {
			t.Errorf("ValidateLogicalPath(%q) = %v, want nil", p, err)
		}
	}

	invalid := map[string]string{
		"":                        "empty",
		"/static/fonts/a.woff2":   "absolute",
		"static\\fonts\\a.woff2":  "backslash",
		"static/../etc/passwd":    "traversal",
		"static/./a.woff2":        "dot segment",
		"static//a.woff2":         "empty segment",
		"static/fonts/":           "trailing slash",
		"fonts/a.woff2":           "outside static/",
		"static":                  "bare static",
		"media/a.png":             "outside static/",
		"static/fo\x00nt.woff2":   "NUL",
		"static/fo\nnt.woff2":     "newline",
		"C:/static/fonts/a.woff2": "drive letter",
		"static/C:/fonts/a.woff2": "embedded colon",
		"../static/fonts/a.woff2": "leading traversal",
		"static/fonts/..":         "trailing traversal",
		// Paths are not URLs: values that parse differently as URLs or
		// after percent-decoding must be rejected, not deduplicated wrong.
		"static/image.png?variant=1": "query syntax",
		"static/image.png#fragment":  "fragment syntax",
		"static/%2e%2e/secret":       "percent-encoded traversal",
		"static/a%2fb.png":           "percent-encoded slash",
		"static/a%20b.png":           "any percent escape",
		"static/fo\u0085nt.woff2":    "C1 control (NEL)",
		"static/fo\u009cnt.woff2":    "C1 control (ST)",
		// CSS url("...") delimiters and other non-unreserved characters:
		// these would break or escape a quoted CSS context when rendered.
		"static/fonts/custom\").woff2": "double quote and paren",
		"static/fonts/custom'.woff2":   "single quote",
		"static/fonts/my font.woff2":   "space",
		"static/fonts/a(b).woff2":      "parentheses",
		"static/fonts/a&b.woff2":       "ampersand",
		"static/fonts/a;b.woff2":       "semicolon",
	}
	for p, why := range invalid {
		if err := ValidateLogicalPath(p); err == nil {
			t.Errorf("ValidateLogicalPath(%q) = nil, want error (%s)", p, why)
		}
	}
}

func TestValidateOutputPath(t *testing.T) {
	valid := []string{"favicon.ico", "static/fonts/inter-400.woff2", "a/b/c"}
	for _, p := range valid {
		if err := ValidateOutputPath(p); err != nil {
			t.Errorf("ValidateOutputPath(%q) = %v, want nil", p, err)
		}
	}
	invalid := []string{
		"",
		"/favicon.ico", // site-relative, not absolute
		"a//b",
		"a/../b",
		"a/./b",
		"a\\b",
		"a\x00b",
		"a?x=1",
		"a#frag",
		"%2e%2e/b",
		"a\u0085b",
	}
	for _, p := range invalid {
		if err := ValidateOutputPath(p); err == nil {
			t.Errorf("ValidateOutputPath(%q) = nil, want error", p)
		}
	}
}

func TestAssetValidate(t *testing.T) {
	if err := validAsset().Validate(); err != nil {
		t.Fatalf("valid asset rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Asset)
	}{
		{"traversal path", func(a *Asset) { a.LogicalPath = "static/../a.woff2" }},
		{"query in path", func(a *Asset) { a.LogicalPath = "static/a.woff2?v=1" }},
		{"empty content type", func(a *Asset) { a.ContentType = "" }},
		{"junk content type", func(a *Asset) { a.ContentType = "not a mime" }},
		{"bare content type", func(a *Asset) { a.ContentType = "woff2" }},
		{"uppercase hash", func(a *Asset) { a.SHA256 = strings.ToUpper(a.SHA256) }},
		{"short hash", func(a *Asset) { a.SHA256 = "abc123" }},
		{"empty hash", func(a *Asset) { a.SHA256 = "" }},
		{"negative size", func(a *Asset) { a.Size = -1 }},
		{"absolute output path", func(a *Asset) { a.OutputPath = "/favicon.ico" }},
		{"traversal output path", func(a *Asset) { a.OutputPath = "../favicon.ico" }},
		{"fragment output path", func(a *Asset) { a.OutputPath = "favicon.ico#x" }},
	}
	for _, tt := range tests {
		a := validAsset()
		tt.mutate(&a)
		if err := a.Validate(); err == nil {
			t.Errorf("%s: Validate() = nil, want error", tt.name)
		}
	}
}

func TestAssetValidateAllowsContentTypeParams(t *testing.T) {
	a := validAsset()
	a.ContentType = "text/plain; charset=utf-8"
	if err := a.Validate(); err != nil {
		t.Fatalf("content type with params rejected: %v", err)
	}
}

func TestEffectiveOutputPath(t *testing.T) {
	a := validAsset()
	if got := a.EffectiveOutputPath(); got != "static/fonts/inter-400.woff2" {
		t.Fatalf("default output path = %q", got)
	}
	a.OutputPath = "favicon.ico"
	if got := a.EffectiveOutputPath(); got != "favicon.ico" {
		t.Fatalf("explicit output path = %q", got)
	}
}

func TestNewManifestCanonicalizes(t *testing.T) {
	a := validAsset()
	b := validAsset()
	b.LogicalPath = "static/a.txt"

	m, err := NewManifest(a, b)
	if err != nil {
		t.Fatalf("NewManifest: %v", err)
	}
	if m[0].LogicalPath != "static/a.txt" {
		t.Fatalf("NewManifest must sort; got %q first", m[0].LogicalPath)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("canonical manifest failed validation: %v", err)
	}
	// The input slice is not mutated.
	if a.LogicalPath != "static/fonts/inter-400.woff2" {
		t.Error("NewManifest mutated its input")
	}
}

func TestManifestValidateRequiresCanonicalOrder(t *testing.T) {
	a := validAsset()
	b := validAsset()
	b.LogicalPath = "static/a.txt"

	// Descending order is individually valid but not canonical.
	if err := (Manifest{a, b}).Validate(); err == nil {
		t.Error("unsorted manifest accepted: a deterministic manifest must be sorted")
	}
	if err := (Manifest{b, a}).Validate(); err != nil {
		t.Errorf("sorted manifest rejected: %v", err)
	}
}

func TestManifestValidateDuplicates(t *testing.T) {
	a := validAsset()

	if err := (Manifest{a, a}).Validate(); err == nil {
		t.Error("duplicate logical path accepted")
	}

	// Distinct logical paths colliding on the effective output path.
	c := validAsset()
	c.LogicalPath = "static/other.woff2"
	c.OutputPath = "static/fonts/inter-400.woff2"
	if _, err := NewManifest(a, c); err == nil {
		t.Error("duplicate output path accepted")
	}

	bad := validAsset()
	bad.SHA256 = "nope"
	if _, err := NewManifest(bad); err == nil {
		t.Error("manifest with invalid asset accepted")
	}
}

func TestMergeOverridesByOutputPath(t *testing.T) {
	builtin := validAsset()
	builtin.LogicalPath = BuiltinPrefix + "favicon.ico"
	builtin.OutputPath = "favicon.ico"

	font := validAsset()

	base, err := NewManifest(builtin, font)
	if err != nil {
		t.Fatal(err)
	}

	// A user favicon on the same output path replaces the built-in entry.
	userFavicon := Asset{
		LogicalPath: "static/user-favicon.ico",
		ContentType: "image/x-icon",
		SHA256:      Sum([]byte("user bytes")),
		Size:        10,
		OutputPath:  "favicon.ico",
	}
	extra := Asset{
		LogicalPath: "static/fonts/custom.woff2",
		ContentType: "font/woff2",
		SHA256:      Sum([]byte("custom")),
		Size:        6,
	}

	merged, err := Merge(base, Manifest{userFavicon, extra})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if len(merged) != 3 {
		t.Fatalf("merged has %d entries, want 3 (override replaced, extra appended)", len(merged))
	}
	var atFavicon *Asset
	for i := range merged {
		if merged[i].EffectiveOutputPath() == "favicon.ico" {
			atFavicon = &merged[i]
		}
	}
	if atFavicon == nil || atFavicon.SHA256 != userFavicon.SHA256 {
		t.Error("override did not replace the built-in favicon entry")
	}
	if err := merged.Validate(); err != nil {
		t.Fatalf("merged manifest not canonical: %v", err)
	}
}

func TestIsBuiltinPath(t *testing.T) {
	if !IsBuiltinPath("static/leafpress/fonts/inter-400.woff2") {
		t.Error("builtin path not recognized")
	}
	if IsBuiltinPath("static/fonts/inter-400.woff2") {
		t.Error("user path misclassified as builtin")
	}
}

func TestEscapedURLPath(t *testing.T) {
	// Canonical paths are escape-free: identity.
	if got := EscapedURLPath("static/fonts/inter-400.woff2"); got != "static/fonts/inter-400.woff2" {
		t.Fatalf("EscapedURLPath changed a canonical path: %q", got)
	}
	// Defense in depth: a hostile segment renders inert, slashes preserved.
	got := EscapedURLPath(`static/fonts/a").woff2`)
	if strings.Contains(got, `"`) {
		t.Fatalf("EscapedURLPath left a quote unescaped: %q", got)
	}
	if strings.Count(got, "/") != 2 {
		t.Fatalf("EscapedURLPath must preserve segment separators: %q", got)
	}
}

func TestValidatePathWindowsPortability(t *testing.T) {
	invalid := map[string]string{
		"static/CON":             "bare device name",
		"static/con.woff2":       "device name with extension, lowercase",
		"static/fonts/COM1.txt":  "COM device with extension",
		"static/NUL/a.woff2":     "device name as directory",
		"static/fonts/aux":       "aux lowercase",
		"static/fonts/font.":     "trailing dot",
		"static/trailing./a.txt": "trailing dot in directory",
	}
	for p, why := range invalid {
		if err := ValidateLogicalPath(p); err == nil {
			t.Errorf("ValidateLogicalPath(%q) = nil, want error (%s)", p, why)
		}
	}
	// Names merely containing device substrings are fine.
	valid := []string{"static/fonts/console.woff2", "static/fonts/communal.txt", "static/nullable.md"}
	for _, p := range valid {
		if err := ValidateLogicalPath(p); err != nil {
			t.Errorf("ValidateLogicalPath(%q) = %v, want nil", p, err)
		}
	}
}

func TestIsBuiltinPathMatchesBareDirectory(t *testing.T) {
	if !IsBuiltinPath("static/leafpress") {
		t.Error("bare static/leafpress must count as the reserved namespace")
	}
	if IsBuiltinPath("static/leafpress-extras/a.txt") {
		t.Error("sibling directory misclassified as reserved")
	}
}

func TestManifestValidateRejectsCaseFoldCollisions(t *testing.T) {
	a := validAsset()
	b := validAsset()
	b.LogicalPath = "static/fonts/Inter-400.woff2" // differs only by case

	if _, err := NewManifest(a, b); err == nil {
		t.Error("case-fold logical collision accepted: macOS/Windows would materialize one file")
	}

	// Distinct logical paths whose effective output paths case-collide.
	c := validAsset()
	c.LogicalPath = "static/other.ico"
	c.OutputPath = "Favicon.ico"
	d := validAsset()
	d.LogicalPath = "static/another.ico"
	d.OutputPath = "favicon.ico"
	if _, err := NewManifest(c, d); err == nil {
		t.Error("case-fold output collision accepted")
	}
}
