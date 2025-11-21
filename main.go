// ABOUTME: Entry point for CRM MCP server and CLI
// ABOUTME: Routes to MCP server or CLI commands based on arguments
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/harperreed/crm-mcp/cli"
	"github.com/harperreed/crm-mcp/db"
)

const version = "0.1.0"

func main() {
	// Global flags
	showVersion := flag.Bool("version", false, "Show version and exit")
	dbPath := flag.String("db-path", "", "Database path (default: ~/.local/share/crm/crm.db)")
	initOnly := flag.Bool("init", false, "Initialize database and exit")

	// Parse global flags but don't fail on unknown (for subcommands)
	flag.CommandLine.Parse(os.Args[1:])

	// Handle version flag
	if *showVersion {
		fmt.Printf("crm-mcp version %s\n", version)
		os.Exit(0)
	}

	// Determine database path
	var finalDBPath string
	if *dbPath != "" {
		finalDBPath = *dbPath
	} else {
		finalDBPath = filepath.Join(xdg.DataHome, "crm", "crm.db")
	}

	// Initialize database
	database, err := db.OpenDatabase(finalDBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	log.Printf("CRM database: %s", finalDBPath)

	// Handle init-only flag
	if *initOnly {
		log.Println("Database initialized successfully")
		os.Exit(0)
	}

	// Get remaining args after flags
	args := flag.Args()

	// If no command specified, show usage
	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	// Route to command
	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "mcp":
		// Start MCP server
		if err := cli.MCPCommand(database); err != nil {
			log.Fatalf("MCP server failed: %v", err)
		}

	// Contact commands
	case "add-contact":
		if err := cli.AddContactCommand(database, commandArgs); err != nil {
			log.Fatalf("Error: %v", err)
		}
	case "list-contacts":
		if err := cli.ListContactsCommand(database, commandArgs); err != nil {
			log.Fatalf("Error: %v", err)
		}

	// Company commands
	case "add-company":
		if err := cli.AddCompanyCommand(database, commandArgs); err != nil {
			log.Fatalf("Error: %v", err)
		}
	case "list-companies":
		if err := cli.ListCompaniesCommand(database, commandArgs); err != nil {
			log.Fatalf("Error: %v", err)
		}

	// Deal commands
	case "add-deal":
		if err := cli.AddDealCommand(database, commandArgs); err != nil {
			log.Fatalf("Error: %v", err)
		}
	case "list-deals":
		if err := cli.ListDealsCommand(database, commandArgs); err != nil {
			log.Fatalf("Error: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`crm-mcp v%s - CRM tool with MCP server and CLI

USAGE:
  crm-mcp [global flags] <command> [command flags]

GLOBAL FLAGS:
  --version              Show version and exit
  --db-path <path>       Database path (default: ~/.local/share/crm/crm.db)
  --init                 Initialize database and exit

MCP SERVER:
  mcp                    Start MCP server for Claude Desktop (stdio)

CONTACT COMMANDS:
  add-contact            Add a new contact
    --name <name>          Contact name (required)
    --email <email>        Email address
    --phone <phone>        Phone number
    --company <company>    Company name
    --notes <notes>        Notes about contact

  list-contacts          List contacts
    --query <text>         Search by name or email
    --company <company>    Filter by company name
    --limit <n>            Max results (default: 50)

COMPANY COMMANDS:
  add-company            Add a new company
    --name <name>          Company name (required)
    --domain <domain>      Company domain (e.g., acme.com)
    --industry <industry>  Industry
    --notes <notes>        Notes about company

  list-companies         List companies
    --query <text>         Search by name or domain
    --limit <n>            Max results (default: 50)

DEAL COMMANDS:
  add-deal               Add a new deal
    --title <title>        Deal title (required)
    --company <company>    Company name (required)
    --amount <cents>       Deal amount in cents
    --currency <code>      Currency code (default: USD)
    --stage <stage>        Stage (default: prospecting)
    --notes <notes>        Initial notes

  list-deals             List deals
    --stage <stage>        Filter by stage
    --company <company>    Filter by company name
    --limit <n>            Max results (default: 50)

EXAMPLES:
  # Start MCP server for Claude Desktop
  crm-mcp mcp

  # Add a contact
  crm-mcp add-contact --name "John Smith" --email "john@acme.com" --company "Acme Corp"

  # List all contacts at Acme Corp
  crm-mcp list-contacts --company "Acme Corp"

  # Add a deal
  crm-mcp add-deal --title "Enterprise License" --company "Acme Corp" --amount 5000000

  # List deals in negotiation stage
  crm-mcp list-deals --stage negotiation

`, version)
}
