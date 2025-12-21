package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Config represents the leafpress.json configuration
type Config struct {
	Title       string    `json:"title"`
	BaseURL     string    `json:"baseURL"`
	OutputDir   string    `json:"outputDir"`
	Port        int       `json:"port"`
	Nav         []NavItem `json:"nav"`
	Theme       Theme     `json:"theme"`
	Graph       bool      `json:"graph"`
	GraphOnHome bool      `json:"graphOnHome"`
	TOC         bool      `json:"toc"`
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
	Accent      string `json:"accent"`
	StickyNav   bool   `json:"stickyNav"`
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
			FontHeading: "Crimson Pro",
			FontBody:    "Inter",
			FontMono:    "JetBrains Mono",
			Accent:      "#4a9eff",
			StickyNav:   true,
		},
		Graph:       false,
		GraphOnHome: false,
		TOC:         true,
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
		cfg.Theme.Accent = "#4a9eff"
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
	dangerousPaths := []string{"/", "/etc", "/bin", "/usr", "/var", "/sys", "/proc"}
	for _, dangerous := range dangerousPaths {
		if absPath == dangerous || strings.HasPrefix(absPath, dangerous+string(filepath.Separator)) {
			return fmt.Errorf("output directory cannot be set to system path: %s", absPath)
		}
	}

	// Validate accent color format (hex color)
	hexColorRegex := regexp.MustCompile(`^#[0-9A-Fa-f]{3}([0-9A-Fa-f]{3})?$`)
	if !hexColorRegex.MatchString(c.Theme.Accent) {
		return fmt.Errorf("accent color must be a valid hex color (e.g., #4a9eff), got %s", c.Theme.Accent)
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
