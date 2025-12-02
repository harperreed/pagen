# Office OS Foundation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build core Office OS primitives - unified objects table, BaseObject structs, CRUD operations, and migration script.

**Architecture:** Single `objects` table with JSONB fields. Generic CRUD operations. Type-safe Go structs that serialize to/from JSON. Migration script nukes old DB.

**Tech Stack:** Go 1.21+, SQLite with JSON functions, uuid for IDs

---

## Task 1: Create Core Schema

**Files:**
- Create: `db/objects.go`
- Create: `db/objects_test.go`
- Modify: `db/schema.go` (replace entire InitSchema function)

**Step 1: Write failing test for objects table creation**

Create `db/objects_test.go`:

```go
package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestObjectsTableCreation(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}
	defer database.Close()

	if err := InitSchema(database); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Verify objects table exists
	var tableName string
	err = database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='objects'").Scan(&tableName)
	if err != nil {
		t.Fatalf("objects table not found: %v", err)
	}
	if tableName != "objects" {
		t.Errorf("Expected table name 'objects', got %s", tableName)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/harper/Public/src/personal/pagen/.worktrees/office-os-foundation
go test ./db -run TestObjectsTableCreation -v
```

Expected: FAIL (objects table doesn't exist)

**Step 3: Replace schema with objects table**

Modify `db/schema.go` - replace entire `InitSchema` function:

```go
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
```

**Step 4: Run test to verify it passes**

```bash
go test ./db -run TestObjectsTableCreation -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add db/schema.go db/objects_test.go
git commit -m "feat: create unified objects table schema

Replace specialized CRM tables with single objects table.
Supports Office OS object model with JSONB fields."
```

---

## Task 2: BaseObject Structs

**Files:**
- Create: `models/objects.go`
- Create: `models/objects_test.go`

**Step 1: Write failing test for BaseObject marshaling**

Create `models/objects_test.go`:

```go
package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBaseObjectMarshaling(t *testing.T) {
	obj := BaseObject{
		ID:        "test-id",
		Kind:      KindRecord,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		CreatedBy: "user-1",
		ACL: []Permission{
			{ActorID: "user-1", Role: RoleOwner},
		},
		Tags: []string{"test", "crm"},
	}

	// Marshal to JSON
	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded BaseObject
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ID != obj.ID {
		t.Errorf("Expected ID %s, got %s", obj.ID, decoded.ID)
	}
	if decoded.Kind != obj.Kind {
		t.Errorf("Expected Kind %s, got %s", obj.Kind, decoded.Kind)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./models -run TestBaseObjectMarshaling -v
```

Expected: FAIL (BaseObject not defined)

**Step 3: Create BaseObject struct**

Create `models/objects.go`:

```go
package models

import "time"

type ObjectKind string

const (
	KindUser         ObjectKind = "user"
	KindRecord       ObjectKind = "record"
	KindTask         ObjectKind = "task"
	KindEvent        ObjectKind = "event"
	KindMessage      ObjectKind = "message"
	KindActivity     ObjectKind = "activity"
	KindNotification ObjectKind = "notification"
)

type Role string

const (
	RoleOwner     Role = "owner"
	RoleEditor    Role = "editor"
	RoleCommenter Role = "commenter"
	RoleViewer    Role = "viewer"
)

type Permission struct {
	ActorID string `json:"actorId"`
	Role    Role   `json:"role"`
}

type BaseObject struct {
	ID        string       `json:"id"`
	Kind      ObjectKind   `json:"kind"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	CreatedBy string       `json:"createdBy"`
	ACL       []Permission `json:"acl"`
	Tags      []string     `json:"tags,omitempty"`
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./models -run TestBaseObjectMarshaling -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add models/objects.go models/objects_test.go
git commit -m "feat: add BaseObject and ObjectKind types

Core types for Office OS object model with JSON serialization."
```

---

## Task 3: RecordObject (CRM Schemas)

**Files:**
- Modify: `models/objects.go`
- Modify: `models/objects_test.go`

**Step 1: Write failing test for RecordObject**

Add to `models/objects_test.go`:

```go
func TestRecordObjectSerialization(t *testing.T) {
	contact := RecordObject{
		BaseObject: BaseObject{
			ID:        "contact-1",
			Kind:      KindRecord,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			CreatedBy: "user-1",
			ACL: []Permission{
				{ActorID: "user-1", Role: RoleOwner},
			},
		},
		Title:    "Sarah Chen",
		SchemaID: SchemaContact,
		Fields: map[string]interface{}{
			"firstName": "Sarah",
			"lastName":  "Chen",
			"email":     "sarah@acme.com",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(contact)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify fields are present
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if raw["title"] != "Sarah Chen" {
		t.Errorf("Expected title 'Sarah Chen', got %v", raw["title"])
	}

	fields := raw["fields"].(map[string]interface{})
	if fields["email"] != "sarah@acme.com" {
		t.Errorf("Expected email sarah@acme.com, got %v", fields["email"])
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./models -run TestRecordObjectSerialization -v
```

Expected: FAIL (RecordObject not defined)

**Step 3: Add RecordObject struct**

Add to `models/objects.go`:

```go
const (
	SchemaContact     = "schema:crm_contact"
	SchemaAccount     = "schema:crm_account"
	SchemaOpportunity = "schema:crm_opportunity"
)

type RecordObject struct {
	BaseObject
	Title    string                 `json:"title"`
	SchemaID string                 `json:"schemaId,omitempty"`
	Fields   map[string]interface{} `json:"fields"`
	Links    []LinkRef              `json:"links,omitempty"`
}

type LinkType string

const (
	LinkReferences  LinkType = "references"
	LinkAttachedTo  LinkType = "attached_to"
	LinkSubtaskOf   LinkType = "subtask_of"
	LinkDuplicateOf LinkType = "duplicate_of"
	LinkRelatedTo   LinkType = "related_to"
)

type LinkRef struct {
	Type     LinkType `json:"type"`
	TargetID string   `json:"targetId"`
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./models -run TestRecordObjectSerialization -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add models/objects.go models/objects_test.go
git commit -m "feat: add RecordObject and CRM schemas

Supports contacts, accounts, opportunities as RecordObjects."
```

---

## Task 4: CRUD Operations

**Files:**
- Modify: `db/objects.go`
- Modify: `db/objects_test.go`

**Step 1: Write failing test for CreateObject**

Add to `db/objects_test.go`:

```go
func TestCreateObject(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}
	defer database.Close()

	if err := InitSchema(database); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	contact := &models.RecordObject{
		BaseObject: models.BaseObject{
			ID:        "contact-test",
			Kind:      models.KindRecord,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			CreatedBy: "user-1",
			ACL: []models.Permission{
				{ActorID: "user-1", Role: models.RoleOwner},
			},
		},
		Title:    "Test Contact",
		SchemaID: models.SchemaContact,
		Fields: map[string]interface{}{
			"email": "test@example.com",
		},
	}

	if err := CreateObject(database, contact); err != nil {
		t.Fatalf("Failed to create object: %v", err)
	}

	// Verify it was created
	var id string
	err = database.QueryRow("SELECT id FROM objects WHERE id = ?", contact.ID).Scan(&id)
	if err != nil {
		t.Fatalf("Object not found in database: %v", err)
	}
	if id != "contact-test" {
		t.Errorf("Expected ID contact-test, got %s", id)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./db -run TestCreateObject -v
```

Expected: FAIL (CreateObject not defined)

**Step 3: Implement CreateObject**

Create `db/objects.go`:

```go
package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/harperreed/pagen/models"
)

func CreateObject(db *sql.DB, obj interface{}) error {
	var base models.BaseObject
	var fields map[string]interface{}

	// Extract BaseObject and fields based on type
	switch v := obj.(type) {
	case *models.RecordObject:
		base = v.BaseObject
		// Serialize RecordObject fields
		data := map[string]interface{}{
			"schemaId": v.SchemaID,
			"title":    v.Title,
		}
		for k, val := range v.Fields {
			data[k] = val
		}
		fields = data
	default:
		return fmt.Errorf("unsupported object type: %T", obj)
	}

	// Serialize ACL and tags
	aclJSON, err := json.Marshal(base.ACL)
	if err != nil {
		return fmt.Errorf("failed to marshal ACL: %w", err)
	}

	tagsJSON := "null"
	if len(base.Tags) > 0 {
		data, err := json.Marshal(base.Tags)
		if err != nil {
			return fmt.Errorf("failed to marshal tags: %w", err)
		}
		tagsJSON = string(data)
	}

	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}

	query := `
		INSERT INTO objects (id, kind, created_at, updated_at, created_by, acl, tags, fields)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = db.Exec(query,
		base.ID,
		base.Kind,
		base.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		base.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		base.CreatedBy,
		string(aclJSON),
		tagsJSON,
		string(fieldsJSON),
	)

	return err
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./db -run TestCreateObject -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add db/objects.go db/objects_test.go
git commit -m "feat: add CreateObject CRUD operation

Serializes objects to unified objects table with JSONB."
```

---

## Task 5: GetObject Operation

**Files:**
- Modify: `db/objects.go`
- Modify: `db/objects_test.go`

**Step 1: Write failing test for GetObject**

Add to `db/objects_test.go`:

```go
func TestGetObject(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}
	defer database.Close()

	if err := InitSchema(database); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Create test object
	original := &models.RecordObject{
		BaseObject: models.BaseObject{
			ID:        "contact-get-test",
			Kind:      models.KindRecord,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			CreatedBy: "user-1",
			ACL: []models.Permission{
				{ActorID: "user-1", Role: models.RoleOwner},
			},
		},
		Title:    "Get Test",
		SchemaID: models.SchemaContact,
		Fields: map[string]interface{}{
			"email": "get@example.com",
		},
	}

	if err := CreateObject(database, original); err != nil {
		t.Fatalf("Failed to create: %v", err)
	}

	// Get it back
	retrieved := &models.RecordObject{}
	if err := GetObject(database, "contact-get-test", retrieved); err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	if retrieved.Title != "Get Test" {
		t.Errorf("Expected title 'Get Test', got %s", retrieved.Title)
	}
	if retrieved.Fields["email"] != "get@example.com" {
		t.Errorf("Expected email get@example.com, got %v", retrieved.Fields["email"])
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./db -run TestGetObject -v
```

Expected: FAIL (GetObject not defined)

**Step 3: Implement GetObject**

Add to `db/objects.go`:

```go
func GetObject(db *sql.DB, id string, dest interface{}) error {
	query := `
		SELECT id, kind, created_at, updated_at, created_by, acl, tags, fields
		FROM objects
		WHERE id = ?
	`

	var (
		objID      string
		kind       string
		createdAt  string
		updatedAt  string
		createdBy  string
		aclJSON    string
		tagsJSON   sql.NullString
		fieldsJSON string
	)

	err := db.QueryRow(query, id).Scan(
		&objID, &kind, &createdAt, &updatedAt, &createdBy,
		&aclJSON, &tagsJSON, &fieldsJSON,
	)
	if err != nil {
		return err
	}

	// Parse timestamps
	createdTime, _ := time.Parse(time.RFC3339, createdAt)
	updatedTime, _ := time.Parse(time.RFC3339, updatedAt)

	// Parse ACL
	var acl []models.Permission
	if err := json.Unmarshal([]byte(aclJSON), &acl); err != nil {
		return fmt.Errorf("failed to parse ACL: %w", err)
	}

	// Parse tags
	var tags []string
	if tagsJSON.Valid && tagsJSON.String != "null" {
		if err := json.Unmarshal([]byte(tagsJSON.String), &tags); err != nil {
			return fmt.Errorf("failed to parse tags: %w", err)
		}
	}

	// Parse fields
	var fields map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return fmt.Errorf("failed to parse fields: %w", err)
	}

	// Populate destination based on type
	switch v := dest.(type) {
	case *models.RecordObject:
		v.ID = objID
		v.Kind = models.ObjectKind(kind)
		v.CreatedAt = createdTime
		v.UpdatedAt = updatedTime
		v.CreatedBy = createdBy
		v.ACL = acl
		v.Tags = tags

		// Extract schema-specific fields
		if schemaID, ok := fields["schemaId"].(string); ok {
			v.SchemaID = schemaID
			delete(fields, "schemaId")
		}
		if title, ok := fields["title"].(string); ok {
			v.Title = title
			delete(fields, "title")
		}
		v.Fields = fields
	default:
		return fmt.Errorf("unsupported destination type: %T", dest)
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./db -run TestGetObject -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add db/objects.go db/objects_test.go
git commit -m "feat: add GetObject CRUD operation

Deserializes objects from JSONB fields."
```

---

## Task 6: ListObjects Query

**Files:**
- Modify: `db/objects.go`
- Modify: `db/objects_test.go`

**Step 1: Write failing test for ListObjects**

Add to `db/objects_test.go`:

```go
func TestListObjects(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}
	defer database.Close()

	if err := InitSchema(database); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Create multiple contacts
	for i := 1; i <= 3; i++ {
		contact := &models.RecordObject{
			BaseObject: models.BaseObject{
				ID:        fmt.Sprintf("contact-%d", i),
				Kind:      models.KindRecord,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
				CreatedBy: "user-1",
				ACL: []models.Permission{
					{ActorID: "user-1", Role: models.RoleOwner},
				},
			},
			Title:    fmt.Sprintf("Contact %d", i),
			SchemaID: models.SchemaContact,
			Fields:   map[string]interface{}{},
		}
		if err := CreateObject(database, contact); err != nil {
			t.Fatalf("Failed to create: %v", err)
		}
	}

	// List all contacts
	var contacts []models.RecordObject
	err = ListObjects(database, models.KindRecord, models.SchemaContact, &contacts)
	if err != nil {
		t.Fatalf("Failed to list: %v", err)
	}

	if len(contacts) != 3 {
		t.Errorf("Expected 3 contacts, got %d", len(contacts))
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./db -run TestListObjects -v
```

Expected: FAIL (ListObjects not defined)

**Step 3: Implement ListObjects**

Add to `db/objects.go`:

```go
func ListObjects(db *sql.DB, kind models.ObjectKind, schemaID string, dest interface{}) error {
	query := `
		SELECT id, kind, created_at, updated_at, created_by, acl, tags, fields
		FROM objects
		WHERE kind = ? AND json_extract(fields, '$.schemaId') = ?
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, kind, schemaID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Determine destination slice type
	switch v := dest.(type) {
	case *[]models.RecordObject:
		*v = []models.RecordObject{}

		for rows.Next() {
			var (
				objID      string
				objKind    string
				createdAt  string
				updatedAt  string
				createdBy  string
				aclJSON    string
				tagsJSON   sql.NullString
				fieldsJSON string
			)

			if err := rows.Scan(&objID, &objKind, &createdAt, &updatedAt, &createdBy, &aclJSON, &tagsJSON, &fieldsJSON); err != nil {
				return err
			}

			rec := models.RecordObject{}
			if err := populateRecordObject(&rec, objID, objKind, createdAt, updatedAt, createdBy, aclJSON, tagsJSON, fieldsJSON); err != nil {
				return err
			}

			*v = append(*v, rec)
		}
	default:
		return fmt.Errorf("unsupported destination type: %T", dest)
	}

	return rows.Err()
}

func populateRecordObject(dest *models.RecordObject, id, kind, createdAt, updatedAt, createdBy, aclJSON string, tagsJSON sql.NullString, fieldsJSON string) error {
	createdTime, _ := time.Parse(time.RFC3339, createdAt)
	updatedTime, _ := time.Parse(time.RFC3339, updatedAt)

	var acl []models.Permission
	if err := json.Unmarshal([]byte(aclJSON), &acl); err != nil {
		return err
	}

	var tags []string
	if tagsJSON.Valid && tagsJSON.String != "null" {
		if err := json.Unmarshal([]byte(tagsJSON.String), &tags); err != nil {
			return err
		}
	}

	var fields map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return err
	}

	dest.ID = id
	dest.Kind = models.ObjectKind(kind)
	dest.CreatedAt = createdTime
	dest.UpdatedAt = updatedTime
	dest.CreatedBy = createdBy
	dest.ACL = acl
	dest.Tags = tags

	if schemaID, ok := fields["schemaId"].(string); ok {
		dest.SchemaID = schemaID
		delete(fields, "schemaId")
	}
	if title, ok := fields["title"].(string); ok {
		dest.Title = title
		delete(fields, "title")
	}
	dest.Fields = fields

	return nil
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./db -run TestListObjects -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add db/objects.go db/objects_test.go
git commit -m "feat: add ListObjects query operation

Query objects by kind and schemaId with JSON extraction."
```

---

## Task 7: Migration Script

**Files:**
- Create: `cmd/migrate-to-office-os/main.go`

**Step 1: Create migration script**

Create `cmd/migrate-to-office-os/main.go`:

```go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/harperreed/pagen/db"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("=== Office OS Migration ===")
	fmt.Println()
	fmt.Println("WARNING: This will DELETE your existing pagen database and create a fresh Office OS schema.")
	fmt.Println("All existing data (contacts, companies, deals) will be LOST.")
	fmt.Println("You will need to re-sync from Google to populate the new schema.")
	fmt.Println()
	fmt.Print("Type 'DELETE' to confirm: ")

	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "DELETE" {
		fmt.Println("Migration cancelled.")
		return
	}

	// Get database path
	dbPath := filepath.Join(xdg.DataHome, "crm", "crm.db")

	fmt.Printf("\nDeleting database: %s\n", dbPath)
	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to delete database: %v", err)
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// Open new database
	fmt.Println("Creating new Office OS schema...")
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Initialize Office OS schema
	if err := db.InitSchema(database); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	fmt.Println("\n✓ Migration complete!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'pagen sync init' to authenticate with Google")
	fmt.Println("  2. Run 'pagen sync' to populate the new schema")
	fmt.Println()
}
```

**Step 2: Test migration script**

```bash
go run cmd/migrate-to-office-os/main.go
# Type something other than DELETE to verify cancellation
```

Expected: "Migration cancelled."

**Step 3: Commit**

```bash
git add cmd/migrate-to-office-os/main.go
git commit -m "feat: add Office OS migration script

Deletes old database and creates fresh Office OS schema.
Requires re-sync from Google."
```

---

## Task 8: Update README

**Files:**
- Modify: `README.md`

**Step 1: Add migration notice to README**

Add to top of `README.md` after title:

```markdown
## ⚠️ Breaking Change: Office OS Migration

**Version 0.5.0+ uses a new unified data model.**

If upgrading from an earlier version:
1. Run `go run cmd/migrate-to-office-os/main.go` to migrate
2. Re-authenticate: `pagen sync init`
3. Re-sync data: `pagen sync`

Your old data will be lost. The migration re-creates everything from Google sync.
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add Office OS migration warning to README"
```

---

## Task 9: Integration Test

**Files:**
- Create: `db/objects_integration_test.go`

**Step 1: Write integration test**

Create `db/objects_integration_test.go`:

```go
package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/harperreed/pagen/models"
	_ "github.com/mattn/go-sqlite3"
)

func TestObjectCRUDIntegration(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}
	defer database.Close()

	if err := InitSchema(database); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Create a contact
	contact := &models.RecordObject{
		BaseObject: models.BaseObject{
			ID:        "integration-test-contact",
			Kind:      models.KindRecord,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			CreatedBy: "user-1",
			ACL: []models.Permission{
				{ActorID: "user-1", Role: models.RoleOwner},
			},
			Tags: []string{"test", "integration"},
		},
		Title:    "Integration Test Contact",
		SchemaID: models.SchemaContact,
		Fields: map[string]interface{}{
			"firstName": "Integration",
			"lastName":  "Test",
			"email":     "integration@test.com",
		},
	}

	// Create
	if err := CreateObject(database, contact); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get
	retrieved := &models.RecordObject{}
	if err := GetObject(database, contact.ID, retrieved); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Title != contact.Title {
		t.Errorf("Expected title %s, got %s", contact.Title, retrieved.Title)
	}
	if retrieved.Fields["email"] != "integration@test.com" {
		t.Errorf("Email mismatch: %v", retrieved.Fields["email"])
	}
	if len(retrieved.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(retrieved.Tags))
	}

	// List
	var contacts []models.RecordObject
	if err := ListObjects(database, models.KindRecord, models.SchemaContact, &contacts); err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}
}
```

**Step 2: Run integration test**

```bash
go test ./db -run TestObjectCRUDIntegration -v
```

Expected: PASS (all CRUD operations work together)

**Step 3: Commit**

```bash
git add db/objects_integration_test.go
git commit -m "test: add Office OS CRUD integration test

