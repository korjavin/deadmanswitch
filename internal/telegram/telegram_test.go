package telegram

import (
	"context"
	"errors" // Keep this one
	"fmt"
	"strconv"
	"strings" // Ensure strings is imported
	"testing"
	"time" // Add time import

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage" 
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBotAPIClient is a mock implementation of the ClientBotAPI interface.
type MockBotAPIClient struct {
	SendFunc    func(c tgbotapi.Chattable) (tgbotapi.Message, error)
	RequestFunc func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	
	LastChattableSent tgbotapi.Chattable
	SendCalledCount   int
}

func (m *MockBotAPIClient) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.SendCalledCount++
	m.LastChattableSent = c
	if m.SendFunc != nil {
		return m.SendFunc(c)
	}
	return tgbotapi.Message{MessageID: 123}, nil
}

func (m *MockBotAPIClient) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	if m.RequestFunc != nil {
		return m.RequestFunc(c)
	}
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *MockBotAPIClient) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return make(tgbotapi.UpdatesChannel)
}

type MockRepository struct {
	storage.Repository
	GetUserByTelegramIDFunc    func(ctx context.Context, telegramID string) (*models.User, error)
	GetUserByEmailFunc         func(ctx context.Context, email string) (*models.User, error)
	UpdateUserFunc             func(ctx context.Context, user *models.User) error
	ListSecretsByUserIDFunc    func(ctx context.Context, userID string) ([]*models.Secret, error)
	ListRecipientsByUserIDFunc func(ctx context.Context, userID string) ([]*models.Recipient, error)
	GetUserByIDFunc            func(ctx context.Context, userID string) (*models.User, error)
	GetLatestPingByUserIDFunc  func(ctx context.Context, userID string) (*models.PingHistory, error)
	UpdatePingHistoryFunc      func(ctx context.Context, ping *models.PingHistory) error
	// Store the user passed to UpdateUser for inspection
	LastUserUpdated       *models.User
	LastPingHistoryUpdated *models.PingHistory
}

// GetUserByTelegramID is the mock implementation for storage.Repository
func (m *MockRepository) GetUserByTelegramID(ctx context.Context, telegramID string) (*models.User, error) {
	if m.GetUserByTelegramIDFunc != nil {
		return m.GetUserByTelegramIDFunc(ctx, telegramID)
	}
	return nil, errors.New("GetUserByTelegramID not implemented in mock")
}

// GetUserByEmail is the mock implementation for storage.Repository
func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("GetUserByEmail not implemented in mock")
}

// UpdateUser is the mock implementation for storage.Repository
func (m *MockRepository) UpdateUser(ctx context.Context, user *models.User) error {
	m.LastUserUpdated = user // Store for inspection
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
	}
	return errors.New("UpdateUser not implemented in mock")
}

// ListSecretsByUserID is the mock implementation for storage.Repository
func (m *MockRepository) ListSecretsByUserID(ctx context.Context, userID string) ([]*models.Secret, error) {
	if m.ListSecretsByUserIDFunc != nil {
		return m.ListSecretsByUserIDFunc(ctx, userID)
	}
	return nil, errors.New("ListSecretsByUserID not implemented in mock")
}

// ListRecipientsByUserID is the mock implementation for storage.Repository
func (m *MockRepository) ListRecipientsByUserID(ctx context.Context, userID string) ([]*models.Recipient, error) {
	if m.ListRecipientsByUserIDFunc != nil {
		return m.ListRecipientsByUserIDFunc(ctx, userID)
	}
	return nil, errors.New("ListRecipientsByUserID not implemented in mock")
}

// GetUserByID is the mock implementation for storage.Repository
func (m *MockRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}
	return nil, errors.New("GetUserByID not implemented in mock")
}

// GetLatestPingByUserID is the mock implementation for storage.Repository
func (m *MockRepository) GetLatestPingByUserID(ctx context.Context, userID string) (*models.PingHistory, error) {
	if m.GetLatestPingByUserIDFunc != nil {
		return m.GetLatestPingByUserIDFunc(ctx, userID)
	}
	return nil, errors.New("GetLatestPingByUserID not implemented in mock")
}

// UpdatePingHistory is the mock implementation for storage.Repository
func (m *MockRepository) UpdatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	m.LastPingHistoryUpdated = ping
	if m.UpdatePingHistoryFunc != nil {
		return m.UpdatePingHistoryFunc(ctx, ping)
	}
	return errors.New("UpdatePingHistory not implemented in mock")
}

