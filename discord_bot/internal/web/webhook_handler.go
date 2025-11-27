package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		EventType string        `json:"event_type"`
		Market    models.Market `json:"market"`
	}

	if err := json.Unmarshal(requestBody, &payload); err != nil {
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

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		EventType string        `json:"event_type"`
		Market    models.Market `json:"market"`
	}

	if err := json.Unmarshal(requestBody, &payload); err != nil {
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

// HandleRegisterWebhook handles POST /discord/webhooks/register
func (h *WebhookHandler) HandleRegisterWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		ChannelID         string   `json:"channel_id"`
		WebhookURL        string   `json:"webhook_url"`
		Events            []string `json:"events"`
		Frequency         string   `json:"frequency"`
		AllowedCategories []string `json:"allowed_categories"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if payload.ChannelID == "" || payload.WebhookURL == "" {
		http.Error(w, `{"error": "channel_id and webhook_url are required"}`, http.StatusBadRequest)
		return
	}

	// create registration
	reg := &models.WebhookRegistration{
		ChannelID:         payload.ChannelID,
		WebhookURL:        payload.WebhookURL,
		Events:            payload.Events,
		Frequency:         payload.Frequency,
		AllowedCategories: payload.AllowedCategories,
	}

	saved, err := h.subscriptionService.RegisterWebhook(reg)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to save webhook registration: %v", err))
		http.Error(w, `{"error": "Failed to register webhook"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if b, err := json.Marshal(saved); err == nil {
		w.Write(b)
	} else {
		w.Write([]byte(`{"ok": true}`))
	}
}

func (h *WebhookHandler) HandleUnregisterWebhookByPath(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, `{"error": "id required"}`, http.StatusBadRequest)
		return
	}
	id := parts[len(parts)-1]
	if id == "" {
		http.Error(w, `{"error": "id required"}`, http.StatusBadRequest)
		return
	}
	if err := h.subscriptionService.UnregisterWebhook(id); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to unregister webhook: %v", err))
		http.Error(w, `{"error": "Failed to unregister webhook"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WebhookHandler) HandleEventNewMarket(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	var payload struct {
		MarketID    string `json:"market_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Creator     string `json:"creator"`
		Category    string `json:"category"`
		Outcomes    []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"outcomes"`
		StartTime string  `json:"start_time"`
		EndTime   string  `json:"end_time"`
		Volume    float64 `json:"volume"`
		Link      string  `json:"link"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	st, _ := time.Parse(time.RFC3339, payload.StartTime)
	et, _ := time.Parse(time.RFC3339, payload.EndTime)
	outs := make([]string, 0, len(payload.Outcomes))
	for _, o := range payload.Outcomes {
		outs = append(outs, o.Name)
	}
	market := models.Market{
		ID:          payload.MarketID,
		Title:       payload.Title,
		Description: payload.Description,
		Outcomes:    outs,
		Percentages: []float64{},
		Category:    payload.Category,
		Creator:     payload.Creator,
		Volume:      payload.Volume,
		StartTime:   st,
		EndTime:     et,
		Status:      "active",
		Link:        payload.Link,
	}
	msg := h.marketService.CreateMarketAnnouncement(&market)
	h.sendToSubscribedChannels(msg, &market)
	h.sendToSubscribedUsers(msg, &market)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"accepted": true}`))
}

func (h *WebhookHandler) HandleEventMarketUpdate(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	var payload struct {
		MarketID       string  `json:"market_id"`
		Title          string  `json:"title"`
		Volume         float64 `json:"volume"`
		VolumeDeltaPct float64 `json:"volume_delta_pct"`
		TimeLeft       string  `json:"time_left"`
		EndTime        string  `json:"end_time"`
		Link           string  `json:"link"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	et, _ := time.Parse(time.RFC3339, payload.EndTime)
	market := models.Market{
		ID:      payload.MarketID,
		Title:   payload.Title,
		Volume:  payload.Volume,
		EndTime: et,
		Status:  "active",
		Link:    payload.Link,
	}
	msg := h.marketService.CreateMarketUpdateMessage(&market)
	h.sendToSubscribedChannels(msg, &market)
	h.sendToSubscribedUsers(msg, &market)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"accepted": true}`))
}

func (h *WebhookHandler) HandleEventTradingStart(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	var eventPayload struct {
		MarketID      string   `json:"market_id"`
		Title         string   `json:"title"`
		Description   string   `json:"description"`
		Duration      string   `json:"duration"`
		OutcomesCount int      `json:"outcomes_count"`
		Outcomes      []string `json:"outcomes"`
		Link          string   `json:"link"`
	}
	if err := json.Unmarshal(requestBody, &eventPayload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	market := models.Market{ID: eventPayload.MarketID, Title: eventPayload.Title, Description: eventPayload.Description, Outcomes: eventPayload.Outcomes, Link: eventPayload.Link}
	messageBody := h.marketService.CreateTradingStartMessage(&market)
	h.sendToSubscribedChannels(messageBody, &market)
	h.sendToSubscribedUsers(messageBody, &market)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"accepted": true}`))
}

