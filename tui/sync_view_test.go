// ABOUTME: Tests for sync view functionality
// ABOUTME: Verifies sync state display and command handling
package tui

import (
	"database/sql"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"

	"github.com/harperreed/pagen/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}
	if err := db.InitSchema(database); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}
	return database
}

func TestSyncViewRendering(t *testing.T) {
	// Create test database
	database := setupTestDB(t)
	defer func() { _ = database.Close() }()

	// Create model
	m := NewModel(database)
	m.entityType = EntitySync

	// Render the view
	output := m.renderSyncView()

	// Verify basic structure
	if output == "" {
		t.Fatal("Sync view should not be empty")
	}

	// Should contain title
	if !contains(output, "Google Sync Management") && !contains(output, "Sync Management") {
		t.Error("Sync view should contain title")
	}
}

func TestSyncViewWithStates(t *testing.T) {
	// Create test database
	database := setupTestDB(t)
	defer func() { _ = database.Close() }()

	// Add some sync states
	_ = db.UpdateSyncStatus(database, "contacts", "idle", nil)
	_ = db.UpdateSyncToken(database, "contacts", "test-token")

	errMsg := "test error"
	_ = db.UpdateSyncStatus(database, "gmail", "error", &errMsg)

	// Create model
	m := NewModel(database)
	m.entityType = EntitySync

	// Load states
	m.loadSyncStates()

	// Verify states were loaded
	if len(m.syncStates) == 0 {
		t.Error("Should have loaded sync states")
	}

	// Render the view
	output := m.renderSyncView()

	// Should show service names
	if !contains(output, "Contacts") && !contains(output, "contacts") {
		t.Error("Should show contacts service")
	}
}

func TestSyncKeyNavigation(t *testing.T) {
	// Create test database
	database := setupTestDB(t)
	defer func() { _ = database.Close() }()

	m := NewModel(database)
	m.entityType = EntitySync

	// Test up/down navigation
	m.selectedService = 1
	updated, _ := m.handleSyncKeys(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.selectedService != 0 {
		t.Errorf("Expected selectedService=0, got %d", m.selectedService)
	}

	updated, _ = m.handleSyncKeys(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selectedService != 1 {
		t.Errorf("Expected selectedService=1, got %d", m.selectedService)
	}

	// Test escape key
	m.viewMode = ViewList
	m.entityType = EntitySync
	updated, _ = m.handleSyncKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.viewMode != ViewList {
		t.Error("Escape should keep view mode as ViewList")
	}
	if m.entityType == EntitySync {
		t.Error("Escape should change entity type away from EntitySync")
	}
}

func TestSyncCompleteMessage(t *testing.T) {
	// Create test database
	database := setupTestDB(t)
	defer func() { _ = database.Close() }()

	m := NewModel(database)

	// Mark a sync as in progress
	m.syncInProgress["contacts"] = true

	// Handle completion
	msg := SyncCompleteMsg{
		Service: "contacts",
		Error:   nil,
	}

	_ = m.handleSyncComplete(msg)

	// Should no longer be in progress
	if m.syncInProgress["contacts"] {
		t.Error("Sync should not be in progress after completion")
	}

	// Should have a message
	if len(m.syncMessages) == 0 {
		t.Error("Should have added a completion message")
	}
}

func TestSyncCompleteWithError(t *testing.T) {
	// Create test database
	database := setupTestDB(t)
	defer func() { _ = database.Close() }()

	m := NewModel(database)

	// Mark a sync as in progress
	m.syncInProgress["gmail"] = true

	// Handle completion with error
	msg := SyncCompleteMsg{
		Service: "gmail",
		Error:   &testError{msg: "test sync error"},
	}

	_ = m.handleSyncComplete(msg)

	// Should no longer be in progress
	if m.syncInProgress["gmail"] {
		t.Error("Sync should not be in progress after error")
	}

	// Should have a message
	if len(m.syncMessages) == 0 {
		t.Error("Should have added an error message")
	}

	// Check that error was recorded in database
	state, _ := db.GetSyncState(database, "gmail")
	if state == nil {
		t.Fatal("Should have sync state")
	}
	if state.Status != "error" {
		t.Errorf("Expected status=error, got %s", state.Status)
	}
	if state.ErrorMessage == nil || *state.ErrorMessage != "test sync error" {
		t.Error("Should have recorded error message")
	}
}

func TestFormatTimeSince(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "just now",
			time:     time.Now().Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "minutes ago",
			time:     time.Now().Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "hours ago",
			time:     time.Now().Add(-2 * time.Hour),
			expected: "2 hours ago",
		},
		{
			name:     "days ago",
			time:     time.Now().Add(-3 * 24 * time.Hour),
			expected: "3 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeSince(tt.time)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSyncMessageAddition(t *testing.T) {
	m := NewModel(nil)

	m.addSyncMessage("Test message 1")
	m.addSyncMessage("Test message 2")

	if len(m.syncMessages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(m.syncMessages))
	}

	// Check that messages contain timestamps
	if !contains(m.syncMessages[0], "Test message 1") {
		t.Error("First message should contain content")
	}
	if !contains(m.syncMessages[1], "Test message 2") {
		t.Error("Second message should contain content")
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
