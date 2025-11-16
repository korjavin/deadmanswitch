package migrations

import (
	"database/sql"
	"log"
)

// AddAccessCodesTable creates the access_codes table for secure access code storage
func AddAccessCodesTable(db *sql.DB) error {
	log.Println("Adding access_codes table...")

	query := `
	CREATE TABLE IF NOT EXISTS access_codes (
		id TEXT PRIMARY KEY,
		code TEXT NOT NULL,  -- Store hashed version
		recipient_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		delivery_event_id TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		used_at TIMESTAMP,
		attempt_count INTEGER DEFAULT 0,
		max_attempts INTEGER DEFAULT 5,
		FOREIGN KEY (recipient_id) REFERENCES recipients(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (delivery_event_id) REFERENCES delivery_events(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_access_codes_code ON access_codes(code);
	CREATE INDEX IF NOT EXISTS idx_access_codes_expires_at ON access_codes(expires_at);
	CREATE INDEX IF NOT EXISTS idx_access_codes_recipient_id ON access_codes(recipient_id);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Failed to create access_codes table: %v", err)
		return err
	}

	log.Println("access_codes table added successfully")
	return nil
}