func (h *WebhookHandler) HandleEventTradingEnd(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	var eventPayload struct {
		MarketID    string `json:"market_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Outcomes    []struct {
			ID   string  `json:"id"`
			Name string  `json:"name"`
			Pct  float64 `json:"pct"`
		} `json:"outcomes"`
		FinalPool float64 `json:"final_pool"`
		Link      string  `json:"link"`
	}
	if err := json.Unmarshal(requestBody, &eventPayload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	outcomeNames := make([]string, 0, len(eventPayload.Outcomes))
	for _, outcome := range eventPayload.Outcomes {
		outcomeNames = append(outcomeNames, outcome.Name)
	}
	market := models.Market{ID: eventPayload.MarketID, Title: eventPayload.Title, Description: eventPayload.Description, Outcomes: outcomeNames, Volume: eventPayload.FinalPool, Link: eventPayload.Link}
	messageBody := h.marketService.CreateTradingEndMessage(&market)
	h.sendToSubscribedChannels(messageBody, &market)
	h.sendToSubscribedUsers(messageBody, &market)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"accepted": true}`))
}

func (h *WebhookHandler) HandleEventMarketResolved(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	var payload struct {
		MarketID       string  `json:"market_id"`
		Title          string  `json:"title"`
		WinningOutcome string  `json:"winning_outcome"`
		TotalPool      float64 `json:"total_pool"`
		Link           string  `json:"link"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	market := models.Market{ID: payload.MarketID, Title: payload.Title, ResolvedOutcome: payload.WinningOutcome, Volume: payload.TotalPool, Link: payload.Link}
	msg := h.marketService.CreateMarketResolutionMessage(&market)
	h.sendToSubscribedChannels(msg, &market)
	h.sendToSubscribedUsers(msg, &market)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"accepted": true}`))
}

func (h *WebhookHandler) HandleEventMarketBuy(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	var payload struct {
		MarketID string  `json:"market_id"`
		Title    string  `json:"title"`
		Amount   float64 `json:"amount"`
		Outcome  string  `json:"outcome"`
		Buyer    string  `json:"buyer"`
		Link     string  `json:"link"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	msg := h.marketService.CreateMarketBuyMessage(payload.MarketID, payload.Title, payload.Amount, payload.Outcome, payload.Buyer, payload.Link)
	market := models.Market{ID: payload.MarketID, Title: payload.Title, Link: payload.Link}
	h.sendToSubscribedChannels(msg, &market)
	h.sendToSubscribedUsers(msg, &market)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"accepted": true}`))
}

func (h *WebhookHandler) HandleSubscribeMarket(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		DiscordUserID string `json:"discord_user_id"`
		MarketID      string `json:"market_id"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if err := h.subscriptionService.SubscribeToMarket(payload.DiscordUserID, payload.MarketID); err != nil {
		http.Error(w, `{"error": "Failed to subscribe"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"subscribed": true}`))
}

func (h *WebhookHandler) HandleUnsubscribeMarket(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		DiscordUserID string `json:"discord_user_id"`
		MarketID      string `json:"market_id"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if err := h.subscriptionService.UnsubscribeFromMarket(payload.DiscordUserID, payload.MarketID); err != nil {
		http.Error(w, `{"error": "Failed to unsubscribe"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"subscribed": false}`))
}

