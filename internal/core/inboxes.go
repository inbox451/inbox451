package core

import (
	"inbox451/internal/models"
)

type InboxService struct {
	core *Core
}

func NewInboxService(core *Core) InboxService {
	return InboxService{core: core}
}

func (s *InboxService) Create(inbox models.Inbox) (models.Inbox, error) {
	s.core.Logger.Info("Creating new inbox for project %d: %s", inbox.ProjectID, inbox.Email)

	newInbox, err := s.core.Repository.CreateInbox(inbox)
	if err != nil {
		s.core.Logger.Error("Failed to create inbox: %v", err)
		return models.Inbox{}, err
	}

	s.core.Logger.Info("Successfully created inbox with ID: %d", newInbox.ID)
	return newInbox, nil
}

func (s *InboxService) Get(id int) (models.Inbox, error) {
	s.core.Logger.Debug("Fetching inbox with ID: %d", id)

	inbox, err := s.core.Repository.GetInbox(id)
	if err != nil {
		s.core.Logger.Error("Failed to fetch inbox: %v", err)
		return inbox, err
	}

	return inbox, nil
}

func (s *InboxService) Update(inbox models.Inbox) (models.Inbox, error) {
	s.core.Logger.Info("Updating inbox with ID: %d", inbox.ID)

	update, err := s.core.Repository.UpdateInbox(inbox)
	if err != nil {
		s.core.Logger.Error("Failed to update inbox: %v", err)
		return update, err
	}

	s.core.Logger.Info("Successfully updated inbox with ID: %d", inbox.ID)
	return update, nil
}

func (s *InboxService) Delete(id int) error {
	s.core.Logger.Info("Deleting inbox with ID: %d", id)

	if err := s.core.Repository.DeleteInbox(id); err != nil {
		s.core.Logger.Error("Failed to delete inbox: %v", err)
		return err
	}

	s.core.Logger.Info("Successfully deleted inbox with ID: %d", id)
	return nil
}

func (s *InboxService) ListByProject(projectID, limit, offset int) (models.PaginatedResponse, error) {
	s.core.Logger.Info("Listing inboxes for project %d with limit: %d and offset: %d", projectID, limit, offset)

	inboxes, total, err := s.core.Repository.ListInboxesByProject(projectID, limit, offset)
	if err != nil {
		s.core.Logger.Error("Failed to list inboxes: %v", err)
		return models.PaginatedResponse{}, err
	}

	response := models.PaginatedResponse{
		Data: inboxes,
	}
	response.Pagination.Total = total
	response.Pagination.Limit = limit
	response.Pagination.Offset = offset

	s.core.Logger.Info("Successfully retrieved %d inboxes (total: %d)", len(inboxes), total)
	return response, nil
}
