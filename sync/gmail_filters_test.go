// ABOUTME: Comprehensive unit tests for Gmail filtering logic
// ABOUTME: Tests query building, email filtering, and address parsing
package sync

import (
	"testing"
	"time"

	"google.golang.org/api/gmail/v1"
)

// TestBuildHighSignalQuery tests Gmail query construction with different dates.
func TestBuildHighSignalQuery(t *testing.T) {
	tests := []struct {
		name      string
		userEmail string
		since     time.Time
		want      string
	}{
		{
			name:      "basic query with specific date",
			userEmail: "user@example.com",
			since:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			want:      "(from:me is:replied) OR (to:me is:replied) OR is:starred after:2024/01/15 -in:spam -in:trash",
		},
		{
			name:      "query with different year",
			userEmail: "test@test.com",
			since:     time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
			want:      "(from:me is:replied) OR (to:me is:replied) OR is:starred after:2023/12/31 -in:spam -in:trash",
		},
		{
			name:      "query with early month/day",
			userEmail: "foo@bar.com",
			since:     time.Date(2024, 3, 5, 10, 30, 0, 0, time.UTC),
			want:      "(from:me is:replied) OR (to:me is:replied) OR is:starred after:2024/03/05 -in:spam -in:trash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildHighSignalQuery(tt.userEmail, tt.since)
			if got != tt.want {
				t.Errorf("BuildHighSignalQuery() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestIsHighSignalEmail tests the main filtering logic.
func TestIsHighSignalEmail(t *testing.T) {
	tests := []struct {
		name      string
		message   *gmail.Message
		userEmail string
		wantOk    bool
		wantMsg   string
	}{
		{
			name:      "nil message",
			message:   nil,
			userEmail: "user@example.com",
			wantOk:    false,
			wantMsg:   "nil message",
		},
		{
			name: "automated sender - noreply",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "noreply@example.com"},
						{Name: "To", Value: "user@example.com"},
						{Name: "Subject", Value: "Welcome to our service"},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    false,
			wantMsg:   "automated sender",
		},
		{
			name: "automated sender - notifications",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "notifications@github.com"},
						{Name: "To", Value: "user@example.com"},
						{Name: "Subject", Value: "New issue opened"},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    false,
			wantMsg:   "automated sender",
		},
		{
			name: "group email - 5 recipients in To",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "person@example.com"},
						{Name: "To", Value: "a@x.com, b@x.com, c@x.com, d@x.com, e@x.com"},
						{Name: "Subject", Value: "Team meeting"},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    false,
			wantMsg:   "group email (5 recipients)",
		},
		{
			name: "group email - combined To and Cc",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "person@example.com"},
						{Name: "To", Value: "a@x.com, b@x.com, c@x.com"},
						{Name: "Cc", Value: "d@x.com, e@x.com"},
						{Name: "Subject", Value: "Project update"},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    false,
			wantMsg:   "group email (5 recipients)",
		},
		{
			name: "calendar invite - text/calendar MIME type",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					MimeType: "text/calendar",
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "person@example.com"},
						{Name: "To", Value: "user@example.com"},
						{Name: "Subject", Value: "Meeting tomorrow"},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    false,
			wantMsg:   "calendar invite",
		},
		{
			name: "calendar invite - invitation prefix",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "person@example.com"},
						{Name: "To", Value: "user@example.com"},
						{Name: "Subject", Value: "Invitation: Team Sync"},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    false,
			wantMsg:   "calendar invite",
		},
		{
			name: "auto-generated subject - empty",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "person@example.com"},
						{Name: "To", Value: "user@example.com"},
						{Name: "Subject", Value: ""},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    false,
			wantMsg:   "auto-generated subject",
		},
		{
			name: "auto-generated subject - out of office",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "person@example.com"},
						{Name: "To", Value: "user@example.com"},
						{Name: "Subject", Value: "Out of office: vacation"},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    false,
			wantMsg:   "auto-generated subject",
		},
		{
			name: "high signal email - normal conversation",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "colleague@example.com"},
						{Name: "To", Value: "user@example.com"},
						{Name: "Subject", Value: "Quick question about the project"},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    true,
			wantMsg:   "",
		},
		{
			name: "high signal email - small group",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "From", Value: "person@example.com"},
						{Name: "To", Value: "a@x.com, b@x.com"},
						{Name: "Cc", Value: "c@x.com"},
						{Name: "Subject", Value: "Discussion topic"},
					},
				},
			},
			userEmail: "user@example.com",
			wantOk:    true,
			wantMsg:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, gotMsg := IsHighSignalEmail(tt.message, tt.userEmail)
			if gotOk != tt.wantOk {
				t.Errorf("IsHighSignalEmail() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotMsg != tt.wantMsg {
				t.Errorf("IsHighSignalEmail() msg = %q, want %q", gotMsg, tt.wantMsg)
			}
		})
	}
}

