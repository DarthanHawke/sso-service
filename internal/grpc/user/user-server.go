package user

import (
	"context"
	"errors"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	ssogrpc "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type User interface {
	Register(ctx context.Context, fullName, email, password string) (uuid.UUID, error)
	Login(ctx context.Context, email, password string) (uuid.UUID, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, fulName, email, password string) (*models.User, error)
}

type UserServerAPI struct {
	ssogrpc.UnimplementedUserServiceServer
	user User
}

func NewUserServer(gRPC *grpc.Server, user User) {
	ssogrpc.RegisterUserServiceServer(gRPC, &UserServerAPI{user: user})
}

func (s *UserServerAPI) Register(ctx context.Context, req *ssogrpc.RegisterRequest) (*ssogrpc.RegisterResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	userId, err := s.user.Register(ctx, req.GetName(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssogrpc.RegisterResponse{UserId: userId.String()}, nil
}

func (s *UserServerAPI) Login(ctx context.Context, req *ssogrpc.LoginRequest) (*ssogrpc.LoginResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	userId, err := s.user.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, ssoerrors.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssogrpc.LoginResponse{UserId: userId.String()}, nil
}

func (s *UserServerAPI) GetProfile(ctx context.Context, req *ssogrpc.GetProfileRequest) (*ssogrpc.GetProfileResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	userProfile, err := s.user.GetProfile(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &ssogrpc.GetProfileResponse{User: convertUserToProto(userProfile)}, nil
}

func (s *UserServerAPI) UpdateProfile(ctx context.Context, req *ssogrpc.UpdateProfileRequest) (*ssogrpc.UpdateProfileResponse, error) {
	if *req.Name == "" && *req.Email == "" && *req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "incorrect data")
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	userProfile, err := s.user.UpdateProfile(ctx, userID, req.GetName(), req.GetEmail(), *req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change user")
	}

	return &ssogrpc.UpdateProfileResponse{User: convertUserToProto(userProfile)}, nil
}

func convertUserToProto(u *models.User) (user *ssogrpc.User) {
	if u == nil {
		return nil
	}
	return &ssogrpc.User{
		Id:        u.ID.String(),
		Email:     u.Email,
		FullName:  u.FullName,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}
