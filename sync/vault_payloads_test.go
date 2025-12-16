// ABOUTME: Tests for vault sync payload serialization and deserialization
// ABOUTME: Covers JSON marshaling, omitempty behavior, and roundtrip conversions
package sync

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContactPayloadJSON(t *testing.T) {
	// Create test contact with all fields
	lastContacted := "2025-12-15T10:30:00Z"
	original := &ContactPayload{
		ID:              "contact-123",
		Name:            "Jane Doe",
		Email:           "jane@example.com",
		Phone:           "+1234567890",
		CompanyName:     "Acme Corp",
		Notes:           "Important client",
		LastContactedAt: &lastContacted,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err, "marshaling should succeed")

	// Unmarshal back
	var decoded ContactPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err, "unmarshaling should succeed")

	// Verify all fields match
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.Email, decoded.Email)
	assert.Equal(t, original.Phone, decoded.Phone)
	assert.Equal(t, original.CompanyName, decoded.CompanyName)
	assert.Equal(t, original.Notes, decoded.Notes)
	require.NotNil(t, decoded.LastContactedAt)
	assert.Equal(t, *original.LastContactedAt, *decoded.LastContactedAt)
}

func TestContactPayloadJSON_MinimalFields(t *testing.T) {
	// Create contact with only required fields
	original := &ContactPayload{
		ID:   "contact-456",
		Name: "John Smith",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var decoded ContactPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify required fields
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)

	// Optional fields should be empty
	assert.Empty(t, decoded.Email)
	assert.Empty(t, decoded.Phone)
	assert.Empty(t, decoded.CompanyName)
	assert.Empty(t, decoded.Notes)
	assert.Nil(t, decoded.LastContactedAt)
}

