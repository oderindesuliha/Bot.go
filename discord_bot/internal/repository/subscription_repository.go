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
func (r *InMemorySubscriptionRepository) GetSubscription(discordUserID string) (*models.Subscription, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	subscription, exists := r.subscriptions[discordUserID]
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
func (r *InMemorySubscriptionRepository) SaveSubscription(subscription *models.Subscription) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.subscriptions[subscription.DiscordUserID] = subscription
	return nil
}

// DeleteSubscription deletes a subscription
func (r *InMemorySubscriptionRepository) DeleteSubscription(discordUserID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.subscriptions, discordUserID)
	return nil
}

// GetAllSubscriptions retrieves all subscriptions
func (r *InMemorySubscriptionRepository) GetAllSubscriptions() ([]*models.Subscription, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	subscriptions := make([]*models.Subscription, 0, len(r.subscriptions))
	for _, subscription := range r.subscriptions {
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

// GetChannelConfig retrieves a channel configuration by channel ID
func (r *InMemorySubscriptionRepository) GetChannelConfig(channelID string) (*models.ChannelConfig, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	config, exists := r.channels[channelID]
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
func (r *InMemorySubscriptionRepository) SaveChannelConfig(config *models.ChannelConfig) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.channels[config.ChannelID] = config
	return nil
}

// GetAllChannelConfigs retrieves all channel configurations
func (r *InMemorySubscriptionRepository) GetAllChannelConfigs() ([]*models.ChannelConfig, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	configs := make([]*models.ChannelConfig, 0, len(r.channels))
	for _, config := range r.channels {
		configs = append(configs, config)
	}

	return configs, nil
}

// SaveWebhookRegistration stores or updates a webhook registration
func (r *InMemorySubscriptionRepository) SaveWebhookRegistration(registration *models.WebhookRegistration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.webhooks[registration.ID] = registration
	return nil
}

// GetWebhookRegistration retrieves a webhook registration by id
func (r *InMemorySubscriptionRepository) GetWebhookRegistration(id string) (*models.WebhookRegistration, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	reg, exists := r.webhooks[id]
	if !exists {
		return nil, nil
	}
	return reg, nil
}

// DeleteWebhookRegistration deletes a webhook registration by id
func (r *InMemorySubscriptionRepository) DeleteWebhookRegistration(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.webhooks, id)
	return nil
}

// GetAllWebhookRegistrations returns all webhook registrations
func (r *InMemorySubscriptionRepository) GetAllWebhookRegistrations() ([]*models.WebhookRegistration, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	regs := make([]*models.WebhookRegistration, 0, len(r.webhooks))
	for _, reg := range r.webhooks {
		regs = append(regs, reg)
	}

	return regs, nil
}

// GetWebhookRegistrationsByChannel returns registrations for a specific channel
func (r *InMemorySubscriptionRepository) GetWebhookRegistrationsByChannel(channelID string) ([]*models.WebhookRegistration, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	regs := []*models.WebhookRegistration{}
	for _, reg := range r.webhooks {
		if reg.ChannelID == channelID {
			regs = append(regs, reg)
		}
	}
	return regs, nil
}
