# Google Integration Design

**Date:** 2025-11-28
**Goal:** Transform pagen from manual CRM to automated relationship intelligence engine by syncing with Google Contacts, Calendar, and Gmail

## Problem Statement

Currently, pagen requires manual data entry for all contacts, interactions, and deals. This creates friction and means the CRM falls out of date quickly. Users already have rich relationship data in their Google ecosystem (contacts, calendar events, email conversations) that should automatically populate and update pagen.

## Solution Overview

Build a local-only sync engine that:
1. Imports contacts from Google Contacts (curated foundation)
2. Logs interactions from Google Calendar events (high-signal meetings)
3. Enriches from Gmail (email conversations you actually care about)
4. Auto-discovers deals and relationships using intelligent pattern matching
5. Surfaces suggestions for fuzzy discoveries (you review and approve)

All data stored locally in SQLite. Manual sync on-demand via `pagen sync` command.

## Core Principles

- **Local-only:** All data in `~/.local/share/pagen/`, OAuth tokens in `~/.local/share/pagen/google-credentials.json`
- **Manual sync:** User runs `pagen sync` when they want fresh data (no background daemons)
- **High-signal only:** Skip email noise, focus on replied messages and calendar events
- **Certain vs fuzzy:** Auto-import obvious stuff (calendar attendees → contacts), suggest uncertain stuff (email mentions money → maybe a deal)
- **No duplicates:** Smart matching prevents creating same contact twice from different sources

## Data Model

### New Tables

**sync_state** - Track last sync per Google service
```sql
CREATE TABLE sync_state (
    service TEXT PRIMARY KEY,        -- 'gmail', 'calendar', 'contacts'
    last_sync_time TIMESTAMP,
    last_sync_token TEXT,            -- For incremental syncs (Calendar syncToken, Gmail historyId)
    status TEXT,                     -- 'idle', 'syncing', 'error'
    error_message TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**sync_log** - Prevent re-importing same items
```sql
CREATE TABLE sync_log (
    id TEXT PRIMARY KEY,
    source_service TEXT,             -- 'gmail', 'calendar', 'contacts'
    source_id TEXT,                  -- Google's ID (message ID, event ID, contact resourceName)
    entity_type TEXT,                -- 'contact', 'interaction', 'company'
    entity_id TEXT,                  -- Our pagen entity ID (references contacts/companies/etc)
    imported_at TIMESTAMP,
    metadata TEXT,                   -- JSON blob for service-specific data
    UNIQUE(source_service, source_id)
);

CREATE INDEX idx_sync_log_source ON sync_log(source_service, source_id);
CREATE INDEX idx_sync_log_entity ON sync_log(entity_type, entity_id);
```

**suggestions** - Hold fuzzy discoveries for review
```sql
CREATE TABLE suggestions (
    id TEXT PRIMARY KEY,
    type TEXT,                       -- 'deal', 'relationship', 'company'
    confidence REAL,                 -- 0.0-1.0 how sure we are
    source_service TEXT,             -- 'gmail', 'calendar', 'contacts'
    source_id TEXT,                  -- Link back to source (event ID, message ID)
    source_data TEXT,                -- JSON of extracted data (event title, email subject, etc)
    status TEXT,                     -- 'pending', 'accepted', 'rejected'
    created_at TIMESTAMP,
    reviewed_at TIMESTAMP
);

