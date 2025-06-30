package session

import (
	"context"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"

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
	jwtManager    JWTManager
	hasher        HashManager
}

type JWTManager interface {
	GenerateAccessToken(userID []uint8) (string, error)
	GenerateRefreshToken() (string, error)
	GetAccessTokenTTL() time.Duration
	GetRefreshTokenTTL() time.Duration
}

type HashManager interface {
	HashToken(token string) string
}

type SessionManage interface {
	CreateSession(ctx context.Context, userID []uint8, refreshToken string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, sessionID []uint8) error
	DeleteAllUserSessions(ctx context.Context, userID []uint8) error
}

type SessionGet interface {
	GetSessionByToken(ctx context.Context, refreshToken string) (models.Session, error)
	GetUserSessions(ctx context.Context, userID []uint8) ([]models.Session, error)
	IsSessionValid(ctx context.Context, refreshToken string) (bool, error)
}

func NewSessionService(
	logger *zap.Logger,
	sessionManage SessionManage,
	sessionGet SessionGet,
	jwtManager JWTManager,
	hasher HashManager,
) *SessionService {
	return &SessionService{
		logger:        logger.With(zap.String("component", "sso_service")),
		sessionManage: sessionManage,
		sessionGet:    sessionGet,
		jwtManager:    jwtManager,
		hasher:        hasher,
	}
}

// CreateSession - cоздание сессии
func (s *SessionService) CreateSession(ctx context.Context, userID []uint8) (string, string, error) {
	const op = "service.session.CreateRole"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("creating new session")

	// Генерация токенов
	refreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		s.logger.Error("generate refresh token", zap.Uint8s("UserID", userID), zap.Error(err))

		return nilToken, nilToken, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}
	accessToken, err := s.jwtManager.GenerateAccessToken(userID)
	if err != nil {
		s.logger.Error("generate accsess token", zap.Uint8s("UserID", userID), zap.Error(err))

		return nilToken, nilToken, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Хеширование токена
	refreshTokenHash := s.hasher.HashToken(refreshToken)

	expiresAt := time.Now().Add(s.jwtManager.GetRefreshTokenTTL())

	if err := s.sessionManage.CreateSession(ctx, userID, refreshTokenHash, expiresAt); err != nil {
		s.logger.Error("hashing token", zap.Uint8s("UserID", userID), zap.Error(err))

		return nilToken, nilToken, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully created session",
		zap.Uint8s("UserID", userID),
	)
	return accessToken, refreshToken, nil
}

// RefreshSession - обновление сессии
func (s *SessionService) RefreshSession(
	ctx context.Context,
	userID []uint8,
	refreshToken string,
) (string, string, error) {
	const op = "service.session.RefreshSession"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("refreshing session")

	// Хеширование токена
	refreshTokenHash := s.hasher.HashToken(refreshToken)

	oldSession, err := s.sessionGet.GetSessionByToken(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrSessionNotFound) {
			s.logger.Warn("sessinon dont get", zap.Uint8s("UserID", userID), zap.Error(err))

			return nilToken, nilToken, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
		}
		s.logger.Warn("sessinon dont get", zap.Uint8s("UserID", userID), zap.Error(err))

		return nilToken, nilToken, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Проверяем срок действия
	if time.Now().After(oldSession.ExpiresAt) {
		return nilToken, nilToken, ssoerrors.ErrSessionOld
	}

	// Удаляем старую сессию
	if err := s.sessionManage.DeleteSession(ctx, oldSession.ID); err != nil {
		s.logger.Error("cannot delete session",
			zap.Uint8s("UserID", userID),
			zap.Uint8s("SessionID", oldSession.ID),
			zap.Error(err),
		)

		return nilToken, nilToken, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Создаем новую
	return s.CreateSession(ctx, userID)
}

// Logout - выход (удаление сессии)
func (s *SessionService) Logout(ctx context.Context, userID []uint8, sessionID []uint8) error {
	const op = "service.session.Logout"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("logout session")

	if err := s.sessionManage.DeleteSession(ctx, sessionID); err != nil {
		if errors.Is(err, ssoerrors.ErrSessionNotFound) {
			s.logger.Warn("cannot delete session", zap.Uint8s("SessionID", sessionID), zap.Error(err))

			return fmt.Errorf("%s: %w", op, ssoerrors.ErrSessionNotFound)
		}
		s.logger.Error("cannot delete session", zap.Uint8s("SessionID", sessionID), zap.Error(err))

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully logout",
		zap.Uint8s("UserID", userID),
	)
	return nil
}

// LogoutAll - выход со всех устройств
func (s *SessionService) LogoutAll(ctx context.Context, userID []uint8) error {
	const op = "service.session.LogoutAll"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("logout all session")

	if err := s.sessionManage.DeleteAllUserSessions(ctx, userID); err != nil {
		s.logger.Error("cannot delete sessions", zap.Uint8s("UserID", userID), zap.Error(err))

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully logout from all devices",
		zap.Uint8s("UserID", userID),
	)
	return nil
}
