package bot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// onReady handles the ready event when the bot connects to Discord
func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	logrus.Infof("Bot is ready! Connected as %s#%s", r.User.Username, r.User.Discriminator)

	// Set bot status
	err := s.UpdateGameStatus(0, "Type /help for commands")
	if err != nil {
		logrus.Errorf("Error setting bot status: %v", err)
	}

	// Initialize guilds map
	for _, guild := range r.Guilds {
		b.guildMutex.Lock()
		b.Guilds[guild.ID] = guild
		b.guildMutex.Unlock()
	}

	// Update stats
	b.updateStats()
}

// onGuildCreate handles when the bot joins a new guild
func (b *Bot) onGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	logrus.Infof("Bot joined guild: %s (ID: %s)", g.Name, g.ID)

	// Add guild to map
	b.guildMutex.Lock()
	b.Guilds[g.ID] = g.Guild
	b.guildMutex.Unlock()

	// Update stats
	b.updateStats()

	// Send welcome message to the first text channel we can send to
	for _, channel := range g.Channels {
		if channel.Type == discordgo.ChannelTypeGuildText {
			// Check if we have permission to send messages in this channel
			perms, err := s.State.UserChannelPermissions(s.State.User.ID, channel.ID)
			if err != nil {
				logrus.Warnf("Error checking permissions: %v", err)
				continue
			}

			if perms&discordgo.PermissionSendMessages != 0 {
				// Create welcome embed
				embed := &discordgo.MessageEmbed{
					Title:       "Thanks for adding me!",
					Description: "Use `/help` to see available commands.",
					Color:       0x00AAFF,
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Discord Bot Template",
					},
				}

				_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
				if err != nil {
					logrus.Warnf("Error sending welcome message: %v", err)
				}

				break
			}
		}
	}
}

// onGuildDelete handles when the bot leaves a guild
func (b *Bot) onGuildDelete(s *discordgo.Session, g *discordgo.GuildDelete) {
	logrus.Infof("Bot left guild: %s (ID: %s)", g.Name, g.ID)

	// Remove guild from map
	b.guildMutex.Lock()
	delete(b.Guilds, g.ID)
	b.guildMutex.Unlock()

	// Update stats
	b.updateStats()
}

// onMessageCreate handles when a message is created in a channel the bot has access to
func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if message starts with the command prefix
	if strings.HasPrefix(m.Content, b.Config.CommandPrefix) {
		// Handle prefix command
		cmdString := strings.TrimPrefix(m.Content, b.Config.CommandPrefix)
		cmdName := strings.Fields(cmdString)[0]
		cmdArgs := strings.Fields(cmdString)[1:]

		b.Commands.HandlePrefixCommand(s, m, cmdName, cmdArgs)
	}

	// Handle message reactions (example)
	if strings.Contains(strings.ToLower(m.Content), "hello bot") {
		// React with a wave emoji
		err := s.MessageReactionAdd(m.ChannelID, m.ID, "👋")
		if err != nil {
			logrus.Warnf("Error adding reaction: %v", err)
		}
	}
}

// onInteractionCreate handles Discord interactions (slash commands, buttons, etc.)
func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		// Handle slash command
		b.Commands.HandleSlashCommand(s, i)

	case discordgo.InteractionMessageComponent:
		// Handle button or select menu
		data := i.MessageComponentData()
		switch data.ComponentType {
		case discordgo.ButtonComponent:
			b.handleButtonInteraction(s, i, data)
		case discordgo.SelectMenuComponent:
			b.handleSelectMenuInteraction(s, i, data)
		}
	}
}

// handleButtonInteraction handles button click interactions
func (b *Bot) handleButtonInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.MessageComponentInteractionData) {
	// Log the interaction
	guildID := ""
	if i.GuildID != "" {
		guildID = i.GuildID
	}

	interactionData := map[string]interface{}{
		"custom_id": data.CustomID,
	}

	err := b.Repository.LogInteraction(guildID, i.ChannelID, i.Member.User.ID, "button", data.CustomID, interactionData)
	if err != nil {
		logrus.Errorf("Error logging button interaction: %v", err)
	}

	// Handle different button IDs
	switch data.CustomID {
	case "example_button":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You clicked the example button!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

	default:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown button interaction.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

// handleSelectMenuInteraction handles select menu interactions
func (b *Bot) handleSelectMenuInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.MessageComponentInteractionData) {
	// Log the interaction
	guildID := ""
	if i.GuildID != "" {
		guildID = i.GuildID
	}

	interactionData := map[string]interface{}{
		"custom_id": data.CustomID,
		"values":    data.Values,
	}

	err := b.Repository.LogInteraction(guildID, i.ChannelID, i.Member.User.ID, "select_menu", data.CustomID, interactionData)
	if err != nil {
		logrus.Errorf("Error logging select menu interaction: %v", err)
	}

	// Handle different select menu IDs
	switch data.CustomID {
	case "example_select":
		response := "You selected: " + strings.Join(data.Values, ", ")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

	default:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown select menu interaction.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}