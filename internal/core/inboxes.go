package core

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"inbox451/internal/models"
)

type InboxService struct {
	core *Core
}

func NewInboxService(core *Core) InboxService {
	return InboxService{core: core}
}

func (s *InboxService) Create(ctx context.Context, inbox *models.Inbox) error {
	// Check if EmailDomain is configured
	if s.core.Config.Server.EmailDomain != "" {
		// If inbox.Email doesn't contain "@", append the domain
		if !strings.Contains(inbox.Email, "@") {
			originalEmail := inbox.Email
			inbox.Email = fmt.Sprintf("%s@%s", inbox.Email, s.core.Config.Server.EmailDomain)
			s.core.Logger.Info("Auto-appended domain to inbox email: %s -> %s", originalEmail, inbox.Email)
		}

		// Validate that the email ends with the configured domain
		expectedSuffix := "@" + s.core.Config.Server.EmailDomain
		if !strings.HasSuffix(inbox.Email, expectedSuffix) {
			return &APIError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Inbox email must end with %s", expectedSuffix),
			}
		}
	}

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
	// Check if EmailDomain is configured
	if s.core.Config.Server.EmailDomain != "" {
		// If inbox.Email doesn't contain "@", append the domain
		if !strings.Contains(inbox.Email, "@") {
			originalEmail := inbox.Email
			inbox.Email = fmt.Sprintf("%s@%s", inbox.Email, s.core.Config.Server.EmailDomain)
			s.core.Logger.Info("Auto-appended domain to inbox email: %s -> %s", originalEmail, inbox.Email)
		}

		// Validate that the email ends with the configured domain
		expectedSuffix := "@" + s.core.Config.Server.EmailDomain
		if !strings.HasSuffix(inbox.Email, expectedSuffix) {
			return &APIError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("Inbox email must end with %s", expectedSuffix),
			}
		}
	}

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
func (s *InboxService) ListByUser(ctx context.Context, userID int) ([]*models.Inbox, error) {
	s.core.Logger.Info("Listing inboxes for user %d", userID)

	inboxes, err := s.core.Repository.ListInboxesByUser(ctx, userID)
	if err != nil {
		s.core.Logger.Error("Failed to list inboxes by user: %v", err)
		return nil, err
	}

	s.core.Logger.Info("Successfully retrieved %d inboxes for user %d", len(inboxes), userID)
	return inboxes, nil
}

func (s *InboxService) GetByEmailAndUser(ctx context.Context, email string, userID int) (*models.Inbox, error) {
	s.core.Logger.Debug("Fetching inbox with email %s for user %d", email, userID)

	inbox, err := s.core.Repository.GetInboxByEmailAndUser(ctx, email, userID)
	if err != nil {
		s.core.Logger.Error("Failed to fetch inbox by email and user: %v", err)
		return nil, err
	}

	if inbox == nil {
		s.core.Logger.Info("Inbox not found with email %s for user %d", email, userID)
		return nil, ErrNotFound
	}

	s.core.Logger.Info("Successfully retrieved inbox %d with email %s for user %d", inbox.ID, email, userID)
	return inbox, nil
}

func (s *InboxService) GetByEmailWithWildcard(ctx context.Context, to string) (*models.Inbox, error) {
	s.core.Logger.Info("Fetching inbox by email with wildcard: %s", to)

	// first check exact match
	inbox, err := s.core.Repository.GetInboxByEmail(ctx, to)
	if err != nil {
		s.core.Logger.Error("Failed to fetch inbox by email: %v", err)
		return nil, err
	}

	if inbox != nil {
		s.core.Logger.Info("Found inbox by exact email match: %s", to)
		return inbox, nil
	}

	// if not found, check for wildcard match
	// kermit.the.frog@example.org we just look for the domain part and until the first '.'
	// we would be looking for kermit@example.org
	atIndex := strings.LastIndex(to, "@")
	if atIndex == -1 {
		s.core.Logger.Warn("Invalid email format, missing '@': %s", to)
		return nil, errors.New("invalid email format, missing '@'")
	}

	recipientPart := to[:atIndex] // Get the part before '@'
	domainPart := to[atIndex:]    // Get the domain part after '@'

	dotIndex := strings.Index(recipientPart, ".")
	if dotIndex == -1 {
		s.core.Logger.Warn("Email not found, exact match failed, and no wildcard to match to: %s", to)
		return nil, ErrNotFound
	}

	baseEmail := recipientPart[:dotIndex] + domainPart

	s.core.Logger.Info("Looking for inbox for wildcard match. inbox=%s wildcard=%s", baseEmail, to)

	return s.core.Repository.GetInboxByEmail(ctx, baseEmail)

}
