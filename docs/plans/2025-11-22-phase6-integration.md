# Phase 6: Integration & Testing Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Integrate all visualization features, create comprehensive scenario tests, update documentation, and verify everything works together

**Architecture:** Final integration testing, documentation updates, and release preparation

**Tech Stack:** Bash scenario tests, Go testing, documentation updates

---

## Task 6.1: Create comprehensive scenario tests

**Files:**
- Create: `.scratch/test_all_visualization.sh`
- Create: `.scratch/test_integration.sh`

**Step 1: Create comprehensive visualization test**

Create `.scratch/test_all_visualization.sh`:

```bash
#!/bin/bash
set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PAGEN CRM - Comprehensive Visualization Test"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

export DB=/tmp/test_viz_all_$$.db
ERRORS=0

# Helper function to check output
check_output() {
    local description="$1"
    local command="$2"
    local expected="$3"

    echo -n "Testing: $description... "
    OUTPUT=$(eval "$command" 2>&1)
    if echo "$OUTPUT" | grep -q "$expected"; then
        echo "✓"
    else
        echo "✗"
        echo "  Expected to find: $expected"
        echo "  Got: $OUTPUT"
        ERRORS=$((ERRORS + 1))
    fi
}

echo "=== Setup Test Data ==="
./pagen --db-path $DB crm add-company --name "Acme Corp" --domain "acme.com" --industry "Software"
./pagen --db-path $DB crm add-company --name "TechStart Inc" --domain "techstart.io" --industry "SaaS"
./pagen --db-path $DB crm add-company --name "BigCo Ltd" --domain "bigco.com" --industry "Enterprise"
echo "  ✓ Created 3 companies"

./pagen --db-path $DB crm add-contact --name "Alice Johnson" --email "alice@acme.com" --company "Acme Corp"
./pagen --db-path $DB crm add-contact --name "Bob Smith" --email "bob@techstart.io" --company "TechStart Inc"
./pagen --db-path $DB crm add-contact --name "Carol White" --email "carol@acme.com" --company "Acme Corp"
./pagen --db-path $DB crm add-contact --name "Dave Brown" --email "dave@bigco.com" --company "BigCo Ltd"
./pagen --db-path $DB crm add-contact --name "Eve Davis" --email "eve@techstart.io" --company "TechStart Inc"
echo "  ✓ Created 5 contacts"

./pagen --db-path $DB crm add-deal --title "Enterprise License" --company "Acme Corp" --contact "Alice Johnson" --amount 500000 --stage "negotiation"
./pagen --db-path $DB crm add-deal --title "Startup Package" --company "TechStart Inc" --contact "Bob Smith" --amount 50000 --stage "prospecting"
./pagen --db-path $DB crm add-deal --title "Premium Features" --company "Acme Corp" --contact "Carol White" --amount 250000 --stage "proposal"
./pagen --db-path $DB crm add-deal --title "Corporate Deal" --company "BigCo Ltd" --contact "Dave Brown" --amount 1000000 --stage "qualification"
./pagen --db-path $DB crm add-deal --title "Small Deal" --company "TechStart Inc" --amount 25000 --stage "prospecting"
echo "  ✓ Created 5 deals"

./pagen --db-path $DB crm link-contacts --contact1 "Alice Johnson" --contact2 "Carol White" --type "colleague" --context "Work together at Acme"
./pagen --db-path $DB crm link-contacts --contact1 "Bob Smith" --contact2 "Eve Davis" --type "colleague" --context "TechStart team"
echo "  ✓ Created 2 relationships"

echo ""
echo "=== Testing Terminal Dashboard (pagen viz) ==="
check_output \
    "Dashboard header" \
    "./pagen --db-path $DB viz" \
    "PAGEN CRM DASHBOARD"

check_output \
    "Stats - contacts count" \
    "./pagen --db-path $DB viz" \
    "5 contacts"

check_output \
    "Stats - companies count" \
    "./pagen --db-path $DB viz" \
    "3 companies"

check_output \
    "Stats - deals count" \
    "./pagen --db-path $DB viz" \
    "5 deals"

check_output \
    "Pipeline overview section" \
    "./pagen --db-path $DB viz" \
    "PIPELINE OVERVIEW"

check_output \
    "Stats section" \
    "./pagen --db-path $DB viz" \
    "STATS"

echo ""
echo "=== Testing GraphViz Graphs (pagen viz graph) ==="
check_output \
    "Contact graph - DOT header" \
    "./pagen --db-path $DB viz graph contacts" \
    "digraph"

check_output \
    "Contact graph - has nodes" \
    "./pagen --db-path $DB viz graph contacts" \
    "label="

check_output \
    "Company graph - requires ID" \
    "./pagen --db-path $DB viz graph company 2>&1 || true" \
    "company ID or name required"

check_output \
    "Pipeline graph - DOT header" \
    "./pagen --db-path $DB viz graph pipeline" \
    "digraph"

check_output \
    "Pipeline graph - has deals" \
    "./pagen --db-path $DB viz graph pipeline" \
    "Enterprise License"

echo ""
echo "=== Testing CRUD Operations ==="

# Update contact
ALICE_ID=$(./pagen --db-path $DB crm find-contacts --query "Alice" 2>/dev/null | grep -o '[0-9a-f-]\{36\}' | head -1)
./pagen --db-path $DB crm update-contact --name "Alice J. Johnson" --phone "555-1234" "$ALICE_ID" >/dev/null 2>&1

check_output \
    "Update contact - name changed" \
    "./pagen --db-path $DB crm find-contacts --query 'Alice J'" \
    "Alice J. Johnson"

check_output \
    "Update contact - phone added" \
    "./pagen --db-path $DB crm find-contacts --query 'Alice J'" \
    "555-1234"

# Update company
ACME_ID=$(./pagen --db-path $DB crm find-companies --query "Acme" 2>/dev/null | grep -o '[0-9a-f-]\{36\}' | head -1)
./pagen --db-path $DB crm update-company --domain "acmecorp.com" "$ACME_ID" >/dev/null 2>&1

check_output \
    "Update company - domain changed" \
    "./pagen --db-path $DB crm find-companies --query 'Acme'" \
    "acmecorp.com"

# Update deal
DEAL_ID=$(./pagen --db-path $DB crm find-deals --query "Enterprise" 2>/dev/null | grep -o '[0-9a-f-]\{36\}' | head -1)
./pagen --db-path $DB crm update-deal --stage "closed_won" "$DEAL_ID" >/dev/null 2>&1

check_output \
    "Update deal - stage changed" \
    "./pagen --db-path $DB crm find-deals --query 'Enterprise'" \
    "closed_won"

# Delete relationship
REL_ID=$(./pagen --db-path $DB crm query --entity-type relationship 2>/dev/null | grep -o '[0-9a-f-]\{36\}' | head -1)
./pagen --db-path $DB crm delete-relationship "$REL_ID" >/dev/null 2>&1

check_output \
    "Delete relationship - count decreased" \
    "./pagen --db-path $DB crm query --entity-type relationship 2>/dev/null | grep -c 'colleague' || echo 0" \
    "1"

echo ""
echo "=== Testing Delete Protection ==="

# Try to delete company with deals
ACME_DELETE_OUTPUT=$(./pagen --db-path $DB crm delete-company "$ACME_ID" 2>&1 || true)
if echo "$ACME_DELETE_OUTPUT" | grep -q "cannot delete company with"; then
    echo "Testing: Delete protection for companies... ✓"
else
    echo "Testing: Delete protection for companies... ✗"
    echo "  Expected error about deleting company with deals"
    ERRORS=$((ERRORS + 1))
fi

# Delete a deal first, then company should work
./pagen --db-path $DB crm delete-deal "$DEAL_ID" >/dev/null 2>&1

echo ""
echo "=== Testing MCP Tools (via query) ==="

check_output \
    "Query contacts" \
    "./pagen --db-path $DB crm query --entity-type contact" \
    "Alice"

check_output \
    "Query companies" \
    "./pagen --db-path $DB crm query --entity-type company" \
    "Acme"

check_output \
    "Query deals" \
    "./pagen --db-path $DB crm query --entity-type deal" \
    "Startup Package"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ]; then
    echo "  ✓ ALL TESTS PASSED"
else
    echo "  ✗ $ERRORS TEST(S) FAILED"
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Cleanup
rm $DB

exit $ERRORS
```

