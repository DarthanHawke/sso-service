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
	"github.com/jmoiron/sqlx"
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
	ip, userAgent string,
	expiresAt time.Time,
) error {
	const op = "storage.session.CreateSession"
	return sessionDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, `
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
			)`,
			userID,
			refreshToken,
			ip,
			userAgent,
			expiresAt,
		)

		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
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
) ([]models.Session, error) {
	const op = "storage.session.GetUserSessions"

	var sessions []models.Session
	err := sessionDB.db.SelectContext(ctx, &sessions, `
        SELECT 
            id, 
            user_id, 
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

	return sessions, nil
}

// DeleteSession удаляет конкретную сессию по ID
func (sessionDB *SessionDataBase) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	const op = "storage.session.DeleteSession"

	return sessionDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, `
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
	})
}

// DeleteAllUserSessions удаляет все сессии пользователя
func (sessionDB *SessionDataBase) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	const op = "storage.session.DeleteAllUserSessions"

	return sessionDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Сначала удаляем все refresh токены пользователя
		_, err := tx.ExecContext(ctx, `
			DELETE FROM sessions 
			WHERE user_id = $1`,
			userID,
		)

		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
}
