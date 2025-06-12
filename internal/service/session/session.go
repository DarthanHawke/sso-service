package session

import (
	"context"
	"errors"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/lib/hash"

	"sso-service/internal/lib/jwt"
	"sso-service/internal/models"

	"time"

	"go.uber.org/zap"
)

const (
	nilToken = ""
)

type SessionService struct {
	logger        *zap.Logger
	sessionManage SessionManage
	sessionGet    SessionGet
	jwtManager    jwt.JWTManager
	tokenTTL      time.Duration
	hasher        hash.PasswordHasher
}

type SessionManage interface {
	CreateSession(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, sessionID int64) error
	DeleteAllUserSessions(ctx context.Context, userID int64) error
}

type SessionGet interface {
	GetSessionByToken(ctx context.Context, refreshToken string) (models.Session, error)
	GetSessionByUserId(ctx context.Context, userID int64) (models.Session, error)
	GetUserSessions(ctx context.Context, userID int64) ([]models.Session, error)
	IsSessionValid(ctx context.Context, refreshToken string) (bool, error)
}

func NewSessionService(
	logger *zap.Logger,
	sessionManage SessionManage,
	sessionGet SessionGet,
	jwtManager jwt.JWTManager,
	tokenTTL time.Duration,
	hasher hash.PasswordHasher,
) *SessionService {
	return &SessionService{
		logger:        logger,
		sessionManage: sessionManage,
		sessionGet:    sessionGet,
		jwtManager:    jwtManager,
		tokenTTL:      tokenTTL,
		hasher:        hasher,
	}
}

// Создание сессии
func (s *SessionService) CreateSession(ctx context.Context, userID int64) (string, string, error) {
	//TO DO: use logger

	// Генерация токенов
	refreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nilToken, nilToken, ssoerrors.ErrInternal
	}
	accessToken, err := s.jwtManager.GenerateAccessToken(userID)
	if err != nil {
		return nilToken, nilToken, ssoerrors.ErrInternal
	}

	// Хеширование токена
	refreshTokenHash, err := s.hasher.Hash(refreshToken)
	if err != nil {
		return nilToken, nilToken, ssoerrors.ErrInternal
	}

	expiresAt := time.Now().Add(s.tokenTTL)

	if err := s.sessionManage.CreateSession(ctx, userID, refreshTokenHash, expiresAt); err != nil {
		return nilToken, nilToken, ssoerrors.ErrInternal
	}

	return accessToken, refreshToken, nil
}

// Обновление сессии
func (s *SessionService) RefreshSession(
	ctx context.Context,
	refreshToken string,
	userID int64,
) (string, string, error) {
	//TO DO: use logger

	oldSession, err := s.sessionGet.GetSessionByUserId(ctx, userID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrSessionNotFound) {
			return nilToken, nilToken, ssoerrors.ErrInvalidToken
		}
		return nilToken, nilToken, ssoerrors.ErrInternal
	}

	if !s.hasher.Compare(refreshToken, oldSession.RefreshTokenHash) {
		return nilToken, nilToken, ssoerrors.ErrInvalidToken
	}

	// Удаляем старую сессию
	if err := s.sessionManage.DeleteSession(ctx, oldSession.ID); err != nil {
		return nilToken, nilToken, ssoerrors.ErrInternal
	}

	// Создаем новую
	return s.CreateSession(ctx, userID)
}

// Выход (удаление сессии)
func (s *SessionService) Logout(ctx context.Context, sessionID int64) error {
	//TO DO: use logger

	if err := s.sessionManage.DeleteSession(ctx, sessionID); err != nil {
		if errors.Is(err, ssoerrors.ErrSessionNotFound) {
			return ssoerrors.ErrSessionNotFound
		}
		return ssoerrors.ErrInternal
	}
	return nil
}

// Выход со всех устройств
func (s *SessionService) LogoutAll(ctx context.Context, userID int64) error {
	//TO DO: use logger

	if err := s.sessionManage.DeleteAllUserSessions(ctx, userID); err != nil {
		return ssoerrors.ErrInternal
	}
	return nil
}
