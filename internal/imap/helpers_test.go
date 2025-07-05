package imap

import (
	"strings"
	"testing"
	"time"

	"inbox451/internal/models"

	"github.com/emersion/go-imap"
	"github.com/stretchr/testify/assert"
	null "github.com/volatiletech/null/v9"
)

func TestParseEmailToAddress(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		expectedAddr *imap.Address
		expectErr    bool
		expectedErr  string
	}{
		{
			name:  "valid email with name",
			email: "John Doe <john.doe@example.com>",
			expectedAddr: &imap.Address{
				PersonalName: "John Doe",
				AtDomainList: "",
				MailboxName:  "john.doe",
				HostName:     "example.com",
			},
			expectErr: false,
		},
		{
			name:  "valid email without name",
			email: "john.doe@example.com",
			expectedAddr: &imap.Address{
				PersonalName: "",
				AtDomainList: "",
				MailboxName:  "john.doe",
				HostName:     "example.com",
			},
			expectErr: false,
		},
		{
			name:  "email with special characters in name",
			email: "\"John O'Doe\" <john@example.com>",
			expectedAddr: &imap.Address{
				PersonalName: "John O'Doe",
				AtDomainList: "",
				MailboxName:  "john",
				HostName:     "example.com",
			},
			expectErr: false,
		},
		{
			name:  "simple email fallback",
			email: "simple@example.com",
			expectedAddr: &imap.Address{
				PersonalName: "",
				AtDomainList: "",
				MailboxName:  "simple",
				HostName:     "example.com",
			},
			expectErr: false,
		},
		{
			name:         "invalid email without @",
			email:        "invalidemail",
			expectedAddr: nil,
			expectErr:    true,
			expectedErr:  "invalid email format",
		},
		{
			name:         "invalid email with multiple @",
			email:        "user@@example.com",
			expectedAddr: nil,
			expectErr:    true,
			expectedErr:  "invalid email format",
		},
		{
			name:         "empty email",
			email:        "",
			expectedAddr: nil,
			expectErr:    true,
			expectedErr:  "invalid email format",
		},
		{
			name:  "email with subdomain",
			email: "user@mail.example.com",
			expectedAddr: &imap.Address{
				PersonalName: "",
				AtDomainList: "",
				MailboxName:  "user",
				HostName:     "mail.example.com",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := parseEmailToAddress(tt.email)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, addr)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, addr)
				assert.Equal(t, tt.expectedAddr.PersonalName, addr.PersonalName)
				assert.Equal(t, tt.expectedAddr.AtDomainList, addr.AtDomainList)
				assert.Equal(t, tt.expectedAddr.MailboxName, addr.MailboxName)
				assert.Equal(t, tt.expectedAddr.HostName, addr.HostName)
			}
		})
	}
}

func TestReconstructRFC822(t *testing.T) {
	now := time.Now()
	message := &models.Message{
		Base: models.Base{
			ID:        "test-message-1",
			CreatedAt: null.TimeFrom(now),
		},
		InboxID:  "test-inbox-1",
		Sender:   "sender@example.com",
		Receiver: "receiver@example.com",
		Subject:  "Test Subject",
		Body:     "This is a test message body.",
	}

	result := reconstructRFC822(message)

	// Check that the RFC822 message contains expected headers
	assert.Contains(t, result, "From: sender@example.com")
	assert.Contains(t, result, "To: receiver@example.com")
	assert.Contains(t, result, "Subject: Test Subject")
	assert.Contains(t, result, "Content-Type: text/plain; charset=UTF-8")
	assert.Contains(t, result, "MIME-Version: 1.0")
	assert.Contains(t, result, "This is a test message body.")

	// Check that date is formatted properly
	expectedDate := now.Format(time.RFC1123Z)
	assert.Contains(t, result, "Date: "+expectedDate)

	// Check that headers and body are separated by empty line
	assert.Contains(t, result, "\r\n\r\n")

	// Check that the body comes after the headers
	parts := strings.Split(result, "\r\n\r\n")
	assert.Len(t, parts, 2)
	assert.Equal(t, "This is a test message body.", parts[1])
}

