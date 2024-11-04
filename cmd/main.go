package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"mercury/internal/api"
	"mercury/internal/config"
	"mercury/internal/core"
	"mercury/internal/imap"
	"mercury/internal/smtp"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/providers/env"
	_ "github.com/lib/pq"
)

var (
	db *sqlx.DB
	lo = log.New(os.Stderr, "", 0)
)

func initDB(cfg *config.Config) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("postgres", cfg.Database.URL)
		if err == nil {
			break
		}
		fmt.Printf("Failed to connect to database, retrying in 2 seconds... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d retries: %v", maxRetries, err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return db, nil
}

type serverError struct {
	server string
	err    error
}

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type ServerInstance struct {
	server Server
	name   string
}

func startServers(core *core.Core) error {
	// Create error channel for servers
	errChan := make(chan serverError, 3)

	// Create a channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Initialize all servers
	servers := []ServerInstance{
		{server: api.NewServer(core), name: "HTTP"},
		{server: smtp.NewServer(core), name: "SMTP"},
		{server: imap.NewServer(core), name: "IMAP"},
	}

	// Start all servers
	for _, s := range servers {
		go func(s ServerInstance) {
			core.Logger.Info("Starting %s server", s.name)
			if err := s.server.ListenAndServe(); err != nil {
				if err != http.ErrServerClosed {
					errChan <- serverError{server: s.name, err: err}
				}
			}
		}(s)
	}

	// Wait for shutdown signal or error
	select {
	case <-stop:
		core.Logger.Info("Received shutdown signal")
		return handleGracefulShutdown(core, servers)
	case err := <-errChan:
		return fmt.Errorf("%s server error: %v", err.server, err.err)
	}
}

func handleGracefulShutdown(core *core.Core, servers []ServerInstance) error {
	core.Logger.Info("Initiating graceful shutdown...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create error channel for shutdown errors
	errChan := make(chan error, len(servers))

	// Shutdown all servers concurrently
	for _, s := range servers {
		go func(s ServerInstance) {
			core.Logger.Info("Shutting down %s server...", s.name)
			if err := s.server.Shutdown(ctx); err != nil {
				errChan <- fmt.Errorf("%s shutdown error: %v", s.name, err)
				return
			}
			core.Logger.Info("%s server shutdown completed", s.name)
			errChan <- nil
		}(s)
	}

	// Wait for all servers to shutdown or timeout
	var shutdownErrors []error
	for i := 0; i < len(servers); i++ {
		if err := <-errChan; err != nil {
			shutdownErrors = append(shutdownErrors, err)
		}
	}

	// Handle any shutdown errors
	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown errors: %v", shutdownErrors)
	}

	core.Logger.Info("Graceful shutdown completed")
	return nil
}

func init() {
	ko := initFlags()
	initConfigFiles(ko.Strings("config"), ko)

	// Load environment variables
	if err := ko.Load(env.Provider("MERCURY_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "MERCURY_")), "__", ".", -1)
	}), nil); err != nil {
		lo.Fatalf("error loading config from env: %v", err)
	}

	// Check if the DB schema is installed.
	if ok, err := checkSchema(db); err != nil {
		log.Fatalf("error checking schema in DB: %v", err)
	} else if !ok {
		lo.Fatal("the database does not appear to be setup. Run --install.")
	}
}

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config/default.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create core
	core, err := core.NewCore(cfg, db)
	if err != nil {
		fmt.Printf("Failed to create core: %v\n", err)
		os.Exit(1)
	}

	core.Logger.Info("Starting application with configuration from %s", *configFile)

	// Start all servers
	if err := startServers(core); err != nil {
		core.Logger.Fatal("Server error: %v", err)
	}
}
