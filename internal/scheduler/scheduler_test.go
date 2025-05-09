package scheduler

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/email"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// MockRepository is a mock implementation of the storage.Repository interface
type MockRepository struct {
	users                 []*models.User
	recipients            []*models.Recipient
	secrets               []*models.Secret
	secretAssignments     []*models.SecretAssignment
	pingHistories         []*models.PingHistory
	pingVerifications     []*models.PingVerification
	deliveryEvents        []*models.DeliveryEvent
	auditLogs             []*models.AuditLog
	sessions              []*models.Session
	usersForPinging       []*models.User
	usersWithExpiredPings []*models.User

	// Custom behavior functions
	GetLatestPingByUserIDFunc  func(ctx context.Context, userID string) (*models.PingHistory, error)
	CreatePingHistoryFunc      func(ctx context.Context, ping *models.PingHistory) error
	CreatePingVerificationFunc func(ctx context.Context, verification *models.PingVerification) error
	CreateAuditLogFunc         func(ctx context.Context, log *models.AuditLog) error
	ListUsersFunc              func(ctx context.Context) ([]*models.User, error)
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:                 make([]*models.User, 0),
		recipients:            make([]*models.Recipient, 0),
		secrets:               make([]*models.Secret, 0),
		secretAssignments:     make([]*models.SecretAssignment, 0),
		pingHistories:         make([]*models.PingHistory, 0),
		pingVerifications:     make([]*models.PingVerification, 0),
		deliveryEvents:        make([]*models.DeliveryEvent, 0),
		auditLogs:             make([]*models.AuditLog, 0),
		sessions:              make([]*models.Session, 0),
		usersForPinging:       make([]*models.User, 0),
		usersWithExpiredPings: make([]*models.User, 0),
	}
}

func (m *MockRepository) GetUsersForPinging(ctx context.Context) ([]*models.User, error) {
	return m.usersForPinging, nil
}

func (m *MockRepository) GetUsersWithExpiredPings(ctx context.Context) ([]*models.User, error) {
	return m.usersWithExpiredPings, nil
}

func (m *MockRepository) CreatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	if m.CreatePingHistoryFunc != nil {
		return m.CreatePingHistoryFunc(ctx, ping)
	}
	m.pingHistories = append(m.pingHistories, ping)
	return nil
}

func (m *MockRepository) CreatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	if m.CreatePingVerificationFunc != nil {
		return m.CreatePingVerificationFunc(ctx, verification)
	}
	m.pingVerifications = append(m.pingVerifications, verification)
	return nil
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *models.User) error {
	for i, u := range m.users {
		if u.ID == user.ID {
			m.users[i] = user
			return nil
		}
	}
	m.users = append(m.users, user)
	return nil
}

