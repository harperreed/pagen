// ABOUTME: Google Contacts API importer
// ABOUTME: Fetches and imports contacts from Google People API with deduplication
package sync

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/harperreed/pagen/db"
	"github.com/harperreed/pagen/models"
	"google.golang.org/api/people/v1"
)

type ContactsImporter struct {
	db      *sql.DB
	matcher *ContactMatcher
}

type GoogleContact struct {
	ResourceName string
	Name         string
	Email        string
	Phone        string
	Company      string
	JobTitle     string
	Notes        string
}

func NewContactsImporter(database *sql.DB) *ContactsImporter {
	return &ContactsImporter{
		db: database,
	}
}

// ImportContact imports a single contact from Google.
func (ci *ContactsImporter) ImportContact(gc *GoogleContact) (bool, error) {
	// Check for existing contact
	existing, found := ci.matcher.FindMatch(gc.Email, gc.Name)
	if found {
		// Update existing contact if needed
		_, err := ci.updateContact(existing, gc)
		if err != nil {
			return false, err
		}

		// Log sync for existing contact
		if err := ci.logSync(gc.ResourceName, existing.ID); err != nil {
			return false, fmt.Errorf("failed to log sync: %w", err)
		}

		return false, nil
	}

	// Create new contact
	contact := &models.Contact{
		Name:  gc.Name,
		Email: gc.Email,
		Phone: gc.Phone,
		Notes: gc.Notes,
	}

	// Handle company
	if gc.Company != "" {
		company, err := ci.findOrCreateCompany(gc.Company)
		if err != nil {
			return false, fmt.Errorf("failed to handle company: %w", err)
		}
		contact.CompanyID = &company.ID
	}

	// Create contact
	if err := db.CreateContact(ci.db, contact); err != nil {
		return false, fmt.Errorf("failed to create contact: %w", err)
	}

	// Log sync
	if err := ci.logSync(gc.ResourceName, contact.ID); err != nil {
		return false, fmt.Errorf("failed to log sync: %w", err)
	}

	// Add to matcher to prevent duplicates within the same import session
	ci.matcher.AddContact(contact)

	return true, nil
}

func (ci *ContactsImporter) updateContact(existing *models.Contact, gc *GoogleContact) (bool, error) {
	// Load fresh copy from database to avoid working with stale cache data
	freshContact, err := db.GetContact(ci.db, existing.ID)
	if err != nil {
		return false, fmt.Errorf("failed to load contact: %w", err)
	}

	// Only update if Google data is more complete
	updated := false

	if gc.Phone != "" && freshContact.Phone == "" {
		freshContact.Phone = gc.Phone
		updated = true
	}

	if gc.Notes != "" && freshContact.Notes == "" {
		freshContact.Notes = gc.Notes
		updated = true
	}

	// Update company if contact doesn't have one
	if gc.Company != "" && freshContact.CompanyID == nil {
		company, err := ci.findOrCreateCompany(gc.Company)
		if err != nil {
			return false, fmt.Errorf("failed to handle company: %w", err)
		}
		freshContact.CompanyID = &company.ID
		updated = true
	}

	if !updated {
		return false, nil
	}

	if err := db.UpdateContact(ci.db, freshContact.ID, freshContact); err != nil {
		return false, err
	}

	// Update the matcher's cache with the new data
	ci.matcher.AddContact(freshContact)

	return true, nil
}

func (ci *ContactsImporter) findOrCreateCompany(name string) (*models.Company, error) {
	// Try to find existing company
	company, err := db.FindCompanyByName(ci.db, name)
	if err != nil {
		return nil, err
	}

	if company != nil {
		return company, nil
	}

	// Create new company
	newCompany := &models.Company{
		Name: name,
	}

	// Use INSERT OR IGNORE to handle race condition where multiple contacts
	// with the same company are processed in the same batch
	if err := db.CreateCompany(ci.db, newCompany); err != nil {
		// If creation failed due to uniqueness constraint, try finding it again
		// (another concurrent import may have created it)
		company, findErr := db.FindCompanyByName(ci.db, name)
		if findErr != nil {
			return nil, fmt.Errorf("failed to create or find company: %w (original error: %w)", findErr, err)
		}
		if company != nil {
			return company, nil
		}
		return nil, err
	}

	return newCompany, nil
}

