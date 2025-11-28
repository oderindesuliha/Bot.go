package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "coral-bot/discord_bot/internal/repository"
    "coral-bot/discord_bot/internal/services"
    "coral-bot/discord_bot/internal/utils"
    "coral-bot/discord_bot/internal/web"
)

func setupHandler() *web.WebhookHandler {
    logger := utils.NewLogger()
    repo := repository.NewInMemorySubscriptionRepository()
    marketService := services.NewMarketService("", logger)
    subscriptionService := services.NewSubscriptionService(repo, logger)
    return web.NewWebhookHandler(marketService, subscriptionService, logger)
}

func TestEventNewMarketAccepted(t *testing.T) {
    h := setupHandler()
    payload := map[string]interface{}{
        "market_id":  "m1",
        "title":      "Test Market",
        "description": "Desc",
        "creator":    "u1",
        "category":   "politics",
        "outcomes":   []map[string]string{{"id": "o1", "name": "Yes"}, {"id": "o2", "name": "No"}},
        "start_time": time.Now().Add(-time.Hour).Format(time.RFC3339),
        "end_time":   time.Now().Add(time.Hour).Format(time.RFC3339),
        "volume":     100.0,
        "link":       "https://example.com/m1",
    }
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/discord/events/new-market", bytes.NewBuffer(b))
    rec := httptest.NewRecorder()
    h.HandleEventNewMarket(rec, req)
    if rec.Code != http.StatusAccepted {
        t.Fatalf("expected %d got %d", http.StatusAccepted, rec.Code)
    }
}

func TestEventMarketUpdateAccepted(t *testing.T) {
    h := setupHandler()
    payload := map[string]interface{}{
        "market_id":       "m1",
        "title":           "Test Market",
        "volume":          200.0,
        "volume_delta_pct": 5.0,
        "time_left":       "1h",
        "end_time":        time.Now().Add(time.Hour).Format(time.RFC3339),
        "link":            "https://example.com/m1",
    }
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/discord/events/market-update", bytes.NewBuffer(b))
    rec := httptest.NewRecorder()
    h.HandleEventMarketUpdate(rec, req)
    if rec.Code != http.StatusAccepted {
        t.Fatalf("expected %d got %d", http.StatusAccepted, rec.Code)
    }
}

func TestEventTradingStartAccepted(t *testing.T) {
    h := setupHandler()
    payload := map[string]interface{}{
        "market_id":    "m1",
        "title":        "Test Market",
        "description":  "Desc",
        "duration":     "2h",
        "outcomes_count": 2,
        "outcomes":     []string{"Yes", "No"},
        "link":         "https://example.com/m1",
    }
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/discord/events/trading-start", bytes.NewBuffer(b))
    rec := httptest.NewRecorder()
    h.HandleEventTradingStart(rec, req)
    if rec.Code != http.StatusAccepted {
        t.Fatalf("expected %d got %d", http.StatusAccepted, rec.Code)
    }
}

func TestEventTradingEndAccepted(t *testing.T) {
    h := setupHandler()
    payload := map[string]interface{}{
        "market_id":   "m1",
        "title":       "Test Market",
        "description": "Desc",
        "outcomes":    []map[string]interface{}{{"id": "o1", "name": "Yes", "pct": 60.0}, {"id": "o2", "name": "No", "pct": 40.0}},
        "final_pool":  500.0,
        "link":        "https://example.com/m1",
    }
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/discord/events/trading-end", bytes.NewBuffer(b))
    rec := httptest.NewRecorder()
    h.HandleEventTradingEnd(rec, req)
    if rec.Code != http.StatusAccepted {
        t.Fatalf("expected %d got %d", http.StatusAccepted, rec.Code)
    }
}

func TestEventMarketResolvedAccepted(t *testing.T) {
    h := setupHandler()
    payload := map[string]interface{}{
        "market_id":      "m1",
        "title":          "Test Market",
        "winning_outcome": "Yes",
        "total_pool":     1000.0,
        "link":           "https://example.com/m1",
    }
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/discord/events/market-resolved", bytes.NewBuffer(b))
    rec := httptest.NewRecorder()
    h.HandleEventMarketResolved(rec, req)
    if rec.Code != http.StatusAccepted {
        t.Fatalf("expected %d got %d", http.StatusAccepted, rec.Code)
    }
}

func TestEventMarketBuyAccepted(t *testing.T) {
    h := setupHandler()
    payload := map[string]interface{}{
        "market_id": "m1",
        "title":     "Test Market",
        "amount":    42.50,
        "outcome":   "Yes",
        "buyer":     "u1",
        "link":      "https://example.com/m1",
    }
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/discord/events/market-buy", bytes.NewBuffer(b))
    rec := httptest.NewRecorder()
    h.HandleEventMarketBuy(rec, req)
    if rec.Code != http.StatusAccepted {
        t.Fatalf("expected %d got %d", http.StatusAccepted, rec.Code)
    }
}

func TestNotificationsDMRequiresDiscordSession(t *testing.T) {
    h := setupHandler()
    payload := map[string]interface{}{
        "discord_user_id": "user-1",
        "type":            "market_update",
        "payload":         map[string]interface{}{"market_id": "m1", "title": "Test", "link": "https://example.com/m1"},
    }
    b, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/discord/notifications/dm", bytes.NewBuffer(b))
    rec := httptest.NewRecorder()
    h.HandleNotificationsDM(rec, req)
    if rec.Code != http.StatusServiceUnavailable {
        t.Fatalf("expected %d got %d", http.StatusServiceUnavailable, rec.Code)
    }
}

