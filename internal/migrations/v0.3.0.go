package migrations

import (
	"log"

	"inbox451/internal/config"

	"github.com/jmoiron/sqlx"
)

func V0_3_0(db *sqlx.DB, config *config.Config, log *log.Logger) error {
	log.Print("running migration v0.3.0")

	schema := []string{
		// Add is_deleted column to messages table
		`ALTER TABLE messages ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT false`,

		// Create index for efficient filtering by inbox_id and is_deleted
		`CREATE INDEX IF NOT EXISTS idx_messages_inbox_id_is_deleted ON messages (inbox_id, is_deleted)`,

		// Create additional indexes for IMAP filtering operations
		`CREATE INDEX IF NOT EXISTS idx_messages_inbox_id_is_read_is_deleted ON messages (inbox_id, is_read, is_deleted)`,
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
	}()

	for _, s := range schema {
		_, err = tx.Exec(s)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Print("migration v0.3.0 complete")

	return nil
}
