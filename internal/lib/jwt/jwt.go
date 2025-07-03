package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sso-service/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Конфиг
type TokenConfig struct {
	accessTokenTTL  time.Duration // Время жизни Access токена
	refreshTokenTTL time.Duration // Время жизни Refresh токена
	issuer          string        // Идентификатор издателя
}

// Ключи
type TokenKeys struct {
	privateKey *rsa.PrivateKey // Приватный ключ для подписи
	publicKey  *rsa.PublicKey  // Публичный ключ для проверки
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
func (g *TokenGenerator) GenerateAccessToken(userID, sessionID uuid.UUID) (string, error) {
	claims := models.AccessTokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.config.accessTokenTTL)),
			Issuer:    g.config.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(g.keys.privateKey)
}

// GenerateRefreshToken создает Refresh токен
func (g *TokenGenerator) GenerateRefreshToken() (string, error) {
	claims := models.RefreshTokenClaims{
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.config.refreshTokenTTL)),
			Issuer:    g.config.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(g.keys.privateKey)
}
