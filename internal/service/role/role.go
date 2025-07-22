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
	CreateEntity(ctx context.Context, id uuid.UUID, entityType string) error
	DeleteEntity(ctx context.Context, id uuid.UUID) error
	CreateRelation(ctx context.Context, sourceID, targetID uuid.UUID, relationType string) error
	DeleteRelation(ctx context.Context, sourceID, targetID uuid.UUID, relationType string) error
	AddPermission(ctx context.Context, name, description string) (uuid.UUID, error)
	AssignPermission(ctx context.Context, permissionID uuid.UUID, relationType string) error
}

type RoleInfoManage interface {
	CheckPermission(ctx context.Context, subjectID, objectID, permissionID uuid.UUID) (bool, error)
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

	s.logger.With(
		zap.String("op", op),
		zap.String("entity_id", id.String()),
		zap.String("entity_type", entityType),
	)

	s.logger.Info("creating new entity")

	// Создание сущности
	err := s.roleManage.CreateEntity(ctx, id, entityType)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrEntityExists) {
			s.logger.Warn("entity exists",
				zap.String("entity_id", id.String()),
				zap.Error(ssoerrors.ErrEntityExists),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrEntityExists)
		}
		s.logger.Error("creating entity",
			zap.String("entity_id", id.String()),
			zap.String("entity_type", entityType),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully created entity",
		zap.String("entity_id", id.String()),
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

	s.logger.With(
		zap.String("op", op),
		zap.String("entity_id", id.String()),
	)

	s.logger.Info("deleting entity")

	err := s.roleManage.DeleteEntity(ctx, id)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrEntityNotFound) {
			s.logger.Warn("entity not found",
				zap.String("entity_id", id.String()),
				zap.Error(ssoerrors.ErrEntityNotFound),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrEntityNotFound)
		}
		s.logger.Error("deleting entity",
			zap.String("entity_id", id.String()),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully deleted entity",
		zap.String("entity_id", id.String()),
	)
	return nil
}

// CreateRelation создает связь между сущностями
func (s *RoleService) CreateRelation(
	ctx context.Context,
	sourceID, targetID uuid.UUID,
	relationType string,
) error {
	const op = "service.role.CreateRelation"

	s.logger.With(
		zap.String("op", op),
		zap.String("source_id", sourceID.String()),
		zap.String("target_id", targetID.String()),
		zap.String("relation_type", relationType),
	)

	s.logger.Info("creating relation")

	err := s.roleManage.CreateRelation(ctx, sourceID, targetID, relationType)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRelationExists) {
			s.logger.Warn("relation exists",
				zap.String("source_id", sourceID.String()),
				zap.String("target_id", targetID.String()),
				zap.String("relation_type", relationType),
				zap.Error(ssoerrors.ErrRelationExists),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRelationExists)
		}
		s.logger.Error("creating relation",
			zap.String("source_id", sourceID.String()),
			zap.String("target_id", targetID.String()),
			zap.String("relation_type", relationType),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully created relation",
		zap.String("source_id", sourceID.String()),
		zap.String("target_id", targetID.String()),
		zap.String("relation_type", relationType),
	)
	return nil
}

// DeleteRelation удаляет связь между сущностями
func (s *RoleService) DeleteRelation(
	ctx context.Context,
	sourceID, targetID uuid.UUID,
	relationType string,
) error {
	const op = "service.role.DeleteRelation"

	s.logger.With(
		zap.String("op", op),
		zap.String("source_id", sourceID.String()),
		zap.String("target_id", targetID.String()),
		zap.String("relation_type", relationType),
	)

	s.logger.Info("deleting relation")

	err := s.roleManage.DeleteRelation(ctx, sourceID, targetID, relationType)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrRelationNotFound) {
			s.logger.Warn("relation not found",
				zap.String("source_id", sourceID.String()),
				zap.String("target_id", targetID.String()),
				zap.String("relation_type", relationType),
				zap.Error(ssoerrors.ErrRelationNotFound),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrRelationNotFound)
		}
		s.logger.Error("deleting relation",
			zap.String("source_id", sourceID.String()),
			zap.String("target_id", targetID.String()),
			zap.String("relation_type", relationType),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully deleted relation",
		zap.String("source_id", sourceID.String()),
		zap.String("target_id", targetID.String()),
		zap.String("relation_type", relationType),
	)
	return nil
}

