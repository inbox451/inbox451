package migrations

import (
	"database/sql"
	"fmt"
	"log"

	"inbox451/internal/config"

	"github.com/jmoiron/sqlx"
)

func V0_1_0(db *sqlx.DB, config *config.Config, log *log.Logger) error {
	log.Print("Running migration v0.1.0")
	schema := []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY
		)`,

		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
				CREATE TYPE user_role AS ENUM ('user', 'admin');
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'project_role') THEN
				CREATE TYPE project_role AS ENUM ('user', 'admin');
			END IF;
		END
		$$`,

		`CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL CHECK (LENGTH(name) >= 2),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			status VARCHAR(50) NOT NULL DEFAULT 'inactive',
			role user_role NOT NULL DEFAULT 'user',
			password_login BOOLEAN NOT NULL DEFAULT false,
			loggedin_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS project_users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role project_role NOT NULL DEFAULT 'user',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(project_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS inboxes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			email VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS tokens (
        	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE,
			last_used_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		`
		CREATE INDEX idx_user_tokens_token ON tokens(token);
		CREATE INDEX idx_user_tokens_user_id ON tokens(user_id);
		`,

		`CREATE TABLE IF NOT EXISTS forward_rules (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
			sender VARCHAR(255),
			receiver VARCHAR(255),
			subject VARCHAR(200),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CHECK (sender IS NOT NULL OR receiver IS NOT NULL OR subject IS NOT NULL)
		)`,

		`CREATE TABLE IF NOT EXISTS messages (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			inbox_id UUID NOT NULL REFERENCES inboxes(id) ON DELETE CASCADE,
			sender VARCHAR(255) NOT NULL,
			receiver VARCHAR(255) NOT NULL,
			subject VARCHAR(200) NOT NULL,
			body TEXT NOT NULL,
			is_read BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE sessions (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		    session_id VARCHAR(255) NOT NULL,
		    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		    data JSONB DEFAULT '{}'::jsonb,
		    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		    ip_address VARCHAR(45),
		    user_agent TEXT,
		    is_active BOOLEAN DEFAULT true,
		    UNIQUE (session_id)
		)`,

		`
		CREATE INDEX idx_sessions_session_id ON sessions(session_id);
		CREATE INDEX idx_sessions_user_id ON sessions(user_id);
		CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
		`,
	}

	// Start a transaction
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure proper rollback handling
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// We can only log this error since we can't return it
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	// Execute the schema
	for _, query := range schema {
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("Failed to execute schema: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}

	return nil
}
