package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type SessionDataBase struct {
	db *Database
	// TO DO: Redis Cache
}

func NewSessionDataBase(db *Database) *SessionDataBase {
	return &SessionDataBase{
		db: db,
	}
}

func (sessionDB *SessionDataBase) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	refreshToken string,
	expiresAt time.Time,
) error {
	const op = "storage.session.CreateSession"

	ip, _ := ctx.Value("ip").(string)

	userAgent, _ := ctx.Value("user_agent").(string)

	stmt, err := sessionDB.db.Prepare(`
        INSERT INTO sessions (
            id, 
            user_id, 
            refresh_token_hash, 
            user_ip, 
            user_agent, 
            expires_at,
			created_at
        ) VALUES (
            gen_random_uuid(), 
            $1, $2, $3, $4, $5, NOW()
        )`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx,
		userID,
		refreshToken,
		ip,        // Получаем IP из контекста
		userAgent, // Получаем User-Agent из контекста
		expiresAt,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetSessionByToken возвращает сессию по refresh-токену
func (sessionDB *SessionDataBase) GetSessionByToken(
	ctx context.Context,
	refreshToken string,
) (*models.Session, error) {
	const op = "storage.session.GetSessionByToken"

	var session models.Session
	err := sessionDB.db.GetContext(ctx, &session, `
        SELECT 
            id, 
            user_id, 
            refresh_token_hash, 
            user_ip, 
            user_agent, 
            expires_at,
			created_at
        FROM sessions
        WHERE refresh_token_hash = $1 AND expires_at > NOW()`,
		refreshToken,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrSessionNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &session, nil
}

// GetUserSessions возвращает все активные сессии пользователя
func (sessionDB *SessionDataBase) GetUserSessions(
	ctx context.Context,
	userID uuid.UUID,
) (*[]models.Session, error) {
	const op = "storage.session.GetUserSessions"

	var sessions []models.Session
	err := sessionDB.db.SelectContext(ctx, &sessions, `
        SELECT 
            id, 
            user_id, 
            refresh_token_hash, 
            user_ip, 
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

	return &sessions, nil
}

// DeleteSession удаляет конкретную сессию по ID
func (sessionDB *SessionDataBase) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	const op = "storage.session.DeleteSession"

	result, err := sessionDB.db.ExecContext(ctx, `
        DELETE FROM sessions 
        WHERE id = $1`,
		sessionID,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrSessionNotFound)
	}

	return nil
}

// DeleteAllUserSessions удаляет все сессии пользователя
func (sessionDB *SessionDataBase) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	const op = "storage.session.DeleteAllUserSessions"

	_, err := sessionDB.db.ExecContext(ctx, `
        DELETE FROM sessions 
        WHERE user_id = $1`,
		userID,
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// IsSessionValid проверяет валидность сессии
func (sessionDB *SessionDataBase) IsSessionValid(ctx context.Context, refreshToken string) (bool, error) {
	const op = "storage.session.IsSessionValid"

	var exists bool
	err := sessionDB.db.GetContext(ctx, &exists, `
        SELECT EXISTS (
            SELECT 1 FROM sessions 
            WHERE refresh_token_hash = $1 AND expires_at > NOW()
        )`,
		refreshToken,
	)

	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}
