package role

import (
	"context"
	"sso-service/internal/models"

	ssogrpc "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Role interface {
	CreateEntity(ctx context.Context, id uuid.UUID, entityType string) error
	DeleteEntity(ctx context.Context, id uuid.UUID) error
	CreateRelation(ctx context.Context, sourceID, targetID uuid.UUID, relationType string) error
	DeleteRelation(ctx context.Context, sourceID, targetID uuid.UUID, relationType string) error
	AddPermission(ctx context.Context, name, description string) (uuid.UUID, error)
	AssignPermission(ctx context.Context, permissionID uuid.UUID, relationType string) error
	CheckPermission(ctx context.Context, subjectID, objectID, permissionID uuid.UUID) (bool, error)
	GetAllPermissions(ctx context.Context) (*[]models.Permission, error)
	GetUserRelations(ctx context.Context, userID uuid.UUID) (*[]models.Relation, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) (*[]models.Permission, error)
	GetPermissionsForRelationType(ctx context.Context, relationType string) (*[]models.Permission, error)
	GetEntityRelations(ctx context.Context, entityID uuid.UUID) (*[]models.Relation, error)
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
) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.GetId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	if req.GetEntityType() == "" {
		return nil, status.Error(codes.InvalidArgument, "entity type is required")
	}

	err = s.role.CreateEntity(ctx, id, req.GetEntityType())
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

	permissionID, err := uuid.Parse(req.GetPermissionId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid permission UUID format: %v", err)
	}

	hasPermission, err := s.role.CheckPermission(ctx, subjectID, objectID, permissionID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check permission")
	}

	return &ssogrpc.CheckPermissionResponse{
		HasPermission: hasPermission,
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

	return &ssogrpc.GetAllPermissionsResponse{
		Permissions: convertPermissionsToProto(permissions),
	}, nil
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

	return &ssogrpc.GetUserRelationsResponse{
		Relations: convertRelationsToProto(relations),
	}, nil
}

func (s *RoleServerAPI) GetUserPermissions(
	ctx context.Context,
	req *ssogrpc.GetUserPermissionsRequest,
) (*ssogrpc.GetUserPermissionsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user UUID format: %v", err)
	}

	permissions, err := s.role.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user permissions")
	}

	return &ssogrpc.GetUserPermissionsResponse{
		Permission: convertPermissionsToProto(permissions),
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

	return &ssogrpc.GetPermissionsForRelationTypeResponse{
		Permissions: convertPermissionsToProto(permissions),
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

	return &ssogrpc.GetEntityRelationsResponse{
		Relations: convertRelationsToProto(relations),
	}, nil
}

func convertPermissionsToProto(p *[]models.Permission) []*ssogrpc.Permission {
	if p == nil {
		return nil
	}

	var permissions []*ssogrpc.Permission
	for _, permission := range *p {
		permissions = append(permissions, convertPermissionToProto(&permission))
	}

	return permissions
}

func convertPermissionToProto(p *models.Permission) *ssogrpc.Permission {
	if p == nil {
		return nil
	}
	return &ssogrpc.Permission{
		Id:          &ssogrpc.UUID{Value: p.ID.String()},
		Name:        p.Name,
		Description: p.Description,
	}
}

func convertRelationsToProto(r *[]models.Relation) []*ssogrpc.Relation {
	if r == nil {
		return nil
	}

	var relations []*ssogrpc.Relation
	for _, relation := range *r {
		relations = append(relations, convertRelationToProto(&relation))
	}

	return relations
}

func convertRelationToProto(r *models.Relation) *ssogrpc.Relation {
	if r == nil {
		return nil
	}
	return &ssogrpc.Relation{
		SourceId:     &ssogrpc.UUID{Value: r.SourceID.String()},
		TargetId:     &ssogrpc.UUID{Value: r.TargetID.String()},
		RelationType: r.RelationType,
	}
}
