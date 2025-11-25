package models

import "time"

// ChannelConfig represents configuration for a Discord channel
type ChannelConfig struct {
	ChannelID           string    `json:"channel_id"`
	FeedEnabled         bool      `json:"feed_enabled"`
	AllowedCategories   []string  `json:"allowed_categories"`
	FrequencyMode       string    `json:"frequency_mode"` // low, medium, high
	LastUpdateTimestamp time.Time `json:"last_update_timestamp"`
}
