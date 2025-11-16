package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot represents a Telegram bot service
type Bot struct {
	bot      *tgbotapi.BotAPI
	config   *config.Config
	repo     storage.Repository
	handlers map[string]CommandHandler
	updates  tgbotapi.UpdatesChannel
	mu       sync.RWMutex
}

// CommandHandler is a function that handles a telegram command
type CommandHandler func(ctx context.Context, message *tgbotapi.Message, args string) error

// NewBot creates a new Telegram bot
func NewBot(cfg *config.Config, repo storage.Repository) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	bot.Debug = cfg.Debug

	log.Printf("Authorized on Telegram bot account %s", bot.Self.UserName)

	// Store the bot username in the config
	cfg.TelegramBotUsername = "@" + bot.Self.UserName

	b := &Bot{
		bot:      bot,
		config:   cfg,
		repo:     repo,
		handlers: make(map[string]CommandHandler),
	}

	// Register command handlers
	b.registerHandlers()

	return b, nil
}

// registerHandlers registers the bot's command handlers
func (b *Bot) registerHandlers() {
	b.handlers = map[string]CommandHandler{
		"start":   b.handleStart,
		"help":    b.handleHelp,
		"status":  b.handleStatus,
		"verify":  b.handleVerify,
		"connect": b.handleConnect,
	}
}

// StartListening starts the bot's update loop
func (b *Bot) StartListening(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)
	b.updates = updates

	for {
		select {
		case update := <-updates:
			go b.handleUpdate(ctx, update)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handleUpdate processes a Telegram update
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	// Handle messages
	if update.Message != nil {
		b.handleMessage(ctx, update.Message)
		return
	}

	// Handle callback queries (button presses)
	if update.CallbackQuery != nil {
		b.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}
}

// handleMessage processes a Telegram message
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	// Log the message
	if message.From != nil {
		log.Printf("Received message from %s (%s): %s", message.From.UserName, message.From.FirstName, message.Text)
	}

	// Handle commands
	if message.IsCommand() {
		command := message.Command()
		args := message.CommandArguments()

		handler, exists := b.handlers[command]
		if exists {
			if err := handler(ctx, message, args); err != nil {
				log.Printf("Error handling command %s: %v", command, err)
				if sendErr := b.sendErrorMessage(message.Chat.ID, "An error occurred processing your command"); sendErr != nil {
					log.Printf("Failed to send error message: %v", sendErr)
				}
			}
			return
		}

		// Unknown command
		if err := b.sendMessage(message.Chat.ID, "Unknown command. Type /help for available commands."); err != nil {
			log.Printf("Failed to send unknown command message: %v", err)
		}
		return
	}

	// Update user activity for non-command messages from registered users
	if message.From != nil {
		user, err := b.repo.GetUserByTelegramID(ctx, strconv.FormatInt(message.From.ID, 10))
		if err == nil {
			// User exists, update their activity
			user.LastActivity = time.Now().UTC()
			if err := b.repo.UpdateUser(ctx, user); err != nil {
				log.Printf("Error updating user activity: %v", err)
			}

			// Mark any pending pings as responded
			latestPing, err := b.repo.GetLatestPingByUserID(ctx, user.ID)
			if err == nil && latestPing.Status == "sent" {
				now := time.Now().UTC()
				latestPing.Status = "responded"
				latestPing.RespondedAt = &now
				if err := b.repo.UpdatePingHistory(ctx, latestPing); err != nil {
					log.Printf("Error updating ping history: %v", err)
				}
			}
		}
	}

	// Default response for non-command messages
	if err := b.sendMessage(message.Chat.ID, "I only respond to commands. Type /help for available commands."); err != nil {
		log.Printf("Failed to send default response: %v", err)
	}
}

