package role

import (
	"context"
	"sso-service/internal/models"

	ssogrpc "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Role interface {
	CreateRole(ctx context.Context, name string, permissions []string) (*models.Role, error)
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
	RevokeRole(ctx context.Context, userID, roleID uuid.UUID) error
	CheckPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
}

type RoleServerAPI struct {
	ssogrpc.UnimplementedRoleServiceServer
	role Role
}

func NewRoleServer(gRPC *grpc.Server, role Role) {
	ssogrpc.RegisterRoleServiceServer(gRPC, &RoleServerAPI{role: role})
}

func (s *RoleServerAPI) CreateRole(
	ctx context.Context,
	req *ssogrpc.CreateRoleRequest,
) (*ssogrpc.CreateRoleResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid name role")
	}

	for _, permission := range req.Permissions {
		if permission == "" {
			return nil, status.Error(codes.InvalidArgument, "invalid role")
		}
	}

	role, err := s.role.CreateRole(ctx, req.GetName(), req.GetPermissions())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get tokens")
	}
	return &ssogrpc.CreateRoleResponse{Role: convertRoleToProto(role)}, nil
}

func (s *RoleServerAPI) AssignRole(
	ctx context.Context,
	req *ssogrpc.AssignRoleRequest,
) (*ssogrpc.AssignRoleResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.role.AssignRole(ctx, userID, roleID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to assign role")
	}

	return &ssogrpc.AssignRoleResponse{}, nil
}

func (s *RoleServerAPI) RevokeRole(
	ctx context.Context,
	req *ssogrpc.RevokeRoleRequest,
) (*ssogrpc.RevokeRoleResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.role.RevokeRole(ctx, userID, roleID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke role")
	}

	return &ssogrpc.RevokeRoleResponse{}, nil
}

func (s *RoleServerAPI) CheckPermission(
	ctx context.Context,
	req *ssogrpc.CheckPermissionRequest,
) (*ssogrpc.CheckPermissionResponse, error) {
	if req.Permission == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	permission, err := s.role.CheckPermission(ctx, userID, req.GetPermission())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssogrpc.CheckPermissionResponse{HasPermission: permission}, nil
}

func (s *RoleServerAPI) GetUserPermissions(
	ctx context.Context,
	req *ssogrpc.GetUserPermissionsRequest,
) (*ssogrpc.GetUserPermissionsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	permissions, err := s.role.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssogrpc.GetUserPermissionsResponse{Permissions: permissions}, nil
}

func convertRoleToProto(r *models.Role) (role *ssogrpc.Role) {
	if r == nil {
		return nil
	}
	return &ssogrpc.Role{
		Id:          r.ID.String(),
		Name:        r.Name,
		Permissions: r.Permissions,
		Description: r.Description,
	}
}
