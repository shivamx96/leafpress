package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const ManifestFile = ".leafpress-deploy-state.json"

// DeploymentRecord represents a single deployment
type DeploymentRecord struct {
	Timestamp     time.Time         `json:"timestamp"`
	Provider      string            `json:"provider"`
	DeployID      string            `json:"deployID"`
	URL           string            `json:"url"`
	FileCount     int               `json:"fileCount"`
	FilesDeployed map[string]string `json:"filesDeployed"` // path -> SHA1 hash (deployed files for reference)
	SourceFiles   map[string]string `json:"sourceFiles"`   // path -> SHA1 hash (source files for comparison)
}

// DeploymentManifest tracks deployment history and current state
type DeploymentManifest struct {
	LastDeploy    *DeploymentRecord  `json:"lastDeploy"`
	DeployHistory []DeploymentRecord `json:"deployHistory"`
}

// NewDeploymentManifest creates an empty manifest
func NewDeploymentManifest() *DeploymentManifest {
	return &DeploymentManifest{
		LastDeploy:    nil,
		DeployHistory: make([]DeploymentRecord, 0),
	}
}

// LoadDeploymentManifest loads the manifest from disk
func LoadDeploymentManifest(projectRoot string) (*DeploymentManifest, error) {
	path := filepath.Join(projectRoot, ManifestFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No manifest yet, return empty
			return NewDeploymentManifest(), nil
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest DeploymentManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// Save writes the manifest to disk
func (m *DeploymentManifest) Save(projectRoot string) error {
	path := filepath.Join(projectRoot, ManifestFile)

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// RecordDeployment records a new deployment
func (m *DeploymentManifest) RecordDeployment(result *DeployResult, provider string, filesDeployed map[string]string, sourceFiles map[string]string) {
	record := DeploymentRecord{
		Timestamp:     result.DeployedAt,
		Provider:      provider,
		DeployID:      result.DeployID,
		URL:           result.URL,
		FileCount:     len(filesDeployed),
		FilesDeployed: filesDeployed,
		SourceFiles:   sourceFiles,
	}

	// Add to history (keep last 10 deployments)
	m.DeployHistory = append(m.DeployHistory, record)
	if len(m.DeployHistory) > 10 {
		m.DeployHistory = m.DeployHistory[len(m.DeployHistory)-10:]
	}

	// Update last deploy
	m.LastDeploy = &record
}

// GetPendingFiles compares current source files against last deployment
// Returns a list of changed files (path -> current hash)
func (m *DeploymentManifest) GetPendingFiles(currentFiles map[string]string) map[string]string {
	if m.LastDeploy == nil {
		// No previous deployment, all files are pending
		return currentFiles
	}

	pending := make(map[string]string)

	// Use SourceFiles for comparison if available, fallback to FilesDeployed for backwards compatibility
	previousFiles := m.LastDeploy.SourceFiles
	if previousFiles == nil {
		previousFiles = m.LastDeploy.FilesDeployed
	}

	// Check for new or modified files
	for path, hash := range currentFiles {
		if oldHash, exists := previousFiles[path]; !exists || oldHash != hash {
			pending[path] = hash
		}
	}

	// Check for deleted files (in previous but not in current)
	for path := range previousFiles {
		if _, exists := currentFiles[path]; !exists {
			pending[path] = "deleted"
		}
	}

	return pending
}

// LastDeploymentTime returns when the last deployment occurred
func (m *DeploymentManifest) LastDeploymentTime() *time.Time {
	if m.LastDeploy == nil {
		return nil
	}
	return &m.LastDeploy.Timestamp
}

// TimeSinceLastDeploy returns a human-readable time string
func (m *DeploymentManifest) TimeSinceLastDeploy() string {
	if m.LastDeploy == nil {
		return "never"
	}

	since := time.Since(m.LastDeploy.Timestamp)

	if since < time.Minute {
		return "just now"
	}
	if since < time.Hour {
		minutes := int(since.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	if since < 24*time.Hour {
		hours := int(since.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	days := int(since.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