// handleCallbackQuery processes a callback query (button press)
func (b *Bot) handleCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	// Extract the data from the callback
	data := query.Data

	// Log the callback
	log.Printf("Received callback query from %s: %s", query.From.UserName, data)

	// Parse the callback data
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		log.Printf("Invalid callback data format: %s", data)
		return
	}

	action := parts[0]

	switch action {
	case "verify":
		if len(parts) < 3 {
			log.Printf("Invalid verify callback format: %s", data)
			return
		}
		userID := parts[1]
		pingID := parts[2]

		// Get the user
		user, err := b.repo.GetUserByID(ctx, userID)
		if err != nil {
			log.Printf("Error getting user %s: %v", userID, err)
			if answerErr := b.answerCallbackQuery(query.ID, "Error: User not found"); answerErr != nil {
				log.Printf("Failed to answer callback query: %v", answerErr)
			}
			return
		}

		// Get the ping
		var ping *models.PingHistory
		if pingID != "0" {
			ping, err = b.repo.GetLatestPingByUserID(ctx, userID)
			if err != nil {
				log.Printf("Error getting ping for user %s: %v", userID, err)
				// Continue anyway, we'll create a new ping response
			}
		}

		// Update user's last activity
		user.LastActivity = time.Now().UTC()
		if err := b.repo.UpdateUser(ctx, user); err != nil {
			log.Printf("Error updating user activity: %v", err)
		}

		// Update ping status if it exists
		if ping != nil && ping.Status == "sent" {
			now := time.Now().UTC()
			ping.Status = "responded"
			ping.RespondedAt = &now
			if err := b.repo.UpdatePingHistory(ctx, ping); err != nil {
				log.Printf("Error updating ping: %v", err)
			}
		}

		// Send confirmation message
		if err := b.editMessageText(query.Message.Chat.ID, query.Message.MessageID,
			"✅ Thank you for confirming your status. Your Dead Man's Switch has been reset."); err != nil {
			log.Printf("Failed to edit message: %v", err)
		}
		if err := b.answerCallbackQuery(query.ID, "Verification successful"); err != nil {
			log.Printf("Failed to answer callback query: %v", err)
		}

	default:
		log.Printf("Unknown callback action: %s", action)
		if err := b.answerCallbackQuery(query.ID, "Invalid action"); err != nil {
			log.Printf("Failed to answer callback query: %v", err)
		}
	}
}

// SendPingMessage sends a ping message to a user
func (b *Bot) SendPingMessage(ctx context.Context, user *models.User, pingID string) error {
	if user.TelegramID == "" {
		return fmt.Errorf("user has no associated Telegram ID")
	}

	chatID, err := strconv.ParseInt(user.TelegramID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid Telegram ID: %w", err)
	}

	// Create inline keyboard with verification button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("I'm OK - Confirm", fmt.Sprintf("verify:%s:%s", user.ID, pingID)),
		),
	)

	message := fmt.Sprintf(
		"🔔 *Dead Man's Switch Check-In*\n\n"+
			"Please confirm you're okay by pressing the button below.\n\n"+
			"If you don't respond within %d days, your pre-configured secrets will be sent to your designated recipients.",
		user.PingDeadline,
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err = b.bot.Send(msg)
	return err
}

// Command handlers

func (b *Bot) handleStart(ctx context.Context, message *tgbotapi.Message, args string) error {
	var response string

	// Check if user already exists
	tgID := strconv.FormatInt(message.From.ID, 10)
	_, err := b.repo.GetUserByTelegramID(ctx, tgID)

	if err == nil {
		// User exists
		response = fmt.Sprintf("Welcome back, %s! Your Dead Man's Switch is active. Type /status to see your current settings.", message.From.FirstName)
	} else if err == storage.ErrNotFound {
		// New user
		response = fmt.Sprintf(
			"Welcome to Dead Man's Switch, %s!\n\n"+
				"This bot helps ensure your sensitive information is only shared if you're unable to respond to regular check-ins.\n\n"+
				"To connect this bot to your account, please visit https://%s and use the code:\n\n"+
				"%s\n\n"+
				"Or use the /connect command with your email: /connect your@email.com",
			message.From.FirstName, b.config.BaseDomain, tgID,
		)
	} else {
		// Database error
		return fmt.Errorf("database error: %w", err)
	}

	return b.sendMessage(message.Chat.ID, response)
}

func (b *Bot) handleHelp(ctx context.Context, message *tgbotapi.Message, args string) error {
	helpText := `
*Dead Man's Switch Bot Commands*

/start - Start the bot and get initial instructions
/help - Show this help message
/status - Check your current settings and status
/verify - Manually verify that you're okay
/connect [email] - Connect your Telegram account to your web account

For more information, visit our web interface at https://` + b.config.BaseDomain

	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
	msg.ParseMode = "Markdown"
	_, err := b.bot.Send(msg)
	return err
}

