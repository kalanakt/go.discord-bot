package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// CommandHandler manages bot commands
type CommandHandler struct {
	Bot            *Bot
	PrefixCommands map[string]PrefixCommand
	SlashCommands  map[string]SlashCommand
}

// PrefixCommand represents a text-based command
type PrefixCommand struct {
	Name        string
	Description string
	Usage       string
	Handler     func(s *discordgo.Session, m *discordgo.MessageCreate, args []string)
}

// SlashCommand represents a slash command
type SlashCommand struct {
	Command     *discordgo.ApplicationCommand
	Handler     func(s *discordgo.Session, i *discordgo.InteractionCreate)
	Permissions int64
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(bot *Bot) *CommandHandler {
	handler := &CommandHandler{
		Bot:            bot,
		PrefixCommands: make(map[string]PrefixCommand),
		SlashCommands:  make(map[string]SlashCommand),
	}

	// Register commands
	handler.registerCommands()

	return handler
}

// registerCommands registers all bot commands
func (h *CommandHandler) registerCommands() {
	// Register prefix commands
	h.registerPrefixCommands()

	// Register slash commands
	h.registerSlashCommands()
}

// registerPrefixCommands registers all prefix commands
func (h *CommandHandler) registerPrefixCommands() {
	// Help command
	h.PrefixCommands["help"] = PrefixCommand{
		Name:        "help",
		Description: "Shows the help message",
		Usage:       "help [command]",
		Handler:     h.helpCommand,
	}

	// Ping command
	h.PrefixCommands["ping"] = PrefixCommand{
		Name:        "ping",
		Description: "Checks if the bot is online",
		Usage:       "ping",
		Handler:     h.pingCommand,
	}

	// Info command
	h.PrefixCommands["info"] = PrefixCommand{
		Name:        "info",
		Description: "Shows information about the bot",
		Usage:       "info",
		Handler:     h.infoCommand,
	}

	// Play command (example for voice)
	h.PrefixCommands["play"] = PrefixCommand{
		Name:        "play",
		Description: "Plays audio in a voice channel",
		Usage:       "play [URL or search term]",
		Handler:     h.playCommand,
	}
}

// registerSlashCommands defines all slash commands
func (h *CommandHandler) registerSlashCommands() {
	// Help command
	h.SlashCommands["help"] = SlashCommand{
		Command: &discordgo.ApplicationCommand{
			Name:        "help",
			Description: "Shows the help message",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "The command to get help for",
					Required:    false,
				},
			},
		},
		Handler:     h.helpSlashCommand,
		Permissions: 0, // No special permissions required
	}

	// Ping command
	h.SlashCommands["ping"] = SlashCommand{
		Command: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Checks if the bot is online",
		},
		Handler:     h.pingSlashCommand,
		Permissions: 0, // No special permissions required
	}

	// Info command
	h.SlashCommands["info"] = SlashCommand{
		Command: &discordgo.ApplicationCommand{
			Name:        "info",
			Description: "Shows information about the bot",
		},
		Handler:     h.infoSlashCommand,
		Permissions: 0, // No special permissions required
	}

	// Example button command
	h.SlashCommands["button"] = SlashCommand{
		Command: &discordgo.ApplicationCommand{
			Name:        "button",
			Description: "Shows an example button",
		},
		Handler:     h.buttonSlashCommand,
		Permissions: 0, // No special permissions required
	}

	// Example select menu command
	h.SlashCommands["select"] = SlashCommand{
		Command: &discordgo.ApplicationCommand{
			Name:        "select",
			Description: "Shows an example select menu",
		},
		Handler:     h.selectSlashCommand,
		Permissions: 0, // No special permissions required
	}

	// Example role management command
	h.SlashCommands["role"] = SlashCommand{
		Command: &discordgo.ApplicationCommand{
			Name:        "role",
			Description: "Manages roles for a user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "Adds a role to a user",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to add the role to",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionRole,
							Name:        "role",
							Description: "The role to add",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "Removes a role from a user",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to remove the role from",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionRole,
							Name:        "role",
							Description: "The role to remove",
							Required:    true,
						},
					},
				},
			},
		},
		Handler:     h.roleSlashCommand,
		Permissions: discordgo.PermissionManageRoles, // Requires manage roles permission
	}
}

