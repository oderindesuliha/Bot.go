package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"coral-bot/discord_bot/internal/models"
	"coral-bot/discord_bot/internal/repository"
	"coral-bot/discord_bot/internal/utils"
)

// SubscriptionService defines the interface for subscription-related operations
type SubscriptionService interface {
	SubscribeToMarket(discordUserID, marketID string) error
	UnsubscribeFromMarket(discordUserID, marketID string) error
	SubscribeToCreator(discordUserID, creator string) error
	UnsubscribeFromCreator(discordUserID, creator string) error
	GetUserSubscriptions(discordUserID string) (*models.Subscription, error)
	GetAllSubscriptions() ([]*models.Subscription, error)

	// Channel configuration
	UpdateChannelConfig(config *models.ChannelConfig) error
	GetChannelConfig(channelID string) (*models.ChannelConfig, error)
	GetAllChannelConfigs() ([]*models.ChannelConfig, error)

	// Notification logic
	ShouldNotifyUser(subscription *models.Subscription, market *models.Market) bool
	SendNotificationToUser(discordUserID string, message string) error

	// Webhook registration management
	RegisterWebhook(registration *models.WebhookRegistration) (*models.WebhookRegistration, error)
	UnregisterWebhook(id string) error
	GetWebhookRegistration(id string) (*models.WebhookRegistration, error)
	ListWebhookRegistrations() ([]*models.WebhookRegistration, error)
	ListWebhookRegistrationsByChannel(channelID string) ([]*models.WebhookRegistration, error)
}

