// This file implements a basic SMTP Mail Transfer Agent (MTA) using the emersion/go-smtp library.
//
// Purpose:
// - Accepts incoming SMTP connections from remote MTAs over port 25.
// - Handles server-to-server email delivery for inbound messages.
// - Does not support SMTP AUTH (authentication) as this is not required for standard MTA behavior.
// - Validates recipients, accepts messages for local domains, and rejects relaying for unauthorized domains.

package mta

import (
	"bytes"
	"context"
	"github.com/emersion/go-message"
	"github.com/emersion/go-smtp"
	"inbox451/internal/core"
	"inbox451/internal/models"
	"io"
	"strings"
	"time"
)

type MTAServer struct {
	core *core.Core
	smtp *smtp.Server
}

type MTABackend struct {
	core *core.Core
}

type MTASession struct {
	core *core.Core
	from string
	to   string
}

func NewServer(core *core.Core) *MTAServer {
	backend := &MTABackend{core: core}
	smtpServer := smtp.NewServer(backend)

	// TODO: Review configuration options and place them in core.Config
	smtpServer.Addr = core.Config.Server.SMTP.Hostname + ":" + core.Config.Server.SMTP.MTA.Port
	smtpServer.Domain = core.Config.Server.SMTP.Domain
	smtpServer.ReadTimeout = 5 * time.Second
	smtpServer.WriteTimeout = 10 * time.Second
	smtpServer.MaxMessageBytes = 10 * 1024 * 1024
	smtpServer.MaxRecipients = 100
	smtpServer.AllowInsecureAuth = core.Config.Server.SMTP.AllowInsecureAuth

	return &MTAServer{
		core: core,
		smtp: smtpServer,
	}
}

func (s *MTAServer) ListenAndServe() error {
	s.core.Logger.Info("MTA: Starting SMTP server on %s", s.smtp.Addr)
	if err := s.smtp.ListenAndServe(); err != nil {
		s.core.Logger.Error("MTA: Error starting SMTP server: %v", err)
		return err
	}
	return nil
}

func (s *MTAServer) Shutdown(ctx context.Context) error {
	s.core.Logger.Info("MTA: Shutting down SMTP server")
	if err := s.smtp.Shutdown(ctx); err != nil {
		s.core.Logger.Error("MTA: Error shutting down SMTP server: %v", err)
		return err
	}
	s.core.Logger.Info("MTA: SMTP server shutdown complete")
	return nil
}

func (backend *MTABackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	backend.core.Logger.Info("MTA: New connection from %s", c.Conn().RemoteAddr())
	session := &MTASession{core: backend.core}
	session.Reset()
	return session, nil
}

func (s *MTASession) Mail(from string, opts *smtp.MailOptions) error {
	s.core.Logger.Info("MTA: Mail from %s", from)
	s.from = from
	return nil
}

func (s *MTASession) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.core.Logger.Info("MTA: Recipient to %s", to)

	// Validate the domain of the recipient email address
	expectedDomain := "@" + s.core.Config.Server.EmailDomain
	if strings.HasSuffix(to, expectedDomain) == false {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 7, 1},
			Message:      "Relay not permitted for domain, message refused",
		}
	}

	// TODO: Understand how much time we can wait here and make it configurable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inbox, err := s.core.InboxService.GetByEmailWithWildcard(ctx, to)
	if err != nil {
		s.core.Logger.Error("MTA: Error fetching inbox for %s: %v", to, err)
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{4, 3, 0},
			Message:      "Temporary error while processing recipient",
		}
	}

	if inbox == nil {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "Recipient address rejected: User unknown",
		}
	}

	s.core.Logger.Info("MSA: Recipient %s accepted for (inbox ID: %d)", to, inbox.ID)

	s.to = to
	return nil
}

func (s *MTASession) Reset() {
	s.from = ""
	s.to = ""
}

func (s *MTASession) Data(r io.Reader) error {
	s.core.Logger.Info("MTA: Data received from %s to %s", s.from, s.to)

	// TODO: Understand how much time we can wait here and make it configurable
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	buffer := new(bytes.Buffer)
	if _, err := buffer.ReadFrom(r); err != nil {
		s.core.Logger.Error("MTA: Error reading data: %v", err)
		return &smtp.SMTPError{
			Code:         554,
			EnhancedCode: smtp.EnhancedCode{5, 3, 0},
			Message:      "Message content could not be read",
		}
	}

	// parse the message content, body, headers, etc.
	msg, err := message.Read(bytes.NewReader(buffer.Bytes()))
	if err != nil {
		s.core.Logger.Error("MTA: Error parsing message: %v", err)
		return &smtp.SMTPError{
			Code:         554,
			EnhancedCode: smtp.EnhancedCode{5, 3, 0},
			Message:      "Message content could not be parsed",
		}
	}

	header := msg.Header
	body := new(bytes.Buffer)
	if _, err := body.ReadFrom(msg.Body); err != nil {
		s.core.Logger.Error("MTA: Error reading message body: %v", err)
		return &smtp.SMTPError{
			Code:         554,
			EnhancedCode: smtp.EnhancedCode{5, 3, 0},
			Message:      "Message body could not be read",
		}
	}

	inbox, err := s.core.InboxService.GetByEmailWithWildcard(ctx, s.to)
	if err != nil {
		s.core.Logger.Error("MTA: Error fetching inbox for %s: %v", s.to, err)
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{4, 3, 0},
			Message:      "Temporary error while processing recipient",
		}
	}

	if inbox == nil {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "Recipient address rejected: User unknown",
		}
	}

	message := &models.Message{
		InboxID:  inbox.ID,
		Sender:   s.from,
		Receiver: s.to,
		Subject:  header.Get("Subject"),
		Body:     body.String(),
		IsRead:   false,
	}

	if err := s.core.MessageService.Store(ctx, message); err != nil {
		s.core.Logger.Error("MTA: Error storing message: %v", err)
		return &smtp.SMTPError{
			Code:         554,
			EnhancedCode: smtp.EnhancedCode{5, 3, 0},
			Message:      "Message could not be stored",
		}
	}

	s.core.Logger.Info("MTA: Message stored successfully for %s", s.to)
	return nil
}

func (s *MTASession) Logout() error {
	return nil
}

func (s *MTASession) AuthPlain(username, password string) error {
	// Since this is a simple MTA server, we do not support authentication.
	return &smtp.SMTPError{
		Code:         554,
		EnhancedCode: smtp.EnhancedCode{5, 5, 1},
		Message:      "Command not implemented",
	}
}
