package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// AddPasskeyFlags adds the backup_eligible/backup_state columns to the passkeys
// table. These mirror the WebAuthn BE/BS authenticator flags. go-webauthn
// validates BackupEligible equality during login, so persisting the value at
// registration is required for cloud-synced passkeys (iCloud Keychain, Google
// Password Manager, 1Password) to authenticate.
func AddPasskeyFlags(db *sql.DB) error {
	log.Println("Running migration: Adding backup_eligible/backup_state to passkeys table")

	for _, col := range []string{"backup_eligible", "backup_state"} {
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('passkeys')
			WHERE name = ?
		`, col).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check if %s column exists: %w", col, err)
		}
		if count > 0 {
			log.Printf("%s column already exists, skipping", col)
			continue
		}
		if _, err := db.Exec(fmt.Sprintf(`
			ALTER TABLE passkeys
			ADD COLUMN %s BOOLEAN NOT NULL DEFAULT 0
		`, col)); err != nil {
			return fmt.Errorf("failed to add %s column: %w", col, err)
		}
		log.Printf("Successfully added %s to passkeys", col)
	}

	return nil
}
