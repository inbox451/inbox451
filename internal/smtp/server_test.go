package smtp

import (
	"io"
	"testing"

	"inbox451/internal/config"
	"inbox451/internal/core"
	"inbox451/internal/logger"
	"inbox451/internal/mocks"

	"github.com/stretchr/testify/assert"
)

func setupSMTPTestCore(t *testing.T) (*core.Core, *mocks.Repository) {
	mockRepo := mocks.NewRepository(t)
	logger := logger.New(io.Discard, logger.DEBUG)

	core := &core.Core{
		Config:     &config.Config{},
		Logger:     logger,
		Repository: mockRepo,
	}

	// Set default SMTP port
	core.Config.Server.SMTP.Port = ":1025"

	return core, mockRepo
}

func TestNewServer_EmailDomainConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		emailDomain  string
		smtpHostname string
		expectDomain string
		expectLog    string
	}{
		{
			name:         "EmailDomain takes precedence over SMTP.Hostname",
			emailDomain:  "mail.example.com",
			smtpHostname: "smtp.example.com",
			expectDomain: "mail.example.com",
			expectLog:    "SMTP server domain set to:",
		},
		{
			name:         "Falls back to SMTP.Hostname when EmailDomain is empty",
			emailDomain:  "",
			smtpHostname: "smtp.example.com",
			expectDomain: "smtp.example.com",
			expectLog:    "EmailDomain not configured, falling back to SMTP.Hostname:",
		},
		{
			name:         "Warning when both are empty",
			emailDomain:  "",
			smtpHostname: "",
			expectDomain: "",
			expectLog:    "Neither EmailDomain nor SMTP.Hostname configured",
		},
		{
			name:         "Uses EmailDomain when set",
			emailDomain:  "example.com",
			smtpHostname: "",
			expectDomain: "example.com",
			expectLog:    "SMTP server domain set to:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, _ := setupSMTPTestCore(t)
			core.Config.Server.EmailDomain = tt.emailDomain
			core.Config.Server.SMTP.Hostname = tt.smtpHostname

			server := NewServer(core)
			assert.NotNil(t, server)
			assert.Equal(t, tt.expectDomain, server.smtp.Domain)
		})
	}
}

func TestSmtpServer_Integration(t *testing.T) {
	tests := []struct {
		name        string
		emailDomain string
		setupCore   func(*core.Core)
	}{
		{
			name:        "Server starts with EmailDomain configured",
			emailDomain: "test.example.com",
			setupCore: func(c *core.Core) {
				c.Config.Server.EmailDomain = "test.example.com"
			},
		},
		{
			name:        "Server starts without EmailDomain",
			emailDomain: "",
			setupCore: func(c *core.Core) {
				c.Config.Server.SMTP.Hostname = "smtp.example.com"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, _ := setupSMTPTestCore(t)
			tt.setupCore(core)

			server := NewServer(core)
			assert.NotNil(t, server)
			assert.NotNil(t, server.smtp)
			assert.Equal(t, ":1025", server.smtp.Addr)
		})
	}
}
