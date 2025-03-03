package storage

import (
	"database/sql"
	"errors"

	"inbox451/internal/models"
)

func (r *repository) CreateInbox(inbox models.Inbox) (models.Inbox, error) {
	var inboxID int
	err := r.queries.CreateInbox.QueryRow(inbox.ProjectID, inbox.Email).
		Scan(&inbox.ID)
	if err != nil {
		return models.Inbox{}, handleDBError(err)
	}
	return r.GetInbox(inboxID)
}

func (r *repository) GetInbox(id int) (models.Inbox, error) {
	var inbox models.Inbox
	err := r.queries.GetInbox.Get(&inbox, id)
	if err != nil {
		return models.Inbox{}, handleDBError(err)
	}
	return inbox, nil
}

func (r *repository) GetInboxByEmail(email string) (models.Inbox, error) {
	var inbox models.Inbox
	err := r.queries.GetInboxByEmail.Get(&inbox, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Inbox{}, nil
		}
		return models.Inbox{}, err
	}
	return inbox, nil
}

func (r *repository) UpdateInbox(inbox models.Inbox) (models.Inbox, error) {
	result, err := r.queries.UpdateInbox.Exec(inbox.Email, inbox.ID)
	if err != nil {
		return models.Inbox{}, handleDBError(err)
	}
	if err := handleRowsAffected(result); err != nil {
		return models.Inbox{}, err
	}
	return r.GetInbox(inbox.ID)
}

func (r *repository) DeleteInbox(id int) error {
	result, err := r.queries.DeleteInbox.Exec(id)
	if err != nil {
		return handleDBError(err)
	}
	return handleRowsAffected(result)
}

func (r *repository) ListInboxesByProject(projectID, limit, offset int) ([]models.Inbox, int, error) {
	var total int
	var inboxes []models.Inbox

	err := r.queries.CountInboxesByProject.Get(&total, projectID)
	if err != nil {
		return nil, 0, err
	}

	err = r.queries.ListInboxesByProject.Select(&inboxes, projectID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return inboxes, total, nil
}
