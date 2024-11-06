package core

import (
	"context"
	"mercury/internal/models"
)

type InboxService struct {
	core *Core
}

func NewInboxService(core *Core) InboxService {
	return InboxService{core: core}
}

func (s *InboxService) Create(ctx context.Context, inbox *models.Inbox) error {
	s.core.Logger.Info("Creating new inbox for project %d: %s", inbox.ProjectID, inbox.Email)

	if err := s.core.Repository.CreateInbox(ctx, inbox); err != nil {
		s.core.Logger.Error("Failed to create inbox: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully created inbox with ID: %d", inbox.ID)
	return nil
}

func (s *InboxService) Get(ctx context.Context, id int) (*models.Inbox, error) {
	s.core.Logger.Debug("Fetching inbox with ID: %d", id)

	inbox, err := s.core.Repository.GetInbox(ctx, id)
	if err != nil {
		s.core.Logger.Error("Failed to fetch inbox: %v", err)
		return nil, err
	}

	if inbox == nil {
		s.core.Logger.Info("Inbox not found with ID: %d", id)
		return nil, ErrNotFound
	}

	return inbox, nil
}

func (s *InboxService) Update(ctx context.Context, inbox *models.Inbox) error {
	s.core.Logger.Info("Updating inbox with ID: %d", inbox.ID)

	if err := s.core.Repository.UpdateInbox(ctx, inbox); err != nil {
		s.core.Logger.Error("Failed to update inbox: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully updated inbox with ID: %d", inbox.ID)
	return nil
}

func (s *InboxService) Delete(ctx context.Context, id int) error {
	s.core.Logger.Info("Deleting inbox with ID: %d", id)

	if err := s.core.Repository.DeleteInbox(ctx, id); err != nil {
		s.core.Logger.Error("Failed to delete inbox: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted inbox with ID: %d", id)
	return nil
}

func (s *InboxService) ListByProject(ctx context.Context, projectID, limit, offset int) (*models.PaginatedResponse, error) {
	s.core.Logger.Info("Listing inboxes for project %d with limit: %d and offset: %d", projectID, limit, offset)

	inboxes, total, err := s.core.Repository.ListInboxesByProject(ctx, projectID, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list inboxes: %v", err)
		return nil, err
	}

	response := &models.PaginatedResponse{
		Data: inboxes,
	}
	response.Pagination.Total = total
	response.Pagination.Limit = limit
	response.Pagination.Offset = offset

	s.core.Logger.Info("Successfully retrieved %d inboxes (total: %d)", len(inboxes), total)
	return response, nil
}
