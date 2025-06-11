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

type SessionService struct {
	logger        *zap.Logger
	sessionManage SessionManage
	sessionGet    SessionGet
	jwtManager    jwt.JWTManager
	tokenTTL      time.Duration
	hasher        hash.PasswordHasher
}

type SessionManage interface {
	CreateSession(ctx context.Context, userID, refreshToken string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteAllUserSessions(ctx context.Context, userID string) error
}

type SessionGet interface {
	GetSessionByToken(ctx context.Context, refreshToken string) (models.Session, error)
	GetUserSessions(ctx context.Context, userID string) ([]models.Session, error)
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
func (s *SessionService) CreateSession(ctx context.Context, userID string) (models.Session, error) {
	//TO DO: use logger

	// Генерация токенов
	refreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return models.Session{}, ssoerrors.ErrInternal
	}

	// Хеширование токена
	refreshToken, err = s.hasher.Hash(refreshToken)
	if err != nil {
		return models.Session{}, ssoerrors.ErrInternal
	}

	expiresAt := time.Now().Add(s.tokenTTL)

	// Сохранение сессии
	session := models.Session{
		UserID:           userID,
		RefreshTokenHash: refreshToken,
		ExpiresAt:        expiresAt,
	}

	if err := s.sessionManage.CreateSession(ctx, userID, refreshToken, expiresAt); err != nil {
		return models.Session{}, ssoerrors.ErrInternal
	}

	return session, nil
}

// Обновление сессии
func (s *SessionService) RefreshSession(
	ctx context.Context,
	refreshToken string,
) (models.Session, error) {
	//TO DO: use logger

	oldSession, err := s.sessionGet.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrSessionNotFound) {
			return models.Session{}, ssoerrors.ErrInvalidToken
		}
		return models.Session{}, ssoerrors.ErrInternal
	}

	// Удаляем старую сессию
	if err := s.sessionManage.DeleteSession(ctx, oldSession.ID); err != nil {
		return models.Session{}, ssoerrors.ErrInternal
	}

	// Создаем новую
	return s.CreateSession(ctx, oldSession.UserID)
}

// Выход (удаление сессии)
func (s *SessionService) Logout(ctx context.Context, sessionID string) error {
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
func (s *SessionService) LogoutAll(ctx context.Context, userID string) error {
	//TO DO: use logger

	if err := s.sessionManage.DeleteAllUserSessions(ctx, userID); err != nil {
		return ssoerrors.ErrInternal
	}
	return nil
}
