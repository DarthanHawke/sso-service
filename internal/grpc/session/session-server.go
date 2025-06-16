package session

import (
	"context"

	ssov1 "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Session interface {
	CreateSession(ctx context.Context, userID []uint8) (string, string, error)
	RefreshSession(ctx context.Context, userID []uint8, refreshToken string) (string, string, error)
	Logout(ctx context.Context, sessionID []uint8) error
	LogoutAll(ctx context.Context, userID []uint8) error
}

type SessionServerAPI struct {
	ssov1.UnimplementedSessionServiceServer
	session Session
}

func NewSessionServer(gRPC *grpc.Server, session Session) {
	ssov1.RegisterSessionServiceServer(gRPC, &SessionServerAPI{session: session})
}

func (s *SessionServerAPI) CreateSession(ctx context.Context, req *ssov1.CreateSessionRequest) (*ssov1.CreateSessionResponse, error) {
	accsessToken, refreshToken, err := s.session.CreateSession(ctx, []uint8(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get tokens")
	}
	return &ssov1.CreateSessionResponse{AccessToken: accsessToken, RefreshToken: refreshToken}, nil
}

func (s *SessionServerAPI) RefreshSession(ctx context.Context, req *ssov1.RefreshSessionRequest) (*ssov1.RefreshSessionResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	accsessToken, refreshToken, err := s.session.RefreshSession(ctx, []uint8(req.GetUserId()), req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get tokens")
	}

	return &ssov1.RefreshSessionResponse{AccessToken: accsessToken, RefreshToken: refreshToken}, nil
}

func (s *SessionServerAPI) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {
	err := s.session.Logout(ctx, []uint8(req.GetSessionId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssov1.LogoutResponse{}, nil
}

func (s *SessionServerAPI) LogoutAll(ctx context.Context, req *ssov1.LogoutAllRequest) (*ssov1.LogoutAllResponse, error) {
	err := s.session.LogoutAll(ctx, []uint8(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssov1.LogoutAllResponse{}, nil
}
