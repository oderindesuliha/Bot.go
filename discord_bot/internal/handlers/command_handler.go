package handlers

import (
	"fmt"
	"strings"

	"coral-bot/discord_bot/internal/services"
	"coral-bot/discord_bot/internal/utils"

	"github.com/bwmarrin/discordgo"
)

// CommandHandler handles Discord slash commands
type CommandHandler struct {
	marketService       services.MarketService
	subscriptionService services.SubscriptionService
	logger              *utils.Logger
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(
	marketService services.MarketService,
	subscriptionService services.SubscriptionService,
	logger *utils.Logger,
) *CommandHandler {
	return &CommandHandler{
		marketService:       marketService,
		subscriptionService: subscriptionService,
		logger:              logger,
	}
}

// RegisterCommands registers all slash commands with Discord
func (h *CommandHandler) RegisterCommands(session *discordgo.Session) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "subscribe_market",
			Description: "Subscribe to notifications for a specific market",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "market_id",
					Description: "The ID of the market to subscribe to",
					Required:    true,
				},
			},
		},
		{
			Name:        "unsubscribe_market",
			Description: "Unsubscribe from notifications for a specific market",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "market_id",
					Description: "The ID of the market to unsubscribe from",
					Required:    true,
				},
			},
		},
		{
			Name:        "subscribe_creator",
			Description: "Subscribe to notifications for a specific creator",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "creator",
					Description: "The name of the creator to subscribe to",
					Required:    true,
				},
			},
		},
		{
			Name:        "unsubscribe_creator",
			Description: "Unsubscribe from notifications for a specific creator",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "creator",
					Description: "The name of the creator to unsubscribe from",
					Required:    true,
				},
			},
		},
		{
			Name:        "list_subscriptions",
			Description: "List all your current subscriptions",
		},
		{
			Name:        "market",
			Description: "Get information about a specific market",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "market_id",
					Description: "The ID of the market to get information for",
					Required:    true,
				},
			},
		},
		{
			Name:        "help",
			Description: "Display help information",
		},
		{
			Name:        "channel_feed_new_markets",
			Description: "Enable or disable new market announcements in this channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "setting",
					Description: "on or off",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "on", Value: "on"},
						{Name: "off", Value: "off"},
					},
				},
			},
		},
		{
			Name:        "channel_feed_categories",
			Description: "Set allowed categories for market feed in this channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "categories",
					Description: "Comma-separated list of allowed categories",
					Required:    true,
				},
			},
		},
		{
			Name:        "channel_feed_frequency",
			Description: "Set the frequency of market updates in this channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "level",
					Description: "low, medium, or high",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "low", Value: "low"},
						{Name: "medium", Value: "medium"},
						{Name: "high", Value: "high"},
					},
				},
			},
		},
		{
			Name:        "channel_settings",
			Description: "Display current channel settings",
		},
	}

	h.logger.Info("Registering commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, cmd := range commands {
		rc, err := session.ApplicationCommandCreate(session.State.User.ID, "", cmd)
		if err != nil {
			h.logger.Error(fmt.Sprintf("Cannot create '%s' command: %v", cmd.Name, err))
			return err
		}
		registeredCommands[i] = rc
	}

	h.logger.Info(fmt.Sprintf("Successfully registered %d commands", len(registeredCommands)))
	return nil
}

// HandleInteraction handles incoming slash command interactions
func (h *CommandHandler) HandleInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}

	command := interaction.ApplicationCommandData()
	userID := interaction.Member.User.ID

	h.logger.Info(fmt.Sprintf("Handling command: %s from user: %s", command.Name, userID))

	switch command.Name {
	case "subscribe_market":
		h.handleSubscribeMarket(session, interaction, userID, command.Options[0].StringValue())
	case "unsubscribe_market":
		h.handleUnsubscribeMarket(session, interaction, userID, command.Options[0].StringValue())
	case "subscribe_creator":
		h.handleSubscribeCreator(session, interaction, userID, command.Options[0].StringValue())
	case "unsubscribe_creator":
		h.handleUnsubscribeCreator(session, interaction, userID, command.Options[0].StringValue())
	case "list_subscriptions":
		h.handleListSubscriptions(session, interaction, userID)
	case "market":
		h.handleGetMarket(session, interaction, command.Options[0].StringValue())
	case "help":
		h.handleHelp(session, interaction)
	case "channel_feed_new_markets":
		h.handleChannelFeedNewMarkets(session, interaction, interaction.ChannelID, command.Options[0].StringValue())
	case "channel_feed_categories":
		h.handleChannelFeedCategories(session, interaction, interaction.ChannelID, command.Options[0].StringValue())
	case "channel_feed_frequency":
		h.handleChannelFeedFrequency(session, interaction, interaction.ChannelID, command.Options[0].StringValue())
	case "channel_settings":
		h.handleChannelSettings(session, interaction, interaction.ChannelID)
	default:
		h.respondToInteraction(session, interaction, "Unknown command")
	}
}

