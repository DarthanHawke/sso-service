package role

import (
	"context"
	"slices"
	"sso-service/internal/models"

	ssogrpc "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	nilStr = ""
)

type Role interface {
	CreateRole(ctx context.Context, name string, permissions []string, description string) (uuid.UUID, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	GetRoleByID(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	GetRoleByName(ctx context.Context, roleName string) (*models.Role, error)
	ListRoles(ctx context.Context, limit, offset int) (*[]models.Role, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, description string) error
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RevokeRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) (*[]models.Role, error)
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
	if req.Name == nilStr {
		return nil, status.Error(codes.InvalidArgument, "invalid name role")
	}

	if slices.Contains(req.Permissions, nilStr) {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	roleId, err := s.role.CreateRole(ctx, req.GetName(), req.GetPermissions(), req.GetDescription())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create role")
	}
	return &ssogrpc.CreateRoleResponse{RoleId: roleId.String()}, nil
}

func (s *RoleServerAPI) DeleteRole(
	ctx context.Context,
	req *ssogrpc.DeleteRoleRequest,
) (*ssogrpc.DeleteRoleResponse, error) {
	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.role.DeleteRole(ctx, roleID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete role")
	}
	return &ssogrpc.DeleteRoleResponse{}, nil
}

func (s *RoleServerAPI) GetRoleByID(
	ctx context.Context,
	req *ssogrpc.GetRoleByIDRequest,
) (*ssogrpc.GetRoleByIDResponse, error) {
	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	role, err := s.role.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get role")
	}
	return &ssogrpc.GetRoleByIDResponse{Role: convertRoleToProto(role)}, nil
}

func (s *RoleServerAPI) GetRoleByName(
	ctx context.Context,
	req *ssogrpc.GetRoleByNameRequest,
) (*ssogrpc.GetRoleByNameResponse, error) {
	if req.Name == nilStr {
		return nil, status.Error(codes.InvalidArgument, "invalid name")
	}

	role, err := s.role.GetRoleByName(ctx, req.GetName())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get role")
	}
	return &ssogrpc.GetRoleByNameResponse{Role: convertRoleToProto(role)}, nil
}

func (s *RoleServerAPI) ListRoles(
	ctx context.Context,
	req *ssogrpc.ListRolesRequest,
) (*ssogrpc.ListRolesResponse, error) {
	roles, err := s.role.ListRoles(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get list roles")
	}
	return &ssogrpc.ListRolesResponse{Role: convertRolesToProto(roles)}, nil
}

func (s *RoleServerAPI) UpdateRole(
	ctx context.Context,
	req *ssogrpc.UpdateRoleRequest,
) (*ssogrpc.UpdateRoleResponse, error) {
	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.role.UpdateRole(ctx, roleID, req.GetDescription())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update role")
	}
	return &ssogrpc.UpdateRoleResponse{}, nil
}

func (s *RoleServerAPI) AssignRoleToUser(
	ctx context.Context,
	req *ssogrpc.AssignRoleToUserRequest,
) (*ssogrpc.AssignRoleToUserResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.role.AssignRoleToUser(ctx, userID, roleID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to assign role")
	}

	return &ssogrpc.AssignRoleToUserResponse{}, nil
}

func (s *RoleServerAPI) RevokeRoleFromUser(
	ctx context.Context,
	req *ssogrpc.RevokeRoleFromUserRequest,
) (*ssogrpc.RevokeRoleFromUserResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.role.RevokeRoleFromUser(ctx, userID, roleID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke role")
	}

	return &ssogrpc.RevokeRoleFromUserResponse{}, nil
}

func (s *RoleServerAPI) GetUserRoles(
	ctx context.Context,
	req *ssogrpc.GetUserRolesRequest,
) (*ssogrpc.GetUserRolesResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	roles, err := s.role.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssogrpc.GetUserRolesResponse{Role: convertRolesToProto(roles)}, nil
}

func convertRolesToProto(r *[]models.Role) []*ssogrpc.Role {
	if r == nil {
		return nil
	}

	var roles []*ssogrpc.Role
	for _, role := range *r {
		roles = append(roles, convertRoleToProto(&role))
	}

	return roles
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
