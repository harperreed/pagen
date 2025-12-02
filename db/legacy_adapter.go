// ABOUTME: Adapter functions for converting between legacy models and Office OS objects
// ABOUTME: Provides bidirectional conversion for Contact, Company, and Deal types

package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/harperreed/pagen/models"
)

// Object type constants for CRM entities.
const (
	ObjectTypeCompany = "Company"
	ObjectTypeContact = "Contact"
	ObjectTypeDeal    = "Deal"
)

// Relationship type constants for CRM relationships.
const (
	RelTypeWorksAt     = "works_at"
	RelTypeDealContact = "deal_contact"
	RelTypeDealCompany = "deal_company"
	RelTypeKnows       = "knows"
)

// CompanyToObject converts a models.Company to an Office OS Object.
func CompanyToObject(company *models.Company) *Object {
	return &Object{
		ID:   company.ID.String(),
		Type: ObjectTypeCompany,
		Name: company.Name,
		Metadata: map[string]interface{}{
			"domain":   company.Domain,
			"industry": company.Industry,
			"notes":    company.Notes,
		},
		CreatedAt: company.CreatedAt,
		UpdatedAt: company.UpdatedAt,
	}
}

// ObjectToCompany converts an Office OS Object to a models.Company.
func ObjectToCompany(obj *Object) (*models.Company, error) {
	if obj.Type != ObjectTypeCompany {
		return nil, fmt.Errorf("object is not a Company type: %s", obj.Type)
	}

	id, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid company ID: %w", err)
	}

	company := &models.Company{
		ID:        id,
		Name:      obj.Name,
		CreatedAt: obj.CreatedAt,
		UpdatedAt: obj.UpdatedAt,
	}

	// Extract metadata fields with type assertions
	if domain, ok := obj.Metadata["domain"].(string); ok {
		company.Domain = domain
	}
	if industry, ok := obj.Metadata["industry"].(string); ok {
		company.Industry = industry
	}
	if notes, ok := obj.Metadata["notes"].(string); ok {
		company.Notes = notes
	}

	return company, nil
}

// ContactToObject converts a models.Contact to an Office OS Object.
func ContactToObject(contact *models.Contact) *Object {
	metadata := map[string]interface{}{
		"email": contact.Email,
		"phone": contact.Phone,
		"notes": contact.Notes,
	}

	if contact.CompanyID != nil {
		metadata["company_id"] = contact.CompanyID.String()
	}

	if contact.LastContactedAt != nil {
		metadata["last_contacted_at"] = contact.LastContactedAt.Format(time.RFC3339)
	}

	return &Object{
		ID:        contact.ID.String(),
		Type:      ObjectTypeContact,
		Name:      contact.Name,
		Metadata:  metadata,
		CreatedAt: contact.CreatedAt,
		UpdatedAt: contact.UpdatedAt,
	}
}

// ObjectToContact converts an Office OS Object to a models.Contact.
func ObjectToContact(obj *Object) (*models.Contact, error) {
	if obj.Type != ObjectTypeContact {
		return nil, fmt.Errorf("object is not a Contact type: %s", obj.Type)
	}

	id, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid contact ID: %w", err)
	}

	contact := &models.Contact{
		ID:        id,
		Name:      obj.Name,
		CreatedAt: obj.CreatedAt,
		UpdatedAt: obj.UpdatedAt,
	}

	// Extract metadata fields with type assertions
	if email, ok := obj.Metadata["email"].(string); ok {
		contact.Email = email
	}
	if phone, ok := obj.Metadata["phone"].(string); ok {
		contact.Phone = phone
	}
	if notes, ok := obj.Metadata["notes"].(string); ok {
		contact.Notes = notes
	}

	// Parse company_id if present
	if companyIDStr, ok := obj.Metadata["company_id"].(string); ok && companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err == nil {
			contact.CompanyID = &companyID
		}
	}

	// Parse last_contacted_at if present
	if lastContactedStr, ok := obj.Metadata["last_contacted_at"].(string); ok && lastContactedStr != "" {
		lastContacted, err := time.Parse(time.RFC3339, lastContactedStr)
		if err == nil {
			contact.LastContactedAt = &lastContacted
		}
	}

	return contact, nil
}

