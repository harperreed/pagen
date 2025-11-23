# Phase 5: Web UI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add read-only web dashboard at `http://localhost:8080` using Go templates, HTMX, and Tailwind CSS

**Architecture:** `pagen web` starts HTTP server serving embedded templates with HTMX for partial updates

**Tech Stack:** Go html/template (embedded), HTMX (CDN), Tailwind CSS (CDN), Go stdlib net/http

---

## Task 5.1: Create web package and template structure

**Files:**
- Create: `web/server.go`
- Create: `web/templates/layout.html`
- Create: `web/templates/dashboard.html`

**Step 1: Create web package with HTTP server**

Create `web/server.go`:

```go
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

	"github.com/harperreed/pagen/db"
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
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
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

	// Partials for HTMX
	http.HandleFunc("/partials/contact-detail", s.handleContactDetail)
	http.HandleFunc("/partials/company-detail", s.handleCompanyDetail)
	http.HandleFunc("/partials/deal-detail", s.handleDealDetail)
	http.HandleFunc("/partials/graph", s.handleGraphPartial)

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
		"Stats": stats,
		"Title": "Dashboard",
	}

	s.renderTemplate(w, "dashboard.html", data)
}

func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	err := s.templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
```

**Step 2: Create base layout template**

Create `web/templates/layout.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Pagen CRM</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50">
    <nav class="bg-purple-600 text-white p-4">
        <div class="container mx-auto flex items-center justify-between">
            <h1 class="text-2xl font-bold">Pagen CRM</h1>
            <div class="space-x-4">
                <a href="/" class="hover:underline">Dashboard</a>
                <a href="/contacts" class="hover:underline">Contacts</a>
                <a href="/companies" class="hover:underline">Companies</a>
                <a href="/deals" class="hover:underline">Deals</a>
                <a href="/graphs" class="hover:underline">Graphs</a>
            </div>
        </div>
    </nav>

    <main class="container mx-auto p-6">
        {{template "content" .}}
    </main>

    <footer class="bg-gray-800 text-white p-4 mt-12">
        <div class="container mx-auto text-center">
            <p>Pagen CRM - Read-Only Dashboard</p>
        </div>
    </footer>
</body>
</html>
```

**Step 3: Create dashboard template**

Create `web/templates/dashboard.html`:

```html
{{define "content"}}
<div class="space-y-6">
    <!-- Header -->
    <div class="bg-white shadow rounded-lg p-6">
        <h2 class="text-3xl font-bold text-gray-800">Dashboard</h2>
        <p class="text-gray-600">Overview of your CRM data</p>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="bg-white shadow rounded-lg p-6">
            <div class="flex items-center">
                <div class="text-4xl mr-4">📇</div>
                <div>
                    <p class="text-gray-600 text-sm">Contacts</p>
                    <p class="text-3xl font-bold text-gray-800">{{.Stats.TotalContacts}}</p>
                </div>
            </div>
        </div>

        <div class="bg-white shadow rounded-lg p-6">
            <div class="flex items-center">
                <div class="text-4xl mr-4">🏢</div>
                <div>
                    <p class="text-gray-600 text-sm">Companies</p>
                    <p class="text-3xl font-bold text-gray-800">{{.Stats.TotalCompanies}}</p>
                </div>
            </div>
        </div>

        <div class="bg-white shadow rounded-lg p-6">
            <div class="flex items-center">
                <div class="text-4xl mr-4">💼</div>
                <div>
                    <p class="text-gray-600 text-sm">Deals</p>
                    <p class="text-3xl font-bold text-gray-800">{{.Stats.TotalDeals}}</p>
                </div>
            </div>
        </div>
    </div>

    <!-- Pipeline Overview -->
    <div class="bg-white shadow rounded-lg p-6">
        <h3 class="text-2xl font-bold text-gray-800 mb-4">Pipeline Overview</h3>
        <div class="space-y-3">
            {{range $stage, $stats := .Stats.PipelineByStage}}
            <div>
                <div class="flex justify-between mb-1">
                    <span class="text-sm font-medium text-gray-700">{{$stats.Stage}}</span>
                    <span class="text-sm text-gray-600">{{$stats.Count}} deals (${{divide $stats.Amount 100000}}K)</span>
                </div>
                <div class="w-full bg-gray-200 rounded-full h-2.5">
                    <div class="bg-purple-600 h-2.5 rounded-full" style="width: {{multiply (divide $stats.Count 20) 100}}%"></div>
                </div>
            </div>
            {{end}}
        </div>
    </div>

    <!-- Needs Attention -->
    {{if or .Stats.StaleContacts .Stats.StaleDeals}}
    <div class="bg-yellow-50 border-l-4 border-yellow-400 p-6">
        <h3 class="text-xl font-bold text-yellow-800 mb-3">⚠️ Needs Attention</h3>
        <div class="space-y-2">
            {{if .Stats.StaleContacts}}
            <p class="text-yellow-700">
                <span class="font-semibold">{{len .Stats.StaleContacts}}</span> contacts - no contact in 30+ days
            </p>
            {{end}}
            {{if .Stats.StaleDeals}}
            <p class="text-yellow-700">
                <span class="font-semibold">{{len .Stats.StaleDeals}}</span> deals - stale (no activity in 14+ days)
            </p>
            {{end}}
        </div>
    </div>
    {{end}}
</div>
{{end}}
```

