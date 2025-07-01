package session

import (
	"context"

	ssogrpc "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Session interface {
	CreateSession(ctx context.Context, userID uuid.UUID) (string, string, error)
	RefreshSession(ctx context.Context, userID uuid.UUID, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
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
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	accsessToken, refreshToken, err := s.session.CreateSession(ctx, userID)
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

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	accsessToken, refreshToken, err := s.session.RefreshSession(ctx, userID, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get tokens")
	}

	return &ssogrpc.RefreshSessionResponse{AccessToken: accsessToken, RefreshToken: refreshToken}, nil
}

func (s *SessionServerAPI) Logout(
	ctx context.Context,
	req *ssogrpc.LogoutRequest,
) (*ssogrpc.LogoutResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	sessionID, err := uuid.Parse(req.GetSessionId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.session.Logout(ctx, userID, sessionID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssogrpc.LogoutResponse{}, nil
}

func (s *SessionServerAPI) LogoutAll(
	ctx context.Context,
	req *ssogrpc.LogoutAllRequest,
) (*ssogrpc.LogoutAllResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.session.LogoutAll(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssogrpc.LogoutAllResponse{}, nil
}
