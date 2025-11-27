# Pagen - Personal Agent Toolkit

A personal agent toolkit with CRM capabilities. Works both as a Model Context Protocol (MCP) server for Claude Desktop AND as a standalone CLI for direct terminal use.

## Features

- **Contact Management** - Track people with full interaction history
- **Company Management** - Organize contacts by company with industry tracking
- **Deal Pipeline** - Manage sales from prospecting to closed
- **Relationship Tracking** - Map connections between contacts (colleagues, friends, etc.)
- **Follow-Up Tracking** - Never lose touch with your network through smart cadence tracking
- **Universal Query** - Flexible searching across all entity types

## Installation

### Homebrew (macOS)

```bash
# Add the tap
brew tap harperreed/tap

# Install pagen
brew install pagen
```

### Pre-built Binaries

Download the latest release from [GitHub Releases](https://github.com/harperreed/pagen/releases)

### Build from Source

```bash
git clone https://github.com/harperreed/pagen.git
cd pagen
CGO_ENABLED=1 go build -o pagen
```

## Usage

Pagen can be used in two ways:

### 1. MCP Server for Claude Desktop

Configure Claude Desktop to use the MCP server.

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "pagen": {
      "command": "/path/to/pagen",
      "args": ["mcp"]
    }
  }
}
```

Then restart Claude Desktop and you'll have access to all 19 CRM tools through natural language.

### 2. Interactive TUI (Default)

Launch the full-screen interactive terminal interface:

```bash
pagen
```

Features:
- **Tab** - Switch between Contacts/Companies/Deals
- **Arrow keys** - Navigate rows
- **Enter** - View details
- **n** - Create new entity
- **e** - Edit selected entity
- **d** - Delete selected entity
- **g** - View graph for entity
- **/** - Search/filter
- **q** - Quit

### 3. CLI for Direct Terminal Use

Use the CLI directly for quick CRM operations:

```bash
# Add a company
pagen crm add-company --name "Acme Corp" --industry "Software" --domain "acme.com"

# Add a contact
pagen crm add-contact --name "John Smith" --email "john@acme.com" --company "Acme Corp"

# List contacts
pagen crm list-contacts

# List contacts at specific company
pagen crm list-contacts --company "Acme Corp"

# Add a deal
pagen crm add-deal --title "Enterprise License" --company "Acme Corp" --amount 5000000 --stage "negotiation"

# List deals
pagen crm list-deals

# List deals in specific stage
pagen crm list-deals --stage "negotiation"

# List companies
pagen crm list-companies
```

## Visualization Features

### Terminal Dashboard

View a static ASCII dashboard with pipeline overview, stats, and alerts:

```bash
pagen viz
```

Output:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  PAGEN CRM DASHBOARD
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PIPELINE OVERVIEW
  prospecting    ████░░░░░░  12 ($45K)
  qualification  ██████░░░░  18 ($120K)
  negotiation    ████████░░   8 ($250K)

STATS
  📇 45 contacts  🏢 12 companies  💼 43 deals

NEEDS ATTENTION
  ⚠️  8 contacts - no contact in 30+ days
  ⚠️  3 deals - stale (no activity in 14+ days)
```

### Read-Only Web UI

Start the web dashboard server:

```bash
pagen web [--port 8080]
```

Visit `http://localhost:8080` in your browser.

Pages:
- `/` - Dashboard with stats and pipeline
- `/contacts` - Searchable contacts table
- `/companies` - Companies with org charts
- `/deals` - Deals with stage filtering
- `/graphs` - Interactive graph generation

All pages use HTMX for partial updates (no full page reloads).

### GraphViz Visualizations

Generate relationship graphs in DOT format:

```bash
# All contact relationships
pagen viz graph contacts

# Specific contact's network
pagen viz graph contacts <contact-id>

# Company org chart
pagen viz graph company <company-id-or-name>

# Deal pipeline flow
pagen viz graph pipeline
```

