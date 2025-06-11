package role

import (
	"context"
	"sso-service/internal/models"
	//ssov1 "github.com/DarthanHawke/protos-payment-system/gen/go/role"
)

type Role interface {
	CreateRole(ctx context.Context, name string, permissions []string) (*models.Role, error)
	AssignRole(ctx context.Context, userID, roleID string) error
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	RevokeRole(ctx context.Context, userID, roleID string) error
}

type RoleServerAPI struct {
	//ssov1.UnimplementedRoleServiceServer
	//role server: need change proto
	//role Role
}
