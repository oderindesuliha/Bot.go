package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"coral-bot/discord_bot/internal/models"
	"coral-bot/discord_bot/internal/utils"
)

// MarketService defines the interface for market-related operations
type MarketService interface {
	FetchMarket(marketID string) (*models.Market, error)
	FetchAllMarkets() ([]*models.Market, error)
	CreateMarketAnnouncement(market *models.Market) string
	CreateMarketUpdateMessage(market *models.Market) string
	CreateTradingStartMessage(market *models.Market) string
	CreateTradingEndMessage(market *models.Market) string
	CreateMarketResolutionMessage(market *models.Market) string
	CreateMarketBuyMessage(marketID string, title string, amount float64, outcome string, buyer string, link string) string
	ShouldSendUpdate(market *models.Market, frequency string, lastUpdate time.Time) bool
}

// MarketServiceImpl implements MarketService
type MarketServiceImpl struct {
	baseURL string
	logger  *utils.Logger
	client  *http.Client
}

// NewMarketService creates a new market service
func NewMarketService(baseURL string, logger *utils.Logger) *MarketServiceImpl {
	return &MarketServiceImpl{
		baseURL: baseURL,
		logger:  logger,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// FetchMarket fetches a market by ID from the backend API
func (service *MarketServiceImpl) FetchMarket(marketID string) (*models.Market, error) {
	if service.baseURL == "" {
		service.logger.Warning("Backend URL not configured, returning mock market data")
		// Return mock data for testing
		return &models.Market{
			ID:          marketID,
			Title:       "Test Market",
			Description: "This is a test market",
			Outcomes:    []string{"Yes", "No"},
			Percentages: []float64{50.0, 50.0},
			Category:    "Test",
			Creator:     "Test Creator",
			Volume:      1000.0,
			StartTime:   time.Now(),
			EndTime:     time.Now().Add(24 * time.Hour),
			Status:      "active",
			Link:        "https://coral.markets/market/" + marketID,
		}, nil
	}

	url := fmt.Sprintf("%s/markets/%s", service.baseURL, marketID)
	resp, err := service.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	var market models.Market
	if err := json.NewDecoder(resp.Body).Decode(&market); err != nil {
		return nil, fmt.Errorf("failed to decode market response: %w", err)
	}

	return &market, nil
}

// FetchAllMarkets fetches all markets from the backend API
func (service *MarketServiceImpl) FetchAllMarkets() ([]*models.Market, error) {
	if service.baseURL == "" {
		service.logger.Warning("Backend URL not configured, returning mock markets")
		// Return mock data for testing
		return []*models.Market{
			{
				ID:          "1",
				Title:       "Test Market 1",
				Description: "This is a test market",
				Outcomes:    []string{"Yes", "No"},
				Percentages: []float64{50.0, 50.0},
				Category:    "Test",
				Creator:     "Test Creator",
				Volume:      1000.0,
				StartTime:   time.Now(),
				EndTime:     time.Now().Add(24 * time.Hour),
				Status:      "active",
				Link:        "https://coral.markets/market/1",
			},
			{
				ID:          "2",
				Title:       "Test Market 2",
				Description: "This is another test market",
				Outcomes:    []string{"Option A", "Option B", "Option C"},
				Percentages: []float64{33.3, 33.3, 33.4},
				Category:    "Test",
				Creator:     "Another Creator",
				Volume:      2500.0,
				StartTime:   time.Now().Add(-1 * time.Hour),
				EndTime:     time.Now().Add(12 * time.Hour),
				Status:      "active",
				Link:        "https://coral.markets/market/2",
			},
		}, nil
	}

	url := fmt.Sprintf("%s/markets", service.baseURL)
	resp, err := service.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch markets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend returned status %d", resp.StatusCode)
	}

	var markets []*models.Market
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return nil, fmt.Errorf("failed to decode markets response: %w", err)
	}

	return markets, nil
}

// CreateMarketAnnouncement creates a formatted announcement message for a new market
func (service *MarketServiceImpl) CreateMarketAnnouncement(market *models.Market) string {
	message := fmt.Sprintf(
		"🎉 **NEW MARKET ALERT** 🎉\n\n"+
			"**%s**\n"+
			"%s\n\n"+
			"📊 Volume: $%.2f\n"+
			"⏰ Time Left: %s\n\n"+
			"**Outcomes:**\n",
		market.Title,
		market.Description,
		market.Volume,
		time.Until(market.EndTime).String(),
	)

	for i, outcome := range market.Outcomes {
		percentage := 0.0
		if i < len(market.Percentages) {
			percentage = market.Percentages[i]
		}
		message += fmt.Sprintf("- %s (%.1f%%)\n", outcome, percentage)
	}

	message += fmt.Sprintf("\n🔗 [View on Coral Markets](%s)", market.Link)
	return message
}