Verifies Create, Get, List operations work end-to-end."
```

---

## Task 10: Run All Tests

**Step 1: Run full test suite**

```bash
go test ./... -v
```

Expected: All tests PASS

**Step 2: Build binary**

```bash
CGO_ENABLED=1 go build -o pagen
```

Expected: Build succeeds

**Step 3: Final commit and push**

```bash
git add .
git commit -m "feat: complete Office OS foundation

Foundation includes:
- Unified objects table with JSONB fields
- BaseObject, RecordObject, LinkRef types
- Create, Get, List CRUD operations
- CRM schemas (contact, account, opportunity)
- Migration script
- Comprehensive test coverage

Ready for Activity and Task system integration."
git push -u origin feature/office-os-foundation
```

---

## Completion Checklist

- [ ] Objects table schema created
- [ ] BaseObject types defined
- [ ] RecordObject with CRM schemas
- [ ] CreateObject operation
- [ ] GetObject operation
- [ ] ListObjects query
- [ ] Migration script
- [ ] README updated with migration warning
- [ ] Integration tests passing
- [ ] All tests passing
- [ ] Foundation branch pushed

## Next Steps

After foundation is complete:
1. **Merge to main** (breaks existing database)
2. **Run migration script**
3. **Activity system** can build on top (reads/writes ActivityObject)
4. **Task system** can build on top (reads/writes TaskObject)
