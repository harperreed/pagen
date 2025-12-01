// ABOUTME: Contact deduplication and matching logic
// ABOUTME: Finds existing contacts by email to prevent duplicates during sync
package sync

import (
	"strings"

	"github.com/harperreed/pagen/models"
)

type ContactMatcher struct {
	byEmail map[string]*models.Contact
}

// NewContactMatcher creates a matcher from existing contacts.
func NewContactMatcher(contacts []models.Contact) *ContactMatcher {
	m := &ContactMatcher{
		byEmail: make(map[string]*models.Contact),
	}

	for i := range contacts {
		email := normalizeEmail(contacts[i].Email)
		if email != "" {
			m.byEmail[email] = &contacts[i]
		}
	}

	return m
}

// FindMatch looks for existing contact by email.
func (m *ContactMatcher) FindMatch(email, name string) (*models.Contact, bool) {
	normalized := normalizeEmail(email)
	if normalized == "" {
		return nil, false
	}

	contact, found := m.byEmail[normalized]
	return contact, found
}

// AddContact adds a newly created contact to the matcher to prevent duplicates
// within the same import session.
func (m *ContactMatcher) AddContact(contact *models.Contact) {
	email := normalizeEmail(contact.Email)
	if email != "" {
		m.byEmail[email] = contact
	}
}

// normalizeEmail converts email to lowercase for comparison.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// extractDomain extracts domain from email address.
func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
