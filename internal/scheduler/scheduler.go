package scheduler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/korjavin/deadmanswitch/internal/activity"
	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/email"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// Task represents a scheduled task
type Task struct {
	ID         string
	Name       string
	NextRun    time.Time
	Duration   time.Duration
	RunOnStart bool
	Handler    TaskHandler
}

// TaskHandler is a function that runs a task
type TaskHandler func(ctx context.Context) error

// EmailClient is an interface for email clients
type EmailClient interface {
	SendPingEmail(email, name, verificationCode string) error
	SendSecretDeliveryEmail(recipientEmail, recipientName, message, accessCode string) error
	SendEmail(options *email.MessageOptions) error
	SendEmailSimple(to []string, subject, body string, isHTML bool) error
}

// TelegramBot is an interface for telegram bots
type TelegramBot interface {
	SendPingMessage(ctx context.Context, user *models.User, pingID string) error
}

// Scheduler handles periodic tasks
type Scheduler struct {
	tasks            map[string]*Task
	repo             storage.Repository
	emailClient      EmailClient
	telegramBot      TelegramBot
	config           *config.Config
	activityRegistry *activity.Registry
	mu               sync.RWMutex
	stopChan         chan struct{}
	deliveryLock     sync.Mutex
}

// NewScheduler creates a new scheduler
func NewScheduler(
	repo storage.Repository,
	emailClient EmailClient,
	telegramBot TelegramBot,
	config *config.Config,
) *Scheduler {
	// Create activity registry and register providers
	activityRegistry := activity.NewRegistry()
	activityRegistry.Register(activity.NewGitHubProvider())

	return &Scheduler{
		tasks:            make(map[string]*Task),
		repo:             repo,
		emailClient:      emailClient,
		telegramBot:      telegramBot,
		config:           config,
		activityRegistry: activityRegistry,
		stopChan:         make(chan struct{}),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	// Register the periodic tasks
	if err := s.registerTasks(); err != nil {
		return err
	}

	// Run tasks immediately if specified
	for _, task := range s.tasks {
		if task.RunOnStart {
			go func(t *Task) {
				log.Printf("Running task on start: %s", t.Name)
				if err := t.Handler(ctx); err != nil {
					log.Printf("Error running task %s: %v", t.Name, err)
				}
			}(task)
		}
	}

	// Start the scheduler loop
	go s.startLoop(ctx)

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler...")
	close(s.stopChan)
}

// registerTasks adds all the tasks to the scheduler
func (s *Scheduler) registerTasks() error {
	// Task for checking and sending pings
	s.AddTask(&Task{
		ID:         uuid.New().String(),
		Name:       "PingTask",
		Duration:   5 * time.Minute, // Check for pending pings every 5 minutes
		RunOnStart: true,
		Handler:    s.pingTask,
	})

	// Task for sending escalating reminders as deadlines approach
	s.AddTask(&Task{
		ID:         uuid.New().String(),
		Name:       "ReminderTask",
		Duration:   30 * time.Minute, // Check for approaching deadlines every 30 minutes
		RunOnStart: true,
		Handler:    s.reminderTask,
	})

	// Task for checking expired pings
	s.AddTask(&Task{
		ID:         uuid.New().String(),
		Name:       "DeadSwitchTask",
		Duration:   15 * time.Minute, // Check for expired switches every 15 minutes
		RunOnStart: true,
		Handler:    s.deadSwitchTask,
	})

	// Task for checking external activity (GitHub, etc.)
	s.AddTask(&Task{
		ID:         uuid.New().String(),
		Name:       "ExternalActivityTask",
		Duration:   1 * time.Hour, // Check external activity hourly
		RunOnStart: true,
		Handler:    s.externalActivityTask,
	})

	// Task for re-encrypting secret questions
	s.AddTask(&Task{
		ID:         uuid.New().String(),
		Name:       "ReencryptQuestionsTask",
		Duration:   1 * time.Hour, // Re-encrypt questions hourly
		RunOnStart: true,
		Handler:    s.ReencryptQuestionsTask,
	})

	// Task for cleaning up expired access codes
	s.AddTask(&Task{
		ID:         uuid.New().String(),
		Name:       "CleanupAccessCodesTask",
		Duration:   24 * time.Hour, // Clean up expired codes daily
		RunOnStart: false,
		Handler:    s.cleanupAccessCodesTask,
	})

	// Task for cleaning expired sessions
	s.AddTask(&Task{
		ID:         uuid.New().String(),
		Name:       "CleanupTask",
		Duration:   24 * time.Hour, // Run daily
		RunOnStart: false,
		Handler:    s.cleanupTask,
	})

	return nil
}

// AddTask adds a task to the scheduler
func (s *Scheduler) AddTask(task *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.NextRun = time.Now().Add(task.Duration)
	s.tasks[task.ID] = task
	log.Printf("Added task: %s, next run at: %s", task.Name, task.NextRun.Format(time.RFC3339))
}

// startLoop runs the scheduler loop
func (s *Scheduler) startLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Scheduler started")

	for {
		select {
		case <-ticker.C:
			s.checkTasks(ctx)
		case <-s.stopChan:
			log.Println("Scheduler stopped")
			return
		case <-ctx.Done():
			log.Println("Scheduler context cancelled")
			return
		}
	}
}

