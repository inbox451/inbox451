package smtp

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"inbox451/internal/core"
	"inbox451/internal/models"

	"github.com/emersion/go-message"
	"github.com/emersion/go-smtp"
	"golang.org/x/net/context"
)

type SmtpServer struct {
	core *core.Core
	smtp *smtp.Server
}

type SmtpBackend struct {
	core *core.Core
}

type SmtpSession struct {
	core *core.Core
	from string
	to   string
}

func NewServer(core *core.Core) *SmtpServer {
	be := &SmtpBackend{core: core}
	s := smtp.NewServer(be)

	s.Addr = core.Config.Server.SMTP.Port

	// Determine the domain to use for SMTP server
	if core.Config.Server.EmailDomain != "" {
		s.Domain = core.Config.Server.EmailDomain
		core.Logger.Info("SMTP server domain set to: %s", s.Domain)
	} else if core.Config.Server.SMTP.Hostname != "" {
		s.Domain = core.Config.Server.SMTP.Hostname
		core.Logger.Warn("EmailDomain not configured, falling back to SMTP.Hostname: %s. Consider setting EmailDomain instead.", s.Domain)
	} else {
		// s.Domain will be empty, go-smtp will use os.Hostname()
		core.Logger.Warn("Neither EmailDomain nor SMTP.Hostname configured. SMTP server will use OS hostname, which might not be desirable.")
	}

	s.AllowInsecureAuth = true

	return &SmtpServer{
		core: core,
		smtp: s,
	}
}

func (s *SmtpServer) ListenAndServe() error {
	return s.smtp.ListenAndServe()
}

func (be *SmtpBackend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &SmtpSession{core: be.core}, nil
}

func (s *SmtpSession) Mail(from string, _ *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *SmtpSession) Rcpt(to string, _ *smtp.RcptOptions) error {
	s.to = to
	return nil
}

func (s *SmtpSession) Data(r io.Reader) error {
	// Create a context with timeout for the email processing
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse the email to get the subject
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return err
	}

	// Parse email headers and body
	msg, err := message.Read(bytes.NewReader(buf.Bytes()))
	if err != nil {
		s.core.Logger.Error("Failed to parse email: %v", err)
		return err
	}

	header := msg.Header
	body := new(bytes.Buffer)
	if _, err := io.Copy(body, msg.Body); err != nil {
		s.core.Logger.Error("Failed to read message body: %v", err)
		return err
	}

	message := &models.Message{
		Body:     body.String(),
		Sender:   s.from,
		Receiver: s.to,
		Subject:  header.Get("Subject"),
		InboxID:  0, // We need to look up the inbox ID based on the recipient email
	}

	// Look up the inbox ID based on the recipient email
	inbox, err := s.core.Repository.GetInboxByEmail(ctx, s.to)
	if err != nil {
		s.core.Logger.Error("Failed to find inbox for email %s: %v", s.to, err)
		return err
	}
	if inbox == nil {
		s.core.Logger.Error("No inbox found for email %s", s.to)
		return fmt.Errorf("no inbox found for recipient")
	}

	message.InboxID = inbox.ID

	s.core.Logger.Info("Received email from %s to %s", s.from, s.to)

	if err := s.core.StoreMessage(message); err != nil {
		s.core.Logger.Error("Failed to store message: %v", err)
		return err
	}

	s.core.Logger.Debug("Stored message successfully")
	return nil
}

func (s *SmtpSession) Reset() {}

func (s *SmtpSession) Logout() error {
	return nil
}

func (s *SmtpSession) AuthPlain(username, password string) error {
	return nil // TODO: For now, accept all auth
}

func (s *SmtpServer) Shutdown(ctx context.Context) error {
	return s.smtp.Close()
}
