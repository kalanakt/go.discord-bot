package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// helpCommand handles the help prefix command
func (h *CommandHandler) helpCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	var response string

	if len(args) > 0 {
		// Help for specific command
		cmdName := strings.ToLower(args[0])
		cmd, exists := h.PrefixCommands[cmdName]
		if exists {
			response = fmt.Sprintf("**%s**\nDescription: %s\nUsage: `%s%s`", 
				cmd.Name, cmd.Description, h.Bot.Config.CommandPrefix, cmd.Usage)
		} else {
			response = fmt.Sprintf("Command `%s` not found. Use `%shelp` to see all commands.", 
				cmdName, h.Bot.Config.CommandPrefix)
		}
	} else {
		// General help
		response = "**Available Commands:**\n"
		for _, cmd := range h.PrefixCommands {
			response += fmt.Sprintf("`%s%s` - %s\n", 
				h.Bot.Config.CommandPrefix, cmd.Name, cmd.Description)
		}
		response += "\nUse `/help [command]` for more information about a specific command."
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: response,
		Color:       0x00AAFF,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Discord Bot Template",
		},
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		logrus.Errorf("Error sending help message: %v", err)
	}
}

// pingCommand handles the ping prefix command
func (h *CommandHandler) pingCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Calculate latency
	start := time.Now()
	msg, err := s.ChannelMessageSend(m.ChannelID, "Pinging...")
	if err != nil {
		logrus.Errorf("Error sending ping message: %v", err)
		return
	}

	latency := time.Since(start)

	// Edit message with latency
	_, err = s.ChannelMessageEdit(m.ChannelID, msg.ID, fmt.Sprintf("Pong! Latency: %s", latency.Round(time.Millisecond)))
	if err != nil {
		logrus.Errorf("Error editing ping message: %v", err)
	}
}

// infoCommand handles the info prefix command
func (h *CommandHandler) infoCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Get bot stats
	guildsCount := len(h.Bot.GetGuilds())
	uptime := h.Bot.GetUptime().Round(time.Second)

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title: "Bot Information",
		Color: 0x00AAFF,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Servers",
				Value:  fmt.Sprintf("%d", guildsCount),
				Inline: true,
			},
			{
				Name:   "Uptime",
				Value:  uptime.String(),
				Inline: true,
			},
			{
				Name:   "Library",
				Value:  "discordgo",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Discord Bot Template",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embed)
	if err != nil {
		logrus.Errorf("Error sending info message: %v", err)
	}
}

// playCommand handles the play prefix command (example for voice)
func (h *CommandHandler) playCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	// Check if a search term or URL was provided
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Please provide a URL or search term.")
		return
	}

	// Find the user's voice channel
	var voiceChannelID string
	if m.GuildID != "" {
		// Get guild
		guild, err := s.State.Guild(m.GuildID)
		if err != nil {
			logrus.Errorf("Error getting guild: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Error finding your voice channel.")
			return
		}

		// Find the user in a voice channel
		for _, vs := range guild.VoiceStates {
			if vs.UserID == m.Author.ID {
				voiceChannelID = vs.ChannelID
				break
			}
		}

		if voiceChannelID == "" {
			s.ChannelMessageSend(m.ChannelID, "You must be in a voice channel to use this command.")
			return
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "This command can only be used in a server.")
		return
	}

	// In a real implementation, you would:
	// 1. Connect to the voice channel
	// 2. Download/stream the audio
	// 3. Play the audio
	// For this template, we'll just send a message
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Would play audio in voice channel <#%s>: %s", 
		voiceChannelID, strings.Join(args, " ")))
}

// helpSlashCommand handles the help slash command
func (h *CommandHandler) helpSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	var response string

	if len(options) > 0 && options[0].Name == "command" {
		// Help for specific command
		cmdName := strings.ToLower(options[0].StringValue())
		
		// Check slash commands
		slashCmd, slashExists := h.SlashCommands[cmdName]
		
		// Check prefix commands
		prefixCmd, prefixExists := h.PrefixCommands[cmdName]
		
		if slashExists {
			response = fmt.Sprintf("**/%s**\nDescription: %s\n", 
				slashCmd.Command.Name, slashCmd.Command.Description)
			
			// Add options if any
			if len(slashCmd.Command.Options) > 0 {
				response += "\n**Options:**\n"
				for _, opt := range slashCmd.Command.Options {
					requiredText := ""
					if opt.Required {
						requiredText = " (required)"
					}
					response += fmt.Sprintf("`%s`%s - %s\n", opt.Name, requiredText, opt.Description)
				}
			}
		} else if prefixExists {
			response = fmt.Sprintf("**%s%s**\nDescription: %s\nUsage: `%s%s`", 
				h.Bot.Config.CommandPrefix, prefixCmd.Name, prefixCmd.Description, h.Bot.Config.CommandPrefix, prefixCmd.Usage)
		} else {
			response = fmt.Sprintf("Command `%s` not found. Use `/help` to see all commands.", cmdName)
		}
	} else {
		// General help
		response = "**Available Slash Commands:**\n"
		for _, cmd := range h.SlashCommands {
			response += fmt.Sprintf("`/%s` - %s\n", cmd.Command.Name, cmd.Command.Description)
		}
		
		response += "\n**Available Prefix Commands:**\n"
		for _, cmd := range h.PrefixCommands {
			response += fmt.Sprintf("`%s%s` - %s\n", h.Bot.Config.CommandPrefix, cmd.Name, cmd.Description)
		}
		
		response += "\nUse `/help command:name` for more information about a specific command."
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: response,
		Color:       0x00AAFF,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Discord Bot Template",
		},
	}

	// Respond to interaction
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// pingSlashCommand handles the ping slash command
func (h *CommandHandler) pingSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Respond immediately
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong! Calculating latency...",
		},
	})

	// Calculate latency
	// Parse the snowflake ID to get the timestamp
	id, _ := strconv.ParseInt(i.Interaction.ID, 10, 64)
	timestamp := (id >> 22) + 1420070400000
	latency := time.Since(time.UnixMilli(timestamp))

	// Edit the response with latency
	content := fmt.Sprintf("Pong! Latency: %s", latency.Round(time.Millisecond))
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
}