**Step 4: Add template helper functions**

Add to `web/server.go` after the `NewServer` function:

```go
func NewServer(database *sql.DB) (*Server, error) {
	// Helper functions for templates
	funcMap := template.FuncMap{
		"divide": func(a, b int64) int64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"multiply": func(a, b int) int {
			return a * b
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Server{
		db:        database,
		templates: tmpl,
		generator: viz.NewGraphGenerator(database),
	}, nil
}
```

**Step 5: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 6: Commit**

```bash
git add web/server.go web/templates/layout.html web/templates/dashboard.html
git commit -m "feat: add web UI server with dashboard template"
```

---

## Task 5.2: Implement list pages with HTMX

**Files:**
- Create: `web/templates/contacts.html`
- Create: `web/templates/companies.html`
- Create: `web/templates/deals.html`
- Create: `web/templates/partials/contact-detail.html`
- Create: `web/templates/partials/company-detail.html`
- Create: `web/templates/partials/deal-detail.html`
- Modify: `web/server.go` (add handlers)

**Step 1: Create contacts list template**

Create `web/templates/contacts.html`:

```html
{{define "content"}}
<div class="space-y-6">
    <div class="bg-white shadow rounded-lg p-6">
        <h2 class="text-3xl font-bold text-gray-800 mb-4">Contacts</h2>

        <!-- Search -->
        <div class="mb-4">
            <input
                type="text"
                name="q"
                placeholder="Search contacts..."
                class="w-full px-4 py-2 border rounded-lg"
                hx-get="/contacts"
                hx-trigger="keyup changed delay:500ms"
                hx-target="#contacts-table"
            >
        </div>

        <!-- Table -->
        <div id="contacts-table">
            <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                    <tr>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Company</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                    {{range .Contacts}}
                    <tr class="hover:bg-gray-50">
                        <td class="px-6 py-4 whitespace-nowrap">{{.Name}}</td>
                        <td class="px-6 py-4 whitespace-nowrap">{{.Email}}</td>
                        <td class="px-6 py-4 whitespace-nowrap">{{.CompanyName}}</td>
                        <td class="px-6 py-4 whitespace-nowrap">
                            <button
                                class="text-purple-600 hover:text-purple-800"
                                hx-get="/partials/contact-detail?id={{.ID}}"
                                hx-target="#detail-panel"
                            >
                                View
                            </button>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>

    <!-- Detail Panel -->
    <div id="detail-panel"></div>
</div>
{{end}}
```