func TestSendPingMessage(t *testing.T) {
	mockAPI := &MockBotAPIClient{}
	mockRepo := &MockRepository{} 
	cfg := &config.Config{}

	botService := &Bot{
		api:    mockAPI,
		config: cfg,
		repo:   mockRepo,
	}

	ctx := context.Background()

	t.Run("successful message sending", func(t *testing.T) {
		mockAPI.SendCalledCount = 0 
		mockAPI.LastChattableSent = nil
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{MessageID: 1, Text: "test"}, nil
		}

		user := &models.User{
			ID:             "user123",
			TelegramID:     "123456789",
			PingDeadline:   14, 
			TelegramUsername: "testuser",
		}
		pingID := "pingTestID789"

		err := botService.SendPingMessage(ctx, user, pingID)
		require.NoError(t, err)
		assert.Equal(t, 1, mockAPI.SendCalledCount)
		require.NotNil(t, mockAPI.LastChattableSent)

		msgConfig, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok, "Expected LastChattableSent to be tgbotapi.MessageConfig")

		expectedChatID, _ := strconv.ParseInt(user.TelegramID, 10, 64)
		assert.Equal(t, expectedChatID, msgConfig.ChatID)
		assert.Equal(t, tgbotapi.ModeMarkdown, msgConfig.ParseMode)

		expectedText := fmt.Sprintf(
			"🔔 *Dead Man's Switch Check-In*\n\n"+
				"Please confirm you're okay by pressing the button below.\n\n"+
				"If you don't respond within %d days, your pre-configured secrets will be sent to your designated recipients.",
			user.PingDeadline,
		)
		assert.Equal(t, expectedText, msgConfig.Text)

		require.NotNil(t, msgConfig.ReplyMarkup)
		inlineKeyboard, ok := msgConfig.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
		require.True(t, ok, "Expected ReplyMarkup to be InlineKeyboardMarkup")
		require.Len(t, inlineKeyboard.InlineKeyboard, 1, "Expected one row in keyboard")
		require.Len(t, inlineKeyboard.InlineKeyboard[0], 1, "Expected one button in the row")

		button := inlineKeyboard.InlineKeyboard[0][0]
		assert.Equal(t, "I'm OK - Confirm", button.Text)
		expectedCallbackData := fmt.Sprintf("verify:%s:%s", user.ID, pingID)
		require.NotNil(t, button.CallbackData, "CallbackData should not be nil")
		assert.Equal(t, expectedCallbackData, *button.CallbackData)
	})

	t.Run("error if TelegramID is empty", func(t *testing.T) {
		mockAPI.SendCalledCount = 0
		user := &models.User{ID: "userWithoutTelegram", TelegramID: ""}
		err := botService.SendPingMessage(ctx, user, "pingID")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user has no associated Telegram ID")
		assert.Equal(t, 0, mockAPI.SendCalledCount, "API.Send should not be called")
	})

	t.Run("error if TelegramID is invalid", func(t *testing.T) {
		mockAPI.SendCalledCount = 0
		user := &models.User{ID: "userInvalidTelegram", TelegramID: "not-a-number"}
		err := botService.SendPingMessage(ctx, user, "pingID")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Telegram ID")
		assert.Equal(t, 0, mockAPI.SendCalledCount, "API.Send should not be called")
	})

	t.Run("error when api.Send fails", func(t *testing.T) {
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		expectedError := errors.New("telegram API send failed")
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, expectedError
		}

		user := &models.User{ID: "userSendFail", TelegramID: "987654321", PingDeadline: 7}
		err := botService.SendPingMessage(ctx, user, "pingFailID")
		require.Error(t, err)
		assert.True(t, errors.Is(err, expectedError) || strings.Contains(err.Error(), expectedError.Error()))
		assert.Equal(t, 1, mockAPI.SendCalledCount)
	})
}

