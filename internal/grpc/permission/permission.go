package permission

import (
	"context"
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

type Permission interface {
	CreatePermission(ctx context.Context, code, description string) (uuid.UUID, error)
	DeletePermission(ctx context.Context, permID uuid.UUID) error
	GetPermissionByID(ctx context.Context, permID uuid.UUID) (*models.Permission, error)
	GetPermissionByCode(ctx context.Context, code string) (*models.Permission, error)
	ListPermissions(ctx context.Context, limit, offset int) (*[]models.Permission, error)
	UpdatePermission(ctx context.Context, permID uuid.UUID, description string) error
	AddPermissionToRole(ctx context.Context, roleID, permID uuid.UUID) error
	RevokePermissionFromRole(ctx context.Context, roleID, permID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) (*[]models.Permission, error)
	HasRolePermission(ctx context.Context, roleID uuid.UUID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) (*[]models.Permission, error)
	HasUserPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
}

type PermissionServerAPI struct {
	ssogrpc.UnimplementedPermissionServiceServer
	permission Permission
}

func NewRoleServer(gRPC *grpc.Server, permission Permission) {
	ssogrpc.RegisterPermissionServiceServer(gRPC, &PermissionServerAPI{permission: permission})
}

func (s *PermissionServerAPI) CreatePermission(
	ctx context.Context,
	req *ssogrpc.CreatePermissionRequest,
) (*ssogrpc.CreatePermissionResponse, error) {
	if req.Code == nilStr {
		return nil, status.Error(codes.InvalidArgument, "invalid name permissions")
	}

	permissionId, err := s.permission.CreatePermission(ctx, req.GetCode(), req.GetDescription())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create permission")
	}
	return &ssogrpc.CreatePermissionResponse{PermId: permissionId.String()}, nil
}

func (s *PermissionServerAPI) DeletePermission(
	ctx context.Context,
	req *ssogrpc.DeletePermissionRequest,
) (*ssogrpc.DeletePermissionResponse, error) {
	permissionId, err := uuid.Parse(req.GetPermId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.permission.DeletePermission(ctx, permissionId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete permission")
	}
	return &ssogrpc.DeletePermissionResponse{}, nil
}

func (s *PermissionServerAPI) GetPermissionByID(
	ctx context.Context,
	req *ssogrpc.GetPermissionByIDRequest,
) (*ssogrpc.GetPermissionByIDResponse, error) {
	permissionId, err := uuid.Parse(req.GetPermId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	permission, err := s.permission.GetPermissionByID(ctx, permissionId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get permission")
	}
	return &ssogrpc.GetPermissionByIDResponse{Permission: convertPermissionToProto(permission)}, nil
}

func (s *PermissionServerAPI) GetPermissionByCode(
	ctx context.Context,
	req *ssogrpc.GetPermissionByCodeRequest,
) (*ssogrpc.GetPermissionByCodeResponse, error) {
	if req.Code == nilStr {
		return nil, status.Error(codes.InvalidArgument, "invalid name permissions")
	}

	permission, err := s.permission.GetPermissionByCode(ctx, req.GetCode())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get permission")
	}
	return &ssogrpc.GetPermissionByCodeResponse{Permission: convertPermissionToProto(permission)}, nil
}

func (s *PermissionServerAPI) ListPermissions(
	ctx context.Context,
	req *ssogrpc.ListPermissionsRequest,
) (*ssogrpc.ListPermissionsResponse, error) {
	permissions, err := s.permission.ListPermissions(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get list permissions")
	}
	return &ssogrpc.ListPermissionsResponse{Permission: convertPermissionsToProto(permissions)}, nil
}

func (s *PermissionServerAPI) UpdatePermission(
	ctx context.Context,
	req *ssogrpc.UpdatePermissionRequest,
) (*ssogrpc.UpdatePermissionResponse, error) {
	permissionId, err := uuid.Parse(req.GetPermId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.permission.UpdatePermission(ctx, permissionId, req.GetDescription())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update permission")
	}
	return &ssogrpc.UpdatePermissionResponse{}, nil
}

func (s *PermissionServerAPI) AddPermissionToRole(
	ctx context.Context,
	req *ssogrpc.AddPermissionToRoleRequest,
) (*ssogrpc.AddPermissionToRoleResponse, error) {
	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	permissionId, err := uuid.Parse(req.GetPermId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.permission.AddPermissionToRole(ctx, roleID, permissionId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to assign permission")
	}

	return &ssogrpc.AddPermissionToRoleResponse{}, nil
}

func (s *PermissionServerAPI) RevokePermissionFromRole(
	ctx context.Context,
	req *ssogrpc.RevokePermissionFromRoleRequest,
) (*ssogrpc.RevokePermissionFromRoleResponse, error) {
	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	permissionId, err := uuid.Parse(req.GetPermId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.permission.RevokePermissionFromRole(ctx, roleID, permissionId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke permission")
	}

	return &ssogrpc.RevokePermissionFromRoleResponse{}, nil
}

func (s *PermissionServerAPI) GetRolePermissions(
	ctx context.Context,
	req *ssogrpc.GetRolePermissionsRequest,
) (*ssogrpc.GetRolePermissionsResponse, error) {
	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	permissions, err := s.permission.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get role permissions")
	}

	return &ssogrpc.GetRolePermissionsResponse{Permission: convertPermissionsToProto(permissions)}, nil
}

func (s *PermissionServerAPI) HasRolePermission(
	ctx context.Context,
	req *ssogrpc.HasRolePermissionRequest,
) (*ssogrpc.HasRolePermissionResponse, error) {
	roleID, err := uuid.Parse(req.GetRoleId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}
	if req.Code == nilStr {
		return nil, status.Error(codes.InvalidArgument, "invalid name permissions")
	}

	allowed, err := s.permission.HasRolePermission(ctx, roleID, req.GetCode())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check role permissions")
	}

	return &ssogrpc.HasRolePermissionResponse{Allowed: allowed}, nil
}

func (s *PermissionServerAPI) GetUserPermissions(
	ctx context.Context,
	req *ssogrpc.GetUserPermissionsRequest,
) (*ssogrpc.GetUserPermissionsResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	permissions, err := s.permission.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get role permissions")
	}

	return &ssogrpc.GetUserPermissionsResponse{Permission: convertPermissionsToProto(permissions)}, nil
}

func (s *PermissionServerAPI) HasUserPermission(
	ctx context.Context,
	req *ssogrpc.HasUserPermissionRequest,
) (*ssogrpc.HasUserPermissionResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}
	if req.Code == nilStr {
		return nil, status.Error(codes.InvalidArgument, "invalid name permissions")
	}

	allowed, err := s.permission.HasUserPermission(ctx, userID, req.GetCode())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check role permissions")
	}

	return &ssogrpc.HasUserPermissionResponse{Allowed: allowed}, nil
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

func convertPermissionToProto(p *models.Permission) (permission *ssogrpc.Permission) {
	if p == nil {
		return nil
	}
	return &ssogrpc.Permission{
		Id:          p.ID.String(),
		Code:        p.Code,
		Description: p.Description,
	}
}