**Step 2: Create companies list template**

Create `web/templates/companies.html`:

```html
{{define "content"}}
<div class="space-y-6">
    <div class="bg-white shadow rounded-lg p-6">
        <h2 class="text-3xl font-bold text-gray-800 mb-4">Companies</h2>

        <!-- Search -->
        <div class="mb-4">
            <input
                type="text"
                name="q"
                placeholder="Search companies..."
                class="w-full px-4 py-2 border rounded-lg"
                hx-get="/companies"
                hx-trigger="keyup changed delay:500ms"
                hx-target="#companies-table"
            >
        </div>

        <!-- Table -->
        <div id="companies-table">
            <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                    <tr>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Domain</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Industry</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                    {{range .Companies}}
                    <tr class="hover:bg-gray-50">
                        <td class="px-6 py-4 whitespace-nowrap">{{.Name}}</td>
                        <td class="px-6 py-4 whitespace-nowrap">{{.Domain}}</td>
                        <td class="px-6 py-4 whitespace-nowrap">{{.Industry}}</td>
                        <td class="px-6 py-4 whitespace-nowrap">
                            <button
                                class="text-purple-600 hover:text-purple-800"
                                hx-get="/partials/company-detail?id={{.ID}}"
                                hx-target="#detail-panel"
                            >
                                View
                            </button>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>

    <!-- Detail Panel -->
    <div id="detail-panel"></div>
</div>
{{end}}
```

**Step 3: Create deals list template**

Create `web/templates/deals.html`:

```html
{{define "content"}}
<div class="space-y-6">
    <div class="bg-white shadow rounded-lg p-6">
        <h2 class="text-3xl font-bold text-gray-800 mb-4">Deals</h2>

        <!-- Filters -->
        <div class="mb-4 grid grid-cols-2 gap-4">
            <input
                type="text"
                name="q"
                placeholder="Search deals..."
                class="px-4 py-2 border rounded-lg"
                hx-get="/deals"
                hx-trigger="keyup changed delay:500ms"
                hx-target="#deals-table"
            >
            <select
                name="stage"
                class="px-4 py-2 border rounded-lg"
                hx-get="/deals"
                hx-trigger="change"
                hx-target="#deals-table"
            >
                <option value="">All Stages</option>
                <option value="prospecting">Prospecting</option>
                <option value="qualification">Qualification</option>
                <option value="proposal">Proposal</option>
                <option value="negotiation">Negotiation</option>
                <option value="closed_won">Closed Won</option>
                <option value="closed_lost">Closed Lost</option>
            </select>
        </div>

        <!-- Table -->
        <div id="deals-table">
            <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                    <tr>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Title</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Company</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Stage</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                        <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                    </tr>
                </thead>
                <tbody class="bg-white divide-y divide-gray-200">
                    {{range .Deals}}
                    <tr class="hover:bg-gray-50">
                        <td class="px-6 py-4 whitespace-nowrap">{{.Title}}</td>
                        <td class="px-6 py-4 whitespace-nowrap">{{.CompanyName}}</td>
                        <td class="px-6 py-4 whitespace-nowrap">
                            <span class="px-2 py-1 text-xs rounded-full bg-purple-100 text-purple-800">
                                {{.Stage}}
                            </span>
                        </td>
                        <td class="px-6 py-4 whitespace-nowrap">${{divide .Amount 100}} {{.Currency}}</td>
                        <td class="px-6 py-4 whitespace-nowrap">
                            <button
                                class="text-purple-600 hover:text-purple-800"
                                hx-get="/partials/deal-detail?id={{.ID}}"
                                hx-target="#detail-panel"
                            >
                                View
                            </button>
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>
        </div>
    </div>

    <!-- Detail Panel -->
    <div id="detail-panel"></div>
</div>
{{end}}
```

