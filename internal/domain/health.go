package domain

import "time"

// Health represents the health status of the application.
type Health struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Database  string    `json:"database"`
}
