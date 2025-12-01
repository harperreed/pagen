// ABOUTME: Calendar API client setup for Google Calendar integration
// ABOUTME: Creates authenticated Calendar service from OAuth token
package sync

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// NewCalendarClient creates a Google Calendar API service from an OAuth token.
func NewCalendarClient(token *oauth2.Token) (*calendar.Service, error) {
	if token == nil {
		return nil, fmt.Errorf("token cannot be nil")
	}

	ctx := context.Background()
	config := NewOAuthConfig()

	// Create HTTP client from token
	client := config.Client(ctx, token)

	// Create Calendar API service
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return service, nil
}
