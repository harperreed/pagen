# Pagen CRM Visualization Design

**Date:** 2025-11-22
**Status:** Approved
**Author:** Claude + Harper

## Overview

Add comprehensive visualization capabilities to Pagen CRM: terminal dashboard, interactive TUI, read-only web UI, and GraphViz relationship graphs. Expand CRUD operations across all entities for both CLI and MCP interfaces.

## Goals

1. Visualize CRM data in terminal, TUI, and web browser
2. Generate relationship/org/pipeline graphs with GraphViz
3. Complete CRUD operations for all entities
4. Maintain single-binary distribution
5. Keep CLI primary interface, add TUI for interactive work

## Command Structure

```bash
pagen                    # Launch TUI (new default)
pagen crm <command>      # CLI commands (scripting/automation)
pagen web [--port 8080]  # Web dashboard (read-only)
pagen mcp                # MCP server (agents)
pagen viz                # Terminal dashboard (static)
pagen viz graph <type>   # Generate GraphViz graphs
```

## Architecture

### Technology Stack

- **TUI:** bubbletea + lipgloss (styling)
- **Web:** Go templates (embedded) + HTMX (CDN) + Tailwind (CDN)
- **GraphViz:** goccy/go-graphviz (pure Go, no external deps)
- **Embedding:** Go embed package for all templates/static assets

### Single Binary Distribution

All assets embedded in the compiled binary:
```go
//go:embed web/templates/*
var templates embed.FS

//go:embed web/static/*
var static embed.FS
```

No external dependencies. No separate files. One `pagen` binary.

## Features

### 1. Terminal Dashboard (`pagen viz`)

Static ASCII dashboard showing:

**Pipeline Overview:**
- Bar charts for each deal stage
- Deal count and total value per stage
- Visual progress indicators

**Recent Activity:**
- Last 7 days of changes
- Contacts added, deals updated, interactions logged
- Timestamped list

**Stats:**
- Total contacts, companies, deals
- Counts with emoji indicators

**Needs Attention:**
- Contacts with no interaction in 30+ days
- Stale deals with no activity in 14+ days

Output format:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  PAGEN CRM DASHBOARD
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PIPELINE OVERVIEW
  Prospecting    ████░░░░░░  12 ($45K)
  Qualification  ██████░░░░  18 ($120K)
  Negotiation    ████████░░   8 ($250K)
  Closed Won     ██████████   5 ($180K)

RECENT ACTIVITY (Last 7 days)
  • 2025-01-15: Contact added - John Smith (Acme Corp)
  • 2025-01-14: Deal updated - Enterprise License → Negotiation
  • 2025-01-14: Interaction logged - Call with Jane Doe

STATS
  📇 45 contacts  🏢 12 companies  💼 43 deals

NEEDS ATTENTION
  ⚠️  8 contacts - no contact in 30+ days
  ⚠️  3 deals - stale (no activity in 14+ days)
```

### 2. Interactive TUI (`pagen`)

Full-screen terminal interface with keyboard navigation.

**Views:**

1. **List View** (default)
   - Three tabs: Contacts | Companies | Deals
   - Searchable/filterable table
   - Arrow keys navigate rows
   - Enter to view details

2. **Detail View**
   - Full entity information
   - Related entities (company contacts, contact deals, etc.)
   - Actions available via key bindings

3. **Edit View**
   - Form-based editing
   - Tab navigates fields
   - Enter saves, Esc cancels
   - Validation before save

4. **Graph View**
   - Shows DOT source or rendered output
   - For selected entity's relationships

**Key Bindings:**
- `/` - Search/filter current list
- `n` - Create new entity
- `e` - Edit selected entity
- `d` - Delete selected entity (with confirmation)
- `g` - View graph for selected entity
- `Tab` - Switch between entity type tabs
- `Enter` - View details / confirm action
- `Esc` - Go back / cancel
- `q` - Quit application
- `?` - Show help overlay

**Delete Confirmations:**
- All deletes show confirmation modal
- Must type "yes" or press "y" to confirm
- Esc cancels

### 3. Web UI (`pagen web`)

Read-only HTML dashboard served at `http://localhost:8080`.

**Technology:**
- Server-side rendered Go templates
- HTMX for partial updates (no full page reloads)
- Tailwind CSS for styling
- Both loaded via CDN (no build step)

**Pages:**

1. **Dashboard (`/`)**
   - Same stats as terminal dashboard
   - Pipeline chart (HTML/CSS bars)
   - Recent activity feed
   - Stats cards
   - "Needs attention" section

2. **Contacts (`/contacts`)**
   - Searchable table
   - Click row loads detail via HTMX
   - Search updates table without reload

3. **Companies (`/companies`)**
   - Searchable table
   - Detail view shows all contacts and deals
   - Inline GraphViz org chart (SVG)

4. **Deals (`/deals`)**
   - Searchable table
   - Filter by stage dropdown
   - Detail shows timeline of notes

5. **Graphs (`/graphs`)**
   - Dropdowns select graph type and entity
   - Renders SVG inline via HTMX
   - Shows DOT source in collapsible section

**Features:**
- Auto-refresh via HTMX polling (configurable)
- Responsive tables
- Clean, minimal design
- No forms (read-only)

**File Structure:**
```
web/
  templates/
    layout.html           # Base with HTMX/Tailwind CDN
    dashboard.html
    contacts.html
    companies.html
    deals.html
    graphs.html
    partials/
      contact-detail.html
      company-detail.html
      deal-detail.html
      graph.html
```

### 4. GraphViz Visualizations

Three graph types using `goccy/go-graphviz` (pure Go):

