package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	"github.com/lib/pq"
)

type UserDataBase struct {
	db *Database
	// TO DO: Redis Cache
}

// CreateUser создаёт нового пользователя models.User, используя email, passwordHash, fullName string, возвращает userID
func (userDB *UserDataBase) CreateUser(ctx context.Context, email, passwordHash, fullName string) (int64, error) {
	stmt, err := userDB.db.Prepare("INSERT INTO users(email, password_hash, full_name) VALUES(?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}

	result, err := stmt.ExecContext(ctx, email, passwordHash, fullName)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // 23505 = unique_violation
			return 0, fmt.Errorf("%w", ssoerrors.ErrUserExists)
		}

		return 0, fmt.Errorf("%w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}

	return id, nil
}

// GetUserByID возвращант пользователя models.User, используя userID
func (userDB *UserDataBase) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	stmt, err := userDB.db.Prepare(`
		SELECT 
		email, 
		password_hash, 
		full_name, 
		is_active, 
		created_at, 
		updated_at 
		FROM users WHERE id = ?
	`)

	if err != nil {
		return models.User{}, fmt.Errorf("%w", err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%w", ssoerrors.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%w", err)
	}

	return user, nil
}

// GetUserByEmail возвращант пользователя models.User, используя email
func (userDB *UserDataBase) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	stmt, err := userDB.db.Prepare(`
		SELECT 
			email, 
			password_hash, 
			full_name, 
			is_active, 
			created_at, 
			updated_at 
		FROM users WHERE email = ?
	`)

	if err != nil {
		return models.User{}, fmt.Errorf("%w", err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%w", ssoerrors.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%w", err)
	}

	return user, nil
}

// GetListUsers возвращает список всех пользователей
func (userDB *UserDataBase) GetListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	var users []models.User
	err := userDB.db.SelectContext(ctx, &users, `
        SELECT 
        	email, 
			password_hash, 
			full_name, 
			is_active, 
			created_at, 
			updated_at 
        FROM users 
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return users, nil
}

// UpdateUser обновляет email и/или fullName пользователя models.User, используя userID
func (userDB *UserDataBase) UpdateUser(ctx context.Context, userID string, email, fullName string) error {
	stmt, err := userDB.db.Prepare("UPDATE users SET email = $1, full_name = $2 updated_at = NOW() WHERE id = $3")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	result, err := stmt.ExecContext(ctx, email, fullName, userID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ssoerrors.ErrNotFound
	}

	return nil
}

// UpdateUser обновляет email и/или fullName пользователя models.User, используя userID
func (userDB *UserDataBase) UpdatePassword(ctx context.Context, userID, newPasswordHash string) error {
	stmt, err := userDB.db.Prepare("UPDATE users SET password_hash = $1 updated_at = NOW() WHERE id = $2")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	result, err := stmt.ExecContext(ctx, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ssoerrors.ErrNotFound
	}

	return nil
}

// DeleteUser "мягкое" удаление пользователя по userID
func (userDB *UserDataBase) DeleteUser(ctx context.Context, userID string) error {
	// Используем транзакцию, так как нужно удалить связанные данные
	tx, err := userDB.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer tx.Rollback()

	// Удаляем сессии пользователя
	if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}

	// Удаляем связи с ролями
	if _, err := tx.ExecContext(ctx, "DELETE FROM user_roles WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("delete user roles: %w", err)
	}

	// "Мягкое" удаление пользователя (помечаем deleted_at)
	if _, err := tx.ExecContext(ctx,
		"UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		userID,
	); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return tx.Commit()
}
