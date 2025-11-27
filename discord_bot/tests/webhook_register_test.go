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

func TestRegisterUnregisterListWebhooks(t *testing.T) {
    logger := utils.NewLogger()
    repo := repository.NewInMemorySubscriptionRepository()
    marketService := services.NewMarketService("", logger)
    subscriptionService := services.NewSubscriptionService(repo, logger)
    handler := web.NewWebhookHandler(marketService, subscriptionService, logger)

    payload := map[string]interface{}{
        "channel_id":  "channel-1",
        "webhook_url": "https://discordapp.test/webhook/1",
        "events":      []string{"new_market"},
        "frequency":   "low",
    }
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/discord/webhooks/register", bytes.NewBuffer(body))
    rr := httptest.NewRecorder()
    handler.HandleRegisterWebhook(rr, req)

    if rr.Code != http.StatusCreated {
        t.Fatalf("expected status %d got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
    }

    var created models.WebhookRegistration
    if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
    if created.ID == "" {
        t.Fatalf("expected created id, got empty")
    }

    listReq := httptest.NewRequest(http.MethodGet, "/discord/webhooks", nil)
    listRec := httptest.NewRecorder()
    handler.HandleListWebhooks(listRec, listReq)
    if listRec.Code != http.StatusOK {
        t.Fatalf("expected status %d got %d for list", http.StatusOK, listRec.Code)
    }

    var all []models.WebhookRegistration
    if err := json.Unmarshal(listRec.Body.Bytes(), &all); err != nil {
        t.Fatalf("failed to decode list response: %v", err)
    }
    if len(all) == 0 {
        t.Fatalf("expected at least one registration in list")
    }

    delPayload := map[string]string{"id": created.ID}
    delBody, _ := json.Marshal(delPayload)
    delReq := httptest.NewRequest(http.MethodDelete, "/discord/webhooks/unregister", bytes.NewBuffer(delBody))
    delRec := httptest.NewRecorder()
    handler.HandleUnregisterWebhook(delRec, delReq)

    if delRec.Code != http.StatusOK {
        t.Fatalf("expected status %d got %d for delete", http.StatusOK, delRec.Code)
    }

    listRec2 := httptest.NewRecorder()
    handler.HandleListWebhooks(listRec2, httptest.NewRequest(http.MethodGet, "/discord/webhooks", nil))
    var all2 []models.WebhookRegistration
    if err := json.Unmarshal(listRec2.Body.Bytes(), &all2); err != nil {
        t.Fatalf("failed to decode list response after delete: %v", err)
    }

    for _, r := range all2 {
        if r.ID == created.ID {
            t.Fatalf("expected registration to be deleted, but still present")
        }
    }
}