Save to file:
```bash
pagen viz graph contacts > graph.dot
dot -Tsvg graph.dot -o graph.svg
```

## Updated Command Structure

```
pagen                          # Launch interactive TUI (default)
pagen crm <command> [args]     # CLI commands for scripting
pagen mcp                      # MCP server for Claude Desktop
pagen viz                      # Terminal dashboard
pagen viz graph <type> [args]  # Generate GraphViz graphs
pagen web [--port 8080]        # Web UI server
```

### Global Flags

- `--version` - Show version and exit
- `--db-path <path>` - Use custom database path (default: `~/.local/share/crm/crm.db`)
- `--init` - Initialize database and exit (use with `crm` command)

### Available Commands

**Top Level:**
- `pagen` - Launch interactive TUI (default)
- `pagen mcp` - Start MCP server (for Claude Desktop)
- `pagen crm` - CRM management commands
- `pagen viz` - Terminal dashboard
- `pagen viz graph` - Generate GraphViz visualizations
- `pagen web` - Start web UI server

Run `pagen --help` for full help.

## Complete CRUD Operations

### Contacts

```bash
pagen crm add-contact --name "Alice" --email "alice@example.com" [--phone "555-1234"] [--company "CompanyName"] [--notes "Notes"]
pagen crm find-contacts [--query "search"] [--company-id <uuid>]
pagen crm update-contact <id> [--name "New Name"] [--email "new@email.com"] [--phone "555-5678"] [--company "NewCompany"] [--notes "Updated notes"]
pagen crm delete-contact <id>
pagen crm log-interaction --contact <name-or-id> [--note "Met for coffee"]
```

### Companies

```bash
pagen crm add-company --name "Acme Corp" [--domain "acme.com"] [--industry "Software"] [--notes "Notes"]
pagen crm find-companies [--query "search"]
pagen crm update-company <id> [--name "New Name"] [--domain "newdomain.com"] [--industry "NewIndustry"] [--notes "Updated notes"]
pagen crm delete-company <id>  # Fails if company has active deals
```

### Deals

```bash
pagen crm add-deal --title "Enterprise License" --company "Acme Corp" [--contact "Alice"] [--amount 500000] [--currency USD] [--stage prospecting] [--note "Initial outreach"]
pagen crm find-deals [--query "search"]
pagen crm update-deal <id> [--title "New Title"] [--stage negotiation] [--amount 600000]
pagen crm delete-deal <id>  # Cascades to deal notes
pagen crm add-deal-note --deal <id> --note "Follow-up completed"
```

### Relationships

```bash
pagen crm link-contacts --contact1 <id-or-name> --contact2 <id-or-name> [--type "colleague"] [--context "Work together"]
pagen crm find-relationships --contact <id> [--type "colleague"]
pagen crm update-relationship <id> [--type "friend"] [--context "Updated context"]
pagen crm delete-relationship <id>
```

### Query (MCP-style)

```bash
pagen crm query --entity-type <contact|company|deal|relationship> [--query "search"] [--limit 50]
```

### Follow-Up Commands

```bash
# List contacts needing follow-up
pagen followups list [--overdue-only] [--strength weak|medium|strong] [--limit 10]

# Log an interaction
pagen followups log --contact "Alice" --type meeting --notes "Coffee chat"

# Set follow-up cadence
pagen followups set-cadence --contact "Bob" --days 14 --strength strong

# View network health stats
pagen followups stats

# Generate daily digest
pagen followups digest [--format text|json|html]
```

### Follow-Up in TUI

Press `f` to view the Follow-Ups tab showing:
- Prioritized list of contacts needing attention
- Visual indicators (🔴 overdue, 🟡 due soon, 🟢 on track)
- Quick interaction logging with `l`
- Cadence adjustment with `c`

### Follow-Up in Web UI

Visit `/followups` for:
- Filterable table of contacts needing follow-up
- One-click interaction logging via HTMX
- Priority-based sorting

## Database

