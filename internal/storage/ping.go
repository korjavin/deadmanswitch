package storage

import (
	"database/sql"
	"fmt"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// Constants for ping history columns and queries
const (
	pingHistoryColumnsSelect = `
		id, user_id, sent_at, method, status, responded_at
	`
	pingHistoryBaseQuery = `
		SELECT ` + pingHistoryColumnsSelect + `
		FROM ping_history
	`
)

// scanPingHistoryRow scans a single ping history row from a database query
func (r *SQLiteRepository) scanPingHistoryRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.PingHistory, error) {
	ping := &models.PingHistory{}
	err := scanner.Scan(
		&ping.ID, &ping.UserID, &ping.SentAt, &ping.Method, &ping.Status, &ping.RespondedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan ping history row: %w", err)
	}

	return ping, nil
}

// Constants for ping verification columns and queries
const (
	pingVerificationColumnsSelect = `
		id, user_id, code, expires_at, used, created_at
	`
	pingVerificationBaseQuery = `
		SELECT ` + pingVerificationColumnsSelect + `
		FROM ping_verification
	`
)

// scanPingVerificationRow scans a single ping verification row from a database query
func (r *SQLiteRepository) scanPingVerificationRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.PingVerification, error) {
	verification := &models.PingVerification{}
	err := scanner.Scan(
		&verification.ID, &verification.UserID, &verification.Code,
		&verification.ExpiresAt, &verification.Used, &verification.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan ping verification row: %w", err)
	}

	return verification, nil
}
