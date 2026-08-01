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
		"static/leafpress/fonts/crimson-pro-700.woff2",
		"static/images/photo.JPG",
		"static/a/b/c/d.txt",
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
	}
	for p, why := range invalid {
		if err := ValidateLogicalPath(p); err == nil {
			t.Errorf("ValidateLogicalPath(%q) = nil, want error (%s)", p, why)
		}
	}
}

func TestValidatePublicPath(t *testing.T) {
	valid := []string{"/favicon.ico", "/static/fonts/inter-400.woff2", "/a/b/c"}
	for _, p := range valid {
		if err := ValidatePublicPath(p); err != nil {
			t.Errorf("ValidatePublicPath(%q) = %v, want nil", p, err)
		}
	}
	invalid := []string{"favicon.ico", "/a//b", "/a/../b", "/a/./b", "/a\\b", "/a\x00b", "/"}
	for _, p := range invalid {
		if err := ValidatePublicPath(p); err == nil {
			t.Errorf("ValidatePublicPath(%q) = nil, want error", p)
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
		{"empty content type", func(a *Asset) { a.ContentType = "" }},
		{"junk content type", func(a *Asset) { a.ContentType = "not a mime" }},
		{"bare content type", func(a *Asset) { a.ContentType = "woff2" }},
		{"uppercase hash", func(a *Asset) { a.SHA256 = strings.ToUpper(a.SHA256) }},
		{"short hash", func(a *Asset) { a.SHA256 = "abc123" }},
		{"empty hash", func(a *Asset) { a.SHA256 = "" }},
		{"negative size", func(a *Asset) { a.Size = -1 }},
		{"relative public path", func(a *Asset) { a.PublicPath = "favicon.ico" }},
		{"traversal public path", func(a *Asset) { a.PublicPath = "/../favicon.ico" }},
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

func TestEffectivePublicPath(t *testing.T) {
	a := validAsset()
	if got := a.EffectivePublicPath(); got != "/static/fonts/inter-400.woff2" {
		t.Fatalf("default public path = %q", got)
	}
	a.PublicPath = "/favicon.ico"
	if got := a.EffectivePublicPath(); got != "/favicon.ico" {
		t.Fatalf("explicit public path = %q", got)
	}
}

func TestManifestValidate(t *testing.T) {
	a := validAsset()
	b := validAsset()
	b.LogicalPath = "static/fonts/inter-700.woff2"

	if err := (Manifest{a, b}).Validate(); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}

	if err := (Manifest{a, a}).Validate(); err == nil {
		t.Error("duplicate logical path accepted")
	}

	// Distinct logical paths colliding on the served path.
	c := validAsset()
	c.LogicalPath = "static/fonts/other.woff2"
	c.PublicPath = "/static/fonts/inter-400.woff2"
	if err := (Manifest{a, c}).Validate(); err == nil {
		t.Error("duplicate public path accepted")
	}

	bad := validAsset()
	bad.SHA256 = "nope"
	if err := (Manifest{bad}).Validate(); err == nil {
		t.Error("manifest with invalid asset accepted")
	}
}

func TestManifestSort(t *testing.T) {
	a := validAsset()
	b := validAsset()
	b.LogicalPath = "static/a.txt"
	m := Manifest{a, b}
	m.Sort()
	if m[0].LogicalPath != "static/a.txt" {
		t.Fatalf("Sort order wrong: %q first", m[0].LogicalPath)
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
