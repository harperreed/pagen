// ABOUTME: Unit tests for sync daemon mode
// ABOUTME: Tests interval parsing, service selection, and error handling
package cli

import (
	"testing"
	"time"
)

func TestParseServices(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "all services",
			input:    "all",
			expected: []string{"contacts", "calendar", "gmail"},
		},
		{
			name:     "single service",
			input:    "contacts",
			expected: []string{"contacts"},
		},
		{
			name:     "multiple services",
			input:    "contacts,calendar",
			expected: []string{"contacts", "calendar"},
		},
		{
			name:     "spaces around commas",
			input:    "contacts, calendar, gmail",
			expected: []string{"contacts", "calendar", "gmail"},
		},
		{
			name:     "invalid service ignored",
			input:    "contacts,invalid,calendar",
			expected: []string{"contacts", "calendar"},
		},
		{
			name:     "all invalid services",
			input:    "invalid,unknown",
			expected: []string{},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseServices(tt.input)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d services, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			// Check each service
			for i, service := range tt.expected {
				if result[i] != service {
					t.Errorf("expected service[%d] = %s, got %s", i, service, result[i])
				}
			}
		})
	}
}

func TestIntervalValidation(t *testing.T) {
	tests := []struct {
		name        string
		interval    string
		shouldParse bool
		minValid    bool
	}{
		{
			name:        "valid 1 hour",
			interval:    "1h",
			shouldParse: true,
			minValid:    true,
		},
		{
			name:        "valid 15 minutes",
			interval:    "15m",
			shouldParse: true,
			minValid:    true,
		},
		{
			name:        "valid 5 minutes (minimum)",
			interval:    "5m",
			shouldParse: true,
			minValid:    true,
		},
		{
			name:        "invalid 4 minutes (below minimum)",
			interval:    "4m",
			shouldParse: true,
			minValid:    false,
		},
		{
			name:        "invalid 1 minute",
			interval:    "1m",
			shouldParse: true,
			minValid:    false,
		},
		{
			name:        "valid 24 hours",
			interval:    "24h",
			shouldParse: true,
			minValid:    true,
		},
		{
			name:        "invalid format",
			interval:    "invalid",
			shouldParse: false,
			minValid:    false,
		},
		{
			name:        "empty string",
			interval:    "",
			shouldParse: false,
			minValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := time.ParseDuration(tt.interval)

			if tt.shouldParse {
				if err != nil {
					t.Errorf("expected interval to parse, got error: %v", err)
					return
				}

				isValid := duration >= 5*time.Minute
				if isValid != tt.minValid {
					t.Errorf("expected minimum validation = %v, got %v (duration: %s)", tt.minValid, isValid, duration)
				}
			} else {
				if err == nil {
					t.Errorf("expected interval to fail parsing, but it succeeded: %s", duration)
				}
			}
		})
	}
}

func TestFormatTimeSince(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "just now (30 seconds)",
			time:     now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			time:     now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			time:     now.Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			time:     now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			time:     now.Add(-3 * time.Hour),
			expected: "3 hours ago",
		},
		{
			name:     "1 day ago",
			time:     now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "5 days ago",
			time:     now.Add(-5 * 24 * time.Hour),
			expected: "5 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeSince(tt.time)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
