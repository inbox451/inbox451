package config

import (
	"strings"
	"time"

	"inbox451/internal/logger"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type DatabaseConfig struct {
	URL             string        `koanf:"url"`
	MaxOpenConns    int           `koanf:"max_open_conns"`
	MaxIdleConns    int           `koanf:"max_idle_conns"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime"`
}

// OIDCConfig holds OpenID Connect / OAuth2 settings.
// Moved here from internal/auth to break import cycle.
// would have preferred to use "OIDC auth.OIDCConfig"
type OIDCConfig struct {
	Enabled      bool   `koanf:"enabled"`
	ProviderURL  string `koanf:"provider_url"`
	RedirectURL  string `koanf:"redirect_url"` // Note: This might be better constructed at runtime
	ClientID     string `koanf:"client_id"`
	ClientSecret string `koanf:"client_secret"`
}

type Config struct {
	Server struct {
		HTTP struct {
			Port string `koanf:"port"`
		} `koanf:"http"`
		SMTP struct {
			Port     string `koanf:"port"`
			Hostname string `koanf:"hostname"`
			Username string `koanf:"username"`
			Password string `koanf:"password"`
		} `koanf:"smtp"`
		IMAP struct {
			Port     string `koanf:"port"`
			Hostname string `koanf:"hostname"`
		}
	} `koanf:"server"`
	Database DatabaseConfig `koanf:"database"`
	Logging  struct {
		Level  logger.Level `koanf:"level"`
		Format string       `koanf:"format"`
	} `koanf:"logging"`
	OIDC OIDCConfig `koanf:"oidc"`
}

func LoadConfig(configFile string, ko *koanf.Koanf) (*Config, error) {
	if err := ko.Load(file.Provider(configFile), yaml.Parser()); err != nil {
		return nil, err
	}

	// Load environment variables
	// INBOX451_SERVER_HTTP_PORT, INBOX451_SERVER_SMTP_PORT, etc.
	if err := ko.Load(env.Provider("INBOX451_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "INBOX451_")), "_", ".", -1)
	}), nil); err != nil {
		return nil, err
	}

	var config Config
	if err := ko.Unmarshal("", &config); err != nil {
		return nil, err
	}

	return &config, nil
}
