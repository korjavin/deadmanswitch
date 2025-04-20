package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/utils"
)

// CreateSecretQuestion creates a new secret question in the database
func (r *SQLiteRepository) CreateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error {
	if question.ID == "" {
		question.ID = utils.GenerateID()
	}

	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	query := `
		INSERT INTO secret_questions (
			id, secret_assignment_id, question, salt, encrypted_share, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		question.ID,
		question.SecretAssignmentID,
		question.Question,
		question.Salt,
		question.EncryptedShare,
		question.CreatedAt,
		question.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create secret question: %w", err)
	}

	return nil
}

// GetSecretQuestion retrieves a secret question by ID
func (r *SQLiteRepository) GetSecretQuestion(ctx context.Context, id string) (*models.SecretQuestion, error) {
	query := `
		SELECT id, secret_assignment_id, question, salt, encrypted_share, created_at, updated_at
		FROM secret_questions
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var question models.SecretQuestion
	err := row.Scan(
		&question.ID,
		&question.SecretAssignmentID,
		&question.Question,
		&question.Salt,
		&question.EncryptedShare,
		&question.CreatedAt,
		&question.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get secret question: %w", err)
	}

	return &question, nil
}

// UpdateSecretQuestion updates a secret question in the database
func (r *SQLiteRepository) UpdateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error {
	question.UpdatedAt = time.Now()

	query := `
		UPDATE secret_questions
		SET secret_assignment_id = ?, question = ?, salt = ?, encrypted_share = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		question.SecretAssignmentID,
		question.Question,
		question.Salt,
		question.EncryptedShare,
		question.UpdatedAt,
		question.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update secret question: %w", err)
	}

	return nil
}

// DeleteSecretQuestion deletes a secret question from the database
func (r *SQLiteRepository) DeleteSecretQuestion(ctx context.Context, id string) error {
	query := `DELETE FROM secret_questions WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete secret question: %w", err)
	}

	return nil
}

// ListSecretQuestionsByAssignmentID retrieves all secret questions for a specific assignment
func (r *SQLiteRepository) ListSecretQuestionsByAssignmentID(ctx context.Context, assignmentID string) ([]*models.SecretQuestion, error) {
	query := `
		SELECT id, secret_assignment_id, question, salt, encrypted_share, created_at, updated_at
		FROM secret_questions
		WHERE secret_assignment_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secret questions: %w", err)
	}
	defer rows.Close()

	var questions []*models.SecretQuestion
	for rows.Next() {
		var question models.SecretQuestion
		err := rows.Scan(
			&question.ID,
			&question.SecretAssignmentID,
			&question.Question,
			&question.Salt,
			&question.EncryptedShare,
			&question.CreatedAt,
			&question.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret question: %w", err)
		}
		questions = append(questions, &question)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating secret questions: %w", err)
	}

	return questions, nil
}

// CreateSecretQuestionSet creates a new secret question set in the database
func (r *SQLiteRepository) CreateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error {
	if set.ID == "" {
		set.ID = utils.GenerateID()
	}

	now := time.Now()
	set.CreatedAt = now
	set.UpdatedAt = now

	query := `
		INSERT INTO secret_question_sets (
			id, secret_assignment_id, threshold, total_questions, timelock_round, encrypted_blob, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		set.ID,
		set.SecretAssignmentID,
		set.Threshold,
		set.TotalQuestions,
		set.TimelockRound,
		set.EncryptedBlob,
		set.CreatedAt,
		set.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create secret question set: %w", err)
	}

	return nil
}

// GetSecretQuestionSet retrieves a secret question set by ID
func (r *SQLiteRepository) GetSecretQuestionSet(ctx context.Context, id string) (*models.SecretQuestionSet, error) {
	query := `
		SELECT id, secret_assignment_id, threshold, total_questions, timelock_round, encrypted_blob, created_at, updated_at
		FROM secret_question_sets
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var set models.SecretQuestionSet
	err := row.Scan(
		&set.ID,
		&set.SecretAssignmentID,
		&set.Threshold,
		&set.TotalQuestions,
		&set.TimelockRound,
		&set.EncryptedBlob,
		&set.CreatedAt,
		&set.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get secret question set: %w", err)
	}

	return &set, nil
}

// GetSecretQuestionSetByAssignmentID retrieves a secret question set by assignment ID
func (r *SQLiteRepository) GetSecretQuestionSetByAssignmentID(ctx context.Context, assignmentID string) (*models.SecretQuestionSet, error) {
	query := `
		SELECT id, secret_assignment_id, threshold, total_questions, timelock_round, encrypted_blob, created_at, updated_at
		FROM secret_question_sets
		WHERE secret_assignment_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, assignmentID)

	var set models.SecretQuestionSet
	err := row.Scan(
		&set.ID,
		&set.SecretAssignmentID,
		&set.Threshold,
		&set.TotalQuestions,
		&set.TimelockRound,
		&set.EncryptedBlob,
		&set.CreatedAt,
		&set.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get secret question set: %w", err)
	}

	return &set, nil
}

// UpdateSecretQuestionSet updates a secret question set in the database
func (r *SQLiteRepository) UpdateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error {
	set.UpdatedAt = time.Now()

	query := `
		UPDATE secret_question_sets
		SET threshold = ?, total_questions = ?, timelock_round = ?, encrypted_blob = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		set.Threshold,
		set.TotalQuestions,
		set.TimelockRound,
		set.EncryptedBlob,
		set.UpdatedAt,
		set.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update secret question set: %w", err)
	}

	return nil
}

// DeleteSecretQuestionSet deletes a secret question set from the database
func (r *SQLiteRepository) DeleteSecretQuestionSet(ctx context.Context, id string) error {
	query := `DELETE FROM secret_question_sets WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete secret question set: %w", err)
	}

	return nil
}

// ListSecretQuestionSetsNeedingReencryption retrieves all secret question sets that need re-encryption
// based on their timelock round being close to the current time
func (r *SQLiteRepository) ListSecretQuestionSetsNeedingReencryption(ctx context.Context, safeMarginSeconds int64) ([]*models.SecretQuestionSet, error) {
	// Calculate the drand round for the current time + safe margin
	currentTime := time.Now().Add(time.Duration(safeMarginSeconds) * time.Second)
	
	// This is a simplified query that assumes we can calculate the round directly in SQL
	// In a real implementation, we would need to calculate this in Go code
	query := `
		SELECT id, secret_assignment_id, threshold, total_questions, timelock_round, encrypted_blob, created_at, updated_at
		FROM secret_question_sets
		WHERE timelock_round <= ?
		ORDER BY timelock_round ASC
	`

	// For now, we'll pass the current time as a placeholder
	// In the handler, we'll need to filter based on the actual round calculation
	rows, err := r.db.QueryContext(ctx, query, currentTime.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to list secret question sets: %w", err)
	}
	defer rows.Close()

	var sets []*models.SecretQuestionSet
	for rows.Next() {
		var set models.SecretQuestionSet
		err := rows.Scan(
			&set.ID,
			&set.SecretAssignmentID,
			&set.Threshold,
			&set.TotalQuestions,
			&set.TimelockRound,
			&set.EncryptedBlob,
			&set.CreatedAt,
			&set.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret question set: %w", err)
		}
		sets = append(sets, &set)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating secret question sets: %w", err)
	}

	return sets, nil
}
