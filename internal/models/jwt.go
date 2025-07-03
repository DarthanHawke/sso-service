package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims для Access токена
type AccessTokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	SessionID uuid.UUID `json:"session_id"`
	TokenType string    `json:"token_type"`
	jwt.RegisteredClaims
}

// Claims для Refresh токена
type RefreshTokenClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}
