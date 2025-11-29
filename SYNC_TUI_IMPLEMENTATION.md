# Sync TUI Implementation Summary

## Overview
Added a new "Sync" tab to the interactive TUI that displays Google sync status and allows triggering syncs for contacts, calendar, and Gmail services.

## Changes Made

### 1. Core TUI Structure (`tui/tui.go`)
- Added `EntitySync` to the `EntityType` enum
- Added sync-specific state to the `Model` struct:
  - `syncStates []SyncStateDisplay` - Display state for each service
  - `syncInProgress map[string]bool` - Tracks active syncs
  - `syncMessages []string` - Activity log
  - `selectedService int` - Currently selected service in UI
- Created `SyncStateDisplay` struct for UI representation
- Updated `NewModel()` to initialize sync-related maps
- Updated `Update()` to handle `SyncCompleteMsg` and `SyncStartMsg`

### 2. Sync View (`tui/sync_view.go`)
New file implementing the sync view with:

#### Display Features
- Service status table showing:
  - Service name (Calendar, Contacts, Gmail)
  - Current status (Idle, Syncing, Error)
  - Last sync time (human-readable, e.g., "5 minutes ago")
  - Error messages if applicable
- Recent activity log (last 5 messages)
- Visual indicators:
  - ✓ Green for idle/successful syncs
  - ⟳ Yellow for in-progress syncs
  - ✗ Red for errors
  - ▶ Selection indicator

#### Interactive Controls
- **↑/↓ or k/j**: Navigate between services
- **Enter**: Sync selected service
- **a**: Sync all services
- **r**: Refresh sync status
- **Esc**: Return to main view
- **q**: Quit application

#### Background Sync Implementation
- Syncs run in background using bubbletea commands
- Real-time status updates via messages
- Progress tracking with in-progress map
- Error handling with database status updates
- Activity logging with timestamps

### 3. List View Integration (`tui/list_view.go`)
- Added "Sync" to tab bar
- Updated `renderTabs()` to include 5th tab
- Updated `renderTable()` to delegate to `renderSyncView()` for EntitySync
- Updated `handleListKeys()`:
  - Tab navigation now cycles through 5 tabs
  - Added "s" shortcut to jump to sync tab
  - Delegates to `handleSyncKeys()` when in sync view
- Updated help text to include sync shortcut

### 4. Bug Fixes
Fixed compilation errors in existing code:
- `sync/gmail_importer.go`: Fixed `err` variable redeclaration (changed to `syncErr`)
- `cli/sync.go`: Fixed log.Printf format string for DBStats (changed to `%+v`)

### 5. Tests (`tui/sync_view_test.go`)
Added comprehensive test coverage:
- `TestSyncViewRendering`: Verifies basic view rendering
- `TestSyncViewWithStates`: Tests display with sync states in database
- `TestSyncKeyNavigation`: Tests keyboard navigation
- `TestSyncCompleteMessage`: Tests successful sync completion
- `TestSyncCompleteWithError`: Tests error handling
- `TestFormatTimeSince`: Tests time formatting utility
- `TestSyncMessageAddition`: Tests activity log

All tests pass ✓

## Usage

### Accessing the Sync View
1. Launch TUI: `./pagen` (or just `pagen`)
2. Press **Tab** to cycle to the Sync tab, or press **s** to jump directly
3. Use arrow keys to select a service
4. Press **Enter** to sync the selected service, or **a** to sync all

### Sync View States

#### No Sync Data
```
Google Sync Management

No sync data found. Run sync initialization first.

Press 'i' to initialize sync, 'Esc' to go back
```

#### With Sync States
```
Google Sync Management

Service Status

▶ Calendar     ✓ Idle • Last synced 2 hours ago
  Contacts     ✓ Idle • Last synced 5 minutes ago
  Gmail        ⟳ Syncing...

Recent Activity

  [14:32:15] Starting gmail sync...
  [14:31:22] ✓ contacts sync completed
  [14:30:45] Starting contacts sync...

↑/↓: Select service • Enter: Sync selected • a: Sync all • r: Refresh status • Esc: Back • q: Quit
```

#### Error State
```
Google Sync Management

Service Status

  Calendar     ✓ Idle • Last synced 2 hours ago
▶ Contacts     ✓ Idle • Last synced 5 minutes ago
  Gmail        ✗ Error: Authentication failed

Recent Activity

  [14:35:10] ✗ gmail sync failed: no authentication token found
  [14:32:15] ✓ contacts sync completed

↑/↓: Select service • Enter: Sync selected • a: Sync all • r: Refresh status • Esc: Back • q: Quit
```

## Technical Details

### Sync Execution Flow
1. User presses Enter on a service
2. `handleSyncKeys()` calls `syncService()`
3. `syncService()` returns a `tea.Cmd` that:
   - Marks service as in-progress
   - Updates database status to "syncing"
   - Loads OAuth token
   - Creates appropriate API client
   - Runs sync operation
   - Returns `SyncCompleteMsg`
4. `Update()` receives `SyncCompleteMsg`
5. `handleSyncComplete()`:
   - Clears in-progress flag
   - Adds activity message
   - Updates database with result
   - Reloads sync states

### Database Integration
Uses existing `db` package functions:
- `db.GetAllSyncStates()` - Load all service states
- `db.UpdateSyncStatus()` - Update status during sync
- `db.UpdateSyncToken()` - Update token after successful sync

### OAuth Integration
Uses existing `sync` package:
- `sync.LoadToken()` - Load OAuth credentials
- `sync.NewPeopleClient()` - Create Contacts API client
- `sync.NewCalendarClient()` - Create Calendar API client
- `sync.NewGmailClient()` - Create Gmail API client
- `sync.ImportContacts()` - Sync contacts
- `sync.ImportCalendar()` - Sync calendar
- `sync.ImportGmail()` - Sync Gmail

## Key Design Decisions

1. **Non-blocking syncs**: Syncs run as background commands to keep UI responsive
2. **Real-time updates**: Uses bubbletea message passing for live status updates
3. **Error persistence**: Errors are stored in database for history tracking
4. **Activity log**: Recent messages provide context and progress visibility
5. **Consistent styling**: Reuses existing TUI styles and patterns
6. **Keyboard-driven**: All operations accessible via keyboard shortcuts

## Future Enhancements

Potential improvements:
- Add progress bars for individual syncs
- Show record counts (e.g., "Synced 150 contacts")
- Add sync scheduling/automation UI
- Show last sync duration
- Add filtering/search for activity log
- Visual sync history graph
- OAuth re-authentication from TUI

## Files Modified
- `tui/tui.go` - Core model updates
- `tui/list_view.go` - Tab integration
- `sync/gmail_importer.go` - Bug fix
- `cli/sync.go` - Bug fix

## Files Created
- `tui/sync_view.go` - Sync view implementation
- `tui/sync_view_test.go` - Comprehensive tests
- `SYNC_TUI_IMPLEMENTATION.md` - This document

## Testing
All tests pass:
```
go test ./...
ok      github.com/harperreed/pagen/cli
ok      github.com/harperreed/pagen/db
ok      github.com/harperreed/pagen/handlers
ok      github.com/harperreed/pagen/models
ok      github.com/harperreed/pagen/sync
ok      github.com/harperreed/pagen/tui
ok      github.com/harperreed/pagen/viz
```

Build successful:
```
go build -o pagen .
```
