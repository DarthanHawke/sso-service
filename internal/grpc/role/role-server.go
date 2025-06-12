package role

import (
	"context"
	"sso-service/internal/models"

	ssov1 "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"google.golang.org/grpc"
)

type Role interface {
	CreateRole(ctx context.Context, name string, permissions []string) (models.Role, error)
	AssignRole(ctx context.Context, userID, roleID int64) error
	RevokeRole(ctx context.Context, userID, roleID int64) error
	CheckPermission(ctx context.Context, userID int64, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID int64) ([]string, error)
}

type RoleServerAPI struct {
	ssov1.UnimplementedRoleServiceServer
	role Role
}

func NewUserServer(gRPC *grpc.Server, role Role) {
	ssov1.RegisterRoleServiceServer(gRPC, &RoleServerAPI{role: role})
}
