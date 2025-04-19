package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage/migrations"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteRepository implements the Repository interface using SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// SQLiteTx is a transaction wrapper for SQLite
type SQLiteTx struct {
	tx *sql.Tx
	*SQLiteRepository
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Connect to SQLite
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection parameters
	db.SetMaxOpenConns(1) // SQLite only supports one writer at a time
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Create repo
	repo := &SQLiteRepository{db: db}

	// Initialize database
	if err := repo.initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return repo, nil
}

// initialize creates tables if they don't exist
func (r *SQLiteRepository) initialize() error {
	// Create tables
	_, err := r.db.Exec(`
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash BLOB NOT NULL,
		telegram_id TEXT DEFAULT NULL,
		telegram_username TEXT DEFAULT NULL,
		last_activity DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		ping_frequency INTEGER NOT NULL DEFAULT 3,
		ping_deadline INTEGER NOT NULL DEFAULT 14,
		pinging_enabled BOOLEAN NOT NULL DEFAULT 0,
		ping_method TEXT NOT NULL DEFAULT 'both',
		next_scheduled_ping DATETIME DEFAULT NULL,
		totp_secret TEXT DEFAULT NULL,
		totp_enabled BOOLEAN NOT NULL DEFAULT 0,
		totp_verified BOOLEAN NOT NULL DEFAULT 0
	);

	-- Secrets table
	CREATE TABLE IF NOT EXISTS secrets (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		encrypted_data TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		encryption_type TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- Recipients table
	CREATE TABLE IF NOT EXISTS recipients (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		email TEXT NOT NULL,
		name TEXT NOT NULL,
		message TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		phone_number TEXT,
		is_confirmed BOOLEAN NOT NULL DEFAULT 0,
		confirmed_at DATETIME DEFAULT NULL,
		confirmation_code TEXT DEFAULT NULL,
		confirmation_sent_at DATETIME DEFAULT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- SecretAssignments table
	CREATE TABLE IF NOT EXISTS secret_assignments (
		id TEXT PRIMARY KEY,
		secret_id TEXT NOT NULL,
		recipient_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (secret_id) REFERENCES secrets(id) ON DELETE CASCADE,
		FOREIGN KEY (recipient_id) REFERENCES recipients(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE (secret_id, recipient_id)
	);

	-- PingHistory table
	CREATE TABLE IF NOT EXISTS ping_history (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		sent_at DATETIME NOT NULL,
		method TEXT NOT NULL,
		status TEXT NOT NULL,
		responded_at DATETIME DEFAULT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- PingVerification table
	CREATE TABLE IF NOT EXISTS ping_verification (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		code TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		used BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- DeliveryEvent table
	CREATE TABLE IF NOT EXISTS delivery_events (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		recipient_id TEXT NOT NULL,
		sent_at DATETIME NOT NULL,
		status TEXT NOT NULL,
		error_message TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (recipient_id) REFERENCES recipients(id) ON DELETE CASCADE
	);

	-- AuditLog table
	CREATE TABLE IF NOT EXISTS audit_log (
		id TEXT PRIMARY KEY,
		user_id TEXT,
		action TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		ip_address TEXT,
		user_agent TEXT,
		details TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	-- Session table
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token TEXT UNIQUE NOT NULL,
		created_at DATETIME NOT NULL,
		expires_at DATETIME NOT NULL,
		last_activity DATETIME NOT NULL,
		ip_address TEXT,
		user_agent TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- Passkeys table
	CREATE TABLE IF NOT EXISTS passkeys (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		credential_id BLOB NOT NULL,
		public_key BLOB NOT NULL,
		aaguid BLOB,
		sign_count INTEGER NOT NULL DEFAULT 0,
		name TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		last_used_at DATETIME NOT NULL,
		transports TEXT,
		attestation_type TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	-- Create indexes for performance
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
	CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id);
	CREATE INDEX IF NOT EXISTS idx_recipients_user_id ON recipients(user_id);
	CREATE INDEX IF NOT EXISTS idx_secret_assignments_secret_id ON secret_assignments(secret_id);
	CREATE INDEX IF NOT EXISTS idx_secret_assignments_recipient_id ON secret_assignments(recipient_id);
	CREATE INDEX IF NOT EXISTS idx_secret_assignments_user_id ON secret_assignments(user_id);
	CREATE INDEX IF NOT EXISTS idx_ping_history_user_id ON ping_history(user_id);
	CREATE INDEX IF NOT EXISTS idx_ping_verification_code ON ping_verification(code);
	CREATE INDEX IF NOT EXISTS idx_delivery_events_user_id ON delivery_events(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	CREATE INDEX IF NOT EXISTS idx_passkeys_user_id ON passkeys(user_id);
	CREATE INDEX IF NOT EXISTS idx_passkeys_credential_id ON passkeys(credential_id);
	`)

	if err != nil {
		return err
	}

	// Run migrations
	log.Println("Running database migrations...")

	if err := migrations.RunMigrations(r.db); err != nil {
		return err
	}

	log.Println("Database migrations completed")
	return nil
}