// checkTasks checks all tasks to see if they need to run
func (s *Scheduler) checkTasks(ctx context.Context) {
	s.mu.RLock()
	now := time.Now()
	tasksToRun := make([]*Task, 0)

	for _, task := range s.tasks {
		if now.After(task.NextRun) {
			tasksToRun = append(tasksToRun, task)
		}
	}
	s.mu.RUnlock()

	// Run due tasks and update their next run time
	for _, task := range tasksToRun {
		go func(t *Task) {
			log.Printf("Running task: %s", t.Name)
			if err := t.Handler(ctx); err != nil {
				log.Printf("Error running task %s: %v", t.Name, err)
			}

			s.mu.Lock()
			t.NextRun = time.Now().Add(t.Duration)
			log.Printf("Task completed: %s, next run at: %s", t.Name, t.NextRun.Format(time.RFC3339))
			s.mu.Unlock()
		}(task)
	}
}

// pingTask sends pings to users who are due
func (s *Scheduler) pingTask(ctx context.Context) error {
	users, err := s.repo.GetUsersForPinging(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users for pinging: %w", err)
	}

	log.Printf("Found %d users that need to be pinged", len(users))

	for _, user := range users {
		// Skip users that don't have pinging enabled
		if !user.PingingEnabled {
			continue
		}

		// Create ping history record
		ping := &models.PingHistory{
			ID:     uuid.New().String(),
			UserID: user.ID,
			SentAt: time.Now().UTC(),
			Status: "sent",
		}

		// Determine which method to use
		switch user.PingMethod {
		case "telegram":
			if user.TelegramID != "" {
				ping.Method = "telegram"
				if err := s.repo.CreatePingHistory(ctx, ping); err != nil {
					log.Printf("Failed to create ping history for user %s: %v", user.ID, err)
					continue
				}
				if err := s.telegramBot.SendPingMessage(ctx, user, ping.ID); err != nil {
					log.Printf("Failed to send Telegram ping to user %s: %v", user.ID, err)
				}
			} else {
				log.Printf("User %s has telegram method but no telegram ID", user.ID)
			}

		case "email":
			ping.Method = "email"
			if err := s.sendEmailPing(ctx, user, ping); err != nil {
				log.Printf("Failed to send email ping to user %s: %v", user.ID, err)
			}

		case "both", "":
			// Try Telegram first
			if user.TelegramID != "" {
				telegramPing := &models.PingHistory{
					ID:     uuid.New().String(),
					UserID: user.ID,
					SentAt: time.Now().UTC(),
					Method: "telegram",
					Status: "sent",
				}
				if err := s.repo.CreatePingHistory(ctx, telegramPing); err != nil {
					log.Printf("Failed to create telegram ping history for user %s: %v", user.ID, err)
				} else {
					if err := s.telegramBot.SendPingMessage(ctx, user, telegramPing.ID); err != nil {
						log.Printf("Failed to send Telegram ping to user %s: %v", user.ID, err)
					}
				}
			}

			// Also send email
			emailPing := &models.PingHistory{
				ID:     uuid.New().String(),
				UserID: user.ID,
				SentAt: time.Now().UTC(),
				Method: "email",
				Status: "sent",
			}
			if err := s.sendEmailPing(ctx, user, emailPing); err != nil {
				log.Printf("Failed to send email ping to user %s: %v", user.ID, err)
			}

		default:
			log.Printf("Unknown ping method for user %s: %s", user.ID, user.PingMethod)
		}

		// Schedule next ping
		user.NextScheduledPing = time.Now().Add(time.Duration(user.PingFrequency) * 24 * time.Hour)
		if err := s.repo.UpdateUser(ctx, user); err != nil {
			log.Printf("Failed to update next ping time for user %s: %v", user.ID, err)
		}
	}

	return nil
}

