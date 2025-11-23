# Phase 4: Interactive TUI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add full-screen interactive terminal interface using bubbletea with list/detail/edit views and keyboard navigation

**Architecture:** `pagen` command launches TUI with tabs for Contacts/Companies/Deals, keyboard navigation, and CRUD operations

**Tech Stack:** bubbletea (TUI framework), lipgloss (styling), Go stdlib

---

## Task 4.1: Setup bubbletea dependencies and base model

**Files:**
- Modify: `go.mod` (add dependencies)
- Create: `tui/tui.go`

**Step 1: Add bubbletea dependencies**

Run:
```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles/table@latest
go get github.com/charmbracelet/bubbles/textinput@latest
```

Expected: Dependencies added to go.mod

**Step 2: Create TUI package with base model**

Create `tui/tui.go`:

```go
// ABOUTME: Terminal User Interface using bubbletea framework
// ABOUTME: Provides interactive full-screen interface for CRM operations
package tui

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the current TUI view
type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
	ViewEdit
	ViewGraph
)

// EntityType represents the type of entity being viewed
type EntityType int

const (
	EntityContacts EntityType = iota
	EntityCompanies
	EntityDeals
)

// Model is the main bubbletea model
type Model struct {
	db         *sql.DB
	viewMode   ViewMode
	entityType EntityType

	// List view state
	selectedRow int
	searchQuery string

	// Detail view state
	selectedID string

	// Edit view state
	formInputs []textinput.Model
	focusIndex int

	// Graph view state
	graphDOT string

	// UI state
	width  int
	height int
	err    error
}

// NewModel creates a new TUI model
func NewModel(db *sql.DB) Model {
	return Model{
		db:         db,
		viewMode:   ViewList,
		entityType: EntityContacts,
		width:      80,
		height:     24,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	switch m.viewMode {
	case ViewList:
		return m.renderListView()
	case ViewDetail:
		return m.renderDetailView()
	case ViewEdit:
		return m.renderEditView()
	case ViewGraph:
		return m.renderGraphView()
	}
	return ""
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		// TODO: Show help overlay
		return m, nil
	}

	// Delegate to view-specific handlers
	switch m.viewMode {
	case ViewList:
		return m.handleListKeys(msg)
	case ViewDetail:
		return m.handleDetailKeys(msg)
	case ViewEdit:
		return m.handleEditKeys(msg)
	case ViewGraph:
		return m.handleGraphKeys(msg)
	}

	return m, nil
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Background(lipgloss.Color("235")).
			Padding(0, 2)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)
)
```

**Step 3: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 4: Commit**

```bash
git add go.mod go.sum tui/tui.go
git commit -m "feat: add bubbletea TUI base model"
```

---

## Task 4.2: Implement list view with tabs

**Files:**
- Create: `tui/list_view.go`
- Modify: `tui/tui.go` (import list view)

**Step 1: Create list view implementation**

Create `tui/list_view.go`:

