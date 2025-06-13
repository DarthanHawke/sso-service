package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"
	"sso-service/internal/permissions"

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

// InitRoles создаёт базовые роли
func (RoleDB RoleDataBase) InitRoles() error {
	SuperAdminPerms := permissions.All

	adminPerms := []string{
		permissions.UserRead,
		permissions.UserCreate,
		permissions.UserUpdate,
		permissions.PaymentCreate,
		permissions.AdminAccess,
	}

	userPerms := []string{
		permissions.PaymentCreate,
	}

	banUserPerms := []string{}

	_, err := RoleDB.db.Exec(`
        INSERT INTO roles (id, name, permissions) 
        VALUES 
            ('1', 'superadmin', $1),
            ('2', 'admin', $2),
			('3', 'user', $1),
            ('4', 'ban', $2),
		
        ON CONFLICT DO NOTHING
    `, pq.Array(SuperAdminPerms), pq.Array(adminPerms), pq.Array(userPerms), pq.Array(banUserPerms))

	return err
}

// CreateRole создает новую роль с разрешениями
func (RoleDB *RoleDataBase) CreateRole(
	ctx context.Context,
	name string,
	permissions []string,
) (int64, error) {
	result, err := RoleDB.db.ExecContext(ctx, `
        INSERT INTO roles (id, name, permissions)
        VALUES (gen_random_uuid(), $1, $2)
    `, name, pq.Array(permissions)) // pq.Array для массивов PostgreSQL

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return 0, fmt.Errorf("%w", ssoerrors.ErrRoleExists)
		}
		return 0, fmt.Errorf("%w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}

	return id, nil
}

// GetRole возвращает роль по ID
func (RoleDB *RoleDataBase) GetRole(ctx context.Context, roleID int64) (models.Role, error) {

	var role models.Role
	err := RoleDB.db.GetContext(ctx, &role, `
        SELECT id, name, permissions, created_at
        FROM roles
        WHERE id = $1
    `, roleID)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Role{}, fmt.Errorf("%w", ssoerrors.ErrRoleNotFound)
	}

	if err != nil {
		return models.Role{}, fmt.Errorf("%w", err)
	}

	return role, nil
}

// AssignRoleToUser назначает роль пользователю
func (RoleDB *RoleDataBase) AssignRoleToUser(ctx context.Context, userID, roleID int64) error {
	_, err := RoleDB.db.ExecContext(ctx, `
        INSERT INTO user_roles (user_id, role_id)
        VALUES ($1, $2)
        ON CONFLICT (user_id, role_id) DO NOTHING
    `, userID, roleID)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// RevokeRoleFromUser отзывает роль у пользователя
func (RoleDB *RoleDataBase) RevokeRoleFromUser(ctx context.Context, userID, roleID int64) error {
	result, err := RoleDB.db.ExecContext(ctx, `
        DELETE FROM user_roles
        WHERE user_id = $1 AND role_id = $2
    `, userID, roleID)

	if err != nil {
		return fmt.Errorf("%w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w", ssoerrors.ErrRoleNotAssigned)
	}

	return nil
}

// CheckUserPermission проверяет наличие права у пользователя
func (RoleDB *RoleDataBase) CheckUserPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	var exists bool
	err := RoleDB.db.GetContext(ctx, &exists, `
        SELECT EXISTS (
            SELECT 1 FROM user_roles ur
            JOIN roles r ON ur.role_id = r.id
            WHERE ur.user_id = $1 AND $2 = ANY(r.permissions)
        )
    `, userID, permission)

	if err != nil {
		return false, fmt.Errorf("%w", err)
	}

	return exists, nil
}

// ListUserPermissions возвращает все права пользователя
func (RoleDB *RoleDataBase) ListUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	var permissions []string
	err := RoleDB.db.SelectContext(ctx, &permissions, `
        SELECT DISTINCT unnest(r.permissions)
        FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        WHERE ur.user_id = $1
    `, userID)

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return permissions, nil
}