// sendEmailPing sends an email ping to a user
func (s *Scheduler) sendEmailPing(ctx context.Context, user *models.User, ping *models.PingHistory) error {
	// Create verification code
	verification := &models.PingVerification{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Code:      generateVerificationCode(),
		ExpiresAt: time.Now().UTC().Add(time.Duration(user.PingDeadline) * 24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now().UTC(),
	}

	// Save verification code
	if err := s.repo.CreatePingVerification(ctx, verification); err != nil {
		return fmt.Errorf("failed to create ping verification: %w", err)
	}

	// Save ping history
	if err := s.repo.CreatePingHistory(ctx, ping); err != nil {
		return fmt.Errorf("failed to create ping history: %w", err)
	}

	// Send email
	return s.emailClient.SendPingEmail(user.Email, extractNameFromEmail(user.Email), verification.Code)
}

// deadSwitchTask checks for users who have expired deadlines and sends their secrets
func (s *Scheduler) deadSwitchTask(ctx context.Context) error {
	s.deliveryLock.Lock()
	defer s.deliveryLock.Unlock()

	// Get users who have exceeded their ping deadline
	users, err := s.repo.GetUsersWithExpiredPings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users with expired pings: %w", err)
	}

	log.Printf("Found %d users with expired pings", len(users))

	for _, user := range users {
		// Perform a final check of all activity sources before triggering the switch
		log.Printf("Performing final activity check for user %s before triggering switch", user.ID)

		// Check if the user has been active on any external platform
		active := false

		// Check configured activity providers
		configuredProviders := s.activityRegistry.GetConfiguredProviders(user)
		if len(configuredProviders) > 0 {
			activityDetected, err := s.activityRegistry.CheckAnyActivity(ctx, user, user.LastActivity)
			if err != nil {
				log.Printf("Error checking external activity for user %s: %v", user.ID, err)
			} else if activityDetected {
				active = true
				latestActivity := s.activityRegistry.GetLatestActivityTime(ctx, user)
				log.Printf("User %s has been active on external platform at %s, cancelling switch trigger",
					user.ID, latestActivity.Format(time.RFC3339))

				// Update the user's last activity time
				user.LastActivity = latestActivity

				// Reschedule the next ping based on the updated activity time
				user.NextScheduledPing = latestActivity.Add(time.Duration(user.PingFrequency) * 24 * time.Hour)

				if err := s.repo.UpdateUser(ctx, user); err != nil {
					log.Printf("Failed to update user after detecting external activity: %v", err)
				}

				// Create audit log entry
				auditLog := &models.AuditLog{
					ID:        uuid.New().String(),
					UserID:    user.ID,
					Action:    "switch_trigger_cancelled",
					Timestamp: time.Now().UTC(),
					Details: fmt.Sprintf("Switch trigger cancelled due to activity detected on external platform at %s",
						latestActivity.Format(time.RFC3339)),
				}

				if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
					log.Printf("Failed to create audit log for switch cancellation: %v", err)
				}
			}
		}

		// Check for recent ping responses
		latestPing, err := s.repo.GetLatestPingByUserID(ctx, user.ID)
		if err == nil && latestPing != nil && latestPing.Status == "responded" {
			// User has responded to a ping, update their last activity
			if latestPing.RespondedAt != nil && latestPing.RespondedAt.After(user.LastActivity) {
				active = true
				log.Printf("User %s has responded to a ping at %s, cancelling switch trigger",
					user.ID, latestPing.RespondedAt.Format(time.RFC3339))

				// Update the user's last activity time
				user.LastActivity = *latestPing.RespondedAt

				// Reschedule the next ping based on the updated activity time
				user.NextScheduledPing = user.LastActivity.Add(time.Duration(user.PingFrequency) * 24 * time.Hour)

				if err := s.repo.UpdateUser(ctx, user); err != nil {
					log.Printf("Failed to update user after detecting ping response: %v", err)
				}

				// Create audit log entry
				auditLog := &models.AuditLog{
					ID:        uuid.New().String(),
					UserID:    user.ID,
					Action:    "switch_trigger_cancelled",
					Timestamp: time.Now().UTC(),
					Details: fmt.Sprintf("Switch trigger cancelled due to ping response detected at %s",
						latestPing.RespondedAt.Format(time.RFC3339)),
				}

				if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
					log.Printf("Failed to create audit log for switch cancellation: %v", err)
				}
			}
		}

		// If no activity was detected, trigger the switch
		if !active {
			log.Printf("No activity detected for user %s, triggering switch", user.ID)

			// Create audit log entry for switch trigger
			auditLog := &models.AuditLog{
				ID:        uuid.New().String(),
				UserID:    user.ID,
				Action:    "switch_triggered",
				Timestamp: time.Now().UTC(),
				Details:   fmt.Sprintf("Dead man's switch triggered after no activity for %d days", user.PingDeadline),
			}

			if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
				log.Printf("Failed to create audit log for switch trigger: %v", err)
			}

			// Deliver secrets
			if err := s.deliverSecrets(ctx, user); err != nil {
				log.Printf("Failed to deliver secrets for user %s: %v", user.ID, err)
			}
		}
	}

	return nil
}

