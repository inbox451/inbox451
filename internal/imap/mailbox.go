package imap

import (
	"context"
	"errors"
	"time"

	"inbox451/internal/models"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
)

// ImapMailbox implements go-imap/backend.Mailbox interface
type ImapMailbox struct {
	inboxModel *models.Inbox
	user       *ImapUser
	ctx        context.Context
}

// NewImapMailbox creates a new IMAP mailbox
func NewImapMailbox(ctx context.Context, inbox *models.Inbox, user *ImapUser) backend.Mailbox {
	return &ImapMailbox{
		inboxModel: inbox,
		user:       user,
		ctx:        ctx,
	}
}

// Name returns mailbox name (inbox email)
func (m *ImapMailbox) Name() string {
	return m.inboxModel.Email
}

// Info returns mailbox info
func (m *ImapMailbox) Info() (*imap.MailboxInfo, error) {
	info := &imap.MailboxInfo{
		Attributes: []string{},
		Delimiter:  "/",
		Name:       m.inboxModel.Email,
	}
	return info, nil
}

// Status returns mailbox status
func (m *ImapMailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {
	ctx := m.ctx
	status := imap.NewMailboxStatus(m.inboxModel.Email, items)

	// Get total message count (non-deleted)
	falseVal := false
	filters := models.MessageFilters{IsDeleted: &falseVal}
	_, total, err := m.user.core.Repository.ListMessagesByInboxWithFilters(ctx, m.inboxModel.ID, filters, 1, 0)
	if err != nil {
		m.user.core.Logger.Error("Failed to get message count for inbox %s: %v", m.inboxModel.ID, err)
		return status, nil
	}
	status.Messages = uint32(total)

	// Get unread message count (non-deleted and unread)
	filters.IsRead = &falseVal
	_, unreadTotal, err := m.user.core.Repository.ListMessagesByInboxWithFilters(ctx, m.inboxModel.ID, filters, 1, 0)
	if err != nil {
		m.user.core.Logger.Error("Failed to get unread message count for inbox %s: %v", m.inboxModel.ID, err)
	} else {
		status.Unseen = uint32(unreadTotal)
	}

	// Set recent messages count (for simplicity, assume all unseen are recent)
	status.Recent = status.Unseen

	// Get next UID
	maxUID, err := m.user.core.Repository.GetMaxMessageUID(ctx, m.inboxModel.ID)
	if err != nil {
		m.user.core.Logger.Error("Failed to get max UID for inbox %s: %v", m.inboxModel.ID, err)
		status.UidNext = 1
	} else {
		status.UidNext = maxUID + 1
	}

	// Set UID validity (use inbox creation timestamp)
	if m.inboxModel.CreatedAt.Valid {
		status.UidValidity = uint32(m.inboxModel.CreatedAt.Time.Unix())
	} else {
		status.UidValidity = 1
	}

	return status, nil
}

// ListMessages returns a list of messages
func (m *ImapMailbox) ListMessages(uid bool, seqSet *imap.SeqSet, items []imap.FetchItem, ch chan<- *imap.Message) error {
	defer close(ch)
	ctx := m.ctx

	// Resolve sequence set to UIDs
	uids, err := m.resolveSeqSetToUIDs(ctx, seqSet, uid)
	if err != nil {
		m.user.core.Logger.Error("Failed to resolve sequence set: %v", err)
		return err
	}

	if len(uids) == 0 {
		return nil
	}

	// Get messages by UIDs
	dbMessages, err := m.user.core.Repository.GetMessagesByUIDs(ctx, m.inboxModel.ID, uids)
	if err != nil {
		m.user.core.Logger.Error("Failed to get messages by UIDs: %v", err)
		return err
	}

	// Map messages to IMAP format and send to channel
	for seqNum, dbMsg := range dbMessages {
		imapMsg, err := buildImapMessage(dbMsg, uint32(seqNum+1), items)
		if err != nil {
			m.user.core.Logger.Error("Failed to build IMAP message: %v", err)
			continue
		}
		ch <- imapMsg
	}

	return nil
}

// SearchMessages searches for messages matching the given criteria
func (m *ImapMailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {
	ctx := m.ctx

	// For basic implementation, handle common flag searches
	filters := models.MessageFilters{}

	// Always exclude deleted messages by default (unless specifically searching for them)
	searchingDeleted := false
	for _, flag := range criteria.WithFlags {
		if flag == imap.DeletedFlag {
			searchingDeleted = true
			break
		}
	}

	if !searchingDeleted {
		falseVal := false
		filters.IsDeleted = &falseVal
	}

	// Handle flag-based searches
	for _, flag := range criteria.WithFlags {
		switch flag {
		case imap.SeenFlag:
			trueVal := true
			filters.IsRead = &trueVal
		case imap.DeletedFlag:
			trueVal := true
			filters.IsDeleted = &trueVal
		}
	}

	for _, flag := range criteria.WithoutFlags {
		switch flag {
		case imap.SeenFlag:
			falseVal := false
			filters.IsRead = &falseVal
		case imap.DeletedFlag:
			falseVal := false
			filters.IsDeleted = &falseVal
		}
	}

	// Fetch all messages matching filters using pagination
	const batchSize = 100
	var allMessages []*models.Message
	offset := 0

	for {
		messages, totalCount, err := m.user.core.Repository.ListMessagesByInboxWithFilters(ctx, m.inboxModel.ID, filters, batchSize, offset)
		if err != nil {
			m.user.core.Logger.Error("Failed to search messages: %v", err)
			return nil, err
		}

		allMessages = append(allMessages, messages...)

		// Check if we've fetched all messages
		if len(allMessages) >= totalCount || len(messages) < batchSize {
			break
		}

		offset += batchSize
	}

	messages := allMessages

	// Handle search criteria based on header fields, body, etc.
	results := []uint32{}
	for i, msg := range messages {
		// Apply additional search criteria
		if matchesSearchCriteria(msg, criteria) {
			if uid {
				results = append(results, stringToUID(msg.ID))
			} else {
				// For sequence numbers, use 1-based indexing
				results = append(results, uint32(i+1))
			}
		}
	}

	return results, nil
}

// Check does nothing for this implementation
func (m *ImapMailbox) Check() error {
	return nil
}

// ExpungeMessages permanently deletes messages with given UIDs
func (m *ImapMailbox) ExpungeMessages(uids []uint32) error {
	ctx := m.ctx

	// Delete each message
	for _, uid := range uids {
		if err := m.user.core.MessageService.Delete(ctx, uidToStringByLookup(ctx, m, uid)); err != nil {
			m.user.core.Logger.Error("Failed to expunge message %d: %v", uid, err)
			// Continue with other messages even if one fails
		}
	}

	m.user.core.Logger.Info("Expunged %d messages from inbox %s", len(uids), m.inboxModel.ID)
	return nil
}

// CopyMessages is not supported
func (m *ImapMailbox) CopyMessages(uid bool, seqSet *imap.SeqSet, dest string) error {
	return errors.New("copy not supported")
}

// MoveMessages is not supported
func (m *ImapMailbox) MoveMessages(uid bool, seqSet *imap.SeqSet, dest string) error {
	return errors.New("move not supported")
}

// CreateMessage is not supported
func (m *ImapMailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	return errors.New("message creation not supported")
}

// UpdateMessagesFlags handles flag updates for messages
func (m *ImapMailbox) UpdateMessagesFlags(uid bool, seqSet *imap.SeqSet, operation imap.FlagsOp, flags []string) error {
	ctx := m.ctx

	// Resolve sequence set to UIDs
	uids, err := m.resolveSeqSetToUIDs(ctx, seqSet, uid)
	if err != nil {
		return err
	}

	// Track errors for logging summary
	var failedUpdates int

	// Update flags for each message
	for _, messageUID := range uids {
		// For SetFlags operation, first clear all flags
		if operation == imap.SetFlags {
			// Clear Seen flag if not in the new flags list
			seenInFlags := false
			deletedInFlags := false
			for _, flag := range flags {
				if flag == imap.SeenFlag {
					seenInFlags = true
				}
				if flag == imap.DeletedFlag {
					deletedInFlags = true
				}
			}

			// Clear flags that are not in the new set
			if !seenInFlags {
				if err := m.user.core.MessageService.MarkAsUnread(ctx, uidToStringByLookup(ctx, m, messageUID)); err != nil {
					m.user.core.Logger.Error("Failed to mark message %d as unread: %v", messageUID, err)
					failedUpdates++
				}
			}
			if !deletedInFlags {
				if err := m.user.core.MessageService.MarkAsUndeleted(ctx, uidToStringByLookup(ctx, m, messageUID)); err != nil {
					m.user.core.Logger.Error("Failed to mark message %d as undeleted: %v", messageUID, err)
					failedUpdates++
				}
			}
		}

		// Now apply the requested flags
		for _, flag := range flags {
			switch flag {
			case imap.SeenFlag:
				switch operation {
				case imap.AddFlags, imap.SetFlags:
					if err := m.user.core.MessageService.MarkAsRead(ctx, uidToStringByLookup(ctx, m, messageUID)); err != nil {
						m.user.core.Logger.Error("Failed to mark message %d as read: %v", messageUID, err)
						failedUpdates++
					}
				case imap.RemoveFlags:
					if err := m.user.core.MessageService.MarkAsUnread(ctx, uidToStringByLookup(ctx, m, messageUID)); err != nil {
						m.user.core.Logger.Error("Failed to mark message %d as unread: %v", messageUID, err)
						failedUpdates++
					}
				}
			case imap.DeletedFlag:
				switch operation {
				case imap.AddFlags, imap.SetFlags:
					if err := m.user.core.MessageService.MarkAsDeleted(ctx, uidToStringByLookup(ctx, m, messageUID)); err != nil {
						m.user.core.Logger.Error("Failed to mark message %d as deleted: %v", messageUID, err)
						failedUpdates++
					}
				case imap.RemoveFlags:
					if err := m.user.core.MessageService.MarkAsUndeleted(ctx, uidToStringByLookup(ctx, m, messageUID)); err != nil {
						m.user.core.Logger.Error("Failed to mark message %d as undeleted: %v", messageUID, err)
						failedUpdates++
					}
				}
			}
		}
	}

	// Log summary if there were any failures
	if failedUpdates > 0 {
		m.user.core.Logger.Warn("STORE operation completed with %d failed updates out of %d messages", failedUpdates, len(uids))
	}

	return nil
}

// Expunge permanently deletes messages marked as deleted
func (m *ImapMailbox) Expunge() error {
	ctx := m.ctx

	// Get all messages marked as deleted
	trueVal := true
	filters := models.MessageFilters{IsDeleted: &trueVal}
	messages, _, err := m.user.core.Repository.ListMessagesByInboxWithFilters(ctx, m.inboxModel.ID, filters, 0, 0)
	if err != nil {
		m.user.core.Logger.Error("Failed to get deleted messages for expunge: %v", err)
		return err
	}

	// Permanently delete each message
	for _, msg := range messages {
		if err := m.user.core.MessageService.Delete(ctx, msg.ID); err != nil {
			m.user.core.Logger.Error("Failed to expunge message %s: %v", msg.ID, err)
			// Continue with other messages even if one fails
		}
	}

	m.user.core.Logger.Info("Expunged %d messages from inbox %s", len(messages), m.inboxModel.ID)
	return nil
}

// SetSubscribed is not supported
func (m *ImapMailbox) SetSubscribed(subscribed bool) error {
	return errors.New("subscription changes not supported")
}

// resolveSeqSetToUIDs converts sequence set to UIDs
func (m *ImapMailbox) resolveSeqSetToUIDs(ctx context.Context, seqSet *imap.SeqSet, uid bool) ([]uint32, error) {
	if uid {
		// If already UIDs, extract them from sequence set
		var uids []uint32
		for _, seq := range seqSet.Set {
			if seq.Stop == 0 {
				// Single number
				uids = append(uids, seq.Start)
			} else {
				// Range
				for i := seq.Start; i <= seq.Stop; i++ {
					uids = append(uids, i)
				}
			}
		}
		return uids, nil
	}

	// If sequence numbers, we need to get all UIDs including deleted messages for proper sequence mapping
	allUIDs, err := m.user.core.Repository.GetAllMessageUIDsForInboxIncludingDeleted(ctx, m.inboxModel.ID)
	if err != nil {
		return nil, err
	}

	var uids []uint32
	for _, seq := range seqSet.Set {
		if seq.Stop == 0 {
			// Single sequence number
			if int(seq.Start-1) < len(allUIDs) {
				uids = append(uids, allUIDs[seq.Start-1])
			}
		} else {
			// Range of sequence numbers
			for i := seq.Start; i <= seq.Stop && int(i-1) < len(allUIDs); i++ {
				uids = append(uids, allUIDs[i-1])
			}
		}
	}

	return uids, nil
}