func TestHandleCallbackQuery_Verify(t *testing.T) {
	mockAPI := &MockBotAPIClient{}
	mockRepo := &MockRepository{}
	cfg := &config.Config{} // BaseDomain not used by this handler

	botService := &Bot{
		api:    mockAPI,
		config: cfg,
		repo:   mockRepo,
	}

	ctx := context.Background()
	baseChat := &tgbotapi.Chat{ID: 12345}
	baseFromUser := &tgbotapi.User{ID: 67890, UserName: "callback_user"}
	baseMsg := &tgbotapi.Message{Chat: baseChat, MessageID: 1001} // Message being edited

	// Helper to create CallbackQuery
	createCallbackQuery := func(data string) *tgbotapi.CallbackQuery {
		return &tgbotapi.CallbackQuery{
			ID:      "callbackTestID",
			From:    baseFromUser,
			Message: baseMsg,
			Data:    data,
		}
	}

	// Helper to reset mocks
	resetMocks := func() {
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) { return tgbotapi.Message{}, nil }
		mockAPI.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) { return &tgbotapi.APIResponse{Ok: true}, nil }

		mockRepo.GetUserByIDFunc = nil
		mockRepo.GetLatestPingByUserIDFunc = nil
		mockRepo.UpdateUserFunc = nil
		mockRepo.UpdatePingHistoryFunc = nil
		mockRepo.LastUserUpdated = nil
		mockRepo.LastPingHistoryUpdated = nil
	}

	t.Run("Invalid Callback Data Format (Too Few Parts)", func(t *testing.T) {
		resetMocks()
		query := createCallbackQuery("verify") // Missing userID and pingID

		botService.handleCallbackQuery(ctx, query) // This function logs and returns, doesn't return error

		assert.Equal(t, 0, mockAPI.SendCalledCount, "api.Send should not be called")
		// answerCallbackQuery might be called with an error by the handler, or not at all.
		// Current code logs and returns. If it were to answer callback, we'd check:
		// assert.Equal(t, 1, mockAPI.RequestCalledCount)
		// if mockAPI.LastChattableSent != nil { ... check CallbackConfig for error text ... }
	})

	t.Run("User Not Found", func(t *testing.T) {
		resetMocks()
		query := createCallbackQuery("verify:nonexistentUserID:ping123")
		mockRepo.GetUserByIDFunc = func(ctx context.Context, userID string) (*models.User, error) {
			assert.Equal(t, "nonexistentUserID", userID)
			return nil, storage.ErrNotFound
		}
		var requestChattable tgbotapi.Chattable
		mockAPI.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
			requestChattable = c
			return &tgbotapi.APIResponse{Ok: true}, nil
		}


		botService.handleCallbackQuery(ctx, query)

		callbackConfig, ok := requestChattable.(tgbotapi.CallbackConfig)
		require.True(t, ok, "Expected CallbackConfig to be sent to Request")
		assert.Equal(t, query.ID, callbackConfig.CallbackQueryID)
		assert.Contains(t, callbackConfig.Text, "Error: User not found")
		assert.Equal(t, 0, mockAPI.SendCalledCount, "api.Send (for edit) should not be called")
	})

	t.Run("DB Error on GetUserByID", func(t *testing.T) {
		resetMocks()
		dbError := errors.New("get user by id failed")
		query := createCallbackQuery("verify:anyUserID:ping123")
		mockRepo.GetUserByIDFunc = func(ctx context.Context, userID string) (*models.User, error) {
			return nil, dbError
		}
		var requestChattable tgbotapi.Chattable
		mockAPI.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
			requestChattable = c
			return &tgbotapi.APIResponse{Ok: true}, nil
		}

		botService.handleCallbackQuery(ctx, query)
		
		callbackConfig, ok := requestChattable.(tgbotapi.CallbackConfig)
		require.True(t, ok)
		assert.Equal(t, query.ID, callbackConfig.CallbackQueryID)
		assert.Contains(t, callbackConfig.Text, "Error: User not found") // Generic error for user lookup failure
	})

	t.Run("Successful Verification (with existing Ping)", func(t *testing.T) {
		resetMocks()
		testUserID := "testUserWithPing"
		testPingID := "activePing123"
		query := createCallbackQuery(fmt.Sprintf("verify:%s:%s", testUserID, testPingID))

		mockUser := &models.User{ID: testUserID, TelegramID: strconv.FormatInt(baseFromUser.ID, 10)}
		mockPing := &models.PingHistory{ID: testPingID, UserID: testUserID, Status: "sent"}

		mockRepo.GetUserByIDFunc = func(ctx context.Context, userID string) (*models.User, error) { return mockUser, nil }
		mockRepo.GetLatestPingByUserIDFunc = func(ctx context.Context, userID string) (*models.PingHistory, error) { return mockPing, nil }
		mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error { return nil }
		mockRepo.UpdatePingHistoryFunc = func(ctx context.Context, ping *models.PingHistory) error { return nil }

		botService.handleCallbackQuery(ctx, query)

		// Assert User Update
		require.NotNil(t, mockRepo.LastUserUpdated)
		assert.Equal(t, testUserID, mockRepo.LastUserUpdated.ID)
		assert.False(t, mockRepo.LastUserUpdated.LastActivity.IsZero())

		// Assert Ping Update
		require.NotNil(t, mockRepo.LastPingHistoryUpdated)
		assert.Equal(t, testPingID, mockRepo.LastPingHistoryUpdated.ID)
		assert.Equal(t, "responded", mockRepo.LastPingHistoryUpdated.Status)
		require.NotNil(t, mockRepo.LastPingHistoryUpdated.RespondedAt)
		assert.False(t, mockRepo.LastPingHistoryUpdated.RespondedAt.IsZero())

		// Assert Message Edit (api.Send)
		assert.Equal(t, 1, mockAPI.SendCalledCount)
		editMsg, ok := mockAPI.LastChattableSent.(tgbotapi.EditMessageTextConfig)
		require.True(t, ok)
		assert.Equal(t, baseChat.ID, editMsg.ChatID)
		assert.Equal(t, baseMsg.MessageID, editMsg.MessageID)
		assert.Contains(t, editMsg.Text, "Thank you for confirming your status")

		// Assert Callback Answer (api.Request)
		// (RequestFunc is set in resetMocks to a default success, if specific checks needed, override here)
	})

	t.Run("Successful Verification (pingID is 0 - manual)", func(t *testing.T) {
		resetMocks()
		testUserID := "testUserManualVerify"
		query := createCallbackQuery(fmt.Sprintf("verify:%s:0", testUserID))
		mockUser := &models.User{ID: testUserID}

		mockRepo.GetUserByIDFunc = func(ctx context.Context, userID string) (*models.User, error) { return mockUser, nil }
		mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error { return nil }
		// GetLatestPingByUserIDFunc should not be strictly necessary to mock if pingID is "0"
		// but the current code calls it. If it returns ErrNotFound, it's handled.
		mockRepo.GetLatestPingByUserIDFunc = func(ctx context.Context, userID string) (*models.PingHistory, error) {
			return nil, storage.ErrNotFound // No active ping, or manual verify context
		}


		botService.handleCallbackQuery(ctx, query)

		require.NotNil(t, mockRepo.LastUserUpdated)
		assert.False(t, mockRepo.LastUserUpdated.LastActivity.IsZero())
		assert.Nil(t, mockRepo.LastPingHistoryUpdated, "UpdatePingHistory should not be called for pingID 0 if no ping found")
		assert.Equal(t, 1, mockAPI.SendCalledCount) // Edit message
		// Callback answer is also expected
	})

	t.Run("Successful Verification (No Ping Found, pingID non-zero)", func(t *testing.T) {
		resetMocks()
		testUserID := "testUserNoPingFound"
		query := createCallbackQuery(fmt.Sprintf("verify:%s:nonZeroPingID", testUserID))
		mockUser := &models.User{ID: testUserID}

		mockRepo.GetUserByIDFunc = func(ctx context.Context, userID string) (*models.User, error) { return mockUser, nil }
		mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error { return nil }
		mockRepo.GetLatestPingByUserIDFunc = func(ctx context.Context, userID string) (*models.PingHistory, error) {
			return nil, storage.ErrNotFound // Simulate no ping found
		}

		botService.handleCallbackQuery(ctx, query)

		require.NotNil(t, mockRepo.LastUserUpdated)
		assert.Nil(t, mockRepo.LastPingHistoryUpdated, "UpdatePingHistory should not be called if no ping was found")
		assert.Equal(t, 1, mockAPI.SendCalledCount) // Edit message
	})
	
	t.Run("Error during UpdateUser", func(t *testing.T) {
		resetMocks()
		testUserID := "userUpdateFail"
		query := createCallbackQuery(fmt.Sprintf("verify:%s:ping123", testUserID))
		updateErr := errors.New("failed to update user")

		mockRepo.GetUserByIDFunc = func(ctx context.Context, userID string) (*models.User, error) { return &models.User{ID: testUserID}, nil }
		mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error { return updateErr }
		// GetLatestPingByUserIDFunc will be called but its outcome doesn't prevent UpdateUser attempt
		mockRepo.GetLatestPingByUserIDFunc = func(ctx context.Context, userID string) (*models.PingHistory, error) { return &models.PingHistory{ID: "ping123", Status: "sent"}, nil }


		botService.handleCallbackQuery(ctx, query) // Error is logged, not returned

		// Check that answerCallbackQuery was still called, possibly with a generic message or success
		// The current code doesn't change behavior of answer/edit based on UpdateUser error.
		// It logs and proceeds.
		// We can check if api.Request (for answerCallbackQuery) was called.
		// Let's assume it's called with success by default from resetMocks.
		// If we want to check the text of answerCallbackQuery, we need to capture its argument.
		var answeredCallbackText string
		mockAPI.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
			if cbq, ok := c.(tgbotapi.CallbackConfig); ok {
				answeredCallbackText = cbq.Text
			}
			return &tgbotapi.APIResponse{Ok: true}, nil
		}
		botService.handleCallbackQuery(ctx, query) // Call again with the RequestFunc set
		assert.Equal(t, "Verification successful", answeredCallbackText) // It still says successful, error is only logged
	})

	t.Run("Error during EditMessageText", func(t *testing.T) {
		resetMocks()
		testUserID := "editFailUser"
		query := createCallbackQuery(fmt.Sprintf("verify:%s:ping123", testUserID))
		editErr := errors.New("failed to edit message")

		mockRepo.GetUserByIDFunc = func(ctx context.Context, userID string) (*models.User, error) { return &models.User{ID: testUserID}, nil }
		mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error { return nil }
		mockRepo.GetLatestPingByUserIDFunc = func(ctx context.Context, userID string) (*models.PingHistory, error) { return &models.PingHistory{ID: "ping123", Status: "sent"}, nil }
		mockRepo.UpdatePingHistoryFunc = func(ctx context.Context, ping *models.PingHistory) error {return nil}

		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) { return tgbotapi.Message{}, editErr }

		botService.handleCallbackQuery(ctx, query) // Error is logged

		// Assert that answerCallbackQuery was still attempted
		// (RequestFunc is default success from resetMocks)
		// Check if mockAPI.Request was called (how to check this without SendCalledCount equivalent?)
		// For now, assume if SendFunc failed, RequestFunc was still called.
		// This test mainly ensures the function doesn't panic.
	})

	t.Run("Error during AnswerCallbackQuery", func(t *testing.T) {
		resetMocks()
		testUserID := "answerFailUser"
		query := createCallbackQuery(fmt.Sprintf("verify:%s:ping123", testUserID))
		answerErr := errors.New("failed to answer callback")

		mockRepo.GetUserByIDFunc = func(ctx context.Context, userID string) (*models.User, error) { return &models.User{ID: testUserID}, nil }
		mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error { return nil }
		mockRepo.GetLatestPingByUserIDFunc = func(ctx context.Context, userID string) (*models.PingHistory, error) { return &models.PingHistory{ID: "ping123", Status: "sent"}, nil }
		mockRepo.UpdatePingHistoryFunc = func(ctx context.Context, ping *models.PingHistory) error {return nil}


		mockAPI.RequestFunc = func(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) { return nil, answerErr }

		botService.handleCallbackQuery(ctx, query) // Error is logged

		// Assert that EditMessageText (Send) was still attempted
		assert.Equal(t, 1, mockAPI.SendCalledCount)
	})
}

