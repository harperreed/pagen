// ABOUTME: Web UI server with embedded templates
// ABOUTME: Provides read-only dashboard at localhost:8080
package web

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harperreed/pagen/db"
	"github.com/harperreed/pagen/models"
	"github.com/harperreed/pagen/viz"
)

//go:embed templates/*
var templatesFS embed.FS

type Server struct {
	db        *sql.DB
	templates *template.Template
	generator *viz.GraphGenerator
}

func NewServer(database *sql.DB) (*Server, error) {
	// Helper functions for templates
	funcMap := template.FuncMap{
		"divide": func(a, b int64) int64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"multiply": func(a, b int64) int64 {
			return a * b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Server{
		db:        database,
		templates: tmpl,
		generator: viz.NewGraphGenerator(database),
	}, nil
}

func (s *Server) Start(port int) error {
	// Routes
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/contacts", s.handleContacts)
	http.HandleFunc("/companies", s.handleCompanies)
	http.HandleFunc("/deals", s.handleDeals)
	http.HandleFunc("/graphs", s.handleGraphs)
	http.HandleFunc("/followups", s.handleFollowups)

	// Partials for HTMX
	http.HandleFunc("/partials/contact-detail", s.handleContactDetail)
	http.HandleFunc("/partials/company-detail", s.handleCompanyDetail)
	http.HandleFunc("/partials/deal-detail", s.handleDealDetail)
	http.HandleFunc("/partials/graph", s.handleGraphPartial)
	http.HandleFunc("/followups/log/", s.handleFollowupLog)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting web server at http://localhost%s", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := viz.GenerateDashboardStats(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Stats":           stats,
		"Title":           "Dashboard",
		"ContentTemplate": "dashboard-content",
	}

	s.renderTemplate(w, "layout.html", data)
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	// Execute the specified template (usually layout.html)
	// The data map includes ContentTemplate to specify which content block to render
	err := s.templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("Template error rendering %s: %v", name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleContacts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	contacts, err := db.FindContacts(s.db, query, nil, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Enrich with company names
	type ContactView struct {
		ID          string
		Name        string
		Email       string
		CompanyName string
	}

	var contactViews []ContactView
	for _, contact := range contacts {
		companyName := ""
		if contact.CompanyID != nil {
			company, _ := db.GetCompany(s.db, *contact.CompanyID)
			if company != nil {
				companyName = company.Name
			}
		}

		contactViews = append(contactViews, ContactView{
			ID:          contact.ID.String(),
			Name:        contact.Name,
			Email:       contact.Email,
			CompanyName: companyName,
		})
	}

	data := map[string]interface{}{
		"Contacts":        contactViews,
		"Title":           "Contacts",
		"ContentTemplate": "contacts-content",
	}

	s.renderTemplate(w, "layout.html", data)
}

func (s *Server) handleCompanies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	companies, err := db.FindCompanies(s.db, query, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Companies":       companies,
		"Title":           "Companies",
		"ContentTemplate": "companies-content",
	}

	s.renderTemplate(w, "layout.html", data)
}

func (s *Server) handleDeals(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	stage := r.URL.Query().Get("stage")

	deals, err := db.FindDeals(s.db, query, nil, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter by stage if specified
	if stage != "" {
		var filtered []models.Deal
		for _, deal := range deals {
			if deal.Stage == stage {
				filtered = append(filtered, deal)
			}
		}
		deals = filtered
	}

	// Enrich with company names
	type DealView struct {
		ID          string
		Title       string
		CompanyName string
		Stage       string
		Amount      int64
		Currency    string
	}

	var dealViews []DealView
	for _, deal := range deals {
		company, _ := db.GetCompany(s.db, deal.CompanyID)
		companyName := ""
		if company != nil {
			companyName = company.Name
		}

		dealViews = append(dealViews, DealView{
			ID:          deal.ID.String(),
			Title:       deal.Title,
			CompanyName: companyName,
			Stage:       deal.Stage,
			Amount:      deal.Amount,
			Currency:    deal.Currency,
		})
	}

	data := map[string]interface{}{
		"Deals":           dealViews,
		"Title":           "Deals",
		"ContentTemplate": "deals-content",
	}

	s.renderTemplate(w, "layout.html", data)
}

func (s *Server) handleContactDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	contact, err := db.GetContact(s.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	companyName := ""
	if contact.CompanyID != nil {
		company, _ := db.GetCompany(s.db, *contact.CompanyID)
		if company != nil {
			companyName = company.Name
		}
	}

	data := map[string]interface{}{
		"Contact":     contact,
		"CompanyName": companyName,
	}

	s.renderTemplate(w, "partials/contact-detail.html", data)
}

func (s *Server) handleCompanyDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	company, err := db.GetCompany(s.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contacts, _ := db.FindContacts(s.db, "", &id, 100)

	data := map[string]interface{}{
		"Company":  company,
		"Contacts": contacts,
	}

	s.renderTemplate(w, "partials/company-detail.html", data)
}

func (s *Server) handleDealDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	deal, err := db.GetDeal(s.db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	company, _ := db.GetCompany(s.db, deal.CompanyID)
	companyName := ""
	if company != nil {
		companyName = company.Name
	}

	contactName := ""
	if deal.ContactID != nil {
		contact, _ := db.GetContact(s.db, *deal.ContactID)
		if contact != nil {
			contactName = contact.Name
		}
	}

	notes, _ := db.GetDealNotes(s.db, id)

	data := map[string]interface{}{
		"Deal":        deal,
		"CompanyName": companyName,
		"ContactName": contactName,
		"Notes":       notes,
	}

	s.renderTemplate(w, "partials/deal-detail.html", data)
}

func (s *Server) handleGraphs(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":           "Graphs",
		"ContentTemplate": "graphs-content",
	}

	s.renderTemplate(w, "layout.html", data)
}

func (s *Server) handleGraphPartial(w http.ResponseWriter, r *http.Request) {
	graphType := r.URL.Query().Get("type")
	entityIDStr := r.URL.Query().Get("entity_id")

	var dot string
	var err error

	switch graphType {
	case "contacts":
		var contactID *uuid.UUID
		if entityIDStr != "" {
			id, parseErr := uuid.Parse(entityIDStr)
			if parseErr == nil {
				contactID = &id
			}
		}
		dot, err = s.generator.GenerateContactGraph(contactID)

	case "company":
		if entityIDStr == "" {
			http.Error(w, "Company ID required", http.StatusBadRequest)
			return
		}
		companyID, parseErr := uuid.Parse(entityIDStr)
		if parseErr != nil {
			http.Error(w, "Invalid company ID", http.StatusBadRequest)
			return
		}
		dot, err = s.generator.GenerateCompanyGraph(companyID)

	case "pipeline":
		dot, err = s.generator.GeneratePipelineGraph()

	default:
		http.Error(w, "Invalid graph type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"DOT": dot,
	}

	s.renderTemplate(w, "partials/graph.html", data)
}

func (s *Server) handleFollowups(w http.ResponseWriter, r *http.Request) {
	followups, err := db.GetFollowupList(s.db, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Followups []models.FollowupContact
	}{
		Followups: followups,
	}

	err = s.templates.ExecuteTemplate(w, "followups", data)
	if err != nil {
		log.Printf("Template error rendering followups: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleFollowupLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contactID := strings.TrimPrefix(r.URL.Path, "/followups/log/")
	id, err := uuid.Parse(contactID)
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	interaction := &models.InteractionLog{
		ContactID:       id,
		InteractionType: models.InteractionMessage,
		Timestamp:       time.Now(),
		Notes:           "Quick contact via web UI",
	}

	err = db.LogInteraction(s.db, interaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(`<td colspan="5" class="px-4 py-3 text-green-600">✓ Interaction logged</td>`))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
