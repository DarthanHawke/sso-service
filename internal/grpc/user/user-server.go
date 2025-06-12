package user

import (
	"context"
	"sso-service/internal/models"

	ssov1 "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"google.golang.org/grpc"
)

type User interface {
	Register(ctx context.Context, fullName, email, password string) (int64, error)
	Login(ctx context.Context, email, password string) (int64, error)
	GetProfile(ctx context.Context, userID string) (models.User, error)
	UpdateProfile(ctx context.Context, userID, fulName, email, password string) (models.User, error)
}

type UserServerAPI struct {
	ssov1.UnimplementedUserServiceServer
	user User
}

func NewUserServer(gRPC *grpc.Server, user User) {
	ssov1.RegisterUserServiceServer(gRPC, &UserServerAPI{user: user})
}