func (m *MockRepository) ListRecipientsByUserID(ctx context.Context, userID string) ([]*models.Recipient, error) {
	var result []*models.Recipient
	for _, r := range m.recipients {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockRepository) ListSecretAssignmentsByRecipientID(ctx context.Context, recipientID string) ([]*models.SecretAssignment, error) {
	var result []*models.SecretAssignment
	for _, a := range m.secretAssignments {
		if a.RecipientID == recipientID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MockRepository) CreateDeliveryEvent(ctx context.Context, event *models.DeliveryEvent) error {
	m.deliveryEvents = append(m.deliveryEvents, event)
	return nil
}

func (m *MockRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	if m.CreateAuditLogFunc != nil {
		return m.CreateAuditLogFunc(ctx, log)
	}
	m.auditLogs = append(m.auditLogs, log)
	return nil
}

func (m *MockRepository) ListAuditLogsByUserID(ctx context.Context, userID string) ([]*models.AuditLog, error) {
	result := make([]*models.AuditLog, 0)
	for _, log := range m.auditLogs {
		if log.UserID == userID {
			result = append(result, log)
		}
	}
	return result, nil
}

func (m *MockRepository) DeleteExpiredSessions(ctx context.Context) error {
	// Just simulate deleting expired sessions
	return nil
}

// Implement other methods of the Repository interface with empty implementations
func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error { return nil }
func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return nil, nil
}
func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *MockRepository) GetUserByTelegramID(ctx context.Context, telegramID string) (*models.User, error) {
	return nil, nil
}
func (m *MockRepository) DeleteUser(ctx context.Context, id string) error { return nil }
func (m *MockRepository) ListUsers(ctx context.Context) ([]*models.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx)
	}
	return m.users, nil
}
func (m *MockRepository) CreateSecret(ctx context.Context, secret *models.Secret) error { return nil }
func (m *MockRepository) GetSecretByID(ctx context.Context, id string) (*models.Secret, error) {
	return nil, nil
}
func (m *MockRepository) ListSecretsByUserID(ctx context.Context, userID string) ([]*models.Secret, error) {
	return nil, nil
}
func (m *MockRepository) UpdateSecret(ctx context.Context, secret *models.Secret) error { return nil }
func (m *MockRepository) DeleteSecret(ctx context.Context, id string) error             { return nil }
func (m *MockRepository) CreateRecipient(ctx context.Context, recipient *models.Recipient) error {
	return nil
}
func (m *MockRepository) GetRecipientByID(ctx context.Context, id string) (*models.Recipient, error) {
	return nil, nil
}
func (m *MockRepository) UpdateRecipient(ctx context.Context, recipient *models.Recipient) error {
	return nil
}
func (m *MockRepository) DeleteRecipient(ctx context.Context, id string) error { return nil }
func (m *MockRepository) CreateSecretAssignment(ctx context.Context, assignment *models.SecretAssignment) error {
	return nil
}
func (m *MockRepository) GetSecretAssignmentByID(ctx context.Context, id string) (*models.SecretAssignment, error) {
	return nil, nil
}
func (m *MockRepository) ListSecretAssignmentsBySecretID(ctx context.Context, secretID string) ([]*models.SecretAssignment, error) {
	return nil, nil
}
func (m *MockRepository) ListSecretAssignmentsByUserID(ctx context.Context, userID string) ([]*models.SecretAssignment, error) {
	return nil, nil
}
func (m *MockRepository) DeleteSecretAssignment(ctx context.Context, id string) error { return nil }
func (m *MockRepository) UpdatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	return nil
}
func (m *MockRepository) GetLatestPingByUserID(ctx context.Context, userID string) (*models.PingHistory, error) {
	if m.GetLatestPingByUserIDFunc != nil {
		return m.GetLatestPingByUserIDFunc(ctx, userID)
	}

	// Default implementation
	for _, ping := range m.pingHistories {
		if ping.UserID == userID {
			return ping, nil
		}
	}
	return nil, storage.ErrNotFound
}
func (m *MockRepository) ListPingHistoryByUserID(ctx context.Context, userID string) ([]*models.PingHistory, error) {
	return nil, nil
}
func (m *MockRepository) GetPingVerificationByCode(ctx context.Context, code string) (*models.PingVerification, error) {
	return nil, nil
}
func (m *MockRepository) UpdatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	return nil
}
func (m *MockRepository) ListDeliveryEventsByUserID(ctx context.Context, userID string) ([]*models.DeliveryEvent, error) {
	return nil, nil
}
func (m *MockRepository) CreateSession(ctx context.Context, session *models.Session) error {
	return nil
}
func (m *MockRepository) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	return nil, nil
}
func (m *MockRepository) DeleteSession(ctx context.Context, id string) error         { return nil }
func (m *MockRepository) UpdateSessionActivity(ctx context.Context, id string) error { return nil }
func (m *MockRepository) BeginTx(ctx context.Context) (storage.Transaction, error)   { return nil, nil }
func (m *MockRepository) ListPasskeysByUserID(ctx context.Context, userID string) ([]*models.Passkey, error) {
	return nil, nil
}
func (m *MockRepository) ListPasskeys(ctx context.Context) ([]*models.Passkey, error) {
	return nil, nil
}
func (m *MockRepository) GetPasskeyByCredentialID(ctx context.Context, credentialID []byte) (*models.Passkey, error) {
	return nil, nil
}
func (m *MockRepository) CreatePasskey(ctx context.Context, passkey *models.Passkey) error {
	return nil
}
func (m *MockRepository) UpdatePasskey(ctx context.Context, passkey *models.Passkey) error {
	return nil
}
func (m *MockRepository) GetPasskeyByID(ctx context.Context, id string) (*models.Passkey, error) {
	return nil, nil
}
func (m *MockRepository) DeletePasskey(ctx context.Context, id string) error              { return nil }
func (m *MockRepository) DeletePasskeysByUserID(ctx context.Context, userID string) error { return nil }

// MockEmailClient is a mock implementation of the email client
type MockEmailClient struct {
	sentEmails int
}

func (m *MockEmailClient) SendPingEmail(email, name, verificationCode string) error {
	m.sentEmails++
	return nil
}

func (m *MockEmailClient) SendSecretDeliveryEmail(recipientEmail, recipientName, message, accessCode string) error {
	m.sentEmails++
	return nil
}

func (m *MockEmailClient) SendEmail(options *email.MessageOptions) error {
	m.sentEmails++
	return nil
}

func (m *MockEmailClient) SendEmailSimple(to []string, subject, body string, isHTML bool) error {
	m.sentEmails++
	return nil
}

