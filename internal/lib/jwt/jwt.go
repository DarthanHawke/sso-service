package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager interface {
	GenerateAccessToken(userID []uint8) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(tokenString string) (*AccessTokenClaims, error)
	GetAccessTokenTTL() time.Duration
	GetRefreshTokenTTL() time.Duration
}

// Конфиг
type TokenConfig struct {
	accessTokenTTL  time.Duration // Время жизни Access токена (например, 15m)
	refreshTokenTTL time.Duration // Время жизни Refresh токена (например, 720h - 30 дней)
	issuer          string        // Идентификатор издателя (название вашего сервиса)
}

// Ключи
type TokenKeys struct {
	privateKey *rsa.PrivateKey // Приватный ключ для подписи
	publicKey  *rsa.PublicKey  // Публичный ключ для проверки
}

// Claims для Access токена
type AccessTokenClaims struct {
	UserID    []uint8 `json:"user_id"`
	SessionID string  `json:"sid"` // Добавьте ID сессии для инвалидации
	jwt.RegisteredClaims
}

// Генератор токенов
type TokenGenerator struct {
	config *TokenConfig
	keys   *TokenKeys
}

func NewTokenGenerator(
	accessTokenTTL, refreshTokenTTL time.Duration,
	issuer string,
	privateKeyPath, publicKeyPath string,
) (*TokenGenerator, error) {
	keys, err := NewTokenKeys(privateKeyPath, publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load keys: %w", err)
	}
	return &TokenGenerator{
		config: NewTokenConfig(accessTokenTTL, refreshTokenTTL, issuer),
		keys:   keys,
	}, nil
}

func NewTokenConfig(accessTokenTTL, refreshTokenTTL time.Duration, issuer string) *TokenConfig {
	return &TokenConfig{
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		issuer:          issuer,
	}
}

func NewTokenKeys(privateKeyPath, publicKeyPath string) (*TokenKeys, error) {
	privBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	pubBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	privBlock, _ := pem.Decode(privBytes)
	pubBlock, _ := pem.Decode(pubBytes)

	privKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, err
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return &TokenKeys{
		privateKey: privKey.(*rsa.PrivateKey),
		publicKey:  pubKey.(*rsa.PublicKey),
	}, nil
}

func (g *TokenGenerator) GetAccessTokenTTL() time.Duration {
	return g.config.accessTokenTTL
}

func (g *TokenGenerator) GetRefreshTokenTTL() time.Duration {
	return g.config.refreshTokenTTL
}

// GenerateAccessToken создает JWT Access токен
func (g *TokenGenerator) GenerateAccessToken(userID []uint8) (string, error) {
	claims := AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.config.accessTokenTTL)),
			Issuer:    g.config.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(g.keys.privateKey)
}

// GenerateRefreshToken создает Refresh токен (простая UUID + подпись)
func (g *TokenGenerator) GenerateRefreshToken() (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.config.refreshTokenTTL)),
		Issuer:    g.config.issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(g.keys.privateKey)
}

// ValidateAccessToken проверяет Access токен
func (g *TokenGenerator) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.keys.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
