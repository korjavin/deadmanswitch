package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// AddGitHubUsernameField adds the github_username field to the users table
func AddGitHubUsernameField(db *sql.DB) error {
	log.Println("Running migration: Adding github_username field to users table")

	// Check if the column already exists
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users')
		WHERE name = 'github_username'
	`).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to check if github_username column exists: %w", err)
	}

	if count > 0 {
		log.Println("github_username column already exists, skipping migration")
		return nil
	}

	// Add the column
	_, err = db.Exec(`
		ALTER TABLE users
		ADD COLUMN github_username TEXT DEFAULT NULL
	`)

	if err != nil {
		return fmt.Errorf("failed to add github_username column: %w", err)
	}

	log.Println("Successfully added github_username field to users table")
	return nil
}