// MockTelegramBot is a mock implementation of the telegram bot
type MockTelegramBot struct {
	sentMessages int
}

func (m *MockTelegramBot) SendPingMessage(ctx context.Context, user *models.User, pingID string) error {
	m.sentMessages++
	return nil
}

func TestNewScheduler(t *testing.T) {
	repo := NewMockRepository()
	emailClient := &MockEmailClient{}
	telegramBot := &MockTelegramBot{}
	cfg := &config.Config{}

	scheduler := NewScheduler(repo, emailClient, telegramBot, cfg)

	if scheduler == nil {
		t.Fatal("Expected non-nil scheduler")
	}
	if scheduler.repo != repo {
		t.Error("Expected scheduler.repo to be the mock repository")
	}
	if scheduler.emailClient != emailClient {
		t.Error("Expected scheduler.emailClient to be the mock email client")
	}
	if scheduler.telegramBot != telegramBot {
		t.Error("Expected scheduler.telegramBot to be the mock telegram bot")
	}
	if scheduler.config != cfg {
		t.Error("Expected scheduler.config to be the mock config")
	}
	if scheduler.tasks == nil {
		t.Error("Expected scheduler.tasks to be initialized")
	}
	if scheduler.stopChan == nil {
		t.Error("Expected scheduler.stopChan to be initialized")
	}
}

func TestAddTask(t *testing.T) {
	scheduler := NewScheduler(nil, nil, nil, nil)

	task := &Task{
		ID:         "task1",
		Name:       "Test Task",
		Duration:   5 * time.Minute,
		RunOnStart: true,
		Handler:    func(ctx context.Context) error { return nil },
	}

	scheduler.AddTask(task)

	if len(scheduler.tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(scheduler.tasks))
	}

	if scheduler.tasks["task1"] != task {
		t.Error("Expected task to be added to scheduler.tasks")
	}

	if task.NextRun.IsZero() {
		t.Error("Expected task.NextRun to be set")
	}
}

func TestRegisterTasks(t *testing.T) {
	scheduler := NewScheduler(nil, nil, nil, nil)

	err := scheduler.registerTasks()
	if err != nil {
		t.Fatalf("registerTasks failed: %v", err)
	}

	if len(scheduler.tasks) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(scheduler.tasks))
	}

	// Check that the expected tasks are registered
	var hasPingTask, hasReminderTask, hasDeadSwitchTask, hasCleanupTask, hasExternalActivityTask bool
	for _, task := range scheduler.tasks {
		switch task.Name {
		case "PingTask":
			hasPingTask = true
			if task.Duration != 5*time.Minute {
				t.Errorf("Expected PingTask duration to be 5 minutes, got %v", task.Duration)
			}
			if !task.RunOnStart {
				t.Error("Expected PingTask.RunOnStart to be true")
			}
		case "ReminderTask":
			hasReminderTask = true
			if task.Duration != 30*time.Minute {
				t.Errorf("Expected ReminderTask duration to be 30 minutes, got %v", task.Duration)
			}
			if !task.RunOnStart {
				t.Error("Expected ReminderTask.RunOnStart to be true")
			}
		case "DeadSwitchTask":
			hasDeadSwitchTask = true
			if task.Duration != 15*time.Minute {
				t.Errorf("Expected DeadSwitchTask duration to be 15 minutes, got %v", task.Duration)
			}
			if !task.RunOnStart {
				t.Error("Expected DeadSwitchTask.RunOnStart to be true")
			}
		case "CleanupTask":
			hasCleanupTask = true
			if task.Duration != 24*time.Hour {
				t.Errorf("Expected CleanupTask duration to be 24 hours, got %v", task.Duration)
			}
			if task.RunOnStart {
				t.Error("Expected CleanupTask.RunOnStart to be false")
			}
		case "ExternalActivityTask":
			hasExternalActivityTask = true
			if task.Duration != 1*time.Hour {
				t.Errorf("Expected ExternalActivityTask duration to be 1 hour, got %v", task.Duration)
			}
			if !task.RunOnStart {
				t.Error("Expected ExternalActivityTask.RunOnStart to be true")
			}
		}
	}

	if !hasPingTask {
		t.Error("Expected PingTask to be registered")
	}
	if !hasDeadSwitchTask {
		t.Error("Expected DeadSwitchTask to be registered")
	}
	if !hasReminderTask {
		t.Error("Expected ReminderTask to be registered")
	}
	if !hasCleanupTask {
		t.Error("Expected CleanupTask to be registered")
	}
	if !hasExternalActivityTask {
		t.Error("Expected ExternalActivityTask to be registered")
	}
}

