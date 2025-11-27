package models

import "time"

// WebhookRegistration represents a registered Discord webhook for a channel
type WebhookRegistration struct {
	ID                string    `json:"id"`
	ChannelID         string    `json:"channel_id"`
	WebhookURL        string    `json:"webhook_url"`
	Events            []string  `json:"events"`
	Frequency         string    `json:"frequency"` // low|medium|high
	AllowedCategories []string  `json:"allowed_categories"`
	CreatedAt         time.Time `json:"created_at"`
}
