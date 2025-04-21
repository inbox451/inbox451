package main

import (
	"fmt"
	"log"
	"strings"

	"inbox451/internal/core"
	"inbox451/internal/models"

	"github.com/volatiletech/null/v9"
	"golang.org/x/crypto/bcrypt"

	"inbox451/internal/config"

	"github.com/jmoiron/sqlx"
)

func install(db *sqlx.DB, config *config.Config, prompt, idempotent bool) {
	// Check if the database is already initialized.
	// If the database is not initialized, we should get "v0.0.0" as the version.
	version, err := getLastMigrationVersion(db)
	if err != nil {
		logger.Fatalf("Error getting last migration version: %v", err)
	}

	if version != "v0.0.0" {
		if idempotent {
			logger.Printf("Database is already initialized at version %s. Idempotent install skipping.", version)
			return
		}
		logger.Fatalf("Database is already initialized. Current version is %s. Use --upgrade instead.", version)
	}

	if prompt {
		var ok string
		fmt.Printf("** IMPORTANT: This will initialize the database schema.\n")
		fmt.Print("Continue (y/n)?  ")
		if _, err := fmt.Scanf("%s", &ok); err != nil {
			logger.Fatalf("error reading value from terminal: %v", err)
		}
		if strings.ToLower(ok) != "y" {
			fmt.Println("install cancelled")
			return
		}
	}

	// Fetch all available migrations and run them.
	_, toRun, err := getPendingMigrations(db)
	if err != nil {
		logger.Fatalf("Error checking migrations: %v", err)
	}

	// No migrations to run.
	if len(toRun) == 0 {
		logger.Println("No migrations found to run during install.")
	} else {
		logger.Printf("Found %d migration(s) to run for installation.", len(toRun))
		for _, m := range toRun {
			log.Printf("Running migration %s", m.version)
			if err := m.fn(db, config, logger); err != nil {
				log.Fatalf("Error running migration %s: %v", m.version, err)
			}

			if err := recordMigrationVersion(m.version, db); err != nil {
				log.Fatalf("Error recording migration version %s: %v", m.version, err)
			}
		}
		logger.Println("All migrations executed successfully.")
	}

	if err := createDefaultAdminUserIfNeeded(db, logger); err != nil {
		logger.Fatalf("Error creating default admin user: %v", err)
	}

	logger.Println("Installation complete.")
}

func checkSchema(db *sqlx.DB) (bool, error) {
	if _, err := db.Exec(`SELECT version FROM schema_migrations LIMIT 1`); err != nil {
		if isTableNotExistErr(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func checkInstall(db *sqlx.DB) {
	if ok, err := checkSchema(db); err != nil {
		logger.Fatalf("error checking schema in DB: %v", err)
	} else if !ok {
		logger.Fatal("The database does not appear to be setup. Run --install.")
	}
}

func recordMigrationVersion(version string, db *sqlx.DB) error {
	_, err := db.Exec(fmt.Sprintf(`INSERT INTO schema_migrations (version) VALUES('%s')`, version))
	return err
}

func createDefaultAdminUserIfNeeded(db *sqlx.DB, logger *log.Logger) error {
	logger.Println("Checking if default admin user needs to be created...")

	var userCount int
	err := db.Get(&userCount, "SELECT COUNT(*) FROM users")
	if err != nil {
		// If the table doesn't exist yet (e.g., during initial installation before v0.1.0 runs fully),
		// this check might fail. We should ideally run this *after* migrations ensure the table exists.
		// However, the current structure runs it after migrations complete, so this should be fine.
		if isTableNotExistErr(err) {
			logger.Println("Users table does not exist yet, skipping default admin creation (will run after migrations).")
			return nil // Not an error in this context, migrations will create it.
		}
		return fmt.Errorf("failed to count users: %w", err)
	}

	if userCount == 0 {
		logger.Println("No users found. Creating default admin user 'admin'.")

		// Generate a random password
		defaultPassword, err := core.GenerateSecureTokenBase64() // Reuse the token generator
		if err != nil {
			return fmt.Errorf("failed to generate random password: %w", err)
		}
		// Trim the password to a reasonable length
		if len(defaultPassword) > 16 {
			defaultPassword = defaultPassword[:16]
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash default password: %w", err)
		}

		adminUser := &models.User{
			Name:          "Admin User",
			Username:      "admin",
			Email:         "admin@inbox451.dev",
			Status:        "active",
			Role:          "admin",
			PasswordLogin: true,
			Password:      null.StringFrom(string(hashedPassword)),
		}

		// Use a transaction for the insert
		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for admin user creation: %w", err)
		}
		defer tx.Rollback() // Rollback on error

		// Prepare the statement within the transaction
		stmt, err := tx.Preparex(`
            INSERT INTO users (name, username, password, email, status, role, password_login)
            VALUES ($1, $2, $3, $4, $5, $6, $7)`)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement for admin user: %w", err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(
			adminUser.Name,
			adminUser.Username,
			adminUser.Password,
			adminUser.Email,
			adminUser.Status,
			adminUser.Role,
			adminUser.PasswordLogin,
		)
		if err != nil {
			return fmt.Errorf("failed to insert default admin user: %w", err)
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for admin user creation: %w", err)
		}

		// Log the generated password!
		logger.Println("************************************************************")
		logger.Printf("Default admin user created.")
		logger.Printf("Username: admin")
		logger.Printf("E-mail: admin@inbox451.dev")
		logger.Printf("Password: %s", defaultPassword)
		logger.Println("PLEASE STORE THIS PASSWORD SECURELY AND CHANGE IT IF NEEDED.")
		logger.Println("************************************************************")

	} else {
		logger.Println("Users already exist, skipping default admin creation.")
	}

	return nil
}