func (h *WebhookHandler) HandleSubscribeCreator(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		DiscordUserID string `json:"discord_user_id"`
		CreatorID     string `json:"creator_id"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if err := h.subscriptionService.SubscribeToCreator(payload.DiscordUserID, payload.CreatorID); err != nil {
		http.Error(w, `{"error": "Failed to subscribe"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) HandleUnsubscribeCreator(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		DiscordUserID string `json:"discord_user_id"`
		CreatorID     string `json:"creator_id"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if err := h.subscriptionService.UnsubscribeFromCreator(payload.DiscordUserID, payload.CreatorID); err != nil {
		http.Error(w, `{"error": "Failed to unsubscribe"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) HandleGetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, `{"error": "discord_user_id required"}`, http.StatusBadRequest)
		return
	}
	discordUserID := parts[len(parts)-1]
	sub, err := h.subscriptionService.GetUserSubscriptions(discordUserID)
	if err != nil {
		http.Error(w, `{"error": "Failed to get subscriptions"}`, http.StatusInternalServerError)
		return
	}
	resp := struct {
		Markets  []string `json:"markets"`
		Creators []string `json:"creators"`
	}{Markets: sub.SubscribedMarkets, Creators: sub.SubscribedCreators}
	b, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *WebhookHandler) HandleChannelFeedNewMarkets(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		ChannelID string `json:"channel_id"`
		Enabled   bool   `json:"enabled"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	cfg, err := h.subscriptionService.GetChannelConfig(payload.ChannelID)
	if err != nil {
		http.Error(w, `{"error": "Failed to load config"}`, http.StatusInternalServerError)
		return
	}
	cfg.ChannelID = payload.ChannelID
	cfg.FeedEnabled = payload.Enabled
	cfg.LastUpdateTimestamp = time.Now()
	if err := h.subscriptionService.UpdateChannelConfig(cfg); err != nil {
		http.Error(w, `{"error": "Failed to save config"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) HandleChannelFeedCategories(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		ChannelID         string   `json:"channel_id"`
		AllowedCategories []string `json:"allowed_categories"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	cfg, err := h.subscriptionService.GetChannelConfig(payload.ChannelID)
	if err != nil {
		http.Error(w, `{"error": "Failed to load config"}`, http.StatusInternalServerError)
		return
	}
	cfg.ChannelID = payload.ChannelID
	cfg.AllowedCategories = payload.AllowedCategories
	cfg.LastUpdateTimestamp = time.Now()
	if err := h.subscriptionService.UpdateChannelConfig(cfg); err != nil {
		http.Error(w, `{"error": "Failed to save config"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) HandleChannelFeedFrequency(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		ChannelID string `json:"channel_id"`
		Frequency string `json:"frequency"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	cfg, err := h.subscriptionService.GetChannelConfig(payload.ChannelID)
	if err != nil {
		http.Error(w, `{"error": "Failed to load config"}`, http.StatusInternalServerError)
		return
	}
	cfg.ChannelID = payload.ChannelID
	cfg.FrequencyMode = payload.Frequency
	cfg.LastUpdateTimestamp = time.Now()
	if err := h.subscriptionService.UpdateChannelConfig(cfg); err != nil {
		http.Error(w, `{"error": "Failed to save config"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) HandleGetChannelSettings(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, `{"error": "channel_id required"}`, http.StatusBadRequest)
		return
	}
	channelID := parts[len(parts)-1]
	cfg, err := h.subscriptionService.GetChannelConfig(channelID)
	if err != nil {
		http.Error(w, `{"error": "Failed to load config"}`, http.StatusInternalServerError)
		return
	}
	b, _ := json.Marshal(cfg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *WebhookHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	resp := map[string]interface{}{"status": "ok", "time": time.Now().UTC()}
	b, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (h *WebhookHandler) HandleNotificationsDM(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	if h.discordSession == nil {
		http.Error(w, `{"error": "Discord not ready"}`, http.StatusServiceUnavailable)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}
	var payload struct {
		DiscordUserID string                 `json:"discord_user_id"`
		Type          string                 `json:"type"`
		Data          map[string]interface{} `json:"payload"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	var msg string
	switch payload.Type {
	case "market_update":
		m := models.Market{ID: toString(payload.Data["market_id"]), Title: toString(payload.Data["title"]), Link: toString(payload.Data["link"])}
		if v, ok := payload.Data["volume"].(float64); ok {
			m.Volume = v
		}
		msg = h.marketService.CreateMarketUpdateMessage(&m)
	case "trading_start":
		m := models.Market{ID: toString(payload.Data["market_id"]), Title: toString(payload.Data["title"]), Description: toString(payload.Data["description"]), Link: toString(payload.Data["link"])}
		msg = h.marketService.CreateTradingStartMessage(&m)
	case "trading_end":
		m := models.Market{ID: toString(payload.Data["market_id"]), Title: toString(payload.Data["title"]), Link: toString(payload.Data["link"])}
		msg = h.marketService.CreateTradingEndMessage(&m)
	case "market_resolved":
		m := models.Market{ID: toString(payload.Data["market_id"]), Title: toString(payload.Data["title"]), ResolvedOutcome: toString(payload.Data["winning_outcome"]), Link: toString(payload.Data["link"])}
		msg = h.marketService.CreateMarketResolutionMessage(&m)
	case "market_buy":
		amt := toFloat(payload.Data["amount"])
		msg = h.marketService.CreateMarketBuyMessage(toString(payload.Data["market_id"]), toString(payload.Data["title"]), amt, toString(payload.Data["outcome"]), toString(payload.Data["buyer"]), toString(payload.Data["link"]))
	default:
		msg = ""
	}
	if msg == "" {
		http.Error(w, `{"error": "Unsupported type"}`, http.StatusBadRequest)
		return
	}
	ch, err := h.discordSession.UserChannelCreate(payload.DiscordUserID)
	if err != nil {
		http.Error(w, `{"error": "Failed to create DM channel"}`, http.StatusInternalServerError)
		return
	}
	if _, err := h.discordSession.ChannelMessageSend(ch.ID, msg); err != nil {
		http.Error(w, `{"error": "Failed to send DM"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"accepted": true}`))
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toFloat(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

// HandleUnregisterWebhook handles DELETE /discord/webhooks/unregister
func (h *WebhookHandler) HandleUnregisterWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		// accept POST too to make testing simpler (some clients can't send DELETE easily)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to read request body: %v", err))
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to parse JSON: %v", err))
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if payload.ID == "" {
		http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.subscriptionService.UnregisterWebhook(payload.ID); err != nil {
		h.logger.Error(fmt.Sprintf("Failed to unregister webhook: %v", err))
		http.Error(w, `{"error": "Failed to unregister webhook"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok": true}`))
}

// HandleListWebhooks handles GET /discord/webhooks
func (h *WebhookHandler) HandleListWebhooks(w http.ResponseWriter, r *http.Request) {
	if !h.AuthOk(r) {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	regs, err := h.subscriptionService.ListWebhookRegistrations()
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to list webhook registrations: %v", err))
		http.Error(w, `{"error": "Failed to list"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if b, err := json.Marshal(regs); err == nil {
		w.Write(b)
	} else {
		w.Write([]byte("[]"))
	}
}

// StartWebServer starts the webhook server
func (h *WebhookHandler) StartWebServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhooks/new_market", h.HandleNewMarket)
	mux.HandleFunc("/webhooks/market_update", h.HandleMarketUpdate)
	mux.HandleFunc("/webhooks/trading_started", h.HandleTradingStarted)
	mux.HandleFunc("/webhooks/trading_ended", h.HandleTradingEnded)
	mux.HandleFunc("/webhooks/market_resolved", h.HandleMarketResolved)

	// Admin endpoints for managing Discord webhook registrations
	mux.HandleFunc("/discord/webhooks/register", h.HandleRegisterWebhook)
	mux.HandleFunc("/discord/webhooks/unregister", h.HandleUnregisterWebhook)
	mux.HandleFunc("/discord/webhooks", h.HandleListWebhooks)

	mux.HandleFunc("/discord/webhooks/", h.HandleUnregisterWebhookByPath)

	mux.HandleFunc("/discord/events/new-market", h.HandleEventNewMarket)
	mux.HandleFunc("/discord/events/market-update", h.HandleEventMarketUpdate)
	mux.HandleFunc("/discord/events/trading-start", h.HandleEventTradingStart)
	mux.HandleFunc("/discord/events/trading-end", h.HandleEventTradingEnd)
	mux.HandleFunc("/discord/events/market-resolved", h.HandleEventMarketResolved)
	mux.HandleFunc("/discord/events/market-buy", h.HandleEventMarketBuy)

	mux.HandleFunc("/discord/notifications/dm", h.HandleNotificationsDM)

	mux.HandleFunc("/discord/subscribe/market", h.HandleSubscribeMarket)
	mux.HandleFunc("/discord/unsubscribe/market", h.HandleUnsubscribeMarket)
	mux.HandleFunc("/discord/subscribe/creator", h.HandleSubscribeCreator)
	mux.HandleFunc("/discord/unsubscribe/creator", h.HandleUnsubscribeCreator)
	mux.HandleFunc("/discord/subscriptions/", h.HandleGetUserSubscriptions)

	mux.HandleFunc("/discord/channel/feed/new_markets", h.HandleChannelFeedNewMarkets)
	mux.HandleFunc("/discord/channel/feed/categories", h.HandleChannelFeedCategories)
	mux.HandleFunc("/discord/channel/feed/frequency", h.HandleChannelFeedFrequency)
	mux.HandleFunc("/discord/channel/settings/", h.HandleGetChannelSettings)

	mux.HandleFunc("/discord/health", h.HandleHealth)

	h.logger.Info(fmt.Sprintf("Starting webhook server on port %s", port))
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to start webhook server: %v", err))
	}
}
