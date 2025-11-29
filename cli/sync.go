// ABOUTME: Google sync CLI commands
// ABOUTME: Handles OAuth setup, sync operations, and suggestion review
package cli

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/harperreed/pagen/db"
	"github.com/harperreed/pagen/sync"
	"golang.org/x/oauth2"
)

// SyncAllCommand syncs all Google services (contacts, calendar, gmail)
func SyncAllCommand(database *sql.DB) error {
	fmt.Println("=== Syncing All Google Services ===")

	// Load OAuth token once
	token, err := sync.LoadToken()
	if err != nil {
		return fmt.Errorf("no authentication token found. Run 'pagen sync init' first: %w", err)
	}

	// Track overall progress
	totalErrors := 0
	services := []struct {
		name string
		sync func() error
	}{
		{"Contacts", func() error {
			client, err := sync.NewPeopleClient(token)
			if err != nil {
				return fmt.Errorf("failed to create People API client: %w", err)
			}
			return sync.ImportContacts(database, client)
		}},
		{"Calendar", func() error {
			client, err := sync.NewCalendarClient(token)
			if err != nil {
				return fmt.Errorf("failed to create Calendar client: %w", err)
			}
			return sync.ImportCalendar(database, client, false) // incremental
		}},
		{"Gmail", func() error {
			client, err := sync.NewGmailClient(token)
			if err != nil {
				return fmt.Errorf("failed to create Gmail client: %w", err)
			}
			return sync.ImportGmail(database, client, false) // incremental
		}},
	}

	// Sync each service
	for i, service := range services {
		fmt.Printf("[%d/%d] %s\n", i+1, len(services), service.name)
		fmt.Println(strings.Repeat("-", 50))

		if err := service.sync(); err != nil {
			fmt.Printf("✗ %s sync failed: %v\n\n", service.name, err)
			totalErrors++
		} else {
			fmt.Printf("✓ %s sync completed\n\n", service.name)
		}
	}

	// Summary
	fmt.Println(strings.Repeat("=", 50))
	if totalErrors == 0 {
		fmt.Println("✓ All services synced successfully!")
	} else {
		fmt.Printf("⚠ Completed with %d error(s)\n", totalErrors)
	}

	return nil
}

// SyncInitCommand handles OAuth setup
func SyncInitCommand(database *sql.DB, args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	_ = fs.Parse(args)

	ctx := context.Background()

	config, err := sync.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get OAuth config: %w", err)
	}

	// Start local server for OAuth callback
	callbackChan := make(chan *oauth2.Token)
	errChan := make(chan error)

	http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			return
		}

		token, err := config.Exchange(ctx, code)
		if err != nil {
			errChan <- fmt.Errorf("failed to exchange code: %w", err)
			return
		}

		callbackChan <- token
		_, _ = fmt.Fprintf(w, "Authorization successful! You can close this window.")
	})

	server := &http.Server{Addr: ":8080"}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Generate auth URL
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	fmt.Println("Opening browser for Google OAuth...")
	fmt.Printf("\nIf browser doesn't open, visit this URL:\n%s\n\n", authURL)

	// Try to open browser
	_ = openBrowser(authURL)

	// Wait for callback or error
	select {
	case token := <-callbackChan:
		_ = server.Shutdown(ctx)

		if err := sync.SaveToken(token); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		fmt.Printf("\n✓ Authenticated successfully\n")
		fmt.Printf("✓ Tokens saved to %s\n\n", sync.TokenPath())
		fmt.Println("Ready to sync! Run 'pagen sync contacts' to import contacts.")

		return nil

	case err := <-errChan:
		_ = server.Shutdown(ctx)
		return fmt.Errorf("OAuth flow failed: %w", err)
	}
}

// SyncContactsCommand syncs Google Contacts
func SyncContactsCommand(database *sql.DB, args []string) error {
	fs := flag.NewFlagSet("contacts", flag.ExitOnError)
	_ = fs.Parse(args)

	// Load OAuth token
	token, err := sync.LoadToken()
	if err != nil {
		return fmt.Errorf("no authentication token found. Run 'pagen sync init' first: %w", err)
	}

	// Create People API client
	client, err := sync.NewPeopleClient(token)
	if err != nil {
		return fmt.Errorf("failed to create People API client: %w", err)
	}

	// Import contacts
	if err := sync.ImportContacts(database, client); err != nil {
		return fmt.Errorf("contacts sync failed: %w", err)
	}

	return nil
}

