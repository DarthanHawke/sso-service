package user

import (
	"context"
	"errors"
	"fmt"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/lib/validator"
	"sso-service/internal/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserService struct {
	logger     *zap.Logger
	userManage UserManage
	userGet    UserGet
	hasher     Argon2Manager
}

type Argon2Manager interface {
	GenerateHash(data string) (string, error)
	CompareHashAndData(data, encodedHash string) (bool, error)
}

type UserManage interface {
	CreateUser(ctx context.Context, email, passwordHash, fullName string) (uuid.UUID, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserEmail(ctx context.Context, userID uuid.UUID, email string) error
	UpdateUserName(ctx context.Context, userID uuid.UUID, fullName string) error
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
}

type UserGet interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetListUsers(ctx context.Context, limit, offset int) (*[]models.User, error)
}

func NewUserService(
	logger *zap.Logger,
	userManage UserManage,
	userGet UserGet,
	hasher Argon2Manager,
) *UserService {
	return &UserService{
		logger:     logger.With(zap.String("component", "sso_service")),
		userManage: userManage,
		userGet:    userGet,
		hasher:     hasher,
	}
}

// Register регистрирует нового пользователя
func (s *UserService) Register(ctx context.Context, fullName, email, password string) (uuid.UUID, error) {
	const op = "service.user.Register"

	s.logger.With(
		zap.String("op", op),
		zap.String("email: ", email),
	)
	s.logger.Info("registering new user")

	// Валидация
	if err := validator.ValidateEmail(email); err != nil {
		s.logger.Warn("invalide email", zap.String("email", email), zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidEmail)
	}
	if err := validator.ValidatePassword(password); err != nil {
		s.logger.Warn("invalide password", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPasswordTooWeak)
	}

	// Хеширование пароля
	passwordHash, err := s.hasher.GenerateHash(password)
	if err != nil {
		s.logger.Error("hasing password", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Создание пользователя
	id, err := s.userManage.CreateUser(ctx, email, passwordHash, fullName)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			s.logger.Warn("user exisits", zap.String("email", email), zap.Error(err))

			return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
		}
		s.logger.Error("create error", zap.String("email", email), zap.Error(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	s.logger.Debug("Successfully registerd new user",
		zap.String("UserID", id.String()),
	)
	return id, nil
}

// Login - верификация пользователя
func (s *UserService) Login(ctx context.Context, email, password string) (uuid.UUID, error) {
	const op = "service.user.Login"

	s.logger.With(
		zap.String("op", op),
		zap.String("email: ", email),
	)

	s.logger.Info("login user")

	user, err := s.userGet.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			s.logger.Warn("user not found", zap.String("email", email), zap.Error(err))

			return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidCredentials)
		}
		s.logger.Warn("cannot getting user", zap.String("email", email), zap.Error(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Проверка пароля
	passwordStatus, err := s.hasher.CompareHashAndData(password, user.PasswordHash)
	if err != nil {
		s.logger.Error("error to compare passwords", zap.String("email", email), zap.Error(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidCredentials)
	}
	if !passwordStatus {
		s.logger.Warn("password incorrect", zap.String("email", email))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidCredentials)
	}

	s.logger.Debug("Successfully login user",
		zap.String("UserID", user.ID.String()),
	)
	return user.ID, nil
}

// GetProfile - получение профиля пользователя
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const op = "service.user.GetProfile"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting user")

	user, err := s.userGet.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			s.logger.Warn("user not found",
				zap.String("UserID", userID.String()),
				zap.Error(ssoerrors.ErrUserNotFound),
			)

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}
		s.logger.Warn("cannot getting user",
			zap.String("UserID", userID.String()),
			zap.Error(err),
		)

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Скрываем хеш пароля
	user.PasswordHash = ""

	s.logger.Debug("Successfully get user",
		zap.String("UserID", userID.String()),
	)
	return user, nil
}

// UpdateProfile - обновление профиля
func (s *UserService) UpdateProfile(
	ctx context.Context,
	userID uuid.UUID,
	fulName, email, password string,
) (*models.User, error) {
	const op = "service.user.UpdateProfile"

	s.logger.With(
		zap.String("op", op),
	)

	s.logger.Info("getting user")

	if fulName != "" {
		if err := s.userManage.UpdateUserName(ctx, userID, fulName); err != nil {
			if errors.Is(err, ssoerrors.ErrUserExists) {
				s.logger.Warn("user not found", zap.Error(ssoerrors.ErrUserExists))

				return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
			}
			s.logger.Warn("cannot getting user", zap.Error(err))

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
		}
	}
	if email != "" {
		// Валидация
		if err := validator.ValidateEmail(email); err != nil {
			s.logger.Warn("invalide email", zap.String("email", email), zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidEmail)
		}
		if err := s.userManage.UpdateUserEmail(ctx, userID, email); err != nil {
			if errors.Is(err, ssoerrors.ErrUserExists) {
				s.logger.Warn("user not found", zap.Error(ssoerrors.ErrUserExists))

				return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
			}
			s.logger.Warn("cannot getting user", zap.Error(err))

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
		}
	}
	if password != "" {
		// Валидация
		if err := validator.ValidatePassword(password); err != nil {
			s.logger.Warn("invalide password", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPasswordTooWeak)
		}
		// Хеширование пароля
		passwordHash, err := s.hasher.GenerateHash(password)
		if err != nil {
			s.logger.Error("hasing password", zap.Error(err))

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
		}

		if err := s.userManage.UpdateUserPassword(ctx, userID, passwordHash); err != nil {
			if errors.Is(err, ssoerrors.ErrUserExists) {
				s.logger.Warn("user not found", zap.Error(ssoerrors.ErrUserExists))

				return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
			}
			s.logger.Warn("cannot getting user", zap.Error(err))

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
		}
	}
	user, err := s.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	s.logger.Debug("Successfully update user",
		zap.String("UserID", user.ID.String()),
	)
	return user, nil
}
