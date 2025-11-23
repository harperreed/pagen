# Phase 3: Terminal Dashboard Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add static ASCII terminal dashboard showing pipeline overview, stats, recent activity, and attention items

**Architecture:** Single `pagen viz` command that queries the database, calculates metrics, and renders an ASCII dashboard to stdout using Unicode box-drawing characters.

**Tech Stack:** Pure Go stdlib (no dependencies needed), Unicode box-drawing characters for borders

---

## Task 3.1: Add dashboard stats aggregation functions

**Files:**
- Create: `viz/dashboard.go`

**Step 1: Create dashboard package file**

Create `viz/dashboard.go`:

```go
// ABOUTME: Terminal dashboard statistics and rendering
// ABOUTME: Provides ASCII dashboard for CRM overview
package viz

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/harperreed/pagen/db"
	"github.com/harperreed/pagen/models"
)

type DashboardStats struct {
	// Pipeline overview
	PipelineByStage map[string]PipelineStageStats

	// Overall stats
	TotalContacts  int
	TotalCompanies int
	TotalDeals     int

	// Recent activity (last 7 days)
	RecentActivity []ActivityItem

	// Needs attention
	StaleContacts []StaleContact
	StaleDeals    []StaleDeal
}

type PipelineStageStats struct {
	Stage  string
	Count  int
	Amount int64 // in cents
}

type ActivityItem struct {
	Date        time.Time
	Description string
}

type StaleContact struct {
	Name       string
	DaysSince  int
}

type StaleDeal struct {
	Title     string
	DaysSince int
}

func GenerateDashboardStats(database *sql.DB) (*DashboardStats, error) {
	stats := &DashboardStats{
		PipelineByStage: make(map[string]PipelineStageStats),
	}

	// Get pipeline stats
	deals, err := db.FindDeals(database, "", nil, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deals: %w", err)
	}

	for _, deal := range deals {
		stage := deal.Stage
		if stage == "" {
			stage = "unknown"
		}

		pstats := stats.PipelineByStage[stage]
		pstats.Stage = stage
		pstats.Count++
		pstats.Amount += deal.Amount
		stats.PipelineByStage[stage] = pstats
	}

	stats.TotalDeals = len(deals)

	// Get contact stats
	contacts, err := db.FindContacts(database, "", nil, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contacts: %w", err)
	}
	stats.TotalContacts = len(contacts)

	// Get company stats
	companies, err := db.FindCompanies(database, "", 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch companies: %w", err)
	}
	stats.TotalCompanies = len(companies)

	// Find stale contacts (no contact in 30+ days)
	now := time.Now()
	for _, contact := range contacts {
		if contact.LastContactedAt == nil {
			stats.StaleContacts = append(stats.StaleContacts, StaleContact{
				Name:      contact.Name,
				DaysSince: -1, // Never contacted
			})
		} else {
			daysSince := int(now.Sub(*contact.LastContactedAt).Hours() / 24)
			if daysSince > 30 {
				stats.StaleContacts = append(stats.StaleContacts, StaleContact{
					Name:      contact.Name,
					DaysSince: daysSince,
				})
			}
		}
	}

	// Find stale deals (no activity in 14+ days)
	for _, deal := range deals {
		daysSince := int(now.Sub(deal.LastActivityAt).Hours() / 24)
		if daysSince > 14 {
			stats.StaleDeals = append(stats.StaleDeals, StaleDeal{
				Title:     deal.Title,
				DaysSince: daysSince,
			})
		}
	}

	return stats, nil
}
```

**Step 2: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 3: Commit**

```bash
git add viz/dashboard.go
git commit -m "feat: add dashboard stats aggregation"
```

---

## Task 3.2: Add ASCII rendering functions

**Files:**
- Modify: `viz/dashboard.go` (add rendering functions)

**Step 1: Add RenderDashboard function**

Add to `viz/dashboard.go`:

