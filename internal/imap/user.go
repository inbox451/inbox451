package imap

import (
	"context"
	"errors"

	"inbox451/internal/core"
	"inbox451/internal/models"

	"github.com/emersion/go-imap/backend"
)

// ImapUser implements go-imap/backend.User interface
type ImapUser struct {
	userModel *models.User
	core      *core.Core
	ctx       context.Context
}

// NewImapUser creates a new IMAP user
func NewImapUser(ctx context.Context, user *models.User, core *core.Core) backend.User {
	return &ImapUser{
		userModel: user,
		core:      core,
		ctx:       ctx,
	}
}

// Username returns user's username
func (u *ImapUser) Username() string {
	if u.userModel.Username != "" {
		return u.userModel.Username
	}
	return u.userModel.Email
}

// ListMailboxes returns a list of mailboxes for the user
func (u *ImapUser) ListMailboxes(subscribed bool) ([]backend.Mailbox, error) {
	ctx := u.ctx
	inboxes, err := u.core.InboxService.ListByUser(ctx, u.userModel.ID)
	if err != nil {
		u.core.Logger.Error("Failed to list inboxes for user %d: %v", u.userModel.ID, err)
		return nil, err
	}

	mailboxes := make([]backend.Mailbox, len(inboxes))
	for i, inbox := range inboxes {
		mailboxes[i] = NewImapMailbox(ctx, inbox, u)
	}

	u.core.Logger.Info("Listed %d mailboxes for user %d", len(mailboxes), u.userModel.ID)
	return mailboxes, nil
}

// GetMailbox returns a specific mailbox
func (u *ImapUser) GetMailbox(name string) (backend.Mailbox, error) {
	ctx := u.ctx

	// Handle special case for "INBOX" - map to user's first inbox
	if name == "INBOX" {
		inboxes, err := u.core.InboxService.ListByUser(ctx, u.userModel.ID)
		if err != nil {
			u.core.Logger.Error("IMAP GetMailbox: Error fetching inboxes for INBOX special name for user %d: %v", u.userModel.ID, err)
			return nil, err
		}
		if len(inboxes) == 0 {
			u.core.Logger.Warn("IMAP GetMailbox: No inboxes available for user %d, cannot select INBOX.", u.userModel.ID)
			return nil, backend.ErrNoSuchMailbox
		}
		u.core.Logger.Debug("IMAP GetMailbox: Mapping 'INBOX' to user %d's first inbox: %s", u.userModel.ID, inboxes[0].Email)
		return NewImapMailbox(ctx, inboxes[0], u), nil
	}

	// Try to get inbox by email address
	inbox, err := u.core.InboxService.GetByEmailAndUser(ctx, name, u.userModel.ID)
	if err != nil {
		u.core.Logger.Error("Failed to get mailbox %s for user %d: %v", name, u.userModel.ID, err)
		return nil, backend.ErrNoSuchMailbox
	}

	return NewImapMailbox(ctx, inbox, u), nil
}

// CreateMailbox is not supported
func (u *ImapUser) CreateMailbox(name string) error {
	return errors.New("mailbox creation not supported")
}

// DeleteMailbox is not supported
func (u *ImapUser) DeleteMailbox(name string) error {
	return errors.New("mailbox deletion not supported")
}

// RenameMailbox is not supported
func (u *ImapUser) RenameMailbox(existingName, newName string) error {
	return errors.New("mailbox renaming not supported")
}

// Logout handles user logout
func (u *ImapUser) Logout() error {
	u.core.Logger.Info("User %s logged out from IMAP", u.Username())
	return nil
}
