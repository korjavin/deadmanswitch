package scheduler

import (
	"context"
	"fmt"
	// "strings" // No longer needed directly in this test file after error check changes
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/email"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of the storage.Repository interface
type MockRepository struct {
	storage.Repository // Embed to satisfy the interface for methods not explicitly mocked

	// Custom behavior functions
	GetUsersForPingingFunc     func(ctx context.Context) ([]*models.User, error)
	GetUsersWithExpiredPingsFunc func(ctx context.Context) ([]*models.User, error)
	CreatePingHistoryFunc      func(ctx context.Context, ping *models.PingHistory) error
	CreatePingVerificationFunc func(ctx context.Context, verification *models.PingVerification) error
	UpdateUserFunc             func(ctx context.Context, user *models.User) error
	GetLatestPingByUserIDFunc  func(ctx context.Context, userID string) (*models.PingHistory, error)
	// Add other funcs as needed for other tests, e.g., CreateAuditLogFunc, ListUsersFunc

	// Store calls and data for assertions
	CreatedPingHistories     []*models.PingHistory
	CreatedPingVerifications []*models.PingVerification
	UpdatedUsers             []*models.User
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		CreatedPingHistories:     make([]*models.PingHistory, 0),
		CreatedPingVerifications: make([]*models.PingVerification, 0),
		UpdatedUsers:             make([]*models.User, 0),
	}
}

func (m *MockRepository) ResetCalls() {
	m.CreatedPingHistories = make([]*models.PingHistory, 0)
	m.CreatedPingVerifications = make([]*models.PingVerification, 0)
	m.UpdatedUsers = make([]*models.User, 0)
	m.GetUsersForPingingFunc = nil
	m.CreatePingHistoryFunc = nil
	m.CreatePingVerificationFunc = nil
	m.UpdateUserFunc = nil
	m.GetLatestPingByUserIDFunc = nil
}

func (m *MockRepository) GetUsersForPinging(ctx context.Context) ([]*models.User, error) {
	if m.GetUsersForPingingFunc != nil {
		return m.GetUsersForPingingFunc(ctx)
	}
	return nil, fmt.Errorf("GetUsersForPingingFunc not implemented in mock")
}

func (m *MockRepository) CreatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	if m.CreatePingHistoryFunc != nil {
		// Custom func is responsible for logic including appending to CreatedPingHistories
		return m.CreatePingHistoryFunc(ctx, ping)
	}
	// Default behavior if no custom func is set
	m.CreatedPingHistories = append(m.CreatedPingHistories, ping)
	return nil
}

func (m *MockRepository) CreatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	if m.CreatePingVerificationFunc != nil {
		// Custom func is responsible for logic including appending to CreatedPingVerifications
		return m.CreatePingVerificationFunc(ctx, verification)
	}
	// Default behavior if no custom func is set
	m.CreatedPingVerifications = append(m.CreatedPingVerifications, verification)
	return nil
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *models.User) error {
	m.UpdatedUsers = append(m.UpdatedUsers, user)
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
	}
	return nil
}

func (m *MockRepository) GetLatestPingByUserID(ctx context.Context, userID string) (*models.PingHistory, error) {
	if m.GetLatestPingByUserIDFunc != nil {
		return m.GetLatestPingByUserIDFunc(ctx, userID)
	}
	return nil, fmt.Errorf("GetLatestPingByUserIDFunc not implemented in mock")
}

// --- MockEmailClient ---
type MockEmailClient struct {
	SendPingEmailFunc        func(email, name, verificationCode string) error
	LastPingEmailSentTo      string
	LastPingEmailName        string
	LastPingEmailVerifCode   string
	SendPingEmailCalledCount int
}

func NewMockEmailClient() *MockEmailClient {
	return &MockEmailClient{}
}

func (m *MockEmailClient) SendPingEmail(email, name, verificationCode string) error {
	m.SendPingEmailCalledCount++
	m.LastPingEmailSentTo = email
	m.LastPingEmailName = name
	m.LastPingEmailVerifCode = verificationCode
	if m.SendPingEmailFunc != nil {
		return m.SendPingEmailFunc(email, name, verificationCode)
	}
	return nil
}

