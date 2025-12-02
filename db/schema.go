// ABOUTME: Database schema definitions and migrations
// ABOUTME: Handles SQLite table creation and initialization
package db

import (
	"database/sql"
	"time"
)

// Object represents the core entity in Office OS.
type Object struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func InitSchema(db *sql.DB) error {
	// Enable foreign keys for SQLite
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS objects (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		metadata TEXT NOT NULL DEFAULT '{}',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_objects_type ON objects(type);
	CREATE INDEX IF NOT EXISTS idx_objects_created_at ON objects(created_at);

	CREATE TABLE IF NOT EXISTS relationships (
		id TEXT PRIMARY KEY,
		source_id TEXT NOT NULL,
		target_id TEXT NOT NULL,
		type TEXT NOT NULL,
		metadata TEXT NOT NULL DEFAULT '{}',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (source_id) REFERENCES objects(id) ON DELETE CASCADE,
		FOREIGN KEY (target_id) REFERENCES objects(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_relationships_source ON relationships(source_id);
	CREATE INDEX IF NOT EXISTS idx_relationships_target ON relationships(target_id);
	CREATE INDEX IF NOT EXISTS idx_relationships_type ON relationships(type);

	-- Support tables for CRM functionality
	CREATE TABLE IF NOT EXISTS interaction_log (
		id TEXT PRIMARY KEY,
		contact_id TEXT NOT NULL,
		interaction_type TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		notes TEXT,
		sentiment TEXT,
		metadata TEXT DEFAULT '{}',
		FOREIGN KEY (contact_id) REFERENCES objects(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_interaction_log_contact ON interaction_log(contact_id);
	CREATE INDEX IF NOT EXISTS idx_interaction_log_timestamp ON interaction_log(timestamp);

	CREATE TABLE IF NOT EXISTS contact_cadence (
		contact_id TEXT PRIMARY KEY,
		cadence_days INTEGER NOT NULL,
		relationship_strength TEXT NOT NULL,
		priority_score REAL NOT NULL DEFAULT 0,
		last_interaction_date DATETIME,
		next_followup_date DATETIME,
		FOREIGN KEY (contact_id) REFERENCES objects(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS sync_state (
		service TEXT PRIMARY KEY,
		last_sync_time DATETIME,
		last_sync_token TEXT,
		status TEXT NOT NULL,
		error_message TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sync_log (
		id TEXT PRIMARY KEY,
		source_service TEXT NOT NULL,
		source_id TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id TEXT NOT NULL,
		imported_at DATETIME NOT NULL,
		metadata TEXT DEFAULT '{}',
		UNIQUE(source_service, source_id)
	);

	CREATE INDEX IF NOT EXISTS idx_sync_log_source ON sync_log(source_service, source_id);
	CREATE INDEX IF NOT EXISTS idx_sync_log_entity ON sync_log(entity_type, entity_id);

	CREATE TABLE IF NOT EXISTS suggestions (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		confidence REAL NOT NULL,
		source_service TEXT NOT NULL,
		source_id TEXT,
		source_data TEXT,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		reviewed_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_suggestions_status ON suggestions(status);
	CREATE INDEX IF NOT EXISTS idx_suggestions_type ON suggestions(type);
	`

	_, err := db.Exec(schema)
	return err
}
