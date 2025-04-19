package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// Methods related to session management

// GetSessionByID retrieves a session by ID
func (r *SQLiteRepository) GetSessionByID(ctx context.Context, id string) (*models.Session, error) {
	row := r.db.QueryRowContext(ctx, sessionBaseQuery+" WHERE id = ?", id)
	return r.scanSessionRow(row)
}

// ListSessionsByUserID lists all sessions for a user
func (r *SQLiteRepository) ListSessionsByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	rows, err := r.db.QueryContext(ctx, sessionBaseQuery+" WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session, err := r.scanSessionRow(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	return sessions, nil
}

// DeleteSessionsByUserID deletes all sessions for a user
func (r *SQLiteRepository) DeleteSessionsByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}
	return nil
}

// ListActiveSessions lists all active sessions
func (r *SQLiteRepository) ListActiveSessions(ctx context.Context) ([]*models.Session, error) {
	rows, err := r.db.QueryContext(ctx, sessionBaseQuery+` 
		WHERE expires_at > ? 
		ORDER BY last_activity DESC
	`, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to list active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session, err := r.scanSessionRow(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session rows: %w", err)
	}

	return sessions, nil
}