func TestMatchesSearchCriteria(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	message := &models.Message{
		Base: models.Base{
			ID:        "test-message-1",
			CreatedAt: null.TimeFrom(now),
		},
		InboxID:  "test-inbox-1",
		Sender:   "sender@example.com",
		Receiver: "receiver@example.com",
		Subject:  "Important Subject",
		Body:     "This is an important message body with keyword test.",
	}

	tests := []struct {
		name     string
		criteria *imap.SearchCriteria
		expected bool
	}{
		{
			name: "matches header from",
			criteria: &imap.SearchCriteria{
				Header: map[string][]string{
					"from": {"sender@example.com"},
				},
			},
			expected: true,
		},
		{
			name: "matches header from case insensitive",
			criteria: &imap.SearchCriteria{
				Header: map[string][]string{
					"from": {"SENDER@EXAMPLE.COM"},
				},
			},
			expected: true,
		},
		{
			name: "does not match header from",
			criteria: &imap.SearchCriteria{
				Header: map[string][]string{
					"from": {"other@example.com"},
				},
			},
			expected: false,
		},
		{
			name: "matches header to",
			criteria: &imap.SearchCriteria{
				Header: map[string][]string{
					"to": {"receiver@example.com"},
				},
			},
			expected: true,
		},
		{
			name: "matches header subject",
			criteria: &imap.SearchCriteria{
				Header: map[string][]string{
					"subject": {"Important"},
				},
			},
			expected: true,
		},
		{
			name: "matches body search",
			criteria: &imap.SearchCriteria{
				Body: []string{"important"},
			},
			expected: true,
		},
		{
			name: "does not match body search",
			criteria: &imap.SearchCriteria{
				Body: []string{"nonexistent"},
			},
			expected: false,
		},
		{
			name: "matches text search in body",
			criteria: &imap.SearchCriteria{
				Text: []string{"keyword"},
			},
			expected: true,
		},
		{
			name: "matches text search in subject",
			criteria: &imap.SearchCriteria{
				Text: []string{"Subject"},
			},
			expected: true,
		},
		{
			name: "matches text search in sender",
			criteria: &imap.SearchCriteria{
				Text: []string{"sender"},
			},
			expected: true,
		},
		{
			name: "matches text search in receiver",
			criteria: &imap.SearchCriteria{
				Text: []string{"receiver"},
			},
			expected: true,
		},
		{
			name: "does not match text search",
			criteria: &imap.SearchCriteria{
				Text: []string{"nonexistent"},
			},
			expected: false,
		},
		{
			name: "matches since date",
			criteria: &imap.SearchCriteria{
				Since: pastTime,
			},
			expected: true,
		},
		{
			name: "does not match since date",
			criteria: &imap.SearchCriteria{
				Since: futureTime,
			},
			expected: false,
		},
		{
			name: "matches before date",
			criteria: &imap.SearchCriteria{
				Before: futureTime,
			},
			expected: true,
		},
		{
			name: "does not match before date",
			criteria: &imap.SearchCriteria{
				Before: pastTime,
			},
			expected: false,
		},
		{
			name: "matches multiple criteria",
			criteria: &imap.SearchCriteria{
				Header: map[string][]string{
					"from": {"sender"},
				},
				Body:  []string{"important"},
				Since: pastTime,
			},
			expected: true,
		},
		{
			name: "fails multiple criteria",
			criteria: &imap.SearchCriteria{
				Header: map[string][]string{
					"from": {"sender"},
				},
				Body: []string{"nonexistent"},
			},
			expected: false,
		},
		{
			name:     "empty criteria (matches all)",
			criteria: &imap.SearchCriteria{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesSearchCriteria(message, tt.criteria)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildEnvelope(t *testing.T) {
	now := time.Now()
	message := &models.Message{
		Base: models.Base{
			ID:        "test-message-1",
			CreatedAt: null.TimeFrom(now),
		},
		InboxID:  "test-inbox-1",
		Sender:   "John Doe <john@example.com>",
		Receiver: "Jane Smith <jane@example.com>",
		Subject:  "Test Subject",
		Body:     "Test body",
	}

	envelope, err := buildEnvelope(message)

	assert.NoError(t, err)
	assert.NotNil(t, envelope)

	// Check basic fields
	assert.Equal(t, "Test Subject", envelope.Subject)
	assert.Equal(t, now, envelope.Date)

	// Check From field
	assert.Len(t, envelope.From, 1)
	assert.Equal(t, "John Doe", envelope.From[0].PersonalName)
	assert.Equal(t, "john", envelope.From[0].MailboxName)
	assert.Equal(t, "example.com", envelope.From[0].HostName)

	// Check Sender field (should be same as From)
	assert.Len(t, envelope.Sender, 1)
	assert.Equal(t, envelope.From[0], envelope.Sender[0])

	// Check To field
	assert.Len(t, envelope.To, 1)
	assert.Equal(t, "Jane Smith", envelope.To[0].PersonalName)
	assert.Equal(t, "jane", envelope.To[0].MailboxName)
	assert.Equal(t, "example.com", envelope.To[0].HostName)
}

func TestBuildImapMessage(t *testing.T) {
	now := time.Now()
	message := &models.Message{
		Base: models.Base{
			ID:        "test-message-123",
			CreatedAt: null.TimeFrom(now),
		},
		InboxID:   "test-inbox-123",
		Sender:    "sender@example.com",
		Receiver:  "receiver@example.com",
		Subject:   "Test Subject",
		Body:      "Test message body",
		IsRead:    true,
		IsDeleted: false,
	}

	tests := []struct {
		name     string
		items    []imap.FetchItem
		validate func(*testing.T, *imap.Message)
	}{
		{
			name:  "fetch flags",
			items: []imap.FetchItem{imap.FetchFlags},
			validate: func(t *testing.T, msg *imap.Message) {
				flags, ok := msg.Items[imap.FetchFlags].([]string)
				assert.True(t, ok)
				assert.Contains(t, flags, imap.SeenFlag)
				assert.NotContains(t, flags, imap.DeletedFlag)
			},
		},
		{
			name:  "fetch UID",
			items: []imap.FetchItem{imap.FetchUid},
			validate: func(t *testing.T, msg *imap.Message) {
				uid, ok := msg.Items[imap.FetchUid].(uint32)
				assert.True(t, ok)
				assert.Equal(t, uint32(0xd35cc9c2), uid) // Hash of "test-message-123"
			},
		},
		{
			name:  "fetch internal date",
			items: []imap.FetchItem{imap.FetchInternalDate},
			validate: func(t *testing.T, msg *imap.Message) {
				date, ok := msg.Items[imap.FetchInternalDate].(time.Time)
				assert.True(t, ok)
				assert.Equal(t, now, date)
			},
		},
		{
			name:  "fetch RFC822 size",
			items: []imap.FetchItem{imap.FetchRFC822Size},
			validate: func(t *testing.T, msg *imap.Message) {
				size, ok := msg.Items[imap.FetchRFC822Size].(uint32)
				assert.True(t, ok)
				assert.Greater(t, size, uint32(0))
			},
		},
		{
			name:  "fetch envelope",
			items: []imap.FetchItem{imap.FetchEnvelope},
			validate: func(t *testing.T, msg *imap.Message) {
				envelope, ok := msg.Items[imap.FetchEnvelope].(*imap.Envelope)
				assert.True(t, ok)
				assert.Equal(t, "Test Subject", envelope.Subject)
				assert.Equal(t, now, envelope.Date)
			},
		},
		{
			name:  "fetch body structure",
			items: []imap.FetchItem{imap.FetchBodyStructure},
			validate: func(t *testing.T, msg *imap.Message) {
				bodyStructure, ok := msg.Items[imap.FetchBodyStructure].(*imap.BodyStructure)
				assert.True(t, ok)
				assert.Equal(t, "text", bodyStructure.MIMEType)
				assert.Equal(t, "plain", bodyStructure.MIMESubType)
				assert.Equal(t, "UTF-8", bodyStructure.Params["charset"])
			},
		},
		{
			name:  "fetch body",
			items: []imap.FetchItem{imap.FetchBody},
			validate: func(t *testing.T, msg *imap.Message) {
				body, ok := msg.Items[imap.FetchBody].(*strings.Reader)
				assert.True(t, ok)

				// Read the body content
				content := make([]byte, len("Test message body"))
				n, err := body.Read(content)
				assert.NoError(t, err)
				assert.Equal(t, len("Test message body"), n)
				assert.Equal(t, "Test message body", string(content))
			},
		},
		{
			name:  "fetch RFC822",
			items: []imap.FetchItem{imap.FetchRFC822},
			validate: func(t *testing.T, msg *imap.Message) {
				rfc822, ok := msg.Items[imap.FetchRFC822].(*strings.Reader)
				assert.True(t, ok)

				// Read some content to verify it's a valid RFC822 message
				content := make([]byte, 1024)
				n, err := rfc822.Read(content)
				assert.NoError(t, err)
				assert.Greater(t, n, 0)

				rfc822Content := string(content[:n])
				assert.Contains(t, rfc822Content, "From: sender@example.com")
				assert.Contains(t, rfc822Content, "To: receiver@example.com")
				assert.Contains(t, rfc822Content, "Subject: Test Subject")
			},
		},
		{
			name:  "fetch multiple items",
			items: []imap.FetchItem{imap.FetchFlags, imap.FetchUid, imap.FetchInternalDate},
			validate: func(t *testing.T, msg *imap.Message) {
				// Verify all items are present
				assert.Contains(t, msg.Items, imap.FetchFlags)
				assert.Contains(t, msg.Items, imap.FetchUid)
				assert.Contains(t, msg.Items, imap.FetchInternalDate)

				// Verify specific values
				flags := msg.Items[imap.FetchFlags].([]string)
				assert.Contains(t, flags, imap.SeenFlag)

				uid := msg.Items[imap.FetchUid].(uint32)
				assert.Equal(t, uint32(0xd35cc9c2), uid) // Hash of "test-message-123"

				date := msg.Items[imap.FetchInternalDate].(time.Time)
				assert.Equal(t, now, date)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildImapMessage(message, 1, tt.items)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, uint32(1), result.SeqNum)

			tt.validate(t, result)
		})
	}
}

func TestBuildImapMessage_DeletedFlags(t *testing.T) {
	message := &models.Message{
		Base: models.Base{
			ID:        "test-message-1",
			CreatedAt: null.TimeFrom(time.Now()),
		},
		InboxID:   "test-inbox-123",
		Sender:    "sender@example.com",
		Receiver:  "receiver@example.com",
		Subject:   "Test",
		Body:      "Test",
		IsRead:    false,
		IsDeleted: true,
	}

	result, err := buildImapMessage(message, 1, []imap.FetchItem{imap.FetchFlags})

	assert.NoError(t, err)
	flags := result.Items[imap.FetchFlags].([]string)
	assert.Contains(t, flags, imap.DeletedFlag)
	assert.NotContains(t, flags, imap.SeenFlag)
}
