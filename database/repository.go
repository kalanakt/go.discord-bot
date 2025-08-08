package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

// Repository handles database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new database repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CommandLog represents a command usage log entry
type CommandLog struct {
	ID          int64
	GuildID     string
	ChannelID   string
	UserID      string
	CommandName string
	CommandType string
	Arguments   map[string]interface{}
	CreatedAt   time.Time
}

// InteractionEvent represents an interaction event log entry
type InteractionEvent struct {
	ID             int64
	GuildID        string
	ChannelID      string
	UserID         string
	InteractionType string
	ComponentID    string
	Data           map[string]interface{}
	CreatedAt      time.Time
}

// BotStats represents bot statistics
type BotStats struct {
	ID             int64
	GuildsCount    int
	UsersCount     int
	CommandsCount  int
	UptimeSeconds  int
	UpdatedAt      time.Time
}

// LogCommand records a command usage
func (r *Repository) LogCommand(guildID, channelID, userID, commandName, commandType string, arguments map[string]interface{}) error {
	argumentsJSON, err := json.Marshal(arguments)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		"INSERT INTO command_logs (guild_id, channel_id, user_id, command_name, command_type, arguments) VALUES ($1, $2, $3, $4, $5, $6)",
		guildID, channelID, userID, commandName, commandType, argumentsJSON,
	)
	if err != nil {
		logrus.Errorf("Failed to log command: %v", err)
		return err
	}

	// Update command count in bot_stats
	_, err = r.db.Exec("UPDATE bot_stats SET commands_count = commands_count + 1, updated_at = NOW() WHERE id = 1")
	if err != nil {
		logrus.Warnf("Failed to update bot stats: %v", err)
	}

	return nil
}

// LogInteraction records an interaction event
func (r *Repository) LogInteraction(guildID, channelID, userID, interactionType, componentID string, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(
		"INSERT INTO interaction_events (guild_id, channel_id, user_id, interaction_type, component_id, data) VALUES ($1, $2, $3, $4, $5, $6)",
		guildID, channelID, userID, interactionType, componentID, dataJSON,
	)
	if err != nil {
		logrus.Errorf("Failed to log interaction: %v", err)
		return err
	}

	return nil
}

// UpdateBotStats updates the bot statistics
func (r *Repository) UpdateBotStats(guildsCount, usersCount int, uptimeSeconds int) error {
	// Check if stats record exists
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM bot_stats").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Create initial record
		_, err = r.db.Exec(
			"INSERT INTO bot_stats (guilds_count, users_count, commands_count, uptime_seconds) VALUES ($1, $2, 0, $3)",
			guildsCount, usersCount, uptimeSeconds,
		)
	} else {
		// Update existing record
		_, err = r.db.Exec(
			"UPDATE bot_stats SET guilds_count = $1, users_count = $2, uptime_seconds = $3, updated_at = NOW() WHERE id = 1",
			guildsCount, usersCount, uptimeSeconds,
		)
	}

	if err != nil {
		logrus.Errorf("Failed to update bot stats: %v", err)
		return err
	}

	return nil
}

// GetBotStats retrieves the current bot statistics
func (r *Repository) GetBotStats() (*BotStats, error) {
	var stats BotStats
	err := r.db.QueryRow(
		"SELECT id, guilds_count, users_count, commands_count, uptime_seconds, updated_at FROM bot_stats ORDER BY id LIMIT 1",
	).Scan(&stats.ID, &stats.GuildsCount, &stats.UsersCount, &stats.CommandsCount, &stats.UptimeSeconds, &stats.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty stats if no record exists
			return &BotStats{}, nil
		}
		return nil, err
	}

	return &stats, nil
}

// GetRecentCommands retrieves recent command logs
func (r *Repository) GetRecentCommands(limit int) ([]CommandLog, error) {
	rows, err := r.db.Query(
		"SELECT id, guild_id, channel_id, user_id, command_name, command_type, arguments, created_at FROM command_logs ORDER BY created_at DESC LIMIT $1",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []CommandLog
	for rows.Next() {
		var log CommandLog
		var argumentsJSON []byte
		err := rows.Scan(&log.ID, &log.GuildID, &log.ChannelID, &log.UserID, &log.CommandName, &log.CommandType, &argumentsJSON, &log.CreatedAt)
		if err != nil {
			return nil, err
		}

		if len(argumentsJSON) > 0 {
			if err := json.Unmarshal(argumentsJSON, &log.Arguments); err != nil {
				logrus.Warnf("Failed to unmarshal arguments JSON: %v", err)
				log.Arguments = make(map[string]interface{})
			}
		} else {
			log.Arguments = make(map[string]interface{})
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// GetRecentInteractions retrieves recent interaction events
func (r *Repository) GetRecentInteractions(limit int) ([]InteractionEvent, error) {
	rows, err := r.db.Query(
		"SELECT id, guild_id, channel_id, user_id, interaction_type, component_id, data, created_at FROM interaction_events ORDER BY created_at DESC LIMIT $1",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []InteractionEvent
	for rows.Next() {
		var event InteractionEvent
		var dataJSON []byte
		err := rows.Scan(&event.ID, &event.GuildID, &event.ChannelID, &event.UserID, &event.InteractionType, &event.ComponentID, &dataJSON, &event.CreatedAt)
		if err != nil {
			return nil, err
		}

		if len(dataJSON) > 0 {
			if err := json.Unmarshal(dataJSON, &event.Data); err != nil {
				logrus.Warnf("Failed to unmarshal data JSON: %v", err)
				event.Data = make(map[string]interface{})
			}
		} else {
			event.Data = make(map[string]interface{})
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}