// DealToObject converts a models.Deal to an Office OS Object.
func DealToObject(deal *models.Deal) *Object {
	metadata := map[string]interface{}{
		"amount":   deal.Amount,
		"currency": deal.Currency,
		"stage":    deal.Stage,
	}

	// Company ID is required for deals
	metadata["company_id"] = deal.CompanyID.String()

	if deal.ContactID != nil {
		metadata["contact_id"] = deal.ContactID.String()
	}

	if deal.ExpectedCloseDate != nil {
		metadata["expected_close_date"] = deal.ExpectedCloseDate.Format(time.RFC3339)
	}

	metadata["last_activity_at"] = deal.LastActivityAt.Format(time.RFC3339)

	return &Object{
		ID:        deal.ID.String(),
		Type:      ObjectTypeDeal,
		Name:      deal.Title,
		Metadata:  metadata,
		CreatedAt: deal.CreatedAt,
		UpdatedAt: deal.UpdatedAt,
	}
}

// ObjectToDeal converts an Office OS Object to a models.Deal.
func ObjectToDeal(obj *Object) (*models.Deal, error) {
	if obj.Type != ObjectTypeDeal {
		return nil, fmt.Errorf("object is not a Deal type: %s", obj.Type)
	}

	id, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid deal ID: %w", err)
	}

	deal := &models.Deal{
		ID:        id,
		Title:     obj.Name,
		CreatedAt: obj.CreatedAt,
		UpdatedAt: obj.UpdatedAt,
		Currency:  "USD", // Default
	}

	// Extract required company_id
	companyIDStr, ok := obj.Metadata["company_id"].(string)
	if !ok || companyIDStr == "" {
		return nil, fmt.Errorf("deal missing required company_id")
	}
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid company_id: %w", err)
	}
	deal.CompanyID = companyID

	// Extract numeric fields - handle both float64 (from JSON) and int64
	if amount, ok := obj.Metadata["amount"].(float64); ok {
		deal.Amount = int64(amount)
	} else if amount, ok := obj.Metadata["amount"].(int64); ok {
		deal.Amount = amount
	}

	// Extract string fields
	if currency, ok := obj.Metadata["currency"].(string); ok {
		deal.Currency = currency
	}
	if stage, ok := obj.Metadata["stage"].(string); ok {
		deal.Stage = stage
	}

	// Parse optional contact_id
	if contactIDStr, ok := obj.Metadata["contact_id"].(string); ok && contactIDStr != "" {
		contactID, err := uuid.Parse(contactIDStr)
		if err == nil {
			deal.ContactID = &contactID
		}
	}

	// Parse optional expected_close_date
	if closeDateStr, ok := obj.Metadata["expected_close_date"].(string); ok && closeDateStr != "" {
		closeDate, err := time.Parse(time.RFC3339, closeDateStr)
		if err == nil {
			deal.ExpectedCloseDate = &closeDate
		}
	}

	// Parse last_activity_at (required for deals)
	if activityStr, ok := obj.Metadata["last_activity_at"].(string); ok && activityStr != "" {
		activity, err := time.Parse(time.RFC3339, activityStr)
		if err == nil {
			deal.LastActivityAt = activity
		}
	} else {
		// Default to created_at if missing
		deal.LastActivityAt = deal.CreatedAt
	}

	return deal, nil
}

// Helper function to safely get string from metadata.
func getStringFromMetadata(metadata map[string]interface{}, key string) string {
	if val, ok := metadata[key].(string); ok {
		return val
	}
	return ""
}

// Helper function to safely get int64 from metadata (handles JSON number conversion).
func getInt64FromMetadata(metadata map[string]interface{}, key string) int64 {
	if val, ok := metadata[key].(float64); ok {
		return int64(val)
	}
	if val, ok := metadata[key].(int64); ok {
		return val
	}
	if val, ok := metadata[key].(int); ok {
		return int64(val)
	}
	return 0
}

// Helper function to safely get time from metadata.
func getTimeFromMetadata(metadata map[string]interface{}, key string) (*time.Time, error) {
	if val, ok := metadata[key].(string); ok && val != "" {
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}
	return nil, nil
}

// Helper function to safely get UUID from metadata.
func getUUIDFromMetadata(metadata map[string]interface{}, key string) (*uuid.UUID, error) {
	if val, ok := metadata[key].(string); ok && val != "" {
		id, err := uuid.Parse(val)
		if err != nil {
			return nil, err
		}
		return &id, nil
	}
	return nil, nil
}

// Marshal metadata to JSON string for storage.
func marshalMetadata(data interface{}) (string, error) {
	if data == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}
	return string(bytes), nil
}

// Unmarshal metadata from JSON string.
func unmarshalMetadata(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return make(map[string]interface{}), nil
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	return metadata, nil
}
