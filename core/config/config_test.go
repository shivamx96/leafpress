package config

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Defaults ---

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Title != "My Garden" {
		t.Errorf("Title = %q, want %q", cfg.Title, "My Garden")
	}
	if cfg.OutputDir != "_site" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, "_site")
	}
	if cfg.Port != 3000 {
		t.Errorf("Port = %d, want 3000", cfg.Port)
	}
	if !cfg.Graph || !cfg.Search || !cfg.TOC || !cfg.Backlinks || !cfg.Wikilinks || !cfg.RSS {
		t.Error("all features should be enabled by default")
	}
	if cfg.Theme.Accent != "#50ac00" {
		t.Errorf("Accent = %q, want %q", cfg.Theme.Accent, "#50ac00")
	}
}

// --- Load ---

func TestLoad_NoFile(t *testing.T) {
	cfg, err := Load("/nonexistent/leafpress.json")
	if err != nil {
		t.Fatalf("should not error on missing file, got: %v", err)
	}
	if cfg.Title != "My Garden" {
		t.Error("should return defaults when file missing")
	}
}

func TestLoad_MinimalConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leafpress.json")
	os.WriteFile(path, []byte(`{"title": "My Site"}`), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Title != "My Site" {
		t.Errorf("Title = %q, want %q", cfg.Title, "My Site")
	}
	// Defaults should still apply
	if cfg.Port != 3000 {
		t.Errorf("Port = %d, want 3000", cfg.Port)
	}
	if cfg.Theme.FontBody != "Inter" {
		t.Errorf("FontBody = %q, want %q", cfg.Theme.FontBody, "Inter")
	}
}

func TestLoad_DisableFeatures(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "leafpress.json")
	os.WriteFile(path, []byte(`{"title": "Test", "graph": false, "rss": false}`), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Graph {
		t.Error("Graph should be false")
	}
	if cfg.RSS {
		t.Error("RSS should be false")
	}
	if !cfg.Search {
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
		"title":"Rendered Site","graph":false,"rss":false,
		"theme":{"accent":"#123456","navStyle":"sticky"}
	}`)
	cfg, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if cfg.Title != "Rendered Site" || cfg.Graph || cfg.RSS {
		t.Fatalf("overrides not preserved: %+v", cfg)
	}
	if !cfg.Search || !cfg.TOC || !cfg.Backlinks || !cfg.Wikilinks {
		t.Fatal("omitted feature flags should retain CLI defaults")
	}
	if cfg.Theme.Accent != "#123456" || cfg.Theme.NavStyle != "sticky" {
		t.Fatalf("theme overrides not preserved: %+v", cfg.Theme)
	}
	if cfg.Theme.FontBody != "Inter" || cfg.Port != 3000 || cfg.OutputDir != "_site" {
		t.Fatal("omitted config fields should retain CLI defaults")
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

	cfg.Port = 0
	if err := cfg.Validate(); err == nil {
		t.Error("port 0 should be invalid")
	}

	cfg.Port = 65536
	if err := cfg.Validate(); err == nil {
		t.Error("port 65536 should be invalid")
	}

	cfg.Port = 8080
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

	cfg.Nav = []NavItem{{Label: "Home", Path: "/home"}}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid nav should pass, got: %v", err)
	}

	cfg.Nav = []NavItem{{Label: "", Path: "/home"}}
	if err := cfg.Validate(); err == nil {
		t.Error("empty label should fail")
	}

	cfg.Nav = []NavItem{{Label: "Home", Path: ""}}
	if err := cfg.Validate(); err == nil {
		t.Error("empty path should fail")
	}

	cfg.Nav = []NavItem{{Label: "Home", Path: "home"}}
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
		"title": "Test",
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
		"title": "Test",
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
	cfg.Title = "Round Trip Test"

	if err := Write(path, cfg); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Title != "Round Trip Test" {
		t.Errorf("Title = %q, want %q", loaded.Title, "Round Trip Test")
	}
}