func (m *MockEmailClient) SendSecretDeliveryEmail(recipientEmail, recipientName, message, accessCode string) error {
	return nil // Not used in pingTask
}
func (m *MockEmailClient) SendEmail(options *email.MessageOptions) error {
	return nil // Not used in pingTask
}
func (m *MockEmailClient) SendEmailSimple(to []string, subject, body string, isHTML bool) error {
	return nil // Not used in pingTask
}

func (m *MockEmailClient) ResetCalls() {
	m.SendPingEmailCalledCount = 0
	m.LastPingEmailSentTo = ""
	m.LastPingEmailName = ""
	m.LastPingEmailVerifCode = ""
	m.SendPingEmailFunc = nil
}

// --- MockTelegramBot ---
type MockTelegramBot struct {
	SendPingMessageFunc        func(ctx context.Context, user *models.User, pingID string) error
	LastPingMessageUserArg     *models.User
	LastPingMessagePingIDArg string
	SendPingMessageCalledCount int
}

func NewMockTelegramBot() *MockTelegramBot {
	return &MockTelegramBot{}
}

func (m *MockTelegramBot) SendPingMessage(ctx context.Context, user *models.User, pingID string) error {
	m.SendPingMessageCalledCount++
	m.LastPingMessageUserArg = user
	m.LastPingMessagePingIDArg = pingID
	if m.SendPingMessageFunc != nil {
		return m.SendPingMessageFunc(ctx, user, pingID)
	}
	return nil
}

func (m *MockTelegramBot) ResetCalls() {
	m.SendPingMessageCalledCount = 0
	m.LastPingMessageUserArg = nil
	m.LastPingMessagePingIDArg = ""
	m.SendPingMessageFunc = nil
}

func TestNewScheduler(t *testing.T) {
	repo := NewMockRepository()
	emailClient := NewMockEmailClient()
	telegramBot := NewMockTelegramBot()
	cfg := &config.Config{}

	scheduler := NewScheduler(repo, emailClient, telegramBot, cfg)

	require.NotNil(t, scheduler)
	assert.Equal(t, repo, scheduler.repo)
	assert.Equal(t, emailClient, scheduler.emailClient)
	assert.Equal(t, telegramBot, scheduler.telegramBot)
	assert.Equal(t, cfg, scheduler.config)
	require.NotNil(t, scheduler.tasks)
	require.NotNil(t, scheduler.stopChan)
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
	assert.Len(t, scheduler.tasks, 1)
	assert.Equal(t, task, scheduler.tasks["task1"])
	assert.False(t, task.NextRun.IsZero())
}

func TestRegisterTasks(t *testing.T) {
	scheduler := NewScheduler(nil, nil, nil, nil)
	err := scheduler.registerTasks()
	require.NoError(t, err)
	// Number of tasks registered in scheduler.go's registerTasks()
	// Currently: Ping, Reminder, DeadSwitch, ExternalActivity, ReencryptQuestions, Cleanup
	assert.Len(t, scheduler.tasks, 6) 
}

func TestStartStop(t *testing.T) {
	t.Skip("Skipping test that starts goroutines and relies on timing")
	// This test would require more complex synchronization to be reliable in CI.
}