**Step 4: Create detail partials**

Create `web/templates/partials/contact-detail.html`:

```html
<div class="bg-white shadow rounded-lg p-6">
    <div class="flex justify-between items-start mb-4">
        <h3 class="text-2xl font-bold text-gray-800">{{.Contact.Name}}</h3>
        <button class="text-gray-400 hover:text-gray-600" onclick="this.parentElement.parentElement.remove()">✕</button>
    </div>

    <dl class="grid grid-cols-2 gap-4">
        <div>
            <dt class="text-sm font-medium text-gray-500">Email</dt>
            <dd class="mt-1 text-sm text-gray-900">{{.Contact.Email}}</dd>
        </div>
        <div>
            <dt class="text-sm font-medium text-gray-500">Phone</dt>
            <dd class="mt-1 text-sm text-gray-900">{{.Contact.Phone}}</dd>
        </div>
        {{if .CompanyName}}
        <div>
            <dt class="text-sm font-medium text-gray-500">Company</dt>
            <dd class="mt-1 text-sm text-gray-900">{{.CompanyName}}</dd>
        </div>
        {{end}}
        {{if .Contact.LastContactedAt}}
        <div>
            <dt class="text-sm font-medium text-gray-500">Last Contacted</dt>
            <dd class="mt-1 text-sm text-gray-900">{{.Contact.LastContactedAt.Format "2006-01-02"}}</dd>
        </div>
        {{end}}
    </dl>

    {{if .Contact.Notes}}
    <div class="mt-4">
        <dt class="text-sm font-medium text-gray-500">Notes</dt>
        <dd class="mt-1 text-sm text-gray-900">{{.Contact.Notes}}</dd>
    </div>
    {{end}}
</div>
```

Create `web/templates/partials/company-detail.html`:

```html
<div class="bg-white shadow rounded-lg p-6">
    <div class="flex justify-between items-start mb-4">
        <h3 class="text-2xl font-bold text-gray-800">{{.Company.Name}}</h3>
        <button class="text-gray-400 hover:text-gray-600" onclick="this.parentElement.parentElement.remove()">✕</button>
    </div>

    <dl class="grid grid-cols-2 gap-4">
        <div>
            <dt class="text-sm font-medium text-gray-500">Domain</dt>
            <dd class="mt-1 text-sm text-gray-900">{{.Company.Domain}}</dd>
        </div>
        <div>
            <dt class="text-sm font-medium text-gray-500">Industry</dt>
            <dd class="mt-1 text-sm text-gray-900">{{.Company.Industry}}</dd>
        </div>
    </dl>

    {{if .Company.Notes}}
    <div class="mt-4">
        <dt class="text-sm font-medium text-gray-500">Notes</dt>
        <dd class="mt-1 text-sm text-gray-900">{{.Company.Notes}}</dd>
    </div>
    {{end}}

    {{if .Contacts}}
    <div class="mt-6">
        <h4 class="text-lg font-semibold text-gray-800 mb-2">Contacts</h4>
        <ul class="space-y-1">
            {{range .Contacts}}
            <li class="text-sm text-gray-700">• {{.Name}} ({{.Email}})</li>
            {{end}}
        </ul>
    </div>
    {{end}}
</div>
```

Create `web/templates/partials/deal-detail.html`:

