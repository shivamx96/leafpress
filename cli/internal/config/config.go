package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Background represents background configuration that can be a string or object
type Background struct {
	Light string
	Dark  string
}

// Config represents the leafpress.json configuration
type Config struct {
	Title       string    `json:"title"`
	Description string    `json:"description"` // Site-wide meta description
	Author      string    `json:"author"`
	BaseURL     string    `json:"baseURL"`
	Image       string    `json:"image"` // Default OG image path (e.g., "/og-image.png")
	OutputDir   string    `json:"outputDir"`
	Port        int       `json:"port"`
	Nav         []NavItem `json:"nav"`
	Theme       Theme     `json:"theme"`
	Graph       bool      `json:"graph"`
	Search      bool      `json:"search"`
	TOC         bool      `json:"toc"`
	Backlinks   bool      `json:"backlinks"`
	Wikilinks   bool      `json:"wikilinks"`
	Ignore      []string  `json:"ignore"`
}

// NavItem represents a navigation link
type NavItem struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

// Theme represents theme configuration
type Theme struct {
	FontHeading    string     `json:"fontHeading"`
	FontBody       string     `json:"fontBody"`
	FontMono       string     `json:"fontMono"`
	Accent         string     `json:"accent"`
	Background     Background `json:"-"`              // Custom unmarshaling
	NavStyle       string     `json:"navStyle"`       // "base", "sticky", or "glassy"
	NavActiveStyle string     `json:"navActiveStyle"` // "base", "box", or "underlined"
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
			NavStyle:       "glassy",
			NavActiveStyle: "base",
		},
		Graph:     false,
		TOC:       true,
		Backlinks: true,
		Wikilinks: true,
	}
}

// Load reads and parses the config file
func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults if no config file
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

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
		cfg.Theme.NavStyle = "glassy"
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
