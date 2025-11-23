package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/harperreed/pagen/db"
	"github.com/harperreed/pagen/models"
)

func (m Model) renderEditView() string {
	var s strings.Builder

	// Title
	if m.selectedID == "" {
		s.WriteString(titleStyle.Render("NEW " + m.entityTypeName()))
	} else {
		s.WriteString(titleStyle.Render("EDIT " + m.entityTypeName()))
	}
	s.WriteString("\n\n")

	// Form fields
	for i, input := range m.formInputs {
		if i == m.focusIndex {
			s.WriteString("> ")
		} else {
			s.WriteString("  ")
		}
		s.WriteString(input.View())
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Help
	s.WriteString(m.renderEditHelp())

	return s.String()
}

func (m Model) entityTypeName() string {
	switch m.entityType {
	case EntityContacts:
		return "CONTACT"
	case EntityCompanies:
		return "COMPANY"
	case EntityDeals:
		return "DEAL"
	}
	return ""
}

func (m Model) renderEditHelp() string {
	help := []string{
		"Tab: Next field",
		"Enter: Save",
		"Esc: Cancel",
	}
	return helpStyle.Render(strings.Join(help, " • "))
}

func (m Model) handleEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewList
		return m, nil
	case "tab":
		m.focusIndex = (m.focusIndex + 1) % len(m.formInputs)
		m.updateFormFocus()
		return m, nil
	case "enter":
		// Save the entity
		err := m.saveEntity()
		if err != nil {
			m.err = err
		} else {
			m.viewMode = ViewList
		}
		return m, nil
	}

	// Update current input
	var cmd tea.Cmd
	m.formInputs[m.focusIndex], cmd = m.formInputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *Model) initFormInputs() {
	switch m.entityType {
	case EntityContacts:
		m.initContactForm()
	case EntityCompanies:
		m.initCompanyForm()
	case EntityDeals:
		m.initDealForm()
	}

	m.focusIndex = 0
	m.updateFormFocus()
}

func (m *Model) initContactForm() {
	inputs := make([]textinput.Model, 5)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Name"
	inputs[0].CharLimit = 100

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Email"
	inputs[1].CharLimit = 100

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Phone"
	inputs[2].CharLimit = 20

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Company Name"
	inputs[3].CharLimit = 100

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Notes"
	inputs[4].CharLimit = 500

	// If editing, populate fields
	if m.selectedID != "" {
		id, _ := uuid.Parse(m.selectedID)
		contact, _ := db.GetContact(m.db, id)
		if contact != nil {
			inputs[0].SetValue(contact.Name)
			inputs[1].SetValue(contact.Email)
			inputs[2].SetValue(contact.Phone)
			inputs[4].SetValue(contact.Notes)

			if contact.CompanyID != nil {
				company, _ := db.GetCompany(m.db, *contact.CompanyID)
				if company != nil {
					inputs[3].SetValue(company.Name)
				}
			}
		}
	}

	m.formInputs = inputs
}

func (m *Model) initCompanyForm() {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Name"
	inputs[0].CharLimit = 100

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Domain"
	inputs[1].CharLimit = 100

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Industry"
	inputs[2].CharLimit = 100

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Notes"
	inputs[3].CharLimit = 500

	// If editing, populate fields
	if m.selectedID != "" {
		id, _ := uuid.Parse(m.selectedID)
		company, _ := db.GetCompany(m.db, id)
		if company != nil {
			inputs[0].SetValue(company.Name)
			inputs[1].SetValue(company.Domain)
			inputs[2].SetValue(company.Industry)
			inputs[3].SetValue(company.Notes)
		}
	}

	m.formInputs = inputs
}

func (m *Model) initDealForm() {
	inputs := make([]textinput.Model, 6)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Title"
	inputs[0].CharLimit = 100

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Company Name"
	inputs[1].CharLimit = 100

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Contact Name (optional)"
	inputs[2].CharLimit = 100

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Stage (prospecting/qualification/proposal/negotiation/closed_won/closed_lost)"
	inputs[3].CharLimit = 50

	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Amount (in cents)"
	inputs[4].CharLimit = 20

	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Currency (default: USD)"
	inputs[5].CharLimit = 3

	// If editing, populate fields
	if m.selectedID != "" {
		id, _ := uuid.Parse(m.selectedID)
		deal, _ := db.GetDeal(m.db, id)
		if deal != nil {
			inputs[0].SetValue(deal.Title)
			inputs[3].SetValue(deal.Stage)
			inputs[4].SetValue(fmt.Sprintf("%d", deal.Amount))
			inputs[5].SetValue(deal.Currency)

			company, _ := db.GetCompany(m.db, deal.CompanyID)
			if company != nil {
				inputs[1].SetValue(company.Name)
			}

			if deal.ContactID != nil {
				contact, _ := db.GetContact(m.db, *deal.ContactID)
				if contact != nil {
					inputs[2].SetValue(contact.Name)
				}
			}
		}
	}

	m.formInputs = inputs
}

func (m *Model) updateFormFocus() {
	for i := range m.formInputs {
		if i == m.focusIndex {
			m.formInputs[i].Focus()
		} else {
			m.formInputs[i].Blur()
		}
	}
}

func (m Model) saveEntity() error {
	switch m.entityType {
	case EntityContacts:
		return m.saveContact()
	case EntityCompanies:
		return m.saveCompany()
	case EntityDeals:
		return m.saveDeal()
	}
	return nil
}

func (m Model) saveContact() error {
	contact := &models.Contact{
		Name:  m.formInputs[0].Value(),
		Email: m.formInputs[1].Value(),
		Phone: m.formInputs[2].Value(),
		Notes: m.formInputs[4].Value(),
	}

	// Handle company lookup/creation if company_name provided
	if m.formInputs[3].Value() != "" {
		companyName := m.formInputs[3].Value()
		company, err := db.FindCompanyByName(m.db, companyName)
		if err != nil {
			return fmt.Errorf("failed to lookup company: %w", err)
		}

		if company == nil {
			// Create new company
			company = &models.Company{
				Name: companyName,
			}
			if err := db.CreateCompany(m.db, company); err != nil {
				return fmt.Errorf("failed to create company: %w", err)
			}
		}

		contact.CompanyID = &company.ID
	}

	if m.selectedID == "" {
		// Create new
		return db.CreateContact(m.db, contact)
	} else {
		// Update existing
		id, _ := uuid.Parse(m.selectedID)
		return db.UpdateContact(m.db, id, contact)
	}
}

func (m Model) saveCompany() error {
	company := &models.Company{
		Name:     m.formInputs[0].Value(),
		Domain:   m.formInputs[1].Value(),
		Industry: m.formInputs[2].Value(),
		Notes:    m.formInputs[3].Value(),
	}

	if m.selectedID == "" {
		// Create new
		return db.CreateCompany(m.db, company)
	} else {
		// Update existing
		id, _ := uuid.Parse(m.selectedID)
		return db.UpdateCompany(m.db, id, company)
	}
}

func (m Model) saveDeal() error {
	// This is simplified - real implementation needs amount parsing, etc.
	// For now, just return error for TUI deals
	return fmt.Errorf("deal creation/editing in TUI not yet implemented")
}
