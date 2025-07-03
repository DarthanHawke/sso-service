package session

import (
	"context"
	"sso-service/internal/models"

	ssogrpc "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Session interface {
	CreateSession(ctx context.Context, userID uuid.UUID) (*models.UserSession, error)
	RefreshSession(ctx context.Context, userID uuid.UUID, refreshToken string) (*models.UserSession, error)
	Logout(ctx context.Context, userID, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	GetUserSessions(ctx context.Context, userID uuid.UUID) (*[]models.Session, error)
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

	userID, err := uuid.Parse(req.GetUserId())
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

func (s *SessionServerAPI) GetAll(
	ctx context.Context,
	req *ssogrpc.GetAllRequest,
) (*ssogrpc.GetAllResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	sessionsID, err := s.session.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &ssogrpc.GetAllResponse{Session: convertSessionsToProto(sessionsID)}, nil
}

func convertSessionsToProto(s *[]models.Session) []*ssogrpc.Session {
	if s == nil {
		return nil
	}

	var sessnions []*ssogrpc.Session
	for _, session := range *s {
		sessnions = append(sessnions, convertSessionToProto(&session))
	}

	return sessnions
}

func convertSessionToProto(s *models.Session) *ssogrpc.Session {
	if s == nil {
		return nil
	}

	return &ssogrpc.Session{
		Id:        s.ID.String(),
		UserId:    s.UserID.String(),
		ExpiresAt: timestamppb.New(s.ExpiresAt),
	}
}
