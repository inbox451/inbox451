package imap

import (
	"context"

	"inbox451/internal/core"

	"github.com/emersion/go-imap/server"
)

// Server represents the IMAP server
type ImapServer struct {
	core *core.Core
	imap *server.Server
}

// ListenAndServe starts the IMAP server
func (s *ImapServer) ListenAndServe() error {
	return s.imap.ListenAndServe()
}

// Shutdown gracefully shuts down the IMAP server
func (s *ImapServer) Shutdown(ctx context.Context) error {
	return s.imap.Close()
}

// NewServer creates and configures a new IMAP server
func NewServer(core *core.Core) *ImapServer {
	core.Logger.Info("IMAP Server initializing")

	// Create the backend
	backend := NewBackend(core)
	s := server.New(backend)

	// Configure server address
	if core.Config.Server.IMAP.Address != "" {
		s.Addr = core.Config.Server.IMAP.Address
	} else if core.Config.Server.IMAP.Port != "" {
		s.Addr = core.Config.Server.IMAP.Port
	} else {
		s.Addr = ":1143"
	}

	core.Logger.Info("IMAP Server will listen on: %s", s.Addr)

	// Enable debug logging
	s.Debug = core.Logger.Writer()

	// Allow unencrypted plain text authentication for development
	// TODO: Remove this in production
	s.AllowInsecureAuth = true

	return &ImapServer{
		core: core,
		imap: s,
	}
}