```go
package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/harperreed/pagen/db"
	"github.com/harperreed/pagen/models"
)

func (m Model) renderListView() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("PAGEN CRM"))
	s.WriteString("\n\n")

	// Tabs
	s.WriteString(m.renderTabs())
	s.WriteString("\n\n")

	// Table
	s.WriteString(m.renderTable())
	s.WriteString("\n\n")

	// Help
	s.WriteString(m.renderListHelp())

	return s.String()
}

func (m Model) renderTabs() string {
	tabs := []string{"Contacts", "Companies", "Deals"}
	var rendered []string

	for i, tab := range tabs {
		if EntityType(i) == m.entityType {
			rendered = append(rendered, tabActiveStyle.Render(tab))
		} else {
			rendered = append(rendered, tabInactiveStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m Model) renderTable() string {
	switch m.entityType {
	case EntityContacts:
		return m.renderContactsTable()
	case EntityCompanies:
		return m.renderCompaniesTable()
	case EntityDeals:
		return m.renderDealsTable()
	}
	return ""
}

func (m Model) renderContactsTable() string {
	contacts, err := db.FindContacts(m.db, m.searchQuery, nil, 100)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Email", Width: 30},
		{Title: "Company", Width: 20},
	}

	var rows []table.Row
	for _, contact := range contacts {
		companyName := ""
		if contact.CompanyID != nil {
			company, _ := db.GetCompany(m.db, *contact.CompanyID)
			if company != nil {
				companyName = company.Name
			}
		}

		rows = append(rows, table.Row{
			contact.Name,
			contact.Email,
			companyName,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height - 10),
	)

	// Set selected row
	if m.selectedRow < len(rows) {
		t.SetCursor(m.selectedRow)
	}

	return t.View()
}

func (m Model) renderCompaniesTable() string {
	companies, err := db.FindCompanies(m.db, m.searchQuery, 100)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Domain", Width: 30},
		{Title: "Industry", Width: 20},
	}

	var rows []table.Row
	for _, company := range companies {
		rows = append(rows, table.Row{
			company.Name,
			company.Domain,
			company.Industry,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height - 10),
	)

	if m.selectedRow < len(rows) {
		t.SetCursor(m.selectedRow)
	}

	return t.View()
}

func (m Model) renderDealsTable() string {
	deals, err := db.FindDeals(m.db, "", nil, 100)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	columns := []table.Column{
		{Title: "Title", Width: 30},
		{Title: "Company", Width: 25},
		{Title: "Stage", Width: 15},
		{Title: "Amount", Width: 10},
	}

	var rows []table.Row
	for _, deal := range deals {
		company, _ := db.GetCompany(m.db, deal.CompanyID)
		companyName := ""
		if company != nil {
			companyName = company.Name
		}

		amountStr := fmt.Sprintf("$%dK", deal.Amount/100000)

		rows = append(rows, table.Row{
			deal.Title,
			companyName,
			deal.Stage,
			amountStr,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height - 10),
	)

	if m.selectedRow < len(rows) {
		t.SetCursor(m.selectedRow)
	}

	return t.View()
}

func (m Model) renderListHelp() string {
	help := []string{
		"↑/↓: Navigate",
		"Tab: Switch tabs",
		"Enter: View details",
		"/: Search",
		"n: New",
		"q: Quit",
	}
	return helpStyle.Render(strings.Join(help, " • "))
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedRow > 0 {
			m.selectedRow--
		}
	case "down", "j":
		m.selectedRow++
	case "tab":
		m.entityType = (m.entityType + 1) % 3
		m.selectedRow = 0
	case "enter":
		// Switch to detail view
		m.viewMode = ViewDetail
		m.selectedID = m.getSelectedID()
	case "/":
		// TODO: Enter search mode
	case "n":
		// Switch to edit view (new)
		m.viewMode = ViewEdit
		m.selectedID = ""
	}

	return m, nil
}

func (m Model) getSelectedID() string {
	switch m.entityType {
	case EntityContacts:
		contacts, _ := db.FindContacts(m.db, m.searchQuery, nil, 100)
		if m.selectedRow < len(contacts) {
			return contacts[m.selectedRow].ID.String()
		}
	case EntityCompanies:
		companies, _ := db.FindCompanies(m.db, m.searchQuery, 100)
		if m.selectedRow < len(companies) {
			return companies[m.selectedRow].ID.String()
		}
	case EntityDeals:
		deals, _ := db.FindDeals(m.db, "", nil, 100)
		if m.selectedRow < len(deals) {
			return deals[m.selectedRow].ID.String()
		}
	}
	return ""
}
```

**Step 2: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 3: Commit**

```bash
git add tui/list_view.go
git commit -m "feat: add TUI list view with tabs and tables"
```

---

## Task 4.3: Implement detail view

**Files:**
- Create: `tui/detail_view.go`

**Step 1: Create detail view implementation**

Create `tui/detail_view.go`:

