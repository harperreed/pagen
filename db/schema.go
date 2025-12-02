// ABOUTME: Database schema definitions and migrations
// ABOUTME: Handles SQLite table creation and initialization
package db

import (
	"database/sql"
)

func InitSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS objects (
		id TEXT PRIMARY KEY,
		kind TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		created_by TEXT NOT NULL,
		acl TEXT NOT NULL,
		tags TEXT,
		fields TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_objects_kind ON objects(kind);
	CREATE INDEX IF NOT EXISTS idx_objects_created_by ON objects(created_by);
	CREATE INDEX IF NOT EXISTS idx_objects_created_at ON objects(created_at);
	`

	_, err := db.Exec(schema)
	return err
}
