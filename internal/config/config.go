package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the leafpress.json configuration
type Config struct {
	Title     string    `json:"title"`
	BaseURL   string    `json:"baseURL"`
	OutputDir string    `json:"outputDir"`
	Port      int       `json:"port"`
	Nav       []NavItem `json:"nav"`
	Theme     Theme     `json:"theme"`
	Graph     bool      `json:"graph"`
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
		Graph: false,
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