// deliverSecrets delivers a user's secrets to their recipients
func (s *Scheduler) deliverSecrets(ctx context.Context, user *models.User) error {
	// Get all recipients for this user
	recipients, err := s.repo.ListRecipientsByUserID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get recipients for user %s: %w", user.ID, err)
	}

	log.Printf("Delivering secrets for user %s to %d recipients", user.ID, len(recipients))

	for _, recipient := range recipients {
		// Get secret assignments for this recipient
		assignments, err := s.repo.ListSecretAssignmentsByRecipientID(ctx, recipient.ID)
		if err != nil {
			log.Printf("Failed to get secret assignments for recipient %s: %v", recipient.ID, err)
			continue
		}

		if len(assignments) == 0 {
			log.Printf("No secrets assigned to recipient %s", recipient.ID)
			continue
		}

		// Generate access code (plain text - will be sent via email)
		accessCode := generateAccessCode()

		// Hash access code for storage (SHA256 for deterministic lookup)
		hash := sha256.Sum256([]byte(accessCode))
		hashedCodeStr := hex.EncodeToString(hash[:])

		// Use transaction for atomicity
		tx, err := s.repo.BeginTx(ctx)
		if err != nil {
			log.Printf("Failed to begin transaction: %v", err)
			continue
		}

		// Create delivery event
		deliveryEvent := &models.DeliveryEvent{
			ID:          uuid.New().String(),
			UserID:      user.ID,
			RecipientID: recipient.ID,
			SentAt:      time.Now().UTC(),
			Status:      "pending",
		}
		if err := tx.CreateDeliveryEvent(ctx, deliveryEvent); err != nil {
			log.Printf("Failed to create delivery event: %v", err)
			tx.Rollback()
			continue
		}

		// Store access code securely with TTL
		accessCodeModel := &models.AccessCode{
			ID:              uuid.New().String(),
			Code:            hashedCodeStr,
			RecipientID:     recipient.ID,
			UserID:          user.ID,
			DeliveryEventID: deliveryEvent.ID,
			CreatedAt:       time.Now().UTC(),
			ExpiresAt:       time.Now().UTC().Add(time.Duration(s.config.AccessCodeExpirationDays) * 24 * time.Hour),
			MaxAttempts:     s.config.AccessCodeMaxAttempts,
		}

		if err := tx.CreateAccessCode(ctx, accessCodeModel); err != nil {
			log.Printf("Failed to store access code for recipient %s: %v", recipient.ID, err)
			tx.Rollback()
			continue
		}

		// Commit transaction before sending email
		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit transaction: %v", err)
			continue
		}

		// Send delivery email (outside transaction - email is best-effort)
		if err := s.emailClient.SendSecretDeliveryEmail(
			recipient.Email,
			recipient.Name,
			recipient.Message,
			accessCode,
		); err != nil {
			log.Printf("Failed to send delivery email to %s: %v", recipient.Email, err)

			// Update delivery event to failed (email send failed)
			deliveryEvent.Status = "failed"
			deliveryEvent.ErrorMessage = err.Error()
			if updateErr := s.repo.UpdateDeliveryEvent(ctx, deliveryEvent); updateErr != nil {
				log.Printf("Failed to update delivery event: %v", updateErr)
			}

			continue
		}

		// Update delivery event to sent
		deliveryEvent.Status = "sent"
		deliveryEvent.ErrorMessage = ""
		if err := s.repo.UpdateDeliveryEvent(ctx, deliveryEvent); err != nil {
			log.Printf("Failed to update delivery event: %v", err)
		}
	}

	// Disable pinging for this user now that secrets have been delivered
	user.PingingEnabled = false
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		log.Printf("Failed to update user after secret delivery: %v", err)
	}

	// Log the delivery
	log.Printf("Delivered all secrets for user %s", user.ID)

	return nil
}