// SubscriptionServiceImpl implements SubscriptionService
type SubscriptionServiceImpl struct {
    repo   repository.SubscriptionRepository
    logger *utils.Logger
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(repo repository.SubscriptionRepository, logger *utils.Logger) *SubscriptionServiceImpl {
    return &SubscriptionServiceImpl{
        repo:   repo,
        logger: logger,
    }
}

// SubscribeToMarket subscribes a user to a market
func (service *SubscriptionServiceImpl) SubscribeToMarket(discordUserID, marketID string) error {
    subscription, err := service.repo.GetSubscription(discordUserID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Check if already subscribed
	for _, id := range subscription.SubscribedMarkets {
		if id == marketID {
			return nil // Already subscribed
		}
	}

	// Add to subscribed markets
	subscription.SubscribedMarkets = append(subscription.SubscribedMarkets, marketID)

    return service.repo.SaveSubscription(subscription)
}

// UnsubscribeFromMarket unsubscribes a user from a market
func (service *SubscriptionServiceImpl) UnsubscribeFromMarket(discordUserID, marketID string) error {
    subscription, err := service.repo.GetSubscription(discordUserID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Remove from subscribed markets
	newMarkets := []string{}
	for _, id := range subscription.SubscribedMarkets {
		if id != marketID {
			newMarkets = append(newMarkets, id)
		}
	}

	subscription.SubscribedMarkets = newMarkets
    return service.repo.SaveSubscription(subscription)
}

// SubscribeToCreator subscribes a user to a creator
func (service *SubscriptionServiceImpl) SubscribeToCreator(discordUserID, creator string) error {
    subscription, err := service.repo.GetSubscription(discordUserID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Check if already subscribed
	for _, c := range subscription.SubscribedCreators {
		if c == creator {
			return nil // Already subscribed
		}
	}

	// Add to subscribed creators
	subscription.SubscribedCreators = append(subscription.SubscribedCreators, creator)

    return service.repo.SaveSubscription(subscription)
}

// UnsubscribeFromCreator unsubscribes a user from a creator
func (service *SubscriptionServiceImpl) UnsubscribeFromCreator(discordUserID, creator string) error {
    subscription, err := service.repo.GetSubscription(discordUserID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Remove from subscribed creators
	newCreators := []string{}
	for _, c := range subscription.SubscribedCreators {
		if c != creator {
			newCreators = append(newCreators, c)
		}
	}

	subscription.SubscribedCreators = newCreators
    return service.repo.SaveSubscription(subscription)
}

// GetUserSubscriptions gets a user's subscriptions
func (service *SubscriptionServiceImpl) GetUserSubscriptions(discordUserID string) (*models.Subscription, error) {
    return service.repo.GetSubscription(discordUserID)
}

// GetAllSubscriptions gets all subscriptions
func (service *SubscriptionServiceImpl) GetAllSubscriptions() ([]*models.Subscription, error) {
    return service.repo.GetAllSubscriptions()
}

// UpdateChannelConfig updates a channel's configuration
func (service *SubscriptionServiceImpl) UpdateChannelConfig(config *models.ChannelConfig) error {
    return service.repo.SaveChannelConfig(config)
}

// GetChannelConfig gets a channel's configuration
func (service *SubscriptionServiceImpl) GetChannelConfig(channelID string) (*models.ChannelConfig, error) {
    return service.repo.GetChannelConfig(channelID)
}

// GetAllChannelConfigs gets all channel configurations
func (service *SubscriptionServiceImpl) GetAllChannelConfigs() ([]*models.ChannelConfig, error) {
    return service.repo.GetAllChannelConfigs()
}

// ShouldNotifyUser determines if a user should be notified about a market
func (service *SubscriptionServiceImpl) ShouldNotifyUser(subscription *models.Subscription, market *models.Market) bool {
	// Check if user is subscribed to this market
    for _, marketID := range subscription.SubscribedMarkets {
		if marketID == market.ID {
			return true
		}
	}

	// Check if user is subscribed to this creator
    for _, creator := range subscription.SubscribedCreators {
		if creator == market.Creator {
			return true
		}
	}

	return false
}

// SendNotificationToUser sends a notification to a user (placeholder implementation)
func (service *SubscriptionServiceImpl) SendNotificationToUser(discordUserID string, message string) error {
    service.logger.Info(fmt.Sprintf("Would send DM to user %s: %s", discordUserID, message))
    return nil
}

// helper to create a secure random id
func generateID() (string, error) {
    randomBytes := make([]byte, 12)
    if _, err := rand.Read(randomBytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(randomBytes), nil
}

// RegisterWebhook registers a webhook and persists it
func (service *SubscriptionServiceImpl) RegisterWebhook(registration *models.WebhookRegistration) (*models.WebhookRegistration, error) {
	// generate a simple id and set createdAt
	// use time.Now().UnixNano() and fmt.Sprintf random hex
	// generate id and timestamp
	idBytes, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}
	registration.ID = "wh_" + idBytes
	registration.CreatedAt = time.Now()

	if registration.Frequency == "" {
		registration.Frequency = "medium"
	}

    if err := service.repo.SaveWebhookRegistration(registration); err != nil {
        return nil, fmt.Errorf("failed to save webhook registration: %w", err)
    }
    return registration, nil
}

// UnregisterWebhook removes a webhook registration
func (service *SubscriptionServiceImpl) UnregisterWebhook(id string) error {
    return service.repo.DeleteWebhookRegistration(id)
}

// GetWebhookRegistration returns a registration by id
func (service *SubscriptionServiceImpl) GetWebhookRegistration(id string) (*models.WebhookRegistration, error) {
    return service.repo.GetWebhookRegistration(id)
}

// ListWebhookRegistrations lists all registrations
func (service *SubscriptionServiceImpl) ListWebhookRegistrations() ([]*models.WebhookRegistration, error) {
    return service.repo.GetAllWebhookRegistrations()
}

// ListWebhookRegistrationsByChannel lists registrations for a channel
func (service *SubscriptionServiceImpl) ListWebhookRegistrationsByChannel(channelID string) ([]*models.WebhookRegistration, error) {
    return service.repo.GetWebhookRegistrationsByChannel(channelID)
}
