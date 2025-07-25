package msa

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"inbox451/internal/core"
	"inbox451/internal/models"
	"inbox451/internal/util"

	"github.com/emersion/go-message"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
)

type MSAServer struct {
	core *core.Core
	smtp *smtp.Server
}

type MSASession struct {
	core         *core.Core
	to           string
	from         string
	authUsername string
}

type MSABackend struct {
	core *core.Core
}

func (s *MSASession) AuthMechanisms() []string {
	return []string{sasl.Plain}
}

func (backend MSABackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	backend.core.Logger.Info("MSA: New connection from %s", c.Conn().RemoteAddr())
	session := &MSASession{
		core: backend.core,
	}
	session.Reset()
	backend.core.Logger.Info("MSA: New session created")
	return session, nil
}

func (s *MSASession) Auth(mech string) (sasl.Server, error) {
	s.Reset()
	return sasl.NewPlainServer(func(identity, username, password string) error {
		return s.AuthPlain(identity, username, password)
	}), nil
}

func NewServer(core *core.Core) *MSAServer {
	backend := &MSABackend{core: core}
	s := smtp.NewServer(backend)
	s.Addr = core.Config.Server.SMTP.Hostname + ":" + core.Config.Server.SMTP.MSA.Port
	s.Domain = core.Config.Server.SMTP.Domain
	s.Debug = os.Stdout

	s.AllowInsecureAuth = core.Config.Server.SMTP.AllowInsecureAuth

	if core.Config.Server.SMTP.MSA.EnableTLS {
		config, err := util.GetTLSConfig(core, core.Config.Server.TLS.Cert, core.Config.Server.TLS.Key)
		if err != nil {
			core.Logger.Error("MSA: Failed to load TLS configuration: %v . Aborting!", err)
			os.Exit(1)
		}
		s.TLSConfig = config
	}

	return &MSAServer{
		core: core,
		smtp: s,
	}
}

func (s *MSAServer) ListenAndServe() error {
	s.core.Logger.Info("MSA: Starting SMTP server on %s", s.smtp.Addr)
	if err := s.smtp.ListenAndServe(); err != nil {
		s.core.Logger.Error("MSA: Error starting SMTP server: %v", err)
		return err
	}
	return nil
}

func (s *MSAServer) Shutdown(ctx context.Context) error {
	s.core.Logger.Info("MSA: Shutting down SMTP server")
	if err := s.smtp.Shutdown(ctx); err != nil {
		s.core.Logger.Error("MSA: Error shutting down SMTP server: %v", err)
		return err
	}
	return nil
}

func (s *MSASession) Reset() {
	s.from = ""
	s.to = ""
	s.authUsername = ""
}

func (s *MSASession) Logout() error {
	s.core.Logger.Info("MSA: Session logged out")
	return nil
}

func (s *MSASession) AuthPlain(identity, username, password string) error {
	s.core.Logger.Info("MSA: Authentication attempt for username '%s'", username)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	user, err := s.core.UserService.LoginWithToken(ctx, username, password)
	defer cancel()
	if err != nil {
		if errors.Is(err, core.ErrAuthFailed) {
			s.core.Logger.Info("MSA: Authentication failed for username '%s': %v", username, err)
			return &smtp.SMTPError{
				Code:         535,
				EnhancedCode: smtp.EnhancedCode{5, 7, 8},
				Message:      "Authentication credentials invalid",
			}
		}
		if errors.Is(err, core.ErrAccountInactive) {
			s.core.Logger.Info("MSA: Authentication failed because of user account disabled '%s': %v", username, err)
			return &smtp.SMTPError{
				Code:         550,
				EnhancedCode: smtp.EnhancedCode{5, 7, 1},
				Message:      "User account disabled",
			}
		}
		s.core.Logger.Error("MSA: Authentication failed with unexpected error '%s': %v", username, err)
		return &smtp.SMTPError{
			Code:         454,
			EnhancedCode: smtp.EnhancedCode{4, 7, 0},
			Message:      "Temporary authentication failure",
		}
	}
	s.core.Logger.Info("MSA: Authentication successful for username '%s'", username)
	s.authUsername = user.Username
	return nil
}

func (s *MSASession) RequireAuthentication() error {
	if s.authUsername == "" {
		s.core.Logger.Info("MSA: Authentication required before sending email")
		return &smtp.SMTPError{
			Code:         530,
			EnhancedCode: smtp.EnhancedCode{5, 7, 0},
			Message:      "Authentication required",
		}
	}
	return nil
}

func (s *MSASession) Mail(from string, opts *smtp.MailOptions) error {
	s.core.Logger.Info("MSA: Mail from %s", from)
	if err := s.RequireAuthentication(); err != nil {
		return err
	}
	s.from = from
	return nil
}

func (s *MSASession) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.core.Logger.Info("MSA: Recipient to %s", to)
	if err := s.RequireAuthentication(); err != nil {
		return err
	}

	// Validate the domain of the recipient email address
	expectedDomain := "@" + s.core.Config.Server.EmailDomain
	if !strings.HasSuffix(to, expectedDomain) {
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
		if errors.Is(err, core.ErrNotFound) {
			s.core.Logger.Info("MTA: Recipient %s not found in inboxes", to)
			return &smtp.SMTPError{
				Code:         550,
				EnhancedCode: smtp.EnhancedCode{5, 1, 1},
				Message:      "Recipient address rejected: User unknown",
			}
		}

		s.core.Logger.Error("MTA: Error fetching inbox for %s: %v", to, err)
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{4, 3, 0},
			Message:      "Temporary error while processing recipient",
		}
	}

	s.core.Logger.Info("MSA: Recipient %s accepted for user %s (inbox ID: %s)", to, s.authUsername, inbox.ID)
	s.to = to
	return nil
}

func (s *MSASession) Data(r io.Reader) error {
	s.core.Logger.Info("MSA: Processing email data from %s to %s", s.from, s.to)
	if err := s.RequireAuthentication(); err != nil {
		return err
	}

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
		if errors.Is(err, core.ErrNotFound) {
			s.core.Logger.Info("MTA: Recipient %s not found in inboxes", s.to)
			return &smtp.SMTPError{
				Code:         550,
				EnhancedCode: smtp.EnhancedCode{5, 1, 1},
				Message:      "Recipient address rejected: User unknown",
			}
		}

		s.core.Logger.Error("MTA: Error fetching inbox for %s: %v", s.to, err)
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{4, 3, 0},
			Message:      "Temporary error while processing recipient",
		}
	}

	m := &models.Message{
		InboxID:  inbox.ID,
		Sender:   s.from,
		Receiver: s.to,
		Subject:  header.Get("Subject"),
		Body:     body.String(),
		IsRead:   false,
	}

	if err := s.core.MessageService.Store(ctx, m); err != nil {
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
