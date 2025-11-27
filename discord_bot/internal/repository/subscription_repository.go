package repository

import (
	"sync"

	"coral-bot/discord_bot/internal/models"
)

// SubscriptionRepository defines the interface for subscription data operations
type SubscriptionRepository interface {
	GetSubscription(discordUserID string) (*models.Subscription, error)
	SaveSubscription(subscription *models.Subscription) error
	DeleteSubscription(discordUserID string) error
	GetAllSubscriptions() ([]*models.Subscription, error)

	GetChannelConfig(channelID string) (*models.ChannelConfig, error)
	SaveChannelConfig(config *models.ChannelConfig) error
	GetAllChannelConfigs() ([]*models.ChannelConfig, error)

	// Webhook registration methods
	SaveWebhookRegistration(registration *models.WebhookRegistration) error
	GetWebhookRegistration(id string) (*models.WebhookRegistration, error)
	DeleteWebhookRegistration(id string) error
	GetAllWebhookRegistrations() ([]*models.WebhookRegistration, error)
	GetWebhookRegistrationsByChannel(channelID string) ([]*models.WebhookRegistration, error)
}

// InMemorySubscriptionRepository implements SubscriptionRepository using in-memory storage
type InMemorySubscriptionRepository struct {
    subscriptions map[string]*models.Subscription
    channels      map[string]*models.ChannelConfig
    webhooks      map[string]*models.WebhookRegistration
    mutex         sync.RWMutex
}

// NewInMemorySubscriptionRepository creates a new in-memory subscription repository
func NewInMemorySubscriptionRepository() *InMemorySubscriptionRepository {
	return &InMemorySubscriptionRepository{
		subscriptions: make(map[string]*models.Subscription),
		channels:      make(map[string]*models.ChannelConfig),
		webhooks:      make(map[string]*models.WebhookRegistration),
	}
}

// GetSubscription retrieves a subscription by Discord user ID
func (repo *InMemorySubscriptionRepository) GetSubscription(discordUserID string) (*models.Subscription, error) {
    repo.mutex.RLock()
    defer repo.mutex.RUnlock()

    subscription, exists := repo.subscriptions[discordUserID]
	if !exists {
		// Return empty subscription if not found
		return &models.Subscription{
			DiscordUserID:      discordUserID,
			SubscribedMarkets:  []string{},
			SubscribedCreators: []string{},
		}, nil
	}

    return subscription, nil
}

// SaveSubscription saves a subscription
func (repo *InMemorySubscriptionRepository) SaveSubscription(subscription *models.Subscription) error {
    repo.mutex.Lock()
    defer repo.mutex.Unlock()

    repo.subscriptions[subscription.DiscordUserID] = subscription
    return nil
}

// DeleteSubscription deletes a subscription
func (repo *InMemorySubscriptionRepository) DeleteSubscription(discordUserID string) error {
    repo.mutex.Lock()
    defer repo.mutex.Unlock()

    delete(repo.subscriptions, discordUserID)
    return nil
}

// GetAllSubscriptions retrieves all subscriptions
func (repo *InMemorySubscriptionRepository) GetAllSubscriptions() ([]*models.Subscription, error) {
    repo.mutex.RLock()
    defer repo.mutex.RUnlock()

    subscriptions := make([]*models.Subscription, 0, len(repo.subscriptions))
    for _, subscription := range repo.subscriptions {
        subscriptions = append(subscriptions, subscription)
    }

	return subscriptions, nil
}

// GetChannelConfig retrieves a channel configuration by channel ID
func (repo *InMemorySubscriptionRepository) GetChannelConfig(channelID string) (*models.ChannelConfig, error) {
    repo.mutex.RLock()
    defer repo.mutex.RUnlock()

    config, exists := repo.channels[channelID]
	if !exists {
		// Return default config if not found
		return &models.ChannelConfig{
			ChannelID:           channelID,
			FeedEnabled:         true,
			AllowedCategories:   []string{},
			FrequencyMode:       "medium",
			LastUpdateTimestamp: models.ChannelConfig{}.LastUpdateTimestamp,
		}, nil
	}

    return config, nil
}

// SaveChannelConfig saves a channel configuration
func (repo *InMemorySubscriptionRepository) SaveChannelConfig(config *models.ChannelConfig) error {
    repo.mutex.Lock()
    defer repo.mutex.Unlock()

    repo.channels[config.ChannelID] = config
    return nil
}

// GetAllChannelConfigs retrieves all channel configurations
func (repo *InMemorySubscriptionRepository) GetAllChannelConfigs() ([]*models.ChannelConfig, error) {
    repo.mutex.RLock()
    defer repo.mutex.RUnlock()

    configs := make([]*models.ChannelConfig, 0, len(repo.channels))
    for _, config := range repo.channels {
        configs = append(configs, config)
    }

	return configs, nil
}

// SaveWebhookRegistration stores or updates a webhook registration
func (repo *InMemorySubscriptionRepository) SaveWebhookRegistration(registration *models.WebhookRegistration) error {
    repo.mutex.Lock()
    defer repo.mutex.Unlock()

    repo.webhooks[registration.ID] = registration
    return nil
}

// GetWebhookRegistration retrieves a webhook registration by id
func (repo *InMemorySubscriptionRepository) GetWebhookRegistration(id string) (*models.WebhookRegistration, error) {
    repo.mutex.RLock()
    defer repo.mutex.RUnlock()

    reg, exists := repo.webhooks[id]
	if !exists {
		return nil, nil
	}
	return reg, nil
}

// DeleteWebhookRegistration deletes a webhook registration by id
func (repo *InMemorySubscriptionRepository) DeleteWebhookRegistration(id string) error {
    repo.mutex.Lock()
    defer repo.mutex.Unlock()

    delete(repo.webhooks, id)
    return nil
}

// GetAllWebhookRegistrations returns all webhook registrations
func (repo *InMemorySubscriptionRepository) GetAllWebhookRegistrations() ([]*models.WebhookRegistration, error) {
    repo.mutex.RLock()
    defer repo.mutex.RUnlock()

    regs := make([]*models.WebhookRegistration, 0, len(repo.webhooks))
    for _, reg := range repo.webhooks {
        regs = append(regs, reg)
    }

	return regs, nil
}

// GetWebhookRegistrationsByChannel returns registrations for a specific channel
func (repo *InMemorySubscriptionRepository) GetWebhookRegistrationsByChannel(channelID string) ([]*models.WebhookRegistration, error) {
    repo.mutex.RLock()
    defer repo.mutex.RUnlock()

    regs := []*models.WebhookRegistration{}
    for _, reg := range repo.webhooks {
        if reg.ChannelID == channelID {
            regs = append(regs, reg)
        }
    }
    return regs, nil
}
