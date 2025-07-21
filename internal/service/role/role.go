package role

import (
	"context"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RoleManage interface {
	CreateRole(ctx context.Context, name string, permissions []string, description string) (uuid.UUID, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	GetRoleByID(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	GetRoleByName(ctx context.Context, roleName string) (*models.Role, error)
	ListRoles(ctx context.Context, limit, offset int) (*[]models.Role, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, description string) error
}

type UserRoleManage interface {
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RevokeRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) (*[]models.Role, error)
}

type RoleService struct {
	logger         *zap.Logger
	roleManage     RoleManage
	userRoleManage UserRoleManage
}

func NewRoleService(
	logger *zap.Logger,
	roleManage RoleManage,
	userRoleManage UserRoleManage,
) *RoleService {
	return &RoleService{
		logger:         logger.With(zap.String("component", "sso_service")),
		roleManage:     roleManage,
		userRoleManage: userRoleManage,
	}
}

// CreateRole создает новую роль
func (s *RoleService) CreateRole(
	ctx context.Context,
	name string,
	permissions []string,
	description string,
) (uuid.UUID, error) {
	const op = "service.role.CreateRole"

	s.logger.With(
		zap.String("op", op),
		zap.String("name role", name),
	)

	s.logger.Info("creating new role")

	// Валидация
	if name == "" {
		s.logger.Warn("invalid name", zap.Error(ssoerrors.ErrRoleName))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleName)
	}

	// Создание роли
	id, err := s.roleManage.CreateRole(ctx, name, permissions, description)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleExists) {
			s.logger.Warn("role exists", zap.String("role name", name))

			return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleExists)
		}
		s.logger.Error("creating role",
			zap.String("role name", name),
			zap.Strings("role name", permissions),
			zap.Error(err),
		)
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully created role",
		zap.String("role_id", id.String()),
	)
	return id, nil
}

// DeleteRole удаляет роль
func (s *RoleService) DeleteRole(
	ctx context.Context,
	roleID uuid.UUID,
) error {
	const op = "service.role.DeleteRole"

	s.logger.With(
		zap.String("op", op),
		zap.String("role id", roleID.String()),
	)

	s.logger.Info("creating new role")

	// Удаление роли
	err := s.roleManage.DeleteRole(ctx, roleID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleHasUsers) {
			s.logger.Warn("role has user",
				zap.String("role id", roleID.String()),
				zap.Error(ssoerrors.ErrRoleHasUsers),
			)

			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleHasUsers)
		}
		if errors.Is(err, ssoerrors.ErrRoleNotFound) {
			s.logger.Warn("role not found",
				zap.String("role id", roleID.String()),
				zap.Error(ssoerrors.ErrRoleNotFound),
			)

			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleNotFound)
		}
		s.logger.Error("deleting role",
			zap.String("role id", roleID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully deleted role",
		zap.String("role_id", roleID.String()),
	)
	return nil
}

// GetRoleByID возвращает роль(models.Role) по ID
func (s *RoleService) GetRoleByID(
	ctx context.Context,
	roleID uuid.UUID,
) (*models.Role, error) {
	const op = "service.role.GetRoleByID"

	s.logger.With(
		zap.String("op", op),
		zap.String("role id", roleID.String()),
	)

	s.logger.Info("getting role")

	// Получение роли
	role, err := s.roleManage.GetRoleByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleNotFound) {
			s.logger.Warn("role not found",
				zap.String("role id", roleID.String()),
				zap.Error(ssoerrors.ErrRoleNotFound),
			)

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleNotFound)
		}
		s.logger.Error("getting role",
			zap.String("role id", roleID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully get role",
		zap.String("role_id", roleID.String()),
	)
	return role, nil
}

// GetRoleByName возвращает роль(models.Role) по name
func (s *RoleService) GetRoleByName(
	ctx context.Context,
	roleName string,
) (*models.Role, error) {
	const op = "service.role.GetRoleByName"

	s.logger.With(
		zap.String("op", op),
		zap.String("role name", roleName),
	)

	s.logger.Info("getting role")

	// Получение роли
	role, err := s.roleManage.GetRoleByName(ctx, roleName)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleNotFound) {
			s.logger.Warn("role not found",
				zap.String("role name", roleName),
				zap.Error(ssoerrors.ErrRoleNotFound),
			)

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrRoleNotFound)
		}
		s.logger.Error("getting role",
			zap.String("role name", roleName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully get role",
		zap.String("role name", roleName),
	)
	return role, nil
}

// ListRoles все существующие роли
func (s *RoleService) ListRoles(
	ctx context.Context,
	limit, offset int,
) (*[]models.Role, error) {
	const op = "service.role.ListRoles"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting list roles")

	// Получение роли
	roles, err := s.roleManage.ListRoles(ctx, limit, offset)
	if err != nil {
		s.logger.Error("getting roles", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully get list roles")
	return roles, nil
}

// UpdateRole обновляет роль
func (s *RoleService) UpdateRole(
	ctx context.Context,
	roleID uuid.UUID,
	description string,
) error {
	const op = "service.role.UpdateRole"

	s.logger.With(
		zap.String("op", op),
		zap.String("role id", roleID.String()),
	)

	s.logger.Info("updating role")

	// Создание роли
	err := s.roleManage.UpdateRole(ctx, roleID, description)
	if err != nil {
		s.logger.Error("creating role",
			zap.String("role id", roleID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully update role",
		zap.String("role id", roleID.String()),
	)
	return nil
}

// AssignRoleToUser назначает роль пользователю с проверками
func (s *RoleService) AssignRoleToUser(
	ctx context.Context,
	userID, roleID uuid.UUID,
) error {
	const op = "service.role.AssignRoleToUser"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("assigning role")

	// Проверяем существование роли
	_, err := s.roleManage.GetRoleByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleNotFound) {
			s.logger.Warn("getting role", zap.String("role id", roleID.String()), zap.Error(err))

			return ssoerrors.ErrRoleNotFound
		}
		s.logger.Warn("getting role", zap.String("role id", roleID.String()), zap.Error(err))

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Назначаем роль
	err = s.userRoleManage.AssignRoleToUser(ctx, userID, roleID)
	if err != nil {
		s.logger.Error("assigning role",
			zap.String("role id", roleID.String()),
			zap.String("user id", userID.String()),
			zap.Error(err),
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Debug("Successfully assign role to user",
		zap.String("role id", roleID.String()),
		zap.String("user id", userID.String()),
	)
	return nil
}

// RevokeRoleFromUser отзывает роль у пользователя
func (s *RoleService) RevokeRoleFromUser(
	ctx context.Context,
	userID, roleID uuid.UUID,
) error {
	const op = "service.role.RevokeRoleFromUser"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("revoking role user")

	err := s.userRoleManage.RevokeRoleFromUser(ctx, userID, roleID)
	if err != nil {
		s.logger.Error("revoking role",
			zap.String("role id", roleID.String()),
			zap.String("user id", userID.String()),
			zap.Error(err),
		)

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully revoke role from user",
		zap.String("role id", roleID.String()),
		zap.String("user id", userID.String()),
	)
	return nil
}

// GetUserRoles возвращает все роли пользователя
func (s *RoleService) GetUserRoles(
	ctx context.Context,
	userID uuid.UUID,
) (*[]models.Role, error) {
	const op = "service.role.GetUserRoles"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting user roles")

	roles, err := s.userRoleManage.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.Warn("cant get",
			zap.String("user id", userID.String()),
			zap.Error(err),
		)

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully getting user roles",
		zap.String("user id", userID.String()),
	)
	return roles, nil
}