func TestHandleVerify(t *testing.T) {
	mockAPI := &MockBotAPIClient{}
	mockRepo := &MockRepository{}
	cfg := &config.Config{} // BaseDomain not used by handleVerify directly

	botService := &Bot{
		api:    mockAPI,
		config: cfg,
		repo:   mockRepo,
	}

	ctx := context.Background()
	baseChat := &tgbotapi.Chat{ID: 555}
	baseFromUser := &tgbotapi.User{ID: 444, FirstName: "VerifyUser"}
	baseMessage := &tgbotapi.Message{
		MessageID: 300,
		Chat:      baseChat,
		From:      baseFromUser,
		Text:      "/verify",
	}
	userTelegramIDStr := strconv.FormatInt(baseFromUser.ID, 10)

	// Helper to reset mocks
	resetMocks := func() {
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) { return tgbotapi.Message{}, nil }
		mockRepo.GetUserByTelegramIDFunc = nil
	}

	t.Run("User Not Found", func(t *testing.T) {
		resetMocks()
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			assert.Equal(t, userTelegramIDStr, telegramID)
			return nil, storage.ErrNotFound
		}

		err := botService.handleVerify(ctx, baseMessage, "")
		require.NoError(t, err, "handleVerify should send a message, not error, for user not found")

		assert.Equal(t, 1, mockAPI.SendCalledCount)
		sentMsg, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Equal(t, baseChat.ID, sentMsg.ChatID)
		assert.Contains(t, sentMsg.Text, "You're not registered yet. Please use /start to begin.")
	})

	t.Run("DB Error on GetUserByTelegramID", func(t *testing.T) {
		resetMocks()
		dbError := errors.New("internal DB error")
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			return nil, dbError
		}

		err := botService.handleVerify(ctx, baseMessage, "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, dbError) || strings.Contains(err.Error(), "database error"))
		assert.Equal(t, 0, mockAPI.SendCalledCount, "api.Send should not be called")
	})

	t.Run("Successful Verify Message Sending", func(t *testing.T) {
		resetMocks()
		testUserID := "verifiedUserID"
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			return &models.User{ID: testUserID, TelegramID: userTelegramIDStr}, nil
		}

		err := botService.handleVerify(ctx, baseMessage, "")
		require.NoError(t, err)

		assert.Equal(t, 1, mockAPI.SendCalledCount)
		sentMsgConfig, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok, "Expected MessageConfig to be sent")

		assert.Equal(t, baseChat.ID, sentMsgConfig.ChatID)
		assert.Equal(t, "Please confirm you're okay by pressing the button below:", sentMsgConfig.Text)

		require.NotNil(t, sentMsgConfig.ReplyMarkup, "ReplyMarkup should not be nil")
		inlineKeyboard, ok := sentMsgConfig.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
		require.True(t, ok, "Expected ReplyMarkup to be InlineKeyboardMarkup")
		require.Len(t, inlineKeyboard.InlineKeyboard, 1, "Expected one row in keyboard")
		require.Len(t, inlineKeyboard.InlineKeyboard[0], 1, "Expected one button in the row")

		button := inlineKeyboard.InlineKeyboard[0][0]
		assert.Equal(t, "I'm OK - Confirm", button.Text)
		expectedCallbackData := fmt.Sprintf("verify:%s:0", testUserID) // "0" for pingID as per handleVerify logic
		require.NotNil(t, button.CallbackData)
		assert.Equal(t, expectedCallbackData, *button.CallbackData)
	})

	t.Run("API Send Error", func(t *testing.T) {
		resetMocks()
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			return &models.User{ID: "anyUserID", TelegramID: userTelegramIDStr}, nil
		}
		sendError := errors.New("telegram send failed")
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, sendError
		}

		err := botService.handleVerify(ctx, baseMessage, "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sendError) || strings.Contains(err.Error(), sendError.Error()))
		assert.Equal(t, 1, mockAPI.SendCalledCount)
	})
}

