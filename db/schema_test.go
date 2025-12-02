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

	// Verify objects table exists
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='objects'").Scan(&name)
	if err != nil {
		t.Errorf("Table objects not found: %v", err)
	}

	// Verify indexes exist
	indexes := []string{
		"idx_objects_kind",
		"idx_objects_created_by",
		"idx_objects_created_at",
	}
	for _, idx := range indexes {
		var indexName string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&indexName)
		if err != nil {
			t.Errorf("Index %s not found: %v", idx, err)
		}
	}
}