func (ci *ContactsImporter) logSync(sourceID string, entityID uuid.UUID) error {
	syncLog := &models.SyncLog{
		ID:            uuid.New(),
		SourceService: "contacts",
		SourceID:      sourceID,
		EntityType:    "contact",
		EntityID:      entityID,
	}

	query := `
		INSERT OR IGNORE INTO sync_log (id, source_service, source_id, entity_type, entity_id)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := ci.db.Exec(query,
		syncLog.ID.String(),
		syncLog.SourceService,
		syncLog.SourceID,
		syncLog.EntityType,
		syncLog.EntityID.String(),
	)

	return err
}

// ImportContacts fetches and imports contacts from Google People API.
func ImportContacts(database *sql.DB, client *people.Service) error {
	const contactsService = "contacts"

	// Update sync state to 'syncing'
	fmt.Println("Syncing Google Contacts...")
	if err := db.UpdateSyncStatus(database, contactsService, "syncing", nil); err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	// Load all existing contacts for matching (do this ONCE, not per contact)
	allContacts, err := db.FindContacts(database, "", nil, 20000)
	if err != nil {
		errMsg := err.Error()
		_ = db.UpdateSyncStatus(database, contactsService, "error", &errMsg)
		return fmt.Errorf("failed to load existing contacts: %w", err)
	}

	// Create importer with pre-loaded matcher
	importer := NewContactsImporter(database)
	importer.matcher = NewContactMatcher(allContacts)

	// Fetch contacts with pagination
	totalFetched := 0
	totalProcessed := 0
	newContacts := 0
	updatedContacts := 0
	pageToken := ""

	for {
		// Build request
		call := client.People.Connections.List("people/me").
			PageSize(1000).
			PersonFields("names,emailAddresses,phoneNumbers,organizations,biographies")

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		// Fetch page
		response, err := call.Do()
		if err != nil {
			errMsg := fmt.Sprintf("failed to fetch contacts: %v", err)
			_ = db.UpdateSyncStatus(database, contactsService, "error", &errMsg)
			return fmt.Errorf("failed to fetch contacts: %w", err)
		}

		// Handle nil or empty response
		if response == nil || response.Connections == nil {
			break
		}

		// Count contacts in this page
		pageCount := len(response.Connections)
		totalFetched += pageCount

		// Process contacts
		for _, person := range response.Connections {
			// Convert People API person to GoogleContact
			gc := convertPerson(person)

			// Skip contacts without email or name (both are required)
			if gc.Email == "" || gc.Name == "" {
				continue
			}

			// Check if already synced
			exists, err := db.CheckSyncLogExists(database, contactsService, person.ResourceName)
			if err != nil {
				fmt.Printf("  ✗ Failed to check sync log for %q: %v\n", gc.Name, err)
				continue
			}

			if exists {
				// Already imported, skip
				continue
			}

			// Import contact
			isNew, err := importer.ImportContact(gc)
			if err != nil {
				fmt.Printf("  ✗ Failed to import contact %q: %v\n", gc.Name, err)
				continue
			}

			totalProcessed++
			if isNew {
				newContacts++
			} else {
				updatedContacts++
			}
		}

		// Check for next page
		pageToken = response.NextPageToken
		if pageToken == "" {
			break
		}

		// Show progress if we're processing contacts
		if totalProcessed > 0 {
			fmt.Printf("  → Processed %d new contacts so far...\n", totalProcessed)
		}
	}

	// Update sync state to 'idle' on success
	if err := db.UpdateSyncStatus(database, contactsService, "idle", nil); err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	// Print summary
	fmt.Printf("\n✓ Fetched %d contacts from Google\n", totalFetched)
	if totalProcessed == 0 {
		fmt.Println("  ✓ No new contacts to import (all up to date)")
	} else {
		fmt.Printf("  ✓ Processed %d new contacts\n", totalProcessed)
		if newContacts > 0 {
			fmt.Printf("  ✓ Created %d new contacts\n", newContacts)
		}
		if updatedContacts > 0 {
			fmt.Printf("  ✓ Updated %d existing contacts\n", updatedContacts)
		}
	}

	return nil
}

// convertPerson converts a People API Person to GoogleContact.
func convertPerson(person *people.Person) *GoogleContact {
	gc := &GoogleContact{
		ResourceName: person.ResourceName,
	}

	// Extract name
	if len(person.Names) > 0 && person.Names[0].DisplayName != "" {
		gc.Name = person.Names[0].DisplayName
	}

	// Extract email (prefer primary, otherwise first available)
	for _, email := range person.EmailAddresses {
		if email.Value != "" {
			// If we haven't set an email yet, use this one
			if gc.Email == "" {
				gc.Email = email.Value
			}
			// If this is the primary email, use it and stop looking
			if email.Metadata != nil && email.Metadata.Primary {
				gc.Email = email.Value
				break
			}
		}
	}

	// Extract phone (prefer primary, otherwise first available)
	for _, phone := range person.PhoneNumbers {
		if phone.Value != "" {
			// If we haven't set a phone yet, use this one
			if gc.Phone == "" {
				gc.Phone = phone.Value
			}
			// If this is the primary phone, use it and stop looking
			if phone.Metadata != nil && phone.Metadata.Primary {
				gc.Phone = phone.Value
				break
			}
		}
	}

	// Extract organization/company
	if len(person.Organizations) > 0 {
		org := person.Organizations[0]
		if org.Name != "" {
			gc.Company = org.Name
		}
		if org.Title != "" {
			gc.JobTitle = org.Title
		}
	}

	// Extract notes from biography
	if len(person.Biographies) > 0 && person.Biographies[0].Value != "" {
		gc.Notes = person.Biographies[0].Value
	}

	return gc
}
