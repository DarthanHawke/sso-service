package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type RoleDataBase struct {
	db *Database
	// TO DO: Redis Cache
}

func NewRoleDataBase(db *Database) *RoleDataBase {
	return &RoleDataBase{
		db: db,
	}
}

// CreateRole создает новую роль с разрешениями и описанием
func (roleDB *RoleDataBase) CreateRole(
	ctx context.Context,
	name string,
	permissions []string,
	description string,
) (uuid.UUID, error) {
	const op = "storage.role.CreateRole"

	var roleID uuid.UUID

	err := roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Создаем роль
		err := tx.QueryRowxContext(ctx,
			"INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id",
			name, description,
		).Scan(&roleID)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleExists)
			}
			return fmt.Errorf("%s: %w", op, err)
		}

		// Добавляем разрешения
		if len(permissions) > 0 {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO role_permissions (role_id, permission_id)
				SELECT $1, p.id FROM permissions p WHERE p.code = ANY($2)`,
				roleID, pq.Array(permissions),
			)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return roleID, nil
}

// DeleteRole удаляет роль
func (roleDB *RoleDataBase) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	const op = "storage.role.DeleteRole"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Проверяем, есть ли пользователи с этой ролью
		var userCount int
		err := tx.GetContext(ctx, &userCount,
			"SELECT COUNT(*) FROM user_roles WHERE role_id = $1", roleID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if userCount > 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleHasUsers)
		}

		// Удаляем роль
		result, err := tx.ExecContext(ctx, "DELETE FROM roles WHERE id = $1", roleID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleNotFound)
		}

		return nil
	})
}

// GetRoleByID возвращает роль по ID
func (roleDB *RoleDataBase) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*models.Role, error) {
	const op = "storage.role.GetRoleByID"

	var role models.Role
	err := roleDB.db.GetContext(ctx, &role,
		"SELECT id, name, description FROM roles WHERE id = $1", roleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &role, nil
}

// GetRoleByName возвращает роль по Name
func (roleDB *RoleDataBase) GetRoleByName(ctx context.Context, roleName string) (*models.Role, error) {
	const op = "storage.role.GetRoleByName"

	var role models.Role
	err := roleDB.db.GetContext(ctx, &role,
		"SELECT id, name, description FROM roles WHERE name = $1", roleName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &role, nil
}

// ListRoles возвращает рсписок ролей с заданными limit и offset
func (roleDB *RoleDataBase) ListRoles(ctx context.Context, limit, offset int) (*[]models.Role, error) {
	const op = "storage.role.ListRoles"

	var roles []models.Role
	err := roleDB.db.SelectContext(ctx, &roles,
		"SELECT id, name, description FROM roles ORDER BY name LIMIT $1 OFFSET $2",
		limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &roles, nil
}

// UpdateRole обновляет описание роли
func (roleDB *RoleDataBase) UpdateRole(
	ctx context.Context,
	roleID uuid.UUID,
	description string,
) error {
	const op = "storage.role.UpdateRole"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		result, err := roleDB.db.ExecContext(ctx,
			"UPDATE roles SET description = $1, updated_at = NOW() WHERE id = $2",
			description, roleID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleNotFound)
		}
		return nil
	})
}

// AssignRoleToUser назначает роль пользователю
func (roleDB *RoleDataBase) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	const op = "storage.role.AssignRoleToUser"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			userID, roleID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})
}

// RevokeRoleFromUser отзывает роль у пользователя
func (roleDB *RoleDataBase) RevokeRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	const op = "storage.role.RevokeRoleFromUser"

	return roleDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx,
			"DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2",
			userID, roleID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})
}

// GetUserRoles получить роли пользователя
func (roleDB *RoleDataBase) GetUserRoles(ctx context.Context, userID uuid.UUID) (*[]models.Role, error) {
	const op = "storage.role.GetUserRoles"

	var roles []models.Role
	err := roleDB.db.SelectContext(ctx, &roles, `
		SELECT r.id, r.name, r.description 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &roles, nil
}

// HasUserRole проверяет наличие права у пользователя
func (roleDB *RoleDataBase) HasUserRole(
	ctx context.Context,
	userID uuid.UUID,
	roleName string,
) (bool, error) {
	const op = "storage.role.HasUserRole"

	var exists bool
	err := roleDB.db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1 AND r.name = $2
		)`, userID, roleName)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}