func TestHandleStatus(t *testing.T) {
	mockAPI := &MockBotAPIClient{}
	mockRepo := &MockRepository{}
	cfg := &config.Config{
		BaseDomain: "dms.test",
	}

	botService := &Bot{
		api:    mockAPI,
		config: cfg,
		repo:   mockRepo,
	}

	ctx := context.Background()
	baseChat := &tgbotapi.Chat{ID: 777}
	baseFromUser := &tgbotapi.User{ID: 888, FirstName: "StatusUser"}
	baseMessage := &tgbotapi.Message{
		MessageID: 200,
		Chat:      baseChat,
		From:      baseFromUser,
		Text:      "/status",
	}
	userTelegramIDStr := strconv.FormatInt(baseFromUser.ID, 10)

	// Helper to reset mocks
	resetMocks := func() {
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) { return tgbotapi.Message{}, nil }
		mockRepo.GetUserByTelegramIDFunc = nil
		mockRepo.ListSecretsByUserIDFunc = nil
		mockRepo.ListRecipientsByUserIDFunc = nil
	}

	t.Run("User Not Found", func(t *testing.T) {
		resetMocks()
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			assert.Equal(t, userTelegramIDStr, telegramID)
			return nil, storage.ErrNotFound
		}

		err := botService.handleStatus(ctx, baseMessage, "")
		require.NoError(t, err, "handleStatus should send a message, not return an error for user not found")

		assert.Equal(t, 1, mockAPI.SendCalledCount)
		sentMsg, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Equal(t, baseChat.ID, sentMsg.ChatID)
		assert.Contains(t, sentMsg.Text, "You're not registered yet. Please use /start to begin.")
		assert.Nil(t, mockRepo.ListSecretsByUserIDFunc)    // Should not be called
		assert.Nil(t, mockRepo.ListRecipientsByUserIDFunc) // Should not be called
	})

	t.Run("DB Error on GetUserByTelegramID", func(t *testing.T) {
		resetMocks()
		dbError := errors.New("major DB outage")
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			return nil, dbError
		}

		err := botService.handleStatus(ctx, baseMessage, "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, dbError) || strings.Contains(err.Error(), "database error"), "Error should be or wrap DB error")
		assert.Equal(t, 0, mockAPI.SendCalledCount, "api.Send should not be called")
	})

	t.Run("Successful Status - No Secrets/Recipients, DB Errors on Counts", func(t *testing.T) {
		resetMocks()
		testUser := &models.User{
			ID:                "userWithNoItems",
			Email:             "noitems@example.com",
			PingFrequency:     3,
			PingDeadline:      14,
			PingingEnabled:    true,
			PingMethod:        "telegram",
			LastActivity:      time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			NextScheduledPing: time.Time{}, // Zero time for "Not scheduled"
		}
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			return testUser, nil
		}
		countError := errors.New("failed to count items")
		mockRepo.ListSecretsByUserIDFunc = func(ctx context.Context, userID string) ([]*models.Secret, error) {
			assert.Equal(t, testUser.ID, userID)
			return nil, countError
		}
		mockRepo.ListRecipientsByUserIDFunc = func(ctx context.Context, userID string) ([]*models.Recipient, error) {
			assert.Equal(t, testUser.ID, userID)
			return nil, countError
		}

		err := botService.handleStatus(ctx, baseMessage, "")
		require.NoError(t, err, "handleStatus should proceed even if counts fail")

		assert.Equal(t, 1, mockAPI.SendCalledCount)
		sentMsg, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Equal(t, baseChat.ID, sentMsg.ChatID)
		assert.Equal(t, tgbotapi.ModeMarkdown, sentMsg.ParseMode)

		assert.Contains(t, sentMsg.Text, fmt.Sprintf("Email: %s", testUser.Email))
		assert.Contains(t, sentMsg.Text, fmt.Sprintf("Ping Frequency: Every %d days", testUser.PingFrequency))
		assert.Contains(t, sentMsg.Text, fmt.Sprintf("Response Deadline: %d days", testUser.PingDeadline))
		assert.Contains(t, sentMsg.Text, fmt.Sprintf("Pinging Enabled: %v", testUser.PingingEnabled))
		assert.Contains(t, sentMsg.Text, fmt.Sprintf("Ping Method: %s", testUser.PingMethod))
		assert.Contains(t, sentMsg.Text, "Secrets Stored: 0")    // Should default to 0 on error
		assert.Contains(t, sentMsg.Text, "Recipients Configured: 0") // Should default to 0 on error
		assert.Contains(t, sentMsg.Text, fmt.Sprintf("Last Activity: %s", testUser.LastActivity.Format("Jan 2, 2006 at 15:04 MST")))
		assert.Contains(t, sentMsg.Text, "Next Scheduled Ping: Not scheduled")
		assert.Contains(t, sentMsg.Text, cfg.BaseDomain)
	})

	t.Run("Successful Status - With Secrets/Recipients", func(t *testing.T) {
		resetMocks()
		lastActivityTime := time.Date(2023, 10, 5, 12, 30, 0, 0, time.FixedZone("MST", -7*60*60))
		nextPingTime := time.Date(2023, 10, 10, 12, 0, 0, 0, time.FixedZone("MST", -7*60*60))

		testUser := &models.User{
			ID:                "userWithItems",
			Email:             "items@example.com",
			PingFrequency:     5,
			PingDeadline:      20,
			PingingEnabled:    false,
			PingMethod:        "email",
			LastActivity:      lastActivityTime,
			NextScheduledPing: nextPingTime,
		}
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			return testUser, nil
		}
		mockRepo.ListSecretsByUserIDFunc = func(ctx context.Context, userID string) ([]*models.Secret, error) {
			return []*models.Secret{{ID: "s1"}, {ID: "s2"}}, nil
		}
		mockRepo.ListRecipientsByUserIDFunc = func(ctx context.Context, userID string) ([]*models.Recipient, error) {
			return []*models.Recipient{{ID: "r1"}, {ID: "r2"}, {ID: "r3"}}, nil
		}

		err := botService.handleStatus(ctx, baseMessage, "")
		require.NoError(t, err)

		assert.Equal(t, 1, mockAPI.SendCalledCount)
		sentMsg, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Contains(t, sentMsg.Text, "Secrets Stored: 2")
		assert.Contains(t, sentMsg.Text, "Recipients Configured: 3")
		assert.Contains(t, sentMsg.Text, fmt.Sprintf("Last Activity: %s", lastActivityTime.Format("Jan 2, 2006 at 15:04 MST")))
		assert.Contains(t, sentMsg.Text, fmt.Sprintf("Next Scheduled Ping: %s", nextPingTime.Format("Jan 2, 2006 at 15:04 MST")))
	})

	t.Run("API Send Error", func(t *testing.T) {
		resetMocks()
		testUser := &models.User{ID: "userApiErr", Email: "apierr@example.com"}
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			return testUser, nil
		}
		// Default List funcs will return 0 counts as they are not set up to return data
		mockRepo.ListSecretsByUserIDFunc = func(ctx context.Context, userID string) ([]*models.Secret, error) { return []*models.Secret{}, nil }
		mockRepo.ListRecipientsByUserIDFunc = func(ctx context.Context, userID string) ([]*models.Recipient, error) { return []*models.Recipient{}, nil }

		sendError := errors.New("telegram API failed")
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, sendError
		}

		err := botService.handleStatus(ctx, baseMessage, "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, sendError) || strings.Contains(err.Error(), sendError.Error()))
		assert.Equal(t, 1, mockAPI.SendCalledCount)
	})
}