func TestStartStop(t *testing.T) {
	// Skip this test as it starts goroutines that can cause issues
	t.Skip("Skipping test that starts goroutines")

	scheduler := NewScheduler(NewMockRepository(), &MockEmailClient{}, &MockTelegramBot{}, &config.Config{})

	// Start the scheduler
	ctx := context.Background()
	err := scheduler.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the scheduler
	scheduler.Stop()

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)

	// No real assertions here, just making sure it doesn't panic
}

func TestPingTask(t *testing.T) {
	// Skip this test as it requires more complex mocking
	t.Skip("Skipping ping task test")

	repo := NewMockRepository()
	emailClient := &MockEmailClient{}
	telegramBot := &MockTelegramBot{}
	cfg := &config.Config{}

	_ = NewScheduler(repo, emailClient, telegramBot, cfg)

	// Add users for pinging
	user1 := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		PingingEnabled: true,
		PingMethod:     "email",
		PingFrequency:  3,
	}
	user2 := &models.User{
		ID:             "user2",
		Email:          "user2@example.com",
		TelegramID:     "tg123",
		PingingEnabled: true,
		PingMethod:     "telegram",
		PingFrequency:  3,
	}
	user3 := &models.User{
		ID:             "user3",
		Email:          "user3@example.com",
		TelegramID:     "tg456",
		PingingEnabled: true,
		PingMethod:     "both",
		PingFrequency:  3,
	}
	user4 := &models.User{
		ID:             "user4",
		Email:          "user4@example.com",
		PingingEnabled: false, // Disabled
		PingMethod:     "email",
		PingFrequency:  3,
	}

	repo.usersForPinging = []*models.User{user1, user2, user3, user4}

	// In a real test, we would run the ping task and check the results
	// ctx := context.Background()
	// err := scheduler.pingTask(ctx)
	// if err != nil {
	// 	t.Fatalf("pingTask failed: %v", err)
	// }

	// // Check that ping histories were created
	// if len(repo.pingHistories) < 3 {
	// 	t.Errorf("Expected at least 3 ping histories, got %d", len(repo.pingHistories))
	// }

	// // Check that ping verifications were created
	// if len(repo.pingVerifications) < 2 {
	// 	t.Errorf("Expected at least 2 ping verifications, got %d", len(repo.pingVerifications))
	// }

	// // Check that emails were sent
	// if emailClient.sentEmails < 2 {
	// 	t.Errorf("Expected at least 2 emails to be sent, got %d", emailClient.sentEmails)
	// }

	// // Check that telegram messages were sent
	// if telegramBot.sentMessages < 2 {
	// 	t.Errorf("Expected at least 2 telegram messages to be sent, got %d", telegramBot.sentMessages)
	// }

	// // Check that next ping times were updated
	// for _, user := range repo.users {
	// 	if user.PingingEnabled && user.NextScheduledPing.IsZero() {
	// 		t.Errorf("Expected NextScheduledPing to be set for user %s", user.ID)
	// 	}
	// }
}

func TestDeadSwitchTaskSkipped(t *testing.T) {
	// Skip this test as it requires more complex mocking
	t.Skip("Skipping dead switch task test")

	repo := NewMockRepository()
	emailClient := &MockEmailClient{}
	telegramBot := &MockTelegramBot{}
	cfg := &config.Config{}

	_ = NewScheduler(repo, emailClient, telegramBot, cfg)

	// Add users with expired pings
	user1 := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		PingingEnabled: true,
	}
	repo.usersWithExpiredPings = []*models.User{user1}

	// Add recipients for user1
	recipient1 := &models.Recipient{
		ID:     "recipient1",
		UserID: "user1",
		Email:  "recipient1@example.com",
		Name:   "Recipient 1",
	}
	recipient2 := &models.Recipient{
		ID:     "recipient2",
		UserID: "user1",
		Email:  "recipient2@example.com",
		Name:   "Recipient 2",
	}
	repo.recipients = []*models.Recipient{recipient1, recipient2}

	// Add secret assignments
	assignment1 := &models.SecretAssignment{
		ID:          "assignment1",
		UserID:      "user1",
		SecretID:    "secret1",
		RecipientID: "recipient1",
	}
	assignment2 := &models.SecretAssignment{
		ID:          "assignment2",
		UserID:      "user1",
		SecretID:    "secret2",
		RecipientID: "recipient2",
	}
	repo.secretAssignments = []*models.SecretAssignment{assignment1, assignment2}

	// In a real test, we would run the dead switch task and check the results
	// ctx := context.Background()
	// err := scheduler.deadSwitchTask(ctx)
	// if err != nil {
	// 	t.Fatalf("deadSwitchTask failed: %v", err)
	// }

	// // Check that delivery events were created
	// if len(repo.deliveryEvents) != 2 {
	// 	t.Errorf("Expected 2 delivery events, got %d", len(repo.deliveryEvents))
	// }

	// // Check that emails were sent
	// if emailClient.sentEmails != 2 {
	// 	t.Errorf("Expected 2 emails to be sent, got %d", emailClient.sentEmails)
	// }

	// // Check that pinging was disabled for the user
	// if user1.PingingEnabled {
	// 	t.Error("Expected PingingEnabled to be false after delivery")
	// }
}