**Step 2: Create integration test**

Create `.scratch/test_integration.sh`:

```bash
#!/bin/bash
set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PAGEN CRM - Integration Test"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

export DB=/tmp/test_integration_$$.db

echo "=== Test Workflow: From CLI to Visualization ==="
echo ""

# Step 1: Create data via CLI
echo "Step 1: Creating CRM data via CLI..."
./pagen --db-path $DB crm add-company --name "TestCo" --domain "test.co"
./pagen --db-path $DB crm add-contact --name "Test User" --email "test@test.co" --company "TestCo"
./pagen --db-path $DB crm add-deal --title "Test Deal" --company "TestCo" --contact "Test User" --amount 100000 --stage "prospecting"
echo "  ✓ Data created"
echo ""

# Step 2: Verify via find commands
echo "Step 2: Verifying data via find commands..."
COMPANY_COUNT=$(./pagen --db-path $DB crm find-companies --query "Test" 2>/dev/null | grep -c "TestCo" || echo 0)
CONTACT_COUNT=$(./pagen --db-path $DB crm find-contacts --query "Test" 2>/dev/null | grep -c "Test User" || echo 0)
DEAL_COUNT=$(./pagen --db-path $DB crm find-deals --query "Test" 2>/dev/null | grep -c "Test Deal" || echo 0)

if [ "$COMPANY_COUNT" -eq 1 ] && [ "$CONTACT_COUNT" -eq 1 ] && [ "$DEAL_COUNT" -eq 1 ]; then
    echo "  ✓ All entities found"
else
    echo "  ✗ Missing entities (Company: $COMPANY_COUNT, Contact: $CONTACT_COUNT, Deal: $DEAL_COUNT)"
    exit 1
fi
echo ""

# Step 3: Update via CLI
echo "Step 3: Updating data via CLI..."
CONTACT_ID=$(./pagen --db-path $DB crm find-contacts --query "Test User" 2>/dev/null | grep -o '[0-9a-f-]\{36\}' | head -1)
./pagen --db-path $DB crm update-contact --phone "555-0000" "$CONTACT_ID"

DEAL_ID=$(./pagen --db-path $DB crm find-deals --query "Test Deal" 2>/dev/null | grep -o '[0-9a-f-]\{36\}' | head -1)
./pagen --db-path $DB crm update-deal --stage "negotiation" "$DEAL_ID"
echo "  ✓ Updates completed"
echo ""

# Step 4: Visualize via dashboard
echo "Step 4: Checking terminal dashboard..."
DASHBOARD_OUTPUT=$(./pagen --db-path $DB viz)
if echo "$DASHBOARD_OUTPUT" | grep -q "1 contacts" && echo "$DASHBOARD_OUTPUT" | grep -q "1 deals"; then
    echo "  ✓ Dashboard shows correct stats"
else
    echo "  ✗ Dashboard stats incorrect"
    exit 1
fi
echo ""

# Step 5: Generate graph
echo "Step 5: Generating relationship graph..."
GRAPH_OUTPUT=$(./pagen --db-path $DB viz graph contacts)
if echo "$GRAPH_OUTPUT" | grep -q "digraph" && echo "$GRAPH_OUTPUT" | grep -q "Test User"; then
    echo "  ✓ Graph generated with contact data"
else
    echo "  ✗ Graph generation failed"
    exit 1
fi
echo ""

# Step 6: Query via MCP-style
echo "Step 6: Querying via MCP-style commands..."
QUERY_OUTPUT=$(./pagen --db-path $DB crm query --entity-type contact)
if echo "$QUERY_OUTPUT" | grep -q "Test User" && echo "$QUERY_OUTPUT" | grep -q "555-0000"; then
    echo "  ✓ MCP query returns updated data"
else
    echo "  ✗ MCP query failed"
    exit 1
fi
echo ""

# Step 7: Delete workflow
echo "Step 7: Testing delete cascade..."
./pagen --db-path $DB crm delete-deal "$DEAL_ID"
DEAL_COUNT_AFTER=$(./pagen --db-path $DB crm find-deals --query "Test" 2>/dev/null | grep -c "Test Deal" || echo 0)

if [ "$DEAL_COUNT_AFTER" -eq 0 ]; then
    echo "  ✓ Deal deleted successfully"
else
    echo "  ✗ Deal still exists after delete"
    exit 1
fi

./pagen --db-path $DB crm delete-contact "$CONTACT_ID"
CONTACT_COUNT_AFTER=$(./pagen --db-path $DB crm find-contacts --query "Test" 2>/dev/null | grep -c "Test User" || echo 0)

if [ "$CONTACT_COUNT_AFTER" -eq 0 ]; then
    echo "  ✓ Contact deleted successfully"
else
    echo "  ✗ Contact still exists after delete"
    exit 1
fi
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✓ INTEGRATION TEST PASSED"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Cleanup
rm $DB
```

