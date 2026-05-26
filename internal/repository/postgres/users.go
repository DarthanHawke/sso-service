package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"authn-gate-service/internal/lib/errors/apperr"
	"authn-gate-service/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UserRepository struct {
	db *Database
}

func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{db: db}
}

// Create создаёт нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	const op = "repository.postgres.UserRepository.Create"

	query := `
        INSERT INTO users (id, email, password_hash, full_name, email_verified, disabled)
        VALUES (:id, :email, :password_hash, :full_name, :email_verified, :disabled)
        RETURNING created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, map[string]any{
		"id":             user.ID,
		"email":          user.Email,
		"password_hash":  user.PasswordHash,
		"full_name":      user.FullName,
		"email_verified": user.EmailVerified,
		"disabled":       user.Disabled,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, apperr.ErrUserExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&user.CreatedAt, &user.UpdatedAt); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

// GetByID возвращает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "repository.postgres.UserRepository.GetByID"

	query := `
        SELECT id, email, email_verified, full_name, disabled, password_hash, 
               created_at, updated_at, last_login_at
        FROM users
        WHERE id = $1`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, apperr.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// GetByEmail возвращает пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "repository.postgres.UserRepository.GetByEmail"

	query := `
        SELECT id, email, email_verified, full_name, disabled, password_hash, 
               created_at, updated_at, last_login_at
        FROM users
        WHERE email = $1`

	user := &models.User{}
	err := r.db.GetContext(ctx, user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, apperr.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// Update обновляет данные пользователя.
// Обновляет все поля, кроме ID.
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	const op = "repository.postgres.UserRepository.Update"

	query := `
        UPDATE users 
        SET email = :email,
            email_verified = :email_verified,
            full_name = :full_name,
            disabled = :disabled,
            password_hash = :password_hash,
            updated_at = :updated_at,
            last_login_at = :last_login_at
        WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("%s: %w", op, apperr.ErrUserExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: %w", op, apperr.ErrUserNotFound)
	}

	return nil
}

// Delete удаляет пользователя по ID (hard delete)
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.UserRepository.Delete"

	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rows == 0 {
		return fmt.Errorf("%s: %w", op, apperr.ErrUserNotFound)
	}

	return nil
}
