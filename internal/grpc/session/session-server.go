package session

import (
	"context"
	"sso-service/internal/models"

	ssov1 "github.com/DarthanHawke/protos-payment-system/gen/go/auth"
	"google.golang.org/grpc"
)

type Session interface {
	CreateSession(ctx context.Context, userID string) (models.Session, error)
	RefreshSession(ctx context.Context, refreshToken string) (models.Session, error)
	Logout(ctx context.Context, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
}

type UserServerAPI struct {
	ssov1.UnimplementedAuthServiceServer
	session Session
}

func NewSessionServer(gRPC *grpc.Server, session Session) {
	ssov1.RegisterAuthServiceServer(gRPC, &UserServerAPI{session: session})
}
