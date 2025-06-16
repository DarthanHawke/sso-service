package user

import (
	"context"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/lib/hash"
	"sso-service/internal/models"

	"go.uber.org/zap"
)

type UserService struct {
	logger     *zap.Logger
	userManage UserManage
	userGet    UserGet
	hasher     hash.Argon2Manager
}

type UserManage interface {
	CreateUser(ctx context.Context, email, passwordHash, fullName string) ([]uint8, error)
	DeleteUser(ctx context.Context, userID []uint8) error
	UpdateUser(ctx context.Context, userID []uint8, email, fullName string) error
	UpdatePassword(ctx context.Context, userID []uint8, newPasswordHash string) error
}

type UserGet interface {
	GetUserByID(ctx context.Context, userID []uint8) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetListUsers(ctx context.Context, limit, offset int) ([]models.User, error)
}

func NewUserService(
	logger *zap.Logger,
	userManage UserManage,
	userGet UserGet,
	hasher hash.Argon2Manager,
) *UserService {
	return &UserService{
		logger:     logger,
		userManage: userManage,
		userGet:    userGet,
		hasher:     hasher,
	}
}

func (s *UserService) Register(ctx context.Context, fullName, email, password string) ([]uint8, error) {
	const op = "service.role.Register"

	s.logger.With(
		zap.String("op", op),
		zap.String("email: ", email),
	)

	s.logger.Info("registering new user")

	// Валидация
	/*if err := validator.ValidateEmail(email); err != nil {
		s.logger.Info("invalide email", zap.Error(err))

		return errId, fmt.Errorf("%s: %w", op, err)
	}

	if err := validator.ValidatePassword(password); err != nil {
		s.logger.Info("invalide password", zap.Error(err))

		return errId, fmt.Errorf("%s: %w", op, err)
	}*/

	// Хеширование пароля
	passwordHash, err := s.hasher.GenerateHash(password)
	if err != nil {
		s.logger.Error("hasing password", zap.Error(err))

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Создание пользователя
	id, err := s.userManage.CreateUser(ctx, email, passwordHash, fullName)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			s.logger.Error("user exisits", zap.Error(ssoerrors.ErrUserExists))

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
		}
		s.logger.Error("create error", zap.Error(err))

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}
	return id, nil
}

// Аутентификация пользователя
func (s *UserService) Login(ctx context.Context, email, password string) ([]uint8, error) {
	const op = "service.role.Login"

	s.logger.With(
		zap.String("op", op),
		zap.String("email: ", email),
	)

	s.logger.Info("login user")

	user, err := s.userGet.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			s.logger.Warn("user not found", zap.Error(ssoerrors.ErrUserNotFound))

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidCredentials)
		}
		s.logger.Warn("cannot getting user", zap.Error(err))

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Проверка пароля
	passwordStatus, err := s.hasher.CompareHashAndData(password, user.PasswordHash)
	if err != nil {
		s.logger.Warn("error to compare passwords", zap.Error(err))

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidCredentials)
	}
	if !passwordStatus {
		s.logger.Warn("password incorrect")

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidCredentials)
	}
	return user.ID, nil
}

// Получение профиля пользователя
func (s *UserService) GetProfile(ctx context.Context, userID []uint8) (models.User, error) {
	const op = "service.role.GetProfile"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting user")

	user, err := s.userGet.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			s.logger.Warn("user not found", zap.Error(ssoerrors.ErrUserNotFound))

			return models.User{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}
		s.logger.Warn("cannot getting user", zap.Error(err))

		return models.User{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Скрываем хеш пароля
	user.PasswordHash = ""
	return user, nil
}

// Обновление профиля
func (s *UserService) UpdateProfile(
	ctx context.Context,
	userID []uint8,
	fulName, email, password string,
) (models.User, error) {
	const op = "service.role.UpdateProfile"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting user")

	if err := s.userManage.UpdateUser(ctx, userID, email, fulName); err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			s.logger.Warn("user not found", zap.Error(ssoerrors.ErrUserExists))

			return models.User{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
		}
		s.logger.Warn("cannot getting user", zap.Error(err))

		return models.User{}, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}
	return s.GetProfile(ctx, userID)
}