// handleSubscribeMarket handles the subscribe_market command
func (h *CommandHandler) handleSubscribeMarket(session *discordgo.Session, interaction *discordgo.InteractionCreate, userID, marketID string) {
	err := h.subscriptionService.SubscribeToMarket(userID, marketID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to subscribe user %s to market %s: %v", userID, marketID, err))
		h.respondToInteraction(session, interaction, "Failed to subscribe to market")
		return
	}

	response := fmt.Sprintf("You have been subscribed to market `%s`", marketID)
	h.respondToInteraction(session, interaction, response)
}

// handleUnsubscribeMarket handles the unsubscribe_market command
func (h *CommandHandler) handleUnsubscribeMarket(session *discordgo.Session, interaction *discordgo.InteractionCreate, userID, marketID string) {
	err := h.subscriptionService.UnsubscribeFromMarket(userID, marketID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to unsubscribe user %s from market %s: %v", userID, marketID, err))
		h.respondToInteraction(session, interaction, "Failed to unsubscribe from market")
		return
	}

	response := fmt.Sprintf("You have been unsubscribed from market `%s`", marketID)
	h.respondToInteraction(session, interaction, response)
}

// handleSubscribeCreator handles the subscribe_creator command
func (h *CommandHandler) handleSubscribeCreator(session *discordgo.Session, interaction *discordgo.InteractionCreate, userID, creator string) {
	err := h.subscriptionService.SubscribeToCreator(userID, creator)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to subscribe user %s to creator %s: %v", userID, creator, err))
		h.respondToInteraction(session, interaction, "Failed to subscribe to creator")
		return
	}

	response := fmt.Sprintf("You have been subscribed to creator `%s`", creator)
	h.respondToInteraction(session, interaction, response)
}

// handleUnsubscribeCreator handles the unsubscribe_creator command
func (h *CommandHandler) handleUnsubscribeCreator(session *discordgo.Session, interaction *discordgo.InteractionCreate, userID, creator string) {
	err := h.subscriptionService.UnsubscribeFromCreator(userID, creator)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to unsubscribe user %s from creator %s: %v", userID, creator, err))
		h.respondToInteraction(session, interaction, "Failed to unsubscribe from creator")
		return
	}

	response := fmt.Sprintf("You have been unsubscribed from creator `%s`", creator)
	h.respondToInteraction(session, interaction, response)
}

// handleListSubscriptions handles the list_subscriptions command
func (h *CommandHandler) handleListSubscriptions(session *discordgo.Session, interaction *discordgo.InteractionCreate, userID string) {
	subscription, err := h.subscriptionService.GetUserSubscriptions(userID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get subscriptions for user %s: %v", userID, err))
		h.respondToInteraction(session, interaction, "Failed to retrieve subscriptions")
		return
	}

	if len(subscription.SubscribedMarkets) == 0 && len(subscription.SubscribedCreators) == 0 {
		h.respondToInteraction(session, interaction, "You have no subscriptions")
		return
	}

	var response strings.Builder
	response.WriteString("**Your Subscriptions:**\n\n")

	if len(subscription.SubscribedMarkets) > 0 {
		response.WriteString("**Markets:**\n")
		for _, marketID := range subscription.SubscribedMarkets {
			response.WriteString(fmt.Sprintf("- `%s`\n", marketID))
		}
		response.WriteString("\n")
	}

	if len(subscription.SubscribedCreators) > 0 {
		response.WriteString("**Creators:**\n")
		for _, creator := range subscription.SubscribedCreators {
			response.WriteString(fmt.Sprintf("- `%s`\n", creator))
		}
	}

	h.respondToInteraction(session, interaction, response.String())
}

// handleGetMarket handles the market command
func (h *CommandHandler) handleGetMarket(session *discordgo.Session, interaction *discordgo.InteractionCreate, marketID string) {
	market, err := h.marketService.FetchMarket(marketID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to fetch market %s: %v", marketID, err))
		h.respondToInteraction(session, interaction, "Failed to retrieve market information")
		return
	}

	announcement := h.marketService.CreateMarketAnnouncement(market)
	h.respondToInteraction(session, interaction, announcement)
}