CREATE INDEX idx_suggestions_status ON suggestions(status);
CREATE INDEX idx_suggestions_type ON suggestions(type);
```

### Existing Tables (no changes)

Leverage existing schema:
- `contacts` - populated from Google Contacts, Calendar attendees, Gmail senders
- `companies` - extracted from email domains and Google Contacts org field
- `interaction_log` - Calendar events and Gmail replies logged here
- `contact_cadence` - Auto-updated when interactions imported

## Architecture

### Sync Command Structure

```
pagen sync [service]           # Sync specific service or all
pagen sync init                # First-time OAuth setup
pagen sync contacts            # Just Google Contacts
pagen sync calendar            # Just Google Calendar
pagen sync gmail               # Just Gmail
pagen sync --initial           # Full 6-month import (first time)
pagen sync status              # Show last sync times and errors
pagen sync review              # Review pending suggestions
pagen sync accept <id>         # Accept a suggestion
pagen sync reject <id>         # Reject a suggestion
```

### Core Components

**1. Sync Orchestrator** (`sync/orchestrator.go`)
- Coordinates which services to sync
- Handles OAuth token refresh (using `golang.org/x/oauth2`)
- Manages sync_state table
- Reports progress to user

**2. Service Importers** (one per Google service)

`sync/contacts_importer.go`:
- Fetches from Google People API: `people.connections.list`
- Creates contacts with: name, email, phone, company, job title, notes
- Deduplicates using email matching
- Pagination: 1000 contacts per request

`sync/calendar_importer.go`:
- Fetches from Google Calendar API: `events.list`
- For each event:
  - Create/update contacts from attendees (except self)
  - Log interaction (type=meeting, timestamp=event start)
  - Pass to analyzer for deal detection
- Incremental sync using `updatedMin` parameter
- Skips: all-day events, declined events, cancelled events

`sync/gmail_importer.go`:
- Fetches from Gmail API: `users.messages.list`
- **Filters applied:**
  - Only messages you replied to: `is:sent OR (in:inbox -is:sent)`
  - Exclude spam/trash: `-in:spam -in:trash`
  - Minimum thread length: 2+ messages (back-and-forth)
- For each message:
  - Create/update contacts from sender/recipients
  - Log interaction (type=email, timestamp=message date)
  - Pass to analyzer for deal detection
- Pagination: 100 messages per request

**3. Entity Matcher** (`sync/matcher.go`)

Contact matching priority:
1. Exact email match (case-insensitive)
2. Normalized email match (strip dots, lowercase)
3. Name + company fuzzy match (if email unknown)

Merge strategy:
- Prefer Google Contacts data (user-curated)
- Enrich with Gmail/Calendar data (fill missing fields)
- Never overwrite manually entered data

Company detection:
1. Email domain extraction (`alice@acme.com` → `acme.com`)
2. Domain → company name mapping (hardcoded common domains, else capitalize domain)
3. Google Contacts "organization" field (highest priority)
4. Dedupe by normalized name (remove Inc, LLC, Corp, spaces, lowercase)

**4. Intelligence Layer** (`sync/analyzer.go`)

**Deal Detection** (creates suggestions):

High-confidence signals:
- Calendar event title contains: "demo", "proposal", "contract", "sales", "partnership", "pitch"
- Email subject contains: $ amount, "deal", "agreement", "negotiation"
- Multi-company calendar meeting (attendees from 2+ different domains)

Suggestion includes:
- Detected company name
- Suggested deal title (event title or email subject)
- Amount (if parsed from text)
- Stage guess (e.g., "demo" → proposal stage)
- Confidence score

**Relationship Inference** (creates suggestions):

Signals:
- CC'd together on 3+ emails → "colleagues"
- Same company domain → "colleagues"
- Recurring calendar events together → extract context from event titles
- Google Contacts "relationship" field → direct import (no suggestion)

**Company Extraction:**

From email signatures (parse last 4 lines of email body):
- Look for company name after sender name
- Match against known company suffixes (Inc, LLC, Corp, Ltd)
- Extract website URLs (http://acme.com → Acme Corp)

### OAuth & API Integration

**OAuth Flow:**

```bash
pagen sync init
# Opens browser to Google OAuth consent screen
# User grants permissions
# Tokens saved to ~/.local/share/pagen/google-credentials.json
# Auto-refresh handled by oauth2 library
```

**Required Scopes:**
```
https://www.googleapis.com/auth/contacts.readonly
https://www.googleapis.com/auth/calendar.readonly
https://www.googleapis.com/auth/gmail.readonly
```

**Token Storage** (XDG-compliant):
```
~/.local/share/pagen/google-credentials.json
```

Format:
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "expiry": "2025-11-29T10:00:00Z"
}
```

