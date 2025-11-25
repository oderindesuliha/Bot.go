package models

// Subscription represents a user's subscription to markets or creators
type Subscription struct {
	DiscordUserID      string   `json:"discord_user_id"`
	SubscribedMarkets  []string `json:"subscribed_markets"`  // market IDs
	SubscribedCreators []string `json:"subscribed_creators"` // creator names
}
