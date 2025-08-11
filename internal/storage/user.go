package storage

import (
	"context"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
)

type UserDataBase struct {
	db *Database
	// TO DO: Redis Cache
}

func NewUserDataBase(db *Database) *UserDataBase {
	return &UserDataBase{
		db: db,
	}
}

// CreateUser создаёт нового пользователя models.User, используя email, passwordHash, fullName string, возвращает userID
func (userDB *UserDataBase) CreateUser(ctx context.Context, email, passwordHash, fullName string) (uuid.UUID, error) {
	const op = "storage.user.CreateUser"

	var id uuid.UUID

	err := userDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		err := tx.QueryRowxContext(ctx, `
            INSERT INTO users
                (email, password_hash, full_name, created_at, updated_at) 
            VALUES
                ($1, $2, $3, NOW(), NOW()) 
            RETURNING id`,
			email, passwordHash, fullName,
		).Scan(&id)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
			}
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// GetUserByID возвращант пользователя models.User, используя userID
func (userDB *UserDataBase) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const op = "storage.user.GetUserByID"

	var user models.User
	err := userDB.db.GetContext(ctx, &user, `
        SELECT 
            id,
            email, 
			password_hash,
            full_name, 
            created_at, 
            updated_at 
        FROM users WHERE id = $1
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

// GetUserByEmail возвращант пользователя models.User, используя email
func (userDB *UserDataBase) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "storage.user.GetUserByEmail"

	var user models.User
	err := userDB.db.GetContext(ctx, &user, `
        SELECT 
            id,
            email, 
			password_hash,
            full_name, 
            created_at, 
            updated_at 
        FROM users WHERE email = $1
    `, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

// GetListUsers возвращает список всех пользователей
func (userDB *UserDataBase) GetListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	const op = "storage.user.GetListUsers"

	var users []models.User

	query := `
        SELECT 
            id,
            email, 
            full_name, 
            created_at, 
            updated_at 
        FROM users 
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
    `

	// Если limit > 0, добавляем LIMIT и OFFSET
	if limit > 0 {
		query += " LIMIT $1 OFFSET $2"
		err := userDB.db.SelectContext(ctx, &users, query, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	} else {
		// Если limit не указан, выполняем запрос без ограничений
		err := userDB.db.SelectContext(ctx, &users, query)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return users, nil
}

// UpdateUser обновляет email и/или fullName пользователя models.User, используя userID
func (userDB *UserDataBase) UpdateUserEmail(ctx context.Context, userID uuid.UUID, email string) error {
	const op = "storage.user.UpdateUserEmail"

	return userDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, `
            UPDATE users 
            SET email = $1, updated_at = NOW() 
            WHERE id = $2 AND deleted_at IS NULL`,
			email, userID,
		)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
			}
			return fmt.Errorf("%s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}

		return nil
	})
}

// UpdateUser обновляет email и/или fullName пользователя models.User, используя userID
func (userDB *UserDataBase) UpdateUserName(ctx context.Context, userID uuid.UUID, fullName string) error {
	const op = "storage.user.UpdateUserName"

	return userDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, `
            UPDATE users 
            SET full_name = $1, updated_at = NOW() 
            WHERE id = $2 AND deleted_at IS NULL`,
			fullName, userID,
		)

		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}

		return nil
	})
}

// UpdateUser обновляет email и/или fullName пользователя models.User, используя userID
func (userDB *UserDataBase) UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
	const op = "storage.user.UpdateUserPassword"

	return userDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, `
            UPDATE users 
            SET password_hash = $1, updated_at = NOW() 
            WHERE id = $2 AND deleted_at IS NULL`,
			newPasswordHash, userID,
		)

		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}

		return nil
	})
}

// DeleteUserSoft "мягкое" удаление пользователя по userID
func (userDB *UserDataBase) DeleteUserSoft(ctx context.Context, userID uuid.UUID) error {
	const op = "storage.user.DeleteUserSoft"

	return userDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Удаляем сессии пользователя
		if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = $1", userID); err != nil {
			return fmt.Errorf("%s: failed to delete sessions: %w", op, err)
		}

		// Удаляем связи с ролями
		if _, err := tx.ExecContext(ctx, "DELETE FROM user_roles WHERE user_id = $1", userID); err != nil {
			return fmt.Errorf("%s: failed to delete user roles: %w", op, err)
		}

		// "Мягкое" удаление пользователя (устанавливаем deleted_at)
		result, err := tx.ExecContext(ctx,
			"UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
			userID,
		)
		if err != nil {
			return fmt.Errorf("%s: failed to soft delete user: %w", op, err)
		}

		// Проверяем, что пользователь действительно был обновлен
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: failed to get affected rows: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}

		return nil
	})
}

// DeleteUser удаляет пользователя (с каскадным удалением связанных данных)
func (userDB *UserDataBase) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "storage.role.DeleteUser"

	return userDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}

		return nil
	})
}
