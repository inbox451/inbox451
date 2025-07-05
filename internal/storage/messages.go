package storage

import (
	"context"
	"hash/fnv"

	"inbox451/internal/models"
)

// stringToUID converts a string ID (UUID) to a uint32 UID for IMAP
// Uses FNV-1a hash to ensure consistent mapping
func stringToUID(id string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(id))
	return h.Sum32()
}

func (r *repository) CreateMessage(ctx context.Context, message *models.Message) error {
	err := r.queries.CreateMessage.QueryRowContext(ctx,
		message.InboxID, message.Sender, message.Receiver, message.Subject, message.Body).
		Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)
	return handleDBError(err)
}

func (r *repository) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	var message models.Message
	err := r.queries.GetMessage.GetContext(ctx, &message, id)
	if err != nil {
		return nil, handleDBError(err)
	}
	return &message, nil
}

func (r *repository) ListMessagesByInbox(ctx context.Context, inboxID string, limit, offset int) ([]*models.Message, int, error) {
	var total int
	err := r.queries.CountMessagesByInbox.GetContext(ctx, &total, inboxID)
	if err != nil {
		return nil, 0, handleDBError(err)
	}

	messages := []*models.Message{}

	if total > 0 {
		err = r.queries.ListMessagesByInbox.SelectContext(ctx, &messages, inboxID, limit, offset)
		if err != nil {
			return nil, 0, handleDBError(err)
		}
	}

	return messages, total, nil
}

func (r *repository) UpdateMessageReadStatus(ctx context.Context, messageID string, isRead bool) error {
	result, err := r.queries.UpdateMessageReadStatus.ExecContext(ctx, isRead, messageID)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

func (r *repository) DeleteMessage(ctx context.Context, messageID string) error {
	result, err := r.queries.DeleteMessage.ExecContext(ctx, messageID)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

func (r *repository) ListMessagesByInboxWithFilter(ctx context.Context, inboxID string, isRead *bool, limit, offset int) ([]*models.Message, int, error) {
	var total int
	var err error

	if isRead == nil {
		err = r.queries.CountMessagesByInbox.GetContext(ctx, &total, inboxID)
	} else {
		err = r.queries.CountMessagesByInboxWithReadFilter.GetContext(ctx, &total, inboxID, *isRead)
	}
	if err != nil {
		return nil, 0, handleDBError(err)
	}

	messages := []*models.Message{}

	if total > 0 {
		if isRead == nil {
			err = r.queries.ListMessagesByInbox.SelectContext(ctx, &messages, inboxID, limit, offset)
		} else {
			err = r.queries.ListMessagesByInboxWithReadFilter.SelectContext(ctx, &messages, inboxID, *isRead, limit, offset)
		}
		if err != nil {
			return nil, 0, handleDBError(err)
		}
	}

	return messages, total, nil
}

// UpdateMessageDeletedStatus updates the is_deleted flag for a message
func (r *repository) UpdateMessageDeletedStatus(ctx context.Context, messageID string, isDeleted bool) error {
	result, err := r.queries.UpdateMessageDeletedStatus.ExecContext(ctx, isDeleted, messageID)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

// ListMessagesByInboxWithFilters returns messages with both read and deleted filters
func (r *repository) ListMessagesByInboxWithFilters(ctx context.Context, inboxID string, filters models.MessageFilters, limit, offset int) ([]*models.Message, int, error) {
	var total int
	err := r.queries.CountMessagesByInboxWithFilters.GetContext(ctx, &total, inboxID, filters.IsRead, filters.IsDeleted)
	if err != nil {
		return nil, 0, handleDBError(err)
	}

	messages := []*models.Message{}

	if total > 0 {
		err = r.queries.ListMessagesByInboxWithFilters.SelectContext(ctx, &messages, inboxID, filters.IsRead, filters.IsDeleted, limit, offset)
		if err != nil {
			return nil, 0, handleDBError(err)
		}
	}

	return messages, total, nil
}

// GetMessagesByUIDs returns messages by their IDs (UIDs in IMAP context)
// Now works with UUID strings by getting all messages and filtering by UID hash
func (r *repository) GetMessagesByUIDs(ctx context.Context, inboxID string, uids []uint32) ([]*models.Message, error) {
	if len(uids) == 0 {
		return []*models.Message{}, nil
	}

	// Get all messages for the inbox and filter by UID hash
	filters := models.MessageFilters{} // Get all messages including deleted
	allMessages, _, err := r.ListMessagesByInboxWithFilters(ctx, inboxID, filters, 0, 0)
	if err != nil {
		return nil, err
	}

	// Filter messages that match the requested UIDs
	var result []*models.Message
	uidSet := make(map[uint32]bool)
	for _, uid := range uids {
		uidSet[uid] = true
	}

	for _, msg := range allMessages {
		msgUID := stringToUID(msg.ID)
		if uidSet[msgUID] {
			result = append(result, msg)
		}
	}

	return result, nil
}

// GetAllMessageUIDsForInbox returns all message IDs for an inbox (excluding deleted)
func (r *repository) GetAllMessageUIDsForInbox(ctx context.Context, inboxID string) ([]uint32, error) {
	// Get all non-deleted messages
	falseVal := false
	filters := models.MessageFilters{IsDeleted: &falseVal}
	messages, _, err := r.ListMessagesByInboxWithFilters(ctx, inboxID, filters, 0, 0)
	if err != nil {
		return nil, err
	}

	// Convert string IDs to UIDs using hash
	uids := make([]uint32, len(messages))
	for i, msg := range messages {
		uids[i] = stringToUID(msg.ID)
	}

	return uids, nil
}

// GetAllMessageUIDsForInboxIncludingDeleted returns all message IDs for an inbox (including deleted)
// This is used for IMAP sequence number mapping where deleted messages are still addressable until expunged
func (r *repository) GetAllMessageUIDsForInboxIncludingDeleted(ctx context.Context, inboxID string) ([]uint32, error) {
	// Get all messages including deleted
	filters := models.MessageFilters{}
	messages, _, err := r.ListMessagesByInboxWithFilters(ctx, inboxID, filters, 0, 0)
	if err != nil {
		return nil, err
	}

	// Convert string IDs to UIDs using hash
	uids := make([]uint32, len(messages))
	for i, msg := range messages {
		uids[i] = stringToUID(msg.ID)
	}

	return uids, nil
}

// GetMaxMessageUID returns the highest message UID in an inbox
// Since UIDs are now hash-based, we need to get all messages and find the max UID
func (r *repository) GetMaxMessageUID(ctx context.Context, inboxID string) (uint32, error) {
	// Get all messages
	filters := models.MessageFilters{}
	messages, _, err := r.ListMessagesByInboxWithFilters(ctx, inboxID, filters, 0, 0)
	if err != nil {
		return 0, err
	}

	if len(messages) == 0 {
		return 0, nil
	}

	// Find the maximum UID
	var maxUID uint32
	for _, msg := range messages {
		uid := stringToUID(msg.ID)
		if uid > maxUID {
			maxUID = uid
		}
	}

	return maxUID, nil
}
