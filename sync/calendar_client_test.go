package sync

import (
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestNewCalendarClient(t *testing.T) {
	// Create a mock token
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Test creating calendar client
	service, err := NewCalendarClient(token)
	if err != nil {
		t.Fatalf("NewCalendarClient failed: %v", err)
	}

	if service == nil {
		t.Fatal("expected service, got nil")
	}
}

func TestNewCalendarClientNilToken(t *testing.T) {
	// Test that nil token returns error
	service, err := NewCalendarClient(nil)
	if err == nil {
		t.Fatal("expected error for nil token, got nil")
	}

	if service != nil {
		t.Error("expected nil service for nil token")
	}
}
