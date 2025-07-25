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
	CreateEntity(ctx context.Context, entityType string) error
	CreateEntityWithID(ctx context.Context, id uuid.UUID, entityType string) error
	DeleteEntity(ctx context.Context, id uuid.UUID) error
	CreateRelation(ctx context.Context, sourceID, targetID uuid.UUID, relationType string) error
	DeleteRelation(ctx context.Context, sourceID, targetID uuid.UUID, relationType string) error
	AddPermission(ctx context.Context, name, description string) (uuid.UUID, error)
	AssignPermission(ctx context.Context, permissionID uuid.UUID, relationType string) error
	RevokePermission(ctx context.Context, permissionID uuid.UUID, relationType string) error
}

type RoleInfoManage interface {
	CheckPermission(ctx context.Context, subjectID, objectID uuid.UUID, permissionName string) (bool, error)
	GetPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	GetEntityID(ctx context.Context, entityType string) (uuid.UUID, error)
	GetAllPermissions(ctx context.Context) (*[]models.Permission, error)
	GetUserRelations(ctx context.Context, userID uuid.UUID) (*[]models.Relation, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) (*[]models.Permission, error)
	GetPermissionsForRelationType(ctx context.Context, relationType string) (*[]models.Permission, error)
	GetEntityRelations(ctx context.Context, entityID uuid.UUID) (*[]models.Relation, error)
}

type RoleService struct {
	logger         *zap.Logger
	roleManage     RoleManage
	roleInfoManage RoleInfoManage
}

func NewRoleService(
	logger *zap.Logger,
	roleManage RoleManage,
	roleInfoManage RoleInfoManage,
) *RoleService {
	return &RoleService{
		logger:         logger.With(zap.String("component", "sso_service")),
		roleManage:     roleManage,
		roleInfoManage: roleInfoManage,
	}
}

// CreateEntity создает новую сущность (роль/пользователя/группу)
func (s *RoleService) CreateEntity(
	ctx context.Context,
	id uuid.UUID,
	entityType string,
) error {
	const op = "service.role.CreateEntity"

	logger := s.logger.With(
		zap.String("op", op),
	)

	logger.Info("creating new entity")

	// Создание сущности
	var err error
	switch {
	case id == uuid.Nil:
		err = s.roleManage.CreateEntity(ctx, entityType)
	case id != uuid.Nil:
		err = s.roleManage.CreateEntityWithID(ctx, id, entityType)
	}
	if err != nil {
		if errors.Is(err, ssoerrors.ErrEntityExists) {
			logger.Warn("entity exists",
				zap.Error(ssoerrors.ErrEntityExists),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrEntityExists)
		}
		logger.Error("creating entity",
			zap.String("entity_type", entityType),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully created entity",
		zap.String("entity_type", entityType),
	)
	return nil
}

// DeleteEntity удаляет сущность
func (s *RoleService) DeleteEntity(
	ctx context.Context,
	id uuid.UUID,
) error {
	const op = "service.role.DeleteEntity"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("entity_id", id.String()),
	)

	logger.Info("deleting entity")

	err := s.roleManage.DeleteEntity(ctx, id)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrEntityNotFound) {
			logger.Warn("entity not found",
				zap.Error(ssoerrors.ErrEntityNotFound),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrEntityNotFound)
		}
		logger.Error("deleting entity",
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully deleted entity")
	return nil
}

// CreateRelation создает связь между сущностями
func (s *RoleService) CreateRelation(
	ctx context.Context,
	sourceID, targetID uuid.UUID,
	relationType string,
) error {
	const op = "service.role.CreateRelation"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("source_id", sourceID.String()),
		zap.String("target_id", targetID.String()),
		zap.String("relation_type", relationType),
	)

	logger.Info("creating relation")

	err := s.roleManage.CreateRelation(ctx, sourceID, targetID, relationType)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRelationExists) {
			logger.Warn("relation exists",
				zap.Error(ssoerrors.ErrRelationExists),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRelationExists)
		}
		logger.Error("creating relation",
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully created relation")
	return nil
}

