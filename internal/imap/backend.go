package imap

import (
	"context"

	"inbox451/internal/core"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
)

// ImapBackend implements go-imap/backend interface
type ImapBackend struct {
	core *core.Core
}

// NewBackend creates a new IMAP backend
func NewBackend(core *core.Core) backend.Backend {
	return &ImapBackend{core: core}
}

// Login handles user authentication
func (be *ImapBackend) Login(connInfo *imap.ConnInfo, username string, password string) (backend.User, error) {
	be.core.Logger.Info("IMAP Token Login attempt for username: %s (using token auth)", username)

	ctx := context.Background()
	user, err := be.core.UserService.LoginWithToken(ctx, username, password)
	if err != nil {
		be.core.Logger.Warn("IMAP Token Login failed for username: %s, error: %v", username, err)
		return nil, backend.ErrInvalidCredentials
	}

	be.core.Logger.Info("IMAP Token Login successful for username: %s", username)
	return NewImapUser(ctx, user, be.core), nil
}