// TestParseHeaders tests header extraction from gmail.MessagePart.
func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name    string
		payload *gmail.MessagePart
		want    map[string]string
	}{
		{
			name:    "nil payload",
			payload: nil,
			want:    map[string]string{},
		},
		{
			name: "nil headers",
			payload: &gmail.MessagePart{
				Headers: nil,
			},
			want: map[string]string{},
		},
		{
			name: "empty headers",
			payload: &gmail.MessagePart{
				Headers: []*gmail.MessagePartHeader{},
			},
			want: map[string]string{},
		},
		{
			name: "standard email headers",
			payload: &gmail.MessagePart{
				Headers: []*gmail.MessagePartHeader{
					{Name: "From", Value: "sender@example.com"},
					{Name: "To", Value: "recipient@example.com"},
					{Name: "Subject", Value: "Test email"},
					{Name: "Date", Value: "Mon, 1 Jan 2024 12:00:00 +0000"},
				},
			},
			want: map[string]string{
				"From":    "sender@example.com",
				"To":      "recipient@example.com",
				"Subject": "Test email",
				"Date":    "Mon, 1 Jan 2024 12:00:00 +0000",
			},
		},
		{
			name: "headers with cc and bcc",
			payload: &gmail.MessagePart{
				Headers: []*gmail.MessagePartHeader{
					{Name: "From", Value: "sender@example.com"},
					{Name: "To", Value: "recipient@example.com"},
					{Name: "Cc", Value: "cc@example.com"},
					{Name: "Bcc", Value: "bcc@example.com"},
					{Name: "Subject", Value: "Multi-recipient"},
				},
			},
			want: map[string]string{
				"From":    "sender@example.com",
				"To":      "recipient@example.com",
				"Cc":      "cc@example.com",
				"Bcc":     "bcc@example.com",
				"Subject": "Multi-recipient",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHeaders(tt.payload)
			if len(got) != len(tt.want) {
				t.Errorf("parseHeaders() got %d headers, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseHeaders()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

// TestIsAutomatedSender tests automated sender detection.
func TestIsAutomatedSender(t *testing.T) {
	tests := []struct {
		name string
		from string
		want bool
	}{
		{
			name: "empty from",
			from: "",
			want: true,
		},
		{
			name: "noreply",
			from: "noreply@example.com",
			want: true,
		},
		{
			name: "no-reply with hyphen",
			from: "no-reply@service.com",
			want: true,
		},
		{
			name: "donotreply",
			from: "donotreply@company.com",
			want: true,
		},
		{
			name: "do-not-reply with hyphens",
			from: "do-not-reply@site.com",
			want: true,
		},
		{
			name: "notifications",
			from: "notifications@github.com",
			want: true,
		},
		{
			name: "notify",
			from: "notify@app.com",
			want: true,
		},
		{
			name: "mailer-daemon",
			from: "MAILER-DAEMON@mail.example.com",
			want: true,
		},
		{
			name: "postmaster",
			from: "postmaster@domain.com",
			want: true,
		},
		{
			name: "bounces",
			from: "bounces@mailing.com",
			want: true,
		},
		{
			name: "unsubscribe",
			from: "unsubscribe@newsletter.com",
			want: true,
		},
		{
			name: "newsletter",
			from: "newsletter@company.com",
			want: true,
		},
		{
			name: "marketing",
			from: "marketing@business.com",
			want: true,
		},
		{
			name: "normal user",
			from: "john.doe@example.com",
			want: false,
		},
		{
			name: "person with name",
			from: "Jane Smith <jane@example.com>",
			want: false,
		},
		{
			name: "case insensitive - NOREPLY",
			from: "NOREPLY@EXAMPLE.COM",
			want: true,
		},
		{
			name: "partial match in domain",
			from: "support@notifications-service.com",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAutomatedSender(tt.from)
			if got != tt.want {
				t.Errorf("isAutomatedSender(%q) = %v, want %v", tt.from, got, tt.want)
			}
		})
	}
}

// TestCountRecipients tests recipient counting.
func TestCountRecipients(t *testing.T) {
	tests := []struct {
		name        string
		headerValue string
		want        int
	}{
		{
			name:        "empty string",
			headerValue: "",
			want:        0,
		},
		{
			name:        "single recipient",
			headerValue: "user@example.com",
			want:        1,
		},
		{
			name:        "two recipients",
			headerValue: "a@x.com, b@x.com",
			want:        2,
		},
		{
			name:        "five recipients",
			headerValue: "a@x.com, b@x.com, c@x.com, d@x.com, e@x.com",
			want:        5,
		},
		{
			name:        "recipients with extra spaces",
			headerValue: "a@x.com,  b@x.com,   c@x.com",
			want:        3,
		},
		{
			name:        "recipients with names",
			headerValue: "Alice <a@x.com>, Bob <b@x.com>, Charlie <c@x.com>",
			want:        3,
		},
		{
			name:        "trailing comma",
			headerValue: "a@x.com, b@x.com, ",
			want:        2,
		},
		{
			name:        "only whitespace after comma",
			headerValue: "a@x.com,    ",
			want:        1,
		},
		{
			name:        "multiple commas in a row",
			headerValue: "a@x.com,,,b@x.com",
			want:        2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countRecipients(tt.headerValue)
			if got != tt.want {
				t.Errorf("countRecipients(%q) = %d, want %d", tt.headerValue, got, tt.want)
			}
		})
	}
}

// TestIsCalendarInvite tests calendar invite detection.
func TestIsCalendarInvite(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		message *gmail.Message
		want    bool
	}{
		{
			name:    "nil payload",
			subject: "Meeting",
			message: &gmail.Message{Payload: nil},
			want:    false,
		},
		{
			name:    "text/calendar MIME type",
			subject: "Meeting",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					MimeType: "text/calendar",
				},
			},
			want: true,
		},
		{
			name:    "invitation prefix",
			subject: "Invitation: Team Meeting",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{},
			},
			want: true,
		},
		{
			name:    "invite prefix",
			subject: "Invite: Coffee chat",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{},
			},
			want: true,
		},
		{
			name:    "calendar prefix",
			subject: "Calendar: Weekly sync",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{},
			},
			want: true,
		},
		{
			name:    "updated invitation",
			subject: "Updated invitation: Project review",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{},
			},
			want: true,
		},
		{
			name:    "canceled event",
			subject: "Canceled event: Team lunch",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{},
			},
			want: true,
		},
		{
			name:    "case insensitive - INVITATION",
			subject: "INVITATION: IMPORTANT MEETING",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{},
			},
			want: true,
		},
		{
			name:    "invitation in middle of subject",
			subject: "Please see invitation below",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{},
			},
			want: false,
		},
		{
			name:    "normal email",
			subject: "Let's meet tomorrow",
			message: &gmail.Message{
				Payload: &gmail.MessagePart{
					MimeType: "text/plain",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCalendarInvite(tt.subject, tt.message)
			if got != tt.want {
				t.Errorf("isCalendarInvite(%q, message) = %v, want %v", tt.subject, got, tt.want)
			}
		})
	}
}

