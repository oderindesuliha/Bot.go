package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"coral-bot/discord_bot/internal/models"
	"coral-bot/discord_bot/internal/services"
	"coral-bot/discord_bot/internal/utils"

	"github.com/bwmarrin/discordgo"
)

// WebhookHandler handles incoming webhooks from the backend
type WebhookHandler struct {
	marketService       services.MarketService
	subscriptionService services.SubscriptionService
	logger              *utils.Logger
	discordSession      *discordgo.Session // Store the Discord session to send messages
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	marketService services.MarketService,
	subscriptionService services.SubscriptionService,
	logger *utils.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		marketService:       marketService,
		subscriptionService: subscriptionService,
		logger:              logger,
	}
}

// SetDiscordSession sets the Discord session for sending messages
func (h *WebhookHandler) SetDiscordSession(session *discordgo.Session) {
	h.discordSession = session
}

// AuthOk checks if the request is properly authenticated
func (h *WebhookHandler) AuthOk(r *http.Request) bool {
	apiKey := r.Header.Get("X-API-Key")
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && bearer[:7] == "Bearer " {
		bearer = bearer[7:]
	}

	requiredAPIKey := os.Getenv("CORAL_API_KEY")
	requiredToken := os.Getenv("CORAL_TOKEN")

	if requiredAPIKey != "" && apiKey == requiredAPIKey {
		return true
	}
	if requiredToken != "" && bearer == requiredToken {
		return true
	}
	return requiredAPIKey == "" && requiredToken == ""
}

// sendToSubscribedChannels sends a message to all subscribed channels
func (h *WebhookHandler) sendToSubscribedChannels(message string, market *models.Market) {
	if h.discordSession == nil {
		h.logger.Error("Discord session not set")
		return
	}

	// Get all channel configurations
	channels, err := h.subscriptionService.GetAllChannelConfigs()
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get channel configs: %v", err))
		return
	}

	for _, channelConfig := range channels {
		// Check if feed is enabled for this channel
		if !channelConfig.FeedEnabled {
			continue
		}

		// Check if market category is allowed
		if len(channelConfig.AllowedCategories) > 0 {
			allowed := false
			for _, category := range channelConfig.AllowedCategories {
				if category == market.Category {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}

		// Send message to channel
		_, err := h.discordSession.ChannelMessageSend(channelConfig.ChannelID, message)
		if err != nil {
			h.logger.Error(fmt.Sprintf("Failed to send message to channel %s: %v", channelConfig.ChannelID, err))
		} else {
			h.logger.Info(fmt.Sprintf("Sent message to channel %s", channelConfig.ChannelID))
		}
	}
}

// sendToSubscribedUsers sends a DM to all subscribed users
func (h *WebhookHandler) sendToSubscribedUsers(message string, market *models.Market) {
	if h.discordSession == nil {
		h.logger.Error("Discord session not set")
		return
	}

	// Get all subscriptions
	subscriptions, err := h.subscriptionService.GetAllSubscriptions()
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to get subscriptions: %v", err))
		return
	}

	for _, subscription := range subscriptions {
		// Check if user is subscribed to this market or creator
		shouldNotify := h.subscriptionService.ShouldNotifyUser(subscription, market)
		if !shouldNotify {
			continue
		}

		// Send DM to user
		channel, err := h.discordSession.UserChannelCreate(subscription.DiscordUserID)
		if err != nil {
			h.logger.Error(fmt.Sprintf("Failed to create DM channel for user %s: %v", subscription.DiscordUserID, err))
			continue
		}

		_, err = h.discordSession.ChannelMessageSend(channel.ID, message)
		if err != nil {
			h.logger.Error(fmt.Sprintf("Failed to send DM to user %s: %v", subscription.DiscordUserID, err))
		} else {
			h.logger.Info(fmt.Sprintf("Sent DM to user %s", subscription.DiscordUserID))
		}
	}
}

