package storage

import (
	"context"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// Add fields to MockRepository struct
func init() {
	// This is just to ensure the compiler knows these fields exist
	// They will be properly initialized in NewMockRepository
	_ = &MockRepository{
		SecretQuestions:    make([]*models.SecretQuestion, 0),
		SecretQuestionSets: make([]*models.SecretQuestionSet, 0),
	}
}

// SecretQuestion methods for MockRepository
func (m *MockRepository) CreateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error {
	m.SecretQuestions = append(m.SecretQuestions, question)
	return nil
}

func (m *MockRepository) GetSecretQuestion(ctx context.Context, id string) (*models.SecretQuestion, error) {
	for _, q := range m.SecretQuestions {
		if q.ID == id {
			return q, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error {
	for i, q := range m.SecretQuestions {
		if q.ID == question.ID {
			m.SecretQuestions[i] = question
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) DeleteSecretQuestion(ctx context.Context, id string) error {
	for i, q := range m.SecretQuestions {
		if q.ID == id {
			m.SecretQuestions = append(m.SecretQuestions[:i], m.SecretQuestions[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) ListSecretQuestionsByAssignmentID(ctx context.Context, assignmentID string) ([]*models.SecretQuestion, error) {
	var result []*models.SecretQuestion
	for _, q := range m.SecretQuestions {
		if q.SecretAssignmentID == assignmentID {
			result = append(result, q)
		}
	}
	return result, nil
}

// SecretQuestionSet methods for MockRepository
func (m *MockRepository) CreateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error {
	m.SecretQuestionSets = append(m.SecretQuestionSets, set)
	return nil
}

func (m *MockRepository) GetSecretQuestionSet(ctx context.Context, id string) (*models.SecretQuestionSet, error) {
	for _, s := range m.SecretQuestionSets {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) GetSecretQuestionSetByAssignmentID(ctx context.Context, assignmentID string) (*models.SecretQuestionSet, error) {
	for _, s := range m.SecretQuestionSets {
		if s.SecretAssignmentID == assignmentID {
			return s, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error {
	for i, s := range m.SecretQuestionSets {
		if s.ID == set.ID {
			m.SecretQuestionSets[i] = set
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) DeleteSecretQuestionSet(ctx context.Context, id string) error {
	for i, s := range m.SecretQuestionSets {
		if s.ID == id {
			m.SecretQuestionSets = append(m.SecretQuestionSets[:i], m.SecretQuestionSets[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) ListSecretQuestionSetsNeedingReencryption(ctx context.Context, safeMarginSeconds int64) ([]*models.SecretQuestionSet, error) {
	// In a mock implementation, we'll just return all sets
	// In a real implementation, we would filter based on the timelock round
	return m.SecretQuestionSets, nil
}

// SecretQuestion methods for MockTransaction
func (t *MockTransaction) CreateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error {
	return t.repo.CreateSecretQuestion(ctx, question)
}

func (t *MockTransaction) GetSecretQuestion(ctx context.Context, id string) (*models.SecretQuestion, error) {
	return t.repo.GetSecretQuestion(ctx, id)
}

func (t *MockTransaction) UpdateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error {
	return t.repo.UpdateSecretQuestion(ctx, question)
}

func (t *MockTransaction) DeleteSecretQuestion(ctx context.Context, id string) error {
	return t.repo.DeleteSecretQuestion(ctx, id)
}

func (t *MockTransaction) ListSecretQuestionsByAssignmentID(ctx context.Context, assignmentID string) ([]*models.SecretQuestion, error) {
	return t.repo.ListSecretQuestionsByAssignmentID(ctx, assignmentID)
}

// SecretQuestionSet methods for MockTransaction
func (t *MockTransaction) CreateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error {
	return t.repo.CreateSecretQuestionSet(ctx, set)
}

func (t *MockTransaction) GetSecretQuestionSet(ctx context.Context, id string) (*models.SecretQuestionSet, error) {
	return t.repo.GetSecretQuestionSet(ctx, id)
}

func (t *MockTransaction) GetSecretQuestionSetByAssignmentID(ctx context.Context, assignmentID string) (*models.SecretQuestionSet, error) {
	return t.repo.GetSecretQuestionSetByAssignmentID(ctx, assignmentID)
}

func (t *MockTransaction) UpdateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error {
	return t.repo.UpdateSecretQuestionSet(ctx, set)
}

func (t *MockTransaction) DeleteSecretQuestionSet(ctx context.Context, id string) error {
	return t.repo.DeleteSecretQuestionSet(ctx, id)
}

func (t *MockTransaction) ListSecretQuestionSetsNeedingReencryption(ctx context.Context, safeMarginSeconds int64) ([]*models.SecretQuestionSet, error) {
	return t.repo.ListSecretQuestionSetsNeedingReencryption(ctx, safeMarginSeconds)
}
