package permission

import (
	"context"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PermissionsManage interface {
	CreatePermission(ctx context.Context, code, description string) (uuid.UUID, error)
	DeletePermission(ctx context.Context, permID uuid.UUID) error
	GetPermissionByID(ctx context.Context, permID uuid.UUID) (*models.Permission, error)
	GetPermissionByCode(ctx context.Context, code string) (*models.Permission, error)
	ListPermissions(ctx context.Context, limit, offset int) (*[]models.Permission, error)
	UpdatePermission(ctx context.Context, permID uuid.UUID, description string) error
}

type PermRoleManage interface {
	AddPermissionToRole(ctx context.Context, roleID, permID uuid.UUID) error
	RevokePermissionFromRole(ctx context.Context, roleID, permID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) (*[]models.Permission, error)
	HasRolePermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) (*[]models.Permission, error)
	HasUserPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

var (
	nilBool = false
)

type PermissionService struct {
	logger            *zap.Logger
	permissionsManage PermissionsManage
	permRoleManage    PermRoleManage
}

func NewPermissionService(
	logger *zap.Logger,
	permissionsManage PermissionsManage,
	permRoleManage PermRoleManage,
) *PermissionService {
	return &PermissionService{
		logger:            logger.With(zap.String("component", "sso_service")),
		permissionsManage: permissionsManage,
		permRoleManage:    permRoleManage,
	}
}

// CreatePermission создает новое разрешение
func (s *PermissionService) CreatePermission(
	ctx context.Context,
	code, description string,
) (uuid.UUID, error) {
	const op = "service.role.CreatePermission"

	s.logger.With(
		zap.String("op", op),
		zap.String("permission code", code),
	)

	s.logger.Info("creating new permission")

	// Валидация
	if code == "" {
		s.logger.Warn("invalid code", zap.Error(ssoerrors.ErrPermissionName))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionName)
	}

	// Создание роли
	id, err := s.permissionsManage.CreatePermission(ctx, code, description)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRoleExists) {
			s.logger.Warn("permission exists", zap.String("permission code", code))

			return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionExists)
		}
		s.logger.Error("creating permission",
			zap.String("permission code", code),

			zap.Error(err),
		)
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully created permission",
		zap.String("permission_id", id.String()),
	)
	return id, nil
}

// DeletePermission удаляет разрешение
func (s *PermissionService) DeletePermission(
	ctx context.Context,
	permID uuid.UUID,
) error {
	const op = "service.role.DeletePermission"

	s.logger.With(
		zap.String("op", op),
		zap.String("permission id", permID.String()),
	)

	s.logger.Info("creating new role")

	// Удаление роли
	err := s.permissionsManage.DeletePermission(ctx, permID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrPermissionHasRoles) {
			s.logger.Warn("permission has role",
				zap.String("permission id", permID.String()),
				zap.Error(ssoerrors.ErrPermissionHasRoles),
			)

			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionHasRoles)
		}
		if errors.Is(err, ssoerrors.ErrPermissionNotFound) {
			s.logger.Warn("permission not found",
				zap.String("permission id", permID.String()),
				zap.Error(ssoerrors.ErrPermissionNotFound),
			)

			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}
		s.logger.Error("deleting permission",
			zap.String("permission id", permID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully deleted permission",
		zap.String("permission id", permID.String()),
	)
	return nil
}

// GetPermissionByID возвращает разрешение(models.Permission) по ID
func (s *PermissionService) GetPermissionByID(
	ctx context.Context,
	permID uuid.UUID,
) (*models.Permission, error) {
	const op = "service.role.GetPermissionByID"

	s.logger.With(
		zap.String("op", op),
		zap.String("permission id", permID.String()),
	)

	s.logger.Info("getting permission")

	// Получение роли
	permission, err := s.permissionsManage.GetPermissionByID(ctx, permID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrPermissionNotFound) {
			s.logger.Warn("permission not found",
				zap.String("permission id", permID.String()),
				zap.Error(ssoerrors.ErrPermissionNotFound),
			)

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}
		s.logger.Error("getting permission",
			zap.String("permission id", permID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully get permission",
		zap.String("permission id", permID.String()),
	)
	return permission, nil
}

// GetPermissionByID возвращает разрешение(models.Permission) по code
func (s *PermissionService) GetPermissionByCode(
	ctx context.Context,
	code string,
) (*models.Permission, error) {
	const op = "service.role.GetPermissionByID"

	s.logger.With(
		zap.String("op", op),
		zap.String("permission code", code),
	)

	s.logger.Info("getting permission")

	// Получение роли
	permission, err := s.permissionsManage.GetPermissionByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrPermissionNotFound) {
			s.logger.Warn("permission not found",
				zap.String("permission code", code),
				zap.Error(ssoerrors.ErrPermissionNotFound),
			)

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}
		s.logger.Error("getting permission",
			zap.String("permission code", code),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully get permission",
		zap.String("permission code", code),
	)
	return permission, nil
}

