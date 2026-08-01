package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/shivamx96/leafpress/core/assets"
)

// Background represents background configuration that can be a string or object
type Background struct {
	Light string
	Dark  string
}

// Config represents the leafpress.json configuration
type Config struct {
	Title       string       `json:"title"`
	Description string       `json:"description"` // Site-wide meta description
	Author      string       `json:"author"`
	BaseURL     string       `json:"baseURL"`
	Image       string       `json:"image"` // Default OG image path (e.g., "/og-image.png")
	OutputDir   string       `json:"outputDir"`
	Port        int          `json:"port"`
	Nav         []NavItem    `json:"nav"`
	Theme       Theme        `json:"theme"`
	Graph       bool         `json:"graph"`
	Search      bool         `json:"search"`
	TOC         bool         `json:"toc"`
	Backlinks   bool         `json:"backlinks"`
	Wikilinks   bool         `json:"wikilinks"`
	RSS         bool         `json:"rss"`
	Ignore      []string     `json:"ignore"`
	HeadExtra   string       `json:"headExtra"` // Custom HTML to inject in <head>
	Deploy      DeployConfig `json:"deploy"`    // Deployment configuration
}

// DeployConfig holds deployment settings
type DeployConfig struct {
	Provider string            `json:"provider"` // e.g., "github-pages", "netlify", "vercel"
	Settings map[string]string `json:"settings"` // Provider-specific settings
}

// NavItem represents a navigation link
type NavItem struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

// Theme represents theme configuration
type Theme struct {
	FontHeading string `json:"fontHeading"`
	FontBody    string `json:"fontBody"`
	FontMono    string `json:"fontMono"`
	// Fonts declares custom local font files under static/fonts/.
	Fonts []FontFace `json:"fonts,omitempty"`
	// RemoteFonts is a deprecated compatibility escape hatch: when true,
	// families outside the bundled set load from Google Fonts as before.
	// The default is self-contained output — unknown families produce a
	// warning and fall back to the CSS system stacks. This flag will be
	// removed.
	RemoteFonts    bool       `json:"remoteFonts,omitempty"`
	Accent         string     `json:"accent"`
	Background     Background `json:"-"`              // Custom unmarshaling
	NavStyle       string     `json:"navStyle"`       // "base", "sticky", or "glassy"
	NavActiveStyle string     `json:"navActiveStyle"` // "base", "box", or "underlined"
}

// FontFace declares one custom local font file under static/fonts/. Families
// declared here are self-hosted: they never load from a remote provider.
// Weight, Style, and Display are optional; consumers treat empty values as
// "400", "normal", and "swap".
type FontFace struct {
	Family  string `json:"family"`            // CSS font-family name
	File    string `json:"file"`              // logical path under static/fonts/
	Weight  string `json:"weight,omitempty"`  // "400" or a variable range "400 700"
	Style   string `json:"style,omitempty"`   // "normal", "italic", or "oblique"
	Display string `json:"display,omitempty"` // CSS font-display value
}

// DeclaresFamily reports whether the theme declares family as a custom
// local font.
func (t Theme) DeclaresFamily(family string) bool {
	for _, f := range t.Fonts {
		if f.Family == family {
			return true
		}
	}
	return false
}

// MarshalJSON round-trips the background field UnmarshalJSON accepts: the
// string form when only light is set (dark keeps defaults), the object form
// when dark is customized, omitted entirely when unset. Without this, the
// `json:"-"` tag on Background made config.Write silently drop it.
func (t Theme) MarshalJSON() ([]byte, error) {
	type Alias Theme
	aux := struct {
		Background any `json:"background,omitempty"`
		Alias
	}{
		Alias: (Alias)(t),
	}
	switch {
	case t.Background.Light != "" && t.Background.Dark == "":
		aux.Background = t.Background.Light
	case t.Background.Light != "" || t.Background.Dark != "":
		aux.Background = map[string]string{
			"light": t.Background.Light,
			"dark":  t.Background.Dark,
		}
	}
	return json.Marshal(aux)
}

