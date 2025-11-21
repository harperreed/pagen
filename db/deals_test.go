// ABOUTME: Tests for deal and deal note database operations
// ABOUTME: Covers CRUD operations, stage updates, and note management
package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/harperreed/crm-mcp/models"
)

func TestCreateDeal(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create company
	company := &models.Company{Name: "Deal Corp"}
	if err := CreateCompany(db, company); err != nil {
		t.Fatalf("CreateCompany failed: %v", err)
	}

	deal := &models.Deal{
		Title:     "Big Deal",
		Amount:    100000,
		Currency:  "USD",
		Stage:     models.StageProspecting,
		CompanyID: company.ID,
	}

	if err := CreateDeal(db, deal); err != nil {
		t.Fatalf("CreateDeal failed: %v", err)
	}

	if deal.ID == uuid.Nil {
		t.Error("Deal ID was not set")
	}

	if deal.LastActivityAt.IsZero() {
		t.Error("LastActivityAt was not set")
	}
}

func TestUpdateDeal(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	company := &models.Company{Name: "Deal Corp"}
	if err := CreateCompany(db, company); err != nil {
		t.Fatalf("CreateCompany failed: %v", err)
	}

	deal := &models.Deal{
		Title:     "Test Deal",
		Stage:     models.StageProspecting,
		CompanyID: company.ID,
		Currency:  "USD",
	}

	if err := CreateDeal(db, deal); err != nil {
		t.Fatalf("CreateDeal failed: %v", err)
	}

	// Update stage
	deal.Stage = models.StageNegotiation
	deal.Amount = 50000

	if err := UpdateDeal(db, deal); err != nil {
		t.Fatalf("UpdateDeal failed: %v", err)
	}

	// Verify update
	found, err := GetDeal(db, deal.ID)
	if err != nil {
		t.Fatalf("GetDeal failed: %v", err)
	}

	if found.Stage != models.StageNegotiation {
		t.Errorf("Expected stage %s, got %s", models.StageNegotiation, found.Stage)
	}

	if found.Amount != 50000 {
		t.Errorf("Expected amount 50000, got %d", found.Amount)
	}
}

func TestAddDealNote(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	company := &models.Company{Name: "Note Corp"}
	if err := CreateCompany(db, company); err != nil {
		t.Fatalf("CreateCompany failed: %v", err)
	}

	deal := &models.Deal{
		Title:     "Note Deal",
		Stage:     models.StageProspecting,
		CompanyID: company.ID,
		Currency:  "USD",
	}

	if err := CreateDeal(db, deal); err != nil {
		t.Fatalf("CreateDeal failed: %v", err)
	}

	note := &models.DealNote{
		DealID:  deal.ID,
		Content: "Had a great call today",
	}

	if err := AddDealNote(db, note); err != nil {
		t.Fatalf("AddDealNote failed: %v", err)
	}

	if note.ID == uuid.Nil {
		t.Error("Note ID was not set")
	}

	// Verify note
	notes, err := GetDealNotes(db, deal.ID)
	if err != nil {
		t.Fatalf("GetDealNotes failed: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}

	if notes[0].Content != note.Content {
		t.Error("Note content mismatch")
	}
}
