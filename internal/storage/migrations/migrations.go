package migrations

import (
	"database/sql"
	"log"
)

// RunMigrations runs all database migrations
func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	// Add GitHub username field to users table
	if err := AddGitHubUsernameField(db); err != nil {
		return err
	}

	// Add access codes table
	if err := AddAccessCodesTable(db); err != nil {
		return err
	}

	log.Println("All migrations completed successfully")
	return nil
}
