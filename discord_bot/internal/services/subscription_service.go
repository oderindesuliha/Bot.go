package services

import (
	"fmt"

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
func (s *SubscriptionServiceImpl) SubscribeToMarket(discordUserID, marketID string) error {
	subscription, err := s.repo.GetSubscription(discordUserID)
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

	return s.repo.SaveSubscription(subscription)
}

// UnsubscribeFromMarket unsubscribes a user from a market
func (s *SubscriptionServiceImpl) UnsubscribeFromMarket(discordUserID, marketID string) error {
	subscription, err := s.repo.GetSubscription(discordUserID)
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
	return s.repo.SaveSubscription(subscription)
}

// SubscribeToCreator subscribes a user to a creator
func (s *SubscriptionServiceImpl) SubscribeToCreator(discordUserID, creator string) error {
	subscription, err := s.repo.GetSubscription(discordUserID)
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

	return s.repo.SaveSubscription(subscription)
}

// UnsubscribeFromCreator unsubscribes a user from a creator
func (s *SubscriptionServiceImpl) UnsubscribeFromCreator(discordUserID, creator string) error {
	subscription, err := s.repo.GetSubscription(discordUserID)
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
	return s.repo.SaveSubscription(subscription)
}

// GetUserSubscriptions gets a user's subscriptions
func (s *SubscriptionServiceImpl) GetUserSubscriptions(discordUserID string) (*models.Subscription, error) {
	return s.repo.GetSubscription(discordUserID)
}

// GetAllSubscriptions gets all subscriptions
func (s *SubscriptionServiceImpl) GetAllSubscriptions() ([]*models.Subscription, error) {
	return s.repo.GetAllSubscriptions()
}

// UpdateChannelConfig updates a channel's configuration
func (s *SubscriptionServiceImpl) UpdateChannelConfig(config *models.ChannelConfig) error {
	return s.repo.SaveChannelConfig(config)
}

// GetChannelConfig gets a channel's configuration
func (s *SubscriptionServiceImpl) GetChannelConfig(channelID string) (*models.ChannelConfig, error) {
	return s.repo.GetChannelConfig(channelID)
}

// GetAllChannelConfigs gets all channel configurations
func (s *SubscriptionServiceImpl) GetAllChannelConfigs() ([]*models.ChannelConfig, error) {
	return s.repo.GetAllChannelConfigs()
}

// ShouldNotifyUser determines if a user should be notified about a market
func (s *SubscriptionServiceImpl) ShouldNotifyUser(subscription *models.Subscription, market *models.Market) bool {
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
func (s *SubscriptionServiceImpl) SendNotificationToUser(discordUserID string, message string) error {
	// This would be implemented with actual Discord DM sending
	s.logger.Info(fmt.Sprintf("Would send DM to user %s: %s", discordUserID, message))
	return nil
}
