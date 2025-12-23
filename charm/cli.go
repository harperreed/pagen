// ABOUTME: CLI commands for Charm KV sync operations
// ABOUTME: Simplified sync with SSH key auth - no login/logout needed

package charm

import (
	"flag"
	"fmt"

	"github.com/charmbracelet/charm/kv"
)

// SyncLinkCommand links this device to a Charm account
// Uses SSH key auth - charm handles this automatically via SSH keys.
func SyncLinkCommand(args []string) error {
	fs := flag.NewFlagSet("sync link", flag.ExitOnError)
	_ = fs.Parse(args)

	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Linking to Charm Cloud (%s)...\n\n", cfg.Host)
	fmt.Println("Charm uses SSH key authentication.")

	// Get client to test connection
	c, err := GetClient()
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	// Test connection by syncing
	if err := c.Sync(); err != nil {
		return fmt.Errorf("link failed: %w", err)
	}

	// Get and display ID using the client's cached method
	id, err := c.ID()
	if err != nil {
		fmt.Println("✓ Device linked (ID unavailable)")
	} else {
		fmt.Printf("✓ Linked to account: %s\n", id)
	}

	fmt.Printf("✓ Auto-sync: %v\n", cfg.AutoSync)
	fmt.Println("\nYour device is now syncing with Charm Cloud!")

	return nil
}

// SyncStatusCommand shows current sync configuration and status.
func SyncStatusCommand(args []string) error {
	fs := flag.NewFlagSet("sync status", flag.ExitOnError)
	_ = fs.Parse(args)

	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return showSyncStatus(cfg)
}

func showSyncStatus(cfg *Config) error {
	fmt.Println("Charm Sync Status")
	fmt.Println("─────────────────")
	fmt.Printf("Server:    %s\n", cfg.Host)
	fmt.Printf("Auto-sync: %v\n", cfg.AutoSync)

	// Get client to check connection status
	c, err := GetClient()
	if err != nil {
		fmt.Println("\nStatus: Not connected")
		fmt.Println("\nCharm uses SSH keys for authentication - no login required!")
		return nil //nolint:nilerr // Not connected is a valid state
	}

	// Use client's cached ID method
	id, err := c.ID()
	if err != nil {
		fmt.Println("\nStatus: Connected (ID unavailable)")
	} else {
		fmt.Println("\nStatus: Connected to Charm Cloud")
		fmt.Printf("ID:        %s\n", id)
	}

	// Show KV stats
	keys, err := c.Keys()
	if err == nil {
		fmt.Printf("Keys:      %d\n", len(keys))
	}

	fmt.Println("\nCharm uses SSH keys for authentication - no login required!")
	fmt.Println("Sync happens automatically in the background.")

	return nil
}

// SyncUnlinkCommand disconnects this device from the Charm account
// Note: Charm doesn't provide a direct "unlink" API - users should remove
// SSH keys from their Charm account to fully unlink.
func SyncUnlinkCommand(args []string) error {
	fs := flag.NewFlagSet("sync unlink", flag.ExitOnError)
	_ = fs.Parse(args)

	fmt.Println("To unlink your device from Charm Cloud:")
	fmt.Println()
	fmt.Println("  1. Remove this device's SSH key from your Charm account")
	fmt.Println("  2. Delete local charm data: rm -rf ~/.local/share/charm")
	fmt.Println()
	fmt.Println("Local pagen data will be preserved in ~/.local/share/pagen")

	return nil
}

// SyncWipeCommand completely resets the KV store
// WARNING: This deletes all local data!
func SyncWipeCommand(args []string) error {
	fs := flag.NewFlagSet("sync wipe", flag.ExitOnError)
	confirm := fs.Bool("confirm", false, "Confirm data wipe")
	_ = fs.Parse(args)

	if !*confirm {
		fmt.Println("WARNING: This will delete ALL local data!")
		fmt.Println()
		fmt.Println("To confirm, run:")
		fmt.Println("  pagen sync wipe --confirm")
		return nil
	}

	c, err := GetClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	// Reset the KV store
	if err := c.Reset(); err != nil {
		return fmt.Errorf("failed to reset KV store: %w", err)
	}

	fmt.Println("✓ All data wiped")
	fmt.Println("Your Charm account is still linked.")
	fmt.Println("You can start adding data again.")

	return nil
}

