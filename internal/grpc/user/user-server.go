package user

import (
	"context"
	"errors"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/models"

	ssov1 "github.com/DarthanHawke/protos-payment-system/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type User interface {
	Register(ctx context.Context, fullName, email, password string) ([]uint8, error)
	Login(ctx context.Context, email, password string) ([]uint8, error)
	GetProfile(ctx context.Context, userID []uint8) (models.User, error)
	UpdateProfile(ctx context.Context, userID []uint8, fulName, email, password string) (models.User, error)
}

type UserServerAPI struct {
	ssov1.UnimplementedUserServiceServer
	user User
}

func NewUserServer(gRPC *grpc.Server, user User) {
	ssov1.RegisterUserServiceServer(gRPC, &UserServerAPI{user: user})
}

func (s *UserServerAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
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

	return &ssov1.RegisterResponse{UserId: string(userId)}, nil
}

func (s *UserServerAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
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

	return &ssov1.LoginResponse{UserId: string(userId)}, nil
}

func (s *UserServerAPI) GetProfile(ctx context.Context, req *ssov1.GetProfileRequest) (*ssov1.GetProfileResponse, error) {
	userProfile, err := s.user.GetProfile(ctx, []uint8(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &ssov1.GetProfileResponse{User: convertUserToProto(&userProfile)}, nil
}

func (s *UserServerAPI) UpdateProfile(ctx context.Context, req *ssov1.UpdateProfileRequest) (*ssov1.UpdateProfileResponse, error) {
	if *req.Name == "" && *req.Email == "" && *req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "incorrect data")
	}

	userProfile, err := s.user.UpdateProfile(ctx, []uint8(req.GetUserId()), req.GetName(), req.GetEmail(), *req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change user")
	}

	return &ssov1.UpdateProfileResponse{User: convertUserToProto(&userProfile)}, nil
}

func convertUserToProto(u *models.User) (user *ssov1.User) {
	if u == nil {
		return nil
	}
	return &ssov1.User{
		Id:        string(u.ID),
		Email:     u.Email,
		FullName:  u.FullName,
		Roles:     u.Roles,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}
