package models

import "time"

// CoinTransaction представляет операцию с монетами.
type CoinTransaction struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	Amount        int       `json:"amount"`
	Type          string    `json:"type"` // например, "transfer_sent" или "transfer_received"
	RelatedUserID *int64    `json:"related_user_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