**API Rate Limits:**
- Google Contacts: 600 requests/minute
- Google Calendar: 1M requests/day
- Gmail: 1B quota units/day

Strategy: Batch requests, respect rate limits, retry with exponential backoff on 429 errors.

### Configuration

**Config File** (`~/.config/pagen/google-sync.toml`):

```toml
[sync]
initial_history_days = 180  # 6 months for first sync

[gmail]
enabled = true
reply_only = true           # Only import emails you replied to
min_thread_length = 2       # Require back-and-forth exchanges
max_recipients = 5          # Skip emails to >5 people (likely spam)

[calendar]
enabled = true
import_past_events = true
skip_all_day_events = true
skip_declined_events = true

[contacts]
enabled = true

[filters]
# Domains to ignore (newsletters, notifications)
ignore_domains = [
  "noreply.com",
  "notifications.google.com",
  "no-reply.com"
]

# Keywords that suggest a deal
deal_keywords = [
  "demo",
  "proposal",
  "contract",
  "sales",
  "partnership",
  "pitch",
  "agreement"
]

# Company domain overrides
[company_domains]
"gmail.com" = "skip"        # Don't create company for Gmail users
"google.com" = "Google"
```

## User Experience

### Initial Setup

```bash
# First-time setup
$ pagen sync init

Opening browser for Google OAuth...
✓ Authenticated successfully
✓ Tokens saved to ~/.local/share/pagen/google-credentials.json

Ready to sync! Run 'pagen sync --initial' to import last 6 months.
```

### First Sync

```bash
$ pagen sync --initial

Syncing Google Contacts...
  → Fetching contacts...
  ✓ Fetched 247 contacts
  ✓ Created 12 new contacts
  ✓ Updated 8 existing contacts
  ✓ Skipped 227 duplicates (already in pagen)

Syncing Google Calendar (last 6 months)...
  → Fetching events...
  ✓ Fetched 156 events
  ✓ Created 45 new contacts from attendees
  ✓ Logged 156 interactions
  ✓ Found 3 potential deals

Syncing Gmail (replied messages only)...
  → Fetching threads...
  ✓ Fetched 89 message threads
  ✓ Created 5 new contacts
  ✓ Logged 89 interactions
  ✓ Found 2 potential deals

Summary:
  Contacts: 62 created, 8 updated
  Interactions: 245 logged
  Suggestions: 5 pending review

Run 'pagen sync review' to see suggestions.
```

### Incremental Sync

```bash
$ pagen sync

Syncing Google Contacts...
  ✓ No changes since last sync

Syncing Google Calendar...
  ✓ 3 new events
  ✓ Logged 3 interactions

Syncing Gmail...
  ✓ 7 new messages
  ✓ Logged 7 interactions

Summary:
  Interactions: 10 logged

All up to date!
```

### Review Suggestions

```bash
$ pagen sync review

Pending Suggestions (5):

[1] Deal: "Acme Corp - Product Demo"
    Source: Calendar event (2025-11-15)
    Confidence: 0.85
    Company: Acme Corp
    Amount: Unknown
    Stage: proposal

[2] Deal: "TechStart Partnership Discussion"
    Source: Gmail thread (2025-11-20)
    Confidence: 0.72
    Company: TechStart
    Amount: $50,000 (extracted from email)
    Stage: negotiation

[3] Relationship: John Smith <-> Sarah Johnson
    Type: colleagues
    Confidence: 0.90
    Source: CC'd together on 8 emails

Commands:
  pagen sync accept <id>  - Accept and create entity
  pagen sync reject <id>  - Reject suggestion
  pagen sync accept-all   - Accept all pending
```

### TUI Integration