func TestCleanupTaskSkipped(t *testing.T) {
	// Skip this test as it's very simple and just calls DeleteExpiredSessions
	t.Skip("Skipping cleanup task test")

	repo := NewMockRepository()
	scheduler := NewScheduler(repo, &MockEmailClient{}, &MockTelegramBot{}, &config.Config{})

	// Run the cleanup task
	ctx := context.Background()
	err := scheduler.cleanupTask(ctx)
	if err != nil {
		t.Fatalf("cleanupTask failed: %v", err)
	}

	// Not much to assert here since we're just calling DeleteExpiredSessions
	// which is mocked to do nothing
}

func TestHelperFunctions(t *testing.T) {
	// Test generateVerificationCode
	code := generateVerificationCode()
	if len(code) != 16 {
		t.Errorf("Expected verification code length to be 16, got %d", len(code))
	}

	// Test generateAccessCode
	accessCode := generateAccessCode()
	if len(accessCode) != 36 {
		t.Errorf("Expected access code length to be 36, got %d", len(accessCode))
	}

	// Test extractNameFromEmail
	tests := []struct {
		email    string
		expected string
	}{
		{"john.doe@example.com", "John Doe"},
		{"jane_smith@example.com", "Jane Smith"},
		{"user@example.com", "User"},
		{"first.middle.last@example.com", "First Middle Last"},
		{"@invalid", ""},
	}

	for _, test := range tests {
		result := extractNameFromEmail(test.email)
		if result != test.expected {
			t.Errorf("extractNameFromEmail(%s) = %s, expected %s", test.email, result, test.expected)
		}
	}

	// Test formatDuration
	durationTests := []struct {
		duration time.Duration
		expected string
	}{
		{1 * time.Hour, "1h 0m"},
		{90 * time.Minute, "1h 30m"},
		{30 * time.Minute, "30m"},
		{45 * time.Second, "1m"}, // Rounds to minutes
		{0, "0m"},
	}

	for _, test := range durationTests {
		result := formatDuration(test.duration)
		if result != test.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", test.duration, result, test.expected)
		}
	}
}

func TestPingTaskWithEmailMethod(t *testing.T) {
	// Create a mock repository
	repo := NewMockRepository()

	// Create a mock email client
	emailClient := &MockEmailClient{}

	// Create a mock telegram bot
	telegramBot := &MockTelegramBot{}

	// Create a scheduler
	scheduler := NewScheduler(repo, emailClient, telegramBot, &config.Config{})

	// Add a user for pinging with email method
	user := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		PingingEnabled: true,
		PingMethod:     "email",
		PingFrequency:  3,
		PingDeadline:   7,
	}

	// Add the user to the mock repository's usersForPinging slice
	repo.usersForPinging = []*models.User{user}

	// Run the ping task
	ctx := context.Background()
	err := scheduler.pingTask(ctx)
	if err != nil {
		t.Fatalf("pingTask failed: %v", err)
	}

	// Check that a ping history was created
	if len(repo.pingHistories) != 1 {
		t.Fatalf("Expected 1 ping history, got %d", len(repo.pingHistories))
	}

	// Check that the ping history has the correct user ID and method
	pingHistory := repo.pingHistories[0]
	if pingHistory.UserID != user.ID {
		t.Errorf("Expected ping history user ID to be %s, got %s", user.ID, pingHistory.UserID)
	}
	if pingHistory.Method != "email" {
		t.Errorf("Expected ping history method to be 'email', got '%s'", pingHistory.Method)
	}

	// Check that a ping verification was created
	if len(repo.pingVerifications) != 1 {
		t.Fatalf("Expected 1 ping verification, got %d", len(repo.pingVerifications))
	}

	// Check that the ping verification has the correct user ID
	pingVerification := repo.pingVerifications[0]
	if pingVerification.UserID != user.ID {
		t.Errorf("Expected ping verification user ID to be %s, got %s", user.ID, pingVerification.UserID)
	}

	// Check that an email was sent
	if emailClient.sentEmails != 1 {
		t.Errorf("Expected 1 email to be sent, got %d", emailClient.sentEmails)
	}

	// Check that the user's next scheduled ping was updated
	if user.NextScheduledPing.IsZero() {
		t.Error("Expected NextScheduledPing to be set")
	}
}

