package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Defaults ---

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.ContractVersion != ContractVersionLatest {
		t.Errorf("ContractVersion = %d, want %d", cfg.ContractVersion, ContractVersionLatest)
	}
	if cfg.Site.Title != "My Garden" {
		t.Errorf("Title = %q, want %q", cfg.Site.Title, "My Garden")
	}
	if cfg.Build.OutputDir != "_site" {
		t.Errorf("OutputDir = %q, want %q", cfg.Build.OutputDir, "_site")
	}
	if cfg.Build.Port != 3000 {
		t.Errorf("Port = %d, want 3000", cfg.Build.Port)
	}
	if cfg.Navigation.Mode != NavAutomatic {
		t.Errorf("Navigation.Mode = %q, want %q", cfg.Navigation.Mode, NavAutomatic)
	}
	if !cfg.Features.Graph || !cfg.Features.Search || !cfg.Features.TOC || !cfg.Features.Backlinks || !cfg.Features.Wikilinks || !cfg.Features.RSS {
		t.Error("all features should be enabled by default")
	}
	if cfg.Theme.Accent != "#50ac00" {
		t.Errorf("Accent = %q, want %q", cfg.Theme.Accent, "#50ac00")
	}
	if cfg.Theme.FontHeading != "Bricolage Grotesque" {
		t.Errorf("FontHeading = %q, want %q", cfg.Theme.FontHeading, "Bricolage Grotesque")
	}
}

// --- Load ---

func TestLoad_NoFile(t *testing.T) {
	cfg, err := Load("/nonexistent/leafpress.json")
	if err != nil {
		t.Fatalf("should not error on missing file, got: %v", err)
	}
	if cfg.Site.Title != "My Garden" {
		t.Error("should return defaults when file missing")
	}
}

func TestLoad_MinimalConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leafpress.json")
	os.WriteFile(path, []byte(`{"site": {"title": "My Site"}}`), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Site.Title != "My Site" {
		t.Errorf("Title = %q, want %q", cfg.Site.Title, "My Site")
	}
	// Defaults should still apply
	if cfg.Build.Port != 3000 {
		t.Errorf("Port = %d, want 3000", cfg.Build.Port)
	}
	if cfg.Theme.FontBody != "Inter" {
		t.Errorf("FontBody = %q, want %q", cfg.Theme.FontBody, "Inter")
	}
}

