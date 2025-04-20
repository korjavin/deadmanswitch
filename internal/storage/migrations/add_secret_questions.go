package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// AddSecretQuestionsTables adds the tables for secret questions and question sets
func AddSecretQuestionsTables(db *sql.DB) error {
	log.Println("Running migration: Adding secret questions tables")

	// Check if the tables already exist
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master
		WHERE type='table' AND name='secret_questions'
	`).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to check if secret_questions table exists: %w", err)
	}

	if count > 0 {
		log.Println("secret_questions table already exists, skipping migration")
		return nil
	}

	// Create the tables
	_, err = db.Exec(`
	-- SecretQuestions table
	CREATE TABLE IF NOT EXISTS secret_questions (
		id TEXT PRIMARY KEY,
		secret_assignment_id TEXT NOT NULL,
		question TEXT NOT NULL,
		salt BLOB NOT NULL,
		encrypted_share BLOB NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (secret_assignment_id) REFERENCES secret_assignments(id) ON DELETE CASCADE
	);

	-- SecretQuestionSets table
	CREATE TABLE IF NOT EXISTS secret_question_sets (
		id TEXT PRIMARY KEY,
		secret_assignment_id TEXT NOT NULL,
		threshold INTEGER NOT NULL,
		total_questions INTEGER NOT NULL,
		timelock_round INTEGER NOT NULL,
		encrypted_blob BLOB NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (secret_assignment_id) REFERENCES secret_assignments(id) ON DELETE CASCADE
	);

	-- Create indexes for performance
	CREATE INDEX IF NOT EXISTS idx_secret_questions_assignment_id ON secret_questions(secret_assignment_id);
	CREATE INDEX IF NOT EXISTS idx_secret_question_sets_assignment_id ON secret_question_sets(secret_assignment_id);
	CREATE INDEX IF NOT EXISTS idx_secret_question_sets_timelock_round ON secret_question_sets(timelock_round);
	`)

	if err != nil {
		return fmt.Errorf("failed to create secret questions tables: %w", err)
	}

	log.Println("Successfully created secret questions tables")
	return nil
}
