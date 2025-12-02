// ABOUTME: Adapter functions for converting between legacy models and Office OS objects
// ABOUTME: Provides bidirectional conversion for Contact, Company, and Deal types

package db

import (
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
		ID:        company.ID.String(),
		Kind:      ObjectTypeCompany,
		CreatedAt: company.CreatedAt,
		UpdatedAt: company.UpdatedAt,
		CreatedBy: "system",
		ACL:       `[{"actorId":"system","role":"owner"}]`,
		Tags:      "[]",
		Fields: map[string]interface{}{
			"name":     company.Name,
			"domain":   company.Domain,
			"industry": company.Industry,
			"notes":    company.Notes,
		},
	}
}

// ObjectToCompany converts an Office OS Object to a models.Company.
func ObjectToCompany(obj *Object) (*models.Company, error) {
	if obj.Kind != ObjectTypeCompany {
		return nil, fmt.Errorf("object is not a Company type: %s", obj.Kind)
	}

	id, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid company ID: %w", err)
	}

	company := &models.Company{
		ID:        id,
		CreatedAt: obj.CreatedAt,
		UpdatedAt: obj.UpdatedAt,
	}

	// Extract fields with type assertions
	if name, ok := obj.Fields["name"].(string); ok {
		company.Name = name
	}
	if domain, ok := obj.Fields["domain"].(string); ok {
		company.Domain = domain
	}
	if industry, ok := obj.Fields["industry"].(string); ok {
		company.Industry = industry
	}
	if notes, ok := obj.Fields["notes"].(string); ok {
		company.Notes = notes
	}

	return company, nil
}

// ContactToObject converts a models.Contact to an Office OS Object.
func ContactToObject(contact *models.Contact) *Object {
	fields := map[string]interface{}{
		"name":  contact.Name,
		"email": contact.Email,
		"phone": contact.Phone,
		"notes": contact.Notes,
	}

	if contact.CompanyID != nil {
		fields["company_id"] = contact.CompanyID.String()
	}

	if contact.LastContactedAt != nil {
		fields["last_contacted_at"] = contact.LastContactedAt.Format(time.RFC3339Nano)
	}

	return &Object{
		ID:        contact.ID.String(),
		Kind:      ObjectTypeContact,
		CreatedAt: contact.CreatedAt,
		UpdatedAt: contact.UpdatedAt,
		CreatedBy: "system",
		ACL:       `[{"actorId":"system","role":"owner"}]`,
		Tags:      "[]",
		Fields:    fields,
	}
}

// ObjectToContact converts an Office OS Object to a models.Contact.
func ObjectToContact(obj *Object) (*models.Contact, error) {
	if obj.Kind != ObjectTypeContact {
		return nil, fmt.Errorf("object is not a Contact type: %s", obj.Kind)
	}

	id, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid contact ID: %w", err)
	}

	contact := &models.Contact{
		ID:        id,
		CreatedAt: obj.CreatedAt,
		UpdatedAt: obj.UpdatedAt,
	}

	// Extract fields with type assertions
	if name, ok := obj.Fields["name"].(string); ok {
		contact.Name = name
	}
	if email, ok := obj.Fields["email"].(string); ok {
		contact.Email = email
	}
	if phone, ok := obj.Fields["phone"].(string); ok {
		contact.Phone = phone
	}
	if notes, ok := obj.Fields["notes"].(string); ok {
		contact.Notes = notes
	}

	// Parse company_id if present
	if companyIDStr, ok := obj.Fields["company_id"].(string); ok && companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err == nil {
			contact.CompanyID = &companyID
		}
	}

	// Parse last_contacted_at if present
	if lastContactedStr, ok := obj.Fields["last_contacted_at"].(string); ok && lastContactedStr != "" {
		lastContacted, err := time.Parse(time.RFC3339Nano, lastContactedStr)
		if err == nil {
			contact.LastContactedAt = &lastContacted
		}
	}

	return contact, nil
}

// DealToObject converts a models.Deal to an Office OS Object.
func DealToObject(deal *models.Deal) *Object {
	fields := map[string]interface{}{
		"title":    deal.Title,
		"amount":   deal.Amount,
		"currency": deal.Currency,
		"stage":    deal.Stage,
	}

	// Company ID is required for deals
	fields["company_id"] = deal.CompanyID.String()

	if deal.ContactID != nil {
		fields["contact_id"] = deal.ContactID.String()
	}

	if deal.ExpectedCloseDate != nil {
		fields["expected_close_date"] = deal.ExpectedCloseDate.Format(time.RFC3339Nano)
	}

	fields["last_activity_at"] = deal.LastActivityAt.Format(time.RFC3339Nano)

	return &Object{
		ID:        deal.ID.String(),
		Kind:      ObjectTypeDeal,
		CreatedAt: deal.CreatedAt,
		UpdatedAt: deal.UpdatedAt,
		CreatedBy: "system",
		ACL:       `[{"actorId":"system","role":"owner"}]`,
		Tags:      "[]",
		Fields:    fields,
	}
}

// ObjectToDeal converts an Office OS Object to a models.Deal.
func ObjectToDeal(obj *Object) (*models.Deal, error) {
	if obj.Kind != ObjectTypeDeal {
		return nil, fmt.Errorf("object is not a Deal type: %s", obj.Kind)
	}

	id, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid deal ID: %w", err)
	}

	deal := &models.Deal{
		ID:        id,
		CreatedAt: obj.CreatedAt,
		UpdatedAt: obj.UpdatedAt,
		Currency:  "USD", // Default
	}

	// Extract title from fields
	if title, ok := obj.Fields["title"].(string); ok {
		deal.Title = title
	}

	// Extract required company_id
	companyIDStr, ok := obj.Fields["company_id"].(string)
	if !ok || companyIDStr == "" {
		return nil, fmt.Errorf("deal missing required company_id")
	}
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid company_id: %w", err)
	}
	deal.CompanyID = companyID

	// Extract numeric fields - handle both float64 (from JSON) and int64
	if amount, ok := obj.Fields["amount"].(float64); ok {
		deal.Amount = int64(amount)
	} else if amount, ok := obj.Fields["amount"].(int64); ok {
		deal.Amount = amount
	}

	// Extract string fields
	if currency, ok := obj.Fields["currency"].(string); ok {
		deal.Currency = currency
	}
	if stage, ok := obj.Fields["stage"].(string); ok {
		deal.Stage = stage
	}

	// Parse optional contact_id
	if contactIDStr, ok := obj.Fields["contact_id"].(string); ok && contactIDStr != "" {
		contactID, err := uuid.Parse(contactIDStr)
		if err == nil {
			deal.ContactID = &contactID
		}
	}

	// Parse optional expected_close_date
	if closeDateStr, ok := obj.Fields["expected_close_date"].(string); ok && closeDateStr != "" {
		closeDate, err := time.Parse(time.RFC3339Nano, closeDateStr)
		if err == nil {
			deal.ExpectedCloseDate = &closeDate
		}
	}

	// Parse last_activity_at (required for deals)
	if activityStr, ok := obj.Fields["last_activity_at"].(string); ok && activityStr != "" {
		activity, err := time.Parse(time.RFC3339Nano, activityStr)
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
