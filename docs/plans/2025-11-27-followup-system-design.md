# Follow-Up System Design

**Date:** 2025-11-27
**Goal:** Help users maintain their personal network by tracking and reminding them to follow up with contacts

## Problem Statement

The biggest challenge with personal networking is forgetting to follow up with people. Users lose touch with valuable relationships because they lack:
- Visibility into who needs attention
- Automatic reminders across multiple touchpoints
- Intelligence about relationship priority and health

## Solution Overview

A comprehensive follow-up tracking system that:
1. Tracks interaction cadence and relationship strength per contact
2. Computes priority scores to surface who needs attention most
3. Surfaces follow-ups across all interfaces (TUI, Web, CLI, MCP)
4. Provides proactive reminders via digest, notifications, and Claude integration
5. Offers network health insights and analytics

## Data Model

### New Tables

**contact_cadence:**
```sql
CREATE TABLE contact_cadence (
    contact_id TEXT PRIMARY KEY,
    cadence_days INTEGER DEFAULT 30,
    relationship_strength TEXT CHECK(relationship_strength IN ('weak', 'medium', 'strong')),
    priority_score REAL,
    last_interaction_date TIMESTAMP,
    next_followup_date TIMESTAMP,
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);
```

**interaction_log:**
```sql
CREATE TABLE interaction_log (
    id TEXT PRIMARY KEY,
    contact_id TEXT NOT NULL,
    interaction_type TEXT CHECK(interaction_type IN ('meeting', 'call', 'email', 'message', 'event')),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    sentiment TEXT CHECK(sentiment IN ('positive', 'neutral', 'negative')),
    FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
);
```

### Priority Scoring Algorithm

Priority score (0-100) computed from:
- **Days overdue:** `max(0, days_since_last_contact - cadence_days) * 2`
- **Relationship multiplier:** weak=1x, medium=1.5x, strong=2x
- **Interaction consistency:** Bonus for historically consistent contact patterns

Formula: `(days_overdue * 2) * relationship_multiplier + consistency_bonus`

### Cadence Logic

- **Default cadence:** 30 days for new contacts
- **Configurable per contact:** Users can set custom cadence (7, 14, 30, 60, 90 days)
- **Auto-adjustment:** Relationship strength influences default (strong=14d, medium=30d, weak=60d)
- **Auto-update:** Logging an interaction resets `last_interaction_date` and recalculates `next_followup_date`

## Interface Integration

### TUI Enhancements

**New "Follow-Ups" Tab:**
- Accessible via 'f' key or Tab navigation
- Columns: Name | Days Since | Priority | Relationship | Next Action
- Sorted by priority score (highest first)
- Visual indicators:
  - 🔴 Overdue (>7 days past cadence)
  - 🟡 Due Soon (within 3 days of cadence)
  - 🟢 On Track

**Keyboard Shortcuts:**
- `l` - Quick log interaction (updates timestamp, clears from overdue list)
- `c` - Adjust cadence for selected contact
- `s` - Set relationship strength
- `/` - Filter by relationship strength or priority threshold

**Startup Banner:**
```
⚠️  You have 3 overdue follow-ups (press 'f' to view)
```

### Web Dashboard

**New `/followups` Page:**
- Prioritized table view (same data as TUI)
- Filter controls: Relationship strength, Priority threshold, Overdue only
- Inline "Log Interaction" button (HTMX partial update)
- Search/filter by contact name

**Dashboard Widget:**
- "Top 5 Follow-Ups" section on homepage
- Shows highest priority contacts with quick action buttons

**Contact Detail Page Enhancement:**
- Interaction timeline visualization
- Next follow-up date display
- Cadence setting controls
- Relationship strength selector

### CLI Commands

```bash
# List follow-ups
pagen followups list [--overdue-only] [--limit 10] [--strength weak|medium|strong]

# Log interaction
pagen followups log --contact "Alice" --type meeting [--notes "Coffee chat"] [--sentiment positive]

# Set cadence
pagen followups set-cadence --contact "Bob" --days 14

# Set relationship strength
pagen followups set-strength --contact "Carol" --strength strong

# Daily digest
pagen followups digest [--format text|json|html]

# Stats
pagen followups stats
```

### MCP Tools for Claude

**New Tools:**
- `get_followup_list` - Returns prioritized contacts needing follow-up with context
- `log_interaction` - Enhanced interaction logging with auto-follow-up updates
- `suggest_followup` - AI suggests who to contact based on conversation context
- `get_network_health` - Returns relationship health metrics