```go
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/harperreed/pagen/db"
)

var (
	fieldLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Width(20)

	fieldValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)

func (m Model) renderDetailView() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("DETAIL VIEW"))
	s.WriteString("\n\n")

	// Entity details
	switch m.entityType {
	case EntityContacts:
		s.WriteString(m.renderContactDetail())
	case EntityCompanies:
		s.WriteString(m.renderCompanyDetail())
	case EntityDeals:
		s.WriteString(m.renderDealDetail())
	}

	s.WriteString("\n\n")

	// Help
	s.WriteString(m.renderDetailHelp())

	return s.String()
}

func (m Model) renderContactDetail() string {
	id, err := uuid.Parse(m.selectedID)
	if err != nil {
		return fmt.Sprintf("Error: invalid ID: %v", err)
	}

	contact, err := db.GetContact(m.db, id)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	var s strings.Builder

	s.WriteString(m.renderField("Name", contact.Name))
	s.WriteString(m.renderField("Email", contact.Email))
	s.WriteString(m.renderField("Phone", contact.Phone))

	if contact.CompanyID != nil {
		company, _ := db.GetCompany(m.db, *contact.CompanyID)
		if company != nil {
			s.WriteString(m.renderField("Company", company.Name))
		}
	}

	if contact.LastContactedAt != nil {
		s.WriteString(m.renderField("Last Contacted", contact.LastContactedAt.Format("2006-01-02")))
	}

	s.WriteString(m.renderField("Notes", contact.Notes))

	// Related entities
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("RELATIONSHIPS"))
	s.WriteString("\n")

	relationships, _ := db.FindContactRelationships(m.db, id.String(), "")
	for _, rel := range relationships {
		s.WriteString(fmt.Sprintf("  • %s (%s)\n", rel.Context, rel.RelationshipType))
	}

	return s.String()
}

func (m Model) renderCompanyDetail() string {
	id, err := uuid.Parse(m.selectedID)
	if err != nil {
		return fmt.Sprintf("Error: invalid ID: %v", err)
	}

	company, err := db.GetCompany(m.db, id)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	var s strings.Builder

	s.WriteString(m.renderField("Name", company.Name))
	s.WriteString(m.renderField("Domain", company.Domain))
	s.WriteString(m.renderField("Industry", company.Industry))
	s.WriteString(m.renderField("Notes", company.Notes))

	// Contacts at company
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("CONTACTS"))
	s.WriteString("\n")

	contacts, _ := db.FindContacts(m.db, "", &id, 100)
	for _, contact := range contacts {
		s.WriteString(fmt.Sprintf("  • %s (%s)\n", contact.Name, contact.Email))
	}

	return s.String()
}

func (m Model) renderDealDetail() string {
	id, err := uuid.Parse(m.selectedID)
	if err != nil {
		return fmt.Sprintf("Error: invalid ID: %v", err)
	}

	deal, err := db.GetDeal(m.db, id)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	var s strings.Builder

	s.WriteString(m.renderField("Title", deal.Title))

	company, _ := db.GetCompany(m.db, deal.CompanyID)
	if company != nil {
		s.WriteString(m.renderField("Company", company.Name))
	}

	if deal.ContactID != nil {
		contact, _ := db.GetContact(m.db, *deal.ContactID)
		if contact != nil {
			s.WriteString(m.renderField("Contact", contact.Name))
		}
	}

	s.WriteString(m.renderField("Stage", deal.Stage))
	s.WriteString(m.renderField("Amount", fmt.Sprintf("$%d %s", deal.Amount/100, deal.Currency)))

	if deal.ExpectedCloseDate != nil {
		s.WriteString(m.renderField("Expected Close", deal.ExpectedCloseDate.Format("2006-01-02")))
	}

	// Notes
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("NOTES"))
	s.WriteString("\n")

	notes, _ := db.GetDealNotes(m.db, id)
	for _, note := range notes {
		s.WriteString(fmt.Sprintf("  • [%s] %s\n", note.CreatedAt.Format("2006-01-02"), note.Content))
	}

	return s.String()
}

func (m Model) renderField(label, value string) string {
	if value == "" {
		value = "-"
	}
	return fmt.Sprintf("%s %s\n",
		fieldLabelStyle.Render(label+":"),
		fieldValueStyle.Render(value))
}

func (m Model) renderDetailHelp() string {
	help := []string{
		"Esc: Back",
		"e: Edit",
		"d: Delete",
		"g: View graph",
		"q: Quit",
	}
	return helpStyle.Render(strings.Join(help, " • "))
}

func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewList
	case "e":
		m.viewMode = ViewEdit
	case "d":
		// TODO: Show delete confirmation
	case "g":
		m.viewMode = ViewGraph
		// TODO: Generate graph DOT
	}

	return m, nil
}
```

