package sync

import (
	"testing"

	"github.com/google/uuid"
	"github.com/harperreed/pagen/models"
)

func TestMatchContactByEmail(t *testing.T) {
	existing := []models.Contact{
		{ID: uuid.New(), Name: "Alice", Email: "alice@example.com"},
		{ID: uuid.New(), Name: "Bob", Email: "bob@example.com"},
	}

	matcher := NewContactMatcher(existing)

	// Test exact match
	match, found := matcher.FindMatch("alice@example.com", "")
	if !found {
		t.Error("expected to find match for alice@example.com")
	}
	if match.Email != "alice@example.com" {
		t.Errorf("expected alice@example.com, got %s", match.Email)
	}

	// Test no match
	_, found = matcher.FindMatch("charlie@example.com", "")
	if found {
		t.Error("expected no match for charlie@example.com")
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Alice@Example.com", "alice@example.com"},
		{"alice.smith@example.com", "alice.smith@example.com"},
		{"ALICE@EXAMPLE.COM", "alice@example.com"},
	}

	for _, tt := range tests {
		result := normalizeEmail(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"alice@example.com", "example.com"},
		{"bob@acme.co.uk", "acme.co.uk"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		result := extractDomain(tt.email)
		if result != tt.expected {
			t.Errorf("extractDomain(%q) = %q, want %q", tt.email, result, tt.expected)
		}
	}
}
