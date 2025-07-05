package storage

import (
	"context"

	"inbox451/internal/models"

	"github.com/lib/pq"
)


func (r *repository) CreateMessage(ctx context.Context, message *models.Message) error {
	err := r.queries.CreateMessage.QueryRowContext(ctx,
		message.InboxID, message.Sender, message.Receiver, message.Subject, message.Body).
		Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt, &message.UID)
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
func (r *repository) GetMessagesByUIDs(ctx context.Context, inboxID string, uids []uint32) ([]*models.Message, error) {
	if len(uids) == 0 {
		return []*models.Message{}, nil
	}

	var messages []*models.Message
	err := r.queries.GetMessagesByUIDs.SelectContext(ctx, &messages, inboxID, pq.Array(uids))
	if err != nil {
		return nil, handleDBError(err)
	}

	return messages, nil
}

// GetAllMessageUIDsForInbox returns all message IDs for an inbox (excluding deleted)
func (r *repository) GetAllMessageUIDsForInbox(ctx context.Context, inboxID string) ([]uint32, error) {
	var uids []uint32
	err := r.queries.GetAllMessageUIDsForInbox.SelectContext(ctx, &uids, inboxID)
	if err != nil {
		return nil, handleDBError(err)
	}

	return uids, nil
}

// GetAllMessageUIDsForInboxIncludingDeleted returns all message IDs for an inbox (including deleted)
// This is used for IMAP sequence number mapping where deleted messages are still addressable until expunged
func (r *repository) GetAllMessageUIDsForInboxIncludingDeleted(ctx context.Context, inboxID string) ([]uint32, error) {
	var uids []uint32
	err := r.queries.GetAllMessageUIDsForInboxIncludingDeleted.SelectContext(ctx, &uids, inboxID)
	if err != nil {
		return nil, handleDBError(err)
	}

	return uids, nil
}

// GetMaxMessageUID returns the highest message UID in an inbox
func (r *repository) GetMaxMessageUID(ctx context.Context, inboxID string) (uint32, error) {
	var maxUID uint32
	err := r.queries.GetMaxMessageUID.GetContext(ctx, &maxUID, inboxID)
	return maxUID, handleDBError(err)
}

func (r *repository) GetMessageIDFromUID(ctx context.Context, inboxID string, uid uint32) (string, error) {
	var messageID string
	err := r.queries.GetMessageIDFromUID.GetContext(ctx, &messageID, inboxID, uid)
	return messageID, handleDBError(err)
}