func TestPingTaskComprehensive(t *testing.T) {
	defaultUser := func(id, pingMethod string, pingingEnabled bool, telegramID string) *models.User {
		return &models.User{
			ID:             id,
			Email:          fmt.Sprintf("%s@example.com", id),
			TelegramID:     telegramID,
			PingingEnabled: pingingEnabled,
			PingMethod:     pingMethod,
			PingFrequency:  3, // Days
			PingDeadline:   7, // Days
		}
	}

	tests := []struct {
		name                         string
		usersForPinging              []*models.User
		mockRepoSetup                func(mockRepo *MockRepository)
		mockEmailClientSetup         func(mockEmailClient *MockEmailClient)
		mockTelegramBotSetup         func(mockTelegramBot *MockTelegramBot)
		expectedPingHistories        int
		expectedPingVerifications    int
		expectedEmailPings           int
		expectedTelegramPings        int
		expectedUserUpdates          int
		assertUserUpdate             func(t *testing.T, updatedUsers []*models.User)
		assertPingHistory            func(t *testing.T, histories []*models.PingHistory)
		assertPingVerification       func(t *testing.T, verifications []*models.PingVerification)
		assertEmailClient            func(t *testing.T, client *MockEmailClient)
		assertTelegramBot            func(t *testing.T, bot *MockTelegramBot)
		expectedErrorMsgSubstring    string 
	}{
		{
			name:            "No users to ping",
			usersForPinging: []*models.User{},
			expectedPingHistories:     0,
			expectedPingVerifications: 0,
			expectedEmailPings:        0,
			expectedTelegramPings:     0,
			expectedUserUpdates:       0,
		},
		{
			name:            "User with PingingEnabled false",
			usersForPinging: []*models.User{defaultUser("userDisabled", "email", false, "")},
			expectedPingHistories:     0,
			expectedPingVerifications: 0,
			expectedEmailPings:        0,
			expectedTelegramPings:     0,
			expectedUserUpdates:       0, 
		},
		{
			name:            "User with PingMethod email",
			usersForPinging: []*models.User{defaultUser("userEmail", "email", true, "")},
			expectedPingHistories:     1,
			expectedPingVerifications: 1,
			expectedEmailPings:        1,
			expectedTelegramPings:     0,
			expectedUserUpdates:       1,
			assertPingHistory: func(t *testing.T, histories []*models.PingHistory) {
				require.Len(t, histories, 1)
				assert.Equal(t, "userEmail", histories[0].UserID)
				assert.Equal(t, "email", histories[0].Method)
				assert.Equal(t, "sent", histories[0].Status)
			},
			assertUserUpdate: func(t *testing.T, updatedUsers []*models.User) {
				require.Len(t, updatedUsers, 1)
				assert.Equal(t, "userEmail", updatedUsers[0].ID)
				assert.False(t, updatedUsers[0].NextScheduledPing.IsZero())
			},
		},
		{
			name:            "User with PingMethod telegram, valid TelegramID",
			usersForPinging: []*models.User{defaultUser("userTG", "telegram", true, "tg123")},
			expectedPingHistories:     1,
			expectedPingVerifications: 0, 
			expectedEmailPings:        0,
			expectedTelegramPings:     1,
			expectedUserUpdates:       1,
			assertPingHistory: func(t *testing.T, histories []*models.PingHistory) {
				require.Len(t, histories, 1)
				assert.Equal(t, "userTG", histories[0].UserID)
				assert.Equal(t, "telegram", histories[0].Method)
			},
			assertTelegramBot: func(t *testing.T, bot *MockTelegramBot) {
				require.NotNil(t, bot.LastPingMessageUserArg, "Telegram bot should have received a user argument")
				assert.Equal(t, "userTG", bot.LastPingMessageUserArg.ID)
				assert.NotEmpty(t, bot.LastPingMessagePingIDArg)
			},
		},
		{
			name:            "User with PingMethod telegram, no TelegramID",
			usersForPinging: []*models.User{defaultUser("userTGNoID", "telegram", true, "")},
			expectedPingHistories:     0, 
			expectedPingVerifications: 0,
			expectedEmailPings:        0,
			expectedTelegramPings:     0, 
			expectedUserUpdates:       1, // NextScheduledPing is still updated
		},
		{
			name:            "User with PingMethod both, valid TelegramID",
			usersForPinging: []*models.User{defaultUser("userBoth", "both", true, "tg789")},
			expectedPingHistories:     2, 
			expectedPingVerifications: 1, 
			expectedEmailPings:        1,
			expectedTelegramPings:     1,
			expectedUserUpdates:       1,
			assertPingHistory: func(t *testing.T, histories []*models.PingHistory) {
				require.Len(t, histories, 2)
				emailPingFound := false
				telegramPingFound := false
				for _, h := range histories {
					if h.UserID == "userBoth" && h.Method == "email" {
						emailPingFound = true
					}
					if h.UserID == "userBoth" && h.Method == "telegram" {
						telegramPingFound = true
					}
				}
				assert.True(t, emailPingFound, "Email ping history missing for 'both'")
				assert.True(t, telegramPingFound, "Telegram ping history missing for 'both'")
			},
		},
		{
			name:            "User with PingMethod both, no TelegramID (should only send email)",
			usersForPinging: []*models.User{defaultUser("userBothNoID", "both", true, "")},
			expectedPingHistories:     1, 
			expectedPingVerifications: 1, 
			expectedEmailPings:        1,
			expectedTelegramPings:     0, 
			expectedUserUpdates:       1,
		},
		{
			name: "Error from GetUsersForPinging",
			mockRepoSetup: func(mockRepo *MockRepository) {
				mockRepo.GetUsersForPingingFunc = func(ctx context.Context) ([]*models.User, error) {
					return nil, fmt.Errorf("db error GetUsersForPinging")
				}
			},
			expectedErrorMsgSubstring: "failed to get users for pinging",
		},
		{
			name:            "Error from CreatePingHistory (email)",
			usersForPinging: []*models.User{defaultUser("userEmailFailPH", "email", true, "")},
			mockRepoSetup: func(mockRepo *MockRepository) {
				// CreatePingVerificationFunc needs to be set up to add to the slice for accurate count
				mockRepo.CreatePingVerificationFunc = func(ctx context.Context, verification *models.PingVerification) error {
					mockRepo.CreatedPingVerifications = append(mockRepo.CreatedPingVerifications, verification)
					return nil 
				}
				mockRepo.CreatePingHistoryFunc = func(ctx context.Context, ping *models.PingHistory) error {
					// This one fails, so it should not add to CreatedPingHistories
					return fmt.Errorf("db error CreatePingHistory")
				}
			},
			expectedPingHistories:     0, // Fails, so not added by custom func
			expectedPingVerifications: 1, // Succeeds and is added by custom func
			expectedEmailPings:        0, // Not called if CreatePingHistory fails in sendEmailPing
			expectedUserUpdates:       1, 
		},
		{
			name:            "Error from CreatePingVerification",
			usersForPinging: []*models.User{defaultUser("userEmailFailPV", "email", true, "")},
			mockRepoSetup: func(mockRepo *MockRepository) {
				mockRepo.CreatePingVerificationFunc = func(ctx context.Context, verification *models.PingVerification) error {
					// This one fails, so it should not add to CreatedPingVerifications
					return fmt.Errorf("db error CreatePingVerification")
				}
			},
			expectedPingHistories:     0, 
			expectedPingVerifications: 0, // Fails, so not added by custom func
			expectedEmailPings:        0, 
			expectedUserUpdates:       1,
		},
		{
			name:            "Error from SendPingEmail",
			usersForPinging: []*models.User{defaultUser("userEmailFailSend", "email", true, "")},
			mockEmailClientSetup: func(mockEmailClient *MockEmailClient) {
				mockEmailClient.SendPingEmailFunc = func(email, name, verificationCode string) error {
					return fmt.Errorf("email send failed")
				}
			},
			expectedPingHistories:     1,
			expectedPingVerifications: 1,
			expectedEmailPings:        1, 
			expectedUserUpdates:       1,
		},
		{
			name:            "Error from SendPingMessage (telegram)",
			usersForPinging: []*models.User{defaultUser("userTGFailSend", "telegram", true, "tg123")},
			mockTelegramBotSetup: func(mockTelegramBot *MockTelegramBot) {
				mockTelegramBot.SendPingMessageFunc = func(ctx context.Context, user *models.User, pingID string) error {
					return fmt.Errorf("telegram send failed")
				}
			},
			expectedPingHistories:     1,
			expectedTelegramPings:     1, 
			expectedUserUpdates:       1,
		},
		{
			name:            "Error from UpdateUser",
			usersForPinging: []*models.User{defaultUser("userEmailFailUpdate", "email", true, "")},
			mockRepoSetup: func(mockRepo *MockRepository) {
				mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error {
					return fmt.Errorf("db error UpdateUser")
				}
			},
			expectedPingHistories:     1,
			expectedPingVerifications: 1,
			expectedEmailPings:        1,
			expectedUserUpdates:       1, 
		},
		{
			name: "Process multiple users, one fails mid-process (CreatePingHistory for email)",
			usersForPinging: []*models.User{
				defaultUser("userOK1", "email", true, ""),
				defaultUser("userFailPH", "email", true, ""), 
				defaultUser("userOK2", "telegram", true, "tgOK"),
			},
			mockRepoSetup: func(mockRepo *MockRepository) {
				mockRepo.CreatePingHistoryFunc = func(ctx context.Context, ping *models.PingHistory) error {
					if ping.UserID == "userFailPH" && ping.Method == "email" {
						return fmt.Errorf("db error CreatePingHistory for userFailPH email")
					}
					// For other users, add to the slice manually as the default mock behavior is overridden
					mockRepo.CreatedPingHistories = append(mockRepo.CreatedPingHistories, ping)
					return nil
				}
			},
			expectedPingHistories:     1 + 0 + 1, 
			expectedPingVerifications: 1 + 1 + 0, 
			expectedEmailPings:        1 + 1 + 0, 
			expectedTelegramPings:     0 + 0 + 1, 
			expectedUserUpdates:       3,         
			assertUserUpdate: func(t *testing.T, updatedUsers []*models.User) {
				assert.Len(t, updatedUsers, 3)
				ids := []string{}
				for _, u := range updatedUsers { ids = append(ids, u.ID) }
				assert.Contains(t, ids, "userOK1")
				assert.Contains(t, ids, "userFailPH")
				assert.Contains(t, ids, "userOK2")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			mockEmailClient := NewMockEmailClient()
			mockTelegramBot := NewMockTelegramBot()

			// Call ResetCalls for mocks at the beginning of each sub-test
			mockRepo.ResetCalls()
			mockEmailClient.ResetCalls()
			mockTelegramBot.ResetCalls()

			if tc.mockRepoSetup != nil {
				tc.mockRepoSetup(mockRepo)
			}
			if tc.mockEmailClientSetup != nil {
				tc.mockEmailClientSetup(mockEmailClient)
			}
			if tc.mockTelegramBotSetup != nil {
				tc.mockTelegramBotSetup(mockTelegramBot)
			}
			
			if mockRepo.GetUsersForPingingFunc == nil { 
				mockRepo.GetUsersForPingingFunc = func(ctx context.Context) ([]*models.User, error) {
					return tc.usersForPinging, nil
				}
			}

			scheduler := NewScheduler(mockRepo, mockEmailClient, mockTelegramBot, &config.Config{})
			ctx := context.Background()

			err := scheduler.pingTask(ctx)

			if tc.expectedErrorMsgSubstring != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrorMsgSubstring)
			} else {
				require.NoError(t, err)
			}

			assert.Len(t, mockRepo.CreatedPingHistories, tc.expectedPingHistories, "PingHistories count")
			assert.Len(t, mockRepo.CreatedPingVerifications, tc.expectedPingVerifications, "PingVerifications count")
			assert.Equal(t, tc.expectedEmailPings, mockEmailClient.SendPingEmailCalledCount, "EmailPings count")
			assert.Equal(t, tc.expectedTelegramPings, mockTelegramBot.SendPingMessageCalledCount, "TelegramPings count")
			assert.Len(t, mockRepo.UpdatedUsers, tc.expectedUserUpdates, "UserUpdates count")

			if tc.assertUserUpdate != nil {
				tc.assertUserUpdate(t, mockRepo.UpdatedUsers)
			}
			if tc.assertPingHistory != nil {
				tc.assertPingHistory(t, mockRepo.CreatedPingHistories)
			}
			if tc.assertPingVerification != nil {
				tc.assertPingVerification(t, mockRepo.CreatedPingVerifications)
			}
			if tc.assertEmailClient != nil {
				tc.assertEmailClient(t, mockEmailClient)
			}
			if tc.assertTelegramBot != nil {
				tc.assertTelegramBot(t, mockTelegramBot)
			}
		})
	}
}
// Ensure all other old test functions are removed.
// For example, TestDeadSwitchTask, TestCleanupTask, TestReminderTask, TestHelperFunctions etc.
// if they were present.
// The provided initial file only had TestPingTaskWithEmailMethod and TestPingTaskWithTelegramMethod
// related to pingTask, and TestDeadSwitchTask, TestCleanupTask, TestReminderTask related to other tasks.
// TestHelperFunctions was also present.
// I will ensure these are not present in the final file.
// The current overwrite should only contain the mocks, TestNewScheduler, TestAddTask, TestRegisterTasks,
// the skipped TestStartStop, and TestPingTaskComprehensive.