// SyncNowCommand performs an immediate sync.
func SyncNowCommand(args []string) error {
	fs := flag.NewFlagSet("sync now", flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Show verbose output")
	_ = fs.Parse(args)

	c, err := GetClient()
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	if *verbose {
		fmt.Println("Syncing with server...")
	}

	if err := c.Sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if *verbose {
		fmt.Println("✓ Sync complete")
	} else {
		fmt.Println("✓ Synced")
	}

	return nil
}

// SetAutoSyncCommand enables or disables auto-sync.
func SetAutoSyncCommand(args []string) error {
	fs := flag.NewFlagSet("sync auto", flag.ExitOnError)
	enable := fs.Bool("enable", false, "Enable auto-sync")
	disable := fs.Bool("disable", false, "Disable auto-sync")
	_ = fs.Parse(args)

	if !*enable && !*disable {
		fmt.Println("Usage: pagen sync auto --enable|--disable")
		return nil
	}

	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if *enable {
		if err := cfg.SetAutoSync(true); err != nil {
			return fmt.Errorf("failed to enable auto-sync: %w", err)
		}
		fmt.Println("✓ Auto-sync enabled")
	} else if *disable {
		if err := cfg.SetAutoSync(false); err != nil {
			return fmt.Errorf("failed to disable auto-sync: %w", err)
		}
		fmt.Println("✓ Auto-sync disabled")
	}

	return nil
}

// SyncRepairCommand runs database repair operations.
func SyncRepairCommand(args []string) error {
	fs := flag.NewFlagSet("sync repair", flag.ExitOnError)
	force := fs.Bool("force", false, "Force repair even if integrity check passes")
	_ = fs.Parse(args)

	fmt.Println("Running database repair...")

	result, err := kv.Repair(AppName, *force)
	if err != nil {
		return fmt.Errorf("repair failed: %w", err)
	}

	// Show repair results
	fmt.Println("\nRepair Results:")
	fmt.Printf("  WAL Checkpointed: %v\n", result.WalCheckpointed)
	fmt.Printf("  SHM Removed:      %v\n", result.ShmRemoved)
	fmt.Printf("  Integrity OK:     %v\n", result.IntegrityOK)
	fmt.Printf("  Vacuumed:         %v\n", result.Vacuumed)

	if result.IntegrityOK {
		fmt.Println("\n✓ Database is healthy")
	} else {
		fmt.Println("\n⚠ Database integrity check failed")
	}

	return nil
}

// SyncResetCommand resets the local database while preserving cloud data.
func SyncResetCommand(args []string) error {
	fs := flag.NewFlagSet("sync reset", flag.ExitOnError)
	_ = fs.Parse(args)

	// Ask for confirmation
	fmt.Println("⚠ WARNING: This will reset your local database!")
	fmt.Println("  - Local database will be cleared")
	fmt.Println("  - Cloud data will be preserved")
	fmt.Println("  - Next sync will download fresh data from cloud")
	fmt.Print("\nContinue? (y/N): ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	if response != "y" && response != "Y" {
		fmt.Println("Cancelled")
		return nil
	}

	if err := kv.Reset(AppName); err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}

	fmt.Println("\n✓ Local database reset")
	fmt.Println("Run 'pagen sync now' to download fresh data from cloud")

	return nil
}

// SyncWipeDBCommand completely wipes both local and cloud database.
func SyncWipeDBCommand(args []string) error {
	fs := flag.NewFlagSet("sync wipedb", flag.ExitOnError)
	_ = fs.Parse(args)

	// Ask for typed confirmation
	fmt.Println("⚠⚠⚠ DANGER: This will PERMANENTLY DELETE all data! ⚠⚠⚠")
	fmt.Println("  - Local database will be deleted")
	fmt.Println("  - Cloud backups will be deleted")
	fmt.Println("  - This action CANNOT be undone")
	fmt.Print("\nType 'wipe' to confirm: ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	if response != "wipe" {
		fmt.Println("Cancelled")
		return nil
	}

	result, err := kv.Wipe(AppName)
	if err != nil {
		return fmt.Errorf("wipe failed: %w", err)
	}

	// Show wipe results
	fmt.Println("\nWipe Results:")
	fmt.Printf("  Cloud Backups Deleted: %v\n", result.CloudBackupsDeleted)
	fmt.Printf("  Local Files Deleted:   %v\n", result.LocalFilesDeleted)

	fmt.Println("\n✓ All data wiped")
	fmt.Println("You can start fresh by adding new data")

	return nil
}
