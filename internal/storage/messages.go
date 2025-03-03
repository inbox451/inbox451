package storage

import (
	"inbox451/internal/models"
)

func (r *repository) CreateMessage(message models.Message) (models.Message, error) {
	var messageId int
	err := r.queries.CreateMessage.QueryRow(
		message.InboxID,
		message.Sender,
		message.Receiver,
		message.Subject,
		message.Body).Scan(&messageId)

	if err != nil {
		return models.Message{}, handleDBError(err)
	}

	return r.GetMessage(messageId)
}

func (r *repository) GetMessage(messageId int) (models.Message, error) {
	var message models.Message
	err := r.queries.GetMessage.Get(&message, messageId)
	if err != nil {
		return models.Message{}, handleDBError(err)
	}
	return message, nil
}

func (r *repository) ListMessagesByInbox(inboxId, limit, offset int) ([]models.Message, int, error) {
	var total int
	var messages []models.Message

	err := r.queries.CountMessagesByInbox.Get(&total, inboxId)
	if err != nil {
		return nil, 0, handleDBError(err)
	}

	err = r.queries.ListMessagesByInbox.Select(&messages, inboxId, limit, offset)
	if err != nil {
		return nil, 0, handleDBError(err)
	}

	return messages, total, nil
}

func (r *repository) UpdateMessageReadStatus(messageId int, isRead bool) (models.Message, error) {
	result, err := r.queries.UpdateMessageReadStatus.Exec(isRead, messageId)

	if err != nil {
		return models.Message{}, handleDBError(err)
	}

	if err := handleRowsAffected(result); err != nil {
		return models.Message{}, err
	}

	return r.GetMessage(messageId)
}

func (r *repository) DeleteMessage(messageId int) error {
	result, err := r.queries.DeleteMessage.Exec(messageId)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

func (r *repository) ListMessagesByInboxWithFilter(inboxId int, isRead *bool, limit, offset int) ([]models.Message, int, error) {
	var total int
	var err error
	var messages []models.Message

	if isRead == nil {
		err = r.queries.CountMessagesByInbox.Get(&total, &inboxId)
	} else {
		err = r.queries.CountMessagesByInboxWithReadFilter.Get(&total, inboxId, *isRead)
	}
	if err != nil {
		return nil, 0, handleDBError(err)
	}

	if isRead == nil {
		err = r.queries.ListMessagesByInbox.Select(&messages, inboxId, limit, offset)
	} else {
		err = r.queries.ListMessagesByInboxWithReadFilter.Select(&messages, inboxId, *isRead, limit, offset)
	}

	if err != nil {
		return nil, 0, handleDBError(err)
	}

	return messages, total, nil
}
