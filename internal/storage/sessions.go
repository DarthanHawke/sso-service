package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso-service/internal/models"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type Sessions interface {
	// Создание
	CreateSession(ctx context.Context, userID, refreshToken string, expiresAt time.Time) error

	// Чтение
	GetSessionByToken(ctx context.Context, refreshToken string) (models.Session, error)
	GetUserSessions(ctx context.Context, userID string) ([]models.Session, error)

	// Удаление
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteAllUserSessions(ctx context.Context, userID string) error

	// Проверки
	IsSessionValid(ctx context.Context, refreshToken string) (bool, error)
}

type SessionsDataBase struct {
	db *Database
	// TO DO: Redis Cache
}

func (sessionsDB *SessionsDataBase) CreateSession(
	ctx context.Context,
	userID, refreshToken string,
	expiresAt time.Time,
) error {
	stmt, err := sessionsDB.db.Prepare(`
        INSERT INTO sessions (
            id, 
            user_id, 
            refresh_token, 
            ip, 
            user_agent, 
            expires_at
        ) VALUES (
            gen_random_uuid(), 
            $1, $2, $3, $4, $5
        )`)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	_, err = stmt.ExecContext(ctx,
		userID,
		refreshToken,
		ctx.Value("ip").(string),         // Получаем IP из контекста
		ctx.Value("user_agent").(string), // Получаем User-Agent из контекста
		expiresAt,
	)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// GetSessionByToken возвращает сессию по refresh-токену
func (sessionsDB *SessionsDataBase) GetSessionByToken(
	ctx context.Context,
	refreshToken string,
) (models.Session, error) {
	var session models.Session
	err := sessionsDB.db.GetContext(ctx, &session, `
        SELECT 
            id, 
            user_id, 
            refresh_token, 
            ip, 
            user_agent, 
            expires_at, 
            created_at
        FROM sessions
        WHERE refresh_token = $1 AND expires_at > NOW()`,
		refreshToken,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Session{}, fmt.Errorf("%w", ErrSessionNotFound)
	}

	if err != nil {
		return models.Session{}, fmt.Errorf("%w", err)
	}

	return session, nil
}

// GetUserSessions возвращает все активные сессии пользователя
func (sessionsDB *SessionsDataBase) GetUserSessions(
	ctx context.Context,
	userID string,
) ([]models.Session, error) {
	const op = "repository.GetUserSessions"

	var sessions []models.Session
	err := sessionsDB.db.SelectContext(ctx, &sessions, `
        SELECT 
            id, 
            user_id, 
            refresh_token, 
            ip, 
            user_agent, 
            expires_at, 
            created_at
        FROM sessions
        WHERE user_id = $1 AND expires_at > NOW()
        ORDER BY created_at DESC`,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sessions, nil
}

// DeleteSession удаляет конкретную сессию по ID
func (sessionsDB *SessionsDataBase) DeleteSession(ctx context.Context, sessionID string) error {
	result, err := sessionsDB.db.ExecContext(ctx, `
        DELETE FROM sessions 
        WHERE id = $1`,
		sessionID,
	)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w", ErrSessionNotFound)
	}

	return nil
}

// DeleteAllUserSessions удаляет все сессии пользователя
func (sessionsDB *SessionsDataBase) DeleteAllUserSessions(ctx context.Context, userID string) error {
	_, err := sessionsDB.db.ExecContext(ctx, `
        DELETE FROM sessions 
        WHERE user_id = $1`,
		userID,
	)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// IsSessionValid проверяет валидность сессии
func (sessionsDB *SessionsDataBase) IsSessionValid(ctx context.Context, refreshToken string) (bool, error) {
	var exists bool
	err := sessionsDB.db.GetContext(ctx, &exists, `
        SELECT EXISTS (
            SELECT 1 FROM sessions 
            WHERE refresh_token = $1 AND expires_at > NOW()
        )`,
		refreshToken,
	)

	if err != nil {
		return false, fmt.Errorf(" %w", err)
	}

	return exists, nil
}