func TestPingTaskWithTelegramMethod(t *testing.T) {
	// Create a mock repository
	repo := NewMockRepository()

	// Create a mock email client
	emailClient := &MockEmailClient{}

	// Create a mock telegram bot
	telegramBot := &MockTelegramBot{}

	// Create a scheduler
	scheduler := NewScheduler(repo, emailClient, telegramBot, &config.Config{})

	// Add a user for pinging with telegram method
	user := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		TelegramID:     "tg123",
		PingingEnabled: true,
		PingMethod:     "telegram",
		PingFrequency:  3,
		PingDeadline:   7,
	}

	// Add the user to the mock repository's usersForPinging slice
	repo.usersForPinging = []*models.User{user}

	// Run the ping task
	ctx := context.Background()
	err := scheduler.pingTask(ctx)
	if err != nil {
		t.Fatalf("pingTask failed: %v", err)
	}

	// Check that a ping history was created
	if len(repo.pingHistories) != 1 {
		t.Fatalf("Expected 1 ping history, got %d", len(repo.pingHistories))
	}

	// Check that the ping history has the correct user ID and method
	pingHistory := repo.pingHistories[0]
	if pingHistory.UserID != user.ID {
		t.Errorf("Expected ping history user ID to be %s, got %s", user.ID, pingHistory.UserID)
	}
	if pingHistory.Method != "telegram" {
		t.Errorf("Expected ping history method to be 'telegram', got '%s'", pingHistory.Method)
	}

	// Check that a telegram message was sent
	if telegramBot.sentMessages != 1 {
		t.Errorf("Expected 1 telegram message to be sent, got %d", telegramBot.sentMessages)
	}

	// Check that the user's next scheduled ping was updated
	if user.NextScheduledPing.IsZero() {
		t.Error("Expected NextScheduledPing to be set")
	}
}

func TestDeadSwitchTask(t *testing.T) {
	// Create a mock repository
	repo := NewMockRepository()

	// Create a mock email client
	emailClient := &MockEmailClient{}

	// Create a mock telegram bot
	telegramBot := &MockTelegramBot{}

	// Create a scheduler
	scheduler := NewScheduler(repo, emailClient, telegramBot, &config.Config{})

	// Add a user with expired pings
	user := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		PingingEnabled: true,
	}

	// Add the user to the mock repository's usersWithExpiredPings slice
	repo.usersWithExpiredPings = []*models.User{user}

	// Add recipients for the user
	recipient1 := &models.Recipient{
		ID:     "recipient1",
		UserID: "user1",
		Email:  "recipient1@example.com",
		Name:   "Recipient 1",
	}
	recipient2 := &models.Recipient{
		ID:     "recipient2",
		UserID: "user1",
		Email:  "recipient2@example.com",
		Name:   "Recipient 2",
	}
	repo.recipients = []*models.Recipient{recipient1, recipient2}

	// Add secret assignments
	assignment1 := &models.SecretAssignment{
		ID:          "assignment1",
		UserID:      "user1",
		SecretID:    "secret1",
		RecipientID: "recipient1",
	}
	assignment2 := &models.SecretAssignment{
		ID:          "assignment2",
		UserID:      "user1",
		SecretID:    "secret2",
		RecipientID: "recipient2",
	}
	repo.secretAssignments = []*models.SecretAssignment{assignment1, assignment2}

	// Run the dead switch task
	ctx := context.Background()
	err := scheduler.deadSwitchTask(ctx)
	if err != nil {
		t.Fatalf("deadSwitchTask failed: %v", err)
	}

	// Check that delivery events were created
	if len(repo.deliveryEvents) != 2 {
		t.Errorf("Expected 2 delivery events, got %d", len(repo.deliveryEvents))
	}

	// Check that emails were sent
	if emailClient.sentEmails != 2 {
		t.Errorf("Expected 2 emails to be sent, got %d", emailClient.sentEmails)
	}

	// Check that pinging was disabled for the user
	if user.PingingEnabled {
		t.Error("Expected PingingEnabled to be false after delivery")
	}
}

func TestCleanupTask(t *testing.T) {
	// Create a mock repository
	repo := NewMockRepository()

	// Create a scheduler
	scheduler := NewScheduler(repo, &MockEmailClient{}, &MockTelegramBot{}, &config.Config{})

	// Run the cleanup task
	ctx := context.Background()
	err := scheduler.cleanupTask(ctx)
	if err != nil {
		t.Fatalf("cleanupTask failed: %v", err)
	}

	// Not much to assert here since we're just calling DeleteExpiredSessions
	// which is mocked to do nothing
}

