// ABOUTME: WAL concurrency tests for Charm KV
// ABOUTME: Verifies multiple connections can access the database simultaneously

package charm

import (
	"sync"
	"testing"

	"github.com/charmbracelet/charm/kv"
)

func TestWALConcurrentConnections(t *testing.T) {
	// Test that multiple KV connections can open the same database concurrently.
	// This verifies the WAL mode fix prevents SQLITE_BUSY errors.
	tmpDir := t.TempDir()
	t.Setenv("CHARM_DATA_DIR", tmpDir)

	// First, initialize the database with a single connection.
	initKV, err := kv.OpenWithDefaults("pagen-wal-test")
	if err != nil {
		t.Fatalf("failed to initialize: %v", err)
	}
	_ = initKV.Close()

	const numConnections = 3
	const writesPerConnection = 5

	var wg sync.WaitGroup
	errors := make(chan error, numConnections*(writesPerConnection+1))

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each goroutine opens its own KV (simulates separate processes)
			kvStore, err := kv.OpenWithDefaults("pagen-wal-test")
			if err != nil {
				errors <- err
				return
			}
			defer func() { _ = kvStore.Close() }()

			// Perform writes
			for j := 0; j < writesPerConnection; j++ {
				key := []byte("test-key")
				value := []byte("test-value")
				if err := kvStore.Set(key, value); err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Collect any errors
	errs := make([]error, 0, numConnections*writesPerConnection)
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		t.Errorf("concurrent connections produced %d errors, first: %v", len(errs), errs[0])
	}
}
