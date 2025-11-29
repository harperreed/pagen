// ABOUTME: Google sync CLI commands
// ABOUTME: Handles OAuth setup, sync operations, and suggestion review
package cli

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/harperreed/pagen/db"
	"github.com/harperreed/pagen/sync"
	"golang.org/x/oauth2"
)

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
