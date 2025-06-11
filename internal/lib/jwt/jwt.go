package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager interface {
	GenerateAccessToken(userID string, roles []string) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
}

// Конфиг
type TokenConfig struct {
	AccessTokenExpiry  time.Duration // Время жизни Access токена (например, 15m)
	RefreshTokenExpiry time.Duration // Время жизни Refresh токена (например, 720h - 30 дней)
	Issuer             string        // Идентификатор издателя (название вашего сервиса)
}

// Ключи
type TokenKeys struct {
	PrivateKey *rsa.PrivateKey // Приватный ключ для подписи
	PublicKey  *rsa.PublicKey  // Публичный ключ для проверки
}

// Claims для Access токена
type AccessTokenClaims struct {
	UserID    string   `json:"user_id"`
	Roles     []string `json:"roles"`
	SessionID string   `json:"sid"` // Добавьте ID сессии для инвалидации
	jwt.RegisteredClaims
}

// Генератор токенов
type TokenGenerator struct {
	config TokenConfig
	keys   TokenKeys
}

func NewTokenGenerator(cfg TokenConfig, keys TokenKeys) *TokenGenerator {
	return &TokenGenerator{
		config: cfg,
		keys:   keys,
	}
}

// GenerateAccessToken создает JWT Access токен
func (g *TokenGenerator) GenerateAccessToken(userID string, roles []string) (string, error) {
	claims := AccessTokenClaims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.config.AccessTokenExpiry)),
			Issuer:    g.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(g.keys.PrivateKey)
}

// GenerateRefreshToken создает Refresh токен (простая UUID + подпись)
func (g *TokenGenerator) GenerateRefreshToken() (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.config.RefreshTokenExpiry)),
		Issuer:    g.config.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(g.keys.PrivateKey)
}

// ValidateAccessToken проверяет Access токен
func (g *TokenGenerator) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.keys.PublicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
