package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MockProvider is a test provider that simulates deployment
// It copies files to a local directory instead of deploying remotely
type MockProvider struct {
	// Configurable behavior for testing
	ShouldFailAuth   bool
	ShouldFailDeploy bool
	AuthDelay        time.Duration
	DeployDelay      time.Duration

	// Track calls for assertions
	AuthCalls   int
	DeployCalls int
	LastDeploy  *DeployContext
}

// NewMockProvider creates a new mock provider for testing
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) Name() string {
	return "mock"
}

func (m *MockProvider) DisplayName() string {
	return "Mock Provider"
}

func (m *MockProvider) Description() string {
	return "Test provider for development and testing"
}

func (m *MockProvider) NeedsAuth() bool {
	return true
}

func (m *MockProvider) Authenticate(ctx context.Context) (*Credentials, error) {
	m.AuthCalls++

	if m.AuthDelay > 0 {
		select {
		case <-time.After(m.AuthDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.ShouldFailAuth {
		return nil, fmt.Errorf("mock authentication failed")
	}

	return &Credentials{
		Provider:    "mock",
		AccessToken: "mock-token-12345",
		Username:    "mock-user",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *MockProvider) ValidateCredentials(ctx context.Context, creds *Credentials) error {
	if creds == nil || creds.AccessToken == "" {
		return fmt.Errorf("invalid credentials")
	}
	if creds.Provider != "mock" {
		return fmt.Errorf("credentials not for mock provider")
	}
	return nil
}

func (m *MockProvider) Configure(ctx context.Context, creds *Credentials) (*ProviderConfig, error) {
	return &ProviderConfig{
		Provider: "mock",
		Settings: map[string]string{
			SettingRepo:   "mock-user/mock-repo",
			SettingBranch: "gh-pages",
		},
	}, nil
}

func (m *MockProvider) Deploy(ctx context.Context, cfg *DeployContext) (*DeployResult, error) {
	m.DeployCalls++
	m.LastDeploy = cfg

	if m.DeployDelay > 0 {
		select {
		case <-time.After(m.DeployDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.ShouldFailDeploy {
		return nil, fmt.Errorf("mock deployment failed")
	}

	if cfg.DryRun {
		return &DeployResult{
			URL:        "https://mock-user.github.io/mock-repo",
			DeployID:   "dry-run",
			DeployedAt: time.Now(),
			Message:    "Dry run completed successfully",
		}, nil
	}

	// Simulate deployment by copying files to a temp directory
	deployDir := filepath.Join(os.TempDir(), "leafpress-mock-deploy", cfg.Config.Settings[SettingRepo])
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create mock deploy dir: %w", err)
	}

	// Count files for the message
	fileCount := 0
	err := filepath.Walk(cfg.BuildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count files: %w", err)
	}

	return &DeployResult{
		URL:        "https://mock-user.github.io/mock-repo",
		DeployID:   fmt.Sprintf("mock-%d", time.Now().Unix()),
		DeployedAt: time.Now(),
		Message:    fmt.Sprintf("Deployed %d files to mock provider", fileCount),
	}, nil
}

// Reset clears all tracked state for test isolation
func (m *MockProvider) Reset() {
	m.ShouldFailAuth = false
	m.ShouldFailDeploy = false
	m.AuthDelay = 0
	m.DeployDelay = 0
	m.AuthCalls = 0
	m.DeployCalls = 0
	m.LastDeploy = nil
}

func init() {
	// Only register in test/dev environments
	if os.Getenv("LEAFPRESS_ENABLE_MOCK_PROVIDER") == "1" {
		Register(NewMockProvider())
	}
}
