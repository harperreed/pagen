# Google Sync Phase 2: Calendar Integration

**Date:** 2025-11-28
**Goal:** Import Google Calendar events as interactions and create contacts from attendees

## Overview

Phase 2 builds on Phase 1's OAuth foundation to add Calendar integration. We focus on clean data import: events → interactions, attendees → contacts. No deal detection or intelligence layer yet - that's deferred to Phase 4.

## Scope

**What we're building:**
- Google Calendar API integration
- Event → interaction mapping
- Attendee → contact creation/updates
- Incremental sync using sync tokens
- CLI: `pagen sync calendar`

**What we're NOT building (yet):**
- Deal detection from event titles
- Suggestions creation
- Intelligence/analysis layer
- TUI integration

## Architecture

### Calendar Importer (`sync/calendar_importer.go`)

**Core Functions:**

```go
// ImportCalendar fetches and imports calendar events
func ImportCalendar(db *sql.DB, client *calendar.Service, initial bool) error

// processEvent converts a Calendar event to pagen entities
func processEvent(db *sql.DB, event *calendar.Event, matcher *ContactMatcher) error

// extractAttendees creates/updates contacts from event attendees
func extractAttendees(db *sql.DB, event *calendar.Event, matcher *ContactMatcher) ([]uuid.UUID, error)

// logInteraction creates interaction_log entry for the event
func logInteraction(db *sql.DB, contactIDs []uuid.UUID, event *calendar.Event) error
```

### Event Filtering

**Skip these events:**
- All-day events (`event.Start.Date != ""`)
- Declined events (`attendee.ResponseStatus == "declined"` for self)
- Cancelled events (`event.Status == "cancelled"`)
- Solo events (`len(event.Attendees) <= 1`)
- Events older than configured history window

### Data Mapping

**Calendar Event → Interaction Log:**
```
event.Summary              → notes
event.Start.DateTime       → timestamp
"meeting"                  → type
event.Id                   → sync_log.source_id
{
  "location": event.Location,
  "duration_minutes": duration,
  "attendee_count": count,
  "calendar_event_id": event.Id
}                          → metadata (JSON)
```

**Calendar Attendee → Contact:**
```
attendee.Email             → email (primary key for matching)
attendee.DisplayName       → name (if not already set)
```

### Incremental Sync

**First sync (`--initial` flag):**
- Use `timeMin` parameter: 6 months ago
- Fetch all events in that window
- Save `syncToken` from response

**Subsequent syncs:**
- Use `syncToken` from `sync_state.last_sync_token`
- Only fetches changed/new events
- Updates sync token after successful sync

**Token invalidation handling:**
- If sync token invalid (410 error), fall back to `updatedMin` with last sync time
- Log warning to user
- Continue with time-based incremental sync

## API Integration

### Google Calendar API

**Package:** `google.golang.org/api/calendar/v3`

**Key Methods:**
```go
service.Events.List(calendarId).
    TimeMin(sixMonthsAgo).
    MaxResults(250).
    SingleEvents(true).
    OrderBy("startTime").
    Do()

service.Events.List(calendarId).
    SyncToken(lastToken).
    Do()
```

**Pagination:**
- Max 250 events per request
- Use `PageToken` for next page
- Loop until `NextPageToken` is empty

**Rate Limits:**
- 1M requests/day (very generous)
- No special throttling needed

### OAuth Scope

Already configured in Phase 1:
```
https://www.googleapis.com/auth/calendar.readonly
```

## Database Operations

### Sync State

**On sync start:**
```sql
UPDATE sync_state
SET status = 'syncing',
    error_message = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE service = 'calendar';
```

**On sync complete:**
```sql
UPDATE sync_state
SET status = 'idle',
    last_sync_time = CURRENT_TIMESTAMP,
    last_sync_token = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE service = 'calendar';
```

### Duplicate Prevention

**Check before import:**
```sql
SELECT entity_id FROM sync_log
WHERE source_service = 'calendar'
  AND source_id = ?
```

If exists, skip event (already imported).

### Contact Updates

**Use existing `ContactMatcher` from Phase 1:**
- Check if contact exists by email
- If exists: update name if empty
- If not exists: create new contact
- Return contact ID for interaction logging

## CLI Commands

### New Commands

```bash
pagen sync calendar              # Incremental sync
pagen sync calendar --initial    # Initial sync (6 months)
```

### Updated Commands

```bash
pagen sync                       # Syncs contacts + calendar
pagen sync --initial             # Initial sync for both
pagen sync status                # Show last sync times (now includes calendar)
```

### Command Implementation

**File:** `cli/sync.go`

```go
func SyncCalendarCommand(database *sql.DB, args []string) error {
    // Parse --initial flag
    initial := hasFlag(args, "--initial")

    // Load OAuth token
    token, err := sync.LoadToken()
    if err != nil {
        return fmt.Errorf("not authenticated. Run 'pagen sync init' first")
    }

    // Create Calendar API client
    client := sync.NewCalendarClient(token)

    // Run import
    return sync.ImportCalendar(database, client, initial)
}
```

## User Experience

### Initial Sync

```bash
$ pagen sync calendar --initial

Syncing Google Calendar (last 6 months)...
  → Fetching events...
  ✓ Fetched 156 events
  ✓ Skipped 23 all-day events
  ✓ Skipped 8 declined events
  ✓ Skipped 12 solo events

  → Processing 113 meetings...
  ✓ Created 45 new contacts from attendees
  ✓ Logged 113 interactions
  ✓ Updated cadence for 67 contacts

Summary:
  Events imported: 113
  Contacts created: 45
  Interactions logged: 113

Sync token saved. Next sync will be incremental.
```

