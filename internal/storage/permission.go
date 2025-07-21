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

type PermissionDataBase struct {
	db *Database
	// TO DO: Redis Cache
}

func NewPermissionDataBase(db *Database) *PermissionDataBase {
	return &PermissionDataBase{
		db: db,
	}
}

// CreatePermission создает новое разрешение в системе
func (permissionDB *PermissionDataBase) CreatePermission(
	ctx context.Context,
	code, description string,
) (uuid.UUID, error) {
	const op = "storage.role.CreatePermission"

	var permissionID uuid.UUID

	err := permissionDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Проверяем, не существует ли уже разрешение с таким кодом
		var exists bool
		err := tx.GetContext(ctx, &exists,
			"SELECT EXISTS(SELECT 1 FROM permissions WHERE code = $1)", code)
		if err != nil {
			return fmt.Errorf("%s: failed to check permission existence: %w", op, err)
		}

		if exists {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionExists)
		}

		// Создаем новое разрешение
		err = tx.QueryRowxContext(ctx,
			"INSERT INTO permissions (code, description) VALUES ($1, $2) RETURNING id",
			code, description,
		).Scan(&permissionID)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionExists)
			}
			return fmt.Errorf("%s: failed to create permission: %w", op, err)
		}

		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return permissionID, nil
}

// DeletePermission удаляет разрешение
func (permissionDB *PermissionDataBase) DeletePermission(ctx context.Context, permID uuid.UUID) error {
	const op = "storage.role.DeletePermission"

	return permissionDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Проверяем, есть ли роль с этим разрешением
		var roleCount int
		err := tx.GetContext(ctx, &roleCount,
			"SELECT COUNT(*) FROM role_permissions WHERE permission_id = $1", permID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if roleCount > 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionHasRoles)
		}
		result, err := tx.ExecContext(ctx, "DELETE FROM permissions WHERE id = $1", permID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}

		return nil
	})
}

// GetPermissionByID возвращает разрешение по ID
func (permissionDB *PermissionDataBase) GetPermissionByID(
	ctx context.Context,
	permID uuid.UUID,
) (*models.Permission, error) {
	const op = "storage.role.GetPermissionByID"

	var perm models.Permission
	err := permissionDB.db.GetContext(ctx, &perm,
		"SELECT id, code, description FROM permissions WHERE id = $1", permID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &perm, nil
}

// GetPermissionByCode возвращает разрешение по code
func (permissionDB *PermissionDataBase) GetPermissionByCode(
	ctx context.Context,
	code string,
) (*models.Permission, error) {
	const op = "storage.role.GetPermissionByCode"

	var perm models.Permission
	err := permissionDB.db.GetContext(ctx, &perm,
		"SELECT id, code, description FROM permissions WHERE code = $1", code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &perm, nil
}

// ListPermissions возвращает все разрешения с заданными limit и offset
func (permissionDB *PermissionDataBase) ListPermissions(
	ctx context.Context,
	limit, offset int,
) (*[]models.Permission, error) {
	const op = "storage.role.ListPermissions"

	var perms []models.Permission
	err := permissionDB.db.SelectContext(ctx, &perms,
		"SELECT id, code, description FROM permissions ORDER BY code LIMIT $1 OFFSET $2",
		limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &perms, nil
}

// UpdatePermission обновляет описание роли
func (permissionDB *PermissionDataBase) UpdatePermission(
	ctx context.Context,
	permID uuid.UUID,
	description string,
) error {
	const op = "storage.role.UpdateRole"

	return permissionDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		result, err := permissionDB.db.ExecContext(ctx,
			"UPDATE permissions SET description = $1, updated_at = NOW() WHERE id = $2",
			description, permID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}
		return nil
	})
}

// AddPermissionToRole назначает разрешение роли
func (permissionDB *PermissionDataBase) AddPermissionToRole(ctx context.Context, roleID, permID uuid.UUID) error {
	const op = "storage.role.AddPermissionToRole"

	return permissionDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			roleID, permID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})
}

// RevokePermissionFromRole отзывает разрешение у роли
func (permissionDB *PermissionDataBase) RevokePermissionFromRole(ctx context.Context, roleID, permID uuid.UUID) error {
	const op = "storage.role.RevokePermissionFromRole"

	return permissionDB.db.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx,
			"DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2",
			roleID, permID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	})
}

// GetRolePermissions возвращает разрешения роли
func (permissionDB *PermissionDataBase) GetRolePermissions(
	ctx context.Context,
	roleID uuid.UUID,
) (*[]models.Permission, error) {
	const op = "storage.role.GetRolePermissions"

	var perms []models.Permission
	err := permissionDB.db.SelectContext(ctx, &perms, `
		SELECT p.id, p.code, p.description 
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1`, roleID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &perms, nil
}

// HasRolePermission проверяет наличие разрешения у роли
func (permissionDB *PermissionDataBase) HasRolePermission(
	ctx context.Context,
	roleID uuid.UUID,
	permission string,
) (bool, error) {
	const op = "storage.role.HasRolePermission"

	var exists bool
	err := permissionDB.db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM role_permissions rp
			JOIN permissions p ON rp.permission_id = p.id
			WHERE rp.role_id = $1 AND p.code = $2
		)`, roleID, permission)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}

// GetUserPermissions возвращает все права пользователя
func (permissionDB *PermissionDataBase) GetUserPermissions(
	ctx context.Context,
	userID uuid.UUID,
) (*[]models.Permission, error) {
	const op = "storage.role.GetUserPermissions"

	var perms []models.Permission
	err := permissionDB.db.SelectContext(ctx, &perms, `
		SELECT p.id, p.code, p.description 
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1`, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &perms, nil
}

// HasUserPermission проверяет наличие права у пользователя
func (permissionDB *PermissionDataBase) HasUserPermission(
	ctx context.Context,
	userID uuid.UUID,
	permission string,
) (bool, error) {
	const op = "storage.role.HasPermission"

	var exists bool
	err := permissionDB.db.GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			JOIN role_permissions rp ON ur.role_id = rp.role_id
			JOIN permissions p ON rp.permission_id = p.id
			WHERE ur.user_id = $1 AND p.code = $2
		)`, userID, permission)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
}
