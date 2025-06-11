package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso-service/internal/models"
	"sso-service/internal/permissions"

	"github.com/lib/pq"
)

var (
	ErrRoleExists      = errors.New("role already exists")
	ErrRoleNotFound    = errors.New("role not found")
	ErrRoleNotAssigned = errors.New("role is not assigned to the user")
)

type Role interface {
	// Роли
	InitRoles() error
	CreateRole(ctx context.Context, name string, permissions []string) error
	GetRole(ctx context.Context, roleID string) (models.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RevokeRoleFromUser(ctx context.Context, userID, roleID string) error

	// Права
	CheckUserPermission(ctx context.Context, userID, permission string) (bool, error)
	ListUserPermissions(ctx context.Context, userID string) ([]string, error)
}

type RoleDataBase struct {
	db *Database
	// TO DO: Redis Cache
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
) error {
	_, err := RoleDB.db.ExecContext(ctx, `
        INSERT INTO roles (id, name, permissions)
        VALUES (gen_random_uuid(), $1, $2)
    `, name, pq.Array(permissions)) // pq.Array для массивов PostgreSQL

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return fmt.Errorf("%w", ErrRoleExists)
		}
		return fmt.Errorf("%w", err)
	}

	return nil
}

// GetRole возвращает роль по ID
func (RoleDB *RoleDataBase) GetRole(ctx context.Context, roleID string) (models.Role, error) {

	var role models.Role
	err := RoleDB.db.GetContext(ctx, &role, `
        SELECT id, name, permissions, created_at
        FROM roles
        WHERE id = $1
    `, roleID)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Role{}, fmt.Errorf("%w", ErrRoleNotFound)
	}

	if err != nil {
		return models.Role{}, fmt.Errorf("%w", err)
	}

	return role, nil
}

// AssignRoleToUser назначает роль пользователю
func (RoleDB *RoleDataBase) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
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
func (RoleDB *RoleDataBase) RevokeRoleFromUser(ctx context.Context, userID, roleID string) error {
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
		return fmt.Errorf("%w", ErrRoleNotAssigned)
	}

	return nil
}

// CheckUserPermission проверяет наличие права у пользователя
func (RoleDB *RoleDataBase) CheckUserPermission(ctx context.Context, userID, permission string) (bool, error) {
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
func (RoleDB *RoleDataBase) ListUserPermissions(ctx context.Context, userID string) ([]string, error) {
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