**Step 3: Make scripts executable and test**

Run:
```bash
chmod +x .scratch/test_all_visualization.sh
chmod +x .scratch/test_integration.sh
```

Run: `./.scratch/test_all_visualization.sh`
Expected: All tests pass

Run: `./.scratch/test_integration.sh`
Expected: Integration test passes

**Step 4: Commit**

```bash
git add .scratch/test_all_visualization.sh .scratch/test_integration.sh
git commit -m "test: add comprehensive visualization and integration tests"
```

---

## Task 6.2: Update README documentation

**Files:**
- Modify: `README.md`

**Step 1: Read current README**

Read `README.md` to understand current structure.

**Step 2: Update README with visualization features**

Update `README.md` to add new sections:

```markdown
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

### Interactive TUI

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

## MCP Tools

Total: **19 tools** for Claude Desktop integration

New visualization tool:
- `generate_graph` - Generate GraphViz DOT for contact networks, company org charts, or deal pipelines

All CRUD operations available via MCP:
- Contact: `add_contact`, `find_contacts`, `update_contact`, `delete_contact`, `log_contact_interaction`
- Company: `add_company`, `find_companies`, `update_company`, `delete_company`
- Deal: `create_deal`, `find_deals`, `update_deal`, `delete_deal`, `add_deal_note`
- Relationship: `link_contacts`, `find_contact_relationships`, `update_relationship`, `remove_relationship`
- Query: `query_crm`

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
```

