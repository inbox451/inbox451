package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inbox451/internal/api"
	"inbox451/internal/assets"
	"inbox451/internal/config"
	"inbox451/internal/core"
	"inbox451/internal/imap"
	"inbox451/internal/smtp"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	_ "github.com/lib/pq"
	"github.com/spf13/pflag"
)

var (
	logger  = log.New(os.Stderr, "", 0)
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func initDB(cfg *config.Config) (*sqlx.DB, *sql.DB, error) {
	var db *sqlx.DB
	var stdDB *sql.DB
	var err error

	const maxRetries = 5
	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("postgres", cfg.Database.URL)
		if err == nil {
			break
		}
		fmt.Printf("Failed to connect to database. Retrying in 2 seconds... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database after %d retries: %w", maxRetries, err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// TODO: Check if the database is reachable
	stdDB = db.DB

	return db, stdDB, nil
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

func startServers(core *core.Core, db *sql.DB) error {
	// Create error channel for servers
	const numServers = 3
	errChan := make(chan serverError, numServers)

	// Create a channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Initialize all servers
	servers := []ServerInstance{
		{server: api.NewServer(core, db), name: "HTTP"},
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

	// Wait for all servers to shut down or timeout
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

func initFlags() *koanf.Koanf {
	ko := koanf.New(".")

	f := pflag.NewFlagSet("config", pflag.ContinueOnError)

	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}

	f.String("config", "config.yml", "path to the config file")
	f.Bool("idempotent", false, "make --install run only if the database isn't already setup")
	f.Bool("install", false, "setup database (first time)")
	f.Bool("upgrade", false, "upgrade database to the current version")
	f.Bool("yes", false, "assume 'yes' to prompts during --install/upgrade")

	if err := f.Parse(os.Args[1:]); err != nil {
		logger.Fatalf("error loading flags: %v", err)
	}

	if err := ko.Load(posflag.Provider(f, ".", ko), nil); err != nil {
		logger.Fatalf("error loading config: %v", err)
	}

	return ko
}

func main() {
	// Get executable path for stuffbin
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	// Initialize asset system
	if err := assets.InitAssets(execPath); err != nil {
		fmt.Printf("Failed to initialize assets: %v\n", err)
		os.Exit(1)
	}

	// Parse command line flags
	ko := initFlags()
	cfg, err := config.LoadConfig(ko.String("config"), ko)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	db, stdDB, err := initDB(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if ko.Bool("install") {
		install(db, cfg, !ko.Bool("yes"), ko.Bool("idempotent"))
		os.Exit(0)
	}

	// Check if the DB schema is installed.
	checkInstall(db)

	if ko.Bool("upgrade") {
		upgrade(db, cfg, !ko.Bool("yes"))
		os.Exit(0)
	}

	// Check DB migrations and up-to-date
	checkUpgrade(db)

	// Create core
	core, err := core.NewCore(cfg, db, version, commit, date)
	if err != nil {
		fmt.Printf("Failed to create core: %v\n", err)
		os.Exit(1)
	}

	// Log version information at startup
	core.Logger.Info("Starting inbox451 version %s (commit: %s, built: %s)", version, commit, date)

	// Start all servers
	if err := startServers(core, stdDB); err != nil {
		core.Logger.Fatal("Server error: %v", err)
	}
}
