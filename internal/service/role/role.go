package role

import (
	"context"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	"go.uber.org/zap"
)

var (
	nilBool = false
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
		logger:            logger.With(zap.String("component", "sso_service")),
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
		s.logger.Warn("invalid name", zap.Error(ssoerrors.ErrRoleName))

		return models.Role{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleName)
	}

	// Создание роли
	id, err := s.roleManage.CreateRole(ctx, name, permissions)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleExists) {
			s.logger.Warn("role exists", zap.String("role name", name))

			return models.Role{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleExists)
		}
		s.logger.Error("creating role",
			zap.String("role name", name),
			zap.Strings("role name", permissions),
			zap.Error(err),
		)
		return models.Role{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Возвращаем созданную роль
	role, err := s.roleManage.GetRole(ctx, id)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleNotFound) {
			s.logger.Warn("getting role", zap.Int64("role id", id), zap.Error(err))

			return models.Role{}, ssoerrors.ErrRoleNotFound
		}
		s.logger.Warn("getting role", zap.Int64("role id", id), zap.Error(err))

		return models.Role{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully created role",
		zap.Int64("role_id", role.ID),
	)
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
			s.logger.Warn("getting role", zap.Int64("role id", roleID), zap.Error(err))

			return ssoerrors.ErrRoleNotFound
		}
		s.logger.Warn("getting role", zap.Int64("role id", roleID), zap.Error(err))

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Назначаем роль
	err = s.roleManage.AssignRoleToUser(ctx, userID, roleID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleNotFound) {
			s.logger.Warn("assigning role",
				zap.Int64("role id", roleID),
				zap.Uint8s("user id", userID),
				zap.Error(err),
			)

			return ssoerrors.ErrRoleNotFound
		}
		s.logger.Error("assigning role",
			zap.Int64("role id", roleID),
			zap.Uint8s("user id", userID),
			zap.Error(err),
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Debug("Successfully assign role to user",
		zap.Int64("role id", roleID),
		zap.Uint8s("user id", userID),
	)
	return nil
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

	perm, err := s.permissionsManage.CheckUserPermission(ctx, userID, permission)
	if err != nil {
		s.logger.Warn("cant check",
			zap.Uint8s("user id", userID),
			zap.String("permission", permission),
			zap.Error(err),
		)

		return nilBool, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully permissions check",
		zap.Uint8s("user id", userID),
		zap.String("permission", permission),
		zap.Bool("res", perm),
	)
	return perm, nil
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

	perms, err := s.permissionsManage.ListUserPermissions(ctx, userID)
	if err != nil {
		s.logger.Warn("cant get",
			zap.Uint8s("user id", userID),
			zap.Error(err),
		)

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully getting permissions user",
		zap.Uint8s("user id", userID),
	)
	return perms, nil
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

	err := s.roleManage.RevokeRoleFromUser(ctx, userID, roleID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleNotAssigned) {
			s.logger.Warn("revoking role",
				zap.Int64("role id", roleID),
				zap.Uint8s("user id", userID),
				zap.Error(err),
			)

			return ssoerrors.ErrRoleNotAssigned
		}
		s.logger.Error("revoking role",
			zap.Int64("role id", roleID),
			zap.Uint8s("user id", userID),
			zap.Error(err),
		)

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully revoke role frome user",
		zap.Int64("role id", roleID),
		zap.Uint8s("user id", userID),
	)
	return nil
}