// infoSlashCommand handles the info slash command
func (h *CommandHandler) infoSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Get bot stats
	guildsCount := len(h.Bot.GetGuilds())
	uptime := h.Bot.GetUptime().Round(time.Second)

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title: "Bot Information",
		Color: 0x00AAFF,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Servers",
				Value:  fmt.Sprintf("%d", guildsCount),
				Inline: true,
			},
			{
				Name:   "Uptime",
				Value:  uptime.String(),
				Inline: true,
			},
			{
				Name:   "Library",
				Value:  "discordgo",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Discord Bot Template",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Respond to interaction
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// buttonSlashCommand handles the button slash command
func (h *CommandHandler) buttonSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Create a message with buttons
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Here's an example button:",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Click Me",
							Style:    discordgo.PrimaryButton,
							CustomID: "example_button",
							Emoji: discordgo.ComponentEmoji{
								Name: "👋",
							},
						},
						discordgo.Button{
							Label: "Visit Website",
							Style: discordgo.LinkButton,
							URL:   "https://github.com/bwmarrin/discordgo",
							Emoji: discordgo.ComponentEmoji{
								Name: "🔗",
							},
						},
					},
				},
			},
		},
	})
}

// selectSlashCommand handles the select menu slash command
func (h *CommandHandler) selectSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Create a message with a select menu
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Here's an example select menu:",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "example_select",
							Placeholder: "Choose an option",
							Options: []discordgo.SelectMenuOption{
								{
									Label:       "Option 1",
									Value:       "option_1",
									Description: "This is the first option",
									Emoji: discordgo.ComponentEmoji{
										Name: "1️⃣",
									},
								},
								{
									Label:       "Option 2",
									Value:       "option_2",
									Description: "This is the second option",
									Emoji: discordgo.ComponentEmoji{
										Name: "2️⃣",
									},
								},
								{
									Label:       "Option 3",
									Value:       "option_3",
									Description: "This is the third option",
									Emoji: discordgo.ComponentEmoji{
										Name: "3️⃣",
									},
								},
							},
						},
					},
				},
			},
		},
	})
}

// roleSlashCommand handles the role management slash command
func (h *CommandHandler) roleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid command usage.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	subcmd := options[0].Name
	subcmdOptions := options[0].Options

	// Get user and role from options
	var userID, roleID string
	for _, opt := range subcmdOptions {
		switch opt.Name {
		case "user":
			userID = opt.UserValue(s).ID
		case "role":
			roleID = opt.RoleValue(s, i.GuildID).ID
		}
	}

	// Check if we have the required permissions
	perms, err := s.State.UserChannelPermissions(s.State.User.ID, i.ChannelID)
	if err != nil || perms&discordgo.PermissionManageRoles == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "I don't have permission to manage roles.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Handle subcommands
	var responseContent string

	switch subcmd {
	case "add":
		// Add role to user
		err = s.GuildMemberRoleAdd(i.GuildID, userID, roleID)
		if err != nil {
			responseContent = fmt.Sprintf("Error adding role: %v", err)
		} else {
			responseContent = fmt.Sprintf("Added role <@&%s> to <@%s>", roleID, userID)
		}

	case "remove":
		// Remove role from user
		err = s.GuildMemberRoleRemove(i.GuildID, userID, roleID)
		if err != nil {
			responseContent = fmt.Sprintf("Error removing role: %v", err)
		} else {
			responseContent = fmt.Sprintf("Removed role <@&%s> from <@%s>", roleID, userID)
		}

	default:
		responseContent = "Unknown subcommand."
	}

	// Respond to interaction
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: responseContent,
		},
	})
}