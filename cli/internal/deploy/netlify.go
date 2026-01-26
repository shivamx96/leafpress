package deploy

import (
	"bytes"
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
	"sync"
	"time"
)

const (
	netlifyAPIBase = "https://api.netlify.com/api/v1"
)

func init() {
	Register(&NetlifyProvider{})
}

// NetlifySite represents a Netlify site
type NetlifySite struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	URL             string `json:"url"`
	PublishedDeploy struct {
		URL string `json:"url"`
	} `json:"published_deploy"`
}

// NetlifyDeploy represents a deployment
type NetlifyDeploy struct {
	ID       string   `json:"id"`
	URL      string   `json:"url"`
	State    string   `json:"state"`
	Required []string `json:"required"`
}

// NetlifyProvider implements deployment to Netlify
type NetlifyProvider struct {
	httpClient *http.Client
}

// Name returns the provider identifier
func (n *NetlifyProvider) Name() string {
	return "netlify"
}

// DisplayName returns human-readable name
func (n *NetlifyProvider) DisplayName() string {
	return "Netlify"
}

// Description returns a short description
func (n *NetlifyProvider) Description() string {
	return "Deploy to Netlify with automatic SSL and edge CDN"
}

// NeedsAuth returns true as Netlify requires authentication
func (n *NetlifyProvider) NeedsAuth() bool {
	return true
}

// Authenticate prompts the user for a Personal Access Token
func (n *NetlifyProvider) Authenticate(ctx context.Context) (*Credentials, error) {
	fmt.Println()
	fmt.Println("  Netlify Personal Access Token")
	fmt.Println("  You can generate a token at: https://app.netlify.com/user/applications and clicking New access token")
	fmt.Println()
	fmt.Print("  Enter your Netlify Personal Access Token: ")

	// Read token from stdin (will be hidden by terminal if called with getpass)
	var token string
	_, err := fmt.Scanln(&token)
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	// Validate the token
	user, err := ValidateNetlifyToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	fmt.Printf("\n  ✓ Authenticated as %s (%s)\n", user.Name, user.Email)

	return &Credentials{
		Provider:    "netlify",
		AccessToken: token,
		Username:    user.Name,
	}, nil
}

// ValidateCredentials checks if the token is still valid
func (n *NetlifyProvider) ValidateCredentials(ctx context.Context, creds *Credentials) error {
	if creds == nil || creds.AccessToken == "" {
		return fmt.Errorf("no credentials provided")
	}

	_, err := ValidateNetlifyToken(ctx, creds.AccessToken)
	return err
}

// Configure is handled by the wizard
func (n *NetlifyProvider) Configure(ctx context.Context, creds *Credentials) (*ProviderConfig, error) {
	return nil, fmt.Errorf("use wizard for configuration")
}

// ListSites fetches the user's Netlify sites
func (n *NetlifyProvider) ListSites(ctx context.Context, token string) ([]NetlifySite, error) {
	if n.httpClient == nil {
		n.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	url := netlifyAPIBase + "/sites"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list sites (status %d): %s", resp.StatusCode, string(body))
	}

	var sites []NetlifySite
	if err := json.NewDecoder(resp.Body).Decode(&sites); err != nil {
		return nil, fmt.Errorf("failed to parse sites response: %w", err)
	}

	return sites, nil
}