**Step 3: Verify README renders correctly**

Preview the markdown to ensure formatting is correct.

**Step 4: Commit**

```bash
git add README.md
git commit -m "docs: update README with visualization features"
```

---

## Task 6.3: Update MCP tool count and verify configuration

**Files:**
- Modify: `README.md` (verify MCP tool count)
- Create: `.scratch/verify_mcp_tools.sh`

**Step 1: Create MCP tool verification script**

Create `.scratch/verify_mcp_tools.sh`:

```bash
#!/bin/bash
set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  PAGEN CRM - MCP Tool Verification"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Count tool registrations in cli/mcp.go
echo "Counting MCP tool registrations..."
TOOL_COUNT=$(grep -c 'mcp.AddTool(server, &mcp.Tool{' cli/mcp.go || echo 0)

echo "Found $TOOL_COUNT tool registrations in cli/mcp.go"
echo ""

# List all tools
echo "Registered MCP tools:"
grep -A 1 'Name:' cli/mcp.go | grep 'Name:' | sed 's/.*Name: "\([^"]*\)".*/  - \1/'
echo ""

# Expected tools
EXPECTED_TOOLS=(
    "add_contact"
    "find_contacts"
    "update_contact"
    "delete_contact"
    "log_contact_interaction"
    "add_company"
    "find_companies"
    "update_company"
    "delete_company"
    "create_deal"
    "find_deals"
    "update_deal"
    "delete_deal"
    "add_deal_note"
    "link_contacts"
    "find_contact_relationships"
    "update_relationship"
    "remove_relationship"
    "query_crm"
    "generate_graph"
)

EXPECTED_COUNT=${#EXPECTED_TOOLS[@]}

echo "Expected tool count: $EXPECTED_COUNT"
echo "Actual tool count: $TOOL_COUNT"
echo ""

if [ "$TOOL_COUNT" -eq "$EXPECTED_COUNT" ]; then
    echo "✓ Tool count matches expected"
else
    echo "✗ Tool count mismatch!"
    echo ""
    echo "Expected tools:"
    for tool in "${EXPECTED_TOOLS[@]}"; do
        echo "  - $tool"
    done
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✓ MCP TOOL VERIFICATION PASSED"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
```

