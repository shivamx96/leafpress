package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// CredentialsStore manages persistent storage of provider credentials
type CredentialsStore struct {
	path string
	data map[string]*Credentials
}

// NewCredentialsStore creates a store using the default config path
func NewCredentialsStore() (*CredentialsStore, error) {
	path, err := defaultCredentialsPath()
	if err != nil {
		return nil, err
	}
	return NewCredentialsStoreAt(path)
}

// NewCredentialsStoreAt creates a store at a specific path
func NewCredentialsStoreAt(path string) (*CredentialsStore, error) {
	store := &CredentialsStore{
		path: path,
		data: make(map[string]*Credentials),
	}

	// Load existing credentials if file exists
	if _, err := os.Stat(path); err == nil {
		if err := store.load(); err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w", err)
		}
	}

	return store, nil
}

// defaultCredentialsPath returns the platform-appropriate config path
func defaultCredentialsPath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config", "leafpress")
	case "linux":
		// Follow XDG Base Directory spec
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			configDir = filepath.Join(xdg, "leafpress")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			configDir = filepath.Join(home, ".config", "leafpress")
		}
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "leafpress")
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config", "leafpress")
	}

	return filepath.Join(configDir, "credentials.json"), nil
}

// Get retrieves credentials for a provider
func (s *CredentialsStore) Get(provider string) (*Credentials, bool) {
	creds, ok := s.data[provider]
	return creds, ok
}

// Set stores credentials for a provider
func (s *CredentialsStore) Set(creds *Credentials) error {
	s.data[creds.Provider] = creds
	return s.save()
}

// Delete removes credentials for a provider
func (s *CredentialsStore) Delete(provider string) error {
	delete(s.data, provider)
	return s.save()
}

// load reads credentials from disk
func (s *CredentialsStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.data)
}

// save writes credentials to disk
func (s *CredentialsStore) save() error {
	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	// Write with restricted permissions (owner read/write only)
	return os.WriteFile(s.path, data, 0600)
}

// Path returns the credentials file path
func (s *CredentialsStore) Path() string {
	return s.path
}

// GetFromEnv attempts to load credentials from environment variables
// Environment variables take precedence over stored credentials
func GetFromEnv(provider string) *Credentials {
	var tokenEnv string
	switch provider {
	case "github-pages":
		tokenEnv = "LEAFPRESS_GITHUB_TOKEN"
	case "netlify":
		tokenEnv = "LEAFPRESS_NETLIFY_TOKEN"
	case "vercel":
		tokenEnv = "LEAFPRESS_VERCEL_TOKEN"
	default:
		return nil
	}

	token := os.Getenv(tokenEnv)
	if token == "" {
		return nil
	}

	return &Credentials{
		Provider:    provider,
		AccessToken: token,
	}
}
