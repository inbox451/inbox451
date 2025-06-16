package storage

import (
	"context"
	"database/sql"
	"errors"

	"inbox451/internal/models"
)

func (r *repository) CreateInbox(ctx context.Context, inbox *models.Inbox) error {
	return r.queries.CreateInbox.QueryRowContext(ctx, inbox.ProjectID, inbox.Email).
		Scan(&inbox.ID, &inbox.CreatedAt, &inbox.UpdatedAt)
}

func (r *repository) GetInbox(ctx context.Context, id string) (*models.Inbox, error) {
	var inbox models.Inbox
	err := r.queries.GetInbox.GetContext(ctx, &inbox, id)
	return &inbox, handleDBError(err)
}

func (r *repository) GetInboxByEmail(ctx context.Context, email string) (*models.Inbox, error) {
	var inbox models.Inbox
	err := r.queries.GetInboxByEmail.GetContext(ctx, &inbox, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &inbox, nil
}

func (r *repository) UpdateInbox(ctx context.Context, inbox *models.Inbox) error {
	result, err := r.queries.UpdateInbox.ExecContext(ctx, inbox.Email, inbox.ID)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

func (r *repository) DeleteInbox(ctx context.Context, id string) error {
	result, err := r.queries.DeleteInbox.ExecContext(ctx, id)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

func (r *repository) ListInboxesByProject(ctx context.Context, projectID string, limit, offset int) ([]*models.Inbox, int, error) {
	var total int
	err := r.queries.CountInboxesByProject.GetContext(ctx, &total, projectID)
	if err != nil {
		return nil, 0, err
	}

	inboxes := []*models.Inbox{}
	if total > 0 {
		err = r.queries.ListInboxesByProject.SelectContext(ctx, &inboxes, projectID, limit, offset)
		if err != nil {
			return nil, 0, err
		}
	}

	return inboxes, total, nil
}

// ListInboxesByUser returns all inboxes accessible to a user through project membership
func (r *repository) ListInboxesByUser(ctx context.Context, userID int) ([]*models.Inbox, error) {
	inboxes := []*models.Inbox{}
	err := r.queries.ListInboxesByUser.SelectContext(ctx, &inboxes, userID)
	if err != nil {
		return nil, handleDBError(err)
	}

	return inboxes, nil
}

// GetInboxByEmailAndUser returns an inbox by email if the user has access to it
func (r *repository) GetInboxByEmailAndUser(ctx context.Context, email string, userID int) (*models.Inbox, error) {
	var inbox models.Inbox
	err := r.queries.GetInboxByEmailAndUser.GetContext(ctx, &inbox, email, userID)
	if err != nil {
		return nil, handleDBError(err)
	}

	return &inbox, nil
}