**Step 2: Make script executable and run**

Run:
```bash
chmod +x .scratch/verify_mcp_tools.sh
./.scratch/verify_mcp_tools.sh
```

Expected: Verification passes with 20 tools total

**Step 3: Update README if needed**

If tool count is different than documented, update README.md with correct count.

**Step 4: Commit**

```bash
git add .scratch/verify_mcp_tools.sh README.md
git commit -m "test: add MCP tool verification script"
```

---

## Task 6.4: Create release checklist and tag

**Files:**
- Create: `docs/RELEASE-CHECKLIST.md`

**Step 1: Create release checklist**

Create `docs/RELEASE-CHECKLIST.md`:

```markdown
# Release Checklist

Use this checklist before creating a new release.

## Pre-Release Verification

### Build & Compilation
- [ ] `make build` completes without errors
- [ ] `make test` passes all tests
- [ ] No compiler warnings

### Scenario Tests
- [ ] `.scratch/test_all_visualization.sh` passes
- [ ] `.scratch/test_integration.sh` passes
- [ ] All existing scenario tests pass:
  - [ ] `test_update_delete_contact.sh`
  - [ ] `test_company_crud.sh`
  - [ ] `test_graphs.sh`

### Manual Testing
- [ ] TUI launches and navigates correctly (`pagen`)
  - [ ] Tab switching works
  - [ ] Arrow key navigation works
  - [ ] Enter shows details
  - [ ] Edit view works
  - [ ] Graph view works
  - [ ] Delete confirmations work
- [ ] Terminal dashboard displays (`pagen viz`)
  - [ ] Stats are correct
  - [ ] Pipeline bars render
  - [ ] Attention items show when applicable
- [ ] Web UI serves correctly (`pagen web`)
  - [ ] Dashboard loads at http://localhost:8080
  - [ ] All navigation links work
  - [ ] HTMX partials load without page refresh
  - [ ] Search filters work
  - [ ] Graph generation works
- [ ] GraphViz graphs generate (`pagen viz graph`)
  - [ ] Contact graphs work
  - [ ] Company graphs work
  - [ ] Pipeline graphs work
- [ ] MCP server starts (`pagen mcp`)
  - [ ] Server responds to requests
  - [ ] All 20 tools are available

### Documentation
- [ ] README.md is up to date
  - [ ] Command examples are accurate
  - [ ] MCP tool count is correct (20)
  - [ ] New features are documented
- [ ] All plan documents are complete
- [ ] CLAUDE.md is accurate (if applicable)

### Code Quality
- [ ] No TODO comments in production code
- [ ] All functions have ABOUTME comments
- [ ] Error handling is consistent
- [ ] No debug logging left in code

### Dependencies
- [ ] `go.mod` and `go.sum` are clean
- [ ] All dependencies are necessary
- [ ] No version conflicts

## Release Process

### Version Bump
- [ ] Update version in relevant files
- [ ] Update CHANGELOG.md (if exists)

### Git
- [ ] All changes committed
- [ ] Working directory is clean
- [ ] On main branch (or release branch)

### Tag & Push
- [ ] Create git tag: `git tag -a v0.X.0 -m "Release v0.X.0"`
- [ ] Push tag: `git push origin v0.X.0`
- [ ] Push commits: `git push`

### Build Release Binaries
- [ ] Build for Linux: `GOOS=linux GOARCH=amd64 go build -o pagen-linux-amd64`
- [ ] Build for macOS: `GOOS=darwin GOARCH=amd64 go build -o pagen-darwin-amd64`
- [ ] Build for macOS ARM: `GOOS=darwin GOARCH=arm64 go build -o pagen-darwin-arm64`
- [ ] Test each binary on target platform

### GitHub Release (if applicable)
- [ ] Create GitHub release from tag
- [ ] Upload binaries
- [ ] Write release notes highlighting new features

## Post-Release
- [ ] Verify release is downloadable
- [ ] Test installation from release
- [ ] Update project documentation links
- [ ] Announce release (if applicable)

## Rollback Plan
If issues are discovered:
1. Document the issue
2. Create hotfix branch
3. Fix and test
4. Create patch release (v0.X.1)
5. Mark broken release as pre-release/draft
```