// externalActivityTask checks for user activity on external platforms
func (s *Scheduler) externalActivityTask(ctx context.Context) error {
	log.Println("Running externalActivityTask")

	// Get all users with pinging enabled
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	log.Printf("Checking external activity for %d users", len(users))

	for _, user := range users {
		// Skip users that don't have pinging enabled
		if !user.PingingEnabled {
			continue
		}

		// Get configured activity providers for this user
		configuredProviders := s.activityRegistry.GetConfiguredProviders(user)
		if len(configuredProviders) == 0 {
			// No external activity providers configured for this user
			continue
		}

		// Check if the user has been active on any platform since their last activity
		active, err := s.activityRegistry.CheckAnyActivity(ctx, user, user.LastActivity)
		if err != nil {
			log.Printf("Error checking activity for user %s: %v", user.ID, err)
			continue
		}

		if active {
			// User has been active on an external platform, update their last activity time
			latestActivity := s.activityRegistry.GetLatestActivityTime(ctx, user)
			log.Printf("User %s has been active on external platform at %s", user.ID, latestActivity.Format(time.RFC3339))

			// Get the provider names that detected activity
			activeProviderNames := make([]string, 0)
			for _, provider := range configuredProviders {
				isActive, _ := provider.CheckActivity(ctx, user, user.LastActivity)
				if isActive {
					activeProviderNames = append(activeProviderNames, provider.Name())
				}
			}

			// Update the user's last activity time
			user.LastActivity = latestActivity

			// Reschedule the next ping based on the updated activity time
			user.NextScheduledPing = latestActivity.Add(time.Duration(user.PingFrequency) * 24 * time.Hour)

			if err := s.repo.UpdateUser(ctx, user); err != nil {
				log.Printf("Failed to update user after external activity: %v", err)
				continue
			}

			// Create detailed audit log entries for each active provider
			for _, providerName := range activeProviderNames {
				auditLog := &models.AuditLog{
					ID:        uuid.New().String(),
					UserID:    user.ID,
					Action:    fmt.Sprintf("%s_activity_detected", strings.ToLower(providerName)),
					Timestamp: time.Now().UTC(),
					Details: fmt.Sprintf("Activity detected on %s at %s, next check-in rescheduled to %s",
						providerName,
						latestActivity.Format(time.RFC3339),
						user.NextScheduledPing.Format(time.RFC3339)),
				}

				if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
					log.Printf("Failed to create audit log for external activity: %v", err)
				}
			}
		}
	}

	return nil
}