```html
<div class="bg-white shadow rounded-lg p-6">
    <div class="flex justify-between items-start mb-4">
        <h3 class="text-2xl font-bold text-gray-800">{{.Deal.Title}}</h3>
        <button class="text-gray-400 hover:text-gray-600" onclick="this.parentElement.parentElement.remove()">✕</button>
    </div>

    <dl class="grid grid-cols-2 gap-4">
        <div>
            <dt class="text-sm font-medium text-gray-500">Company</dt>
            <dd class="mt-1 text-sm text-gray-900">{{.CompanyName}}</dd>
        </div>
        {{if .ContactName}}
        <div>
            <dt class="text-sm font-medium text-gray-500">Contact</dt>
            <dd class="mt-1 text-sm text-gray-900">{{.ContactName}}</dd>
        </div>
        {{end}}
        <div>
            <dt class="text-sm font-medium text-gray-500">Stage</dt>
            <dd class="mt-1 text-sm text-gray-900">
                <span class="px-2 py-1 text-xs rounded-full bg-purple-100 text-purple-800">
                    {{.Deal.Stage}}
                </span>
            </dd>
        </div>
        <div>
            <dt class="text-sm font-medium text-gray-500">Amount</dt>
            <dd class="mt-1 text-sm text-gray-900">${{divide .Deal.Amount 100}} {{.Deal.Currency}}</dd>
        </div>
        {{if .Deal.ExpectedCloseDate}}
        <div>
            <dt class="text-sm font-medium text-gray-500">Expected Close</dt>
            <dd class="mt-1 text-sm text-gray-900">{{.Deal.ExpectedCloseDate.Format "2006-01-02"}}</dd>
        </div>
        {{end}}
    </dl>

    {{if .Notes}}
    <div class="mt-6">
        <h4 class="text-lg font-semibold text-gray-800 mb-2">Notes</h4>
        <ul class="space-y-2">
            {{range .Notes}}
            <li class="text-sm text-gray-700 border-l-2 border-purple-300 pl-3">
                <span class="text-gray-500">[{{.CreatedAt.Format "2006-01-02"}}]</span> {{.Content}}
            </li>
            {{end}}
        </ul>
    </div>
    {{end}}
</div>
```

**Step 5: Add handlers to web/server.go**

Add these handlers to `web/server.go`:

```go
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
		"Contacts": contactViews,
		"Title":    "Contacts",
	}

	s.renderTemplate(w, "contacts.html", data)
}

func (s *Server) handleCompanies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	companies, err := db.FindCompanies(s.db, query, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Companies": companies,
		"Title":     "Companies",
	}

	s.renderTemplate(w, "companies.html", data)
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
		var filtered []*models.Deal
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
		"Deals": dealViews,
		"Title": "Deals",
	}

	s.renderTemplate(w, "deals.html", data)
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
```

**Step 6: Add missing imports to web/server.go**

Add to imports:

```go
import (
	// ... existing imports ...
	"github.com/google/uuid"
	"github.com/harperreed/pagen/models"
)
```

**Step 7: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 8: Commit**

```bash
git add web/templates/contacts.html web/templates/companies.html web/templates/deals.html web/templates/partials/ web/server.go
git commit -m "feat: add web UI list pages with HTMX partials"
```

---

## Task 5.3: Implement graphs page with inline SVG

**Files:**
- Create: `web/templates/graphs.html`
- Create: `web/templates/partials/graph.html`
- Modify: `web/server.go` (add graph handlers)

**Step 1: Create graphs page template**

Create `web/templates/graphs.html`:

