# Coral Markets Discord Bot

A Discord bot for Coral Markets that provides automated market announcements, updates, trading alerts, market resolutions, user subscriptions, and channel admin controls.

## Features

- **New Market Announcements**: Instant notifications when new markets are created
- **Market Updates**: Volume- and time-based updates
- **Trading Alerts**: Notifications for trading start/end times
- **Market Resolution Alerts**: Notifications when markets are resolved
- **User Subscriptions**: Subscribe to specific markets or creators
- **Channel Admin Controls**: Configure channel-specific settings
- **Webhook Integration**: Receives real-time notifications from the Coral Markets backend

## Commands

### User Commands
- `/subscribe_market <market_id>` - Subscribe to notifications for a specific market
- `/unsubscribe_market <market_id>` - Unsubscribe from notifications for a specific market
- `/subscribe_creator <creator>` - Subscribe to notifications for a specific creator
- `/unsubscribe_creator <creator>` - Unsubscribe from notifications for a specific creator
- `/list_subscriptions` - List all your current subscriptions
- `/market <market_id>` - Get information about a specific market
- `/help` - Display help information

### Channel Admin Commands
- `/channel_feed_new_markets <on/off>` - Enable or disable new market announcements
- `/channel_feed_categories <categories>` - Set allowed categories (comma-separated)
- `/channel_feed_frequency <low/medium/high>` - Set update frequency
- `/channel_settings` - Display current channel settings

## Installation

1. Clone the repository
2. Navigate to the `discord_bot` directory
3. Run `go mod tidy` to install dependencies
4. Create a `.env` file with your configuration:
   ```
   DISCORD_BOT_TOKEN=your_discord_bot_token_here
   CORAL_BACKEND_URL=your_backend_api_url_here  # Optional
   CORAL_API_KEY=your_api_key_here  # Optional, for webhook authentication
   CORAL_TOKEN=your_bearer_token_here  # Optional, for webhook authentication
   PORT=3000  # Optional, webhook server port (default: 3000)
   ```
5. Run the bot with `go run main.go`

## Webhook Endpoints

The bot exposes the following webhook endpoints to receive notifications from the backend:

- `POST /webhooks/new_market` - New market created
- `POST /webhooks/market_update` - Market updated
- `POST /webhooks/trading_started` - Trading started
- `POST /webhooks/trading_ended` - Trading ended
- `POST /webhooks/market_resolved` - Market resolved

## Architecture

The bot follows a layered architecture pattern:

- **Handlers**: Process Discord slash commands
- **Web**: Handle incoming webhooks from the backend
- **Services**: Business logic implementation
- **Repository**: Data access layer (in-memory implementation)
- **Models**: Data structures
- **Utils**: Utility functions
- **Config**: Configuration management

## Dependencies

- [discordgo](https://github.com/bwmarrin/discordgo) - Discord API wrapper
- [godotenv](https://github.com/joho/godotenv) - Environment variable loader

## Development

To contribute to this project:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request