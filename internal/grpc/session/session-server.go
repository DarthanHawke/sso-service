package session

import (
	"context"

	ssogrpc "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Session interface {
	CreateSession(ctx context.Context, userID []uint8) (string, string, error)
	RefreshSession(ctx context.Context, userID []uint8, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID []uint8, sessionID []uint8) error
	LogoutAll(ctx context.Context, userID []uint8) error
}

type SessionServerAPI struct {
	ssogrpc.UnimplementedSessionServiceServer
	session Session
}

func NewSessionServer(gRPC *grpc.Server, session Session) {
	ssogrpc.RegisterSessionServiceServer(gRPC, &SessionServerAPI{session: session})
}

func (s *SessionServerAPI) CreateSession(
	ctx context.Context,
	req *ssogrpc.CreateSessionRequest,
) (*ssogrpc.CreateSessionResponse, error) {
	accsessToken, refreshToken, err := s.session.CreateSession(ctx, []uint8(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get tokens")
	}
	return &ssogrpc.CreateSessionResponse{AccessToken: accsessToken, RefreshToken: refreshToken}, nil
}

func (s *SessionServerAPI) RefreshSession(
	ctx context.Context,
	req *ssogrpc.RefreshSessionRequest,
) (*ssogrpc.RefreshSessionResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	accsessToken, refreshToken, err := s.session.RefreshSession(ctx, []uint8(req.GetUserId()), req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get tokens")
	}

	return &ssogrpc.RefreshSessionResponse{AccessToken: accsessToken, RefreshToken: refreshToken}, nil
}

func (s *SessionServerAPI) Logout(ctx context.Context, req *ssogrpc.LogoutRequest) (*ssogrpc.LogoutResponse, error) {
	err := s.session.Logout(ctx, []uint8(req.GetUserId()), []uint8(req.GetSessionId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssogrpc.LogoutResponse{}, nil
}

func (s *SessionServerAPI) LogoutAll(
	ctx context.Context,
	req *ssogrpc.LogoutAllRequest,
) (*ssogrpc.LogoutAllResponse, error) {
	err := s.session.LogoutAll(ctx, []uint8(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssogrpc.LogoutAllResponse{}, nil
}