```html
{{define "content"}}
<div class="space-y-6">
    <div class="bg-white shadow rounded-lg p-6">
        <h2 class="text-3xl font-bold text-gray-800 mb-4">Graphs</h2>

        <!-- Graph Type Selector -->
        <div class="mb-6 grid grid-cols-3 gap-4">
            <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">Graph Type</label>
                <select
                    id="graph-type"
                    class="w-full px-4 py-2 border rounded-lg"
                >
                    <option value="contacts">Contact Relationships</option>
                    <option value="company">Company Org Chart</option>
                    <option value="pipeline">Deal Pipeline</option>
                </select>
            </div>

            <div id="entity-selector" style="display: none;">
                <label class="block text-sm font-medium text-gray-700 mb-2">Select Entity</label>
                <input
                    type="text"
                    id="entity-id"
                    placeholder="Entity ID"
                    class="w-full px-4 py-2 border rounded-lg"
                >
            </div>

            <div class="flex items-end">
                <button
                    class="px-6 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700"
                    onclick="generateGraph()"
                >
                    Generate Graph
                </button>
            </div>
        </div>

        <!-- Graph Display -->
        <div id="graph-display" class="border-t pt-6">
            <p class="text-gray-500 text-center">Select a graph type and click Generate Graph</p>
        </div>
    </div>
</div>

<script>
    // Show/hide entity selector based on graph type
    document.getElementById('graph-type').addEventListener('change', function() {
        const entitySelector = document.getElementById('entity-selector');
        if (this.value === 'company') {
            entitySelector.style.display = 'block';
        } else {
            entitySelector.style.display = 'none';
        }
    });

    function generateGraph() {
        const type = document.getElementById('graph-type').value;
        const entityId = document.getElementById('entity-id').value;

        let url = `/partials/graph?type=${type}`;
        if (entityId) {
            url += `&entity_id=${entityId}`;
        }

        htmx.ajax('GET', url, {target: '#graph-display'});
    }
</script>
{{end}}
```

**Step 2: Create graph partial template**

Create `web/templates/partials/graph.html`:

```html
<div class="space-y-4">
    <div class="flex justify-between items-center">
        <h3 class="text-xl font-semibold text-gray-800">Generated Graph</h3>
        <button
            class="px-4 py-2 text-sm bg-gray-100 text-gray-700 rounded hover:bg-gray-200"
            onclick="toggleDOT()"
        >
            Toggle DOT Source
        </button>
    </div>

    <!-- SVG Display (would need rendering - showing placeholder) -->
    <div class="border rounded-lg p-4 bg-gray-50">
        <p class="text-sm text-gray-600 mb-2">GraphViz DOT Output:</p>
        <pre class="text-xs bg-white p-4 rounded border overflow-x-auto">{{.DOT}}</pre>
    </div>

    <!-- DOT Source (collapsible) -->
    <div id="dot-source" class="hidden">
        <h4 class="text-lg font-semibold text-gray-800 mb-2">DOT Source</h4>
        <pre class="text-xs bg-gray-800 text-green-400 p-4 rounded overflow-x-auto">{{.DOT}}</pre>
    </div>
</div>

<script>
    function toggleDOT() {
        const dotSource = document.getElementById('dot-source');
        dotSource.classList.toggle('hidden');
    }
</script>
```

**Step 3: Add graph handlers to web/server.go**

Add these handlers:

```go
func (s *Server) handleGraphs(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Graphs",
	}

	s.renderTemplate(w, "graphs.html", data)
}

func (s *Server) handleGraphPartial(w http.ResponseWriter, r *http.Request) {
	graphType := r.URL.Query().Get("type")
	entityIDStr := r.URL.Query().Get("entity_id")

	ctx := context.Background()
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
		dot, err = s.generator.GenerateContactGraph(ctx, contactID)

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
		dot, err = s.generator.GenerateCompanyGraph(ctx, &companyID)

	case "pipeline":
		dot, err = s.generator.GeneratePipelineGraph(ctx)

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
```

**Step 4: Add context import to web/server.go**

Add to imports:

```go
import (
	"context"
	// ... other imports ...
)
```

**Step 5: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 6: Commit**

```bash
git add web/templates/graphs.html web/templates/partials/graph.html web/server.go
git commit -m "feat: add web UI graphs page with DOT rendering"
```

---

## Task 5.4: Wire up web command in main.go

**Files:**
- Modify: `main.go` (add web command)
- Create: `.scratch/test_web_manual.sh`

**Step 1: Add web command to main.go**

Add to command handling in `main.go`:

```go
case "web":
	port := 8080
	if len(args) > 1 && args[1] == "--port" && len(args) > 2 {
		fmt.Sscanf(args[2], "%d", &port)
	}

	server, err := web.NewServer(database)
	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}

	if err := server.Start(port); err != nil {
		return fmt.Errorf("web server error: %w", err)
	}
```