// handleHelp handles the help command
func (h *CommandHandler) handleHelp(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	helpText := "**Coral Markets Bot Help**\n\n" +
		"**User Commands:**\n" +
		"- `/subscribe_market <market_id>` - Subscribe to notifications for a specific market\n" +
		"- `/unsubscribe_market <market_id>` - Unsubscribe from notifications for a specific market\n" +
		"- `/subscribe_creator <creator>` - Subscribe to notifications for a specific creator\n" +
		"- `/unsubscribe_creator <creator>` - Unsubscribe from notifications for a specific creator\n" +
		"- `/list_subscriptions` - List all your current subscriptions\n" +
		"- `/market <market_id>` - Get information about a specific market\n" +
		"- `/help` - Display this help message\n\n" +
		"**Channel Admin Commands:**\n" +
		"- `/channel_feed_new_markets <on/off>` - Enable or disable new market announcements\n" +
		"- `/channel_feed_categories <categories>` - Set allowed categories (comma-separated)\n" +
		"- `/channel_feed_frequency <low/medium/high>` - Set update frequency\n" +
		"- `/channel_settings` - Display current channel settings\n\n" +
		"You'll receive notifications for markets and creators you're subscribed to based on your preferences."

	h.respondToInteraction(session, interaction, helpText)
}

// handleChannelFeedNewMarkets handles the channel_feed_new_markets command
func (h *CommandHandler) handleChannelFeedNewMarkets(session *discordgo.Session, interaction *discordgo.InteractionCreate, channelID, setting string) {
	config, err := h.subscriptionService.GetChannelConfig(channelID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get channel config for %s: %v", channelID, err))
		h.respondToInteraction(session, interaction, "Failed to update channel settings")
		return
	}

	enabled := setting == "on"
	config.FeedEnabled = enabled

	err = h.subscriptionService.UpdateChannelConfig(config)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to update channel config for %s: %v", channelID, err))
		h.respondToInteraction(session, interaction, "Failed to update channel settings")
		return
	}

	response := fmt.Sprintf("New market announcements have been turned %s for this channel", setting)
	h.respondToInteraction(session, interaction, response)
}

// handleChannelFeedCategories handles the channel_feed_categories command
func (h *CommandHandler) handleChannelFeedCategories(session *discordgo.Session, interaction *discordgo.InteractionCreate, channelID, categories string) {
	config, err := h.subscriptionService.GetChannelConfig(channelID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get channel config for %s: %v", channelID, err))
		h.respondToInteraction(session, interaction, "Failed to update channel settings")
		return
	}

	categoryList := strings.Split(categories, ",")
	for i, category := range categoryList {
		categoryList[i] = strings.TrimSpace(category)
	}

	config.AllowedCategories = categoryList

	err = h.subscriptionService.UpdateChannelConfig(config)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to update channel config for %s: %v", channelID, err))
		h.respondToInteraction(session, interaction, "Failed to update channel settings")
		return
	}

	response := fmt.Sprintf("Allowed categories have been set to: %s", strings.Join(categoryList, ", "))
	h.respondToInteraction(session, interaction, response)
}

// handleChannelFeedFrequency handles the channel_feed_frequency command
func (h *CommandHandler) handleChannelFeedFrequency(session *discordgo.Session, interaction *discordgo.InteractionCreate, channelID, frequency string) {
	config, err := h.subscriptionService.GetChannelConfig(channelID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get channel config for %s: %v", channelID, err))
		h.respondToInteraction(session, interaction, "Failed to update channel settings")
		return
	}

	config.FrequencyMode = frequency

	err = h.subscriptionService.UpdateChannelConfig(config)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to update channel config for %s: %v", channelID, err))
		h.respondToInteraction(session, interaction, "Failed to update channel settings")
		return
	}

	response := fmt.Sprintf("Update frequency has been set to: %s", frequency)
	h.respondToInteraction(session, interaction, response)
}

// handleChannelSettings handles the channel_settings command
func (h *CommandHandler) handleChannelSettings(session *discordgo.Session, interaction *discordgo.InteractionCreate, channelID string) {
	config, err := h.subscriptionService.GetChannelConfig(channelID)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get channel config for %s: %v", channelID, err))
		h.respondToInteraction(session, interaction, "Failed to retrieve channel settings")
		return
	}

	settings := fmt.Sprintf("**Channel Settings**\n\n"+
		"New Market Announcements: %s\n"+
		"Allowed Categories: %s\n"+
		"Update Frequency: %s\n"+
		"Last Update: %s",
		map[bool]string{true: "Enabled", false: "Disabled"}[config.FeedEnabled],
		func() string {
			if len(config.AllowedCategories) == 0 {
				return "All categories"
			}
			return strings.Join(config.AllowedCategories, ", ")
		}(),
		config.FrequencyMode,
		config.LastUpdateTimestamp.Format("2006-01-02 15:04:05"),
	)

	h.respondToInteraction(session, interaction, settings)
}

// respondToInteraction sends a response to a Discord interaction
func (h *CommandHandler) respondToInteraction(session *discordgo.Session, interaction *discordgo.InteractionCreate, message string) {
	err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to respond to interaction: %v", err))
	}
}
