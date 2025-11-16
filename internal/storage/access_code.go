package storage

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/korjavin/deadmanswitch/internal/crypto"
	"github.com/korjavin/deadmanswitch/internal/models"
)

// CreateAccessCode creates a new access code with hashed code
func (r *SQLiteRepository) CreateAccessCode(ctx context.Context, accessCode *models.AccessCode) error {
	if accessCode.ID == "" {
		accessCode.ID = generateID()
	}

	now := time.Now().UTC()
	accessCode.CreatedAt = now

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO access_codes (
			id, code, recipient_id, user_id, delivery_event_id,
			created_at, expires_at, used_at, attempt_count, max_attempts
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		accessCode.ID, accessCode.Code, accessCode.RecipientID, accessCode.UserID,
		accessCode.DeliveryEventID, accessCode.CreatedAt, accessCode.ExpiresAt,
		accessCode.UsedAt, accessCode.AttemptCount, accessCode.MaxAttempts,
	)

	if err != nil {
		return fmt.Errorf("failed to create access code: %w", err)
	}

	return nil
}

// GetAccessCodeByCode retrieves an access code by its hashed code
func (r *SQLiteRepository) GetAccessCodeByCode(ctx context.Context, code string) (*models.AccessCode, error) {
	// Hash the provided code to search for it
	hashedCode, err := crypto.HashPassword(code, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to hash code: %w", err)
	}
	hashedCodeStr := base64.StdEncoding.EncodeToString(hashedCode)

	accessCode := &models.AccessCode{}
	var usedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, `
		SELECT id, code, recipient_id, user_id, delivery_event_id,
			created_at, expires_at, used_at, attempt_count, max_attempts
		FROM access_codes
		WHERE code = ?
	`, hashedCodeStr).Scan(
		&accessCode.ID, &accessCode.Code, &accessCode.RecipientID, &accessCode.UserID,
		&accessCode.DeliveryEventID, &accessCode.CreatedAt, &accessCode.ExpiresAt,
		&usedAt, &accessCode.AttemptCount, &accessCode.MaxAttempts,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get access code: %w", err)
	}

	if usedAt.Valid {
		accessCode.UsedAt = &usedAt.Time
	}

	return accessCode, nil
}

// VerifyAccessCode verifies an access code and checks expiration and attempts
// It returns the access code if valid, or an error if invalid, expired, or max attempts exceeded
func (r *SQLiteRepository) VerifyAccessCode(ctx context.Context, code string) (*models.AccessCode, error) {
	// First, we need to find all access codes and verify each one since we store hashed codes
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, recipient_id, user_id, delivery_event_id,
			created_at, expires_at, used_at, attempt_count, max_attempts
		FROM access_codes
		WHERE used_at IS NULL
		AND expires_at > ?
	`, time.Now().UTC())

	if err != nil {
		return nil, fmt.Errorf("failed to query access codes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		accessCode := &models.AccessCode{}
		var usedAt sql.NullTime
		var codeStr string

		if err := rows.Scan(
			&accessCode.ID, &codeStr, &accessCode.RecipientID, &accessCode.UserID,
			&accessCode.DeliveryEventID, &accessCode.CreatedAt, &accessCode.ExpiresAt,
			&usedAt, &accessCode.AttemptCount, &accessCode.MaxAttempts,
		); err != nil {
			return nil, fmt.Errorf("failed to scan access code: %w", err)
		}

		if usedAt.Valid {
			accessCode.UsedAt = &usedAt.Time
		}

		// Decode the stored hash
		storedHash, err := base64.StdEncoding.DecodeString(codeStr)
		if err != nil {
			continue // Skip invalid entries
		}

		// Verify the code
		matches, err := crypto.VerifyPassword(code, storedHash)
		if err != nil {
			continue
		}

		if matches {
			accessCode.Code = codeStr

			// Check if already used
			if accessCode.UsedAt != nil {
				return nil, fmt.Errorf("access code already used")
			}

			// Check if expired
			if time.Now().UTC().After(accessCode.ExpiresAt) {
				return nil, fmt.Errorf("access code expired")
			}

			// Check if max attempts exceeded
			if accessCode.AttemptCount >= accessCode.MaxAttempts {
				return nil, fmt.Errorf("access code locked due to too many failed attempts")
			}

			return accessCode, nil
		}
	}

	return nil, ErrNotFound
}

// MarkAccessCodeAsUsed marks an access code as used
func (r *SQLiteRepository) MarkAccessCodeAsUsed(ctx context.Context, id string) error {
	now := time.Now().UTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE access_codes
		SET used_at = ?
		WHERE id = ?
	`, now, id)

	if err != nil {
		return fmt.Errorf("failed to mark access code as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// IncrementAccessCodeAttempts increments the failed attempt counter for an access code
func (r *SQLiteRepository) IncrementAccessCodeAttempts(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE access_codes
		SET attempt_count = attempt_count + 1
		WHERE id = ?
	`, id)

	if err != nil {
		return fmt.Errorf("failed to increment access code attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteExpiredAccessCodes deletes all expired access codes
func (r *SQLiteRepository) DeleteExpiredAccessCodes(ctx context.Context) error {
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		DELETE FROM access_codes
		WHERE expires_at < ?
	`, now)

	if err != nil {
		return fmt.Errorf("failed to delete expired access codes: %w", err)
	}

	return nil
}