// HandleNewMarket handles the new_market webhook
func (h *WebhookHandler) HandleNewMarket(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		EventType string        `json:"event_type"`
		Market    models.Market `json:"market"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if payload.EventType != "new_market" {
		h.logger.Error("Invalid event type")
		http.Error(w, `{"error": "Invalid event type"}`, http.StatusBadRequest)
		return
	}

	// Create announcement message
	announcement := h.marketService.CreateMarketAnnouncement(&payload.Market)
	h.logger.Info(fmt.Sprintf("New market announcement: %s", announcement))

	// Send to subscribed channels and users
	h.sendToSubscribedChannels(announcement, &payload.Market)
	h.sendToSubscribedUsers(announcement, &payload.Market)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok": true}`))
}

// HandleMarketUpdate handles the market_update webhook
func (h *WebhookHandler) HandleMarketUpdate(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		EventType string        `json:"event_type"`
		Market    models.Market `json:"market"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if payload.EventType != "market_update" {
		h.logger.Error("Invalid event type")
		http.Error(w, `{"error": "Invalid event type"}`, http.StatusBadRequest)
		return
	}

	// Create update message
	updateMessage := h.marketService.CreateMarketUpdateMessage(&payload.Market)
	h.logger.Info(fmt.Sprintf("Market update: %s", updateMessage))

	// Send to subscribed channels and users
	h.sendToSubscribedChannels(updateMessage, &payload.Market)
	h.sendToSubscribedUsers(updateMessage, &payload.Market)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok": true}`))
}

// HandleTradingStarted handles the trading_started webhook
func (h *WebhookHandler) HandleTradingStarted(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		EventType string        `json:"event_type"`
		Market    models.Market `json:"market"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if payload.EventType != "trading_started" {
		h.logger.Error("Invalid event type")
		http.Error(w, `{"error": "Invalid event type"}`, http.StatusBadRequest)
		return
	}

	// Create trading start message
	startMessage := h.marketService.CreateTradingStartMessage(&payload.Market)
	h.logger.Info(fmt.Sprintf("Trading started: %s", startMessage))

	// Send to subscribed channels and users
	h.sendToSubscribedChannels(startMessage, &payload.Market)
	h.sendToSubscribedUsers(startMessage, &payload.Market)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok": true}`))
}

// HandleTradingEnded handles the trading_ended webhook
func (h *WebhookHandler) HandleTradingEnded(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		EventType string        `json:"event_type"`
		Market    models.Market `json:"market"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if payload.EventType != "trading_ended" {
		h.logger.Error("Invalid event type")
		http.Error(w, `{"error": "Invalid event type"}`, http.StatusBadRequest)
		return
	}

	// Create trading end message
	endMessage := h.marketService.CreateTradingEndMessage(&payload.Market)
	h.logger.Info(fmt.Sprintf("Trading ended: %s", endMessage))

	// Send to subscribed channels and users
	h.sendToSubscribedChannels(endMessage, &payload.Market)
	h.sendToSubscribedUsers(endMessage, &payload.Market)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok": true}`))
}

// HandleMarketResolved handles the market_resolved webhook
func (h *WebhookHandler) HandleMarketResolved(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		EventType string        `json:"event_type"`
		Market    models.Market `json:"market"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if payload.EventType != "market_resolved" {
		h.logger.Error("Invalid event type")
		http.Error(w, `{"error": "Invalid event type"}`, http.StatusBadRequest)
		return
	}

	// Create resolution message
	resolutionMessage := h.marketService.CreateMarketResolutionMessage(&payload.Market)
	h.logger.Info(fmt.Sprintf("Market resolved: %s", resolutionMessage))

	// Send to subscribed channels and users
	h.sendToSubscribedChannels(resolutionMessage, &payload.Market)
	h.sendToSubscribedUsers(resolutionMessage, &payload.Market)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok": true}`))
}

// StartWebServer starts the webhook server
func (h *WebhookHandler) StartWebServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhooks/new_market", h.HandleNewMarket)
	mux.HandleFunc("/webhooks/market_update", h.HandleMarketUpdate)
	mux.HandleFunc("/webhooks/trading_started", h.HandleTradingStarted)
	mux.HandleFunc("/webhooks/trading_ended", h.HandleTradingEnded)
	mux.HandleFunc("/webhooks/market_resolved", h.HandleMarketResolved)

	h.logger.Info(fmt.Sprintf("Starting webhook server on port %s", port))
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to start webhook server: %v", err))
	}
}
