// Package deploy provides deployment functionality for leafpress sites
package deploy

import (
	"context"
	"sort"
	"time"
)

// Provider defines the interface for deployment providers
type Provider interface {
	// Name returns the provider identifier (e.g., "github-pages")
	Name() string

	// DisplayName returns human-readable name (e.g., "GitHub Pages")
	DisplayName() string

	// Description returns a short description for the selection menu
	Description() string

	// NeedsAuth returns true if authentication is required
	NeedsAuth() bool

	// Authenticate performs the OAuth/auth flow and returns credentials
	Authenticate(ctx context.Context) (*Credentials, error)

	// ValidateCredentials checks if stored credentials are still valid
	ValidateCredentials(ctx context.Context, creds *Credentials) error

	// Configure runs the interactive setup wizard for this provider
	Configure(ctx context.Context, creds *Credentials) (*ProviderConfig, error)

	// Deploy pushes the built site to the hosting provider
	Deploy(ctx context.Context, cfg *DeployContext) (*DeployResult, error)
}

// Credentials holds authentication tokens for a provider
type Credentials struct {
	Provider    string    `toml:"provider"`
	AccessToken string    `toml:"access_token"`
	Username    string    `toml:"username,omitempty"`
	ExpiresAt   time.Time `toml:"expires_at,omitempty"`
}

// ProviderConfig holds provider-specific deployment configuration
type ProviderConfig struct {
	Provider string            `json:"provider"`
	Settings map[string]string `json:"settings"`
}

// DeployContext contains everything needed for a deployment
type DeployContext struct {
	BuildDir string          // Path to _site directory
	Config   *ProviderConfig // Provider-specific config
	Creds    *Credentials    // Authentication credentials
	DryRun   bool            // If true, validate but don't deploy
}

// DeployResult contains information about a completed deployment
type DeployResult struct {
	URL        string    // Live URL of the deployed site
	DeployID   string    // Provider-specific deploy identifier
	DeployedAt time.Time // Timestamp of deployment
	Message    string    // Optional status message
}

// Common provider setting keys
const (
	SettingRepo   = "repo"    // Repository (e.g., "user/repo")
	SettingBranch = "branch"  // Deploy branch (e.g., "gh-pages")
	SettingSiteID = "site_id" // Provider site ID (Netlify/Vercel)
)

// Registry holds all available providers
var registry = make(map[string]Provider)

// Register adds a provider to the registry
func Register(p Provider) {
	registry[p.Name()] = p
}

// Get returns a provider by name
func Get(name string) (Provider, bool) {
	p, ok := registry[name]
	return p, ok
}

// List returns all registered providers sorted by name for consistent ordering
func List() []Provider {
	providers := make([]Provider, 0, len(registry))
	for _, p := range registry {
		providers = append(providers, p)
	}
	// Sort by name for consistent menu ordering
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Name() < providers[j].Name()
	})
	return providers
}