func TestReminderTask(t *testing.T) {
	// Create a mock repository
	repo := NewMockRepository()

	// Create a mock email client
	emailClient := &MockEmailClient{}

	// Create a mock telegram bot
	telegramBot := &MockTelegramBot{}

	// Create a scheduler
	scheduler := NewScheduler(repo, emailClient, telegramBot, &config.Config{})

	// Create a user with an approaching deadline
	now := time.Now().UTC()
	user := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		PingingEnabled: true,
		PingMethod:     "email",
		PingFrequency:  7,
		PingDeadline:   14,
		LastActivity:   now.Add(-13 * 24 * time.Hour), // 13 days ago, deadline in 1 day
	}

	// Add the user to the repository
	repo.users = append(repo.users, user)

	// Create a ping history for the user (more than 12 hours ago)
	pingHistory := &models.PingHistory{
		ID:     "ping1",
		UserID: user.ID,
		SentAt: now.Add(-13 * time.Hour),
		Method: "email",
		Status: "sent",
	}
	repo.pingHistories = append(repo.pingHistories, pingHistory)

	// Mock the ListUsers method to return our user
	repo.ListUsersFunc = func(ctx context.Context) ([]*models.User, error) {
		return []*models.User{user}, nil
	}

	// Set up the function fields to track calls
	repo.CreatePingHistoryFunc = func(ctx context.Context, ping *models.PingHistory) error {
		repo.pingHistories = append(repo.pingHistories, ping)
		return nil
	}

	repo.CreatePingVerificationFunc = func(ctx context.Context, verification *models.PingVerification) error {
		repo.pingVerifications = append(repo.pingVerifications, verification)
		return nil
	}

	repo.CreateAuditLogFunc = func(ctx context.Context, log *models.AuditLog) error {
		repo.auditLogs = append(repo.auditLogs, log)
		return nil
	}

	// Mock the GetLatestPingByUserID method to return our ping history
	repo.GetLatestPingByUserIDFunc = func(ctx context.Context, userID string) (*models.PingHistory, error) {
		if userID == user.ID {
			return pingHistory, nil
		}
		return nil, storage.ErrNotFound
	}

	// Count the initial number of ping histories, verifications, and audit logs
	initialPingHistories := len(repo.pingHistories)
	initialPingVerifications := len(repo.pingVerifications)
	initialAuditLogs := len(repo.auditLogs)
	initialEmailsSent := emailClient.sentEmails

	// Run the reminder task
	ctx := context.Background()
	err := scheduler.reminderTask(ctx)
	if err != nil {
		t.Fatalf("reminderTask failed: %v", err)
	}

	// Check that new ping histories were created
	if len(repo.pingHistories) <= initialPingHistories {
		t.Errorf("Expected new ping histories to be created, got %d (was %d)", len(repo.pingHistories), initialPingHistories)
	}

	// Check that a ping verification was created
	if len(repo.pingVerifications) <= initialPingVerifications {
		t.Errorf("Expected ping verification to be created, got %d (was %d)", len(repo.pingVerifications), initialPingVerifications)
	}

	// Check that an email was sent
	if emailClient.sentEmails <= initialEmailsSent {
		t.Errorf("Expected email to be sent, got %d (was %d)", emailClient.sentEmails, initialEmailsSent)
	}

	// Check that an audit log was created
	if len(repo.auditLogs) <= initialAuditLogs {
		t.Errorf("Expected audit log to be created, got %d (was %d)", len(repo.auditLogs), initialAuditLogs)
	} else {
		// Check that the audit log has the correct action
		auditLog := repo.auditLogs[len(repo.auditLogs)-1] // Get the last audit log
		if !strings.Contains(strings.ToLower(auditLog.Action), "reminder") {
			t.Errorf("Expected audit log action to contain 'reminder', got '%s'", auditLog.Action)
		}
	}
}

