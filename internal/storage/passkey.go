package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// CreatePasskey creates a new passkey
func (r *SQLiteRepository) CreatePasskey(ctx context.Context, passkey *models.Passkey) error {
	if passkey.ID == "" {
		passkey.ID = generateID()
	}

	now := time.Now().UTC()
	passkey.CreatedAt = now
	passkey.LastUsedAt = now

	// Convert transports slice to JSON string
	transportsJSON, err := json.Marshal(passkey.Transports)
	if err != nil {
		return fmt.Errorf("failed to marshal transports: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO passkeys (
			id, user_id, credential_id, public_key, aaguid, sign_count,
			name, created_at, last_used_at, transports, attestation_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		passkey.ID, passkey.UserID, passkey.CredentialID, passkey.PublicKey,
		passkey.AAGUID, passkey.SignCount, passkey.Name, passkey.CreatedAt,
		passkey.LastUsedAt, string(transportsJSON), passkey.AttestationType,
	)

	if err != nil {
		return fmt.Errorf("failed to create passkey: %w", err)
	}

	return nil
}

// GetPasskeyByID retrieves a passkey by ID
func (r *SQLiteRepository) GetPasskeyByID(ctx context.Context, id string) (*models.Passkey, error) {
	passkey := &models.Passkey{}
	var transportsJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT
			id, user_id, credential_id, public_key, aaguid, sign_count,
			name, created_at, last_used_at, transports, attestation_type
		FROM passkeys
		WHERE id = ?
	`, id).Scan(
		&passkey.ID, &passkey.UserID, &passkey.CredentialID, &passkey.PublicKey,
		&passkey.AAGUID, &passkey.SignCount, &passkey.Name, &passkey.CreatedAt,
		&passkey.LastUsedAt, &transportsJSON, &passkey.AttestationType,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get passkey: %w", err)
	}

	// Parse transports JSON
	if transportsJSON != "" {
		if err := json.Unmarshal([]byte(transportsJSON), &passkey.Transports); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transports: %w", err)
		}
	}

	return passkey, nil
}

// GetPasskeyByCredentialID retrieves a passkey by credential ID
func (r *SQLiteRepository) GetPasskeyByCredentialID(ctx context.Context, credentialID []byte) (*models.Passkey, error) {
	passkey := &models.Passkey{}
	var transportsJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT
			id, user_id, credential_id, public_key, aaguid, sign_count,
			name, created_at, last_used_at, transports, attestation_type
		FROM passkeys
		WHERE credential_id = ?
	`, credentialID).Scan(
		&passkey.ID, &passkey.UserID, &passkey.CredentialID, &passkey.PublicKey,
		&passkey.AAGUID, &passkey.SignCount, &passkey.Name, &passkey.CreatedAt,
		&passkey.LastUsedAt, &transportsJSON, &passkey.AttestationType,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get passkey by credential ID: %w", err)
	}

	// Parse transports JSON
	if transportsJSON != "" {
		if err := json.Unmarshal([]byte(transportsJSON), &passkey.Transports); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transports: %w", err)
		}
	}

	return passkey, nil
}

// ListPasskeysByUserID lists all passkeys for a user
func (r *SQLiteRepository) ListPasskeysByUserID(ctx context.Context, userID string) ([]*models.Passkey, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id, user_id, credential_id, public_key, aaguid, sign_count,
			name, created_at, last_used_at, transports, attestation_type
		FROM passkeys
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list passkeys: %w", err)
	}
	defer rows.Close()

	var passkeys []*models.Passkey
	for rows.Next() {
		passkey := &models.Passkey{}
		var transportsJSON string

		if err := rows.Scan(
			&passkey.ID, &passkey.UserID, &passkey.CredentialID, &passkey.PublicKey,
			&passkey.AAGUID, &passkey.SignCount, &passkey.Name, &passkey.CreatedAt,
			&passkey.LastUsedAt, &transportsJSON, &passkey.AttestationType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan passkey row: %w", err)
		}

		// Parse transports JSON
		if transportsJSON != "" {
			if err := json.Unmarshal([]byte(transportsJSON), &passkey.Transports); err != nil {
				return nil, fmt.Errorf("failed to unmarshal transports: %w", err)
			}
		}

		passkeys = append(passkeys, passkey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating passkey rows: %w", err)
	}

	return passkeys, nil
}

// UpdatePasskey updates an existing passkey
func (r *SQLiteRepository) UpdatePasskey(ctx context.Context, passkey *models.Passkey) error {
	// Convert transports slice to JSON string
	transportsJSON, err := json.Marshal(passkey.Transports)
	if err != nil {
		return fmt.Errorf("failed to marshal transports: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE passkeys SET
			public_key = ?,
			aaguid = ?,
			sign_count = ?,
			name = ?,
			last_used_at = ?,
			transports = ?,
			attestation_type = ?
		WHERE id = ?
	`,
		passkey.PublicKey, passkey.AAGUID, passkey.SignCount, passkey.Name,
		passkey.LastUsedAt, string(transportsJSON), passkey.AttestationType,
		passkey.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update passkey: %w", err)
	}

	return nil
}

// DeletePasskey deletes a passkey
func (r *SQLiteRepository) DeletePasskey(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM passkeys WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete passkey: %w", err)
	}
	return nil
}

// DeletePasskeysByUserID deletes all passkeys for a user
func (r *SQLiteRepository) DeletePasskeysByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM passkeys WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user passkeys: %w", err)
	}
	return nil
}

// ListPasskeys lists all passkeys in the database
func (r *SQLiteRepository) ListPasskeys(ctx context.Context) ([]*models.Passkey, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id, user_id, credential_id, public_key, aaguid, sign_count,
			name, created_at, last_used_at, transports, attestation_type
		FROM passkeys
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list all passkeys: %w", err)
	}
	defer rows.Close()

	var passkeys []*models.Passkey
	for rows.Next() {
		passkey := &models.Passkey{}
		var transportsJSON string

		if err := rows.Scan(
			&passkey.ID, &passkey.UserID, &passkey.CredentialID, &passkey.PublicKey,
			&passkey.AAGUID, &passkey.SignCount, &passkey.Name, &passkey.CreatedAt,
			&passkey.LastUsedAt, &transportsJSON, &passkey.AttestationType,
		); err != nil {
			return nil, fmt.Errorf("failed to scan passkey row: %w", err)
		}

		// Parse transports JSON
		if transportsJSON != "" {
			if err := json.Unmarshal([]byte(transportsJSON), &passkey.Transports); err != nil {
				return nil, fmt.Errorf("failed to unmarshal transports: %w", err)
			}
		}

		passkeys = append(passkeys, passkey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating passkey rows: %w", err)
	}

	return passkeys, nil
}