**Step 2: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 3: Commit**

```bash
git add tui/detail_view.go
git commit -m "feat: add TUI detail view with related entities"
```

---

## Task 4.4: Implement edit view with forms

**Files:**
- Create: `tui/edit_view.go`

**Step 1: Create edit view implementation**

Create `tui/edit_view.go`:

```go
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/harperreed/pagen/db"
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
	if m.selectedID == "" {
		// Create new
		return db.AddContact(m.db,
			m.formInputs[0].Value(), // name
			m.formInputs[1].Value(), // email
			m.formInputs[2].Value(), // phone
			m.formInputs[3].Value(), // company
			m.formInputs[4].Value()) // notes
	} else {
		// Update existing
		id, _ := uuid.Parse(m.selectedID)
		return db.UpdateContact(m.db, id,
			m.formInputs[0].Value(), // name
			m.formInputs[1].Value(), // email
			m.formInputs[2].Value(), // phone
			m.formInputs[3].Value(), // company
			m.formInputs[4].Value()) // notes
	}
}

func (m Model) saveCompany() error {
	if m.selectedID == "" {
		// Create new
		return db.AddCompany(m.db,
			m.formInputs[0].Value(), // name
			m.formInputs[1].Value(), // domain
			m.formInputs[2].Value(), // industry
			m.formInputs[3].Value()) // notes
	} else {
		// Update existing
		id, _ := uuid.Parse(m.selectedID)
		return db.UpdateCompany(m.db, id,
			m.formInputs[0].Value(), // name
			m.formInputs[1].Value(), // domain
			m.formInputs[2].Value(), // industry
			m.formInputs[3].Value()) // notes
	}
}

func (m Model) saveDeal() error {
	// This is simplified - real implementation needs amount parsing, etc.
	// For now, just return error for TUI deals
	return fmt.Errorf("deal creation/editing in TUI not yet implemented")
}
```

**Step 2: Update tui.go to initialize forms**

Add to `tui/tui.go` in the `handleListKeys` and `handleDetailKeys` functions:

```go
// In handleListKeys, when "n" is pressed:
case "n":
	m.viewMode = ViewEdit
	m.selectedID = ""
	m.initFormInputs()

// In handleDetailKeys, when "e" is pressed:
case "e":
	m.viewMode = ViewEdit
	m.initFormInputs()
```

**Step 3: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 4: Commit**

```bash
git add tui/edit_view.go tui/tui.go
git commit -m "feat: add TUI edit view with forms"
```

---

## Task 4.5: Implement graph view

**Files:**
- Create: `tui/graph_view.go`

**Step 1: Create graph view implementation**

Create `tui/graph_view.go`:

```go
package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"

	"github.com/harperreed/pagen/viz"
)

func (m Model) renderGraphView() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("GRAPH VIEW"))
	s.WriteString("\n\n")

	// DOT source (scrollable in future)
	if m.graphDOT == "" {
		s.WriteString("Generating graph...\n")
	} else {
		s.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Render(m.graphDOT))
	}

	s.WriteString("\n\n")

	// Help
	s.WriteString(m.renderGraphHelp())

	return s.String()
}

func (m Model) renderGraphHelp() string {
	help := []string{
		"Esc: Back",
		"q: Quit",
	}
	return helpStyle.Render(strings.Join(help, " • "))
}

func (m Model) handleGraphKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.viewMode = ViewDetail
		m.graphDOT = ""
	}

	return m, nil
}

func (m *Model) generateGraph() error {
	generator := viz.NewGraphGenerator(m.db)
	ctx := context.Background()

	var dot string
	var err error

	switch m.entityType {
	case EntityContacts:
		id, _ := uuid.Parse(m.selectedID)
		dot, err = generator.GenerateContactGraph(ctx, &id)
	case EntityCompanies:
		id, _ := uuid.Parse(m.selectedID)
		dot, err = generator.GenerateCompanyGraph(ctx, &id)
	case EntityDeals:
		dot, err = generator.GeneratePipelineGraph(ctx)
	}

	if err != nil {
		return err
	}

	m.graphDOT = dot
	return nil
}
```