// BeginTx starts a new transaction
func (r *SQLiteRepository) BeginTx(ctx context.Context) (Transaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &SQLiteTx{
		tx:               tx,
		SQLiteRepository: &SQLiteRepository{db: r.db},
	}, nil
}

// Commit commits the transaction
func (t *SQLiteTx) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *SQLiteTx) Rollback() error {
	return t.tx.Rollback()
}

// Helper function to generate a new UUID
func generateID() string {
	return uuid.New().String()
}

// ===== User operations =====

// CreateUser creates a new user
func (r *SQLiteRepository) CreateUser(ctx context.Context, user *models.User) error {
	// Generate ID if not provided
	if user.ID == "" {
		user.ID = generateID()
	}

	// Set timestamps
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.LastActivity = now

	// Execute insert query
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (
			id, email, password_hash, telegram_id, telegram_username, github_username,
			last_activity, created_at, updated_at,
			ping_frequency, ping_deadline, pinging_enabled, ping_method, next_scheduled_ping,
			totp_secret, totp_enabled, totp_verified
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		user.ID, user.Email, user.PasswordHash, user.TelegramID, user.TelegramUsername, user.GitHubUsername,
		user.LastActivity, user.CreatedAt, user.UpdatedAt,
		user.PingFrequency, user.PingDeadline, user.PingingEnabled, user.PingMethod, user.NextScheduledPing,
		user.TOTPSecret, user.TOTPEnabled, user.TOTPVerified,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *SQLiteRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT
			id, email, password_hash, telegram_id, telegram_username, github_username,
			last_activity, created_at, updated_at,
			ping_frequency, ping_deadline, pinging_enabled, ping_method, next_scheduled_ping,
			totp_secret, totp_enabled, totp_verified
		FROM users
		WHERE id = ?
	`, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.TelegramID, &user.TelegramUsername, &user.GitHubUsername,
		&user.LastActivity, &user.CreatedAt, &user.UpdatedAt,
		&user.PingFrequency, &user.PingDeadline, &user.PingingEnabled, &user.PingMethod, &user.NextScheduledPing,
		&user.TOTPSecret, &user.TOTPEnabled, &user.TOTPVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *SQLiteRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT
			id, email, password_hash, telegram_id, telegram_username, github_username,
			last_activity, created_at, updated_at,
			ping_frequency, ping_deadline, pinging_enabled, ping_method, next_scheduled_ping,
			totp_secret, totp_enabled, totp_verified
		FROM users
		WHERE email = ?
	`, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.TelegramID, &user.TelegramUsername, &user.GitHubUsername,
		&user.LastActivity, &user.CreatedAt, &user.UpdatedAt,
		&user.PingFrequency, &user.PingDeadline, &user.PingingEnabled, &user.PingMethod, &user.NextScheduledPing,
		&user.TOTPSecret, &user.TOTPEnabled, &user.TOTPVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUserByTelegramID retrieves a user by Telegram ID
func (r *SQLiteRepository) GetUserByTelegramID(ctx context.Context, telegramID string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT
			id, email, password_hash, telegram_id, telegram_username, github_username,
			last_activity, created_at, updated_at,
			ping_frequency, ping_deadline, pinging_enabled, ping_method, next_scheduled_ping,
			totp_secret, totp_enabled, totp_verified
		FROM users
		WHERE telegram_id = ?
	`, telegramID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.TelegramID, &user.TelegramUsername, &user.GitHubUsername,
		&user.LastActivity, &user.CreatedAt, &user.UpdatedAt,
		&user.PingFrequency, &user.PingDeadline, &user.PingingEnabled, &user.PingMethod, &user.NextScheduledPing,
		&user.TOTPSecret, &user.TOTPEnabled, &user.TOTPVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (r *SQLiteRepository) UpdateUser(ctx context.Context, user *models.User) error {
	// Update timestamp
	user.UpdatedAt = time.Now().UTC()
	log.Printf("SQLite: Updating user ID %s with GitHub username: '%s'", user.ID, user.GitHubUsername)

	// Execute update query
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET
			email = ?,
			password_hash = ?,
			telegram_id = ?,
			telegram_username = ?,
			github_username = ?,
			last_activity = ?,
			updated_at = ?,
			ping_frequency = ?,
			ping_deadline = ?,
			pinging_enabled = ?,
			ping_method = ?,
			next_scheduled_ping = ?,
			totp_secret = ?,
			totp_enabled = ?,
			totp_verified = ?
		WHERE id = ?
	`,
		user.Email, user.PasswordHash, user.TelegramID, user.TelegramUsername, user.GitHubUsername,
		user.LastActivity, user.UpdatedAt,
		user.PingFrequency, user.PingDeadline, user.PingingEnabled, user.PingMethod, user.NextScheduledPing,
		user.TOTPSecret, user.TOTPEnabled, user.TOTPVerified,
		user.ID,
	)

	if err != nil {
		log.Printf("SQLite: Error updating user: %v", err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("SQLite: Successfully updated user ID %s", user.ID)
	return nil
}

// DeleteUser deletes a user
func (r *SQLiteRepository) DeleteUser(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers returns all users
func (r *SQLiteRepository) ListUsers(ctx context.Context) ([]*models.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id, email, password_hash, telegram_id, telegram_username, github_username,
			last_activity, created_at, updated_at,
			ping_frequency, ping_deadline, pinging_enabled, ping_method, next_scheduled_ping,
			totp_secret, totp_enabled, totp_verified
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.TelegramID, &user.TelegramUsername, &user.GitHubUsername,
			&user.LastActivity, &user.CreatedAt, &user.UpdatedAt,
			&user.PingFrequency, &user.PingDeadline, &user.PingingEnabled, &user.PingMethod, &user.NextScheduledPing,
			&user.TOTPSecret, &user.TOTPEnabled, &user.TOTPVerified,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

// ===== Secret operations =====

// CreateSecret creates a new secret
func (r *SQLiteRepository) CreateSecret(ctx context.Context, secret *models.Secret) error {
	if secret.ID == "" {
		secret.ID = generateID()
	}

	now := time.Now().UTC()
	secret.CreatedAt = now
	secret.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO secrets (
			id, user_id, name, encrypted_data, created_at, updated_at, encryption_type
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		secret.ID, secret.UserID, secret.Name, secret.EncryptedData,
		secret.CreatedAt, secret.UpdatedAt, secret.EncryptionType,
	)

	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// GetSecretByID retrieves a secret by ID
func (r *SQLiteRepository) GetSecretByID(ctx context.Context, id string) (*models.Secret, error) {
	secret := &models.Secret{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, encrypted_data, created_at, updated_at, encryption_type
		FROM secrets
		WHERE id = ?
	`, id).Scan(
		&secret.ID, &secret.UserID, &secret.Name, &secret.EncryptedData,
		&secret.CreatedAt, &secret.UpdatedAt, &secret.EncryptionType,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return secret, nil
}

// ListSecretsByUserID lists all secrets for a user
func (r *SQLiteRepository) ListSecretsByUserID(ctx context.Context, userID string) ([]*models.Secret, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, encrypted_data, created_at, updated_at, encryption_type
		FROM secrets
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	defer rows.Close()

	var secrets []*models.Secret
	for rows.Next() {
		secret := &models.Secret{}
		if err := rows.Scan(
			&secret.ID, &secret.UserID, &secret.Name, &secret.EncryptedData,
			&secret.CreatedAt, &secret.UpdatedAt, &secret.EncryptionType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan secret row: %w", err)
		}
		secrets = append(secrets, secret)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating secret rows: %w", err)
	}

	return secrets, nil
}

// UpdateSecret updates an existing secret
func (r *SQLiteRepository) UpdateSecret(ctx context.Context, secret *models.Secret) error {
	secret.UpdatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		UPDATE secrets SET
			name = ?,
			encrypted_data = ?,
			updated_at = ?,
			encryption_type = ?
		WHERE id = ? AND user_id = ?
	`,
		secret.Name, secret.EncryptedData, secret.UpdatedAt, secret.EncryptionType,
		secret.ID, secret.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	return nil
}

// DeleteSecret deletes a secret
func (r *SQLiteRepository) DeleteSecret(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM secrets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// ===== Recipient operations =====

// CreateRecipient creates a new recipient
func (r *SQLiteRepository) CreateRecipient(ctx context.Context, recipient *models.Recipient) error {
	if recipient.ID == "" {
		recipient.ID = generateID()
	}

	now := time.Now().UTC()
	recipient.CreatedAt = now
	recipient.UpdatedAt = now

	// Default values for confirmation fields
	recipient.IsConfirmed = false
	recipient.ConfirmedAt = nil
	recipient.ConfirmationCode = ""
	recipient.ConfirmationSentAt = nil

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO recipients (
			id, user_id, email, name, message, created_at, updated_at, phone_number,
			is_confirmed, confirmed_at, confirmation_code, confirmation_sent_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		recipient.ID, recipient.UserID, recipient.Email, recipient.Name,
		recipient.Message, recipient.CreatedAt, recipient.UpdatedAt, recipient.PhoneNumber,
		recipient.IsConfirmed, recipient.ConfirmedAt, recipient.ConfirmationCode, recipient.ConfirmationSentAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create recipient: %w", err)
	}

	return nil
}

// GetRecipientByID retrieves a recipient by ID
func (r *SQLiteRepository) GetRecipientByID(ctx context.Context, id string) (*models.Recipient, error) {
	recipient := &models.Recipient{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, email, name, message, created_at, updated_at, phone_number,
		       is_confirmed, confirmed_at, confirmation_code, confirmation_sent_at
		FROM recipients
		WHERE id = ?
	`, id).Scan(
		&recipient.ID, &recipient.UserID, &recipient.Email, &recipient.Name,
		&recipient.Message, &recipient.CreatedAt, &recipient.UpdatedAt, &recipient.PhoneNumber,
		&recipient.IsConfirmed, &recipient.ConfirmedAt, &recipient.ConfirmationCode, &recipient.ConfirmationSentAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get recipient: %w", err)
	}

	return recipient, nil
}

// ListRecipientsByUserID lists all recipients for a user
func (r *SQLiteRepository) ListRecipientsByUserID(ctx context.Context, userID string) ([]*models.Recipient, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, email, name, message, created_at, updated_at, phone_number,
		       is_confirmed, confirmed_at, confirmation_code, confirmation_sent_at
		FROM recipients
		WHERE user_id = ?
		ORDER BY name ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list recipients: %w", err)
	}
	defer rows.Close()

	var recipients []*models.Recipient
	for rows.Next() {
		recipient := &models.Recipient{}
		if err := rows.Scan(
			&recipient.ID, &recipient.UserID, &recipient.Email, &recipient.Name,
			&recipient.Message, &recipient.CreatedAt, &recipient.UpdatedAt, &recipient.PhoneNumber,
			&recipient.IsConfirmed, &recipient.ConfirmedAt, &recipient.ConfirmationCode, &recipient.ConfirmationSentAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recipient row: %w", err)
		}
		recipients = append(recipients, recipient)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recipient rows: %w", err)
	}

	return recipients, nil
}

// UpdateRecipient updates an existing recipient
func (r *SQLiteRepository) UpdateRecipient(ctx context.Context, recipient *models.Recipient) error {
	recipient.UpdatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		UPDATE recipients SET
			email = ?,
			name = ?,
			message = ?,
			updated_at = ?,
			phone_number = ?,
			is_confirmed = ?,
			confirmed_at = ?,
			confirmation_code = ?,
			confirmation_sent_at = ?
		WHERE id = ? AND user_id = ?
	`,
		recipient.Email, recipient.Name, recipient.Message,
		recipient.UpdatedAt, recipient.PhoneNumber,
		recipient.IsConfirmed, recipient.ConfirmedAt, recipient.ConfirmationCode, recipient.ConfirmationSentAt,
		recipient.ID, recipient.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update recipient: %w", err)
	}

	return nil
}

// DeleteRecipient deletes a recipient
func (r *SQLiteRepository) DeleteRecipient(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM recipients WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete recipient: %w", err)
	}
	return nil
}

// ===== SecretAssignment operations =====

// CreateSecretAssignment creates a new secret assignment
func (r *SQLiteRepository) CreateSecretAssignment(ctx context.Context, assignment *models.SecretAssignment) error {
	if assignment.ID == "" {
		assignment.ID = generateID()
	}

	now := time.Now().UTC()
	assignment.CreatedAt = now
	assignment.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO secret_assignments (
			id, secret_id, recipient_id, user_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		assignment.ID, assignment.SecretID, assignment.RecipientID,
		assignment.UserID, assignment.CreatedAt, assignment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create secret assignment: %w", err)
	}

	return nil
}

// GetSecretAssignmentByID retrieves a secret assignment by ID
func (r *SQLiteRepository) GetSecretAssignmentByID(ctx context.Context, id string) (*models.SecretAssignment, error) {
	assignment := &models.SecretAssignment{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, secret_id, recipient_id, user_id, created_at, updated_at
		FROM secret_assignments
		WHERE id = ?
	`, id).Scan(
		&assignment.ID, &assignment.SecretID, &assignment.RecipientID,
		&assignment.UserID, &assignment.CreatedAt, &assignment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get secret assignment: %w", err)
	}

	return assignment, nil
}

// ListSecretAssignmentsBySecretID lists all assignments for a secret
func (r *SQLiteRepository) ListSecretAssignmentsBySecretID(ctx context.Context, secretID string) ([]*models.SecretAssignment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, secret_id, recipient_id, user_id, created_at, updated_at
		FROM secret_assignments
		WHERE secret_id = ?
	`, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secret assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*models.SecretAssignment
	for rows.Next() {
		assignment := &models.SecretAssignment{}
		if err := rows.Scan(
			&assignment.ID, &assignment.SecretID, &assignment.RecipientID,
			&assignment.UserID, &assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan secret assignment row: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// ListSecretAssignmentsByRecipientID lists all assignments for a recipient
func (r *SQLiteRepository) ListSecretAssignmentsByRecipientID(ctx context.Context, recipientID string) ([]*models.SecretAssignment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, secret_id, recipient_id, user_id, created_at, updated_at
		FROM secret_assignments
		WHERE recipient_id = ?
	`, recipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secret assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*models.SecretAssignment
	for rows.Next() {
		assignment := &models.SecretAssignment{}
		if err := rows.Scan(
			&assignment.ID, &assignment.SecretID, &assignment.RecipientID,
			&assignment.UserID, &assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan secret assignment row: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// ListSecretAssignmentsByUserID lists all assignments for a user
func (r *SQLiteRepository) ListSecretAssignmentsByUserID(ctx context.Context, userID string) ([]*models.SecretAssignment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, secret_id, recipient_id, user_id, created_at, updated_at
		FROM secret_assignments
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secret assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*models.SecretAssignment
	for rows.Next() {
		assignment := &models.SecretAssignment{}
		if err := rows.Scan(
			&assignment.ID, &assignment.SecretID, &assignment.RecipientID,
			&assignment.UserID, &assignment.CreatedAt, &assignment.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan secret assignment row: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// DeleteSecretAssignment deletes a secret assignment
func (r *SQLiteRepository) DeleteSecretAssignment(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM secret_assignments WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete secret assignment: %w", err)
	}
	return nil
}

// ===== Ping operations =====

// CreatePingHistory creates a new ping history entry
func (r *SQLiteRepository) CreatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	if ping.ID == "" {
		ping.ID = generateID()
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ping_history (
			id, user_id, sent_at, method, status, responded_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		ping.ID, ping.UserID, ping.SentAt, ping.Method, ping.Status, ping.RespondedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create ping history: %w", err)
	}

	return nil
}

// UpdatePingHistory updates an existing ping history entry
func (r *SQLiteRepository) UpdatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE ping_history SET
			status = ?,
			responded_at = ?
		WHERE id = ?
	`,
		ping.Status, ping.RespondedAt, ping.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update ping history: %w", err)
	}

	return nil
}

// GetLatestPingByUserID retrieves the latest ping for a user
func (r *SQLiteRepository) GetLatestPingByUserID(ctx context.Context, userID string) (*models.PingHistory, error) {
	ping := &models.PingHistory{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, sent_at, method, status, responded_at
		FROM ping_history
		WHERE user_id = ?
		ORDER BY sent_at DESC
		LIMIT 1
	`, userID).Scan(
		&ping.ID, &ping.UserID, &ping.SentAt, &ping.Method, &ping.Status, &ping.RespondedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get latest ping: %w", err)
	}

	return ping, nil
}

// ListPingHistoryByUserID lists all pings for a user
func (r *SQLiteRepository) ListPingHistoryByUserID(ctx context.Context, userID string) ([]*models.PingHistory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, sent_at, method, status, responded_at
		FROM ping_history
		WHERE user_id = ?
		ORDER BY sent_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list ping history: %w", err)
	}
	defer rows.Close()

	var pings []*models.PingHistory
	for rows.Next() {
		ping := &models.PingHistory{}
		if err := rows.Scan(
			&ping.ID, &ping.UserID, &ping.SentAt, &ping.Method, &ping.Status, &ping.RespondedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan ping history row: %w", err)
		}
		pings = append(pings, ping)
	}

	return pings, nil
}

// ===== Ping verification operations =====

// CreatePingVerification creates a new ping verification
func (r *SQLiteRepository) CreatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	if verification.ID == "" {
		verification.ID = generateID()
	}

	verification.CreatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ping_verification (
			id, user_id, code, expires_at, used, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		verification.ID, verification.UserID, verification.Code,
		verification.ExpiresAt, verification.Used, verification.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create ping verification: %w", err)
	}

	return nil
}

// GetPingVerificationByCode retrieves a ping verification by code
func (r *SQLiteRepository) GetPingVerificationByCode(ctx context.Context, code string) (*models.PingVerification, error) {
	verification := &models.PingVerification{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, code, expires_at, used, created_at
		FROM ping_verification
		WHERE code = ?
	`, code).Scan(
		&verification.ID, &verification.UserID, &verification.Code,
		&verification.ExpiresAt, &verification.Used, &verification.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get ping verification: %w", err)
	}

	return verification, nil
}

// UpdatePingVerification updates an existing ping verification
func (r *SQLiteRepository) UpdatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE ping_verification SET
			used = ?
		WHERE id = ?
	`,
		verification.Used, verification.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update ping verification: %w", err)
	}

	return nil
}

// ===== DeliveryEvent operations =====

// CreateDeliveryEvent creates a new delivery event
func (r *SQLiteRepository) CreateDeliveryEvent(ctx context.Context, event *models.DeliveryEvent) error {
	if event.ID == "" {
		event.ID = generateID()
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO delivery_events (
			id, user_id, recipient_id, sent_at, status, error_message
		) VALUES (?, ?, ?, ?, ?, ?)
	`,
		event.ID, event.UserID, event.RecipientID,
		event.SentAt, event.Status, event.ErrorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to create delivery event: %w", err)
	}

	return nil
}

// ListDeliveryEventsByUserID lists all delivery events for a user
func (r *SQLiteRepository) ListDeliveryEventsByUserID(ctx context.Context, userID string) ([]*models.DeliveryEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, recipient_id, sent_at, status, error_message
		FROM delivery_events
		WHERE user_id = ?
		ORDER BY sent_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list delivery events: %w", err)
	}
	defer rows.Close()

	var events []*models.DeliveryEvent
	for rows.Next() {
		event := &models.DeliveryEvent{}
		if err := rows.Scan(
			&event.ID, &event.UserID, &event.RecipientID,
			&event.SentAt, &event.Status, &event.ErrorMessage,
		); err != nil {
			return nil, fmt.Errorf("failed to scan delivery event row: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// ===== Audit log operations =====

// CreateAuditLog creates a new audit log entry
func (r *SQLiteRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	if log.ID == "" {
		log.ID = generateID()
	}

	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now().UTC()
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_log (
			id, user_id, action, timestamp, ip_address, user_agent, details
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		log.ID, log.UserID, log.Action, log.Timestamp,
		log.IPAddress, log.UserAgent, log.Details,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// ListAuditLogsByUserID lists all audit logs for a user
func (r *SQLiteRepository) ListAuditLogsByUserID(ctx context.Context, userID string) ([]*models.AuditLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, action, timestamp, ip_address, user_agent, details
		FROM audit_log
		WHERE user_id = ?
		ORDER BY timestamp DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		log := &models.AuditLog{}
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.Timestamp,
			&log.IPAddress, &log.UserAgent, &log.Details,
		); err != nil {
			return nil, fmt.Errorf("failed to scan audit log row: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// ===== Session operations =====

// CreateSession creates a new session
func (r *SQLiteRepository) CreateSession(ctx context.Context, session *models.Session) error {
	if session.ID == "" {
		session.ID = generateID()
	}

	now := time.Now().UTC()
	session.CreatedAt = now
	session.LastActivity = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (
			id, user_id, token, created_at, expires_at, last_activity, ip_address, user_agent
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		session.ID, session.UserID, session.Token,
		session.CreatedAt, session.ExpiresAt, session.LastActivity,
		session.IPAddress, session.UserAgent,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByToken retrieves a session by token
func (r *SQLiteRepository) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	session := &models.Session{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, created_at, expires_at, last_activity, ip_address, user_agent
		FROM sessions
		WHERE token = ? AND expires_at > ?
	`, token, time.Now().UTC()).Scan(
		&session.ID, &session.UserID, &session.Token,
		&session.CreatedAt, &session.ExpiresAt, &session.LastActivity,
		&session.IPAddress, &session.UserAgent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// DeleteSession deletes a session
func (r *SQLiteRepository) DeleteSession(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions deletes all expired sessions
func (r *SQLiteRepository) DeleteExpiredSessions(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at <= ?", time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

// UpdateSessionActivity updates a session's last activity time
func (r *SQLiteRepository) UpdateSessionActivity(ctx context.Context, id string) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `
		UPDATE sessions SET
			last_activity = ?
		WHERE id = ?
	`, now, id)

	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	return nil
}

// ===== Scheduler operations =====

// GetUsersForPinging retrieves all users who need to be pinged
func (r *SQLiteRepository) GetUsersForPinging(ctx context.Context) ([]*models.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id, email, password_hash, telegram_id, telegram_username, github_username,
			last_activity, created_at, updated_at,
			ping_frequency, ping_deadline, pinging_enabled, ping_method, next_scheduled_ping
		FROM users
		WHERE pinging_enabled = 1 AND (next_scheduled_ping IS NULL OR next_scheduled_ping <= ?)
		ORDER BY next_scheduled_ping ASC
	`, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to get users for pinging: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.TelegramID, &user.TelegramUsername, &user.GitHubUsername,
			&user.LastActivity, &user.CreatedAt, &user.UpdatedAt,
			&user.PingFrequency, &user.PingDeadline, &user.PingingEnabled, &user.PingMethod, &user.NextScheduledPing,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUsersWithExpiredPings retrieves all users who have not responded to pings and exceeded the deadline
func (r *SQLiteRepository) GetUsersWithExpiredPings(ctx context.Context) ([]*models.User, error) {
	now := time.Now().UTC()
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			u.id, u.email, u.password_hash, u.telegram_id, u.telegram_username, u.github_username,
			u.last_activity, u.created_at, u.updated_at,
			u.ping_frequency, u.ping_deadline, u.pinging_enabled, u.ping_method, u.next_scheduled_ping
		FROM users u
		WHERE u.pinging_enabled = 1
		AND (
			-- User has not been active in ping_deadline days
			u.last_activity < ?
		)
		ORDER BY u.last_activity ASC
	`, now.Add(-24*time.Hour*30)) // 30 days as a maximum deadline
	if err != nil {
		return nil, fmt.Errorf("failed to get users with expired pings: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash, &user.TelegramID, &user.TelegramUsername, &user.GitHubUsername,
			&user.LastActivity, &user.CreatedAt, &user.UpdatedAt,
			&user.PingFrequency, &user.PingDeadline, &user.PingingEnabled, &user.PingMethod, &user.NextScheduledPing,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		// Further filter users who have exceeded their specific deadline
		deadline := user.LastActivity.Add(time.Duration(user.PingDeadline) * 24 * time.Hour)
		if now.After(deadline) {
			users = append(users, user)
		}
	}

	return users, nil
}
