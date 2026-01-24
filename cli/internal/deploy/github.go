package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	githubReposURL = "https://api.github.com/user/repos"
)

// GitHubRepo represents a GitHub repository
type GitHubRepo struct {
	FullName      string `json:"full_name"`
	Name          string `json:"name"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"default_branch"`
	HTMLURL       string `json:"html_url"`
}

// GitHubPagesProvider deploys to GitHub Pages using git
type GitHubPagesProvider struct {
	oauth      *GitHubOAuth
	httpClient *http.Client
}

// NewGitHubPagesProvider creates a new GitHub Pages provider
func NewGitHubPagesProvider() *GitHubPagesProvider {
	return &GitHubPagesProvider{
		oauth: NewGitHubOAuth(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (g *GitHubPagesProvider) Name() string {
	return "github-pages"
}

func (g *GitHubPagesProvider) DisplayName() string {
	return "GitHub Pages"
}

func (g *GitHubPagesProvider) Description() string {
	return "Free for public repos, private repos require GitHub Pro/Team"
}

func (g *GitHubPagesProvider) NeedsAuth() bool {
	return true
}

func (g *GitHubPagesProvider) Authenticate(ctx context.Context) (*Credentials, error) {
	return g.oauth.Authenticate(ctx, func(userCode, verificationURL string) {
		fmt.Println()
		fmt.Printf("  Opening browser to authorize leafpress...\n")
		fmt.Printf("  If browser doesn't open, visit: %s\n", verificationURL)
		fmt.Printf("  And enter code: %s\n", userCode)
		fmt.Println()
		fmt.Println("  Waiting for authorization...")
	})
}

func (g *GitHubPagesProvider) ValidateCredentials(ctx context.Context, creds *Credentials) error {
	if creds == nil || creds.AccessToken == "" {
		return fmt.Errorf("no credentials provided")
	}
	return g.oauth.ValidateToken(ctx, creds.AccessToken)
}

func (g *GitHubPagesProvider) Configure(ctx context.Context, creds *Credentials) (*ProviderConfig, error) {
	// This will be called from the wizard - for now return nil
	// The wizard handles the interactive configuration
	return nil, fmt.Errorf("use wizard for configuration")
}

// ListRepos fetches the user's repositories
func (g *GitHubPagesProvider) ListRepos(ctx context.Context, token string) ([]GitHubRepo, error) {
	var allRepos []GitHubRepo
	page := 1

	for {
		url := fmt.Sprintf("%s?per_page=100&page=%d&sort=updated", githubReposURL, page)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := g.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("failed to list repos: %s", string(body))
		}

		var repos []GitHubRepo
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			return nil, err
		}

		if len(repos) == 0 {
			break
		}

		allRepos = append(allRepos, repos...)
		page++

		// Safety limit
		if page > 10 {
			break
		}
	}

	return allRepos, nil
}

func (g *GitHubPagesProvider) Deploy(ctx context.Context, cfg *DeployContext) (*DeployResult, error) {
	repo := cfg.Config.Settings[SettingRepo]
	branch := cfg.Config.Settings[SettingBranch]
	if branch == "" {
		branch = "gh-pages"
	}

	if cfg.DryRun {
		return &DeployResult{
			URL:        g.buildPagesURL(repo),
			DeployID:   "dry-run",
			DeployedAt: time.Now(),
			Message:    fmt.Sprintf("Would deploy to %s branch %s", repo, branch),
		}, nil
	}

	// Create temp directory for git operations
	tmpDir, err := os.MkdirTemp("", "leafpress-deploy-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Deploy using git
	if err := g.gitDeploy(ctx, cfg.BuildDir, tmpDir, repo, branch, cfg.Creds.AccessToken); err != nil {
		return nil, err
	}

	return &DeployResult{
		URL:        g.buildPagesURL(repo),
		DeployID:   fmt.Sprintf("gh-%d", time.Now().Unix()),
		DeployedAt: time.Now(),
		Message:    fmt.Sprintf("Deployed to %s", branch),
	}, nil
}

// gitDeploy performs the actual git-based deployment
func (g *GitHubPagesProvider) gitDeploy(ctx context.Context, buildDir, tmpDir, repo, branch, token string) error {
	repoURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", token, repo)

	// Try to clone the existing gh-pages branch
	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", branch, repoURL, tmpDir)
	cloneErr := cloneCmd.Run()

	if cloneErr != nil {
		// Branch doesn't exist, initialize new repo
		if err := g.initNewBranch(ctx, tmpDir, repoURL, branch); err != nil {
			return err
		}
	}

	// Remove all existing files (except .git)
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}
	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(tmpDir, entry.Name())); err != nil {
			return fmt.Errorf("failed to clean temp directory: %w", err)
		}
	}

	// Copy build files to temp directory
	if err := copyDir(buildDir, tmpDir); err != nil {
		return fmt.Errorf("failed to copy build files: %w", err)
	}

	// Add .nojekyll to disable Jekyll processing
	nojekyllPath := filepath.Join(tmpDir, ".nojekyll")
	if err := os.WriteFile(nojekyllPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to create .nojekyll: %w", err)
	}

	// Git add all files
	addCmd := exec.CommandContext(ctx, "git", "add", "-A")
	addCmd.Dir = tmpDir
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Git commit
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", fmt.Sprintf("Deploy via leafpress at %s", time.Now().Format(time.RFC3339)))
	commitCmd.Dir = tmpDir
	// Commit might fail if no changes - that's OK
	commitCmd.Run()

	// Git push
	pushCmd := exec.CommandContext(ctx, "git", "push", "origin", branch)
	pushCmd.Dir = tmpDir
	if output, err := pushCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %s", string(output))
	}

	return nil
}

// initNewBranch creates an orphan branch for first deployment
func (g *GitHubPagesProvider) initNewBranch(ctx context.Context, tmpDir, repoURL, branch string) error {
	// Initialize empty repo
	initCmd := exec.CommandContext(ctx, "git", "init")
	initCmd.Dir = tmpDir
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("git init failed: %w", err)
	}

	// Add remote
	remoteCmd := exec.CommandContext(ctx, "git", "remote", "add", "origin", repoURL)
	remoteCmd.Dir = tmpDir
	if err := remoteCmd.Run(); err != nil {
		return fmt.Errorf("git remote add failed: %w", err)
	}

	// Create orphan branch
	checkoutCmd := exec.CommandContext(ctx, "git", "checkout", "--orphan", branch)
	checkoutCmd.Dir = tmpDir
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("git checkout --orphan failed: %w", err)
	}

	return nil
}

// buildPagesURL constructs the GitHub Pages URL for a repo
func (g *GitHubPagesProvider) buildPagesURL(repo string) string {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return ""
	}
	username := parts[0]
	repoName := parts[1]

	// Check if it's a user/org pages repo
	if repoName == username+".github.io" {
		return fmt.Sprintf("https://%s.github.io", username)
	}

	return fmt.Sprintf("https://%s.github.io/%s", username, repoName)
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func init() {
	Register(NewGitHubPagesProvider())
}
