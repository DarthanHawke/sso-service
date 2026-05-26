package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `db:"id"`
	Email         string     `db:"email"`
	EmailVerified bool       `db:"email_verified"`
	FullName      string     `db:"full_name"`
	Disabled      bool       `db:"disabled"`
	PasswordHash  string     `db:"password_hash"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	LastLoginAt   *time.Time `db:"last_login_at"`
}