// reminderTask sends escalating reminders as deadlines approach
func (s *Scheduler) reminderTask(ctx context.Context) error {
	log.Println("Running reminderTask")

	// Get all users with pinging enabled
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	now := time.Now().UTC()
	for _, user := range users {
		// Skip users that don't have pinging enabled
		if !user.PingingEnabled {
			continue
		}

		// Calculate deadline
		deadline := user.LastActivity.Add(time.Duration(user.PingDeadline) * 24 * time.Hour)
		timeUntilDeadline := deadline.Sub(now)

		// Skip if deadline is not approaching
		if timeUntilDeadline > 48*time.Hour {
			continue
		}

		// Get the latest ping for this user
		latestPing, err := s.repo.GetLatestPingByUserID(ctx, user.ID)
		if err != nil {
			// No pings yet, or error fetching
			log.Printf("Error getting latest ping for user %s: %v", user.ID, err)
			continue
		}

		// Skip if the latest ping was sent recently (within last 12 hours) or was already responded to
		if latestPing != nil && (latestPing.Status == "responded" ||
			now.Sub(latestPing.SentAt) < 12*time.Hour) {
			continue
		}

		// Determine the urgency level based on time until deadline
		urgencyLevel := ""
		if timeUntilDeadline <= 12*time.Hour {
			urgencyLevel = "FINAL WARNING"
		} else if timeUntilDeadline <= 24*time.Hour {
			urgencyLevel = "URGENT"
		} else {
			urgencyLevel = "REMINDER"
		}

		// Send appropriate reminders based on user's configured methods
		if user.TelegramID != "" && (user.PingMethod == "telegram" || user.PingMethod == "both" || user.PingMethod == "") {
			// Create a telegram ping with urgency level
			telegramPing := &models.PingHistory{
				ID:     uuid.New().String(),
				UserID: user.ID,
				SentAt: now,
				Method: "telegram",
				Status: "sent",
			}

			if err := s.repo.CreatePingHistory(ctx, telegramPing); err != nil {
				log.Printf("Failed to create telegram reminder history for user %s: %v", user.ID, err)
			} else {
				// TODO: Enhance the SendPingMessage interface to include urgency level
				if err := s.telegramBot.SendPingMessage(ctx, user, telegramPing.ID); err != nil {
					log.Printf("Failed to send Telegram reminder to user %s: %v", user.ID, err)
				}
			}
		}

		// Always send email reminder as a backup
		emailPing := &models.PingHistory{
			ID:     uuid.New().String(),
			UserID: user.ID,
			SentAt: now,
			Method: "email",
			Status: "sent",
		}

		// Create verification code
		verification := &models.PingVerification{
			ID:        uuid.New().String(),
			UserID:    user.ID,
			Code:      generateVerificationCode(),
			ExpiresAt: deadline,
			Used:      false,
			CreatedAt: now,
		}

		// Save verification code
		if err := s.repo.CreatePingVerification(ctx, verification); err != nil {
			log.Printf("Failed to create ping verification for reminder: %v", err)
			continue
		}

		// Save ping history
		if err := s.repo.CreatePingHistory(ctx, emailPing); err != nil {
			log.Printf("Failed to create email reminder history for user %s: %v", user.ID, err)
			continue
		}

		// Send email with urgency level
		// TODO: Enhance the SendPingEmail interface to include urgency level
		if err := s.emailClient.SendPingEmail(user.Email, extractNameFromEmail(user.Email), verification.Code); err != nil {
			log.Printf("Failed to send email reminder to user %s: %v", user.ID, err)
		}

		// Create audit log entry for the reminder
		auditLog := &models.AuditLog{
			ID:        uuid.New().String(),
			UserID:    user.ID,
			Action:    fmt.Sprintf("%s_reminder_sent", strings.ToLower(urgencyLevel)),
			Timestamp: now,
			Details:   fmt.Sprintf("%s reminder sent. Deadline in %s", urgencyLevel, formatDuration(timeUntilDeadline)),
		}

		if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
			log.Printf("Failed to create audit log for reminder: %v", err)
		}
	}

	return nil
}

// cleanupTask handles cleanup operations
func (s *Scheduler) cleanupTask(ctx context.Context) error {
	log.Println("Running cleanupTask")
	// Delete expired sessions
	if err := s.repo.DeleteExpiredSessions(ctx); err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

// cleanupAccessCodesTask handles cleanup of expired access codes
func (s *Scheduler) cleanupAccessCodesTask(ctx context.Context) error {
	log.Println("Running cleanupAccessCodesTask")

	// Delete expired access codes
	if err := s.repo.DeleteExpiredAccessCodes(ctx); err != nil {
		return fmt.Errorf("failed to delete expired access codes: %w", err)
	}

	log.Println("Expired access codes cleaned up successfully")
	return nil
}

// Helper functions

// generateVerificationCode creates a unique code for email ping verification
func generateVerificationCode() string {
	// Use a UUID but make it shorter and more user-friendly
	raw := uuid.New().String()
	return raw[:8] + raw[9:13] + raw[14:18]
}

// generateAccessCode creates a unique access code for secret delivery
func generateAccessCode() string {
	return uuid.New().String()
}

// extractNameFromEmail tries to extract a name from an email address
func extractNameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) < 1 {
		return "User"
	}

	name := parts[0]
	// Replace dots and underscores with spaces
	name = strings.Replace(name, ".", " ", -1)
	name = strings.Replace(name, "_", " ", -1)

	// Capitalize the first letter of each word
	words := strings.Split(name, " ")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
