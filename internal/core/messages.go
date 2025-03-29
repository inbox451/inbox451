package core

import (
	"inbox451/internal/models"
)

type MessageService struct {
	core *Core
}

func NewMessageService(core *Core) MessageService {
	return MessageService{core: core}
}

func (s *MessageService) Store(message models.Message) (models.Message, error) {
	s.core.Logger.Info("Storing new message for inbox %d from %s", message.InboxID, message.Sender)

	message, err := s.core.Repository.CreateMessage(message)
	if err != nil {
		s.core.Logger.Error("Failed to store message: %v", err)
		return message, err
	}

	s.core.Logger.Info("Successfully stored message with ID: %d", message.ID)
	return message, nil
}

func (s *MessageService) Get(id int) (models.Message, error) {
	s.core.Logger.Debug("Fetching message with ID: %d", id)

	message, err := s.core.Repository.GetMessage(id)
	if err != nil {
		s.core.Logger.Error("Failed to fetch message: %v", err)
		return message, err
	}

	return message, nil
}

func (s *MessageService) ListByInbox(inboxID int, limit, offset int, isRead *bool) (models.PaginatedResponse, error) {
	s.core.Logger.Info("Listing messages for inbox %d with limit: %d, offset: %d, isRead: %v",
		inboxID, limit, offset, isRead)

	messages, total, err := s.core.Repository.ListMessagesByInboxWithFilter(inboxID, isRead, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list messages: %v", err)
		return models.PaginatedResponse{}, err
	}

	response := models.PaginatedResponse{
		Data: messages,
	}
	response.Pagination.Total = total
	response.Pagination.Limit = limit
	response.Pagination.Offset = offset

	s.core.Logger.Info("Successfully retrieved %d messages (total: %d)", len(messages), total)
	return response, nil
}

func (s *MessageService) MarkAsRead(messageID int) (models.Message, error) {
	s.core.Logger.Debug("Marking message %d as read", messageID)

	message, err := s.core.Repository.UpdateMessageReadStatus(messageID, true)
	if err != nil {
		s.core.Logger.Error("Failed to mark message as read: %v", err)
		return message, err
	}

	s.core.Logger.Info("Successfully marked message %d as read", messageID)
	return message, nil
}

func (s *MessageService) MarkAsUnread(messageID int) (models.Message, error) {
	s.core.Logger.Debug("Marking message %d as unread", messageID)

	message, err := s.core.Repository.UpdateMessageReadStatus(messageID, false)
	if err != nil {
		s.core.Logger.Error("Failed to mark message as unread: %v", err)
		return message, err
	}

	s.core.Logger.Info("Successfully marked message %d as unread", messageID)
	return message, nil
}

func (s *MessageService) Delete(messageID int) error {
	s.core.Logger.Debug("Deleting message with ID: %d", messageID)

	if err := s.core.Repository.DeleteMessage(messageID); err != nil {
		s.core.Logger.Error("Failed to delete message: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted message with ID: %d", messageID)
	return nil
}
