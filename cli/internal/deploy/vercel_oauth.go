package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// Vercel CLI public client ID - safe to embed
	// This is the same client ID used by the official Vercel CLI
	// Users will see the authorization request from "Vercel CLI"
	vercelClientID = "cl_HYyOPBNtFMfHhaUn9L4QPfTZz6TP47bp"

	// Vercel OAuth endpoints (from OpenID Connect discovery)
	vercelDeviceAuthURL = "https://api.vercel.com/login/oauth/device-authorization"
	vercelTokenURL      = "https://api.vercel.com/login/oauth/token"

	// Vercel REST API for user info
	vercelUserAPIURL = "https://api.vercel.com/v2/user"
)

// VercelDeviceCodeResponse from Vercel's device flow
type VercelDeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// VercelTokenResponse from Vercel's token endpoint
type VercelTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// VercelUserResponse represents the /v2/user API response
type VercelUserResponse struct {
	User struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Email    string `json:"email"`
	} `json:"user"`
}

// VercelOAuth handles the Vercel device OAuth flow
type VercelOAuth struct {
	clientID   string
	httpClient *http.Client
}

// NewVercelOAuth creates a new Vercel OAuth handler
func NewVercelOAuth() *VercelOAuth {
	return &VercelOAuth{
		clientID: vercelClientID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authenticate performs the device OAuth flow
func (v *VercelOAuth) Authenticate(ctx context.Context) (*Credentials, error) {
	// Step 1: Request device code
	deviceCode, err := v.requestDeviceCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}

	// Step 2: Copy code to clipboard and show user
	fmt.Println()

	if err := copyToClipboard(deviceCode.UserCode); err == nil {
		fmt.Printf("  Code copied to clipboard: %s\n", deviceCode.UserCode)
	} else {
		fmt.Printf("  Your code: %s\n", deviceCode.UserCode)
	}

	fmt.Println()
	fmt.Println("  Opening browser to authorize leafpress...")
	fmt.Printf("  If browser doesn't open, visit: %s\n", deviceCode.VerificationURI)
	fmt.Println()
	fmt.Println("  Waiting for authorization...")

	// Try to open browser
	if err := openBrowser(deviceCode.VerificationURI); err != nil {
		// Not fatal, user can manually visit URL
	}

	// Step 3: Poll for token
	token, err := v.pollForToken(ctx, deviceCode)
	if err != nil {
		return nil, err
	}

	// Step 4: Get user info
	userResp, err := v.getUser(ctx, token.AccessToken)
	if err != nil {
		// Not fatal - we have a valid token
		return &Credentials{
			Provider:    "vercel",
			AccessToken: token.AccessToken,
		}, nil
	}

	username := userResp.User.Username
	if username == "" {
		username = userResp.User.Name
	}
	if username == "" {
		username = userResp.User.Email
	}

	return &Credentials{
		Provider:    "vercel",
		AccessToken: token.AccessToken,
		Username:    username,
	}, nil
}

// requestDeviceCode initiates the device flow
func (v *VercelOAuth) requestDeviceCode(ctx context.Context) (*VercelDeviceCodeResponse, error) {
	data := url.Values{
		"client_id": {v.clientID},
		"scope":     {"openid offline_access"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", vercelDeviceAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := v.httpClient.Do(req)
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

	var deviceCode VercelDeviceCodeResponse
	if err := json.Unmarshal(body, &deviceCode); err != nil {
		return nil, err
	}

	return &deviceCode, nil
}

// pollForToken polls Vercel until user authorizes or timeout
func (v *VercelOAuth) pollForToken(ctx context.Context, deviceCode *VercelDeviceCodeResponse) (*VercelTokenResponse, error) {
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

			token, err := v.checkToken(ctx, deviceCode.DeviceCode)
			if err != nil {
				return nil, err
			}

			if token.Error == "" && token.AccessToken != "" {
				return token, nil
			}

			switch token.Error {
			case "authorization_pending":
				// Keep polling
				continue
			case "slow_down":
				// Increase interval per RFC 8628
				interval += 5 * time.Second
				ticker.Reset(interval)
				continue
			case "expired_token":
				return nil, fmt.Errorf("authorization expired - please try again")
			case "access_denied":
				return nil, fmt.Errorf("authorization denied by user")
			default:
				if token.Error != "" {
					return nil, fmt.Errorf("authorization failed: %s - %s", token.Error, token.ErrorDesc)
				}
				// Keep polling if no error but also no token yet
				continue
			}
		}
	}
}

// checkToken attempts to exchange device code for access token
func (v *VercelOAuth) checkToken(ctx context.Context, deviceCode string) (*VercelTokenResponse, error) {
	data := url.Values{
		"client_id":   {v.clientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", vercelTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token VercelTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// getUser fetches the authenticated user's info from /v2/user
func (v *VercelOAuth) getUser(ctx context.Context, token string) (*VercelUserResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", vercelUserAPIURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userResp VercelUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, err
	}

	return &userResp, nil
}
