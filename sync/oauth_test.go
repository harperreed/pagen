package sync

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
)

func TestOAuthConfigCreation(t *testing.T) {
	config := NewOAuthConfig()

	if config == nil {
		t.Fatal("expected config, got nil")
	}

	if len(config.Scopes) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(config.Scopes))
	}

	// Verify required scopes
	requiredScopes := map[string]bool{
		"https://www.googleapis.com/auth/contacts.readonly": false,
		"https://www.googleapis.com/auth/calendar.readonly": false,
		"https://www.googleapis.com/auth/gmail.readonly":    false,
	}

	for _, scope := range config.Scopes {
		if _, ok := requiredScopes[scope]; ok {
			requiredScopes[scope] = true
		}
	}

	for scope, found := range requiredScopes {
		if !found {
			t.Errorf("missing required scope: %s", scope)
		}
	}
}

func TestTokenPathXDG(t *testing.T) {
	path := TokenPath()

	expectedBase := filepath.Join(xdg.DataHome, "pagen")
	if !strings.HasPrefix(path, expectedBase) {
		t.Errorf("expected path under %s, got %s", expectedBase, path)
	}

	if filepath.Base(path) != "google-credentials.json" {
		t.Errorf("expected filename google-credentials.json, got %s", filepath.Base(path))
	}
}