### Incremental Sync

```bash
$ pagen sync calendar

Syncing Google Calendar...
  ✓ 3 new events since last sync
  ✓ 1 updated event
  ✓ Logged 4 interactions

All up to date!
```

### Combined Sync

```bash
$ pagen sync

Syncing Google Contacts...
  ✓ No changes since last sync

Syncing Google Calendar...
  ✓ 3 new events
  ✓ Logged 3 interactions

Summary:
  Interactions: 3 logged

All up to date!
```

## Error Handling

### Calendar API Not Enabled

```
Error: Google Calendar API not enabled

To fix:
1. Visit: https://console.cloud.google.com/apis/library/calendar-json.googleapis.com
2. Enable the Calendar API
3. Run sync again
```

### Sync Token Expired

```
Warning: Sync token expired (you may have deleted events)
Falling back to time-based incremental sync...
✓ Resuming sync from last sync time
```

### Network Errors

```
Error: Failed to fetch calendar events
Retrying in 5 seconds (attempt 1/3)...
```

### Partial Failures

```
Syncing Google Calendar...
  → Fetched 50 events
  ✗ Failed to process event "Team Meeting" (skipping)

  ✓ Processed 49/50 events
  ✓ Logged 49 interactions

Summary:
  Events imported: 49
  Errors: 1 (see logs)

Run 'pagen sync calendar' to retry failed events.
```

## Testing Strategy

### Unit Tests (`sync/calendar_importer_test.go`)

**Test cases:**
1. Event filtering (all-day, declined, cancelled, solo)
2. Attendee extraction (skip self, extract others)
3. Event → interaction mapping
4. Metadata JSON generation
5. Duplicate detection via sync_log

**Mock Calendar API responses:**
- Create fixtures with sample Calendar API JSON
- Test pagination handling
- Test sync token refresh

### Integration Tests

**Test script:** `.scratch/test_google_sync_phase2.sh`

1. Verify calendar sync tables exist
2. Test initial sync (--initial flag)
3. Verify interactions created in DB
4. Test incremental sync (no duplicates)
5. Verify sync token stored correctly

### Manual Testing

**Checklist:**
- [ ] OAuth flow works (use existing token from Phase 1)
- [ ] Initial sync imports last 6 months
- [ ] All-day events skipped
- [ ] Declined events skipped
- [ ] Solo events skipped
- [ ] Attendees become contacts
- [ ] Interactions logged with correct timestamp
- [ ] Incremental sync only fetches new events
- [ ] No duplicate interactions on re-sync
- [ ] Contact cadence updated after interaction

## Implementation Tasks

### Task 1: Calendar API Client Setup
- Add Calendar API scope to OAuth config (already done in Phase 1)
- Create helper function to build Calendar service from token
- Test API connection

### Task 2: Calendar Importer Core
- Implement `ImportCalendar()` function
- Handle pagination
- Store sync token
- Add progress logging

### Task 3: Event Filtering
- Implement skip logic (all-day, declined, cancelled, solo)
- Add tests for filtering edge cases

### Task 4: Attendee → Contact Mapping
- Extract attendees from events
- Use `ContactMatcher` for deduplication
- Create new contacts if needed
- Return contact IDs

### Task 5: Event → Interaction Logging
- Create interaction_log entries
- Store metadata JSON (location, duration, attendee count)
- Link to contacts via contact_id
- Update contact cadence

### Task 6: Sync Log Tracking
- Record imported events in sync_log
- Check for duplicates before import
- Handle event updates (same event ID, different content)

### Task 7: CLI Command
- Add `pagen sync calendar` command
- Support `--initial` flag
- Wire up to main.go router

### Task 8: Integration Tests
- Create test script
- Add unit tests for importer
- Test with mock Calendar API responses

### Task 9: Documentation
- Update README with calendar sync instructions
- Document event filtering rules
- Add troubleshooting section

## Success Criteria

**Functional:**
- [ ] Initial sync imports 6 months of events
- [ ] Incremental sync only fetches new/changed events
- [ ] No duplicate interactions created on re-sync
- [ ] All-day events properly skipped
- [ ] Declined events properly skipped
- [ ] Solo events properly skipped
- [ ] Attendees become contacts (no duplicates)
- [ ] Contact cadence auto-updates after interactions

**Performance:**
- [ ] Initial sync completes in <30 seconds for 200 events
- [ ] Incremental sync completes in <5 seconds

**Quality:**
- [ ] All tests passing
- [ ] Pre-commit hooks passing
- [ ] Error messages are actionable

## Future Enhancements (Not Phase 2)

Defer to later phases:
- Deal detection from event titles (Phase 4)
- Suggestions creation (Phase 4)
- TUI integration (Phase 4)
- Recurring event handling improvements
- Calendar → Deal stage progression
- Event description analysis
- Automatic follow-up suggestions from meeting notes

## Dependencies

**New packages:**
```
google.golang.org/api/calendar/v3  # Google Calendar API
```

**Existing packages (from Phase 1):**
```
golang.org/x/oauth2                 # OAuth (already added)
github.com/adrg/xdg                 # XDG paths (already added)
```

## Open Questions

1. **Should we import recurring events as separate interactions or link them?**
   - Decision: Import each occurrence separately (matches reality)

2. **Should we update existing interactions if event details change?**
   - Decision: No updates for now. Sync log prevents re-import. User can manually edit if needed.

3. **Should we track event response status (accepted/tentative)?**
   - Decision: Not yet. Just track attendance. Can add later if useful.

4. **Should we extract location into a structured field?**
   - Decision: Store in metadata JSON for now. No separate location table yet.
