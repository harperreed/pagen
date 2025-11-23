#!/bin/bash
set -e

echo "=== Testing Contact Update/Delete CLI ==="

export DB=/tmp/test_crud_$$.db

# Setup - capture the ID from the add-contact output
OUTPUT=$(./pagen --db-path $DB crm add-contact --name "John Doe" --email "john@example.com" 2>&1)
echo "$OUTPUT"
CONTACT_ID=$(echo "$OUTPUT" | grep -o '[0-9a-f]\{8\}-[0-9a-f]\{4\}-[0-9a-f]\{4\}-[0-9a-f]\{4\}-[0-9a-f]\{12\}' | head -1)

echo "Contact ID: $CONTACT_ID"

# Test update
echo ""
echo "Testing update-contact..."
./pagen --db-path $DB crm update-contact --name "Jane Doe" --email "jane@example.com" $CONTACT_ID
./pagen --db-path $DB crm list-contacts --query "Jane" | grep "jane@example.com" || { echo "ERROR: Update failed - jane@example.com not found"; exit 1; }
echo "✓ Update successful"

# Test delete
echo ""
echo "Testing delete-contact..."
./pagen --db-path $DB crm delete-contact $CONTACT_ID
! ./pagen --db-path $DB crm list-contacts --query "Jane" | grep "jane@example.com" || { echo "ERROR: Delete failed - contact still exists"; exit 1; }
echo "✓ Delete successful"

# Cleanup
rm $DB

echo ""
echo "✓ Contact update/delete CLI works"
