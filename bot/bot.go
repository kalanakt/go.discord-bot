package bot

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kalanakt/go.discord-bot/config"
	"github.com/kalanakt/go.discord-bot/database"
	"github.com/sirupsen/logrus"
)

// Bot represents the Discord bot instance
type Bot struct {
	Config     *config.Config
	Session    *discordgo.Session
	Repository *database.Repository
	Commands   *CommandHandler
	StartTime  time.Time
	Guilds     map[string]*discordgo.Guild
	guildMutex sync.RWMutex
}

// New creates a new Discord bot instance
func New(cfg *config.Config, db *sql.DB) (*Bot, error) {
	// Create Discord session
	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	// Create bot instance
	bot := &Bot{
		Config:     cfg,
		Session:    session,
		Repository: database.NewRepository(db),
		Guilds:     make(map[string]*discordgo.Guild),
	}

	// Initialize command handler
	bot.Commands = NewCommandHandler(bot)

	// Register event handlers
	session.AddHandler(bot.onReady)
	session.AddHandler(bot.onGuildCreate)
	session.AddHandler(bot.onGuildDelete)
	session.AddHandler(bot.onMessageCreate)
	session.AddHandler(bot.onInteractionCreate)

	// Set intents
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsDirectMessages

	return bot, nil
}

// Start connects the bot to Discord
func (b *Bot) Start() error {
	// Connect to Discord
	if err := b.Session.Open(); err != nil {
		return fmt.Errorf("error opening connection to Discord: %w", err)
	}

	// Register slash commands
	if err := b.Commands.RegisterSlashCommands(); err != nil {
		return fmt.Errorf("error registering slash commands: %w", err)
	}

	// Set start time for uptime tracking
	b.StartTime = time.Now()

	// Start stats updater
	go b.statsUpdater()

	return nil
}

// Stop disconnects the bot from Discord
func (b *Bot) Stop() {
	// Unregister slash commands if in dev mode
	if b.Config.DevMode && b.Config.DevGuildID != "" {
		if err := b.Commands.UnregisterSlashCommands(); err != nil {
			logrus.Errorf("Error unregistering slash commands: %v", err)
		}
	}

	// Close Discord session
	if err := b.Session.Close(); err != nil {
		logrus.Errorf("Error closing Discord session: %v", err)
	}
}

// GetUptime returns the bot's uptime duration
func (b *Bot) GetUptime() time.Duration {
	return time.Since(b.StartTime)
}

// GetGuilds returns a copy of the guilds map
func (b *Bot) GetGuilds() map[string]*discordgo.Guild {
	b.guildMutex.RLock()
	defer b.guildMutex.RUnlock()

	// Create a copy to avoid concurrent map access
	guildsCopy := make(map[string]*discordgo.Guild, len(b.Guilds))
	for id, guild := range b.Guilds {
		guildsCopy[id] = guild
	}

	return guildsCopy
}

// statsUpdater periodically updates bot statistics in the database
func (b *Bot) statsUpdater() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		b.updateStats()
	}
}

// updateStats updates the bot statistics in the database
func (b *Bot) updateStats() {
	b.guildMutex.RLock()
	guildsCount := len(b.Guilds)
	b.guildMutex.RUnlock()

	// Count total users across all guilds
	usersCount := 0
	for _, guild := range b.GetGuilds() {
		usersCount += guild.MemberCount
	}

	// Calculate uptime in seconds
	uptimeSeconds := int(b.GetUptime().Seconds())

	// Update database
	if err := b.Repository.UpdateBotStats(guildsCount, usersCount, uptimeSeconds); err != nil {
		logrus.Errorf("Error updating bot stats: %v", err)
	}
}
