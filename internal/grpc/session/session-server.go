package session

import (
	"context"

	ssov1 "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"google.golang.org/grpc"
)

type Session interface {
	CreateSession(ctx context.Context, userID int64) (string, string, error)
	RefreshSession(ctx context.Context, userID int64, refreshToken string) (string, string, error)
	Logout(ctx context.Context, sessionID string) error
	LogoutAll(ctx context.Context, userID string) error
}

type SessionServerAPI struct {
	ssov1.UnimplementedSessionServiceServer
	session Session
}

func NewSessionServer(gRPC *grpc.Server, session Session) {
	ssov1.RegisterSessionServiceServer(gRPC, &SessionServerAPI{session: session})
}