**Step 2: Add web import to main.go**

Add to imports:

```go
import (
	// ... existing imports ...
	"github.com/harperreed/pagen/web"
)
```

**Step 3: Create manual test script**

Create `.scratch/test_web_manual.sh`:

```bash
#!/bin/bash
set -e

echo "=== Web UI Manual Test Instructions ==="
echo ""
echo "This script creates test data. Then launch the web server manually."
echo ""

export DB=/tmp/test_web_$$.db

# Create test data
./pagen --db-path $DB crm add-company --name "Acme Corp" --domain "acme.com" --industry "Software"
./pagen --db-path $DB crm add-company --name "TechStart Inc" --domain "techstart.io" --industry "SaaS"
./pagen --db-path $DB crm add-contact --name "Alice Johnson" --email "alice@acme.com" --company "Acme Corp"
./pagen --db-path $DB crm add-contact --name "Bob Smith" --email "bob@techstart.io" --company "TechStart Inc"
./pagen --db-path $DB crm add-contact --name "Carol White" --email "carol@acme.com" --company "Acme Corp"
./pagen --db-path $DB crm add-deal --title "Enterprise Deal" --company "Acme Corp" --contact "Alice Johnson" --amount 500000 --stage "negotiation"
./pagen --db-path $DB crm add-deal --title "Startup Package" --company "TechStart Inc" --amount 50000 --stage "prospecting"
./pagen --db-path $DB crm add-deal --title "Premium License" --company "Acme Corp" --amount 250000 --stage "proposal"

echo ""
echo "Test data created in: $DB"
echo ""
echo "To launch web UI, run:"
echo "  ./pagen --db-path $DB web"
echo ""
echo "Then visit: http://localhost:8080"
echo ""
echo "Test checklist:"
echo "  [ ] Dashboard shows correct stats (3 contacts, 2 companies, 3 deals)"
echo "  [ ] Dashboard shows pipeline overview with bars"
echo "  [ ] Contacts page shows all 3 contacts"
echo "  [ ] Clicking 'View' on contact shows detail panel"
echo "  [ ] Search on contacts page filters results"
echo "  [ ] Companies page shows all 2 companies"
echo "  [ ] Company detail shows associated contacts"
echo "  [ ] Deals page shows all 3 deals"
echo "  [ ] Deal detail shows notes (if any)"
echo "  [ ] Graphs page allows selecting graph type"
echo "  [ ] Generating contact graph shows DOT output"
echo "  [ ] Generating pipeline graph shows DOT output"
echo "  [ ] All navigation links work"
echo ""
echo "Cleanup: rm $DB"
```

**Step 4: Make test script executable**

Run: `chmod +x .scratch/test_web_manual.sh`

**Step 5: Build and test**

Run: `make build`
Expected: Compiles successfully

Run: `.scratch/test_web_manual.sh`
Expected: Creates test data and prints instructions

**Step 6: Commit**

```bash
git add main.go .scratch/test_web_manual.sh
git commit -m "feat: wire up web command in main"
```

---

## Success Criteria

- [ ] `pagen web` starts HTTP server at localhost:8080
- [ ] Dashboard shows stats cards with emoji
- [ ] Dashboard shows pipeline overview with progress bars
- [ ] Dashboard shows "Needs Attention" section if applicable
- [ ] Contacts page shows searchable table
- [ ] Clicking "View" on contact loads detail panel via HTMX
- [ ] Search filters contacts without page reload
- [ ] Companies page shows all companies
- [ ] Company detail shows associated contacts
- [ ] Deals page shows all deals with stage badges
- [ ] Deal detail shows notes timeline
- [ ] Graphs page has type selector
- [ ] Generating graphs shows DOT source
- [ ] All pages use consistent styling (Tailwind)
- [ ] All templates are embedded in binary
- [ ] No external files needed to run web server