**Contact Relationship Network:**
- Nodes: Contacts (labeled with names)
- Edges: Relationships (labeled with types)
- Colors: Group by company
- Layout: Use `neato` or `fdp` for network graph

**Company Org Chart:**
- Root node: Company name
- Child nodes: Contacts at company
- Edges: Relationships between contacts
- Layout: Use `dot` for hierarchical

**Deal Pipeline Flow:**
- Nodes: Deals (title + amount)
- Grouped by stage (left to right)
- Could color by company or deal size
- Layout: Use `dot` for left-to-right flow

**CLI Commands:**
```bash
pagen viz graph contacts [--output graph.svg]
pagen viz graph company <name> [--output graph.svg]
pagen viz graph pipeline [--output graph.svg]
pagen viz graph contact <id>  # Specific contact network
```

**MCP Tool:**
```
generate_graph(type, entity_id, format="svg")
→ returns {dot_source, node_count, edge_count}
```

**Integration:**
- **TUI:** Press `g` on entity → renders to temp file or shows DOT
- **Web:** Inline SVG rendered server-side
- **MCP:** Returns DOT source for agent to show/explain

**Rendering:**
- Use `goccy/go-graphviz` to generate SVG/PNG/DOT
- No external `dot` command needed
- Pure Go, embedded in binary

## CRUD Operations

### New CLI Commands

**Contacts:**
```bash
pagen crm update-contact <id> [--name] [--email] [--phone] [--company] [--notes]
pagen crm delete-contact <id>
```

**Companies:**
```bash
pagen crm update-company <id> [--name] [--domain] [--industry] [--notes]
pagen crm delete-company <id>
```

**Deals:**
```bash
pagen crm delete-deal <id>
# update-deal already exists
```

**Relationships:**
```bash
pagen crm update-relationship <id> [--type] [--context]
pagen crm delete-relationship <id>
# remove-relationship may already exist
```

**Deal Notes:**
```bash
pagen crm delete-deal-note <id>
# Notes are historical - no update
```

### New MCP Tools

```
delete_contact(contact_id)
delete_company(company_id)
delete_deal(deal_id)
update_relationship(relationship_id, relationship_type, context)
delete_relationship(relationship_id)
delete_deal_note(note_id)
```

### Cascade Behavior

**Deleting a company:**
- Warn if company has contacts or deals
- Require confirmation
- Set contact.company_id to NULL for affected contacts
- Prevent delete if deals exist (or cascade delete deals)

**Deleting a contact:**
- Warn if contact has relationships or deals
- Remove all relationships involving contact
- Set deal.contact_id to NULL for affected deals

**Deleting a deal:**
- Also delete all associated deal_notes
- Cascade delete notes

### Validation

All updates validate:
- Email format (basic regex)
- UUID existence (foreign keys)
- Required fields present
- Stage values match constants

## Database Schema (No Changes)

Existing schema supports all features:
- contacts, companies, deals, deal_notes, relationships tables
- All tables have id, created_at, updated_at
- Foreign keys for relationships

No migrations needed for visualization features.

## Implementation Phases

### Phase 1: CRUD Completion
- Add missing update/delete functions to db package
- Add CLI commands for update/delete operations
- Add MCP tools for update/delete operations
- Test all CRUD operations via scenarios

### Phase 2: GraphViz Integration
- Add `goccy/go-graphviz` dependency
- Implement graph generation functions
- Add `pagen viz graph` CLI commands
- Add `generate_graph` MCP tool
- Test graph generation for all types

### Phase 3: Terminal Dashboard
- Implement `pagen viz` static dashboard
- Pipeline bar charts
- Recent activity query
- Stats aggregation
- "Needs attention" alerts

### Phase 4: TUI
- Add `bubbletea` dependency
- Implement list view with tabs
- Implement detail view
- Implement edit view with forms
- Implement delete confirmations
- Integrate graph viewing
- Wire up all key bindings

### Phase 5: Web UI
- Create template structure
- Implement layout with HTMX/Tailwind CDN
- Build dashboard page
- Build list pages (contacts, companies, deals)
- Build detail partials with HTMX
- Build graphs page with inline SVG
- Embed all templates in binary
- Add HTTP server with routes

### Phase 6: Integration & Polish
- Test all features end-to-end
- Write scenario tests for each feature
- Update README with new commands
- Update MCP tool count (will be 19 tools total)
- Tag release

## Success Criteria

- [ ] `pagen` launches interactive TUI
- [ ] `pagen viz` shows terminal dashboard
- [ ] `pagen viz graph` generates all graph types
- [ ] `pagen web` serves read-only dashboard at localhost:8080
- [ ] All CRUD operations work in CLI, TUI, and MCP
- [ ] GraphViz graphs render in TUI, web, and MCP
- [ ] Single binary contains all assets
- [ ] No external dependencies beyond Go runtime
- [ ] All features validated with scenario tests

## Non-Goals

- Authentication/authorization (local tool only)
- Multi-user support
- Real-time collaboration
- Mobile-responsive web UI (desktop only)
- Graph editing (read-only visualizations)
- Export to formats other than SVG/PNG/DOT

## Dependencies

New dependencies to add:
```
github.com/goccy/go-graphviz    # GraphViz rendering
github.com/charmbracelet/bubbletea  # TUI framework
github.com/charmbracelet/lipgloss   # TUI styling
```

CDN dependencies (not in go.mod):
- HTMX (https://unpkg.com/htmx.org)
- Tailwind CSS (https://cdn.tailwindcss.com)

## Open Questions

None - design is approved and ready for implementation.

## References

- [goccy/go-graphviz](https://github.com/goccy/go-graphviz) - Pure Go GraphViz
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [HTMX](https://htmx.org/) - HTML-over-the-wire
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS
