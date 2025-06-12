package role

import (
	"context"
	"errors"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	"go.uber.org/zap"
)

type RoleService struct {
	logger            *zap.Logger
	roleManage        RoleManage
	permissionsManage PermissionsManage
}

type RoleManage interface {
	CreateRole(ctx context.Context, name string, permissions []string) (int64, error)
	GetRole(ctx context.Context, roleID int64) (models.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID int64) error
	RevokeRoleFromUser(ctx context.Context, userID, roleID int64) error
}

type PermissionsManage interface {
	CheckUserPermission(ctx context.Context, userID int64, permission string) (bool, error)
	ListUserPermissions(ctx context.Context, userID int64) ([]string, error)
}

func NewRoleService(
	logger *zap.Logger,
	roleManage RoleManage,
	permissionsManage PermissionsManage,
) *RoleService {
	return &RoleService{
		logger:            logger,
		roleManage:        roleManage,
		permissionsManage: permissionsManage,
	}
}

// CreateRole создает новую роль с проверкой уникальности имени
func (s *RoleService) CreateRole(
	ctx context.Context,
	name string,
	permissions []string,
) (models.Role, error) {
	//TO DO: use logger

	// Валидация
	if name == "" {
		return models.Role{}, errors.New("role name cannot be empty")
	}

	// Создание роли
	id, err := s.roleManage.CreateRole(ctx, name, permissions)
	if err != nil {
		return models.Role{}, err // ErrRoleExists или другая ошибка из репозитория
	}

	// Возвращаем созданную роль
	role, err := s.roleManage.GetRole(ctx, id)
	if err != nil {
		return models.Role{}, err
	}

	return role, nil
}

// AssignRole назначает роль пользователю с проверками
func (s *RoleService) AssignRole(
	ctx context.Context,
	userID, roleID int64,
) error {
	//TO DO: use logger

	// Проверяем существование роли
	_, err := s.roleManage.GetRole(ctx, roleID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleNotFound) {
			return ssoerrors.ErrRoleNotFound
		}
		return err
	}

	// Назначаем роль
	return s.roleManage.AssignRoleToUser(ctx, userID, roleID)
}

// CheckPermission проверяет право пользователя
func (s *RoleService) CheckPermission(
	ctx context.Context,
	userID int64, permission string,
) (bool, error) {
	//TO DO: use logger

	return s.permissionsManage.CheckUserPermission(ctx, userID, permission)
}

// GetUserPermissions возвращает все права пользователя
func (s *RoleService) GetUserPermissions(
	ctx context.Context,
	userID int64,
) ([]string, error) {
	//TO DO: use logger

	return s.permissionsManage.ListUserPermissions(ctx, userID)
}

// RevokeRole отзывает роль у пользователя
func (s *RoleService) RevokeRole(
	ctx context.Context,
	userID, roleID int64,
) error {
	//TO DO: use logger

	return s.roleManage.RevokeRoleFromUser(ctx, userID, roleID)
}