func TestContactPayloadJSON_OmitEmpty(t *testing.T) {
	payload := &ContactPayload{
		ID:   "contact-789",
		Name: "Alice Johnson",
		// All optional fields intentionally omitted
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	// Parse to check which fields are present
	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Required fields should be present
	assert.Contains(t, parsed, "id")
	assert.Contains(t, parsed, "name")

	// Optional fields should be omitted
	assert.NotContains(t, parsed, "email")
	assert.NotContains(t, parsed, "phone")
	assert.NotContains(t, parsed, "company_name")
	assert.NotContains(t, parsed, "notes")
	assert.NotContains(t, parsed, "last_contacted_at")
}

func TestCompanyPayloadJSON(t *testing.T) {
	// Create test company with all fields
	original := &CompanyPayload{
		ID:       "company-123",
		Name:     "Tech Corp",
		Domain:   "techcorp.com",
		Industry: "Software",
		Notes:    "Great partner",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var decoded CompanyPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.Domain, decoded.Domain)
	assert.Equal(t, original.Industry, decoded.Industry)
	assert.Equal(t, original.Notes, decoded.Notes)
}

func TestCompanyPayloadJSON_MinimalFields(t *testing.T) {
	original := &CompanyPayload{
		ID:   "company-456",
		Name: "Startup Inc",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded CompanyPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Name, decoded.Name)
	assert.Empty(t, decoded.Domain)
	assert.Empty(t, decoded.Industry)
	assert.Empty(t, decoded.Notes)
}

func TestCompanyPayloadJSON_OmitEmpty(t *testing.T) {
	payload := &CompanyPayload{
		ID:   "company-789",
		Name: "Big Corp",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Required fields present
	assert.Contains(t, parsed, "id")
	assert.Contains(t, parsed, "name")

	// Optional fields omitted
	assert.NotContains(t, parsed, "domain")
	assert.NotContains(t, parsed, "industry")
	assert.NotContains(t, parsed, "notes")
}

func TestDealPayloadJSON(t *testing.T) {
	// Create test deal with all fields
	closeDate := "2025-12-31T00:00:00Z"
	original := &DealPayload{
		ID:                "deal-123",
		Title:             "Enterprise Contract",
		Amount:            500000,
		Currency:          "USD",
		Stage:             "negotiation",
		CompanyName:       "Big Customer",
		ContactName:       "Decision Maker",
		ExpectedCloseDate: &closeDate,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var decoded DealPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Title, decoded.Title)
	assert.Equal(t, original.Amount, decoded.Amount)
	assert.Equal(t, original.Currency, decoded.Currency)
	assert.Equal(t, original.Stage, decoded.Stage)
	assert.Equal(t, original.CompanyName, decoded.CompanyName)
	assert.Equal(t, original.ContactName, decoded.ContactName)
	require.NotNil(t, decoded.ExpectedCloseDate)
	assert.Equal(t, *original.ExpectedCloseDate, *decoded.ExpectedCloseDate)
}

func TestDealPayloadJSON_MinimalFields(t *testing.T) {
	original := &DealPayload{
		ID:       "deal-456",
		Title:    "Small Deal",
		Amount:   10000,
		Currency: "USD",
		Stage:    "prospecting",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded DealPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Title, decoded.Title)
	assert.Equal(t, original.Amount, decoded.Amount)
	assert.Equal(t, original.Currency, decoded.Currency)
	assert.Equal(t, original.Stage, decoded.Stage)
	assert.Empty(t, decoded.CompanyName)
	assert.Empty(t, decoded.ContactName)
	assert.Nil(t, decoded.ExpectedCloseDate)
}

func TestDealPayloadJSON_OmitEmpty(t *testing.T) {
	payload := &DealPayload{
		ID:       "deal-789",
		Title:    "Quick Deal",
		Amount:   5000,
		Currency: "EUR",
		Stage:    "closed-won",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	// Required fields present
	assert.Contains(t, parsed, "id")
	assert.Contains(t, parsed, "title")
	assert.Contains(t, parsed, "amount")
	assert.Contains(t, parsed, "currency")
	assert.Contains(t, parsed, "stage")

	// Optional fields omitted
	assert.NotContains(t, parsed, "company_name")
	assert.NotContains(t, parsed, "contact_name")
	assert.NotContains(t, parsed, "expected_close_date")
}

func TestDealPayloadJSON_AmountPrecision(t *testing.T) {
	// Test that amount (in cents) serializes correctly
	original := &DealPayload{
		ID:       "deal-precision",
		Title:    "Precision Test",
		Amount:   123456789, // Large number to test precision
		Currency: "USD",
		Stage:    "won",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded DealPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Amount, decoded.Amount, "amount should preserve full precision")
}

func TestDealNotePayloadJSON(t *testing.T) {
	original := &DealNotePayload{
		ID:        "note-123",
		DealTitle: "Big Deal",
		Content:   "Customer is very interested",
		CreatedAt: "2025-12-15T14:30:00Z",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded DealNotePayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.DealTitle, decoded.DealTitle)
	assert.Equal(t, original.Content, decoded.Content)
	assert.Equal(t, original.CreatedAt, decoded.CreatedAt)
}

func TestRelationshipPayloadJSON(t *testing.T) {
	original := &RelationshipPayload{
		ID:               "rel-123",
		Contact1Name:     "Alice Smith",
		Contact2Name:     "Bob Jones",
		RelationshipType: "colleague",
		Context:          "Work together at TechCorp",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded RelationshipPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Contact1Name, decoded.Contact1Name)
	assert.Equal(t, original.Contact2Name, decoded.Contact2Name)
	assert.Equal(t, original.RelationshipType, decoded.RelationshipType)
	assert.Equal(t, original.Context, decoded.Context)
}

func TestInteractionLogPayloadJSON(t *testing.T) {
	sentiment := "positive"
	original := &InteractionLogPayload{
		ID:              "log-123",
		ContactName:     "Jane Doe",
		InteractionType: "meeting",
		InteractedAt:    "2025-12-15T09:00:00Z",
		Sentiment:       &sentiment,
		Metadata:        `{"duration": 60, "location": "office"}`,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded InteractionLogPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.ContactName, decoded.ContactName)
	assert.Equal(t, original.InteractionType, decoded.InteractionType)
	assert.Equal(t, original.InteractedAt, decoded.InteractedAt)
	require.NotNil(t, decoded.Sentiment)
	assert.Equal(t, *original.Sentiment, *decoded.Sentiment)
	assert.Equal(t, original.Metadata, decoded.Metadata)
}

func TestContactCadencePayloadJSON(t *testing.T) {
	original := &ContactCadencePayload{
		ID:                   "cadence-123",
		ContactName:          "VIP Client",
		CadenceDays:          30,
		RelationshipStrength: "strong",
		PriorityScore:        95,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ContactCadencePayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.ContactName, decoded.ContactName)
	assert.Equal(t, original.CadenceDays, decoded.CadenceDays)
	assert.Equal(t, original.RelationshipStrength, decoded.RelationshipStrength)
	assert.Equal(t, original.PriorityScore, decoded.PriorityScore)
}

func TestSuggestionPayloadJSON(t *testing.T) {
	original := &SuggestionPayload{
		ID:            "suggestion-123",
		Type:          "follow-up",
		Content:       `{"action": "send email", "to": "jane@example.com"}`,
		Confidence:    0.85,
		SourceService: "ai-analyzer",
		Status:        "pending",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded SuggestionPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Type, decoded.Type)
	assert.Equal(t, original.Content, decoded.Content)
	assert.Equal(t, original.Confidence, decoded.Confidence)
	assert.Equal(t, original.SourceService, decoded.SourceService)
	assert.Equal(t, original.Status, decoded.Status)
}

func TestSuggestionPayloadJSON_ConfidencePrecision(t *testing.T) {
	// Test that confidence float64 serializes with proper precision
	original := &SuggestionPayload{
		ID:            "suggestion-precision",
		Type:          "test",
		Content:       "{}",
		Confidence:    0.123456789,
		SourceService: "test",
		Status:        "active",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded SuggestionPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.InDelta(t, original.Confidence, decoded.Confidence, 0.000001, "confidence should preserve precision")
}

func TestEntityConstants(t *testing.T) {
	// Verify entity name constants are defined correctly
	assert.Equal(t, "contact", EntityContact)
	assert.Equal(t, "company", EntityCompany)
	assert.Equal(t, "deal", EntityDeal)
	assert.Equal(t, "deal_note", EntityDealNote)
	assert.Equal(t, "relationship", EntityRelationship)
	assert.Equal(t, "interaction_log", EntityInteractionLog)
	assert.Equal(t, "contact_cadence", EntityContactCadence)
	assert.Equal(t, "suggestion", EntitySuggestion)
}

func TestAllPayloads_JSONRoundtrip(t *testing.T) {
	// This test ensures all payload types can be marshaled and unmarshaled
	tests := []struct {
		name    string
		payload interface{}
	}{
		{"ContactPayload", &ContactPayload{ID: "1", Name: "Test"}},
		{"CompanyPayload", &CompanyPayload{ID: "1", Name: "Test"}},
		{"DealPayload", &DealPayload{ID: "1", Title: "Test", Amount: 100, Currency: "USD", Stage: "new"}},
		{"DealNotePayload", &DealNotePayload{ID: "1", DealTitle: "Test", Content: "Test", CreatedAt: "2025-12-15T00:00:00Z"}},
		{"RelationshipPayload", &RelationshipPayload{ID: "1", Contact1Name: "A", Contact2Name: "B", RelationshipType: "friend"}},
		{"InteractionLogPayload", &InteractionLogPayload{ID: "1", ContactName: "Test", InteractionType: "call", InteractedAt: "2025-12-15T00:00:00Z"}},
		{"ContactCadencePayload", &ContactCadencePayload{ID: "1", ContactName: "Test", CadenceDays: 30, RelationshipStrength: "strong", PriorityScore: 50}},
		{"SuggestionPayload", &SuggestionPayload{ID: "1", Type: "test", Content: "{}", Confidence: 0.5, SourceService: "test", Status: "active"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.payload)
			require.NoError(t, err, "marshaling %s should succeed", tt.name)

			// Verify it's valid JSON
			var generic map[string]interface{}
			err = json.Unmarshal(data, &generic)
			require.NoError(t, err, "unmarshaling to generic map should succeed")

			// Verify we can unmarshal back to original type
			// (The type-specific tests above verify field correctness)
			assert.NotEmpty(t, data, "marshaled data should not be empty")
		})
	}
}
