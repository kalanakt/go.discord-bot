package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kalanakt/go.discord-bot/bot"
	"github.com/kalanakt/go.discord-bot/config"
	"github.com/kalanakt/go.discord-bot/database"
	"github.com/sirupsen/logrus"
)

func init() {
	// Set up logging
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}
}

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up context with cancellation
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(cfg.DatabaseURL); err != nil {
		logrus.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize bot
	discordBot, err := bot.New(cfg, db)
	if err != nil {
		logrus.Fatalf("Failed to create Discord bot: %v", err)
	}

	// Start the bot
	if err := discordBot.Start(); err != nil {
		logrus.Fatalf("Failed to start Discord bot: %v", err)
	}

	// Wait for termination signal
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Graceful shutdown
	logrus.Info("Shutting down...")
	discordBot.Stop()
	logrus.Info("Shutdown complete")
}