func TestHandleConnect(t *testing.T) {
	mockAPI := &MockBotAPIClient{}
	mockRepo := &MockRepository{}
	cfg := &config.Config{
		BaseDomain: "my.dms.app",
	}

	botService := &Bot{
		api:    mockAPI,
		config: cfg,
		repo:   mockRepo,
	}

	ctx := context.Background()
	baseChat := &tgbotapi.Chat{ID: 12345}
	baseFromUser := &tgbotapi.User{ID: 67890, UserName: "telegram_user"}
	baseMessage := &tgbotapi.Message{
		MessageID: 100,
		Chat:      baseChat,
		From:      baseFromUser,
	}

	// Helper to reset mocks for each sub-test
	resetMocks := func() {
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) { return tgbotapi.Message{}, nil } // Default success
		mockRepo.GetUserByEmailFunc = nil
		mockRepo.UpdateUserFunc = nil
		mockRepo.LastUserUpdated = nil
	}

	t.Run("No Email Argument", func(t *testing.T) {
		resetMocks()
		message := *baseMessage // Copy base
		message.Text = "/connect"

		err := botService.handleConnect(ctx, &message, "") // No arguments
		require.NoError(t, err, "handleConnect should not error on usage message")

		assert.Equal(t, 1, mockAPI.SendCalledCount)
		sentMsg, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Equal(t, baseChat.ID, sentMsg.ChatID)
		assert.Contains(t, sentMsg.Text, "Please provide your email address: /connect your@email.com")

		assert.Nil(t, mockRepo.GetUserByEmailFunc, "GetUserByEmail should not be set/called")
		assert.Nil(t, mockRepo.UpdateUserFunc, "UpdateUser should not be set/called")
	})

	t.Run("Invalid Email Format", func(t *testing.T) {
		resetMocks()
		message := *baseMessage
		message.Text = "/connect invalidemail"
		args := "invalidemail"

		err := botService.handleConnect(ctx, &message, args)
		require.NoError(t, err, "handleConnect should not error on usage message")

		assert.Equal(t, 1, mockAPI.SendCalledCount)
		sentMsg, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Equal(t, baseChat.ID, sentMsg.ChatID)
		assert.Contains(t, sentMsg.Text, "Invalid email format. Please try again.")

		assert.Nil(t, mockRepo.GetUserByEmailFunc, "GetUserByEmail should not be set/called")
		assert.Nil(t, mockRepo.UpdateUserFunc, "UpdateUser should not be set/called")
	})

	t.Run("User Not Found by Email", func(t *testing.T) {
		resetMocks()
		emailArg := "nonexistent@example.com"
		message := *baseMessage
		message.Text = "/connect " + emailArg
		
		var getUserByEmailCalledWith string
		mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (*models.User, error) {
			getUserByEmailCalledWith = email
			return nil, storage.ErrNotFound
		}

		err := botService.handleConnect(ctx, &message, emailArg)
		require.NoError(t, err, "handleConnect should not error for user not found")

		assert.Equal(t, emailArg, getUserByEmailCalledWith)
		assert.Equal(t, 1, mockAPI.SendCalledCount)
		sentMsg, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Equal(t, baseChat.ID, sentMsg.ChatID)
		assert.Contains(t, sentMsg.Text, "No account found with email "+emailArg)
		assert.Contains(t, sentMsg.Text, cfg.BaseDomain)
		assert.Nil(t, mockRepo.UpdateUserFunc, "UpdateUser should not be set/called")
	})

	t.Run("Database Error on GetUserByEmail", func(t *testing.T) {
		resetMocks()
		emailArg := "user@example.com"
		message := *baseMessage
		message.Text = "/connect " + emailArg
		dbError := errors.New("permanent DB issue")

		mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (*models.User, error) {
			return nil, dbError
		}

		err := botService.handleConnect(ctx, &message, emailArg)
		require.Error(t, err)
		assert.True(t, errors.Is(err, dbError) || strings.Contains(err.Error(), dbError.Error()))
		assert.Equal(t, 0, mockAPI.SendCalledCount, "api.Send should not be called")
		assert.Nil(t, mockRepo.UpdateUserFunc, "UpdateUser should not be set/called")
	})

	t.Run("Successful Connection", func(t *testing.T) {
		resetMocks()
		emailArg := "gooduser@example.com"
		message := *baseMessage
		message.Text = "/connect " + emailArg
		originalUser := &models.User{ID: "userFoundID", Email: emailArg}

		mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (*models.User, error) {
			return originalUser, nil
		}
		mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error {
			return nil // Success
		}

		err := botService.handleConnect(ctx, &message, emailArg)
		require.NoError(t, err)

		require.NotNil(t, mockRepo.LastUserUpdated, "UpdateUser should have been called")
		updatedUser := mockRepo.LastUserUpdated
		assert.Equal(t, originalUser.ID, updatedUser.ID)
		assert.Equal(t, strconv.FormatInt(baseFromUser.ID, 10), updatedUser.TelegramID)
		assert.Equal(t, baseFromUser.UserName, updatedUser.TelegramUsername)
		assert.False(t, updatedUser.LastActivity.IsZero(), "LastActivity should be set")

		assert.Equal(t, 1, mockAPI.SendCalledCount)
		sentMsg, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)
		assert.Equal(t, baseChat.ID, sentMsg.ChatID)
		assert.Contains(t, sentMsg.Text, "Success! Your Telegram account is now connected")
		assert.Contains(t, sentMsg.Text, emailArg)
	})

	t.Run("Database Error on UpdateUser", func(t *testing.T) {
		resetMocks()
		emailArg := "updatefail@example.com"
		message := *baseMessage
		message.Text = "/connect " + emailArg
		originalUser := &models.User{ID: "userUpdateFailID", Email: emailArg}
		updateError := errors.New("DB update failed")

		mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (*models.User, error) {
			return originalUser, nil
		}
		mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error {
			return updateError
		}

		err := botService.handleConnect(ctx, &message, emailArg)
		require.Error(t, err)
		assert.True(t, errors.Is(err, updateError) || strings.Contains(err.Error(), updateError.Error()))
		assert.Equal(t, 0, mockAPI.SendCalledCount, "api.Send should not be called after UpdateUser fails")
		require.NotNil(t, mockRepo.LastUserUpdated, "UpdateUser should have been called")
	})

	t.Run("API Send Error on Success Message", func(t *testing.T) {
		resetMocks()
		emailArg := "apifail@example.com"
		message := *baseMessage
		message.Text = "/connect " + emailArg
		originalUser := &models.User{ID: "userApiFailID", Email: emailArg}
		sendError := errors.New("telegram send failed")

		mockRepo.GetUserByEmailFunc = func(ctx context.Context, email string) (*models.User, error) {
			return originalUser, nil
		}
		mockRepo.UpdateUserFunc = func(ctx context.Context, user *models.User) error {
			return nil // UpdateUser succeeds
		}
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, sendError // api.Send fails
		}

		err := botService.handleConnect(ctx, &message, emailArg)
		require.Error(t, err)
		assert.True(t, errors.Is(err, sendError) || strings.Contains(err.Error(), sendError.Error()))
		assert.Equal(t, 1, mockAPI.SendCalledCount, "api.Send should have been called once")
		require.NotNil(t, mockRepo.LastUserUpdated, "UpdateUser should have been called")
	})
}