**Smart Integration:**
When user mentions networking or specific people in Claude:
- Check pagen for overdue follow-ups with those contacts
- Surface contextual reminders: "You haven't connected with Alice in 45 days"
- Enable direct logging: "Log that I emailed Alice about her new role"

## Proactive Reminder System

### Daily Digest

**Command:**
```bash
pagen followups digest --format [text|json|html]
```

**Output Format (text):**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  FOLLOW-UPS FOR 2025-11-27
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔴 OVERDUE (3 contacts)
  Alice Chen        45 days  (priority: 90)
  Bob Martinez      38 days  (priority: 76)
  Carol Johnson     32 days  (priority: 64)

🟡 DUE SOON (2 contacts)
  David Park        27 days  (priority: 54)
  Emma Wilson       26 days  (priority: 52)

💡 SUGGESTIONS
  - Alice: Last met at conference, mentioned new project
  - Bob: Check in about job search
```

**Integration Options:**

1. **macOS Notification:**
```bash
# In crontab: 0 9 * * * /path/to/pagen followups digest --format text | terminal-notifier -title "Follow-Ups" -sound default
```

2. **Email Digest:**
```bash
# In crontab: 0 9 * * * /path/to/pagen followups digest --format html | mail -s "Daily Follow-Ups" you@example.com
```

3. **Slack/Discord:**
```bash
# In crontab: 0 9 * * * /path/to/pagen followups digest --format json | webhook-sender --url $SLACK_WEBHOOK
```

4. **Claude MCP:** Auto-surfaces during conversations about networking

## Analytics & Insights

### Network Health Dashboard

Add to `pagen viz` command:

```
NETWORK HEALTH
  🟢 Strong relationships: 12 (avg contact: 18 days)
  🟡 Medium relationships: 23 (avg contact: 35 days)
  🔴 Weak relationships: 8 (avg contact: 67 days)

  Interaction trends: ↗️ +15% this month
  At-risk contacts: 5 (no contact in 90+ days)

  This week: 8 interactions logged
  This month: 32 interactions logged
```

### GraphViz Enhancements

Enhance `pagen viz graph contacts` to visualize follow-up status:
- **Node colors:** Green (on track), Yellow (due soon), Red (overdue)
- **Node size:** Proportional to relationship strength
- **Edge thickness:** Based on interaction frequency
- **Labels:** Include days since last contact

### Interaction Pattern Analysis

Track and surface insights:
- Average cadence per contact (actual vs configured)
- Temporal patterns (e.g., "Most interactions on Tuesdays")
- Engagement trends (weekly/monthly/quarterly)
- Relationship drift warnings (contacts slipping from strong to weak)

### MCP Stats Integration

Quick stats callable from Claude:
- "You're managing 43 contacts with 5 needing follow-up"
- "Your strongest relationships: Alice, Bob, Carol"
- "Most neglected: David (87 days since contact)"
- "Network engagement: down 30% this quarter"

## Future Enhancements (Not in Scope)

These are explicitly out of scope for the initial implementation but could be added later:

1. **Email/Calendar Integration:** Auto-log interactions from Gmail/Calendar APIs
2. **Birthday/Event Tracking:** Special occasion reminders
3. **Template Messages:** Store follow-up message templates
4. **Task Manager Export:** Push to Todoist, Things, or other systems
5. **Mobile Notifications:** Push notifications to phone
6. **AI-Generated Follow-Up Suggestions:** Claude suggests specific conversation topics

## Implementation Notes

### Migration Strategy

- Add new tables with migrations
- Backfill `last_interaction_date` from existing interaction logs if present
- Default all contacts to 30-day cadence and "medium" strength
- Compute initial priority scores

### Performance Considerations

- Index on `next_followup_date` for fast querying
- Index on `priority_score` for sorted views
- Compute priority scores on-write (when logging interactions) rather than on-read

### Testing Requirements

- Unit tests for priority scoring algorithm
- Integration tests for interaction logging workflow
- E2E tests for TUI follow-up tab
- Web UI tests for HTMX partial updates
- MCP tool tests for Claude integration

## Success Metrics

- Users log interactions regularly (>1/week)
- Follow-up list actively decreases (users acting on reminders)
- Relationship health improves (fewer at-risk contacts)
- Users report staying in touch with important contacts more consistently
