// ABOUTME: Tests for database schema creation and migrations
// ABOUTME: Uses in-memory SQLite for fast isolated tests
package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory db: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	// Verify tables exist
	tables := []string{"contacts", "companies", "deals", "deal_notes"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s not found: %v", table, err)
		}
	}
}

func TestSchemaIncludesFollowupTables(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	// Check contact_cadence table exists
	row := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='contact_cadence'`)
	var tableName string
	err := row.Scan(&tableName)
	if err != nil {
		t.Fatalf("contact_cadence table not found: %v", err)
	}

	// Check interaction_log table exists
	row = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='interaction_log'`)
	err = row.Scan(&tableName)
	if err != nil {
		t.Fatalf("interaction_log table not found: %v", err)
	}

	// Check indexes exist
	row = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='index' AND name='idx_contact_cadence_priority'`)
	err = row.Scan(&tableName)
	if err != nil {
		t.Fatalf("priority index not found: %v", err)
	}
}
