package role

import (
	"context"
	"errors"
	"fmt"
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
	AssignRoleToUser(ctx context.Context, userID []uint8, roleID int64) error
	RevokeRoleFromUser(ctx context.Context, userID []uint8, roleID int64) error
}

type PermissionsManage interface {
	CheckUserPermission(ctx context.Context, userID []uint8, permission string) (bool, error)
	ListUserPermissions(ctx context.Context, userID []uint8) ([]string, error)
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
	const op = "service.role.CreateRole"

	s.logger.With(
		zap.String("op", op),
		zap.String("name role", name),
	)

	s.logger.Info("creating new role")

	// Валидация
	if name == "" {
		s.logger.Info("invalid name", zap.Error(ssoerrors.ErrRoleName))

		return models.Role{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleName)
	}

	// Создание роли
	id, err := s.roleManage.CreateRole(ctx, name, permissions)
	if err != nil {
		s.logger.Error("creating role", zap.Error(err))

		return models.Role{}, fmt.Errorf("%s: %w", op, err)
	}

	// Возвращаем созданную роль
	role, err := s.roleManage.GetRole(ctx, id)
	if err != nil {
		s.logger.Warn("getting role", zap.Error(err))

		return models.Role{}, fmt.Errorf("%s: %w", op, err)
	}

	return role, nil
}

// AssignRole назначает роль пользователю с проверками
func (s *RoleService) AssignRole(
	ctx context.Context,
	userID []uint8,
	roleID int64,
) error {
	const op = "service.role.AssignRole"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("assigning role")

	// Проверяем существование роли
	_, err := s.roleManage.GetRole(ctx, roleID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleNotFound) {
			s.logger.Warn("getting role", zap.Error(err))

			return ssoerrors.ErrRoleNotFound
		}
		s.logger.Warn("getting role", zap.Error(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	// Назначаем роль
	return s.roleManage.AssignRoleToUser(ctx, userID, roleID)
}

// CheckPermission проверяет право пользователя
func (s *RoleService) CheckPermission(
	ctx context.Context,
	userID []uint8, permission string,
) (bool, error) {
	const op = "service.role.CheckPermission"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("checking permission")

	return s.permissionsManage.CheckUserPermission(ctx, userID, permission)
}

// GetUserPermissions возвращает все права пользователя
func (s *RoleService) GetUserPermissions(
	ctx context.Context,
	userID []uint8,
) ([]string, error) {
	const op = "service.role.GetUserPermissions"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting permissions user")

	return s.permissionsManage.ListUserPermissions(ctx, userID)
}

// RevokeRole отзывает роль у пользователя
func (s *RoleService) RevokeRole(
	ctx context.Context,
	userID []uint8,
	roleID int64,
) error {
	const op = "service.role.RevokeRole"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("revoking role user")

	return s.roleManage.RevokeRoleFromUser(ctx, userID, roleID)
}