// SyncCalendarCommand syncs Google Calendar events
func SyncCalendarCommand(database *sql.DB, args []string) error {
	fs := flag.NewFlagSet("calendar", flag.ExitOnError)
	initial := fs.Bool("initial", false, "Full import (last 6 months)")
	_ = fs.Parse(args)

	// Load OAuth token
	token, err := sync.LoadToken()
	if err != nil {
		return fmt.Errorf("no authentication token found. Run 'pagen sync init' first: %w", err)
	}

	// Create Calendar client
	client, err := sync.NewCalendarClient(token)
	if err != nil {
		return fmt.Errorf("failed to create Calendar client: %w", err)
	}

	// Import calendar events
	if err := sync.ImportCalendar(database, client, *initial); err != nil {
		return fmt.Errorf("calendar sync failed: %w", err)
	}

	return nil
}

// SyncGmailCommand syncs Gmail emails
func SyncGmailCommand(database *sql.DB, args []string) error {
	fs := flag.NewFlagSet("gmail", flag.ExitOnError)
	initial := fs.Bool("initial", false, "Import last 30 days")
	_ = fs.Parse(args)

	// Load OAuth token
	token, err := sync.LoadToken()
	if err != nil {
		return fmt.Errorf("no authentication token found. Run 'pagen sync init' first: %w", err)
	}

	// Create Gmail client
	client, err := sync.NewGmailClient(token)
	if err != nil {
		return fmt.Errorf("failed to create Gmail client: %w", err)
	}

	// Import emails
	if err := sync.ImportGmail(database, client, *initial); err != nil {
		return fmt.Errorf("gmail sync failed: %w", err)
	}

	return nil
}

