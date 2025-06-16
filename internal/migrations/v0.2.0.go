package migrations

import (
	"database/sql"
	"fmt"
	"log"

	"inbox451/internal/config"

	"github.com/jmoiron/sqlx"
)

func V0_2_0(db *sqlx.DB, config *config.Config, log *log.Logger) error {
	log.Print("Running migration v0.2.0: Add Authentication")

	schema := []string{

		// Ensure sessions table exists and matches simplesessions requirements
		// See https://github.com/zerodha/simplesessions/blob/v3.0.0/stores/postgres/postgres.go#L4
		// dropping sessions table from first migration to ensure it is created with the correct schema
		`DROP TABLE IF EXISTS sessions CASCADE`, // Drop existing first if schema needs change
		`CREATE TABLE IF NOT EXISTS sessions (
            id TEXT NOT NULL PRIMARY KEY,
            data JSONB DEFAULT '{}'::jsonb NOT NULL,
            created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT now() NOT NULL
        )`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_id ON sessions (id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_id_created_at ON sessions (id, created_at)`,

		// Ensure tokens table exists
		`CREATE TABLE IF NOT EXISTS tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE,
			last_used_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_token ON tokens(token)`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id)`,

		// Add is_deleted column to messages table
		`ALTER TABLE messages ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT false`,

		// Create index for efficient filtering by inbox_id and is_deleted
		`CREATE INDEX IF NOT EXISTS idx_messages_inbox_id_is_deleted ON messages (inbox_id, is_deleted)`,

		// Create additional indexes for IMAP filtering operations
		`CREATE INDEX IF NOT EXISTS idx_messages_inbox_id_is_read_is_deleted ON messages (inbox_id, is_read, is_deleted)`,
	}

	// Start a transaction
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure proper rollback handling
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	// Execute the schema changes
	for _, query := range schema {
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("failed to execute schema update '%s': %w", query, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Print("Finished migration v0.2.0")
	return nil
}