The server uses SQLite and stores data at:
- **Default:** `~/.local/share/crm/crm.db` (XDG data directory)
- **Custom:** Specify with `--db-path` flag

### Database Schema

- **contacts** - People with names, emails, phone numbers, company associations
- **companies** - Organizations with industry and domain information
- **deals** - Sales pipeline with stages, amounts, and expected close dates
- **deal_notes** - Activity logs on deals
- **relationships** - Bidirectional connections between contacts

## MCP Tools

Total: **22 tools** for Claude Desktop integration

### Contact Operations (5 tools)
- `add_contact` - Create new contacts with optional company linking
- `find_contacts` - Search by name, email, or company
- `update_contact` - Modify contact information
- `delete_contact` - Delete a contact and all associated relationships
- `log_contact_interaction` - Record interactions with timestamp tracking

### Company Operations (4 tools)
- `add_company` - Create companies with industry/domain metadata
- `find_companies` - Search by name or domain
- `update_company` - Modify company information
- `delete_company` - Delete a company (must have no active deals)

### Deal Operations (4 tools)
- `create_deal` - Create deals with company and contact associations
- `update_deal` - Modify deal details including stage and amount
- `delete_deal` - Delete a deal and all associated notes
- `add_deal_note` - Add activity notes to deals

### Relationship Operations (4 tools)
- `link_contacts` - Create relationships between contacts
- `find_contact_relationships` - Find all connections for a contact
- `update_relationship` - Update a relationship's type and context
- `remove_relationship` - Delete relationship links

### Follow-Up Operations (3 tools)
- `get_followup_list` - Get prioritized follow-up suggestions
- `log_interaction` - Log interactions and update tracking
- `set_cadence` - Configure follow-up frequency per contact

### Query Operations (1 tool)
- `query_crm` - Universal query across all entity types with flexible filtering

### Visualization Operations (1 tool)
- `generate_graph` - Generate GraphViz DOT for contact networks, company org charts, or deal pipelines

## Example Usage

### In Claude Desktop (via MCP)

Once configured, try these natural language commands:

```
Add a company called Acme Corp in the software industry

Add a contact named John Smith with email john@acme.com at Acme Corp

Create a deal titled "Enterprise License" with Acme Corp for $50,000

Link John Smith and Jane Doe as colleagues who met at a conference

Show me all my contacts at Acme Corp

Query all deals in negotiation stage above $10,000
```

### In Terminal (via CLI)

Quick commands for direct CRM management:

```bash
# Start fresh
pagen --db-path ~/my-crm.db --init crm

# Add companies
pagen crm add-company --name "Acme Corp" --industry "Software"
pagen crm add-company --name "TechStart" --industry "SaaS"

# Add contacts
pagen crm add-contact --name "Alice" --email "alice@acme.com" --company "Acme Corp"
pagen crm add-contact --name "Bob" --email "bob@techstart.io" --company "TechStart"

# Create deals
pagen crm add-deal --title "Enterprise" --company "Acme Corp" --amount 5000000

# Query data
pagen crm list-contacts --company "Acme Corp"
pagen crm list-deals --stage negotiation
```

## Architecture

**Single Binary Distribution:**
- All templates embedded via `go:embed`
- No external files required
- Pure Go dependencies only

**Technology Stack:**
- Database: SQLite with CGO
- TUI: bubbletea + lipgloss
- Web: Go templates + HTMX (CDN) + Tailwind (CDN)
- GraphViz: goccy/go-graphviz (pure Go)
- MCP: Claude Agent SDK

## Testing

Run all scenario tests:

```bash
make test
./.scratch/test_all_visualization.sh
./.scratch/test_integration.sh
```

Manual testing:
- TUI: `./.scratch/test_tui_manual.sh`
- Web: `./.scratch/test_web_manual.sh`

## Development

### Running Tests

```bash
go test ./...
```

### Test Coverage

```bash
go test ./... -cover
```

## Version

Current version: 0.1.0

## License

[Add your license here]
