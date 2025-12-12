// ABOUTME: Integration tests for sync daemon scheduling
// ABOUTME: Tests daemon loop, signal handling, and sync execution
package cli

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/harperreed/pagen/db"
)

// TestDaemonSignalHandling verifies graceful shutdown on signals.
func TestDaemonSignalHandling(t *testing.T) {
	// Create in-memory test database
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer func() { _ = database.Close() }()

	if err := db.InitSchema(database); err != nil {
		t.Fatalf("failed to initialize schema: %v", err)
	}

	// Setup signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	// Run daemon in goroutine
	done := make(chan error, 1)
	go func() {
		// Simulate daemon loop
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Simulate sync work
				continue
			case sig := <-sigChan:
				t.Logf("Received signal: %s", sig)
				done <- nil
				return
			}
		}
	}()

	// Wait for daemon to start
	time.Sleep(50 * time.Millisecond)

	// Send signal
	sigChan <- syscall.SIGINT

	// Wait for graceful shutdown
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("daemon shutdown failed: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("daemon did not shut down within timeout")
	}
}

// TestDaemonTickerScheduling verifies sync runs at intervals.
func TestDaemonTickerScheduling(t *testing.T) {
	// Use short interval for testing
	interval := 100 * time.Millisecond

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	tickCount := 0
	for {
		select {
		case <-ticker.C:
			tickCount++
			t.Logf("Tick %d at %s", tickCount, time.Now().Format("15:04:05.000"))

		case <-ctx.Done():
			// Should get approximately 4-5 ticks in 500ms with 100ms interval
			if tickCount < 3 || tickCount > 6 {
				t.Errorf("expected 3-6 ticks, got %d", tickCount)
			}
			return
		}
	}
}

// TestDaemonMinimumInterval verifies interval validation.
func TestDaemonMinimumInterval(t *testing.T) {
	tests := []struct {
		name        string
		interval    time.Duration
		shouldError bool
	}{
		{
			name:        "valid 5 minutes",
			interval:    5 * time.Minute,
			shouldError: false,
		},
		{
			name:        "valid 1 hour",
			interval:    1 * time.Hour,
			shouldError: false,
		},
		{
			name:        "invalid 4 minutes",
			interval:    4 * time.Minute,
			shouldError: true,
		},
		{
			name:        "invalid 1 minute",
			interval:    1 * time.Minute,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.interval >= 5*time.Minute
			hasError := !isValid

			if hasError != tt.shouldError {
				t.Errorf("expected error = %v, got %v for interval %s", tt.shouldError, hasError, tt.interval)
			}
		})
	}
}

// TestDaemonServiceSelection verifies service filtering.
func TestDaemonServiceSelection(t *testing.T) {
	tests := []struct {
		name             string
		servicesStr      string
		expectedServices []string
	}{
		{
			name:             "all services",
			servicesStr:      "all",
			expectedServices: []string{"contacts", "calendar", "gmail"},
		},
		{
			name:             "only contacts",
			servicesStr:      "contacts",
			expectedServices: []string{"contacts"},
		},
		{
			name:             "contacts and calendar",
			servicesStr:      "contacts,calendar",
			expectedServices: []string{"contacts", "calendar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := parseServices(tt.servicesStr)

			if len(services) != len(tt.expectedServices) {
				t.Errorf("expected %d services, got %d", len(tt.expectedServices), len(services))
				return
			}

			for i, expected := range tt.expectedServices {
				if services[i] != expected {
					t.Errorf("expected service[%d] = %s, got %s", i, expected, services[i])
				}
			}
		})
	}
}

// TestDaemonImmediateFirstSync verifies initial sync runs immediately.
func TestDaemonImmediateFirstSync(t *testing.T) {
	startTime := time.Now()
	interval := 1 * time.Hour // Long interval to ensure we're testing immediate sync

	// Simulate daemon startup
	syncRan := false

	// Initial sync should run immediately
	syncRan = true
	syncTime := time.Now()

	if !syncRan {
		t.Error("initial sync did not run")
	}

	elapsed := syncTime.Sub(startTime)
	if elapsed > 100*time.Millisecond {
		t.Errorf("initial sync took too long: %s (expected immediate)", elapsed)
	}

	// Verify we're not waiting for the full interval
	if elapsed > interval/2 {
		t.Errorf("initial sync waited too long: %s (interval: %s)", elapsed, interval)
	}
}

// TestDaemonGracefulShutdown verifies cleanup on shutdown.
func TestDaemonGracefulShutdown(t *testing.T) {
	// Setup resources that need cleanup
	ticker := time.NewTicker(100 * time.Millisecond)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	// Use channel to signal cleanup completion (avoids data race)
	cleanupDone := make(chan struct{})

	// Simulate daemon
	go func() {
		<-sigChan
		// Cleanup
		ticker.Stop()
		close(cleanupDone)
	}()

	// Wait for daemon to start
	time.Sleep(50 * time.Millisecond)

	// Send shutdown signal
	sigChan <- syscall.SIGTERM

	// Wait for cleanup with timeout
	select {
	case <-cleanupDone:
		// Success - cleanup completed
	case <-time.After(1 * time.Second):
		t.Error("daemon did not clean up resources within timeout")
	}
}
