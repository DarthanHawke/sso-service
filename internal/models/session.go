package models

import "time"

type Session struct {
	ID               []uint8   `json:"id" db:"id"`
	UserID           []uint8   `json:"user_id" db:"user_id"`
	RefreshTokenHash string    `json:"refresh_token_hash" db:"refresh_token_hash"`
	UserIP           string    `json:"user_ip" db:"user_ip"`
	UserAgent        string    `json:"user_agent" db:"user_agent"`
	ExpiresAt        time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}