// RegisterSlashCommands registers slash commands with Discord
func (h *CommandHandler) RegisterSlashCommands() error {
	for _, cmd := range h.SlashCommands {
		var err error

		// Register command globally or to a specific guild based on dev mode
		if h.Bot.Config.DevMode && h.Bot.Config.DevGuildID != "" {
			// Register to specific guild in dev mode
			_, err = h.Bot.Session.ApplicationCommandCreate(
				h.Bot.Session.State.User.ID,
				h.Bot.Config.DevGuildID,
				cmd.Command,
			)
		} else {
			// Register globally in production
			_, err = h.Bot.Session.ApplicationCommandCreate(
				h.Bot.Session.State.User.ID,
				"", // Empty string for global commands
				cmd.Command,
			)
		}

		if err != nil {
			return fmt.Errorf("error creating slash command '%s': %w", cmd.Command.Name, err)
		}

		logrus.Infof("Registered slash command: %s", cmd.Command.Name)
	}

	return nil
}

// UnregisterSlashCommands removes slash commands from Discord
func (h *CommandHandler) UnregisterSlashCommands() error {
	// Only unregister guild commands in dev mode
	if !h.Bot.Config.DevMode || h.Bot.Config.DevGuildID == "" {
		return nil
	}

	// Get all commands for the guild
	commands, err := h.Bot.Session.ApplicationCommands(
		h.Bot.Session.State.User.ID,
		h.Bot.Config.DevGuildID,
	)
	if err != nil {
		return fmt.Errorf("error getting slash commands: %w", err)
	}

	// Delete each command
	for _, cmd := range commands {
		err := h.Bot.Session.ApplicationCommandDelete(
			h.Bot.Session.State.User.ID,
			h.Bot.Config.DevGuildID,
			cmd.ID,
		)
		if err != nil {
			logrus.Errorf("Error deleting slash command '%s': %v", cmd.Name, err)
		} else {
			logrus.Infof("Unregistered slash command: %s", cmd.Name)
		}
	}

	return nil
}

// HandlePrefixCommand handles a prefix command
func (h *CommandHandler) HandlePrefixCommand(s *discordgo.Session, m *discordgo.MessageCreate, cmdName string, args []string) {
	// Check if command exists
	cmd, exists := h.PrefixCommands[strings.ToLower(cmdName)]
	if !exists {
		return
	}

	// Log command usage
	guildID := ""
	if m.GuildID != "" {
		guildID = m.GuildID
	}

	argumentsMap := make(map[string]interface{})
	for i, arg := range args {
		argumentsMap[fmt.Sprintf("arg%d", i+1)] = arg
	}

	err := h.Bot.Repository.LogCommand(guildID, m.ChannelID, m.Author.ID, cmdName, "prefix", argumentsMap)
	if err != nil {
		logrus.Errorf("Error logging command: %v", err)
	}

	// Execute command handler
	cmd.Handler(s, m, args)
}

// HandleSlashCommand handles a slash command
func (h *CommandHandler) HandleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get command name
	cmdName := i.ApplicationCommandData().Name

	// Check if command exists
	cmd, exists := h.SlashCommands[cmdName]
	if !exists {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Check permissions if in a guild
	if i.GuildID != "" && cmd.Permissions != 0 {
		// Get member permissions
		perms, err := s.State.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
		if err != nil {
			logrus.Errorf("Error checking permissions: %v", err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "An error occurred while checking permissions.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Check if user has required permissions
		if perms&cmd.Permissions == 0 {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You don't have permission to use this command.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
	}

	// Log command usage
	guildID := ""
	if i.GuildID != "" {
		guildID = i.GuildID
	}

	// Extract command options as arguments
	data := i.ApplicationCommandData()
	argumentsMap := make(map[string]interface{})

	// Handle options based on command structure
	if len(data.Options) > 0 {
		// Check if this is a subcommand
		if data.Options[0].Type == discordgo.ApplicationCommandOptionSubCommand {
			// Add subcommand name
			argumentsMap["subcommand"] = data.Options[0].Name

			// Add subcommand options
			for _, opt := range data.Options[0].Options {
				argumentsMap[opt.Name] = opt.Value
			}
		} else {
			// Add regular command options
			for _, opt := range data.Options {
				argumentsMap[opt.Name] = opt.Value
			}
		}
	}

	err := h.Bot.Repository.LogCommand(guildID, i.ChannelID, i.Member.User.ID, cmdName, "slash", argumentsMap)
	if err != nil {
		logrus.Errorf("Error logging command: %v", err)
	}

	// Execute command handler
	cmd.Handler(s, i)
}