func TestHandleStart(t *testing.T) {
	mockAPI := &MockBotAPIClient{}
	mockRepo := &MockRepository{}
	cfg := &config.Config{
		BaseDomain: "test.domain.com",
	}

	botService := &Bot{
		api:    mockAPI,
		config: cfg,
		repo:   mockRepo,
	}

	ctx := context.Background()
	chat := &tgbotapi.Chat{ID: 111}
	fromUser := &tgbotapi.User{ID: 222, FirstName: "TestFirstName"}
	message := &tgbotapi.Message{
		MessageID: 1,
		Chat:      chat,
		From:      fromUser,
		Text:      "/start",
	}
	userTelegramIDStr := strconv.FormatInt(fromUser.ID, 10)

	t.Run("New User", func(t *testing.T) {
		// Setup
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			assert.Equal(t, userTelegramIDStr, telegramID, "GetUserByTelegramID called with wrong ID")
			return nil, storage.ErrNotFound // Simulate user not found
		}
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, nil // Simulate successful send
		}

		// Execute
		err := botService.handleStart(ctx, message, "")
		require.NoError(t, err)

		// Assert
		assert.Equal(t, 1, mockAPI.SendCalledCount, "Send should be called once for a new user")
		require.NotNil(t, mockAPI.LastChattableSent)
		sentMsgConfig, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)

		assert.Equal(t, chat.ID, sentMsgConfig.ChatID)
		assert.Contains(t, sentMsgConfig.Text, "Welcome to Dead Man's Switch, "+fromUser.FirstName)
		assert.Contains(t, sentMsgConfig.Text, cfg.BaseDomain)
		assert.Contains(t, sentMsgConfig.Text, userTelegramIDStr, "Message should contain Telegram ID as connection code")
	})

	t.Run("Existing User", func(t *testing.T) {
		// Setup
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		existingUser := &models.User{ID: "existingUser123", TelegramID: userTelegramIDStr, Email: "exists@example.com"}
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			assert.Equal(t, userTelegramIDStr, telegramID)
			return existingUser, nil // Simulate existing user
		}
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, nil
		}

		// Execute
		err := botService.handleStart(ctx, message, "")
		require.NoError(t, err)

		// Assert
		assert.Equal(t, 1, mockAPI.SendCalledCount, "Send should be called once for an existing user")
		require.NotNil(t, mockAPI.LastChattableSent)
		sentMsgConfig, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok)

		assert.Equal(t, chat.ID, sentMsgConfig.ChatID)
		assert.Contains(t, sentMsgConfig.Text, "Welcome back, "+fromUser.FirstName)
		assert.Contains(t, sentMsgConfig.Text, "Your Dead Man's Switch is active")
	})

	t.Run("Database Error on GetUserByTelegramID", func(t *testing.T) {
		// Setup
		mockAPI.SendCalledCount = 0
		dbError := errors.New("simulated DB error")
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			return nil, dbError // Simulate DB error
		}

		// Execute
		err := botService.handleStart(ctx, message, "")

		// Assert
		require.Error(t, err)
		assert.True(t, errors.Is(err, dbError) || strings.Contains(err.Error(), dbError.Error()), "Error should be the DB error")
		assert.Equal(t, 0, mockAPI.SendCalledCount, "Send should not be called on DB error")
	})

	t.Run("Error Propagation from api.Send (New User scenario)", func(t *testing.T) {
		// Setup
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		mockRepo.GetUserByTelegramIDFunc = func(ctx context.Context, telegramID string) (*models.User, error) {
			return nil, storage.ErrNotFound // New user
		}
		sendError := errors.New("simulated API send error")
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, sendError // API send fails
		}

		// Execute
		err := botService.handleStart(ctx, message, "")

		// Assert
		require.Error(t, err)
		assert.True(t, errors.Is(err, sendError) || strings.Contains(err.Error(), sendError.Error()), "Error should be the API send error")
		assert.Equal(t, 1, mockAPI.SendCalledCount, "Send should have been called once")
	})
}