**Step 2: Commit checklist**

```bash
git add docs/RELEASE-CHECKLIST.md
git commit -m "docs: add release checklist"
```

**Step 3: Run through checklist manually**

Verify all items in the checklist for this release.

**Step 4: Create git tag (if ready)**

If all checks pass:
```bash
git tag -a v0.3.0 -m "Release v0.3.0 - Visualization Features

Features:
- Interactive TUI with bubbletea
- Terminal ASCII dashboard
- Read-only web UI with HTMX
- GraphViz relationship graphs
- Complete CRUD operations across all entities
- 20 MCP tools for Claude Desktop integration

Phases completed:
- Phase 1: CRUD completion (all entities)
- Phase 2: GraphViz integration
- Phase 3: Terminal dashboard
- Phase 4: Interactive TUI
- Phase 5: Web UI
- Phase 6: Integration testing and documentation"
```

**Step 5: Commit final changes**

```bash
git add .
git commit -m "chore: prepare for v0.3.0 release"
```

---

## Success Criteria

- [ ] All scenario tests pass without errors
- [ ] Integration test demonstrates full workflow
- [ ] README accurately documents all features
- [ ] MCP tool count is verified (20 tools)
- [ ] Release checklist is complete
- [ ] Git tag created for release
- [ ] All visualization features work together seamlessly
- [ ] No broken functionality from previous releases
- [ ] Documentation is comprehensive and accurate
- [ ] Single binary contains all assets (templates, etc.)
- [ ] No external dependencies beyond Go runtime

## Final Verification Commands

Run these commands in sequence to verify everything works:

```bash
# Build
make build

# Run all tests
make test
./.scratch/test_all_visualization.sh
./.scratch/test_integration.sh

# Verify MCP tools
./.scratch/verify_mcp_tools.sh

# Manual checks
# 1. Launch TUI and navigate all views
./pagen

# 2. Generate dashboard
./pagen viz

# 3. Generate graphs
./pagen viz graph contacts
./pagen viz graph pipeline

# 4. Start web server (check in browser)
./pagen web

# 5. Start MCP server (verify it starts)
./pagen mcp &
MCP_PID=$!
sleep 2
kill $MCP_PID
```

If all commands succeed, the release is ready!