// SyncResetCommand resets a stuck sync state
func SyncResetCommand(database *sql.DB, args []string) error {
	if len(args) == 0 {
		fmt.Println("Usage: pagen sync reset <service>")
		fmt.Println("\nAvailable services:")
		fmt.Println("  calendar - Reset calendar sync state")
		fmt.Println("  contacts - Reset contacts sync state")
		fmt.Println("  all      - Reset all sync states")
		return nil
	}

	service := args[0]

	if service == "all" {
		// Reset all services
		_, err := database.Exec(`UPDATE sync_state SET status='idle', last_sync_time=datetime('now')`)
		if err != nil {
			return fmt.Errorf("failed to reset sync states: %w", err)
		}
		fmt.Println("✓ Reset all sync states to 'idle'")
		return nil
	}

	// Reset specific service
	result, err := database.Exec(`UPDATE sync_state SET status='idle', last_sync_time=datetime('now') WHERE service=?`, service)
	if err != nil {
		return fmt.Errorf("failed to reset sync state for %s: %w", service, err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		fmt.Printf("⚠ No sync state found for service '%s'\n", service)
		return nil
	}

	fmt.Printf("✓ Reset %s sync state to 'idle'\n", service)
	fmt.Println("\nNext sync will be incremental from now.")
	return nil
}

// SyncStatusCommand displays the sync status for all Google services
func SyncStatusCommand(database *sql.DB, args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	_ = fs.Parse(args)

	// Get all sync states
	states, err := db.GetAllSyncStates(database)
	if err != nil {
		return fmt.Errorf("failed to get sync states: %w", err)
	}

	// If no states exist, show helpful message
	if len(states) == 0 {
		fmt.Println("Google Sync Status:")
		fmt.Println("  No sync data found. Run 'pagen sync init' to set up authentication.")
		return nil
	}

	// Display header
	fmt.Println("Google Sync Status:")

	// Define expected services and their display order
	expectedServices := []string{"calendar", "contacts", "gmail"}
	stateMap := make(map[string]db.SyncState)
	for _, state := range states {
		stateMap[state.Service] = state
	}

	// Display each service
	for _, service := range expectedServices {
		state, exists := stateMap[service]
		// Capitalize first letter manually to avoid deprecated strings.Title
		displayService := service
		if len(service) > 0 {
			displayService = strings.ToUpper(service[:1]) + service[1:]
		}

		if !exists {
			// Service not yet synced
			fmt.Printf("  %-10s Not synced yet\n", displayService+":")
			continue
		}

		// Determine status icon
		var icon string
		switch state.Status {
		case "error":
			icon = "✗"
		case "syncing":
			icon = "!"
		default:
			icon = "✓"
		}

		// Format last sync time
		var timeStr string
		if state.LastSyncTime != nil {
			timeStr = formatTimeSince(*state.LastSyncTime)
		} else {
			timeStr = "never"
		}

		// Build status message
		var statusMsg string
		switch state.Status {
		case "error":
			if state.ErrorMessage != nil {
				statusMsg = fmt.Sprintf("Error: %s", *state.ErrorMessage)
			} else {
				statusMsg = "Error (no details)"
			}
		case "syncing":
			statusMsg = "Currently syncing..."
		default:
			// Check if incremental sync is enabled
			if state.LastSyncToken != nil && *state.LastSyncToken != "" {
				statusMsg = fmt.Sprintf("Last synced %s (idle, incremental sync enabled)", timeStr)
			} else {
				statusMsg = fmt.Sprintf("Last synced %s (idle)", timeStr)
			}
		}

		fmt.Printf("  %-10s %s %s\n", displayService+":", icon, statusMsg)
	}

	return nil
}

// formatTimeSince formats a time duration in a human-readable way
func formatTimeSince(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// openBrowser attempts to open URL in default browser
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	command := exec.Command(cmd, args...)
	return command.Start()
}

// SyncDaemonCommand runs sync in daemon mode with configurable interval
func SyncDaemonCommand(database *sql.DB, args []string) error {
	fs := flag.NewFlagSet("daemon", flag.ExitOnError)
	interval := fs.String("interval", "1h", "Sync interval (e.g., 15m, 1h, 4h)")
	servicesStr := fs.String("services", "all", "Comma-separated services to sync (contacts,calendar,gmail,all)")
	_ = fs.Parse(args)

	// Parse interval duration
	duration, err := time.ParseDuration(*interval)
	if err != nil {
		return fmt.Errorf("invalid interval format: %w", err)
	}

	// Validate minimum interval (5 minutes to prevent API hammering)
	if duration < 5*time.Minute {
		return fmt.Errorf("interval must be at least 5 minutes to respect API rate limits")
	}

	// Parse services list
	services := parseServices(*servicesStr)
	if len(services) == 0 {
		return fmt.Errorf("no valid services specified")
	}

	log.Printf("Starting pagen sync daemon")
	log.Printf("  Interval: %s", duration)
	log.Printf("  Services: %s", strings.Join(services, ", "))
	log.Printf("  Database: %+v", database.Stats())

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create ticker for scheduled syncs
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	// Run initial sync immediately
	log.Println("Running initial sync...")
	if err := runDaemonSync(database, services); err != nil {
		log.Printf("Initial sync failed: %v", err)
	}

	// Main daemon loop
	for {
		select {
		case <-ticker.C:
			log.Printf("Starting scheduled sync (interval: %s)", duration)
			if err := runDaemonSync(database, services); err != nil {
				log.Printf("Scheduled sync failed: %v", err)
			}

		case sig := <-sigChan:
			log.Printf("Received signal %s, shutting down gracefully...", sig)
			return nil
		}
	}
}

// parseServices converts comma-separated service string to slice
func parseServices(servicesStr string) []string {
	if servicesStr == "all" {
		return []string{"contacts", "calendar", "gmail"}
	}

	parts := strings.Split(servicesStr, ",")
	var services []string
	validServices := map[string]bool{
		"contacts": true,
		"calendar": true,
		"gmail":    true,
	}

	for _, part := range parts {
		service := strings.TrimSpace(part)
		if validServices[service] {
			services = append(services, service)
		} else {
			log.Printf("Warning: ignoring invalid service '%s'", service)
		}
	}

	return services
}

// runDaemonSync executes sync for specified services
func runDaemonSync(database *sql.DB, services []string) error {
	startTime := time.Now()

	// Load OAuth token once
	token, err := sync.LoadToken()
	if err != nil {
		return fmt.Errorf("no authentication token found. Run 'pagen sync init' first: %w", err)
	}

	// Track errors
	errorCount := 0
	successCount := 0

	// Sync each service
	for _, service := range services {
		serviceStart := time.Now()
		var err error

		switch service {
		case "contacts":
			client, createErr := sync.NewPeopleClient(token)
			if createErr != nil {
				err = fmt.Errorf("failed to create People API client: %w", createErr)
			} else {
				err = sync.ImportContacts(database, client)
			}

		case "calendar":
			client, createErr := sync.NewCalendarClient(token)
			if createErr != nil {
				err = fmt.Errorf("failed to create Calendar client: %w", createErr)
			} else {
				err = sync.ImportCalendar(database, client, false) // incremental
			}

		case "gmail":
			client, createErr := sync.NewGmailClient(token)
			if createErr != nil {
				err = fmt.Errorf("failed to create Gmail client: %w", createErr)
			} else {
				err = sync.ImportGmail(database, client, false) // incremental
			}
		}

		duration := time.Since(serviceStart)

		if err != nil {
			log.Printf("✗ %s sync failed (%.2fs): %v", service, duration.Seconds(), err)
			errorCount++
		} else {
			log.Printf("✓ %s sync completed (%.2fs)", service, duration.Seconds())
			successCount++
		}
	}

	totalDuration := time.Since(startTime)
	log.Printf("Sync cycle completed in %.2fs (%d succeeded, %d failed)",
		totalDuration.Seconds(), successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("%d service(s) failed to sync", errorCount)
	}

	return nil
}
