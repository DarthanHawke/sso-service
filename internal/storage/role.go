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

const (
	errIdRole = -1
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

// CreateRole создает новую роль с разрешениями
func (RoleDB *RoleDataBase) CreateRole(
	ctx context.Context,
	name string,
	permissions []string,
) (int64, error) {
	const op = "storage.user.CreateRole"

	stmt, err := RoleDB.db.Prepare(`
        INSERT INTO roles (name, permissions)
        VALUES ($1, $2)
        RETURNING id
    `)
	if err != nil {
		return errIdRole, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, name, pq.Array(permissions))
	var id int64
	err = row.Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // 23505 = unique_violation
			return errIdRole, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleExists)
		}

		return errIdRole, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// GetRole возвращает роль по ID
func (RoleDB *RoleDataBase) GetRole(ctx context.Context, roleID int64) (models.Role, error) {
	const op = "storage.user.GetRole"

	var role models.Role
	var perms pq.StringArray
	err := RoleDB.db.QueryRowxContext(ctx, `
        SELECT id, name, permissions
        FROM roles
        WHERE id = $1
    `, roleID).Scan(&role.ID, &role.Name, &perms)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Role{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleNotFound)
		}
		return models.Role{}, fmt.Errorf("%s: %w", op, err)
	}

	role.Permissions = []string(perms)
	return role, nil
}

// AssignRoleToUser назначает роль пользователю
func (RoleDB *RoleDataBase) AssignRoleToUser(ctx context.Context, userID []uint8, roleID int64) error {
	const op = "storage.user.AssignRoleToUser"

	_, err := RoleDB.db.ExecContext(ctx, `
        INSERT INTO user_roles (user_id, role_id)
        VALUES ($1, $2)
        ON CONFLICT (user_id, role_id) DO NOTHING
    `, userID, roleID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// RevokeRoleFromUser отзывает роль у пользователя
func (RoleDB *RoleDataBase) RevokeRoleFromUser(ctx context.Context, userID []uint8, roleID int64) error {
	const op = "storage.user.RevokeRoleFromUser"

	result, err := RoleDB.db.ExecContext(ctx, `
        DELETE FROM user_roles
        WHERE user_id = $1 AND role_id = $2
    `, userID, roleID)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleNotAssigned)
	}

	return nil
}

// CheckUserPermission проверяет наличие права у пользователя
func (RoleDB *RoleDataBase) CheckUserPermission(
	ctx context.Context,
	userID []uint8,
	permission string,
) (bool, error) {
	const op = "storage.user.CheckUserPermission"

	var exists bool
	err := RoleDB.db.GetContext(ctx, &exists, `
        SELECT EXISTS (
            SELECT 1 FROM user_roles ur
            JOIN roles r ON ur.role_id = r.id
            WHERE ur.user_id = $1 AND $2 = ANY(r.permissions)
        )
    `, userID, permission)

	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

// ListUserPermissions возвращает все права пользователя
func (RoleDB *RoleDataBase) ListUserPermissions(ctx context.Context, userID []uint8) ([]string, error) {
	const op = "storage.user.ListUserPermissions"

	var permissions []string
	err := RoleDB.db.SelectContext(ctx, &permissions, `
        SELECT DISTINCT unnest(r.permissions)
        FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        WHERE ur.user_id = $1
    `, userID)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return permissions, nil
}
