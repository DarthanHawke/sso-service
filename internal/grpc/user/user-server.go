package user

import (
	"context"
	"sso-service/internal/models"

	ssov1 "github.com/DarthanHawke/protos-payment-system/gen/go/auth"
	"google.golang.org/grpc"
)

type User interface {
	Register(ctx context.Context, email, password, fullName string) (int64, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetProfile(ctx context.Context, userID string) (*models.User, error)
	UpdateProfile(ctx context.Context, userID, email, password, fulName string) error
}

type UserServerAPI struct {
	ssov1.UnimplementedUserServiceServer
	user User
}

func NewUserServer(gRPC *grpc.Server, user User) {
	ssov1.RegisterUserServiceServer(gRPC, &UserServerAPI{user: user})
}