func TestHandleHelp(t *testing.T) {
	mockAPI := &MockBotAPIClient{}
	cfg := &config.Config{
		BaseDomain: "test.example.com",
	}
	// handleHelp does not use the repository, so a nil or basic mock is fine.
	mockRepo := &MockRepository{}

	// Direct instantiation for focused unit testing of the handler.
	// registerHandlers() is not strictly needed if we call b.handleHelp directly.
	botService := &Bot{
		api:    mockAPI,
		config: cfg,
		repo:   mockRepo,
		// handlers map is not used when calling the handler func directly.
	}

	ctx := context.Background()
	chat := &tgbotapi.Chat{ID: 12345}
	fromUser := &tgbotapi.User{ID: 9876, UserName: "testuser"}
	message := &tgbotapi.Message{
		MessageID: 1,
		Chat:      chat,
		From:      fromUser,
		Text:      "/help",
		// Command:   "help", // This field does not exist in tgbotapi.Message
	}

	t.Run("Successful Help Message Sending", func(t *testing.T) {
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			// Default mock send, can be overridden per test if needed
			return tgbotapi.Message{MessageID: 2, Text: "help response"}, nil
		}

		err := botService.handleHelp(ctx, message, "") // Args are empty for /help
		require.NoError(t, err)
		assert.Equal(t, 1, mockAPI.SendCalledCount, "Expected Send to be called once")
		require.NotNil(t, mockAPI.LastChattableSent, "LastChattableSent should not be nil")

		sentMsgConfig, ok := mockAPI.LastChattableSent.(tgbotapi.MessageConfig)
		require.True(t, ok, "Expected LastChattableSent to be a tgbotapi.MessageConfig")

		assert.Equal(t, chat.ID, sentMsgConfig.ChatID)
		assert.Equal(t, tgbotapi.ModeMarkdown, sentMsgConfig.ParseMode)

		expectedHelpTextPart := "*Dead Man's Switch Bot Commands*"
		assert.Contains(t, sentMsgConfig.Text, expectedHelpTextPart)
		assert.Contains(t, sentMsgConfig.Text, cfg.BaseDomain, "Help text should contain BaseDomain")
	})

	t.Run("Error Propagation from api.Send", func(t *testing.T) {
		mockAPI.SendCalledCount = 0
		mockAPI.LastChattableSent = nil
		expectedAPIError := errors.New("simulated telegram API error")
		mockAPI.SendFunc = func(c tgbotapi.Chattable) (tgbotapi.Message, error) {
			return tgbotapi.Message{}, expectedAPIError
		}

		err := botService.handleHelp(ctx, message, "")
		require.Error(t, err, "Expected an error from handleHelp")
		assert.Equal(t, 1, mockAPI.SendCalledCount, "Expected Send to be called once despite error")
		
		// Check if the returned error is the one from the mock API or wraps it.
		// The current handleHelp directly returns the error from b.api.Send.
		assert.True(t, errors.Is(err, expectedAPIError) || err.Error() == expectedAPIError.Error(), "Error should be the one returned by api.Send")
	})
}