func TestSubscribeUnsubscribeMarket(t *testing.T) {
    h := setupHandler()
    sub := map[string]string{"discord_user_id": "u1", "market_id": "m1"}
    b1, _ := json.Marshal(sub)
    req1 := httptest.NewRequest(http.MethodPost, "/discord/subscribe/market", bytes.NewBuffer(b1))
    rec1 := httptest.NewRecorder()
    h.HandleSubscribeMarket(rec1, req1)
    if rec1.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, rec1.Code) }

    b2, _ := json.Marshal(sub)
    req2 := httptest.NewRequest(http.MethodPost, "/discord/unsubscribe/market", bytes.NewBuffer(b2))
    rec2 := httptest.NewRecorder()
    h.HandleUnsubscribeMarket(rec2, req2)
    if rec2.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, rec2.Code) }
}

func TestSubscribeUnsubscribeCreatorAndList(t *testing.T) {
    h := setupHandler()
    sub := map[string]string{"discord_user_id": "u1", "creator_id": "c1"}
    b1, _ := json.Marshal(sub)
    req1 := httptest.NewRequest(http.MethodPost, "/discord/subscribe/creator", bytes.NewBuffer(b1))
    rec1 := httptest.NewRecorder()
    h.HandleSubscribeCreator(rec1, req1)
    if rec1.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, rec1.Code) }

    listReq := httptest.NewRequest(http.MethodGet, "/discord/subscriptions/u1", nil)
    listRec := httptest.NewRecorder()
    h.HandleGetUserSubscriptions(listRec, listReq)
    if listRec.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, listRec.Code) }

    b2, _ := json.Marshal(sub)
    req2 := httptest.NewRequest(http.MethodPost, "/discord/unsubscribe/creator", bytes.NewBuffer(b2))
    rec2 := httptest.NewRecorder()
    h.HandleUnsubscribeCreator(rec2, req2)
    if rec2.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, rec2.Code) }
}

func TestChannelAdminSettings(t *testing.T) {
    h := setupHandler()
    b1, _ := json.Marshal(map[string]interface{}{"channel_id": "ch1", "enabled": true})
    r1 := httptest.NewRequest(http.MethodPost, "/discord/channel/feed/new_markets", bytes.NewBuffer(b1))
    w1 := httptest.NewRecorder()
    h.HandleChannelFeedNewMarkets(w1, r1)
    if w1.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, w1.Code) }

    b2, _ := json.Marshal(map[string]interface{}{"channel_id": "ch1", "allowed_categories": []string{"politics"}})
    r2 := httptest.NewRequest(http.MethodPost, "/discord/channel/feed/categories", bytes.NewBuffer(b2))
    w2 := httptest.NewRecorder()
    h.HandleChannelFeedCategories(w2, r2)
    if w2.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, w2.Code) }

    b3, _ := json.Marshal(map[string]interface{}{"channel_id": "ch1", "frequency": "high"})
    r3 := httptest.NewRequest(http.MethodPost, "/discord/channel/feed/frequency", bytes.NewBuffer(b3))
    w3 := httptest.NewRecorder()
    h.HandleChannelFeedFrequency(w3, r3)
    if w3.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, w3.Code) }

    r4 := httptest.NewRequest(http.MethodGet, "/discord/channel/settings/ch1", nil)
    w4 := httptest.NewRecorder()
    h.HandleGetChannelSettings(w4, r4)
    if w4.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, w4.Code) }
}

func TestHealthEndpointOK(t *testing.T) {
    h := setupHandler()
    req := httptest.NewRequest(http.MethodGet, "/discord/health", nil)
    rec := httptest.NewRecorder()
    h.HandleHealth(rec, req)
    if rec.Code != http.StatusOK { t.Fatalf("expected %d got %d", http.StatusOK, rec.Code) }
}

func TestDeleteWebhookByPath204(t *testing.T) {
    h := setupHandler()
    regBody := map[string]interface{}{"channel_id": "channel-2", "webhook_url": "https://discordapp.test/webhook/2", "events": []string{"new_market"}, "frequency": "low"}
    rb, _ := json.Marshal(regBody)
    regReq := httptest.NewRequest(http.MethodPost, "/discord/webhooks/register", bytes.NewBuffer(rb))
    regRec := httptest.NewRecorder()
    h.HandleRegisterWebhook(regRec, regReq)
    if regRec.Code != http.StatusCreated { t.Fatalf("expected %d got %d", http.StatusCreated, regRec.Code) }

    var created struct{ ID string `json:"id"` }
    _ = json.Unmarshal(regRec.Body.Bytes(), &created)
    delReq := httptest.NewRequest(http.MethodDelete, "/discord/webhooks/"+created.ID, nil)
    delRec := httptest.NewRecorder()
    h.HandleUnregisterWebhookByPath(delRec, delReq)
    if delRec.Code != http.StatusNoContent { t.Fatalf("expected %d got %d", http.StatusNoContent, delRec.Code) }
}

