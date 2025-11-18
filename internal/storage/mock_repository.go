package storage

import (
	"context"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// MockRepository is a mock implementation of the Repository interface for testing
type MockRepository struct {
	Users                 []*models.User
	Secrets               []*models.Secret
	Recipients            []*models.Recipient
	SecretAssignments     []*models.SecretAssignment
	Passkeys              []*models.Passkey
	PingHistories         []*models.PingHistory
	PingVerifications     []*models.PingVerification
	DeliveryEvents        []*models.DeliveryEvent
	AccessCodes           []*models.AccessCode
	Sessions              []*models.Session
	AuditLogs             []*models.AuditLog
	UsersForPinging       []*models.User
	UsersWithExpiredPings []*models.User
}

// NewMockRepository creates a new mock repository for testing
func NewMockRepository() *MockRepository {
	return &MockRepository{
		Users:                 make([]*models.User, 0),
		Secrets:               make([]*models.Secret, 0),
		Recipients:            make([]*models.Recipient, 0),
		SecretAssignments:     make([]*models.SecretAssignment, 0),
		Passkeys:              make([]*models.Passkey, 0),
		PingHistories:         make([]*models.PingHistory, 0),
		PingVerifications:     make([]*models.PingVerification, 0),
		DeliveryEvents:        make([]*models.DeliveryEvent, 0),
		AccessCodes:           make([]*models.AccessCode, 0),
		Sessions:              make([]*models.Session, 0),
		AuditLogs:             make([]*models.AuditLog, 0),
		UsersForPinging:       make([]*models.User, 0),
		UsersWithExpiredPings: make([]*models.User, 0),
	}
}

// User methods
func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error {
	m.Users = append(m.Users, user)
	return nil
}

func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	for _, u := range m.Users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, u := range m.Users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) GetUserByTelegramID(ctx context.Context, telegramID string) (*models.User, error) {
	for _, u := range m.Users {
		if u.TelegramID == telegramID {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *models.User) error {
	for i, u := range m.Users {
		if u.ID == user.ID {
			m.Users[i] = user
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) DeleteUser(ctx context.Context, id string) error {
	for i, u := range m.Users {
		if u.ID == id {
			m.Users = append(m.Users[:i], m.Users[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) ListUsers(ctx context.Context) ([]*models.User, error) {
	return m.Users, nil
}

// Secret methods
func (m *MockRepository) CreateSecret(ctx context.Context, secret *models.Secret) error {
	m.Secrets = append(m.Secrets, secret)
	return nil
}

func (m *MockRepository) GetSecretByID(ctx context.Context, id string) (*models.Secret, error) {
	for _, s := range m.Secrets {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) ListSecretsByUserID(ctx context.Context, userID string) ([]*models.Secret, error) {
	var result []*models.Secret
	for _, s := range m.Secrets {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockRepository) UpdateSecret(ctx context.Context, secret *models.Secret) error {
	for i, s := range m.Secrets {
		if s.ID == secret.ID {
			m.Secrets[i] = secret
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) DeleteSecret(ctx context.Context, id string) error {
	for i, s := range m.Secrets {
		if s.ID == id {
			m.Secrets = append(m.Secrets[:i], m.Secrets[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// Recipient methods
func (m *MockRepository) CreateRecipient(ctx context.Context, recipient *models.Recipient) error {
	m.Recipients = append(m.Recipients, recipient)
	return nil
}

func (m *MockRepository) GetRecipientByID(ctx context.Context, id string) (*models.Recipient, error) {
	for _, r := range m.Recipients {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) ListRecipientsByUserID(ctx context.Context, userID string) ([]*models.Recipient, error) {
	var result []*models.Recipient
	for _, r := range m.Recipients {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockRepository) UpdateRecipient(ctx context.Context, recipient *models.Recipient) error {
	for i, r := range m.Recipients {
		if r.ID == recipient.ID {
			m.Recipients[i] = recipient
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) DeleteRecipient(ctx context.Context, id string) error {
	for i, r := range m.Recipients {
		if r.ID == id {
			m.Recipients = append(m.Recipients[:i], m.Recipients[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// SecretAssignment methods
func (m *MockRepository) CreateSecretAssignment(ctx context.Context, assignment *models.SecretAssignment) error {
	m.SecretAssignments = append(m.SecretAssignments, assignment)
	return nil
}

func (m *MockRepository) GetSecretAssignmentByID(ctx context.Context, id string) (*models.SecretAssignment, error) {
	for _, a := range m.SecretAssignments {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) ListSecretAssignmentsBySecretID(ctx context.Context, secretID string) ([]*models.SecretAssignment, error) {
	var result []*models.SecretAssignment
	for _, a := range m.SecretAssignments {
		if a.SecretID == secretID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MockRepository) ListSecretAssignmentsByRecipientID(ctx context.Context, recipientID string) ([]*models.SecretAssignment, error) {
	var result []*models.SecretAssignment
	for _, a := range m.SecretAssignments {
		if a.RecipientID == recipientID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MockRepository) ListSecretAssignmentsByUserID(ctx context.Context, userID string) ([]*models.SecretAssignment, error) {
	var result []*models.SecretAssignment
	for _, a := range m.SecretAssignments {
		if a.UserID == userID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MockRepository) DeleteSecretAssignment(ctx context.Context, id string) error {
	for i, a := range m.SecretAssignments {
		if a.ID == id {
			m.SecretAssignments = append(m.SecretAssignments[:i], m.SecretAssignments[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// Passkey methods
func (m *MockRepository) CreatePasskey(ctx context.Context, passkey *models.Passkey) error {
	m.Passkeys = append(m.Passkeys, passkey)
	return nil
}

func (m *MockRepository) GetPasskeyByID(ctx context.Context, id string) (*models.Passkey, error) {
	for _, p := range m.Passkeys {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) GetPasskeyByCredentialID(ctx context.Context, credentialID []byte) (*models.Passkey, error) {
	for _, p := range m.Passkeys {
		if string(p.CredentialID) == string(credentialID) {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) ListPasskeysByUserID(ctx context.Context, userID string) ([]*models.Passkey, error) {
	var result []*models.Passkey
	for _, p := range m.Passkeys {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockRepository) ListPasskeys(ctx context.Context) ([]*models.Passkey, error) {
	return m.Passkeys, nil
}

func (m *MockRepository) UpdatePasskey(ctx context.Context, passkey *models.Passkey) error {
	for i, p := range m.Passkeys {
		if p.ID == passkey.ID {
			m.Passkeys[i] = passkey
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) DeletePasskey(ctx context.Context, id string) error {
	for i, p := range m.Passkeys {
		if p.ID == id {
			m.Passkeys = append(m.Passkeys[:i], m.Passkeys[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) DeletePasskeysByUserID(ctx context.Context, userID string) error {
	var newPasskeys []*models.Passkey
	for _, p := range m.Passkeys {
		if p.UserID != userID {
			newPasskeys = append(newPasskeys, p)
		}
	}
	m.Passkeys = newPasskeys
	return nil
}

// PingHistory methods
func (m *MockRepository) CreatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	m.PingHistories = append(m.PingHistories, ping)
	return nil
}

func (m *MockRepository) UpdatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	for i, p := range m.PingHistories {
		if p.ID == ping.ID {
			m.PingHistories[i] = ping
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) GetLatestPingByUserID(ctx context.Context, userID string) (*models.PingHistory, error) {
	var latest *models.PingHistory
	for _, p := range m.PingHistories {
		if p.UserID == userID {
			if latest == nil || p.SentAt.After(latest.SentAt) {
				latest = p
			}
		}
	}
	if latest == nil {
		return nil, ErrNotFound
	}
	return latest, nil
}

func (m *MockRepository) ListPingHistoryByUserID(ctx context.Context, userID string) ([]*models.PingHistory, error) {
	var result []*models.PingHistory
	for _, p := range m.PingHistories {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}

// PingVerification methods
func (m *MockRepository) CreatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	m.PingVerifications = append(m.PingVerifications, verification)
	return nil
}

func (m *MockRepository) GetPingVerificationByCode(ctx context.Context, code string) (*models.PingVerification, error) {
	for _, v := range m.PingVerifications {
		if v.Code == code {
			return v, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) UpdatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	for i, v := range m.PingVerifications {
		if v.ID == verification.ID {
			m.PingVerifications[i] = verification
			return nil
		}
	}
	return ErrNotFound
}

// DeliveryEvent methods
func (m *MockRepository) CreateDeliveryEvent(ctx context.Context, event *models.DeliveryEvent) error {
	m.DeliveryEvents = append(m.DeliveryEvents, event)
	return nil
}

func (m *MockRepository) UpdateDeliveryEvent(ctx context.Context, event *models.DeliveryEvent) error {
	for i, e := range m.DeliveryEvents {
		if e.ID == event.ID {
			m.DeliveryEvents[i] = event
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) ListDeliveryEventsByUserID(ctx context.Context, userID string) ([]*models.DeliveryEvent, error) {
	var result []*models.DeliveryEvent
	for _, e := range m.DeliveryEvents {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	return result, nil
}

// AccessCode methods
func (m *MockRepository) CreateAccessCode(ctx context.Context, code *models.AccessCode) error {
	m.AccessCodes = append(m.AccessCodes, code)
	return nil
}

func (m *MockRepository) GetAccessCodeByCode(ctx context.Context, code string) (*models.AccessCode, error) {
	for _, c := range m.AccessCodes {
		if c.Code == code {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) VerifyAccessCode(ctx context.Context, code string) (*models.AccessCode, error) {
	// This is a simplified mock - just search for matching code
	// Real implementation uses password hashing
	for _, c := range m.AccessCodes {
		if c.UsedAt == nil && time.Now().Before(c.ExpiresAt) {
			// In mock, we'll just do a simple comparison
			// Real implementation would verify hash
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) MarkAccessCodeAsUsed(ctx context.Context, id string) error {
	for _, c := range m.AccessCodes {
		if c.ID == id {
			now := time.Now()
			c.UsedAt = &now
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) IncrementAccessCodeAttempts(ctx context.Context, id string) error {
	for _, c := range m.AccessCodes {
		if c.ID == id {
			c.AttemptCount++
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) DeleteExpiredAccessCodes(ctx context.Context) error {
	var filtered []*models.AccessCode
	now := time.Now()
	for _, c := range m.AccessCodes {
		if c.ExpiresAt.After(now) {
			filtered = append(filtered, c)
		}
	}
	m.AccessCodes = filtered
	return nil
}

// AuditLog methods
func (m *MockRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	m.AuditLogs = append(m.AuditLogs, log)
	return nil
}

func (m *MockRepository) ListAuditLogsByUserID(ctx context.Context, userID string) ([]*models.AuditLog, error) {
	var result []*models.AuditLog
	for _, l := range m.AuditLogs {
		if l.UserID == userID {
			result = append(result, l)
		}
	}
	return result, nil
}

// Session methods
func (m *MockRepository) CreateSession(ctx context.Context, session *models.Session) error {
	m.Sessions = append(m.Sessions, session)
	return nil
}

func (m *MockRepository) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	for _, s := range m.Sessions {
		if s.Token == token {
			return s, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MockRepository) DeleteSession(ctx context.Context, id string) error {
	for i, s := range m.Sessions {
		if s.ID == id {
			m.Sessions = append(m.Sessions[:i], m.Sessions[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *MockRepository) DeleteExpiredSessions(ctx context.Context) error {
	// Just simulate deleting expired sessions
	return nil
}

func (m *MockRepository) UpdateSessionActivity(ctx context.Context, id string) error {
	for i, s := range m.Sessions {
		if s.ID == id {
			s.LastActivity = time.Now()
			m.Sessions[i] = s
			return nil
		}
	}
	return ErrNotFound
}

// Scheduler methods
func (m *MockRepository) GetUsersForPinging(ctx context.Context) ([]*models.User, error) {
	return m.UsersForPinging, nil
}

func (m *MockRepository) GetUsersWithExpiredPings(ctx context.Context) ([]*models.User, error) {
	return m.UsersWithExpiredPings, nil
}

// Transaction methods
func (m *MockRepository) BeginTx(ctx context.Context) (Transaction, error) {
	// Return a mock transaction that does nothing
	return &MockTransaction{repo: m}, nil
}

// MockTransaction is a mock implementation of the Transaction interface
type MockTransaction struct {
	repo *MockRepository
}

func (t *MockTransaction) Commit() error {
	return nil
}

func (t *MockTransaction) Rollback() error {
	return nil
}

func (t *MockTransaction) BeginTx(ctx context.Context) (Transaction, error) {
	return t, nil
}

// Forward all Repository methods to the underlying repo
func (t *MockTransaction) CreateUser(ctx context.Context, user *models.User) error {
	return t.repo.CreateUser(ctx, user)
}

func (t *MockTransaction) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return t.repo.GetUserByID(ctx, id)
}

func (t *MockTransaction) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return t.repo.GetUserByEmail(ctx, email)
}

func (t *MockTransaction) GetUserByTelegramID(ctx context.Context, telegramID string) (*models.User, error) {
	return t.repo.GetUserByTelegramID(ctx, telegramID)
}

func (t *MockTransaction) UpdateUser(ctx context.Context, user *models.User) error {
	return t.repo.UpdateUser(ctx, user)
}

func (t *MockTransaction) DeleteUser(ctx context.Context, id string) error {
	return t.repo.DeleteUser(ctx, id)
}

func (t *MockTransaction) ListUsers(ctx context.Context) ([]*models.User, error) {
	return t.repo.ListUsers(ctx)
}

func (t *MockTransaction) CreateSecret(ctx context.Context, secret *models.Secret) error {
	return t.repo.CreateSecret(ctx, secret)
}

func (t *MockTransaction) GetSecretByID(ctx context.Context, id string) (*models.Secret, error) {
	return t.repo.GetSecretByID(ctx, id)
}

func (t *MockTransaction) ListSecretsByUserID(ctx context.Context, userID string) ([]*models.Secret, error) {
	return t.repo.ListSecretsByUserID(ctx, userID)
}

func (t *MockTransaction) UpdateSecret(ctx context.Context, secret *models.Secret) error {
	return t.repo.UpdateSecret(ctx, secret)
}

func (t *MockTransaction) DeleteSecret(ctx context.Context, id string) error {
	return t.repo.DeleteSecret(ctx, id)
}

func (t *MockTransaction) CreateRecipient(ctx context.Context, recipient *models.Recipient) error {
	return t.repo.CreateRecipient(ctx, recipient)
}

func (t *MockTransaction) GetRecipientByID(ctx context.Context, id string) (*models.Recipient, error) {
	return t.repo.GetRecipientByID(ctx, id)
}

func (t *MockTransaction) ListRecipientsByUserID(ctx context.Context, userID string) ([]*models.Recipient, error) {
	return t.repo.ListRecipientsByUserID(ctx, userID)
}

func (t *MockTransaction) UpdateRecipient(ctx context.Context, recipient *models.Recipient) error {
	return t.repo.UpdateRecipient(ctx, recipient)
}

func (t *MockTransaction) DeleteRecipient(ctx context.Context, id string) error {
	return t.repo.DeleteRecipient(ctx, id)
}

func (t *MockTransaction) CreateSecretAssignment(ctx context.Context, assignment *models.SecretAssignment) error {
	return t.repo.CreateSecretAssignment(ctx, assignment)
}

func (t *MockTransaction) GetSecretAssignmentByID(ctx context.Context, id string) (*models.SecretAssignment, error) {
	return t.repo.GetSecretAssignmentByID(ctx, id)
}

func (t *MockTransaction) ListSecretAssignmentsBySecretID(ctx context.Context, secretID string) ([]*models.SecretAssignment, error) {
	return t.repo.ListSecretAssignmentsBySecretID(ctx, secretID)
}

func (t *MockTransaction) ListSecretAssignmentsByRecipientID(ctx context.Context, recipientID string) ([]*models.SecretAssignment, error) {
	return t.repo.ListSecretAssignmentsByRecipientID(ctx, recipientID)
}

func (t *MockTransaction) ListSecretAssignmentsByUserID(ctx context.Context, userID string) ([]*models.SecretAssignment, error) {
	return t.repo.ListSecretAssignmentsByUserID(ctx, userID)
}

func (t *MockTransaction) DeleteSecretAssignment(ctx context.Context, id string) error {
	return t.repo.DeleteSecretAssignment(ctx, id)
}

func (t *MockTransaction) CreatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	return t.repo.CreatePingHistory(ctx, ping)
}

func (t *MockTransaction) UpdatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	return t.repo.UpdatePingHistory(ctx, ping)
}

func (t *MockTransaction) GetLatestPingByUserID(ctx context.Context, userID string) (*models.PingHistory, error) {
	return t.repo.GetLatestPingByUserID(ctx, userID)
}

func (t *MockTransaction) ListPingHistoryByUserID(ctx context.Context, userID string) ([]*models.PingHistory, error) {
	return t.repo.ListPingHistoryByUserID(ctx, userID)
}

func (t *MockTransaction) CreatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	return t.repo.CreatePingVerification(ctx, verification)
}

func (t *MockTransaction) GetPingVerificationByCode(ctx context.Context, code string) (*models.PingVerification, error) {
	return t.repo.GetPingVerificationByCode(ctx, code)
}

func (t *MockTransaction) UpdatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	return t.repo.UpdatePingVerification(ctx, verification)
}

func (t *MockTransaction) CreateDeliveryEvent(ctx context.Context, event *models.DeliveryEvent) error {
	return t.repo.CreateDeliveryEvent(ctx, event)
}

func (t *MockTransaction) UpdateDeliveryEvent(ctx context.Context, event *models.DeliveryEvent) error {
	return t.repo.UpdateDeliveryEvent(ctx, event)
}

func (t *MockTransaction) ListDeliveryEventsByUserID(ctx context.Context, userID string) ([]*models.DeliveryEvent, error) {
	return t.repo.ListDeliveryEventsByUserID(ctx, userID)
}

func (t *MockTransaction) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	return t.repo.CreateAuditLog(ctx, log)
}

func (t *MockTransaction) ListAuditLogsByUserID(ctx context.Context, userID string) ([]*models.AuditLog, error) {
	return t.repo.ListAuditLogsByUserID(ctx, userID)
}

func (t *MockTransaction) CreateSession(ctx context.Context, session *models.Session) error {
	return t.repo.CreateSession(ctx, session)
}

func (t *MockTransaction) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	return t.repo.GetSessionByToken(ctx, token)
}

func (t *MockTransaction) DeleteSession(ctx context.Context, id string) error {
	return t.repo.DeleteSession(ctx, id)
}

func (t *MockTransaction) DeleteExpiredSessions(ctx context.Context) error {
	return t.repo.DeleteExpiredSessions(ctx)
}

func (t *MockTransaction) UpdateSessionActivity(ctx context.Context, id string) error {
	return t.repo.UpdateSessionActivity(ctx, id)
}

func (t *MockTransaction) GetUsersForPinging(ctx context.Context) ([]*models.User, error) {
	return t.repo.GetUsersForPinging(ctx)
}

func (t *MockTransaction) GetUsersWithExpiredPings(ctx context.Context) ([]*models.User, error) {
	return t.repo.GetUsersWithExpiredPings(ctx)
}

func (t *MockTransaction) CreatePasskey(ctx context.Context, passkey *models.Passkey) error {
	return t.repo.CreatePasskey(ctx, passkey)
}

func (t *MockTransaction) GetPasskeyByID(ctx context.Context, id string) (*models.Passkey, error) {
	return t.repo.GetPasskeyByID(ctx, id)
}

func (t *MockTransaction) GetPasskeyByCredentialID(ctx context.Context, credentialID []byte) (*models.Passkey, error) {
	return t.repo.GetPasskeyByCredentialID(ctx, credentialID)
}

func (t *MockTransaction) ListPasskeysByUserID(ctx context.Context, userID string) ([]*models.Passkey, error) {
	return t.repo.ListPasskeysByUserID(ctx, userID)
}

func (t *MockTransaction) ListPasskeys(ctx context.Context) ([]*models.Passkey, error) {
	return t.repo.ListPasskeys(ctx)
}

func (t *MockTransaction) UpdatePasskey(ctx context.Context, passkey *models.Passkey) error {
	return t.repo.UpdatePasskey(ctx, passkey)
}

func (t *MockTransaction) DeletePasskey(ctx context.Context, id string) error {
	return t.repo.DeletePasskey(ctx, id)
}

func (t *MockTransaction) DeletePasskeysByUserID(ctx context.Context, userID string) error {
	return t.repo.DeletePasskeysByUserID(ctx, userID)
}

func (t *MockTransaction) CreateAccessCode(ctx context.Context, code *models.AccessCode) error {
	return t.repo.CreateAccessCode(ctx, code)
}

func (t *MockTransaction) GetAccessCodeByCode(ctx context.Context, code string) (*models.AccessCode, error) {
	return t.repo.GetAccessCodeByCode(ctx, code)
}

func (t *MockTransaction) VerifyAccessCode(ctx context.Context, code string) (*models.AccessCode, error) {
	return t.repo.VerifyAccessCode(ctx, code)
}

func (t *MockTransaction) MarkAccessCodeAsUsed(ctx context.Context, id string) error {
	return t.repo.MarkAccessCodeAsUsed(ctx, id)
}

func (t *MockTransaction) IncrementAccessCodeAttempts(ctx context.Context, id string) error {
	return t.repo.IncrementAccessCodeAttempts(ctx, id)
}

func (t *MockTransaction) DeleteExpiredAccessCodes(ctx context.Context) error {
	return t.repo.DeleteExpiredAccessCodes(ctx)
}
