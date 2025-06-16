package main

import (
	"fmt"
	"log"
	"strings"

	"inbox451/internal/config"
	"inbox451/internal/migrations"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/mod/semver"
)

type migFunc struct {
	version string
	fn      func(*sqlx.DB, *config.Config, *log.Logger) error
}

var migList = []migFunc{
	{"v0.1.0", migrations.V0_1_0},
	{"v0.2.0", migrations.V0_2_0},
}

func upgrade(db *sqlx.DB, config *config.Config, prompt bool) {
	if prompt {
		var ok string
		fmt.Printf("** IMPORTANT: Take a backup of the database before upgrading.\n")
		fmt.Print("Continue (y/n)?  ")
		if _, err := fmt.Scanf("%s", &ok); err != nil {
			logger.Fatalf("error reading value from terminal: %v", err)
		}
		if strings.ToLower(ok) != "y" {
			fmt.Println("upgrade cancelled")
			return
		}
	}

	_, toRun, err := getPendingMigrations(db)
	if err != nil {
		logger.Fatalf("error checking migrations: %v", err)
	}

	if len(toRun) == 0 {
		logger.Printf("no upgrades to run. Database is up to date.")
	} else {
		logger.Printf("Found %d pending upgrade(s) to run.", len(toRun))
		for _, m := range toRun {
			log.Printf("running migration %s", m.version)
			if err := m.fn(db, config, logger); err != nil {
				log.Fatalf("error running migration %s: %v", m.version, err)
			}

			if err := recordMigrationVersion(m.version, db); err != nil {
				log.Fatalf("error recording migration version %s: %v", m.version, err)
			}
		}
		logger.Println("All migrations executed successfully.")
	}

	// Check if the default admin user exists and create it if not.
	if err := createDefaultAdminUserIfNeeded(db, logger); err != nil {
		logger.Fatalf("Error creating default admin user during upgrade: %v", err)
	}

	log.Printf("upgrade complete")
}

func checkUpgrade(db *sqlx.DB) {
	lastVer, toRun, err := getPendingMigrations(db)
	if err != nil {
		log.Fatalf("error checking migrations: %v", err)
	}

	if len(toRun) == 0 {
		return
	}

	var vers []string
	for _, m := range toRun {
		vers = append(vers, m.version)
	}

	log.Fatalf("there are %d pending database upgrade(s): %v. The last upgrade was %s. Backup the database and run inbox451 --upgrade",
		len(toRun), vers, lastVer)
}

func getPendingMigrations(db *sqlx.DB) (string, []migFunc, error) {
	lastVer, err := getLastMigrationVersion(db)
	if err != nil {
		return "", nil, err
	}

	var toRun []migFunc
	for i, m := range migList {
		if semver.Compare(m.version, lastVer) > 0 {
			toRun = migList[i:]
			break
		}
	}

	return lastVer, toRun, nil
}

// getLastMigrationVersion returns the last migration semver recorded in the DB.
// If there isn't any, `v0.0.0` is returned.
func getLastMigrationVersion(db *sqlx.DB) (string, error) {
	var v string
	if err := db.Get(&v, `
		SELECT COALESCE(
			(SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1),
		'v0.0.0')`); err != nil {
		if isTableNotExistErr(err) {
			return "v0.0.0", nil
		}
		return v, err
	}
	return v, nil
}

// isTableNotExistErr checks if the given error represents a Postgres/pq
// "table does not exist" error.
func isTableNotExistErr(err error) bool {
	if p, ok := err.(*pq.Error); ok {
		if p.Code == "42P01" {
			return true
		}
	}
	return false
}
