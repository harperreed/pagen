// ABOUTME: OAuth configuration and token management for Google APIs
// ABOUTME: Handles OAuth flow, token storage at XDG paths, and auto-refresh
package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// Google OAuth client credentials
	// Users must create their own OAuth app in Google Cloud Console
	// These are placeholders - real values come from environment or config.
	defaultClientID     = "" // Set via GOOGLE_CLIENT_ID env var
	defaultClientSecret = "" // Set via GOOGLE_CLIENT_SECRET env var
)

// NewOAuthConfig creates OAuth2 config for Google APIs.
func NewOAuthConfig() *oauth2.Config {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		clientID = defaultClientID
	}

	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = defaultClientSecret
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "http://localhost:8080/oauth/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/contacts.readonly",
			"https://www.googleapis.com/auth/calendar.readonly",
			"https://www.googleapis.com/auth/gmail.readonly",
		},
		Endpoint: google.Endpoint,
	}
}

// TokenPath returns XDG-compliant path for storing OAuth tokens.
func TokenPath() string {
	return filepath.Join(xdg.DataHome, "pagen", "google-credentials.json")
}

// SaveToken saves OAuth token to XDG data directory.
func SaveToken(token *oauth2.Token) error {
	path := TokenPath()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Write token file with restricted permissions
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}

	return nil
}

// LoadToken loads OAuth token from XDG data directory.
func LoadToken() (*oauth2.Token, error) {
	path := TokenPath()

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open token file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var token oauth2.Token
	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	return &token, nil
}

// GetClient returns an authenticated HTTP client.
func GetClient(ctx context.Context) (*oauth2.Config, error) {
	config := NewOAuthConfig()

	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("google OAuth credentials not configured. Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables")
	}

	return config, nil
}
