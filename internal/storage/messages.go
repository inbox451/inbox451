package storage

import (
	"context"

	"inbox451/internal/models"
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