Add "Suggestions" tab to TUI (press 's' to access):
- List pending suggestions with confidence scores
- Visual indicators for high/medium/low confidence
- Press Enter to view details
- Press 'a' to accept, 'r' to reject
- Press 'A' to accept all

## Error Handling

### Common Errors & Recovery

**OAuth Token Expired:**
```
Error: OAuth token expired
Run 'pagen sync init' to re-authenticate
```

**API Rate Limit Hit:**
```
Warning: Google API rate limit reached
Pausing for 60 seconds...
Resuming sync...
```

**Network Error:**
```
Error: Failed to connect to Google API
Retrying in 5 seconds (attempt 1/3)...
```

**Partial Sync Failure:**
```
Syncing Gmail...
  ✗ Failed to fetch messages: API error 500

Continuing with other services...

Syncing Google Calendar...
  ✓ Success

Error summary:
  Gmail sync failed (see ~/.local/share/pagen/sync.log for details)

Run 'pagen sync gmail' to retry Gmail sync.
```

### Error Storage

Store errors in `sync_state` table:
```sql
UPDATE sync_state
SET status = 'error',
    error_message = 'API timeout after 3 retries'
WHERE service = 'gmail';
```

Errors also logged to `~/.local/share/pagen/sync.log` for debugging.

## Testing Strategy

### Unit Tests

**Contact Matching** (`sync/matcher_test.go`):
- Test email normalization
- Test duplicate detection
- Test name fuzzy matching
- Test company extraction from domains

**Deal Detection** (`sync/analyzer_test.go`):
- Test keyword extraction
- Test amount parsing from text
- Test confidence scoring
- Test stage inference

### Integration Tests

**Mock Google API** (`sync/google_mock.go`):
- Record real API responses as fixtures
- Replay during tests (no network calls)
- Test pagination handling
- Test incremental sync tokens

**Full Sync Flow** (`sync/integration_test.go`):
- Test contacts → calendar → gmail sync order
- Verify no duplicates created on re-sync
- Test error recovery and retry logic
- Verify sync_log prevents re-imports

### Manual Testing

**Test Script** (`.scratch/test_google_sync.sh`):
- Uses test Google account with known data
- Verifies contact import accuracy
- Checks calendar event → interaction mapping
- Validates Gmail filtering (reply-only)
- Confirms suggestion creation

## Implementation Phases

### Phase 1: Foundation (MVP)

**Goal:** OAuth setup and Google Contacts import only

Tasks:
1. Add sync tables to schema (sync_state, sync_log, suggestions)
2. Implement OAuth flow (init command, token storage)
3. Build contacts_importer.go (Google People API)
4. Implement basic entity matcher (email deduplication)
5. Add CLI: `pagen sync init`, `pagen sync contacts`
6. Write unit tests for matcher
7. Create integration test with mock API

**Deliverables:**
- Working Google Contacts import
- No duplicates on re-sync
- Progress reporting in CLI

### Phase 2: Calendar Integration

**Goal:** Import calendar events as interactions

Tasks:
1. Build calendar_importer.go (Google Calendar API)
2. Event → interaction mapping
3. Attendee → contact creation
4. Implement deal detection from event titles
5. Create suggestions for detected deals
6. Add CLI: `pagen sync calendar`
7. Write tests for deal detection logic

**Deliverables:**
- Calendar events logged as interactions
- Deal suggestions created from calendar
- Incremental sync working (updatedMin)

### Phase 3: Gmail Integration

**Goal:** Enrich from email conversations

Tasks:
1. Build gmail_importer.go (Gmail API)
2. Implement reply-only filtering
3. Thread length detection (min 2 messages)
4. Email → interaction logging
5. Enhanced deal detection from email subjects
6. Add CLI: `pagen sync gmail`
7. Write tests for Gmail filtering

**Deliverables:**
- Gmail import working with filters
- Only meaningful conversations imported
- No spam/noise in contacts

### Phase 4: Intelligence Layer

**Goal:** Smart relationship and company detection