func TestDeadSwitchTaskWithActivity(t *testing.T) {
	// Create a mock repository
	repo := NewMockRepository()

	// Create a mock email client
	emailClient := &MockEmailClient{}

	// Create a mock telegram bot
	telegramBot := &MockTelegramBot{}

	// Create a scheduler with a mock activity registry
	scheduler := NewScheduler(repo, emailClient, telegramBot, &config.Config{})

	// Create a user with expired pings but recent activity
	now := time.Now().UTC()
	user := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		PingingEnabled: true,
		PingMethod:     "email",
		PingFrequency:  7,
		PingDeadline:   14,
		LastActivity:   now.Add(-15 * 24 * time.Hour), // 15 days ago, past deadline
	}

	// Add the user to the repository's expired pings list
	repo.usersWithExpiredPings = []*models.User{user}

	// Create a ping history with a recent response
	respondedAt := now.Add(-1 * time.Hour) // 1 hour ago
	pingHistory := &models.PingHistory{
		ID:          "ping1",
		UserID:      user.ID,
		SentAt:      now.Add(-2 * time.Hour),
		Method:      "email",
		Status:      "responded",
		RespondedAt: &respondedAt,
	}
	repo.pingHistories = append(repo.pingHistories, pingHistory)

	// Mock the GetLatestPingByUserID method to return our ping history
	repo.GetLatestPingByUserIDFunc = func(ctx context.Context, userID string) (*models.PingHistory, error) {
		if userID == user.ID {
			return pingHistory, nil
		}
		return nil, storage.ErrNotFound
	}

	// Run the dead switch task
	ctx := context.Background()
	err := scheduler.deadSwitchTask(ctx)
	if err != nil {
		t.Fatalf("deadSwitchTask failed: %v", err)
	}

	// Check that the user's last activity was updated
	if user.LastActivity.Equal(now.Add(-15 * 24 * time.Hour)) {
		t.Error("Expected user's LastActivity to be updated")
	}

	// Check that the user's next scheduled ping was updated
	if user.NextScheduledPing.IsZero() {
		t.Error("Expected user's NextScheduledPing to be updated")
	}

	// Check that an audit log was created
	if len(repo.auditLogs) == 0 {
		t.Error("Expected audit log to be created")
	} else {
		// Check that the audit log has the correct action
		auditLog := repo.auditLogs[0]
		if !strings.Contains(auditLog.Action, "cancelled") {
			t.Errorf("Expected audit log action to contain 'cancelled', got '%s'", auditLog.Action)
		}
	}

	// Check that no delivery events were created (switch was not triggered)
	if len(repo.deliveryEvents) > 0 {
		t.Errorf("Expected no delivery events, got %d", len(repo.deliveryEvents))
	}
}

func TestDeadSwitchTaskWithNoActivity(t *testing.T) {
	// Create a mock repository
	repo := NewMockRepository()

	// Create a mock email client
	emailClient := &MockEmailClient{}

	// Create a mock telegram bot
	telegramBot := &MockTelegramBot{}

	// Create a scheduler
	scheduler := NewScheduler(repo, emailClient, telegramBot, &config.Config{})

	// Create a user with expired pings and no recent activity
	now := time.Now().UTC()
	user := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		PingingEnabled: true,
		PingMethod:     "email",
		PingFrequency:  7,
		PingDeadline:   14,
		LastActivity:   now.Add(-15 * 24 * time.Hour), // 15 days ago, past deadline
	}

	// Add the user to the repository's expired pings list
	repo.usersWithExpiredPings = []*models.User{user}

	// Add a recipient for the user
	recipient := &models.Recipient{
		ID:     "recipient1",
		UserID: user.ID,
		Email:  "recipient@example.com",
		Name:   "Test Recipient",
	}
	repo.recipients = append(repo.recipients, recipient)

	// Add a secret for the user
	secret := &models.Secret{
		ID:     "secret1",
		UserID: user.ID,
		Name:   "Test Secret",
	}
	repo.secrets = append(repo.secrets, secret)

	// Add a secret assignment
	assignment := &models.SecretAssignment{
		ID:          "assignment1",
		UserID:      user.ID,
		SecretID:    secret.ID,
		RecipientID: recipient.ID,
	}
	repo.secretAssignments = append(repo.secretAssignments, assignment)

	// Run the dead switch task
	ctx := context.Background()
	err := scheduler.deadSwitchTask(ctx)
	if err != nil {
		t.Fatalf("deadSwitchTask failed: %v", err)
	}

	// Check that an audit log was created
	if len(repo.auditLogs) == 0 {
		t.Error("Expected audit log to be created")
	} else {
		// Check that the audit log has the correct action
		auditLog := repo.auditLogs[0]
		if !strings.Contains(auditLog.Action, "triggered") {
			t.Errorf("Expected audit log action to contain 'triggered', got '%s'", auditLog.Action)
		}
	}

	// Check that delivery events were created
	if len(repo.deliveryEvents) == 0 {
		t.Error("Expected delivery events to be created")
	}

	// Check that emails were sent
	if emailClient.sentEmails == 0 {
		t.Error("Expected emails to be sent")
	}

	// Check that pinging was disabled for the user
	if user.PingingEnabled {
		t.Error("Expected user's PingingEnabled to be set to false")
	}
}
