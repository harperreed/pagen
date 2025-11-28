#!/bin/bash
# ABOUTME: Integration test script for Google Sync Phase 2
# ABOUTME: Verifies Calendar sync schema, metadata column, and provides manual testing instructions

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database location (XDG standard)
DB_PATH="${XDG_DATA_HOME:-$HOME/.local/share}/crm/crm.db"

echo "=================================="
echo "Google Sync Phase 2 - Integration Test"
echo "=================================="
echo

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}✗ FAIL${NC}: Database not found at $DB_PATH"
    echo "  Please run pagen to initialize the database first"
    exit 1
fi

echo -e "${GREEN}✓${NC} Database found at: $DB_PATH"
echo

# Function to check if a table exists
check_table() {
    local table_name=$1
    local result=$(sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='$table_name';")

    if [ -z "$result" ]; then
        echo -e "${RED}✗ FAIL${NC}: Table '$table_name' does not exist"
        return 1
    else
        echo -e "${GREEN}✓ PASS${NC}: Table '$table_name' exists"
        return 0
    fi
}

# Function to check if a column exists in a table
check_column() {
    local table_name=$1
    local column_name=$2
    local result=$(sqlite3 "$DB_PATH" "PRAGMA table_info($table_name);" | grep -i "$column_name")

    if [ -z "$result" ]; then
        echo -e "${RED}✗ FAIL${NC}: Column '$column_name' does not exist in table '$table_name'"
        return 1
    else
        echo -e "${GREEN}✓ PASS${NC}: Column '$column_name' exists in table '$table_name'"
        return 0
    fi
}

# Function to show table schema
show_schema() {
    local table_name=$1
    echo -e "${BLUE}  Schema:${NC}"
    sqlite3 "$DB_PATH" ".schema $table_name" | sed 's/^/    /'
    echo
}

# Track overall success
ALL_PASSED=true

echo "Checking Phase 1 prerequisites..."
echo "----------------------------------"

# Check sync_state table (from Phase 1)
if check_table "sync_state"; then
    show_schema "sync_state"
else
    ALL_PASSED=false
fi

# Check sync_log table (from Phase 1)
if check_table "sync_log"; then
    show_schema "sync_log"
else
    ALL_PASSED=false
fi

echo "Checking Phase 2 requirements..."
echo "---------------------------------"

# Check interaction_log table has metadata column
if check_table "interaction_log"; then
    if check_column "interaction_log" "metadata"; then
        echo -e "${BLUE}  This column stores Calendar event metadata as JSON${NC}"
        echo -e "${BLUE}  Format: {location, duration_minutes, attendee_count, calendar_event_id}${NC}"
        echo
    else
        ALL_PASSED=false
    fi
    show_schema "interaction_log"
else
    ALL_PASSED=false
fi

# Print overall result
echo "=================================="
if [ "$ALL_PASSED" = true ]; then
    echo -e "${GREEN}✓ ALL SCHEMA CHECKS PASSED${NC}"
    echo

    # Show table row counts
    echo "Table Statistics:"
    echo "-----------------"
    sync_state_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sync_state;")
    sync_log_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sync_log;")
    interaction_log_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM interaction_log;")
    contacts_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM contacts;")

    echo "  sync_state:      $sync_state_count rows"
    echo "  sync_log:        $sync_log_count rows"
    echo "  interaction_log: $interaction_log_count rows"
    echo "  contacts:        $contacts_count rows"
    echo

    # Show calendar sync state if exists
    calendar_state=$(sqlite3 "$DB_PATH" "SELECT service, status, last_sync_time FROM sync_state WHERE service='calendar';" 2>/dev/null || echo "")
    if [ -n "$calendar_state" ]; then
        echo "Calendar Sync Status:"
        echo "---------------------"
        echo "$calendar_state" | awk -F'|' '{printf "  Service: %s\n  Status: %s\n  Last Sync: %s\n", $1, $2, $3}'
        echo
    fi
else
    echo -e "${RED}✗ SOME SCHEMA CHECKS FAILED${NC}"
    echo
    exit 1
fi

# Database Integrity Checks
echo "=================================="
echo "Database Integrity Checks"
echo "=================================="
echo

# Check foreign key constraints
echo -e "${YELLOW}Checking foreign key constraints...${NC}"
fk_violations=$(sqlite3 "$DB_PATH" "PRAGMA foreign_key_check;" 2>&1)
if [ -z "$fk_violations" ]; then
    echo -e "${GREEN}✓ PASS${NC}: No foreign key violations"
else
    echo -e "${RED}✗ FAIL${NC}: Foreign key violations found:"
    echo "$fk_violations"
    ALL_PASSED=false
fi
echo

# Check interaction_log references valid contacts
echo -e "${YELLOW}Checking interaction_log → contacts references...${NC}"
orphaned_interactions=$(sqlite3 "$DB_PATH" "
    SELECT COUNT(*) FROM interaction_log il
    WHERE NOT EXISTS (
        SELECT 1 FROM contacts c WHERE c.id = il.contact_id
    );
")
if [ "$orphaned_interactions" -eq 0 ]; then
    echo -e "${GREEN}✓ PASS${NC}: All interactions reference valid contacts"
else
    echo -e "${RED}✗ FAIL${NC}: Found $orphaned_interactions orphaned interactions"
    ALL_PASSED=false
fi
echo

# Check sync_log references valid entities
echo -e "${YELLOW}Checking sync_log → interaction_log references...${NC}"
calendar_syncs=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sync_log WHERE source_service='calendar';")
orphaned_syncs=$(sqlite3 "$DB_PATH" "
    SELECT COUNT(*) FROM sync_log sl
    WHERE sl.source_service = 'calendar'
    AND NOT EXISTS (
        SELECT 1 FROM interaction_log il WHERE il.id = sl.entity_id
    );
")
if [ "$calendar_syncs" -eq 0 ]; then
    echo -e "${YELLOW}⚠ INFO${NC}: No calendar syncs recorded yet (run 'pagen sync calendar --initial')"
elif [ "$orphaned_syncs" -eq 0 ]; then
    echo -e "${GREEN}✓ PASS${NC}: All calendar sync_log entries reference valid interactions"
else
    echo -e "${RED}✗ FAIL${NC}: Found $orphaned_syncs orphaned sync_log entries"
    ALL_PASSED=false
fi
echo

# Manual Testing Instructions
echo "=================================="
echo "Manual Calendar Sync Testing"
echo "=================================="
echo
echo -e "${YELLOW}Prerequisites:${NC}"
echo "  1. OAuth must be configured (from Phase 1)"
echo "  2. Run: pagen sync init (if not done already)"
echo "  3. Complete OAuth flow in browser"
echo
echo -e "${YELLOW}Step 1: Initial Calendar Sync${NC}"
echo "  # Import last 6 months of calendar events"
echo "  pagen sync calendar --initial"
echo
echo "  Expected behavior:"
echo "  - Fetches all events from last 6 months"
echo "  - Skips all-day events"
echo "  - Skips declined events"
echo "  - Skips solo events (no attendees)"
echo "  - Skips cancelled events"
echo "  - Creates contacts from event attendees"
echo "  - Logs interactions with metadata"
echo "  - Stores sync token for incremental sync"
echo
echo -e "${YELLOW}Step 2: Verify Imported Data${NC}"
echo "  # Check sync state"
echo "  sqlite3 $DB_PATH \"SELECT * FROM sync_state WHERE service='calendar';\""
echo
echo "  # Check imported interactions"
echo "  sqlite3 $DB_PATH \"SELECT id, contact_id, interaction_type, timestamp, notes FROM interaction_log ORDER BY timestamp DESC LIMIT 5;\""
echo
echo "  # Check metadata structure"
echo "  sqlite3 $DB_PATH \"SELECT metadata FROM interaction_log WHERE metadata IS NOT NULL LIMIT 1;\""
echo
echo "  # Expected metadata format:"
echo "  # {\"location\":\"...\",\"duration_minutes\":60,\"attendee_count\":3,\"calendar_event_id\":\"...\"}"
echo
echo -e "${YELLOW}Step 3: Test Incremental Sync${NC}"
echo "  # Run sync again (should only fetch new events)"
echo "  pagen sync calendar"
echo
echo "  Expected behavior:"
echo "  - Uses sync token from database"
echo "  - Only fetches events changed since last sync"
echo "  - No duplicate interactions created"
echo "  - Fast (< 5 seconds)"
echo
echo -e "${YELLOW}Step 4: Verify Duplicate Prevention${NC}"
echo "  # Check sync_log for duplicate detection"
echo "  sqlite3 $DB_PATH \"SELECT source_service, source_id, entity_id FROM sync_log WHERE source_service='calendar' LIMIT 5;\""
echo
echo "  # Run sync again and verify no new duplicates"
echo "  pagen sync calendar"
echo "  sqlite3 $DB_PATH \"SELECT COUNT(*) FROM interaction_log;\""
echo "  # Count should not increase if no new events"
echo
echo -e "${YELLOW}Step 5: Test Contact Creation from Attendees${NC}"
echo "  # Find contacts created from calendar attendees"
echo "  sqlite3 $DB_PATH \"SELECT c.name, c.email FROM contacts c WHERE c.id IN (SELECT DISTINCT contact_id FROM interaction_log);\""
echo
echo "  # Verify attendees became contacts (no duplicates)"
echo "  # Check that contact names are populated from attendee.DisplayName"
echo
echo -e "${YELLOW}Step 6: Test Event Filtering${NC}"
echo "  # Manually create test events in Google Calendar:"
echo "  # - All-day event (should be skipped)"
echo "  # - Solo event with no attendees (should be skipped)"
echo "  # - Event you declined (should be skipped)"
echo "  # - Cancelled event (should be skipped)"
echo "  # - Normal multi-person meeting (should be imported)"
echo
echo "  # Run incremental sync"
echo "  pagen sync calendar"
echo
echo "  # Verify only the normal meeting was imported"
echo "  sqlite3 $DB_PATH \"SELECT COUNT(*) FROM sync_log WHERE source_service='calendar' AND created_at > datetime('now', '-5 minutes');\""
echo "  # Should show 1 new entry"
echo
echo -e "${YELLOW}Step 7: Test Combined Sync${NC}"
echo "  # Run both contacts and calendar sync"
echo "  pagen sync"
echo
echo "  Expected behavior:"
echo "  - Syncs both contacts and calendar"
echo "  - Shows summary for both services"
echo "  - Both complete successfully"
echo
echo "=================================="
echo "Troubleshooting"
echo "=================================="
echo
echo -e "${YELLOW}If sync fails:${NC}"
echo "  1. Check OAuth token is valid: pagen sync status"
echo "  2. Verify Calendar API is enabled in Google Cloud Console"
echo "  3. Check error messages in sync_state table:"
echo "     sqlite3 $DB_PATH \"SELECT error_message FROM sync_state WHERE service='calendar';\""
echo "  4. Re-run with verbose logging (if available)"
echo
echo -e "${YELLOW}If duplicates are created:${NC}"
echo "  1. Check sync_log for duplicate source_id entries"
echo "  2. Verify sync token is being stored correctly"
echo "  3. Clear and re-sync:"
echo "     sqlite3 $DB_PATH \"DELETE FROM interaction_log WHERE id IN (SELECT entity_id FROM sync_log WHERE source_service='calendar');\""
echo "     sqlite3 $DB_PATH \"DELETE FROM sync_log WHERE source_service='calendar';\""
echo "     sqlite3 $DB_PATH \"UPDATE sync_state SET last_sync_token=NULL WHERE service='calendar';\""
echo "     pagen sync calendar --initial"
echo
echo -e "${YELLOW}If metadata is malformed:${NC}"
echo "  1. Check JSON structure:"
echo "     sqlite3 $DB_PATH \"SELECT metadata FROM interaction_log WHERE metadata IS NOT NULL LIMIT 1;\""
echo "  2. Verify it's valid JSON (should parse cleanly)"
echo "  3. Check for expected fields: location, duration_minutes, attendee_count, calendar_event_id"
echo
echo "=================================="
if [ "$ALL_PASSED" = true ]; then
    echo -e "${GREEN}✓ Phase 2 Schema Tests: COMPLETE${NC}"
else
    echo -e "${RED}✗ Phase 2 Schema Tests: FAILED${NC}"
    exit 1
fi
echo "=================================="
