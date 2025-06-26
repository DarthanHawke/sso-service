package models

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims для Access токена
type AccessTokenClaims struct {
	UserID    []uint8 `json:"user_id"`
	SessionID string  `json:"sid"`
	jwt.RegisteredClaims
}
