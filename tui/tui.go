// ABOUTME: Terminal User Interface using bubbletea framework
// ABOUTME: Provides interactive full-screen interface for CRM operations
package tui

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/textinput"
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
	ViewConfirmDelete
)

// EntityType represents the type of entity being viewed
type EntityType int

const (
	EntityContacts EntityType = iota
	EntityCompanies
	EntityDeals
	EntityFollowups
)

// Model is the main bubbletea model
type Model struct {
	db         *sql.DB
	viewMode   ViewMode
	entityType EntityType

	// List view state
	selectedRow int    //nolint:unused // will be used in Task 4.2
	searchQuery string //nolint:unused // will be used in Task 4.2

	// Detail view state
	selectedID string //nolint:unused // will be used in Task 4.3

	// Edit view state
	formInputs []textinput.Model //nolint:unused // will be used in Task 4.4
	focusIndex int               //nolint:unused // will be used in Task 4.4

	// Graph view state
	graphDOT string //nolint:unused // will be used in Task 4.5

	// Delete confirmation state
	deleteConfirmed bool
	deleteMessage   string

	// UI state
	width  int
	height int
	err    error //nolint:unused // will be used in error handling
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
	case ViewConfirmDelete:
		return m.renderConfirmDeleteView()
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
	case ViewConfirmDelete:
		return m.handleConfirmDeleteKeys(msg)
	}

	return m, nil
}

// Styles
var (
	titleStyle = lipgloss.NewStyle(). //nolint:unused // will be used in Task 4.2-4.5
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	tabActiveStyle = lipgloss.NewStyle(). //nolint:unused // will be used in Task 4.2
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Background(lipgloss.Color("235")).
			Padding(0, 2)

	tabInactiveStyle = lipgloss.NewStyle(). //nolint:unused // will be used in Task 4.2
				Foreground(lipgloss.Color("240")).
				Padding(0, 2)

	helpStyle = lipgloss.NewStyle(). //nolint:unused // will be used in Task 4.2-4.5
			Foreground(lipgloss.Color("240")).
			MarginTop(1)
)
