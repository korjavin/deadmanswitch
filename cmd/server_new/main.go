package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/email"
	"github.com/korjavin/deadmanswitch/internal/scheduler"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/telegram"
	"github.com/korjavin/deadmanswitch/internal/web"
)

func main() {
	// Load configuration
	cfg := &config.Config{
		BaseDomain:         "localhost",
		Debug:              true,
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramBotUsername: os.Getenv("TELEGRAM_BOT_USERNAME"),
	}

	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./deadmanswitch.db"
	}
	repo, err := storage.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := repo.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize email client
	emailClient := email.NewClient(
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("EMAIL_FROM"),
	)

	// Initialize Telegram bot
	var telegramBot *telegram.Bot
	if cfg.TelegramBotToken != "" {
		telegramBot, err = telegram.NewBot(cfg.TelegramBotToken, repo)
		if err != nil {
			log.Printf("Warning: Failed to initialize Telegram bot: %v", err)
		}
	}

	// Initialize scheduler
	schedulerInstance := scheduler.NewScheduler(repo, emailClient, telegramBot)

	// Start the scheduler in a goroutine
	go func() {
		if err := schedulerInstance.Start(); err != nil {
			log.Printf("Scheduler error: %v", err)
		}
	}()

	// Initialize and start the web server with the new router
	server := web.NewServerWithRouter(cfg, repo, emailClient, telegramBot, schedulerInstance)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop the scheduler
	schedulerInstance.Stop()

	// Stop the server
	if err := server.Stop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
