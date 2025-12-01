// ABOUTME: Google People API client for contacts sync
// ABOUTME: Creates authenticated People API service for fetching Google Contacts
package sync

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

// NewPeopleClient creates a new Google People API client.
func NewPeopleClient(token *oauth2.Token) (*people.Service, error) {
	if token == nil {
		return nil, fmt.Errorf("token cannot be nil")
	}

	config := NewOAuthConfig()
	client := config.Client(context.Background(), token)

	service, err := people.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create People service: %w", err)
	}

	return service, nil
}