// ListPermissions все существующие разрешения
func (s *PermissionService) ListPermissions(
	ctx context.Context,
	limit, offset int,
) (*[]models.Permission, error) {
	const op = "service.role.ListPermissions"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting list permissions")

	// Получение роли
	permissions, err := s.permissionsManage.ListPermissions(ctx, limit, offset)
	if err != nil {
		s.logger.Error("getting permissions", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully get list permissions")
	return permissions, nil
}

// UpdatePermission обновляет роль
func (s *PermissionService) UpdatePermission(
	ctx context.Context,
	permID uuid.UUID,
	description string,
) error {
	const op = "service.role.UpdatePermission"

	s.logger.With(
		zap.String("op", op),
		zap.String("permission id", permID.String()),
	)

	s.logger.Info("updating permission")

	// Создание роли
	err := s.permissionsManage.UpdatePermission(ctx, permID, description)
	if err != nil {
		s.logger.Error("creating permission",
			zap.String("permission id", permID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully update permission",
		zap.String("permission id", permID.String()),
	)
	return nil
}

// AddPermissionToRole назначает разрешение роли
func (s *PermissionService) AddPermissionToRole(
	ctx context.Context,
	roleID, permID uuid.UUID,
) error {
	const op = "service.role.AssignRoleToUser"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("assigning permission")

	// Проверяем существование разрешения
	_, err := s.permissionsManage.GetPermissionByID(ctx, permID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrPermissionNotFound) {
			s.logger.Warn("getting permission", zap.String("permission id", permID.String()), zap.Error(err))

			return ssoerrors.ErrPermissionNotFound
		}
		s.logger.Warn("getting permission", zap.String("permission id", permID.String()), zap.Error(err))

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}
	// Назначаем роль
	err = s.permRoleManage.AddPermissionToRole(ctx, roleID, permID)
	if err != nil {
		s.logger.Error("assigning permission",
			zap.String("role id", roleID.String()),
			zap.String("permission id", permID.String()),
			zap.Error(err),
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Debug("Successfully assign permission to role",
		zap.String("role id", roleID.String()),
		zap.String("permission id", permID.String()),
	)
	return nil
}

// RevokePermissionFromRole отзывает разрешение у роли
func (s *PermissionService) RevokePermissionFromRole(
	ctx context.Context,
	roleID, permID uuid.UUID,
) error {
	const op = "service.role.RevokePermissionFromRole"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("revoking permission role")

	err := s.permRoleManage.RevokePermissionFromRole(ctx, roleID, permID)
	if err != nil {
		s.logger.Error("revoking permission",
			zap.String("role id", roleID.String()),
			zap.String("permission id", permID.String()),
			zap.Error(err),
		)

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully revoke permission from role",
		zap.String("role id", roleID.String()),
		zap.String("permission id", permID.String()),
	)
	return nil
}

// GetRolePermissions возвращает все разрешения роли
func (s *PermissionService) GetRolePermissions(
	ctx context.Context,
	roleID uuid.UUID,
) (*[]models.Permission, error) {
	const op = "service.role.GetRolePermissions"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting role permissions")

	permissions, err := s.permRoleManage.GetRolePermissions(ctx, roleID)
	if err != nil {
		s.logger.Warn("cant get",
			zap.String("role id", roleID.String()),
			zap.Error(err),
		)

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully getting role permissions",
		zap.String("role id", roleID.String()),
	)
	return permissions, nil
}

// HasRolePermission проверяет право у роли
func (s *PermissionService) HasRolePermission(
	ctx context.Context,
	roleID uuid.UUID, permission string,
) (bool, error) {
	const op = "service.role.HasRolePermission"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("checking permission")

	perm, err := s.permRoleManage.HasRolePermission(ctx, roleID, permission)
	if err != nil {
		s.logger.Warn("cant check",
			zap.String("role id", roleID.String()),
			zap.String("permission", permission),
			zap.Error(err),
		)

		return nilBool, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully permissions check",
		zap.String("role id", roleID.String()),
		zap.String("permission", permission),
		zap.Bool("res", perm),
	)
	return perm, nil
}

// GetUserPermissions возвращает все права пользователя
func (s *PermissionService) GetUserPermissions(
	ctx context.Context,
	userID uuid.UUID,
) (*[]models.Permission, error) {
	const op = "service.role.GetUserPermissions"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting permissions user")

	permissions, err := s.permRoleManage.GetUserPermissions(ctx, userID)
	if err != nil {
		s.logger.Warn("cant get",
			zap.String("user id", userID.String()),
			zap.Error(err),
		)

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully getting permissions user",
		zap.String("user id", userID.String()),
	)
	return permissions, nil
}

// HasUserPermission проверяет право пользователя
func (s *PermissionService) HasUserPermission(
	ctx context.Context,
	userID uuid.UUID, permission string,
) (bool, error) {
	const op = "service.role.HasUserPermission"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("checking permission")

	perm, err := s.permRoleManage.HasUserPermission(ctx, userID, permission)
	if err != nil {
		s.logger.Warn("cant check",
			zap.String("user id", userID.String()),
			zap.String("permission", permission),
			zap.Error(err),
		)

		return nilBool, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully permissions check",
		zap.String("user id", userID.String()),
		zap.String("permission", permission),
		zap.Bool("res", perm),
	)
	return perm, nil
}
