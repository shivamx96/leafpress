package deploy

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	vercelAPIURL      = "https://api.vercel.com"
	vercelUploadURL   = "https://api.vercel.com/v2/files"
	vercelDeployURL   = "https://api.vercel.com/v13/deployments"
	vercelProjectsURL = "https://api.vercel.com/v10/projects"
	vercelUserURL     = "https://api.vercel.com/v2/user"
)

func init() {
	Register(&VercelProvider{})
}

// VercelProvider implements deployment to Vercel
type VercelProvider struct {
	httpClient *http.Client
}

// VercelProject represents a Vercel project
type VercelProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Name returns the provider identifier
func (v *VercelProvider) Name() string {
	return "vercel"
}

// DisplayName returns human-readable name
func (v *VercelProvider) DisplayName() string {
	return "Vercel"
}

// Description returns a short description
func (v *VercelProvider) Description() string {
	return "Deploy to Vercel with automatic SSL and edge network"
}

// NeedsAuth returns true as Vercel requires authentication
func (v *VercelProvider) NeedsAuth() bool {
	return true
}

// Authenticate performs OAuth device flow
func (v *VercelProvider) Authenticate(ctx context.Context) (*Credentials, error) {
	oauth := NewVercelOAuth()
	return oauth.Authenticate(ctx)
}

// ValidateCredentials checks if the token is still valid
func (v *VercelProvider) ValidateCredentials(ctx context.Context, creds *Credentials) error {
	if v.httpClient == nil {
		v.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", vercelUserURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+creds.AccessToken)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid credentials: status %d", resp.StatusCode)
	}

	// Update username from response
	var user struct {
		User struct {
			Username string `json:"username"`
			Name     string `json:"name"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err == nil {
		if user.User.Username != "" {
			creds.Username = user.User.Username
		} else if user.User.Name != "" {
			creds.Username = user.User.Name
		}
	}

	return nil
}

// Configure is handled by the wizard
func (v *VercelProvider) Configure(ctx context.Context, creds *Credentials) (*ProviderConfig, error) {
	return nil, fmt.Errorf("use wizard for configuration")
}

// ListProjects fetches the user's Vercel projects
func (v *VercelProvider) ListProjects(ctx context.Context, token string, teamID string) ([]VercelProject, error) {
	if v.httpClient == nil {
		v.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	url := vercelProjectsURL
	if teamID != "" {
		url += "?teamId=" + teamID
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list projects (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Projects []VercelProject `json:"projects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse projects response: %w", err)
	}

	return result.Projects, nil
}

// Deploy uploads files and creates a deployment
func (v *VercelProvider) Deploy(ctx context.Context, cfg *DeployContext) (*DeployResult, error) {
	if v.httpClient == nil {
		v.httpClient = &http.Client{Timeout: 5 * time.Minute}
	}

	projectName := cfg.Config.Settings["project_name"]
	teamID := cfg.Config.Settings["team_id"]

	if cfg.DryRun {
		return &DeployResult{
			URL:        fmt.Sprintf("https://%s.vercel.app", projectName),
			DeployID:   "dry-run",
			DeployedAt: time.Now(),
			Message:    fmt.Sprintf("Would deploy to Vercel project: %s", projectName),
		}, nil
	}

	// Collect all files to upload
	files, err := v.collectFiles(cfg.BuildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to collect files: %w", err)
	}

	fmt.Printf("  Uploading %d files...\n", len(files))

	// Upload each file and track deployed files
	uploadedFiles := make([]map[string]interface{}, 0, len(files))
	deployedFileMap := make(map[string]string)
	for _, file := range files {
		sha, size, err := v.uploadFile(ctx, cfg.Creds.AccessToken, file.path, teamID)
		if err != nil {
			return nil, fmt.Errorf("failed to upload %s: %w", file.relativePath, err)
		}
		uploadedFiles = append(uploadedFiles, map[string]interface{}{
			"file": file.relativePath,
			"sha":  sha,
			"size": size,
		})
		deployedFileMap["/"+file.relativePath] = sha
	}

	fmt.Println("  Creating deployment...")

	// Create deployment
	deployReq := map[string]interface{}{
		"name":    projectName,
		"files":   uploadedFiles,
		"project": projectName,
		"projectSettings": map[string]interface{}{
			"framework": nil, // Static site, no framework
		},
	}

	deployURL := vercelDeployURL
	if teamID != "" {
		deployURL += "?teamId=" + teamID
	}

	body, err := json.Marshal(deployReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", deployURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Creds.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("deployment failed: %s", string(respBody))
	}

	var deployResp struct {
		ID    string `json:"id"`
		URL   string `json:"url"`
		Ready string `json:"readyState"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		return nil, fmt.Errorf("failed to parse deployment response: %w", err)
	}

	deploymentURL := deployResp.URL
	// Remove any existing protocol and ensure https
	if strings.HasPrefix(deploymentURL, "https://") {
		deploymentURL = strings.TrimPrefix(deploymentURL, "https://")
	}
	if strings.HasPrefix(deploymentURL, "http://") {
		deploymentURL = strings.TrimPrefix(deploymentURL, "http://")
	}
	deploymentURL = "https://" + deploymentURL

	return &DeployResult{
		URL:           deploymentURL,
		DeployID:      deployResp.ID,
		DeployedAt:    time.Now(),
		Message:       "Deployed to Vercel",
		DeployedFiles: deployedFileMap,
	}, nil
}

// fileInfo holds information about a file to upload
type fileInfo struct {
	path         string // Absolute path
	relativePath string // Path relative to build dir (for deployment)
}

// collectFiles walks the build directory and collects all files
func (v *VercelProvider) collectFiles(buildDir string) ([]fileInfo, error) {
	var files []fileInfo

	err := filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(buildDir, path)
		if err != nil {
			return err
		}

		// Convert to forward slashes for Vercel
		relPath = filepath.ToSlash(relPath)

		files = append(files, fileInfo{
			path:         path,
			relativePath: relPath,
		})

		return nil
	})

	return files, err
}

// uploadFile uploads a single file to Vercel and returns its SHA and size
func (v *VercelProvider) uploadFile(ctx context.Context, token, filePath, teamID string) (string, int64, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", 0, err
	}

	// Calculate SHA1
	hash := sha1.Sum(data)
	sha := hex.EncodeToString(hash[:])
	size := int64(len(data))

	// Upload file
	uploadURL := vercelUploadURL
	if teamID != "" {
		uploadURL += "?teamId=" + teamID
	}

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, strings.NewReader(string(data)))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", size))
	req.Header.Set("x-vercel-digest", sha)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	// 200 = uploaded, or file already exists (which is fine)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	return sha, size, nil
}

// BuildVercelURL constructs the Vercel deployment URL for a project
func BuildVercelURL(projectName string) string {
	return fmt.Sprintf("https://%s.vercel.app", projectName)
}
