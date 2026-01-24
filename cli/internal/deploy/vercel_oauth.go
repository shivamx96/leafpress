package deploy

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// VercelAuth handles authentication for Vercel using personal access tokens
type VercelAuth struct{}

// NewVercelAuth creates a new Vercel auth handler
func NewVercelAuth() *VercelAuth {
	return &VercelAuth{}
}

// Authenticate prompts the user to enter their Vercel access token
func (v *VercelAuth) Authenticate(ctx context.Context) (*Credentials, error) {
	fmt.Println()
	fmt.Println("  Vercel requires a personal access token for deployment.")
	fmt.Println()
	fmt.Println("  To create a token:")
	fmt.Println("    1. Go to https://vercel.com/account/tokens")
	fmt.Println("    2. Click 'Create' to generate a new token")
	fmt.Println("    3. Give it a name (e.g., 'leafpress')")
	fmt.Println("    4. Copy the token and paste it below")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("  Enter your Vercel access token: ")
	token, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}
	token = strings.TrimSpace(token)

	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	return &Credentials{
		Provider:    "vercel",
		AccessToken: token,
	}, nil
}
