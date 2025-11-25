package models

import "time"

// Market represents a market in Coral Markets
type Market struct {
	ID              string    `json:"market_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Outcomes        []string  `json:"outcomes"`
	Percentages     []float64 `json:"percentages"`
	Category        string    `json:"category"`
	Creator         string    `json:"creator"`
	Volume          float64   `json:"volume"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Status          string    `json:"status"` // active, closed, resolved
	ResolvedOutcome string    `json:"resolved_outcome,omitempty"`
	Link            string    `json:"link"`
}
