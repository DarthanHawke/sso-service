package role

import (
	"context"
	"errors"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	ssogrpc "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Role interface {
	CreateEntity(ctx context.Context, entityType string) (uuid.UUID, error)
	CreateEntityWithID(ctx context.Context, id uuid.UUID, entityType string) error
	DeleteEntity(ctx context.Context, id uuid.UUID) error
	CreateRelation(ctx context.Context, sourceID, targetID uuid.UUID, relationType string) error
	DeleteRelation(ctx context.Context, sourceID, targetID uuid.UUID, relationType string) error
	AddPermission(ctx context.Context, name, description string) (uuid.UUID, error)
	AssignPermission(ctx context.Context, permissionID uuid.UUID, relationType string) error
	RevokePermission(ctx context.Context, permissionID uuid.UUID, relationType string) error
	CheckPermission(ctx context.Context, subjectID, objectID uuid.UUID, permissionName string) (bool, error)
	GetPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	GetEntityID(ctx context.Context, entityType string) (uuid.UUID, error)
	GetAllEntities(ctx context.Context) ([]models.Entity, error)
	GetAllPermissions(ctx context.Context) ([]models.Permission, error)
	GetUserRelations(ctx context.Context, userID uuid.UUID) ([]models.Relation, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]models.Permission, error)
	GetPermissionsForRelationType(ctx context.Context, relationType string) ([]models.Permission, error)
	GetEntityRelations(ctx context.Context, entityID uuid.UUID) ([]models.Relation, error)
}

type RoleServerAPI struct {
	ssogrpc.UnimplementedRoleServiceServer
	role Role
}

func NewRoleServer(gRPC *grpc.Server, role Role) {
	ssogrpc.RegisterRoleServiceServer(gRPC, &RoleServerAPI{role: role})
}

func (s *RoleServerAPI) CreateEntity(
	ctx context.Context,
	req *ssogrpc.CreateEntityRequest,
) (*ssogrpc.CreateEntityResponse, error) {
	if req.GetEntityType() == "" {
		return nil, status.Error(codes.InvalidArgument, "entity type is required")
	}

	entityID, err := s.role.CreateEntity(ctx, req.GetEntityType())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create entity")
	}
	return &ssogrpc.CreateEntityResponse{Id: &ssogrpc.UUID{Value: entityID.String()}}, nil
}

func (s *RoleServerAPI) CreateEntityWithID(
	ctx context.Context,
	req *ssogrpc.CreateEntityWithIDRequest,
) (*emptypb.Empty, error) {
	if req.GetEntityType() == "" {
		return nil, status.Error(codes.InvalidArgument, "entity type is required")
	}
	id, err := uuid.Parse(req.GetId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}
	err = s.role.CreateEntityWithID(ctx, id, req.GetEntityType())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create entity")
	}
	return &emptypb.Empty{}, nil
}

func (s *RoleServerAPI) DeleteEntity(
	ctx context.Context,
	req *ssogrpc.DeleteEntityRequest,
) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.GetId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.role.DeleteEntity(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete entity")
	}
	return &emptypb.Empty{}, nil
}

func (s *RoleServerAPI) CreateRelation(
	ctx context.Context,
	req *ssogrpc.CreateRelationRequest,
) (*emptypb.Empty, error) {
	sourceID, err := uuid.Parse(req.GetSourceId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid source UUID format: %v", err)
	}

	targetID, err := uuid.Parse(req.GetTargetId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target UUID format: %v", err)
	}

	if req.GetRelationType() == "" {
		return nil, status.Error(codes.InvalidArgument, "relation type is required")
	}

	err = s.role.CreateRelation(ctx, sourceID, targetID, req.GetRelationType())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create relation")
	}
	return &emptypb.Empty{}, nil
}

func (s *RoleServerAPI) DeleteRelation(
	ctx context.Context,
	req *ssogrpc.DeleteRelationRequest,
) (*emptypb.Empty, error) {
	sourceID, err := uuid.Parse(req.GetSourceId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid source UUID format: %v", err)
	}

	targetID, err := uuid.Parse(req.GetTargetId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target UUID format: %v", err)
	}

	if req.GetRelationType() == "" {
		return nil, status.Error(codes.InvalidArgument, "relation type is required")
	}

	err = s.role.DeleteRelation(ctx, sourceID, targetID, req.GetRelationType())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete relation")
	}
	return &emptypb.Empty{}, nil
}

