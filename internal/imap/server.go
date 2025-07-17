package imap

import (
	"context"
	"fmt"

	"inbox451/internal/core"
	"inbox451/internal/util"

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
func NewServer(core *core.Core) (*ImapServer, error) {
	core.Logger.Info("IMAP Server initializing")

	// Create the backend
	backend := NewBackend(core)
	s := server.New(backend)

	// Configure server address
	if core.Config.Server.IMAP.Address != "" {
		s.Addr = core.Config.Server.IMAP.Address
	} else if core.Config.Server.IMAP.Port != "" {
		port := core.Config.Server.IMAP.Port
		if port[0] == ':' {
			s.Addr = port
		} else {
			s.Addr = ":" + port
		}
	} else {
		s.Addr = ":1143"
	}

	core.Logger.Info("IMAP Server will listen on: %s", s.Addr)

	// Enable debug logging
	s.Debug = core.Logger.Writer()

	// Configure TLS if enabled
	if core.Config.Server.IMAP.EnableTLS {
		config, err := util.GetTLSConfig(core, core.Config.Server.TLS.Cert, core.Config.Server.TLS.Key)
		if err != nil {
			return nil, fmt.Errorf("IMAP: Failed to load TLS configuration: %w", err)
		}
		s.TLSConfig = config
	}

	// Allow unencrypted plain text authentication based on config
	s.AllowInsecureAuth = core.Config.Server.IMAP.AllowInsecureAuth

	return &ImapServer{
		core: core,
		imap: s,
	}, nil
}