// DeleteRelation удаляет связь между сущностями
func (s *RoleService) DeleteRelation(
	ctx context.Context,
	sourceID, targetID uuid.UUID,
	relationType string,
) error {
	const op = "service.role.DeleteRelation"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("source_id", sourceID.String()),
		zap.String("target_id", targetID.String()),
		zap.String("relation_type", relationType),
	)

	logger.Info("deleting relation")

	err := s.roleManage.DeleteRelation(ctx, sourceID, targetID, relationType)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRelationNotFound) {
			logger.Warn("relation not found",
				zap.Error(ssoerrors.ErrRelationNotFound),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRelationNotFound)
		}
		logger.Error("deleting relation",
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully deleted relation")
	return nil
}

// AddPermission добавляет новое разрешение в систему
func (s *RoleService) AddPermission(
	ctx context.Context,
	name, description string,
) (uuid.UUID, error) {
	const op = "service.role.AddPermission"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("permission_name", name),
	)

	logger.Info("adding new permission")

	id, err := s.roleManage.AddPermission(ctx, name, description)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrPermissionExists) {
			logger.Warn("permission exists",
				zap.Error(ssoerrors.ErrPermissionExists),
			)
			return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionExists)
		}
		logger.Error("adding permission",
			zap.Error(err),
		)
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully added permission",
		zap.String("permission_id", id.String()),
	)
	return id, nil
}

// AssignPermission назначает разрешение для типа связи
func (s *RoleService) AssignPermission(
	ctx context.Context,
	permissionID uuid.UUID,
	relationType string,
) error {
	const op = "service.role.AssignPermission"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("permission_id", permissionID.String()),
		zap.String("relation_type", relationType),
	)

	logger.Info("assigning permission to relation")

	err := s.roleManage.AssignPermission(ctx, permissionID, relationType)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrPermissionNotFound) {
			logger.Warn("permission not found",
				zap.Error(ssoerrors.ErrPermissionNotFound),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}

		logger.Error("assigning permission",
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully assigned permission to relation")
	return nil
}

// RevokePermission отзывает разрешение у типа связи
func (s *RoleService) RevokePermission(
	ctx context.Context,
	permissionID uuid.UUID,
	relationType string,
) error {
	const op = "service.role.RevokePermission"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("permission_id", permissionID.String()),
		zap.String("relation_type", relationType),
	)

	logger.Info("revoking permission from relation")

	err := s.roleManage.RevokePermission(ctx, permissionID, relationType)
	if err != nil {
		switch {
		case errors.Is(err, ssoerrors.ErrPermissionNotFound):
			logger.Warn("permission not found",
				zap.Error(ssoerrors.ErrPermissionNotFound),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)

		case errors.Is(err, ssoerrors.ErrPermissionNotAssigned):
			logger.Warn("permission was not assigned",
				zap.Error(ssoerrors.ErrPermissionNotAssigned),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotAssigned)

		default:
			logger.Error("revoking permission",
				zap.Error(err),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
		}
	}

	logger.Debug("Successfully revoked permission from relation")
	return nil
}

// CheckPermission проверяет наличие разрешения у субъекта для объекта
func (s *RoleService) CheckPermission(
	ctx context.Context,
	subjectID, objectID uuid.UUID,
	permissionName string,
) (bool, error) {
	const op = "service.role.CheckPermission"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("subject_id", subjectID.String()),
		zap.String("object_id", objectID.String()),
		zap.String("permission_id", permissionName),
	)

	logger.Info("checking permission")

	hasPermission, err := s.roleInfoManage.CheckPermission(ctx, subjectID, objectID, permissionName)
	if err != nil {
		logger.Error("checking permission",
			zap.Error(err),
		)
		return false, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Permission check result",
		zap.Bool("has_permission", hasPermission),
	)
	return hasPermission, nil
}

