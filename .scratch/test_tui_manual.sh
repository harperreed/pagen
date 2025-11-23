#!/bin/bash
set -e

echo "=== TUI Manual Test Instructions ==="
echo ""
echo "This script creates test data. Then launch the TUI manually."
echo ""

export DB=/tmp/test_tui_$$.db

# Create test data
./pagen --db-path $DB crm add-company --name "Acme Corp"
./pagen --db-path $DB crm add-company --name "TechStart Inc"
./pagen --db-path $DB crm add-contact --name "Alice" --email "alice@acme.com" --company "Acme Corp"
./pagen --db-path $DB crm add-contact --name "Bob" --email "bob@techstart.com" --company "TechStart Inc"
./pagen --db-path $DB crm add-deal --title "Enterprise Deal" --company "Acme Corp" --amount 500000 --stage "negotiation"
./pagen --db-path $DB crm add-deal --title "Startup Deal" --company "TechStart Inc" --amount 50000 --stage "prospecting"

echo ""
echo "Test data created in: $DB"
echo ""
echo "To launch TUI, run:"
echo "  ./pagen --db-path $DB"
echo ""
echo "Test checklist:"
echo "  [ ] Tab switches between Contacts/Companies/Deals"
echo "  [ ] Arrow keys navigate rows"
echo "  [ ] Enter shows detail view"
echo "  [ ] Esc returns to list view"
echo "  [ ] 'e' in detail view shows edit form"
echo "  [ ] 'g' in detail view shows graph DOT"
echo "  [ ] 'n' in list view shows new entity form"
echo "  [ ] 'q' quits application"
echo ""
echo "Cleanup: rm $DB"
