package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "coral-bot/discord_bot/internal/models"
    "coral-bot/discord_bot/internal/repository"
    "coral-bot/discord_bot/internal/services"
    "coral-bot/discord_bot/internal/utils"
    "coral-bot/discord_bot/internal/web"
)

func TestWebhookHandler_AuthOk(t *testing.T) {
    logger := utils.NewLogger()
    repo := repository.NewInMemorySubscriptionRepository()
    marketService := services.NewMarketService("", logger)
    subscriptionService := services.NewSubscriptionService(repo, logger)
    handler := web.NewWebhookHandler(marketService, subscriptionService, logger)

    req := httptest.NewRequest("POST", "/webhooks/new_market", nil)
    if !handler.AuthOk(req) {
        t.Error("Expected AuthOk to return true when no auth is required")
    }

    t.Setenv("CORAL_API_KEY", "test-key")
    req = httptest.NewRequest("POST", "/webhooks/new_market", nil)
    req.Header.Set("X-API-Key", "test-key")
    if !handler.AuthOk(req) {
        t.Error("Expected AuthOk to return true when valid API key is provided")
    }

    req.Header.Set("X-API-Key", "invalid-key")
    if handler.AuthOk(req) {
        t.Error("Expected AuthOk to return false when invalid API key is provided")
    }

    t.Setenv("CORAL_API_KEY", "")
    t.Setenv("CORAL_TOKEN", "test-token")
    req = httptest.NewRequest("POST", "/webhooks/new_market", nil)
    req.Header.Set("Authorization", "Bearer test-token")
    if !handler.AuthOk(req) {
        t.Error("Expected AuthOk to return true when valid bearer token is provided")
    }

    req.Header.Set("Authorization", "Bearer invalid-token")
    if handler.AuthOk(req) {
        t.Error("Expected AuthOk to return false when invalid bearer token is provided")
    }
}

func TestWebhookHandler_HandleNewMarket(t *testing.T) {
    logger := utils.NewLogger()
    repo := repository.NewInMemorySubscriptionRepository()
    marketService := services.NewMarketService("", logger)
    subscriptionService := services.NewSubscriptionService(repo, logger)
    handler := web.NewWebhookHandler(marketService, subscriptionService, logger)

    market := models.Market{
        ID:          "test-market",
        Title:       "Test Market",
        Description: "A test market",
        Outcomes:    []string{"Yes", "No"},
        Percentages: []float64{50.0, 50.0},
        Category:    "Test",
        Creator:     "Test Creator",
        Volume:      1000.0,
        Status:      "active",
        Link:        "https://example.com/markets/test-market",
    }

    payload := struct {
        EventType string        `json:"event_type"`
        Market    models.Market `json:"market"`
    }{
        EventType: "new_market",
        Market:    market,
    }

    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        t.Fatalf("Failed to marshal payload: %v", err)
    }

    req := httptest.NewRequest("POST", "/webhooks/new_market", bytes.NewBuffer(payloadBytes))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    handler.HandleNewMarket(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    expected := `{"ok": true}`
    if rr.Body.String() != expected {
        t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }
}