```go
func RenderDashboard(stats *DashboardStats) string {
	var out strings.Builder

	// Header
	out.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	out.WriteString("  PAGEN CRM DASHBOARD\n")
	out.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Pipeline overview
	out.WriteString("PIPELINE OVERVIEW\n")
	renderPipeline(&out, stats.PipelineByStage)
	out.WriteString("\n")

	// Stats
	out.WriteString("STATS\n")
	out.WriteString(fmt.Sprintf("  📇 %d contacts  🏢 %d companies  💼 %d deals\n\n",
		stats.TotalContacts, stats.TotalCompanies, stats.TotalDeals))

	// Needs attention
	if len(stats.StaleContacts) > 0 || len(stats.StaleDeals) > 0 {
		out.WriteString("NEEDS ATTENTION\n")

		if len(stats.StaleContacts) > 0 {
			out.WriteString(fmt.Sprintf("  ⚠️  %d contacts - no contact in 30+ days\n", len(stats.StaleContacts)))
		}

		if len(stats.StaleDeals) > 0 {
			out.WriteString(fmt.Sprintf("  ⚠️  %d deals - stale (no activity in 14+ days)\n", len(stats.StaleDeals)))
		}
	}

	return out.String()
}

func renderPipeline(out *strings.Builder, pipeline map[string]PipelineStageStats) {
	// Define stage order
	stages := []string{
		models.StageProspecting,
		models.StageQualification,
		models.StageProposal,
		models.StageNegotiation,
		models.StageClosedWon,
		models.StageClosedLost,
	}

	// Find max count for scaling
	maxCount := 0
	for _, pstats := range pipeline {
		if pstats.Count > maxCount {
			maxCount = pstats.Count
		}
	}
	if maxCount == 0 {
		maxCount = 1
	}

	// Render each stage
	for _, stage := range stages {
		pstats, exists := pipeline[stage]
		if !exists {
			continue
		}

		// Calculate bar length (0-10 blocks)
		barLength := (pstats.Count * 10) / maxCount

		// Build bar
		bar := strings.Repeat("█", barLength) + strings.Repeat("░", 10-barLength)

		// Format amount in K
		amountK := pstats.Amount / 100000

		out.WriteString(fmt.Sprintf("  %-13s %s  %2d ($%dK)\n",
			stage, bar, pstats.Count, amountK))
	}
}
```

**Step 2: Add strings import**

Add to imports in `viz/dashboard.go`:
```go
import (
	// existing imports...
	"strings"
)
```

**Step 3: Build and test**

Run: `make build`
Expected: Compiles successfully

**Step 4: Commit**

```bash
git add viz/dashboard.go
git commit -m "feat: add ASCII dashboard rendering"
```

---

## Task 3.3: Add viz dashboard CLI command

**Files:**
- Modify: `cli/viz.go` (add DashboardCommand)
- Modify: `main.go` (register viz command)

**Step 1: Add DashboardCommand to cli/viz.go**

Add to `cli/viz.go`:

```go
func VizDashboardCommand(database *sql.DB, args []string) error {
	stats, err := viz.GenerateDashboardStats(database)
	if err != nil {
		return fmt.Errorf("failed to generate dashboard stats: %w", err)
	}

	output := viz.RenderDashboard(stats)
	fmt.Print(output)

	return nil
}
```

**Step 2: Register in main.go**

In `main.go`, update the viz command handling:

```go
case "viz":
	if len(args) < 2 {
		// No subcommand = dashboard
		if err := cli.VizDashboardCommand(database, args[1:]); err != nil {
			return err
		}
	} else if args[1] == "graph" {
		// Existing graph handling...
	}
```

**Step 3: Create test scenario**

Create `.scratch/test_dashboard.sh`:

```bash
#!/bin/bash
set -e

echo "=== Testing Dashboard ==="

export DB=/tmp/test_dashboard_$$.db

# Create test data
./pagen --db-path $DB crm add-company --name "Test Corp"
./pagen --db-path $DB crm add-contact --name "Alice" --company "Test Corp"
./pagen --db-path $DB crm add-deal --title "Big Deal" --company "Test Corp" --amount 100000 --stage "negotiation"

# Generate dashboard
OUTPUT=$(./pagen --db-path $DB viz)

# Verify output
echo "$OUTPUT" | grep "PAGEN CRM DASHBOARD" || exit 1
echo "$OUTPUT" | grep "PIPELINE OVERVIEW" || exit 1
echo "$OUTPUT" | grep "STATS" || exit 1
echo "$OUTPUT" | grep "1 contacts" || exit 1
echo "$OUTPUT" | grep "1 companies" || exit 1
echo "$OUTPUT" | grep "1 deals" || exit 1

# Cleanup
rm $DB

echo "✓ Dashboard test passed"
```

**Step 4: Run test**

Run: `chmod +x .scratch/test_dashboard.sh && ./.scratch/test_dashboard.sh`
Expected: Test passes

**Step 5: Commit**

```bash
git add cli/viz.go main.go .scratch/test_dashboard.sh
git commit -m "feat: add viz dashboard CLI command"
```

---

## Success Criteria

- [ ] `pagen viz` displays ASCII dashboard
- [ ] Pipeline overview shows all stages with bars
- [ ] Stats show total counts with emoji
- [ ] Needs attention section shows stale items
- [ ] Dashboard updates in real-time with current data
- [ ] All tests pass
