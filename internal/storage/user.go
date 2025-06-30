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

func NewUserDataBase(db *Database) *UserDataBase {
	return &UserDataBase{
		db: db,
	}
}

// CreateUser создаёт нового пользователя models.User, используя email, passwordHash, fullName string, возвращает userID
func (userDB *UserDataBase) CreateUser(ctx context.Context, email, passwordHash, fullName string) ([]uint8, error) {
	const op = "storage.user.CreateUser"

	stmt, err := userDB.db.Prepare(`
		INSERT INTO users
				(email, password_hash, full_name, created_at, updated_at) 
			VALUES
				($1, $2, $3, NOW(), NOW()) 
			RETURNING id
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email, passwordHash, fullName)
	var id []uint8
	err = row.Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // 23505 = unique_violation
			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// GetUserByID возвращант пользователя models.User, используя userID
func (userDB *UserDataBase) GetUserByID(ctx context.Context, userID []uint8) (models.User, error) {
	const op = "storage.user.GetUserByID"

	stmt, err := userDB.db.Prepare(`
		SELECT 
			id,
			email, 
			password_hash, 
			full_name, 
			created_at, 
			updated_at 
		FROM users WHERE id = $1
	`)

	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// GetUserByEmail возвращант пользователя models.User, используя email
func (userDB *UserDataBase) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "storage.user.GetUserByEmail"

	stmt, err := userDB.db.Prepare(`
		SELECT 
			id,
			email, 
			password_hash, 
			full_name, 
			created_at, 
			updated_at 
		FROM users WHERE email = $1
	`)

	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// GetListUsers возвращает список всех пользователей
func (userDB *UserDataBase) GetListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	const op = "storage.user.GetListUsers"

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
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

// UpdateUser обновляет email и/или fullName пользователя models.User, используя userID
func (userDB *UserDataBase) UpdateUserEmail(ctx context.Context, userID []uint8, email string) error {
	const op = "storage.user.UpdateUserEmail"

	stmt, err := userDB.db.Prepare("UPDATE users SET email = $1, updated_at = NOW() WHERE id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	result, err := stmt.ExecContext(ctx, email, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return ssoerrors.ErrNotFound
	}

	return nil
}

// UpdateUser обновляет email и/или fullName пользователя models.User, используя userID
func (userDB *UserDataBase) UpdateUserName(ctx context.Context, userID []uint8, fullName string) error {
	const op = "storage.user.UpdateUserName"

	stmt, err := userDB.db.Prepare("UPDATE users SET full_name = $1, updated_at = NOW() WHERE id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	result, err := stmt.ExecContext(ctx, fullName, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return ssoerrors.ErrNotFound
	}

	return nil
}

// UpdateUser обновляет email и/или fullName пользователя models.User, используя userID
func (userDB *UserDataBase) UpdateUserPassword(ctx context.Context, userID []uint8, newPasswordHash string) error {
	const op = "storage.user.UpdateUserPassword"

	stmt, err := userDB.db.Prepare("UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	result, err := stmt.ExecContext(ctx, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return ssoerrors.ErrNotFound
	}

	return nil
}

// DeleteUser "мягкое" удаление пользователя по userID
func (userDB *UserDataBase) DeleteUser(ctx context.Context, userID []uint8) error {
	const op = "storage.user.DeleteUser"

	// Используем транзакцию, так как нужно удалить связанные данные
	tx, err := userDB.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	// Удаляем сессии пользователя
	if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Удаляем связи с ролями
	if _, err := tx.ExecContext(ctx, "DELETE FROM user_roles WHERE user_id = $1", userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// "Мягкое" удаление пользователя (помечаем deleted_at)
	if _, err := tx.ExecContext(ctx,
		"UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		userID,
	); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit()
}