// UnmarshalJSON implements custom JSON unmarshaling for Theme
func (t *Theme) UnmarshalJSON(data []byte) error {
	// Create a temporary struct to avoid recursion
	type Alias Theme
	aux := &struct {
		Background json.RawMessage `json:"background,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle background field
	if len(aux.Background) > 0 {
		// Try to unmarshal as object first
		var bgObj struct {
			Light string `json:"light"`
			Dark  string `json:"dark"`
		}
		if err := json.Unmarshal(aux.Background, &bgObj); err == nil {
			t.Background = Background{
				Light: bgObj.Light,
				Dark:  bgObj.Dark,
			}
		} else {
			// Try as string - only apply to light mode, dark mode keeps defaults
			var bgStr string
			if err := json.Unmarshal(aux.Background, &bgStr); err == nil {
				t.Background = Background{
					Light: bgStr,
					Dark:  "", // Empty means use default dark background
				}
			} else {
				return fmt.Errorf("background must be a string or object with light/dark fields")
			}
		}
	}

	return nil
}

// validateBackground checks if a background value is valid
func validateBackground(bg string) error {
	// Check for common CSS background patterns
	// Allow: hex colors, rgb/rgba, gradients, keywords
	bg = strings.TrimSpace(bg)
	if bg == "" {
		return fmt.Errorf("background cannot be empty")
	}

	// Check for dangerous patterns (script injection)
	dangerous := []string{"<script", "javascript:", "onerror=", "onload="}
	bgLower := strings.ToLower(bg)
	for _, pattern := range dangerous {
		if strings.Contains(bgLower, pattern) {
			return fmt.Errorf("background contains potentially dangerous content")
		}
	}

	// Valid patterns: hex color, rgb/rgba, hsl/hsla, gradients, keywords
	validPatterns := []string{
		`^#[0-9A-Fa-f]{3}([0-9A-Fa-f]{3})?$`, // hex color
		`^rgb\(`,                             // rgb()
		`^rgba\(`,                            // rgba()
		`^hsl\(`,                             // hsl()
		`^hsla\(`,                            // hsla()
		`^linear-gradient\(`,                 // linear-gradient()
		`^radial-gradient\(`,                 // radial-gradient()
		`^conic-gradient\(`,                  // conic-gradient()
		`^repeating-linear-gradient\(`,       // repeating-linear-gradient()
		`^repeating-radial-gradient\(`,       // repeating-radial-gradient()
		`^(transparent|white|black|gray|silver|red|blue|green|yellow|orange)$`, // color keywords
	}

	for _, pattern := range validPatterns {
		matched, _ := regexp.MatchString(pattern, bg)
		if matched {
			return nil
		}
	}

	return fmt.Errorf("invalid CSS background value: %s (must be a hex color, rgb/rgba, gradient, or color keyword)", bg)
}

// CustomFontDir is the logical-path prefix custom font files must live under.
const CustomFontDir = "static/fonts/"

var (
	// fontFamilyRegex matches the safe font-name charset also enforced at
	// the renderer boundary; family names are interpolated into CSS.
	fontFamilyRegex = regexp.MustCompile(`^[A-Za-z0-9 _-]+$`)
	fontWeightRegex = regexp.MustCompile(`^[0-9]{1,4}( [0-9]{1,4})?$`)
)

// fontFileExtensions are the allowed custom font formats.
var fontFileExtensions = map[string]bool{
	".woff2": true,
	".woff":  true,
	".ttf":   true,
	".otf":   true,
}

// validateFontFace checks one custom font declaration. Only the declaration
// shape is validated here — file existence is a CLI concern (the renderer has
// no filesystem; hosted consumers resolve files via their storage manifest).
func validateFontFace(f FontFace) error {
	if f.Family == "" {
		return fmt.Errorf("family is required")
	}
	if !fontFamilyRegex.MatchString(f.Family) {
		return fmt.Errorf("family %q may only contain letters, digits, spaces, hyphens, and underscores", f.Family)
	}
	if f.File == "" {
		return fmt.Errorf("file is required")
	}
	if err := assets.ValidateLogicalPath(f.File); err != nil {
		return fmt.Errorf("file: %w", err)
	}
	if !strings.HasPrefix(f.File, CustomFontDir) {
		return fmt.Errorf("file %q must be under %s", f.File, CustomFontDir)
	}
	ext := strings.ToLower(filepath.Ext(f.File))
	if !fontFileExtensions[ext] {
		return fmt.Errorf("file %q must be a .woff2, .woff, .ttf, or .otf font", f.File)
	}
	if f.Weight != "" {
		if !fontWeightRegex.MatchString(f.Weight) {
			return fmt.Errorf("weight %q must be a number (e.g. 400) or range (e.g. 400 700)", f.Weight)
		}
		parts := strings.Fields(f.Weight)
		values := make([]int, len(parts))
		for i, part := range parts {
			n, err := strconv.Atoi(part)
			if err != nil || n < 1 || n > 1000 {
				return fmt.Errorf("weight %q values must be between 1 and 1000", f.Weight)
			}
			values[i] = n
		}
		if len(values) == 2 && values[0] > values[1] {
			return fmt.Errorf("weight range %q must be low-to-high", f.Weight)
		}
	}
	switch f.Style {
	case "", "normal", "italic", "oblique":
	default:
		return fmt.Errorf("style must be normal, italic, or oblique, got %q", f.Style)
	}
	switch f.Display {
	case "", "auto", "block", "swap", "fallback", "optional":
	default:
		return fmt.Errorf("display must be auto, block, swap, fallback, or optional, got %q", f.Display)
	}
	return nil
}

// Default returns a Config with default values
func Default() *Config {
	return &Config{
		Title:     "My Garden",
		BaseURL:   "",
		OutputDir: "_site",
		Port:      3000,
		Nav:       []NavItem{},
		Theme: Theme{
			FontHeading:    "Crimson Pro",
			FontBody:       "Inter",
			FontMono:       "JetBrains Mono",
			Accent:         "#50ac00",
			NavStyle:       "base",
			NavActiveStyle: "base",
		},
		Graph:     true,
		Search:    true,
		TOC:       true,
		Backlinks: true,
		Wikilinks: true,
		RSS:       true,
	}
}