func (s *RoleServerAPI) AddPermission(
	ctx context.Context,
	req *ssogrpc.AddPermissionRequest,
) (*ssogrpc.AddPermissionResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "permission name is required")
	}

	permissionID, err := s.role.AddPermission(ctx, req.GetName(), req.GetDescription())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to add permission")
	}

	return &ssogrpc.AddPermissionResponse{
		PermissionId: &ssogrpc.UUID{Value: permissionID.String()},
	}, nil
}

func (s *RoleServerAPI) AssignPermission(
	ctx context.Context,
	req *ssogrpc.AssignPermissionRequest,
) (*emptypb.Empty, error) {
	permissionID, err := uuid.Parse(req.GetPermissionId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid permission UUID format: %v", err)
	}

	if req.GetRelationType() == "" {
		return nil, status.Error(codes.InvalidArgument, "relation type is required")
	}

	err = s.role.AssignPermission(ctx, permissionID, req.GetRelationType())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to assign permission")
	}
	return &emptypb.Empty{}, nil
}

func (s *RoleServerAPI) RevokePermission(
	ctx context.Context,
	req *ssogrpc.RevokePermissionRequest,
) (*emptypb.Empty, error) {
	permissionID, err := uuid.Parse(req.GetPermissionId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid permission UUID format: %v", err)
	}

	if req.GetRelationType() == "" {
		return nil, status.Error(codes.InvalidArgument, "relation type is required")
	}

	err = s.role.RevokePermission(ctx, permissionID, req.GetRelationType())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to assign permission")
	}
	return &emptypb.Empty{}, nil
}

func (s *RoleServerAPI) CheckPermission(
	ctx context.Context,
	req *ssogrpc.CheckPermissionRequest,
) (*ssogrpc.CheckPermissionResponse, error) {
	subjectID, err := uuid.Parse(req.GetSubjectId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid subject UUID format: %v", err)
	}

	objectID, err := uuid.Parse(req.GetObjectId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid object UUID format: %v", err)
	}

	hasPermission, err := s.role.CheckPermission(ctx, subjectID, objectID, req.GetName())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check permission")
	}

	return &ssogrpc.CheckPermissionResponse{
		HasPermission: hasPermission,
	}, nil
}

func (s *RoleServerAPI) GetPermissionByName(
	ctx context.Context,
	req *ssogrpc.GetPermissionByNameRequest,
) (*ssogrpc.GetPermissionByNameResponse, error) {
	permission, err := s.role.GetPermissionByName(ctx, req.GetName())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get permissions")
	}

	return &ssogrpc.GetPermissionByNameResponse{
		Permission: &ssogrpc.Permission{
			Id:          &ssogrpc.UUID{Value: permission.ID.String()},
			Name:        permission.Name,
			Description: permission.Description,
		},
	}, nil
}

func (s *RoleServerAPI) GetEntityId(
	ctx context.Context,
	req *ssogrpc.GetEntityIdRequest,
) (*ssogrpc.GetEntityIdResponse, error) {
	if req.GetEntityType() == "" {
		return nil, status.Error(codes.InvalidArgument, "entity type is required")
	}

	id, err := s.role.GetEntityID(ctx, req.GetEntityType())
	if err != nil {
		if errors.Is(err, ssoerrors.ErrEntityNotFound) {
			return nil, status.Error(codes.NotFound, "entity not found")
		}
		return nil, status.Error(codes.Internal, "failed to get entity ID")
	}

	return &ssogrpc.GetEntityIdResponse{
		Id: &ssogrpc.UUID{Value: id.String()},
	}, nil
}

func (s *RoleServerAPI) GetAllEntities(
	ctx context.Context,
	_ *emptypb.Empty,
) (*ssogrpc.GetAllEntitiesResponse, error) {
	entities, err := s.role.GetAllEntities(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get entities")
	}

	protoEntities := make([]*ssogrpc.Entity, 0, len(entities))
	for _, entity := range entities {
		protoEntities = append(protoEntities, &ssogrpc.Entity{
			Id:   &ssogrpc.UUID{Value: entity.ID.String()},
			Type: entity.Type,
		})
	}

	return &ssogrpc.GetAllEntitiesResponse{
		Entities: protoEntities,
	}, nil
}

