package imap

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/mail"
	"strings"
	"time"

	"inbox451/internal/models"

	"github.com/emersion/go-imap"
)

// stringToUID converts a string ID (UUID) to a uint32 UID for IMAP
// Uses FNV-1a hash to ensure consistent mapping
func stringToUID(id string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(id))
	return h.Sum32()
}

// uidToString converts a uint32 UID back to string ID
// This requires looking up the actual message ID from the database
// as the hash is one-way
func uidToString(uid uint32, messages []*models.Message) string {
	for _, msg := range messages {
		if stringToUID(msg.ID) == uid {
			return msg.ID
		}
	}
	return ""
}

// parseEmailToAddress converts email string to IMAP address
func parseEmailToAddress(email string) (*imap.Address, error) {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		// If parsing fails, treat as plain email
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid email format: %s", email)
		}
		return &imap.Address{
			PersonalName: "",
			AtDomainList: "",
			MailboxName:  parts[0],
			HostName:     parts[1],
		}, nil
	}

	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid email format: %s", addr.Address)
	}

	return &imap.Address{
		PersonalName: addr.Name,
		AtDomainList: "",
		MailboxName:  parts[0],
		HostName:     parts[1],
	}, nil
}

// reconstructRFC822 builds a minimal RFC822 message from database data
func reconstructRFC822(dbMsg *models.Message) string {
	var sb strings.Builder

	// Headers
	sb.WriteString("From: " + dbMsg.Sender + "\r\n")
	sb.WriteString("To: " + dbMsg.Receiver + "\r\n")
	sb.WriteString("Subject: " + dbMsg.Subject + "\r\n")
	sb.WriteString("Date: " + dbMsg.CreatedAt.Time.Format(time.RFC1123Z) + "\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("\r\n")

	// Body
	sb.WriteString(dbMsg.Body)

	return sb.String()
}

// matchesSearchCriteria checks if message matches additional search criteria
func matchesSearchCriteria(msg *models.Message, criteria *imap.SearchCriteria) bool {
	// Handle header criteria
	for key, values := range criteria.Header {
		keyLower := strings.ToLower(key)
		for _, value := range values {
			valueLower := strings.ToLower(value)

			switch keyLower {
			case "from":
				if !strings.Contains(strings.ToLower(msg.Sender), valueLower) {
					return false
				}
			case "to":
				if !strings.Contains(strings.ToLower(msg.Receiver), valueLower) {
					return false
				}
			case "subject":
				if !strings.Contains(strings.ToLower(msg.Subject), valueLower) {
					return false
				}
			}
		}
	}

	// Handle body search
	for _, bodyText := range criteria.Body {
		if !strings.Contains(strings.ToLower(msg.Body), strings.ToLower(bodyText)) {
			return false
		}
	}

	// Handle text search (body + headers)
	for _, text := range criteria.Text {
		searchText := strings.ToLower(text)
		if !strings.Contains(strings.ToLower(msg.Body), searchText) &&
			!strings.Contains(strings.ToLower(msg.Subject), searchText) &&
			!strings.Contains(strings.ToLower(msg.Sender), searchText) &&
			!strings.Contains(strings.ToLower(msg.Receiver), searchText) {
			return false
		}
	}

	// Handle date criteria
	if !criteria.Since.IsZero() && msg.CreatedAt.Time.Before(criteria.Since) {
		return false
	}
	if !criteria.Before.IsZero() && msg.CreatedAt.Time.After(criteria.Before) {
		return false
	}

	// If no additional criteria were specified (like headers, body, text, dates),
	// this message matches (flags are handled in the calling function)
	return true
}

// buildEnvelope creates IMAP envelope from database message
func buildEnvelope(dbMsg *models.Message) (*imap.Envelope, error) {
	env := &imap.Envelope{
		Date:    dbMsg.CreatedAt.Time,
		Subject: dbMsg.Subject,
	}

	// Parse sender
	if fromAddr, err := parseEmailToAddress(dbMsg.Sender); err == nil {
		env.From = []*imap.Address{fromAddr}
		env.Sender = []*imap.Address{fromAddr}
	}

	// Parse receiver
	if toAddr, err := parseEmailToAddress(dbMsg.Receiver); err == nil {
		env.To = []*imap.Address{toAddr}
	}

	return env, nil
}

// buildImapMessage converts database message to IMAP message
func buildImapMessage(dbMsg *models.Message, seqNum uint32, items []imap.FetchItem) (*imap.Message, error) {
	imapMsg := &imap.Message{
		SeqNum: seqNum,
		Items:  make(map[imap.FetchItem]any),
		Body:   make(map[*imap.BodySectionName]imap.Literal),
	}

	for _, item := range items {
		switch item {
		case imap.FetchFlags:
			flags := []string{}
			if dbMsg.IsRead {
				flags = append(flags, imap.SeenFlag)
			}
			if dbMsg.IsDeleted {
				flags = append(flags, imap.DeletedFlag)
			}
			// Store flags in both the Items map and the Flags field
			imapMsg.Items[item] = flags
			imapMsg.Flags = flags

		case imap.FetchUid:
			imapMsg.Items[item] = stringToUID(dbMsg.ID)

		case imap.FetchInternalDate:
			imapMsg.Items[item] = dbMsg.CreatedAt.Time

		case imap.FetchRFC822Size:
			// Estimate size (headers + body)
			size := len(dbMsg.Subject) + len(dbMsg.Sender) + len(dbMsg.Receiver) + len(dbMsg.Body) + 200 // headers overhead
			imapMsg.Items[item] = uint32(size)

		case imap.FetchEnvelope:
			env, err := buildEnvelope(dbMsg)
			if err != nil {
				return nil, err
			}
			imapMsg.Items[item] = env
			imapMsg.Envelope = env

		case imap.FetchBodyStructure:
			bodyStructure := &imap.BodyStructure{
				MIMEType:    "text",
				MIMESubType: "plain",
				Params:      map[string]string{"charset": "UTF-8"},
				Size:        uint32(len(dbMsg.Body)),
			}
			imapMsg.Items[item] = bodyStructure

		case imap.FetchBody:
			// Return the plain text body
			imapMsg.Items[item] = strings.NewReader(dbMsg.Body)

		case imap.FetchRFC822:
			// Reconstruct full RFC822 message
			rfc822 := reconstructRFC822(dbMsg)
			imapMsg.Items[item] = strings.NewReader(rfc822)

		default:
			// Handle body section requests
			itemStr := string(item)
			if strings.HasPrefix(itemStr, "BODY[") || strings.HasPrefix(itemStr, "BODY.PEEK[") {
				// Parse the body section request
				section := &imap.BodySectionName{}
				if strings.HasPrefix(itemStr, "BODY.PEEK[") {
					section.Peek = true
				}

				rfc822 := reconstructRFC822(dbMsg)

				// Add to both Body map and Items map for proper IMAP response
				imapMsg.Body[section] = strings.NewReader(rfc822)
				imapMsg.Items[item] = strings.NewReader(rfc822)
			}
		}
	}

	return imapMsg, nil
}

// uidToStringByLookup converts a uint32 UID back to string ID by looking up messages
// This is a helper for the mailbox functions that need to convert UIDs to message IDs
func uidToStringByLookup(ctx context.Context, m *ImapMailbox, uid uint32) string {
	// Get all messages from the inbox to find the matching UID
	filters := models.MessageFilters{} // Get all messages including deleted
	messages, _, err := m.user.core.Repository.ListMessagesByInboxWithFilters(ctx, m.inboxModel.ID, filters, 0, 0)
	if err != nil {
		m.user.core.Logger.Error("Failed to lookup message by UID %d: %v", uid, err)
		return ""
	}

	return uidToString(uid, messages)
}
