// ABOUTME: Calendar event importer from Google Calendar API
// ABOUTME: Handles pagination, sync tokens, and progress logging for calendar events
package sync

import (
	"database/sql"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"

	"github.com/harperreed/pagen/db"
)

const (
	calendarService = "calendar"
	maxResults      = 250 // Google Calendar API max per page
)

// shouldSkipEvent determines if an event should be skipped during import
// Returns (true, reason) if the event should be skipped, (false, "") otherwise
func shouldSkipEvent(event *calendar.Event, userEmail string) (bool, string) {
	// Check for nil event
	if event == nil {
		return true, "nil event"
	}

	// Check for nil start time
	if event.Start == nil {
		return true, "missing start time"
	}

	// Skip all-day events (event.Start.Date is set instead of DateTime)
	if event.Start.Date != "" {
		return true, "all-day event"
	}

	// Skip cancelled events
	if event.Status == "cancelled" {
		return true, "cancelled"
	}

	// Skip declined events (check if user declined)
	// Use Self flag to identify the current user's attendee record
	for _, attendee := range event.Attendees {
		if attendee.Self && attendee.ResponseStatus == "declined" {
			return true, "declined"
		}
	}

	// Skip solo events (0 or 1 attendees)
	attendeeCount := len(event.Attendees)
	if attendeeCount <= 1 {
		return true, fmt.Sprintf("solo event (%d attendee%s)", attendeeCount, pluralize(attendeeCount))
	}

	// Don't skip this event
	return false, ""
}

// pluralize returns "s" if count != 1, otherwise ""
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// ImportCalendar fetches and imports calendar events from Google Calendar
func ImportCalendar(database *sql.DB, client *calendar.Service, initial bool) error {
	// Update sync state to 'syncing'
	fmt.Println("Syncing Google Calendar...")
	if err := db.UpdateSyncStatus(database, calendarService, "syncing", nil); err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	// Get user email from primary calendar for filtering
	calendarInfo, err := client.CalendarList.Get("primary").Do()
	if err != nil {
		errMsg := fmt.Sprintf("failed to get user calendar info: %v", err)
		_ = db.UpdateSyncStatus(database, calendarService, "error", &errMsg)
		return fmt.Errorf("failed to get user calendar info: %w", err)
	}
	userEmail := calendarInfo.Id

	// Get current sync state
	state, err := db.GetSyncState(database, calendarService)
	if err != nil {
		errMsg := err.Error()
		_ = db.UpdateSyncStatus(database, calendarService, "error", &errMsg)
		return fmt.Errorf("failed to get sync state: %w", err)
	}

	// Build the events list call
	call := client.Events.List("primary").
		MaxResults(maxResults).
		SingleEvents(true).
		OrderBy("startTime")

	// Use timeMin for initial sync or syncToken for incremental
	if initial {
		// Initial sync: fetch last 6 months
		sixMonthsAgo := time.Now().AddDate(0, -6, 0)
		call = call.TimeMin(sixMonthsAgo.Format(time.RFC3339))
		fmt.Printf("  → Initial sync (last 6 months)...\n")
	} else if state != nil && state.LastSyncToken != nil {
		// Incremental sync: use sync token
		call = call.SyncToken(*state.LastSyncToken)
		fmt.Printf("  → Incremental sync...\n")
	} else {
		// No sync token available, use timeMin
		sixMonthsAgo := time.Now().AddDate(0, -6, 0)
		call = call.TimeMin(sixMonthsAgo.Format(time.RFC3339))
		fmt.Printf("  → No previous sync found, fetching last 6 months...\n")
	}

	// Fetch events with pagination
	totalEvents := 0
	pageToken := ""

	// Track skip counts by reason
	skipCounts := make(map[string]int)

	for {
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		events, err := call.Do()
		if err != nil {
			// Handle 410 Gone error (invalid sync token)
			if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 410 {
				fmt.Println("  → Sync token invalid, falling back to time-based sync...")

				// Fall back to time-based sync using last sync time or 6 months ago
				var fallbackTime time.Time
				if state != nil && state.LastSyncTime != nil {
					fallbackTime = *state.LastSyncTime
				} else {
					fallbackTime = time.Now().AddDate(0, -6, 0)
				}

				// Rebuild call with timeMin instead of sync token and reset pagination
				call = client.Events.List("primary").
					MaxResults(maxResults).
					SingleEvents(true).
					OrderBy("startTime").
					TimeMin(fallbackTime.Format(time.RFC3339))
				totalEvents = 0

				// Retry the call
				events, err = call.Do()
				if err != nil {
					errMsg := fmt.Sprintf("failed to fetch events after fallback: %v", err)
					_ = db.UpdateSyncStatus(database, calendarService, "error", &errMsg)
					return fmt.Errorf("failed to fetch calendar events after fallback: %w", err)
				}
			} else {
				errMsg := fmt.Sprintf("failed to fetch events: %v", err)
				_ = db.UpdateSyncStatus(database, calendarService, "error", &errMsg)
				return fmt.Errorf("failed to fetch calendar events: %w", err)
			}
		}

		eventCount := len(events.Items)
		totalEvents += eventCount

		if eventCount > 0 {
			pageNum := (totalEvents-eventCount)/maxResults + 1
			fmt.Printf("  → Fetched %d events (page %d)\n", eventCount, pageNum)
		}

		// Process events and apply filters
		for _, event := range events.Items {
			skip, reason := shouldSkipEvent(event, userEmail)
			if skip {
				skipCounts[reason]++
				continue
			}

			// TODO (Task 4-6): Process event - extract attendees and log interaction
			// For now, just count as processed
		}

		// Check for next page
		pageToken = events.NextPageToken
		if pageToken == "" {
			// Last page - save sync token
			if events.NextSyncToken != "" {
				if err := db.UpdateSyncToken(database, calendarService, events.NextSyncToken); err != nil {
					errMsg := err.Error()
					_ = db.UpdateSyncStatus(database, calendarService, "error", &errMsg)
					return fmt.Errorf("failed to update sync token: %w", err)
				}
			}
			break
		}
	}

	// Update sync state to 'idle' on success
	if err := db.UpdateSyncStatus(database, calendarService, "idle", nil); err != nil {
		return fmt.Errorf("failed to update sync status: %w", err)
	}

	// Print summary
	fmt.Printf("\n✓ Fetched %d events\n", totalEvents)

	// Print skip summary if any events were skipped
	if len(skipCounts) > 0 {
		// Calculate total skipped
		totalSkipped := 0
		for _, count := range skipCounts {
			totalSkipped += count
		}

		// Print individual skip reasons
		for reason, count := range skipCounts {
			fmt.Printf("  ✓ Skipped %d %s event%s\n", count, reason, pluralize(count))
		}

		processedCount := totalEvents - totalSkipped
		fmt.Printf("\n  → Processing %d meeting%s...\n", processedCount, pluralize(processedCount))
	}

	fmt.Println("Sync token saved. Next sync will be incremental.")

	return nil
}