func TestLoad_DisableFeatures(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leafpress.json")
	os.WriteFile(path, []byte(`{"site": {"title": "Test"}, "features": {"graph": false, "rss": false}}`), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Features.Graph {
		t.Error("Graph should be false")
	}
	if cfg.Features.RSS {
		t.Error("RSS should be false")
	}
	if !cfg.Features.Search {
		t.Error("Search should still be true (default)")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leafpress.json")
	os.WriteFile(path, []byte(`{invalid json`), 0644)

	_, err := Load(path)
	if err == nil {
		t.Error("should error on invalid JSON")
	}
}

func TestParse_MatchesLoadDefaultsAndOverrides(t *testing.T) {
	raw := []byte(`{
		"site":{"title":"Rendered Site"},
		"features":{"graph":false,"rss":false},
		"theme":{"accent":"#123456","navStyle":"sticky"}
	}`)
	cfg, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Site.Title != "Rendered Site" || cfg.Features.Graph || cfg.Features.RSS {
		t.Fatalf("overrides not preserved: %+v", cfg)
	}
	if !cfg.Features.Search || !cfg.Features.TOC || !cfg.Features.Backlinks || !cfg.Features.Wikilinks {
		t.Fatal("omitted feature flags should retain CLI defaults")
	}
	if cfg.Theme.Accent != "#123456" || cfg.Theme.NavStyle != "sticky" {
		t.Fatalf("theme overrides not preserved: %+v", cfg.Theme)
	}
	if cfg.Theme.FontBody != "Inter" || cfg.Build.Port != 3000 || cfg.Build.OutputDir != "_site" {
		t.Fatal("omitted config fields should retain CLI defaults")
	}
	if cfg.Navigation.Mode != NavAutomatic {
		t.Fatalf("omitted navigation.mode should default to %q, got %q", NavAutomatic, cfg.Navigation.Mode)
	}
}

func TestParse_MinimalEmptyConfig(t *testing.T) {
	cfg, err := Parse([]byte(`{}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Site.Title != "My Garden" {
		t.Errorf("empty config should yield default site, got Title = %q", cfg.Site.Title)
	}
}

func TestParse_UnsupportedContractVersion(t *testing.T) {
	if _, err := Parse([]byte(`{"contractVersion": 3}`)); err == nil {
		t.Error("contractVersion 3 should be rejected")
	}
}

func TestParse_InvalidNavigationMode(t *testing.T) {
	if _, err := Parse([]byte(`{"navigation": {"mode": "sideways"}}`)); err == nil {
		t.Error("navigation mode 'sideways' should be rejected")
	}
}

// Unknown / misspelled keys must be rejected at every nesting level rather
// than silently ignored, so leafpress.json typos surface as build errors.
func TestParse_RejectsUnknownFields(t *testing.T) {
	cases := map[string]string{
		"top-level":  `{"nope": 1}`,
		"site":       `{"site": {"titel": "T"}}`,
		"features":   `{"features": {"grph": true}}`,
		"navigation": `{"navigation": {"modee": "automatic"}}`,
		"build":      `{"build": {"prt": 8080}}`,
		"theme":      `{"theme": {"acent": "#ffffff"}}`,
		"deploy":     `{"deploy": {"provder": "github-pages"}}`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(in)); err == nil {
				t.Errorf("unknown %s field should be rejected", name)
			}
		})
	}
}

// --- Validation ---

func TestValidate_ValidConfig(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Errorf("default config should be valid, got: %v", err)
	}
}

func TestValidate_PortRange(t *testing.T) {
	cfg := Default()

	cfg.Build.Port = 0
	if err := cfg.Validate(); err == nil {
		t.Error("port 0 should be invalid")
	}

	cfg.Build.Port = 65536
	if err := cfg.Validate(); err == nil {
		t.Error("port 65536 should be invalid")
	}

	cfg.Build.Port = 8080
	if err := cfg.Validate(); err != nil {
		t.Errorf("port 8080 should be valid, got: %v", err)
	}
}

func TestValidate_AccentColor(t *testing.T) {
	cfg := Default()

	cfg.Theme.Accent = "#abc"
	if err := cfg.Validate(); err != nil {
		t.Errorf("3-digit hex should be valid, got: %v", err)
	}

	cfg.Theme.Accent = "#abcdef"
	if err := cfg.Validate(); err != nil {
		t.Errorf("6-digit hex should be valid, got: %v", err)
	}

	cfg.Theme.Accent = "red"
	if err := cfg.Validate(); err == nil {
		t.Error("non-hex color should be invalid")
	}

	cfg.Theme.Accent = "#xyz"
	if err := cfg.Validate(); err == nil {
		t.Error("invalid hex should be invalid")
	}
}

func TestValidate_NavStyle(t *testing.T) {
	cfg := Default()

	for _, style := range []string{"base", "sticky", "glassy"} {
		cfg.Theme.NavStyle = style
		if err := cfg.Validate(); err != nil {
			t.Errorf("navStyle %q should be valid, got: %v", style, err)
		}
	}

	cfg.Theme.NavStyle = "floating"
	if err := cfg.Validate(); err == nil {
		t.Error("invalid navStyle should fail")
	}
}

func TestValidate_NavActiveStyle(t *testing.T) {
	cfg := Default()

	for _, style := range []string{"base", "box", "underlined"} {
		cfg.Theme.NavActiveStyle = style
		if err := cfg.Validate(); err != nil {
			t.Errorf("navActiveStyle %q should be valid, got: %v", style, err)
		}
	}

	cfg.Theme.NavActiveStyle = "dotted"
	if err := cfg.Validate(); err == nil {
		t.Error("invalid navActiveStyle should fail")
	}
}

func TestValidate_NavItems(t *testing.T) {
	cfg := Default()

	cfg.Navigation.Items = []NavItem{{Label: "Home", Path: "/home"}}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid nav should pass, got: %v", err)
	}

	cfg.Navigation.Items = []NavItem{{Label: "", Path: "/home"}}
	if err := cfg.Validate(); err == nil {
		t.Error("empty label should fail")
	}

	cfg.Navigation.Items = []NavItem{{Label: "Home", Path: ""}}
	if err := cfg.Validate(); err == nil {
		t.Error("empty path should fail")
	}

	cfg.Navigation.Items = []NavItem{{Label: "Home", Path: "home"}}
	if err := cfg.Validate(); err == nil {
		t.Error("path without leading / should fail")
	}
}

// --- Background ---

func TestValidateBackground(t *testing.T) {
	valid := []string{
		"#fff",
		"#ffffff",
		"rgb(255, 255, 255)",
		"rgba(0, 0, 0, 0.5)",
		"linear-gradient(180deg, #fff 0%, #000 100%)",
		"transparent",
	}
	for _, bg := range valid {
		if err := validateBackground(bg); err != nil {
			t.Errorf("validateBackground(%q) should pass, got: %v", bg, err)
		}
	}

	invalid := []string{
		"",
		`<script>alert('xss')</script>`,
		"javascript:void(0)",
	}
	for _, bg := range invalid {
		if err := validateBackground(bg); err == nil {
			t.Errorf("validateBackground(%q) should fail", bg)
		}
	}
}

// --- Theme unmarshaling ---

func TestTheme_BackgroundObject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leafpress.json")
	os.WriteFile(path, []byte(`{
		"site": {"title": "Test"},
		"theme": {
			"background": {
				"light": "#ffffff",
				"dark": "#1a1a1a"
			}
		}
	}`), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme.Background.Light != "#ffffff" {
		t.Errorf("Light = %q, want %q", cfg.Theme.Background.Light, "#ffffff")
	}
	if cfg.Theme.Background.Dark != "#1a1a1a" {
		t.Errorf("Dark = %q, want %q", cfg.Theme.Background.Dark, "#1a1a1a")
	}
}

func TestTheme_BackgroundString(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leafpress.json")
	os.WriteFile(path, []byte(`{
		"site": {"title": "Test"},
		"theme": {
			"background": "#f0f0f0"
		}
	}`), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme.Background.Light != "#f0f0f0" {
		t.Errorf("Light = %q, want %q", cfg.Theme.Background.Light, "#f0f0f0")
	}
	if cfg.Theme.Background.Dark != "" {
		t.Errorf("Dark should be empty for string background, got %q", cfg.Theme.Background.Dark)
	}
}

// --- Write ---

func TestWrite_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leafpress.json")

	cfg := Default()
	cfg.Site.Title = "Round Trip Test"

	if err := Write(path, cfg); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Site.Title != "Round Trip Test" {
		t.Errorf("Title = %q, want %q", loaded.Site.Title, "Round Trip Test")
	}
	// Nested sections must survive the flat->nested JSON round trip.
	if loaded.ContractVersion != cfg.ContractVersion {
		t.Errorf("ContractVersion = %d, want %d", loaded.ContractVersion, cfg.ContractVersion)
	}
	if loaded.Features != cfg.Features {
		t.Errorf("Features = %+v, want %+v", loaded.Features, cfg.Features)
	}
	if loaded.Navigation.Mode != cfg.Navigation.Mode {
		t.Errorf("Navigation.Mode = %q, want %q", loaded.Navigation.Mode, cfg.Navigation.Mode)
	}
	if loaded.Build.OutputDir != cfg.Build.OutputDir || loaded.Build.Port != cfg.Build.Port {
		t.Errorf("Build = %+v, want %+v", loaded.Build, cfg.Build)
	}
}

func TestCustomFontFaceValidation(t *testing.T) {
	valid := []FontFace{
		{Family: "My Serif", File: "static/fonts/my-serif.woff2"},
		{Family: "My Serif", File: "static/fonts/my-serif-italic.woff2", Weight: "400 700", Style: "italic", Display: "swap"},
		{Family: "Mono_2", File: "static/fonts/sub/dir/mono.otf", Weight: "650"},
		{Family: "Old", File: "static/fonts/old.TTF"},
	}
	for _, f := range valid {
		cfg := Default()
		cfg.Theme.Fonts = []FontFace{f}
		if err := cfg.Validate(); err != nil {
			t.Errorf("valid declaration %+v rejected: %v", f, err)
		}
	}

	invalid := []struct {
		name string
		face FontFace
	}{
		{"missing family", FontFace{File: "static/fonts/a.woff2"}},
		{"unsafe family", FontFace{Family: "Evil\"><script>", File: "static/fonts/a.woff2"}},
		{"missing file", FontFace{Family: "A"}},
		{"outside fonts dir", FontFace{Family: "A", File: "static/other/a.woff2"}},
		{"builtin namespace", FontFace{Family: "A", File: "static/leafpress/fonts/a.woff2"}},
		{"absolute path", FontFace{Family: "A", File: "/static/fonts/a.woff2"}},
		{"traversal", FontFace{Family: "A", File: "static/fonts/../../etc/passwd"}},
		{"remote-ish file", FontFace{Family: "A", File: "static/fonts/https://evil.example/x.woff2"}},
		{"bad extension", FontFace{Family: "A", File: "static/fonts/a.exe"}},
		{"no extension", FontFace{Family: "A", File: "static/fonts/afont"}},
		{"bad weight", FontFace{Family: "A", File: "static/fonts/a.woff2", Weight: "bold"}},
		{"zero weight", FontFace{Family: "A", File: "static/fonts/a.woff2", Weight: "0"}},
		{"huge weight", FontFace{Family: "A", File: "static/fonts/a.woff2", Weight: "1001"}},
		{"inverted range", FontFace{Family: "A", File: "static/fonts/a.woff2", Weight: "700 400"}},
		{"bad style", FontFace{Family: "A", File: "static/fonts/a.woff2", Style: "cursive"}},
		{"bad display", FontFace{Family: "A", File: "static/fonts/a.woff2", Display: "eager"}},
	}
	for _, tt := range invalid {
		cfg := Default()
		cfg.Theme.Fonts = []FontFace{tt.face}
		if err := cfg.Validate(); err == nil {
			t.Errorf("%s: accepted", tt.name)
		}
	}
}

func TestCustomFontsParsedFromJSON(t *testing.T) {
	cfg, err := Parse([]byte(`{
		"theme": {
			"fontBody": "My Serif",
			"fonts": [{"family": "My Serif", "file": "static/fonts/my.woff2", "weight": "400 700"}]
		}
	}`))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(cfg.Theme.Fonts) != 1 || cfg.Theme.Fonts[0].Family != "My Serif" {
		t.Fatalf("fonts not parsed: %+v", cfg.Theme.Fonts)
	}
	if !cfg.Theme.DeclaresFamily("My Serif") || cfg.Theme.DeclaresFamily("Other") {
		t.Error("DeclaresFamily wrong")
	}

	if _, err := Parse([]byte(`{
		"theme": {"fonts": [{"family": "X", "file": "static/evil/../../x.woff2"}]}
	}`)); err == nil {
		t.Error("invalid font declaration accepted by Parse")
	}
}

func TestThemeBackgroundMarshalRoundTrip(t *testing.T) {
	// String form: light only, dark keeps defaults.
	cfg := Default()
	cfg.Theme.Background = Background{Light: "#fdf6e3"}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"background":"#fdf6e3"`) {
		t.Errorf("light-only background should marshal as string, got %s", data)
	}
	back, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if back.Theme.Background != cfg.Theme.Background {
		t.Errorf("round trip = %+v, want %+v", back.Theme.Background, cfg.Theme.Background)
	}

	// Object form: both modes customized.
	cfg.Theme.Background = Background{Light: "#ffffff", Dark: "#101010"}
	data, err = json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	back, err = Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if back.Theme.Background != cfg.Theme.Background {
		t.Errorf("round trip = %+v, want %+v", back.Theme.Background, cfg.Theme.Background)
	}

	// Unset: field omitted entirely.
	cfg.Theme.Background = Background{}
	data, err = json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "background") {
		t.Errorf("unset background should be omitted, got %s", data)
	}
}

func TestWriteLoadRoundTripPreservesTheme(t *testing.T) {
	path := filepath.Join(t.TempDir(), "leafpress.json")
	cfg := Default()
	cfg.Theme.Background = Background{Light: "#fdf6e3", Dark: "#002b36"}
	cfg.Theme.FontBody = "My Serif"
	cfg.Theme.Fonts = []FontFace{{Family: "My Serif", File: "static/fonts/my.woff2", Weight: "400 700"}}

	if err := Write(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Theme.Background != cfg.Theme.Background {
		t.Errorf("background lost in Write/Load: %+v", loaded.Theme.Background)
	}
	if len(loaded.Theme.Fonts) != 1 || loaded.Theme.Fonts[0] != cfg.Theme.Fonts[0] {
		t.Errorf("custom fonts lost in Write/Load: %+v", loaded.Theme.Fonts)
	}
}

func TestThemeFontNamesValidatedCentrally(t *testing.T) {
	cfg := Default()
	cfg.Theme.FontBody = `Evil"; @import url(x); "`
	if err := cfg.Validate(); err == nil {
		t.Error("unsafe font family accepted by config.Validate")
	}
	cfg = Default()
	cfg.Theme.FontBody = ""
	if err := cfg.Validate(); err != nil {
		t.Errorf("empty font (pre-default) should validate: %v", err)
	}
}
