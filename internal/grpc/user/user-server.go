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
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type User interface {
	Register(ctx context.Context, fullName, email, password string) (uuid.UUID, error)
	Login(ctx context.Context, email, password string) (uuid.UUID, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, error)
	UpdateUserName(ctx context.Context, userID uuid.UUID, fulName string) error
	UpdateUserEmail(ctx context.Context, userID uuid.UUID, email string) error
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, password string) error
}

type UserServerAPI struct {
	ssogrpc.UnimplementedUserServiceServer
	user User
}

func NewUserServer(gRPC *grpc.Server, user User) {
	ssogrpc.RegisterUserServiceServer(gRPC, &UserServerAPI{user: user})
}

func (s *UserServerAPI) Register(
	ctx context.Context,
	req *ssogrpc.RegisterRequest,
) (*ssogrpc.RegisterResponse, error) {
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
		if errors.Is(err, ssoerrors.ErrInvalidEmail) {
			return nil, status.Error(codes.InvalidArgument, ssoerrors.ErrInvalidEmail.Error())
		}
		if errors.Is(err, ssoerrors.ErrPasswordTooWeak) {
			return nil, status.Error(codes.InvalidArgument, ssoerrors.ErrPasswordTooWeak.Error())
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssogrpc.RegisterResponse{UserId: &ssogrpc.UUID{Value: userId.String()}}, nil
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

	return &ssogrpc.LoginResponse{UserId: &ssogrpc.UUID{Value: userId.String()}}, nil
}

func (s *UserServerAPI) GetProfile(
	ctx context.Context,
	req *ssogrpc.GetProfileRequest,
) (*ssogrpc.GetProfileResponse, error) {
	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	user, err := s.user.GetProfile(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &ssogrpc.GetProfileResponse{
		User: &ssogrpc.User{
			Id:        &ssogrpc.UUID{Value: user.ID.String()},
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}

func (s *UserServerAPI) GetAllUsers(
	ctx context.Context,
	req *ssogrpc.GetAllUsersRequest,
) (*ssogrpc.GetAllUsersResponse, error) {
	users, err := s.user.GetAllUsers(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get users")
	}

	protoUsers := make([]*ssogrpc.User, 0, len(users))
	for _, user := range users {
		protoUsers = append(protoUsers, &ssogrpc.User{
			Id:        &ssogrpc.UUID{Value: user.ID.String()},
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		})
	}

	return &ssogrpc.GetAllUsersResponse{Users: protoUsers}, nil
}

func (s *UserServerAPI) UpdateName(
	ctx context.Context,
	req *ssogrpc.UpdateNameRequest,
) (*emptypb.Empty, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.user.UpdateUserName(ctx, userID, req.GetName())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change user")
	}

	return nil, nil
}

func (s *UserServerAPI) UpdateEmail(
	ctx context.Context,
	req *ssogrpc.UpdateEmailRequest,
) (*emptypb.Empty, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.user.UpdateUserEmail(ctx, userID, req.GetEmail())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change user")
	}

	return nil, nil
}

func (s *UserServerAPI) UpdatePassword(
	ctx context.Context,
	req *ssogrpc.UpdatePasswordRequest,
) (*emptypb.Empty, error) {
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	userID, err := uuid.Parse(req.GetUserId().GetValue())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid UUID format: %v", err)
	}

	err = s.user.UpdateUserPassword(ctx, userID, req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change user")
	}

	return nil, nil
}
