package session

import (
	"context"
	"sso-service/internal/models"

	ssogrpc "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Session interface {
	CreateSession(ctx context.Context, userID uuid.UUID) (*models.UserSession, error)
	RefreshSession(ctx context.Context, userID uuid.UUID, refreshToken string) (*models.UserSession, error)
	Logout(ctx context.Context, userID, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]models.Session, error)
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
	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	userSession, err := s.session.CreateSession(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get tokens")
	}
	return &ssogrpc.CreateSessionResponse{
		AccessToken:  userSession.AcssesToken,
		RefreshToken: userSession.RefreshToken,
	}, nil
}

func (s *SessionServerAPI) RefreshSession(
	ctx context.Context,
	req *ssogrpc.RefreshSessionRequest,
) (*ssogrpc.RefreshSessionResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	userID, err := uuid.Parse(req.UserId.Value)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	userSession, err := s.session.RefreshSession(ctx, userID, req.GetRefreshToken())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get tokens")
	}

	return &ssogrpc.RefreshSessionResponse{
		AccessToken:  userSession.AcssesToken,
		RefreshToken: userSession.RefreshToken,
	}, nil
}

func (s *SessionServerAPI) Logout(
	ctx context.Context,
	req *ssogrpc.LogoutRequest,
) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	sessionID, err := uuid.Parse(req.GetSessionId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.session.Logout(ctx, userID, sessionID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return nil, nil
}

func (s *SessionServerAPI) LogoutAll(
	ctx context.Context,
	req *ssogrpc.LogoutAllRequest,
) (*emptypb.Empty, error) {
	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.session.LogoutAll(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return nil, nil
}

func (s *SessionServerAPI) GetAll(
	ctx context.Context,
	req *ssogrpc.GetAllRequest,
) (*ssogrpc.GetAllResponse, error) {
	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	sessions, err := s.session.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	protoSessnions := make([]*ssogrpc.Session, 0, len(sessions))
	for _, session := range sessions {
		protoSessnions = append(protoSessnions, &ssogrpc.Session{
			Id:        &ssogrpc.UUID{Value: session.ID.String()},
			UserId:    &ssogrpc.UUID{Value: session.UserID.String()},
			ExpiresAt: timestamppb.New(session.ExpiresAt),
		})
	}

	return &ssogrpc.GetAllResponse{Session: protoSessnions}, nil
}
