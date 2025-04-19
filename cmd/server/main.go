// Package main implements the entry point for the Dead Man's Switch service.
// It initializes configuration, sets up repositories, services, and starts the web server.
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
	// Create a context that's canceled when we receive a termination signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Configure logging
	log.Printf("Starting Dead Man's Switch server, version 1.0.0")
	log.Printf("Debug mode: %v", cfg.Debug)

	// Initialize database
	log.Printf("Initializing database at %s", cfg.DBPath)
	repo, err := storage.NewRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize email client if SMTP is configured
	var emailClient *email.Client
	if cfg.SMTPHost != "" {
		log.Printf("Initializing email client with SMTP server %s", cfg.SMTPHost)
		emailClient, err = email.NewClient(cfg)
		if err != nil {
			log.Printf("Warning: Failed to initialize email client: %v", err)
		}
	} else {
		log.Printf("Warning: SMTP not configured, email notifications will be disabled")
	}

	// Initialize Telegram bot
	log.Printf("Initializing Telegram bot")
	telegramBot, err := telegram.NewBot(cfg, repo)
	if err != nil {
		log.Printf("Warning: Failed to initialize Telegram bot: %v", err)
		log.Printf("Telegram notifications will be disabled")
	} else {
		// Start Telegram bot in a goroutine
		go func() {
			log.Printf("Starting Telegram bot")
			if err := telegramBot.StartListening(ctx); err != nil && err != context.Canceled {
				log.Printf("Telegram bot error: %v", err)
			}
		}()
	}

	// Initialize scheduler
	log.Printf("Initializing scheduler")
	sched := scheduler.NewScheduler(repo, emailClient, telegramBot, cfg)
	if err := sched.Start(ctx); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer sched.Stop()

	// Initialize and start web server
	log.Printf("Initializing web server on domain %s", cfg.BaseDomain)
	webServer := web.NewServer(cfg, repo, emailClient, telegramBot, sched)

	// Start web server in its own goroutine
	go func() {
		log.Printf("Starting web server")
		if err := webServer.Start(); err != nil {
			log.Printf("Web server error: %v", err)
			cancel() // Cancel the context to trigger shutdown
		}
	}()

	// Wait for termination signal or context cancellation
	<-ctx.Done()

	// Perform graceful shutdown
	log.Println("Shutting down...")

	// Give the web server a deadline to finish ongoing requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := webServer.Stop(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server stopped")
}
