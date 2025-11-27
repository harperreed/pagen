package cli

import (
	"database/sql"
	"os"
	"testing"

	"github.com/harperreed/pagen/db"
)

func setupTestCLI(t *testing.T) *sql.DB {
	tmpDB, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmpDB.Close()
	t.Cleanup(func() { _ = os.Remove(tmpDB.Name()) })

	database, err := db.OpenDatabase(tmpDB.Name())
	if err != nil {
		t.Fatal(err)
	}

	return database
}

func TestFollowupListCommand(t *testing.T) {
	database := setupTestCLI(t)
	defer func() { _ = database.Close() }()

	// Will test that command runs without error
	// Detailed output testing will be manual
	err := FollowupListCommand(database, []string{})
	if err != nil {
		t.Errorf("FollowupListCommand failed: %v", err)
	}
}
