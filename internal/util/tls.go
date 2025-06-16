package util

import (
	"crypto/tls"
	"inbox451/internal/core"
	"os"
)

func GetTLSConfig(core *core.Core, certFile string, keyFile string) *tls.Config {
	core.Logger.Info("Loading TLS configuration from cert=%s and key=%s", certFile, keyFile)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		core.Logger.ErrorWithStack(err)
		os.Exit(1)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}
