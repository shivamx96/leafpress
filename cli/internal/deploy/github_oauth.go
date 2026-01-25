package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	// GitHub OAuth App credentials for leafpress
	// This is a public client ID - it's safe to embed
	// Users will see "leafpress" when authorizing
	githubClientID = "Ov23liBzTwDReEfk8lbC" // TODO: Replace with real client ID

	githubDeviceCodeURL = "https://github.com/login/device/code"
	githubTokenURL      = "https://github.com/login/oauth/access_token"
	githubUserURL       = "https://api.github.com/user"

	// Scopes needed for GitHub Pages deployment
	githubScopes = "repo"
)

// DeviceCodeResponse from GitHub's device flow
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// TokenResponse from GitHub's token endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// GitHubUser represents the authenticated user
type GitHubUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

// GitHubOAuth handles the GitHub device OAuth flow
type GitHubOAuth struct {
	clientID   string
	httpClient *http.Client
}

// NewGitHubOAuth creates a new GitHub OAuth handler
func NewGitHubOAuth() *GitHubOAuth {
	return &GitHubOAuth{
		clientID: githubClientID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate performs the device OAuth flow
// Returns credentials on success, or error if user doesn't complete auth
func (g *GitHubOAuth) Authenticate(ctx context.Context, onCode func(userCode, verificationURL string)) (*Credentials, error) {
	// Step 1: Request device code
	deviceCode, err := g.requestDeviceCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}

	// Step 2: Show user the code and open browser
	onCode(deviceCode.UserCode, deviceCode.VerificationURI)
	openBrowser(deviceCode.VerificationURI)

	// Step 3: Poll for token
	token, err := g.pollForToken(ctx, deviceCode)
	if err != nil {
		return nil, err
	}

	// Step 4: Get username
	user, err := g.getUser(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &Credentials{
		Provider:    "github-pages",
		AccessToken: token.AccessToken,
		Username:    user.Login,
	}, nil
}

// requestDeviceCode initiates the device flow
func (g *GitHubOAuth) requestDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	data := url.Values{
		"client_id": {g.clientID},
		"scope":     {githubScopes},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", githubDeviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed: %s", string(body))
	}

	var deviceCode DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceCode); err != nil {
		return nil, err
	}

	return &deviceCode, nil
}

// pollForToken polls GitHub until user authorizes or timeout
func (g *GitHubOAuth) pollForToken(ctx context.Context, deviceCode *DeviceCodeResponse) (*TokenResponse, error) {
	interval := time.Duration(deviceCode.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}

	deadline := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("authorization timed out - please try again")
			}

			token, err := g.checkToken(ctx, deviceCode.DeviceCode)
			if err != nil {
				return nil, err
			}

			if token.Error == "" {
				return token, nil
			}

			switch token.Error {
			case "authorization_pending":
				// Keep polling
				continue
			case "slow_down":
				// Increase interval
				interval += 5 * time.Second
				ticker.Reset(interval)
				continue
			case "expired_token":
				return nil, fmt.Errorf("authorization expired - please try again")
			case "access_denied":
				return nil, fmt.Errorf("authorization denied by user")
			default:
				return nil, fmt.Errorf("authorization failed: %s", token.ErrorDesc)
			}
		}
	}
}

// checkToken attempts to exchange device code for access token
func (g *GitHubOAuth) checkToken(ctx context.Context, deviceCode string) (*TokenResponse, error) {
	data := url.Values{
		"client_id":   {g.clientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", githubTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// getUser fetches the authenticated user's info
func (g *GitHubOAuth) getUser(ctx context.Context, token string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", githubUserURL, nil)
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
		return nil, fmt.Errorf("failed to get user: %s", string(body))
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// ValidateToken checks if a token is still valid
func (g *GitHubOAuth) ValidateToken(ctx context.Context, token string) error {
	_, err := g.getUser(ctx, token)
	return err
}

// openBrowser attempts to open a URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

// copyToClipboard copies text to the system clipboard
func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, fall back to xsel
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "windows":
		cmd = exec.Command("cmd", "/c", "clip")
	default:
		return fmt.Errorf("unsupported platform")
	}

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := pipe.Write([]byte(text)); err != nil {
		return err
	}

	if err := pipe.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}
