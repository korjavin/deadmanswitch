package storage

import (
	"context"
	"errors"

	"github.com/korjavin/deadmanswitch/internal/models"
)

var (
	// ErrNotFound is returned when an item is not found
	ErrNotFound = errors.New("item not found")

	// ErrAlreadyExists is returned when trying to create an item that already exists
	ErrAlreadyExists = errors.New("item already exists")

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Repository defines the interface for database operations
type Repository interface {
	// User operations
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context) ([]*models.User, error)

	// Secret operations
	CreateSecret(ctx context.Context, secret *models.Secret) error
	GetSecretByID(ctx context.Context, id string) (*models.Secret, error)
	ListSecretsByUserID(ctx context.Context, userID string) ([]*models.Secret, error)
	UpdateSecret(ctx context.Context, secret *models.Secret) error
	DeleteSecret(ctx context.Context, id string) error

	// Recipient operations
	CreateRecipient(ctx context.Context, recipient *models.Recipient) error
	GetRecipientByID(ctx context.Context, id string) (*models.Recipient, error)
	ListRecipientsByUserID(ctx context.Context, userID string) ([]*models.Recipient, error)
	UpdateRecipient(ctx context.Context, recipient *models.Recipient) error
	DeleteRecipient(ctx context.Context, id string) error

	// SecretAssignment operations
	CreateSecretAssignment(ctx context.Context, assignment *models.SecretAssignment) error
	GetSecretAssignmentByID(ctx context.Context, id string) (*models.SecretAssignment, error)
	ListSecretAssignmentsBySecretID(ctx context.Context, secretID string) ([]*models.SecretAssignment, error)
	ListSecretAssignmentsByRecipientID(ctx context.Context, recipientID string) ([]*models.SecretAssignment, error)
	ListSecretAssignmentsByUserID(ctx context.Context, userID string) ([]*models.SecretAssignment, error)
	DeleteSecretAssignment(ctx context.Context, id string) error

	// SecretQuestion operations
	CreateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error
	GetSecretQuestion(ctx context.Context, id string) (*models.SecretQuestion, error)
	UpdateSecretQuestion(ctx context.Context, question *models.SecretQuestion) error
	DeleteSecretQuestion(ctx context.Context, id string) error
	ListSecretQuestionsByAssignmentID(ctx context.Context, assignmentID string) ([]*models.SecretQuestion, error)

	// SecretQuestionSet operations
	CreateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error
	GetSecretQuestionSet(ctx context.Context, id string) (*models.SecretQuestionSet, error)
	GetSecretQuestionSetByAssignmentID(ctx context.Context, assignmentID string) (*models.SecretQuestionSet, error)
	UpdateSecretQuestionSet(ctx context.Context, set *models.SecretQuestionSet) error
	DeleteSecretQuestionSet(ctx context.Context, id string) error
	ListSecretQuestionSetsNeedingReencryption(ctx context.Context, safeMarginSeconds int64) ([]*models.SecretQuestionSet, error)

	// Ping operations
	CreatePingHistory(ctx context.Context, ping *models.PingHistory) error
	UpdatePingHistory(ctx context.Context, ping *models.PingHistory) error
	GetLatestPingByUserID(ctx context.Context, userID string) (*models.PingHistory, error)
	ListPingHistoryByUserID(ctx context.Context, userID string) ([]*models.PingHistory, error)

	// Ping verification operations
	CreatePingVerification(ctx context.Context, verification *models.PingVerification) error
	GetPingVerificationByCode(ctx context.Context, code string) (*models.PingVerification, error)
	UpdatePingVerification(ctx context.Context, verification *models.PingVerification) error

	// DeliveryEvent operations
	CreateDeliveryEvent(ctx context.Context, event *models.DeliveryEvent) error
	ListDeliveryEventsByUserID(ctx context.Context, userID string) ([]*models.DeliveryEvent, error)

	// Audit log operations
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	ListAuditLogsByUserID(ctx context.Context, userID string) ([]*models.AuditLog, error)

	// Passkey operations
	CreatePasskey(ctx context.Context, passkey *models.Passkey) error
	GetPasskeyByID(ctx context.Context, id string) (*models.Passkey, error)
	GetPasskeyByCredentialID(ctx context.Context, credentialID []byte) (*models.Passkey, error)
	ListPasskeysByUserID(ctx context.Context, userID string) ([]*models.Passkey, error)
	ListPasskeys(ctx context.Context) ([]*models.Passkey, error)
	UpdatePasskey(ctx context.Context, passkey *models.Passkey) error
	DeletePasskey(ctx context.Context, id string) error
	DeletePasskeysByUserID(ctx context.Context, userID string) error

	// Session operations
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByToken(ctx context.Context, token string) (*models.Session, error)
	DeleteSession(ctx context.Context, id string) error
	DeleteExpiredSessions(ctx context.Context) error
	UpdateSessionActivity(ctx context.Context, id string) error

	// Scheduler operations
	GetUsersForPinging(ctx context.Context) ([]*models.User, error)
	GetUsersWithExpiredPings(ctx context.Context) ([]*models.User, error)

	// Transaction support
	BeginTx(ctx context.Context) (Transaction, error)
}

// Transaction represents a database transaction
type Transaction interface {
	Repository
	Commit() error
	Rollback() error
}

// NewRepository creates a new storage repository
func NewRepository(dbPath string) (Repository, error) {
	// For now, we'll implement a SQLite repository
	return NewSQLiteRepository(dbPath)
}
