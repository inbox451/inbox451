package core

import (
	"context"

	"inbox451/internal/models"
)

type MessageService struct {
	core *Core
}

func NewMessageService(core *Core) MessageService {
	return MessageService{core: core}
}

func (s *MessageService) Store(ctx context.Context, message *models.Message) error {
	s.core.Logger.Info("Storing new message for inbox %d from %s", message.InboxID, message.Sender)

	if err := s.core.Repository.CreateMessage(ctx, message); err != nil {
		s.core.Logger.Error("Failed to store message: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully stored message with ID: %d", message.ID)
	return nil
}

func (s *MessageService) Get(ctx context.Context, id int) (*models.Message, error) {
	s.core.Logger.Debug("Fetching message with ID: %d", id)

	message, err := s.core.Repository.GetMessage(ctx, id)
	if err != nil {
		s.core.Logger.Error("Failed to fetch message: %v", err)
		return nil, err
	}

	if message == nil {
		s.core.Logger.Info("Message not found with ID: %d", id)
		return nil, ErrNotFound
	}

	return message, nil
}

func (s *MessageService) ListByInbox(ctx context.Context, inboxID int, limit, offset int, filters models.MessageFilters) (*models.PaginatedResponse, error) {
	s.core.Logger.Info("Listing messages for inbox %d with limit: %d, offset: %d, filters: %+v",
		inboxID, limit, offset, filters)

	messages, total, err := s.core.Repository.ListMessagesByInboxWithFilters(ctx, inboxID, filters, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list messages: %v", err)
		return nil, err
	}

	response := &models.PaginatedResponse{
		Data: messages,
	}
	response.Pagination.Total = total
	response.Pagination.Limit = limit
	response.Pagination.Offset = offset

	s.core.Logger.Info("Successfully retrieved %d messages (total: %d)", len(messages), total)
	return response, nil
}

func (s *MessageService) MarkAsRead(ctx context.Context, messageID int) error {
	s.core.Logger.Debug("Marking message %d as read", messageID)

	if err := s.core.Repository.UpdateMessageReadStatus(ctx, messageID, true); err != nil {
		s.core.Logger.Error("Failed to mark message as read: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully marked message %d as read", messageID)
	return nil
}

func (s *MessageService) MarkAsUnread(ctx context.Context, messageID int) error {
	s.core.Logger.Debug("Marking message %d as unread", messageID)

	if err := s.core.Repository.UpdateMessageReadStatus(ctx, messageID, false); err != nil {
		s.core.Logger.Error("Failed to mark message as unread: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully marked message %d as unread", messageID)
	return nil
}

func (s *MessageService) MarkAsDeleted(ctx context.Context, messageID int) error {
	s.core.Logger.Debug("Marking message %d as deleted", messageID)

	if err := s.core.Repository.UpdateMessageDeletedStatus(ctx, messageID, true); err != nil {
		s.core.Logger.Error("Failed to mark message as deleted: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully marked message %d as deleted", messageID)
	return nil
}

func (s *MessageService) MarkAsUndeleted(ctx context.Context, messageID int) error {
	s.core.Logger.Debug("Marking message %d as undeleted", messageID)

	if err := s.core.Repository.UpdateMessageDeletedStatus(ctx, messageID, false); err != nil {
		s.core.Logger.Error("Failed to mark message as undeleted: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully marked message %d as undeleted", messageID)
	return nil
}

func (s *MessageService) Delete(ctx context.Context, messageID int) error {
	s.core.Logger.Debug("Deleting message with ID: %d", messageID)

	if err := s.core.Repository.DeleteMessage(ctx, messageID); err != nil {
		s.core.Logger.Error("Failed to delete message: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted message with ID: %d", messageID)
	return nil
}