func (b *Bot) handleStatus(ctx context.Context, message *tgbotapi.Message, args string) error {
	// Get user by Telegram ID
	tgID := strconv.FormatInt(message.From.ID, 10)
	user, err := b.repo.GetUserByTelegramID(ctx, tgID)

	if err == storage.ErrNotFound {
		return b.sendMessage(message.Chat.ID, "You're not registered yet. Please use /start to begin.")
	} else if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Get user's secrets count
	secrets, err := b.repo.ListSecretsByUserID(ctx, user.ID)
	secretCount := 0
	if err == nil {
		secretCount = len(secrets)
	}

	// Get user's recipients count
	recipients, err := b.repo.ListRecipientsByUserID(ctx, user.ID)
	recipientCount := 0
	if err == nil {
		recipientCount = len(recipients)
	}

	// Calculate next ping time
	nextPingTime := "Not scheduled"
	if !user.NextScheduledPing.IsZero() {
		nextPingTime = user.NextScheduledPing.Format("Jan 2, 2006 at 15:04 MST")
	}

	// Prepare status message
	statusText := fmt.Sprintf(
		"*Your Dead Man's Switch Status*\n\n"+
			"Email: %s\n"+
			"Ping Frequency: Every %d days\n"+
			"Response Deadline: %d days\n"+
			"Pinging Enabled: %v\n"+
			"Ping Method: %s\n\n"+
			"Secrets Stored: %d\n"+
			"Recipients Configured: %d\n\n"+
			"Last Activity: %s\n"+
			"Next Scheduled Ping: %s\n\n"+
			"To manage your secrets and recipients, please visit https://%s",
		user.Email,
		user.PingFrequency,
		user.PingDeadline,
		user.PingingEnabled,
		user.PingMethod,
		secretCount,
		recipientCount,
		user.LastActivity.Format("Jan 2, 2006 at 15:04 MST"),
		nextPingTime,
		b.config.BaseDomain,
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, statusText)
	msg.ParseMode = "Markdown"
	_, err = b.bot.Send(msg)
	return err
}

func (b *Bot) handleVerify(ctx context.Context, message *tgbotapi.Message, args string) error {
	// Get user by Telegram ID
	tgID := strconv.FormatInt(message.From.ID, 10)
	user, err := b.repo.GetUserByTelegramID(ctx, tgID)

	if err == storage.ErrNotFound {
		return b.sendMessage(message.Chat.ID, "You're not registered yet. Please use /start to begin.")
	} else if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Create inline keyboard with verification button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("I'm OK - Confirm", fmt.Sprintf("verify:%s:0", user.ID)),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Please confirm you're okay by pressing the button below:")
	msg.ReplyMarkup = keyboard

	_, err = b.bot.Send(msg)
	return err
}

func (b *Bot) handleConnect(ctx context.Context, message *tgbotapi.Message, args string) error {
	if args == "" {
		return b.sendMessage(message.Chat.ID, "Please provide your email address: /connect your@email.com")
	}

	// Validate email format
	email := strings.TrimSpace(args)
	if !strings.Contains(email, "@") {
		return b.sendMessage(message.Chat.ID, "Invalid email format. Please try again.")
	}

	// Check if user exists with this email
	user, err := b.repo.GetUserByEmail(ctx, email)
	if err == storage.ErrNotFound {
		return b.sendMessage(message.Chat.ID, fmt.Sprintf(
			"No account found with email %s. Please register at https://%s first.",
			email, b.config.BaseDomain,
		))
	} else if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Update user's Telegram ID
	tgID := strconv.FormatInt(message.From.ID, 10)
	user.TelegramID = tgID
	user.TelegramUsername = message.From.UserName
	user.LastActivity = time.Now().UTC()

	if err := b.repo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return b.sendMessage(message.Chat.ID, fmt.Sprintf(
		"✅ Success! Your Telegram account is now connected to %s.\n\nType /status to see your current settings.",
		email,
	))
}

// Helper methods

func (b *Bot) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.bot.Send(msg)
	return err
}

func (b *Bot) sendErrorMessage(chatID int64, text string) error {
	return b.sendMessage(chatID, "❌ "+text)
}

func (b *Bot) editMessageText(chatID int64, messageID int, text string) error {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := b.bot.Send(msg)
	return err
}

func (b *Bot) answerCallbackQuery(queryID string, text string) error {
	callback := tgbotapi.NewCallback(queryID, text)
	_, err := b.bot.Request(callback)
	return err
}
