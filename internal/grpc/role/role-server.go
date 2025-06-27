package role

import (
	"context"
	"sso-service/internal/models"

	ssov1 "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Role interface {
	CreateRole(ctx context.Context, name string, permissions []string) (models.Role, error)
	AssignRole(ctx context.Context, userID []uint8, roleID int64) error
	RevokeRole(ctx context.Context, userID []uint8, roleID int64) error
	CheckPermission(ctx context.Context, userID []uint8, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID []uint8) ([]string, error)
}

type RoleServerAPI struct {
	ssov1.UnimplementedRoleServiceServer
	role Role
}

func NewRoleServer(gRPC *grpc.Server, role Role) {
	ssov1.RegisterRoleServiceServer(gRPC, &RoleServerAPI{role: role})
}

func (s *RoleServerAPI) CreateRole(ctx context.Context, req *ssov1.CreateRoleRequest) (*ssov1.CreateRoleResponse, error) {
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
	return &ssov1.CreateRoleResponse{Role: convertRoleToProto(&role)}, nil
}

func (s *RoleServerAPI) AssignRole(ctx context.Context, req *ssov1.AssignRoleRequest) (*ssov1.AssignRoleResponse, error) {
	err := s.role.AssignRole(ctx, []uint8(req.GetUserId()), req.GetRoleId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to assign role")
	}

	return &ssov1.AssignRoleResponse{}, nil
}

func (s *RoleServerAPI) RevokeRole(ctx context.Context, req *ssov1.RevokeRoleRequest) (*ssov1.RevokeRoleResponse, error) {
	err := s.role.RevokeRole(ctx, []uint8(req.GetUserId()), req.GetRoleId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke role")
	}

	return &ssov1.RevokeRoleResponse{}, nil
}

func (s *RoleServerAPI) CheckPermission(ctx context.Context, req *ssov1.CheckPermissionRequest) (*ssov1.CheckPermissionResponse, error) {
	if req.Permission == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}
	permission, err := s.role.CheckPermission(ctx, []uint8(req.GetUserId()), req.GetPermission())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssov1.CheckPermissionResponse{HasPermission: permission}, nil
}

func (s *RoleServerAPI) GetUserPermissions(ctx context.Context, req *ssov1.GetUserPermissionsRequest) (*ssov1.GetUserPermissionsResponse, error) {
	permissions, err := s.role.GetUserPermissions(ctx, []uint8(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssov1.GetUserPermissionsResponse{Permissions: permissions}, nil
}

func convertRoleToProto(r *models.Role) (role *ssov1.Role) {
	if r == nil {
		return nil
	}
	return &ssov1.Role{
		Id:          r.ID,
		Name:        r.Name,
		Permissions: r.Permissions,
		Description: r.Description,
	}
}
