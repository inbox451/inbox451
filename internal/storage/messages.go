package storage

import (
	"context"

	"inbox451/internal/models"

	"github.com/lib/pq"
)

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
func (r *repository) UpdateMessageDeletedStatus(ctx context.Context, messageID int, isDeleted bool) error {
	result, err := r.queries.UpdateMessageDeletedStatus.ExecContext(ctx, isDeleted, messageID)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

// ListMessagesByInboxWithFilters returns messages with both read and deleted filters
func (r *repository) ListMessagesByInboxWithFilters(ctx context.Context, inboxID int, filters models.MessageFilters, limit, offset int) ([]*models.Message, int, error) {
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
func (r *repository) GetMessagesByUIDs(ctx context.Context, inboxID int, uids []uint32) ([]*models.Message, error) {
	if len(uids) == 0 {
		return []*models.Message{}, nil
	}

	// Convert uint32 slice to int slice for the query
	ids := make([]int, len(uids))
	for i, uid := range uids {
		ids[i] = int(uid)
	}

	messages := []*models.Message{}
	err := r.queries.GetMessagesByUIDs.SelectContext(ctx, &messages, inboxID, pq.Array(ids))
	if err != nil {
		return nil, handleDBError(err)
	}

	return messages, nil
}

// GetAllMessageUIDsForInbox returns all message IDs for an inbox (excluding deleted)
func (r *repository) GetAllMessageUIDsForInbox(ctx context.Context, inboxID int) ([]uint32, error) {
	var ids []int
	err := r.queries.GetAllMessageUIDsForInbox.SelectContext(ctx, &ids, inboxID)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Convert int slice to uint32 slice
	uids := make([]uint32, len(ids))
	for i, id := range ids {
		uids[i] = uint32(id)
	}

	return uids, nil
}

// GetAllMessageUIDsForInboxIncludingDeleted returns all message IDs for an inbox (including deleted)
// This is used for IMAP sequence number mapping where deleted messages are still addressable until expunged
func (r *repository) GetAllMessageUIDsForInboxIncludingDeleted(ctx context.Context, inboxID int) ([]uint32, error) {
	var ids []int
	err := r.queries.GetAllMessageUIDsForInboxIncludingDeleted.SelectContext(ctx, &ids, inboxID)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Convert int slice to uint32 slice
	uids := make([]uint32, len(ids))
	for i, id := range ids {
		uids[i] = uint32(id)
	}

	return uids, nil
}

// GetMaxMessageUID returns the highest message ID in an inbox
func (r *repository) GetMaxMessageUID(ctx context.Context, inboxID int) (uint32, error) {
	var maxUID int
	err := r.queries.GetMaxMessageUID.GetContext(ctx, &maxUID, inboxID)
	if err != nil {
		return 0, handleDBError(err)
	}

	return uint32(maxUID), nil
}
