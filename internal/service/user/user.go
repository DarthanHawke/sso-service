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

type Argon2Manager interface {
	GenerateHash(data string) (string, error)
	CompareHashAndData(data, encodedHash string) (bool, error)
}

type UserManage interface {
	CreateUser(ctx context.Context, email, passwordHash, fullName string) (uuid.UUID, error)
	DeleteUserSoft(ctx context.Context, userID uuid.UUID) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type UserUpdate interface {
	UpdateUserEmail(ctx context.Context, userID uuid.UUID, email string) error
	UpdateUserName(ctx context.Context, userID uuid.UUID, fullName string) error
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
}

type UserGet interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetListUsers(ctx context.Context, limit, offset int) ([]models.User, error)
}

type UserService struct {
	logger     *zap.Logger
	userManage UserManage
	userUpdate UserUpdate
	userGet    UserGet
	hasher     Argon2Manager
}

func NewUserService(
	logger *zap.Logger,
	userManage UserManage,
	userUpdate UserUpdate,
	userGet UserGet,
	hasher Argon2Manager,
) *UserService {
	return &UserService{
		logger:     logger.With(zap.String("component", "sso_service")),
		userManage: userManage,
		userUpdate: userUpdate,
		userGet:    userGet,
		hasher:     hasher,
	}
}

// Register регистрирует нового пользователя
func (s *UserService) Register(ctx context.Context, fullName, email, password string) (uuid.UUID, error) {
	const op = "service.user.Register"

	logger := s.logger.With(
		zap.String("op", op),
	)
	logger.Info("registering new user")

	// Валидация
	if err := validator.ValidateEmail(email); err != nil {
		logger.Warn("invalide email", zap.String("email", email), zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidEmail)
	}
	if err := validator.ValidatePassword(password); err != nil {
		logger.Warn("invalide password", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrPasswordTooWeak)
	}

	// Хеширование пароля
	passwordHash, err := s.hasher.GenerateHash(password)
	if err != nil {
		logger.Error("hasing password", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Создание пользователя
	id, err := s.userManage.CreateUser(ctx, email, passwordHash, fullName)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			logger.Warn("user exisits", zap.String("email", email), zap.Error(err))

			return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
		}
		logger.Error("create error", zap.String("email", email), zap.Error(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully registerd new user",
		zap.String("UserID", id.String()),
	)
	return id, nil
}

// Login - верификация пользователя
func (s *UserService) Login(ctx context.Context, email, password string) (uuid.UUID, error) {
	const op = "service.user.Login"

	logger := s.logger.With(
		zap.String("op", op),
	)

	logger.Info("login user")

	user, err := s.userGet.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			logger.Warn("user not found", zap.String("email", email), zap.Error(err))

			return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidCredentials)
		}
		logger.Warn("cannot getting user", zap.String("email", email), zap.Error(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Проверка пароля
	passwordStatus, err := s.hasher.CompareHashAndData(password, user.PasswordHash)
	if err != nil {
		logger.Error("error to compare passwords", zap.String("email", email), zap.Error(err))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidCredentials)
	}
	if !passwordStatus {
		logger.Warn("password incorrect", zap.String("email", email))

		return uuid.Nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidCredentials)
	}

	logger.Debug("Successfully login user",
		zap.String("UserID", user.ID.String()),
	)
	return user.ID, nil
}

// GetProfile - получение профиля пользователя
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const op = "service.user.GetProfile"

	logger := s.logger.With(
		zap.String("op", op),
	)

	logger.Info("getting user")

	user, err := s.userGet.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			logger.Warn("user not found",
				zap.String("UserID", userID.String()),
				zap.Error(ssoerrors.ErrUserNotFound),
			)

			return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrUserNotFound)
		}
		logger.Warn("cannot getting user",
			zap.String("UserID", userID.String()),
			zap.Error(err),
		)

		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Скрываем хеш пароля
	user.PasswordHash = ""

	logger.Debug("Successfully get user",
		zap.String("UserID", userID.String()),
	)
	return user, nil
}

// UpdateUserName - обновление имни профиля
func (s *UserService) UpdateUserName(
	ctx context.Context,
	userID uuid.UUID,
	fulName string,
) error {
	const op = "service.user.UpdateUserName"

	logger := s.logger.With(
		zap.String("op", op),
	)

	logger.Info("updating user")

	if err := s.userUpdate.UpdateUserName(ctx, userID, fulName); err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			logger.Warn("user not found", zap.Error(ssoerrors.ErrUserExists))

			return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
		}
		logger.Warn("cannot update user", zap.Error(err))

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully update user",
		zap.String("UserID", userID.String()),
	)
	return nil
}

// UpdateUserEmail - обновление email профиля
func (s *UserService) UpdateUserEmail(
	ctx context.Context,
	userID uuid.UUID,
	email string,
) error {
	const op = "service.user.UpdateUserEmail"

	logger := s.logger.With(
		zap.String("op", op),
	)

	logger.Info("updating user")

	// Валидация
	if err := validator.ValidateEmail(email); err != nil {
		logger.Warn("invalide email", zap.String("email", email), zap.Error(err))
		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInvalidEmail)
	}
	if err := s.userUpdate.UpdateUserEmail(ctx, userID, email); err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			logger.Warn("user not found", zap.Error(ssoerrors.ErrUserExists))

			return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
		}
		logger.Warn("cannot update user", zap.Error(err))

		return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	logger.Debug("Successfully update user",
		zap.String("UserID", userID.String()),
	)
	return nil
}

// UpdateUserPassword - обновление пароля профиля
func (s *UserService) UpdateUserPassword(
	ctx context.Context,
	userID uuid.UUID,
	password string,
) error {
	const op = "service.user.UpdateUserPassword"

	logger := s.logger.With(
		zap.String("op", op),
	)

	logger.Info("updating user")

	if password != "" {
		// Валидация
		if err := validator.ValidatePassword(password); err != nil {
			logger.Warn("invalide password", zap.Error(err))
			return fmt.Errorf("%s: %w", op, ssoerrors.ErrPasswordTooWeak)
		}
		// Хеширование пароля
		passwordHash, err := s.hasher.GenerateHash(password)
		if err != nil {
			logger.Error("hasing password", zap.Error(err))

			return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
		}

		if err := s.userUpdate.UpdateUserPassword(ctx, userID, passwordHash); err != nil {
			if errors.Is(err, ssoerrors.ErrUserExists) {
				logger.Warn("user not found", zap.Error(ssoerrors.ErrUserExists))

				return fmt.Errorf("%s: %w", op, ssoerrors.ErrUserExists)
			}
			logger.Warn("cannot update user", zap.Error(err))

			return fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
		}
	}

	logger.Debug("Successfully update user",
		zap.String("UserID", userID.String()),
	)
	return nil
}

// Добавляем в UserService
func (s *UserService) GetAllUsers(
	ctx context.Context,
	limit, offset int,
) ([]models.User, error) {
	const op = "service.user.GetAllUsers"

	logger := s.logger.With(
		zap.String("op", op),
	)

	logger.Info("getting all users")

	users, err := s.userGet.GetListUsers(ctx, limit, offset)
	if err != nil {
		logger.Error("failed to get users", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, ssoerrors.ErrInternal)
	}

	// Скрываем хеши паролей
	for i := range users {
		(users)[i].PasswordHash = ""
	}

	logger.Debug("Successfully got users list",
		zap.Int("count", len(users)),
	)
	return users, nil
}