// CreateSite creates a new Netlify site
func (n *NetlifyProvider) CreateSite(ctx context.Context, token, siteName string) (*NetlifySite, error) {
	if n.httpClient == nil {
		n.httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	reqBody := map[string]interface{}{
		"name": siteName,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := netlifyAPIBase + "/sites"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create site (status %d): %s", resp.StatusCode, string(body))
	}

	var site NetlifySite
	if err := json.NewDecoder(resp.Body).Decode(&site); err != nil {
		return nil, fmt.Errorf("failed to parse site response: %w", err)
	}

	return &site, nil
}

// Deploy uploads files and creates a deployment
func (n *NetlifyProvider) Deploy(ctx context.Context, cfg *DeployContext) (*DeployResult, error) {
	if n.httpClient == nil {
		n.httpClient = &http.Client{Timeout: 5 * time.Minute}
	}

	siteID := cfg.Config.Settings[SettingSiteID]
	if siteID == "" {
		return nil, fmt.Errorf("site_id not configured")
	}

	if cfg.DryRun {
		return &DeployResult{
			URL:        fmt.Sprintf("https://%s.netlify.app", strings.TrimSpace(siteID)),
			DeployID:   "dry-run",
			DeployedAt: time.Now(),
			Message:    fmt.Sprintf("Would deploy to Netlify site: %s", siteID),
		}, nil
	}

	// Collect all files
	fmt.Println("  Collecting files...")
	files, err := n.collectFiles(cfg.BuildDir)
	if err != nil {
		return nil, fmt.Errorf("failed to collect files: %w", err)
	}

	fmt.Printf("  Found %d files\n", len(files))

	// Build files manifest with SHA1 hashes and keep hash->file mapping
	filesManifest := make(map[string]string)
	hashToFile := make(map[string]fileInfo) // Map hash to file info for upload
	for _, file := range files {
		hash, err := n.calculateSHA1(file.path)
		if err != nil {
			return nil, fmt.Errorf("failed to hash %s: %w", file.relativePath, err)
		}
		filesManifest["/"+file.relativePath] = hash
		hashToFile[hash] = file
	}

	// Create deploy
	fmt.Println("  Creating deployment...")
	deploy, err := n.createDeploy(ctx, cfg.Creds.AccessToken, siteID, filesManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to create deploy: %w", err)
	}

	if len(deploy.Required) > 0 {
		fmt.Printf("  Uploading %d files...\n", len(deploy.Required))

		// Upload required files in parallel
		if err := n.uploadFiles(ctx, cfg.Creds.AccessToken, deploy.ID, hashToFile, deploy.Required); err != nil {
			return nil, fmt.Errorf("failed to upload files: %w", err)
		}
	} else {
		fmt.Println("  No files need uploading (all cached)")
	}

	// Record all deployed files (both uploaded and cached)
	deployedFiles := make(map[string]string)
	for path, hash := range filesManifest {
		deployedFiles[path] = hash
	}

	// Build deployment URL - ensure no double protocols
	deployURL := deploy.URL
	if deployURL == "" {
		deployURL = fmt.Sprintf("https://%s.netlify.app", siteID)
	} else {
		// Remove any existing protocol (http:// or https://)
		if strings.HasPrefix(deployURL, "https://") {
			deployURL = strings.TrimPrefix(deployURL, "https://")
		}
		if strings.HasPrefix(deployURL, "http://") {
			deployURL = strings.TrimPrefix(deployURL, "http://")
		}
		// Always use https
		deployURL = "https://" + deployURL
	}

	result := &DeployResult{
		URL:        deployURL,
		DeployID:   deploy.ID,
		DeployedAt: time.Now(),
		Message:    "Deployed to Netlify",
	}

	// Store deployed files info for manifest
	result.DeployedFiles = deployedFiles

	return result, nil
}

// collectFiles walks the build directory and collects all files
func (n *NetlifyProvider) collectFiles(buildDir string) ([]fileInfo, error) {
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

		// Convert to forward slashes for Netlify
		relPath = filepath.ToSlash(relPath)

		files = append(files, fileInfo{
			path:         path,
			relativePath: relPath,
		})

		return nil
	})

	return files, err
}

// calculateSHA1 calculates the SHA1 hash of a file
func (n *NetlifyProvider) calculateSHA1(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:]), nil
}

// createDeploy creates a new deployment and gets the list of required files
func (n *NetlifyProvider) createDeploy(ctx context.Context, token, siteID string, files map[string]string) (*NetlifyDeploy, error) {
	reqBody := map[string]interface{}{
		"files": files,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/sites/%s/deploys", netlifyAPIBase, siteID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Netlify returns 200 OK instead of 201 Created
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create deploy (status %d): %s", resp.StatusCode, string(body))
	}

	var deploy NetlifyDeploy
	if err := json.NewDecoder(resp.Body).Decode(&deploy); err != nil {
		return nil, fmt.Errorf("failed to parse deploy response: %w", err)
	}

	return &deploy, nil
}

// uploadFiles uploads files in parallel with a worker pool
func (n *NetlifyProvider) uploadFiles(ctx context.Context, token, deployID string, hashToFile map[string]fileInfo, requiredHashes []string) error {
	// Upload with worker pool (max 10 concurrent)
	maxWorkers := 10
	if len(requiredHashes) < maxWorkers {
		maxWorkers = len(requiredHashes)
	}

	errChan := make(chan error, len(requiredHashes))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)

	for _, hash := range requiredHashes {
		// Check for context cancellation before spawning new goroutines
		select {
		case <-ctx.Done():
			break
		default:
		}

		file, ok := hashToFile[hash]
		if !ok {
			return fmt.Errorf("required file hash %s not found in local files", hash)
		}

		wg.Add(1)
		go func(h string, f fileInfo) {
			defer wg.Done()

			// Check context before acquiring semaphore
			select {
			case <-ctx.Done():
				return
			case semaphore <- struct{}{}:
			}
			defer func() { <-semaphore }()

			if err := n.uploadFile(ctx, token, deployID, h, f); err != nil {
				errChan <- fmt.Errorf("failed to upload %s: %w", f.relativePath, err)
			}
		}(hash, file)
	}

	// Wait for all uploads to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// uploadFile uploads a single file to Netlify by hash
func (n *NetlifyProvider) uploadFile(ctx context.Context, token, deployID, hash string, file fileInfo) error {
	data, err := os.ReadFile(file.path)
	if err != nil {
		return err
	}

	// Upload by hash as the file identifier
	url := fmt.Sprintf("%s/deploys/%s/files/%s", netlifyAPIBase, deployID, hash)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