Tasks:
1. Implement relationship inference (analyzer.go)
2. Add company extraction from email signatures
3. Build suggestion review workflow
4. Add TUI "Suggestions" tab
5. Add CLI: `pagen sync review`, `pagen sync accept/reject`
6. Write tests for relationship detection

**Deliverables:**
- Relationship suggestions from email patterns
- Company extraction working
- UI for reviewing suggestions

### Phase 5: Polish & Optimization

**Goal:** Production-ready UX

Tasks:
1. Add progress bars for long syncs
2. Implement retry logic with exponential backoff
3. Optimize incremental syncs (use sync tokens properly)
4. Add `pagen sync status` command
5. Improve error messages and recovery
6. Add configuration file support (google-sync.toml)
7. Write comprehensive integration tests

**Deliverables:**
- Polished CLI experience
- Reliable error handling
- Fast incremental syncs
- Full test coverage

## Out of Scope (Future Enhancements)

**Not in this design:**
- Bidirectional sync (updating Google from pagen changes)
- Real-time sync via webhooks/push notifications
- Contact photo import
- Full email body analysis (just metadata/subject for now)
- Gmail labels → tags mapping
- Calendar → Deal stage progression tracking
- Email sentiment analysis
- Automatic follow-up suggestions based on email tone

**Why defer:**
- Bidirectional sync adds complexity (conflict resolution)
- Webhooks require persistent server (against local-only principle)
- Email body analysis requires heavier NLP (privacy concerns)
- Focus on MVP: get data in, let user curate

## Success Metrics

### Quantitative

- Import 95%+ of Google Contacts without duplicates
- Calendar events → interactions within 1% error rate
- Gmail filter reduces noise by 80%+ (no spam/newsletters)
- Incremental sync completes in <10 seconds
- Deal detection precision >70% (user accepts most suggestions)

### Qualitative

- User runs `pagen sync` weekly and trusts the data
- Follow-up tracking "just works" (auto-updated from calendar)
- Suggestions feel helpful, not overwhelming
- Error messages are clear and actionable
- Setup takes <5 minutes (OAuth + first sync)

## Dependencies

**New Go Packages:**
```
golang.org/x/oauth2                  # OAuth flow
google.golang.org/api/people/v1      # Google Contacts
google.golang.org/api/calendar/v3    # Google Calendar
google.golang.org/api/gmail/v1       # Gmail
github.com/pelletier/go-toml/v2      # Config file parsing
```

**Already in go.mod:**
```
github.com/adrg/xdg                  # XDG path handling
github.com/google/uuid               # UUID generation
github.com/mattn/go-sqlite3          # Database
```

## Security Considerations

**Token Storage:**
- OAuth tokens stored in `~/.local/share/pagen/` with 600 permissions
- Never log tokens or send over network
- Auto-refresh tokens silently (user doesn't see them)

**Data Privacy:**
- All data stays local (no cloud sync)
- Gmail content not stored (only metadata: sender, recipient, subject, date)
- No analytics or telemetry

**API Scopes:**
- Request minimal scopes (readonly only)
- Never request write access to Google data
- Clear consent screen showing what data is accessed

## Open Questions

1. **Should we parse email signatures to extract company info?**
   - Pro: Better company data
   - Con: Parsing is fragile, many formats
   - **Decision:** Yes, but make it optional (config flag)

2. **How aggressive should deal detection be?**
   - Currently: Only high-confidence keywords
   - Alternative: Use LLM to analyze event/email context
   - **Decision:** Start conservative, can add LLM layer in Phase 5

3. **Should we import Google Contacts groups as tags?**
   - Pro: Leverage user's existing organization
   - Con: pagen doesn't have contact tags yet
   - **Decision:** Defer until we add tag support to pagen

4. **Handle recurring calendar events as one interaction or many?**
   - Option A: One interaction per occurrence
   - Option B: One interaction with "recurring" flag
   - **Decision:** One per occurrence (more accurate interaction frequency)
