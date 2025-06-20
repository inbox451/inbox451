package util

import (
	"crypto/tls"

	"inbox451/internal/core"
)

// Pin minimum version to TLS 1.3 to avoid avoid weak protocols
func GetTLSConfig(core *core.Core, certFile string, keyFile string) (*tls.Config, error) {
	core.Logger.Info("Loading TLS configuration from cert=%s and key=%s", certFile, keyFile)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}, nil
}
