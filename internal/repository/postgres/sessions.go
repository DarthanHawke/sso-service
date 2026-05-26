package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"authn-gate-service/internal/lib/errors/apperr"
	"authn-gate-service/internal/models"

	"github.com/google/uuid"
)

type SessionRepository struct {
	db *Database
}

func NewSessionRepository(db *Database) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create создаёт новую сессию
func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	const op = "repository.postgres.SessionRepository.Create"

	query := `
        INSERT INTO sessions (id, user_id, app_id, user_ip, user_agent, refresh_token_hash, expires_at)
        VALUES (:id, :user_id, :app_id, :user_ip, :user_agent, :refresh_token_hash, :expires_at)
        RETURNING created_at, last_activity_at`

	rows, err := r.db.NamedQueryContext(ctx, query, map[string]any{
		"id":                 session.ID,
		"user_id":            session.UserID,
		"app_id":             session.AppID,
		"user_ip":            session.UserIP,
		"user_agent":         session.UserAgent,
		"refresh_token_hash": session.RefreshTokenHash,
		"expires_at":         session.ExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&session.CreatedAt, &session.LastActivityAt); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

// GetByTokenHash возвращает сессию по хешу refresh-токена
func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	const op = "repository.postgres.SessionRepository.GetByTokenHash"

	query := `
        SELECT id, user_id, app_id, user_ip, user_agent, refresh_token_hash,
               created_at, expires_at, last_activity_at
        FROM sessions
        WHERE refresh_token_hash = $1`

	session := &models.Session{}
	err := r.db.GetContext(ctx, session, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, apperr.ErrSessionNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return session, nil
}

// GetByID возвращает сессию по ID
func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	const op = "repository.postgres.SessionRepository.GetByID"

	query := `
        SELECT id, user_id, app_id, user_ip, user_agent, refresh_token_hash,
               created_at, expires_at, last_activity_at
        FROM sessions
        WHERE id = $1`

	session := &models.Session{}
	err := r.db.GetContext(ctx, session, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, apperr.ErrSessionNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return session, nil
}

// ListByUserID возвращает все активные сессии пользователя.
// Активные = не истёкшие (expires_at > NOW())
func (r *SessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Session, error) {
	const op = "repository.postgres.SessionRepository.ListByUserID"

	query := `
        SELECT id, user_id, app_id, user_ip, user_agent, refresh_token_hash,
               created_at, expires_at, last_activity_at
        FROM sessions
        WHERE user_id = $1 AND expires_at > NOW()
        ORDER BY created_at DESC`

	var sessions []models.Session
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return sessions, nil
}

// Delete удаляет сессию по ID
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.SessionRepository.Delete"

	result, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: %w", op, apperr.ErrSessionNotFound)
	}

	return nil
}

// DeleteAllByUserID удаляет все сессии пользователя
func (r *SessionRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	const op = "repository.postgres.SessionRepository.DeleteAllByUserID"

	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// UpdateLastActivity обновляет время последней активности сессии
func (r *SessionRepository) UpdateLastActivity(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.SessionRepository.UpdateLastActivity"

	result, err := r.db.ExecContext(ctx,
		`UPDATE sessions SET last_activity_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: %w", op, apperr.ErrSessionNotFound)
	}

	return nil
}