// CreateMarketUpdateMessage creates a formatted update message for a market
func (service *MarketServiceImpl) CreateMarketUpdateMessage(market *models.Market) string {
	message := fmt.Sprintf(
		"📈 **MARKET UPDATE** 📈\n\n"+
			"**%s**\n\n"+
			"📊 Volume: $%.2f\n"+
			"⏰ Time Left: %s\n\n"+
			"**Current Probabilities:**\n",
		market.Title,
		market.Volume,
		time.Until(market.EndTime).String(),
	)

	for i, outcome := range market.Outcomes {
		percentage := 0.0
		if i < len(market.Percentages) {
			percentage = market.Percentages[i]
		}
		message += fmt.Sprintf("- %s (%.1f%%)\n", outcome, percentage)
	}

	message += fmt.Sprintf("\n🔗 [View on Coral Markets](%s)", market.Link)
	return message
}

// CreateTradingStartMessage creates a message for when trading starts
func (service *MarketServiceImpl) CreateTradingStartMessage(market *models.Market) string {
	return fmt.Sprintf(
		"🟢 **TRADING STARTED** 🟢\n\n"+
			"**%s**\n\n"+
			"Trading is now open! Place your bets.\n\n"+
			"🔗 [View on Coral Markets](%s)",
		market.Title,
		market.Link,
	)
}

// CreateTradingEndMessage creates a message for when trading ends
func (service *MarketServiceImpl) CreateTradingEndMessage(market *models.Market) string {
	return fmt.Sprintf(
		"🔴 **TRADING CLOSED** 🔴\n\n"+
			"**%s**\n\n"+
			"Betting is now closed. Market will resolve soon.\n\n"+
			"🔗 [View on Coral Markets](%s)",
		market.Title,
		market.Link,
	)
}

// CreateMarketResolutionMessage creates a message for when a market is resolved
func (service *MarketServiceImpl) CreateMarketResolutionMessage(market *models.Market) string {
	resolution := "Market resolved"
	if market.ResolvedOutcome != "" {
		resolution = fmt.Sprintf("Resolved: **%s**", market.ResolvedOutcome)
	}

	return fmt.Sprintf(
		"✅ **MARKET RESOLVED** ✅\n\n"+
			"**%s**\n\n"+
			"%s\n\n"+
			"🔗 [View on Coral Markets](%s)",
		market.Title,
		resolution,
		market.Link,
	)
}

func (s *MarketServiceImpl) CreateMarketBuyMessage(marketID string, title string, amount float64, outcome string, buyer string, link string) string {
	buyerText := buyer
	if buyerText == "" {
		buyerText = "Anonymous"
	}
	return fmt.Sprintf(
		"💸 **MARKET BUY** 💸\n\n"+
			"**%s**\n\n"+
			"Buyer: %s\n"+
			"Amount: $%.2f\n"+
			"Outcome: %s\n\n"+
			"🔗 [View on Coral Markets](%s)",
		title,
		buyerText,
		amount,
		outcome,
		link,
	)
}

// ShouldSendUpdate determines if an update should be sent based on frequency settings
func (service *MarketServiceImpl) ShouldSendUpdate(market *models.Market, frequency string, lastUpdate time.Time) bool {
	if market.Status != "active" {
		return false
	}

	timeSinceLastUpdate := time.Since(lastUpdate)
	timeLeft := time.Until(market.EndTime)

	// For markets closing soon (less than 6 hours), increase update frequency
	if timeLeft < 6*time.Hour {
		return timeSinceLastUpdate >= 15*time.Minute
	}

	switch frequency {
	case "high":
		// Send updates every 30 minutes for active markets
		return timeSinceLastUpdate >= 30*time.Minute
	case "medium":
		// Send updates every hour for active markets
		return timeSinceLastUpdate >= 1*time.Hour
	case "low":
		// Send updates every 3 hours for active markets
		return timeSinceLastUpdate >= 3*time.Hour
	default:
		// Medium frequency for unknown values
		return timeSinceLastUpdate >= 1*time.Hour
	}
}
