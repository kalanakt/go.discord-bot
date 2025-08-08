# Go Discord Bot Template

A complete, production-ready Discord bot template built with Go, featuring a modular structure, comprehensive Discord features, and PostgreSQL integration.

## Features

### Discord Bot
- Built with [discordgo](https://github.com/bwmarrin/discordgo)
- Structured command system (both prefix and slash commands)
- Event listeners (onReady, onMessageCreate, onGuildJoin, etc.)
- Rich embeds, button & select menu interactions
- Reaction collectors
- Voice connection & audio playback
- Permission checks & role management
- Logging & error handling

### Database
- PostgreSQL for storing logs and runtime data
- [goose](https://github.com/pressly/goose) for schema migrations

### Deployment
- Docker and docker-compose for local development
- Heroku and Railway deployment support

## Project Structure

```
├── bot/                  # Discord bot implementation
│   ├── bot.go            # Bot initialization and core functionality
│   ├── commands.go       # Command handler and registration
│   ├── command_handlers.go # Command implementation
│   ├── events.go         # Event handlers
│   └── voice.go          # Voice functionality
├── config/               # Configuration handling
│   └── config.go         # Environment variable loading
├── database/             # Database functionality
│   ├── database.go       # Connection and migration
│   ├── migrations/       # SQL migration files
│   └── repository.go     # Data access layer
├── main.go               # Application entry point
├── Dockerfile            # Docker configuration
├── docker-compose.yml    # Docker Compose configuration
├── Procfile              # Heroku configuration
└── .env.example          # Example environment variables
```

## Setup Instructions

### Prerequisites

- Go 1.18 or higher
- PostgreSQL
- Discord Bot Token (from [Discord Developer Portal](https://discord.com/developers/applications))

### Local Setup

1. Clone the repository

```bash
git clone https://github.com/kalanakt/go-discord-bot.git
cd go-discord-bot
```

2. Create a `.env` file based on `.env.example`

```bash
cp .env.example .env
```

3. Edit the `.env` file with your Discord bot token and other settings

4. Install dependencies

```bash
go mod download
```

5. Run the application

```bash
go run main.go
```

### PostgreSQL Setup

1. Create a PostgreSQL database

```bash
psql -U postgres
CREATE DATABASE discord_bot_db;
CREATE USER discord_bot WITH ENCRYPTED PASSWORD 'discord_bot_password';
GRANT ALL PRIVILEGES ON DATABASE discord_bot_db TO discord_bot;
\q
```

2. Update the `DB_URL` in your `.env` file

3. Migrations will run automatically when the application starts

### Docker Setup

1. Make sure Docker and Docker Compose are installed

2. Create a `.env` file based on `.env.example`

3. Build and start the containers

```bash
docker-compose up -d
```

## Running Locally

### Without Docker

```bash
go run main.go
```

The bot will connect to Discord.

### With Docker

```bash
docker-compose up
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|--------|
| BOT_TOKEN | Discord Bot Token | (required) |
| BOT_PREFIX | Command prefix | ! |
| BOT_DEV_MODE | Development mode | true |
| BOT_DEV_GUILD_ID | Guild ID for dev commands | (optional) |
| DB_URL | PostgreSQL connection URL | (required) |

## Deployment

### Heroku

1. Create a Heroku account and install the Heroku CLI

2. Create a new Heroku app

```bash
heroku create your-app-name
```

3. Add the PostgreSQL addon

```bash
heroku addons:create heroku-postgresql:hobby-dev
```

4. Set the environment variables

```bash
heroku config:set BOT_TOKEN=your_discord_bot_token
heroku config:set BOT_DEV_MODE=false
```

5. Deploy the application

```bash
git push heroku main
```

### Railway

1. Create a Railway account and install the Railway CLI

2. Initialize a new Railway project

```bash
railway init
```

3. Add a PostgreSQL database

```bash
railway add
```

4. Set the environment variables in the Railway console

5. Deploy the application

```bash
railway up
```

## Extending the Bot

### Adding New Commands

1. Open `bot/command_handlers.go`

2. Add a new command handler function

```go
func (b *Bot) handleNewCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
    // Command implementation
}
```

3. Register the command in `bot/commands.go` in the `registerCommands` function

```go
ch.registerPrefixCommand("newcomm", "Description of new command", b.handleNewCommand)
```

### Adding New Slash Commands

1. Open `bot/command_handlers.go`

2. Add a new slash command handler function

```go
func (b *Bot) handleNewSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    // Slash command implementation
}
```

3. Define the command in `bot/commands.go` in the `registerSlashCommands` function

```go
slashCommands = append(slashCommands, &discordgo.ApplicationCommand{
    Name:        "newcomm",
    Description: "Description of new command",
    Options: []*discordgo.ApplicationCommandOption{
        // Command options
    },
})
```

4. Register the handler in the `registerSlashCommandHandlers` function

```go
ch.slashCommandHandlers["newcomm"] = b.handleNewSlashCommand
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.