// TestIsAutoGeneratedSubject tests auto-generated subject detection.
func TestIsAutoGeneratedSubject(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		want    bool
	}{
		{
			name:    "empty subject",
			subject: "",
			want:    true,
		},
		{
			name:    "whitespace only",
			subject: "   ",
			want:    true,
		},
		{
			name:    "very short - 1 char",
			subject: "A",
			want:    true,
		},
		{
			name:    "very short - 2 chars",
			subject: "Re",
			want:    true,
		},
		{
			name:    "exactly 3 chars - should pass",
			subject: "Hey",
			want:    false,
		},
		{
			name:    "automatic reply",
			subject: "Automatic reply: I'm out",
			want:    true,
		},
		{
			name:    "out of office",
			subject: "Out of office: Vacation",
			want:    true,
		},
		{
			name:    "delivery status notification",
			subject: "Delivery Status Notification (Failure)",
			want:    true,
		},
		{
			name:    "returned mail",
			subject: "Returned mail: see transcript for details",
			want:    true,
		},
		{
			name:    "failure notice",
			subject: "Failure notice: message delivery failed",
			want:    true,
		},
		{
			name:    "undelivered mail",
			subject: "Undelivered Mail Returned to Sender",
			want:    true,
		},
		{
			name:    "case insensitive - OUT OF OFFICE",
			subject: "OUT OF OFFICE: HOLIDAY",
			want:    true,
		},
		{
			name:    "normal subject",
			subject: "Quick question about the project",
			want:    false,
		},
		{
			name:    "contains 'out' but not auto-generated",
			subject: "Let's go out for lunch",
			want:    false,
		},
		{
			name:    "contains 'automatic' but not at start",
			subject: "The automatic system is broken",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAutoGeneratedSubject(tt.subject)
			if got != tt.want {
				t.Errorf("isAutoGeneratedSubject(%q) = %v, want %v", tt.subject, got, tt.want)
			}
		})
	}
}

