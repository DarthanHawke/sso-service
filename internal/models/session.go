package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID  `db:"id"`
	UserID           uuid.UUID  `db:"user_id"`
	AppID            string     `db:"app_id"`
	UserIP           string     `db:"user_ip"`
	UserAgent        string     `db:"user_agent"`
	RefreshTokenHash string     `db:"refresh_token_hash"`
	CreatedAt        time.Time  `db:"created_at"`
	ExpiresAt        time.Time  `db:"expires_at"`
	LastActivityAt   *time.Time `db:"last_activity_at"`
}