func (s *RoleServerAPI) GetAllPermissions(
	ctx context.Context,
	_ *emptypb.Empty,
) (*ssogrpc.GetAllPermissionsResponse, error) {
	permissions, err := s.role.GetAllPermissions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get permissions")
	}

	protoPermissions := make([]*ssogrpc.Permission, 0, len(permissions))
	for _, permission := range permissions {
		protoPermissions = append(protoPermissions, &ssogrpc.Permission{
			Id:          &ssogrpc.UUID{Value: permission.ID.String()},
			Name:        permission.Name,
			Description: permission.Description,
		})
	}

	return &ssogrpc.GetAllPermissionsResponse{Permissions: protoPermissions}, nil
}

func (s *RoleServerAPI) GetUserRelations(
	ctx context.Context,
	req *ssogrpc.GetUserRelationsRequest,
) (*ssogrpc.GetUserRelationsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user UUID format: %v", err)
	}

	relations, err := s.role.GetUserRelations(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user relations")
	}

	protoRelations := make([]*ssogrpc.Relation, 0, len(relations))
	for _, relation := range relations {
		protoRelations = append(protoRelations, &ssogrpc.Relation{
			SourceId:     &ssogrpc.UUID{Value: relation.SourceID.String()},
			TargetId:     &ssogrpc.UUID{Value: relation.TargetID.String()},
			RelationType: relation.RelationType,
		})
	}

	return &ssogrpc.GetUserRelationsResponse{Relations: protoRelations}, nil
}

func (s *RoleServerAPI) GetUserPermissions(
	ctx context.Context,
	req *ssogrpc.GetUserPermissionsRequest,
) (*ssogrpc.GetUserPermissionsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user UUID format: %v", err)
	}

	permissions, err := s.role.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user permissions")
	}

	protoPermissions := make([]*ssogrpc.Permission, 0, len(permissions))
	for _, permission := range permissions {
		protoPermissions = append(protoPermissions, &ssogrpc.Permission{
			Id:          &ssogrpc.UUID{Value: permission.ID.String()},
			Name:        permission.Name,
			Description: permission.Description,
		})
	}

	return &ssogrpc.GetUserPermissionsResponse{
		Permissions: protoPermissions,
	}, nil
}

func (s *RoleServerAPI) GetPermissionsForRelationType(
	ctx context.Context,
	req *ssogrpc.GetPermissionsForRelationTypeRequest,
) (*ssogrpc.GetPermissionsForRelationTypeResponse, error) {
	if req.GetRelationType() == "" {
		return nil, status.Error(codes.InvalidArgument, "relation type is required")
	}

	permissions, err := s.role.GetPermissionsForRelationType(ctx, req.GetRelationType())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get permissions for relation type")
	}

	protoPermissions := make([]*ssogrpc.Permission, 0, len(permissions))
	for _, permission := range permissions {
		protoPermissions = append(protoPermissions, &ssogrpc.Permission{
			Id:          &ssogrpc.UUID{Value: permission.ID.String()},
			Name:        permission.Name,
			Description: permission.Description,
		})
	}

	return &ssogrpc.GetPermissionsForRelationTypeResponse{
		Permissions: protoPermissions,
	}, nil
}

func (s *RoleServerAPI) GetEntityRelations(
	ctx context.Context,
	req *ssogrpc.GetEntityRelationsRequest,
) (*ssogrpc.GetEntityRelationsResponse, error) {
	entityID, err := uuid.Parse(req.GetEntityId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity UUID format: %v", err)
	}

	relations, err := s.role.GetEntityRelations(ctx, entityID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get entity relations")
	}

	protoRelations := make([]*ssogrpc.Relation, 0, len(relations))
	for _, relation := range relations {
		protoRelations = append(protoRelations, &ssogrpc.Relation{
			SourceId:     &ssogrpc.UUID{Value: relation.SourceID.String()},
			TargetId:     &ssogrpc.UUID{Value: relation.TargetID.String()},
			RelationType: relation.RelationType,
		})
	}

	return &ssogrpc.GetEntityRelationsResponse{Relations: protoRelations}, nil
}
