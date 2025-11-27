// ABOUTME: Follow-up tracking CLI commands
// ABOUTME: Commands for listing follow-ups, logging interactions, setting cadence
package cli

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/harperreed/pagen/db"
	"github.com/harperreed/pagen/models"
)

// FollowupListCommand lists contacts needing follow-up
func FollowupListCommand(database *sql.DB, args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	overdueOnly := fs.Bool("overdue-only", false, "Show only overdue contacts")
	strength := fs.String("strength", "", "Filter by relationship strength (weak/medium/strong)")
	limit := fs.Int("limit", 10, "Maximum number of contacts to show")
	_ = fs.Parse(args)

	followups, err := db.GetFollowupList(database, *limit)
	if err != nil {
		return fmt.Errorf("failed to get followup list: %w", err)
	}

	// Apply filters
	var filtered []models.FollowupContact
	for _, f := range followups {
		if *overdueOnly && f.PriorityScore <= 0 {
			continue
		}
		if *strength != "" && f.RelationshipStrength != *strength {
			continue
		}
		filtered = append(filtered, f)
	}

	// Print results
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tDAYS SINCE\tPRIORITY\tSTRENGTH\tEMAIL")
	_, _ = fmt.Fprintln(w, "----\t----------\t--------\t--------\t-----")

	for _, f := range filtered {
		indicator := "🟢"
		if f.DaysSinceContact > f.CadenceDays+7 {
			indicator = "🔴"
		} else if f.DaysSinceContact >= f.CadenceDays-3 {
			indicator = "🟡"
		}

		_, _ = fmt.Fprintf(w, "%s %s\t%d\t%.1f\t%s\t%s\n",
			indicator, f.Name, f.DaysSinceContact, f.PriorityScore,
			f.RelationshipStrength, f.Email)
	}

	_ = w.Flush()
	return nil
}

// FollowupStatsCommand shows follow-up statistics
func FollowupStatsCommand(database *sql.DB, args []string) error {
	query := `
		SELECT
			relationship_strength,
			COUNT(*) as count,
			AVG(CAST((julianday('now') - julianday(last_interaction_date)) AS INTEGER)) as avg_days
		FROM contact_cadence
		WHERE last_interaction_date IS NOT NULL
		GROUP BY relationship_strength
	`

	rows, err := database.Query(query)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	fmt.Println("NETWORK HEALTH")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for rows.Next() {
		var strength string
		var count int
		var avgDays sql.NullFloat64

		err := rows.Scan(&strength, &count, &avgDays)
		if err != nil {
			return err
		}

		icon := "🟢"
		switch strength {
		case models.StrengthWeak:
			icon = "🔴"
		case models.StrengthMedium:
			icon = "🟡"
		}

		fmt.Printf("  %s %s relationships: %d (avg contact: %.0f days)\n",
			icon, strength, count, avgDays.Float64)
	}

	return rows.Err()
}