// Load reads and parses the config file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults if no config file
			return Default(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	return Parse(data)
}

// Parse decodes leafpress.json bytes using the exact same default overlay and
// validation as Load. It is the filesystem-free entry point for embedders
// such as leafpress-render, preventing the JSON bridge and CLI from drifting.
func Parse(data []byte) (*Config, error) {
	cfg := Default()

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults for missing values
	if cfg.OutputDir == "" {
		cfg.OutputDir = "_site"
	}
	if cfg.Port == 0 {
		cfg.Port = 3000
	}
	if cfg.Theme.FontHeading == "" {
		cfg.Theme.FontHeading = "Crimson Pro"
	}
	if cfg.Theme.FontBody == "" {
		cfg.Theme.FontBody = "Inter"
	}
	if cfg.Theme.FontMono == "" {
		cfg.Theme.FontMono = "JetBrains Mono"
	}
	if cfg.Theme.Accent == "" {
		cfg.Theme.Accent = "#50ac00"
	}
	if cfg.Theme.NavStyle == "" {
		cfg.Theme.NavStyle = "base"
	}
	if cfg.Theme.NavActiveStyle == "" {
		cfg.Theme.NavActiveStyle = "base"
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Write saves the config to a file
func Write(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Validate checks if the configuration values are valid
func (c *Config) Validate() error {
	// Validate port range
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	// Validate output directory is not a dangerous path
	absPath, err := filepath.Abs(c.OutputDir)
	if err != nil {
		return fmt.Errorf("invalid output directory path: %w", err)
	}
	// Block exact system paths and their direct children (but allow deeper nesting like /var/folders/...)
	dangerousPaths := []string{"/", "/etc", "/bin", "/usr", "/sys", "/proc", "/var/log", "/var/run"}
	for _, dangerous := range dangerousPaths {
		if absPath == dangerous || strings.HasPrefix(absPath, dangerous+string(filepath.Separator)) {
			return fmt.Errorf("output directory cannot be set to system path: %s", absPath)
		}
	}
	// Also block root-level system directories exactly
	rootDirs := []string{"/etc", "/bin", "/usr", "/sys", "/proc", "/var", "/sbin", "/lib", "/boot"}
	for _, dir := range rootDirs {
		if absPath == dir {
			return fmt.Errorf("output directory cannot be set to system path: %s", absPath)
		}
	}

	// Validate accent color format (hex color)
	hexColorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{3}([0-9A-Fa-f]{3})?$`)
	if !hexColorRegex.MatchString(c.Theme.Accent) {
		return fmt.Errorf("accent color must be a valid hex color (e.g., #50ac00), got %s", c.Theme.Accent)
	}

	// Validate background values (basic check for common patterns)
	if c.Theme.Background.Light != "" {
		if err := validateBackground(c.Theme.Background.Light); err != nil {
			return fmt.Errorf("invalid light background: %w", err)
		}
	}
	if c.Theme.Background.Dark != "" {
		if err := validateBackground(c.Theme.Background.Dark); err != nil {
			return fmt.Errorf("invalid dark background: %w", err)
		}
	}

	// Validate navStyle
	validNavStyles := map[string]bool{"base": true, "sticky": true, "glassy": true}
	if !validNavStyles[c.Theme.NavStyle] {
		return fmt.Errorf("navStyle must be 'base', 'sticky', or 'glassy', got '%s'", c.Theme.NavStyle)
	}

	// Validate navActiveStyle
	validNavActiveStyles := map[string]bool{"base": true, "box": true, "underlined": true}
	if !validNavActiveStyles[c.Theme.NavActiveStyle] {
		return fmt.Errorf("navActiveStyle must be 'base', 'box', or 'underlined', got '%s'", c.Theme.NavActiveStyle)
	}

	// Validate theme font family names. They are interpolated into the
	// inline <style> block and the remote font URL, so both the CLI and the
	// renderer must agree on the safe charset. Empty values are allowed
	// here because Parse fills defaults before validating.
	for _, font := range []string{c.Theme.FontHeading, c.Theme.FontBody, c.Theme.FontMono} {
		if font != "" && !fontFamilyRegex.MatchString(font) {
			return fmt.Errorf("theme font %q may only contain letters, digits, spaces, hyphens, and underscores", font)
		}
	}

	// Validate custom font declarations
	for i, face := range c.Theme.Fonts {
		if err := validateFontFace(face); err != nil {
			return fmt.Errorf("theme.fonts[%d]: %w", i, err)
		}
	}

	// Validate nav paths are well-formed
	for i, nav := range c.Nav {
		if nav.Label == "" {
			return fmt.Errorf("nav item %d has empty label", i)
		}
		if nav.Path == "" {
			return fmt.Errorf("nav item %d (%s) has empty path", i, nav.Label)
		}
		if !strings.HasPrefix(nav.Path, "/") {
			return fmt.Errorf("nav path must start with /, got %s for %s", nav.Path, nav.Label)
		}
	}

	return nil
}
