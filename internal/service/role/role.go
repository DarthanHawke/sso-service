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
	CreateRole(ctx context.Context, name string, permissions []string) error
	GetRole(ctx context.Context, roleID string) (models.Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	RevokeRoleFromUser(ctx context.Context, userID, roleID string) error
}

type PermissionsManage interface {
	CheckUserPermission(ctx context.Context, userID, permission string) (bool, error)
	ListUserPermissions(ctx context.Context, userID string) ([]string, error)
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
) (*models.Role, error) {
	//TO DO: use logger

	// Валидация
	if name == "" {
		return nil, errors.New("role name cannot be empty")
	}

	// Создание роли
	err := s.roleManage.CreateRole(ctx, name, permissions)
	if err != nil {
		return nil, err // ErrRoleExists или другая ошибка из репозитория
	}

	// Возвращаем созданную роль
	role, err := s.roleManage.GetRole(ctx, name) // Предполагаем, что GetRole ищет по name
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// AssignRole назначает роль пользователю с проверками
func (s *RoleService) AssignRole(
	ctx context.Context,
	userID, roleID string,
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
	userID, permission string,
) (bool, error) {
	//TO DO: use logger

	return s.permissionsManage.CheckUserPermission(ctx, userID, permission)
}

// GetUserPermissions возвращает все права пользователя
func (s *RoleService) GetUserPermissions(
	ctx context.Context,
	userID string,
) ([]string, error) {
	//TO DO: use logger

	return s.permissionsManage.ListUserPermissions(ctx, userID)
}

// RevokeRole отзывает роль у пользователя
func (s *RoleService) RevokeRole(
	ctx context.Context,
	userID, roleID string,
) error {
	//TO DO: use logger

	return s.roleManage.RevokeRoleFromUser(ctx, userID, roleID)
}
