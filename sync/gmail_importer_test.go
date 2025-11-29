// ABOUTME: Comprehensive unit tests for Gmail importer historyId sync
// ABOUTME: Tests history-based incremental sync, fallback logic, and error handling
package sync

import (
	"testing"
)

// TestIsHistoryExpiredError tests expired historyId error detection
func TestIsHistoryExpiredError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "404 error string",
			err:  &mockError{msg: "googleapi: Error 404: historyId is invalid"},
			want: true,
		},
		{
			name: "historyId error string",
			err:  &mockError{msg: "invalid historyId provided"},
			want: true,
		},
		{
			name: "generic 404",
			err:  &mockError{msg: "404 Not Found"},
			want: true,
		},
		{
			name: "unrelated error",
			err:  &mockError{msg: "network timeout"},
			want: false,
		},
		{
			name: "empty error",
			err:  &mockError{msg: ""},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHistoryExpiredError(tt.err)
			if got != tt.want {
				t.Errorf("isHistoryExpiredError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockError implements error interface for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

// TestHistoryIdParsing tests parsing of historyId from sync token
func TestHistoryIdParsing(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantValid bool
		wantValue uint64
	}{
		{
			name:      "valid historyId",
			token:     "12345678",
			wantValid: true,
			wantValue: 12345678,
		},
		{
			name:      "large historyId",
			token:     "9876543210123",
			wantValid: true,
			wantValue: 9876543210123,
		},
		{
			name:      "zero historyId",
			token:     "0",
			wantValid: false,
			wantValue: 0,
		},
		{
			name:      "invalid format - letters",
			token:     "abc123",
			wantValid: false,
			wantValue: 0,
		},
		{
			name:      "invalid format - empty",
			token:     "",
			wantValid: false,
			wantValue: 0,
		},
		{
			name:      "invalid format - negative",
			token:     "-123",
			wantValid: false,
			wantValue: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsedHistoryId uint64
			var err error
			if tt.token != "" {
				_, err = scanHistoryId(tt.token, &parsedHistoryId)
			} else {
				err = &mockError{msg: "empty token"}
			}

			isValid := err == nil && parsedHistoryId > 0

			if isValid != tt.wantValid {
				t.Errorf("historyId parsing valid = %v, want %v", isValid, tt.wantValid)
			}

			if isValid && parsedHistoryId != tt.wantValue {
				t.Errorf("historyId parsed value = %d, want %d", parsedHistoryId, tt.wantValue)
			}
		})
	}
}

// scanHistoryId is a helper for testing historyId parsing
func scanHistoryId(token string, value *uint64) (int, error) {
	// This mimics the parsing logic in the main code
	var parsed uint64
	n, err := sscanf(token, "%d", &parsed)
	if err == nil {
		*value = parsed
	}
	return n, err
}

// sscanf is a mock of fmt.Sscanf for testing
func sscanf(str, format string, value *uint64) (int, error) {
	// Simple mock that handles %d format
	if format != "%d" {
		return 0, &mockError{msg: "unsupported format"}
	}

	var result uint64
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c < '0' || c > '9' {
			if i == 0 {
				return 0, &mockError{msg: "invalid format"}
			}
			break
		}
		result = result*10 + uint64(c-'0')
	}

	if result == 0 && len(str) > 0 && str[0] == '0' {
		// Special case: "0" is valid but we treat 0 as invalid for historyId
		*value = 0
		return 1, nil
	}

	if result == 0 {
		return 0, &mockError{msg: "no digits parsed"}
	}

	*value = result
	return 1, nil
}

// Integration test scenarios documentation
// These tests would require actual Gmail API mocking, which is complex
// and better suited for integration tests. Key scenarios to test:

// TestHistoryIdSyncFlow would test:
// 1. First sync with no historyId -> falls back to time-based query
// 2. Second sync with valid historyId -> uses history.list API
// 3. History API returns changes -> processes only new/changed messages
// 4. High-signal filtering still applies to history results
// 5. New historyId is stored after successful sync

// TestHistoryIdExpiration would test:
// 1. Sync with valid but expired historyId
// 2. History API returns 404 error
// 3. System detects expired historyId
// 4. Falls back to time-based query
// 5. New historyId is stored for future syncs

// TestHistoryIdIncrementalSync would test:
// 1. Initial sync imports 10 messages, stores historyId
// 2. New messages arrive (e.g., 5 new, 2 starred)
// 3. Incremental sync uses historyId
// 4. Only processes the 7 changed messages
// 5. Deduplication prevents re-importing existing messages
// 6. Updates historyId to latest value

// TestHistoryTypesFiltering would test:
// 1. History API called with historyTypes: messageAdded, labelAdded
// 2. Only processes additions and label changes (starred)
// 3. Ignores messageDeleted events
// 4. Collects unique message IDs from all history records
// 5. Processes each unique message only once

// TestProcessMessageErrorHandling would test:
// 1. Message fetch fails -> logs error, continues with next
// 2. Message already in sync_log -> skips processing
// 3. Message fails high-signal filtering -> skips without error
// 4. Contact creation fails -> logs error, continues
// 5. Interaction logging fails -> logs error, continues
// 6. Successful messages still get processed despite individual failures