// TestExtractEmailAddress tests email address parsing.
func TestExtractEmailAddress(t *testing.T) {
	tests := []struct {
		name       string
		emailField string
		wantName   string
		wantEmail  string
		wantDomain string
	}{
		{
			name:       "empty string",
			emailField: "",
			wantName:   "",
			wantEmail:  "",
			wantDomain: "",
		},
		{
			name:       "plain email",
			emailField: "user@example.com",
			wantName:   "",
			wantEmail:  "user@example.com",
			wantDomain: "example.com",
		},
		{
			name:       "name and email",
			emailField: "John Doe <john@example.com>",
			wantName:   "John Doe",
			wantEmail:  "john@example.com",
			wantDomain: "example.com",
		},
		{
			name:       "quoted name",
			emailField: `"Jane Smith" <jane@example.com>`,
			wantName:   "Jane Smith",
			wantEmail:  "jane@example.com",
			wantDomain: "example.com",
		},
		{
			name:       "name with extra spaces",
			emailField: "  Bob Jones  <bob@example.com>",
			wantName:   "Bob Jones",
			wantEmail:  "bob@example.com",
			wantDomain: "example.com",
		},
		{
			name:       "email with spaces",
			emailField: "Alice <  alice@example.com  >",
			wantName:   "Alice",
			wantEmail:  "alice@example.com",
			wantDomain: "example.com",
		},
		{
			name:       "subdomain",
			emailField: "user@mail.example.com",
			wantName:   "",
			wantEmail:  "user@mail.example.com",
			wantDomain: "mail.example.com",
		},
		{
			name:       "domain case normalization",
			emailField: "user@EXAMPLE.COM",
			wantName:   "",
			wantEmail:  "user@EXAMPLE.COM",
			wantDomain: "example.com",
		},
		{
			name:       "complex name",
			emailField: "Dr. Sarah O'Brien, PhD <sarah@university.edu>",
			wantName:   "Dr. Sarah O'Brien, PhD",
			wantEmail:  "sarah@university.edu",
			wantDomain: "university.edu",
		},
		{
			name:       "no domain - invalid email",
			emailField: "invaliduser",
			wantName:   "",
			wantEmail:  "invaliduser",
			wantDomain: "",
		},
		{
			name:       "multiple @ signs - malformed",
			emailField: "user@@example.com",
			wantName:   "",
			wantEmail:  "user@@example.com",
			wantDomain: "",
		},
		{
			name:       "angle brackets but no email",
			emailField: "Just Name <>",
			wantName:   "Just Name",
			wantEmail:  "",
			wantDomain: "",
		},
		{
			name:       "only opening bracket - malformed",
			emailField: "Name <email@example.com",
			wantName:   "",
			wantEmail:  "Name <email@example.com",
			wantDomain: "example.com", // Domain still extracted from malformed email
		},
		{
			name:       "whitespace around plain email",
			emailField: "  user@example.com  ",
			wantName:   "",
			wantEmail:  "user@example.com",
			wantDomain: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotEmail, gotDomain := ExtractEmailAddress(tt.emailField)
			if gotName != tt.wantName {
				t.Errorf("ExtractEmailAddress(%q) name = %q, want %q", tt.emailField, gotName, tt.wantName)
			}
			if gotEmail != tt.wantEmail {
				t.Errorf("ExtractEmailAddress(%q) email = %q, want %q", tt.emailField, gotEmail, tt.wantEmail)
			}
			if gotDomain != tt.wantDomain {
				t.Errorf("ExtractEmailAddress(%q) domain = %q, want %q", tt.emailField, gotDomain, tt.wantDomain)
			}
		})
	}
}