// GetPermissionByName возвращает разрешение(models.Permission) по name
func (s *RoleService) GetPermissionByName(
	ctx context.Context,
	name string,
) (*models.Permission, error) {
	const op = "service.role.GetPermissionByName"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("permission name", name),
	)

	logger.Info("getting permission")

	// Получение роли
	permission, err := s.roleInfoManage.GetPermissionByName(ctx, name)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrPermissionNotFound) {
			logger.Warn("permission not found",
				zap.Error(ssoerrors.ErrPermissionNotFound),
			)
			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}
		logger.Error("getting permission",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully get permission")
	return permission, nil
}

// GetAllPermissions возвращает все разрешения в системе
func (s *RoleService) GetAllPermissions(
	ctx context.Context,
) (*[]models.Permission, error) {
	const op = "service.role.GetAllPermissions"

	logger := s.logger.With(
		zap.String("op", op),
	)

	logger.Info("getting all permissions")

	permissions, err := s.roleInfoManage.GetAllPermissions(ctx)
	if err != nil {
		logger.Error("getting all permissions",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully retrieved all permissions",
		zap.Int("count", len(*permissions)),
	)
	return permissions, nil
}

// GetEntityID возвращает ID сущности по её типу
func (s *RoleService) GetEntityID(
	ctx context.Context,
	entityType string,
) (uuid.UUID, error) {
	const op = "service.role.GetEntityID"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("entity_type", entityType),
	)

	logger.Info("getting entity ID by type")

	// Получаем ID из storage слоя
	id, err := s.roleInfoManage.GetEntityID(ctx, entityType)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrEntityNotFound) {
			logger.Warn("entity not found",
				zap.Error(ssoerrors.ErrEntityNotFound),
			)
			return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrEntityNotFound)
		}

		logger.Error("failed to get entity ID",
			zap.Error(err),
		)
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("successfully got entity ID",
		zap.String("entity_id", id.String()),
	)
	return id, nil
}

// GetUserRelations возвращает все связи пользователя
func (s *RoleService) GetUserRelations(
	ctx context.Context,
	userID uuid.UUID,
) (*[]models.Relation, error) {
	const op = "service.role.GetUserRelations"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("user_id", userID.String()),
	)

	logger.Info("getting user relations")

	relations, err := s.roleInfoManage.GetUserRelations(ctx, userID)
	if err != nil {
		logger.Error("getting user relations",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully retrieved user relations",
		zap.Int("count", len(*relations)),
	)
	return relations, nil
}

// GetUserPermissions возвращает все разрешения пользователя
func (s *RoleService) GetUserPermissions(
	ctx context.Context,
	userID uuid.UUID,
) (*[]models.Permission, error) {
	const op = "service.role.GetUserPermissions"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("user_id", userID.String()),
	)

	logger.Info("getting user permissions")

	permissions, err := s.roleInfoManage.GetUserPermissions(ctx, userID)
	if err != nil {
		logger.Error("getting user permissions",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully retrieved user permissions",
		zap.Int("count", len(*permissions)),
	)
	return permissions, nil
}

// GetPermissionsForRelationType возвращает разрешения для типа связи
func (s *RoleService) GetPermissionsForRelationType(
	ctx context.Context,
	relationType string,
) (*[]models.Permission, error) {
	const op = "service.role.GetPermissionsForRelationType"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("relation_type", relationType),
	)

	logger.Info("getting permissions for relation type")

	permissions, err := s.roleInfoManage.GetPermissionsForRelationType(ctx, relationType)
	if err != nil {
		logger.Error("getting permissions for relation type",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully retrieved permissions for relation type",
		zap.Int("count", len(*permissions)),
	)
	return permissions, nil
}

// GetEntityRelations возвращает все связи сущности
func (s *RoleService) GetEntityRelations(
	ctx context.Context,
	entityID uuid.UUID,
) (*[]models.Relation, error) {
	const op = "service.role.GetEntityRelations"

	logger := s.logger.With(
		zap.String("op", op),
		zap.String("entity_id", entityID.String()),
	)

	logger.Info("getting entity relations")

	relations, err := s.roleInfoManage.GetEntityRelations(ctx, entityID)
	if err != nil {
		logger.Error("getting entity relations",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully retrieved entity relations",
		zap.Int("count", len(*relations)),
	)
	return relations, nil
}