// AddPermission добавляет новое разрешение в систему
func (s *RoleService) AddPermission(
	ctx context.Context,
	name, description string,
) (uuid.UUID, error) {
	const op = "service.role.AddPermission"

	s.logger.With(
		zap.String("op", op),
		zap.String("permission_name", name),
	)

	s.logger.Info("adding new permission")

	id, err := s.roleManage.AddPermission(ctx, name, description)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrPermissionExists) {
			s.logger.Warn("permission exists",
				zap.String("permission_name", name),
				zap.Error(ssoerrors.ErrPermissionExists),
			)
			return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionExists)
		}
		s.logger.Error("adding permission",
			zap.String("permission_name", name),
			zap.Error(err),
		)
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully added permission",
		zap.String("permission_id", id.String()),
		zap.String("permission_name", name),
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

	s.logger.With(
		zap.String("op", op),
		zap.String("permission_id", permissionID.String()),
		zap.String("relation_type", relationType),
	)

	s.logger.Info("assigning permission to relation")

	err := s.roleManage.AssignPermission(ctx, permissionID, relationType)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrPermissionNotFound) {
			s.logger.Warn("permission not found",
				zap.String("permission_id", permissionID.String()),
				zap.Error(ssoerrors.ErrPermissionNotFound),
			)
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPermissionNotFound)
		}

		s.logger.Error("assigning permission",
			zap.String("permission_id", permissionID.String()),
			zap.String("relation_type", relationType),
			zap.Error(err),
		)
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully assigned permission to relation",
		zap.String("permission_id", permissionID.String()),
		zap.String("relation_type", relationType),
	)
	return nil
}

// CheckPermission проверяет наличие разрешения у субъекта для объекта
func (s *RoleService) CheckPermission(
	ctx context.Context,
	subjectID, objectID, permissionID uuid.UUID,
) (bool, error) {
	const op = "service.role.CheckPermission"

	s.logger.With(
		zap.String("op", op),
		zap.String("subject_id", subjectID.String()),
		zap.String("object_id", objectID.String()),
		zap.String("permission_id", permissionID.String()),
	)

	s.logger.Info("checking permission")

	hasPermission, err := s.roleInfoManage.CheckPermission(ctx, subjectID, objectID, permissionID)
	if err != nil {
		s.logger.Error("checking permission",
			zap.String("subject_id", subjectID.String()),
			zap.String("object_id", objectID.String()),
			zap.String("permission_id", permissionID.String()),
			zap.Error(err),
		)
		return false, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Permission check result",
		zap.String("subject_id", subjectID.String()),
		zap.String("object_id", objectID.String()),
		zap.String("permission_id", permissionID.String()),
		zap.Bool("has_permission", hasPermission),
	)
	return hasPermission, nil
}

// GetAllPermissions возвращает все разрешения в системе
func (s *RoleService) GetAllPermissions(
	ctx context.Context,
) (*[]models.Permission, error) {
	const op = "service.role.GetAllPermissions"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting all permissions")

	permissions, err := s.roleInfoManage.GetAllPermissions(ctx)
	if err != nil {
		s.logger.Error("getting all permissions",
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully retrieved all permissions",
		zap.Int("count", len(*permissions)),
	)
	return permissions, nil
}

// GetUserRelations возвращает все связи пользователя
func (s *RoleService) GetUserRelations(
	ctx context.Context,
	userID uuid.UUID,
) (*[]models.Relation, error) {
	const op = "service.role.GetUserRelations"

	s.logger.With(
		zap.String("op", op),
		zap.String("user_id", userID.String()),
	)

	s.logger.Info("getting user relations")

	relations, err := s.roleInfoManage.GetUserRelations(ctx, userID)
	if err != nil {
		s.logger.Error("getting user relations",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully retrieved user relations",
		zap.String("user_id", userID.String()),
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

	s.logger.With(
		zap.String("op", op),
		zap.String("user_id", userID.String()),
	)

	s.logger.Info("getting user permissions")

	permissions, err := s.roleInfoManage.GetUserPermissions(ctx, userID)
	if err != nil {
		s.logger.Error("getting user permissions",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully retrieved user permissions",
		zap.String("user_id", userID.String()),
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

	s.logger.With(
		zap.String("op", op),
		zap.String("relation_type", relationType),
	)

	s.logger.Info("getting permissions for relation type")

	permissions, err := s.roleInfoManage.GetPermissionsForRelationType(ctx, relationType)
	if err != nil {
		s.logger.Error("getting permissions for relation type",
			zap.String("relation_type", relationType),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully retrieved permissions for relation type",
		zap.String("relation_type", relationType),
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

	s.logger.With(
		zap.String("op", op),
		zap.String("entity_id", entityID.String()),
	)

	s.logger.Info("getting entity relations")

	relations, err := s.roleInfoManage.GetEntityRelations(ctx, entityID)
	if err != nil {
		s.logger.Error("getting entity relations",
			zap.String("entity_id", entityID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully retrieved entity relations",
		zap.String("entity_id", entityID.String()),
		zap.Int("count", len(*relations)),
	)
	return relations, nil
}