**Step 2: Update detail_view.go to generate graph**

Modify `tui/detail_view.go`, in `handleDetailKeys`:

```go
case "g":
	m.viewMode = ViewGraph
	err := m.generateGraph()
	if err != nil {
		m.err = err
		m.viewMode = ViewDetail
	}
```

**Step 3: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 4: Commit**

```bash
git add tui/graph_view.go tui/detail_view.go
git commit -m "feat: add TUI graph view with DOT visualization"
```

---

## Task 4.6: Wire up main command and polish

**Files:**
- Modify: `main.go` (add TUI command)
- Create: `.scratch/test_tui_manual.sh`

**Step 1: Add TUI command to main.go**

Modify `main.go`, add TUI handling before the existing command parsing:

```go
func main() {
	// ... existing flag parsing ...

	database, err := db.InitDB(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// If no command specified, launch TUI
	if len(flag.Args()) == 0 {
		tuiModel := tui.NewModel(database)
		p := tea.NewProgram(tuiModel, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatalf("TUI error: %v", err)
		}
		return
	}

	// Otherwise handle CLI commands
	args := flag.Args()
	command := args[0]

	// ... existing command handling ...
}
```

**Step 2: Add missing import**

Add to imports in `main.go`:

```go
import (
	// ... existing imports ...
	tea "github.com/charmbracelet/bubbletea"
	"github.com/harperreed/pagen/tui"
)
```

**Step 3: Create manual test script**

Create `.scratch/test_tui_manual.sh`:

```bash
#!/bin/bash
set -e

echo "=== TUI Manual Test Instructions ==="
echo ""
echo "This script creates test data. Then launch the TUI manually."
echo ""

export DB=/tmp/test_tui_$$.db

# Create test data
./pagen --db-path $DB crm add-company --name "Acme Corp"
./pagen --db-path $DB crm add-company --name "TechStart Inc"
./pagen --db-path $DB crm add-contact --name "Alice" --email "alice@acme.com" --company "Acme Corp"
./pagen --db-path $DB crm add-contact --name "Bob" --email "bob@techstart.com" --company "TechStart Inc"
./pagen --db-path $DB crm add-deal --title "Enterprise Deal" --company "Acme Corp" --contact "Alice" --amount 500000 --stage "negotiation"
./pagen --db-path $DB crm add-deal --title "Startup Deal" --company "TechStart Inc" --amount 50000 --stage "prospecting"

echo ""
echo "Test data created in: $DB"
echo ""
echo "To launch TUI, run:"
echo "  ./pagen --db-path $DB"
echo ""
echo "Test checklist:"
echo "  [ ] Tab switches between Contacts/Companies/Deals"
echo "  [ ] Arrow keys navigate rows"
echo "  [ ] Enter shows detail view"
echo "  [ ] Esc returns to list view"
echo "  [ ] 'e' in detail view shows edit form"
echo "  [ ] 'g' in detail view shows graph DOT"
echo "  [ ] 'n' in list view shows new entity form"
echo "  [ ] 'q' quits application"
echo ""
echo "Cleanup: rm $DB"
```

**Step 4: Make test script executable**

Run: `chmod +x .scratch/test_tui_manual.sh`

**Step 5: Build and test**

Run: `make build`
Expected: Compiles successfully

Run: `.scratch/test_tui_manual.sh`
Expected: Creates test data and prints instructions

**Step 6: Commit**

```bash
git add main.go .scratch/test_tui_manual.sh
git commit -m "feat: wire up TUI as default command"
```

---

## Success Criteria

- [ ] `pagen` launches interactive TUI
- [ ] Tab key switches between Contacts/Companies/Deals tabs
- [ ] Arrow keys navigate table rows
- [ ] Enter shows detail view with related entities
- [ ] Esc returns to previous view
- [ ] 'e' key opens edit form with populated fields
- [ ] 'n' key opens new entity form
- [ ] 'g' key shows GraphViz DOT output
- [ ] Forms save successfully on Enter
- [ ] 'q' quits the application
- [ ] All entity types work in all views
