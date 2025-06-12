package user

import (
	"context"
	"errors"
	ssoerrors "sso-service/internal/lib/errors"
	"sso-service/internal/lib/hash"
	"sso-service/internal/lib/validator"
	"sso-service/internal/models"

	"go.uber.org/zap"
)

const (
	errId = -1
)

type UserService struct {
	logger     *zap.Logger
	userManage UserManage
	userGet    UserGet
	hasher     hash.PasswordHasher
}

type UserManage interface {
	CreateUser(ctx context.Context, email, passwordHash, fullName string) (int64, error)
	DeleteUser(ctx context.Context, userID string) error
	UpdateUser(ctx context.Context, userID string, email, fullName string) error
	UpdatePassword(ctx context.Context, userID, newPasswordHash string) error
}

type UserGet interface {
	GetUserByID(ctx context.Context, userID string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetListUsers(ctx context.Context, limit, offset int) ([]models.User, error)
}

func NewUserService(
	logger *zap.Logger,
	userManage UserManage,
	userGet UserGet,
	hasher hash.PasswordHasher,
) *UserService {
	return &UserService{
		logger:     logger,
		userManage: userManage,
		userGet:    userGet,
		hasher:     hasher,
	}
}

func (s *UserService) Register(ctx context.Context, fullName, email, password string) (int64, error) {
	//TO DO: use logger

	// Валидация
	if err := validator.ValidateEmail(email); err != nil {
		return errId, err
	}

	if err := validator.ValidatePassword(password); err != nil {
		return errId, err
	}

	// Хеширование пароля
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return errId, ssoerrors.ErrInternal
	}

	// Создание пользователя
	id, err := s.userManage.CreateUser(ctx, email, passwordHash, fullName)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			return errId, ssoerrors.ErrUserExists
		}
		return errId, ssoerrors.ErrInternal
	}
	return id, nil
}

// Аутентификация пользователя
func (s *UserService) Login(ctx context.Context, email, password string) (int64, error) {
	//TO DO: use logger

	user, err := s.userGet.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			return errId, ssoerrors.ErrInvalidCredentials
		}
		return errId, ssoerrors.ErrInternal
	}

	// Проверка пароля
	if !s.hasher.Compare(password, user.PasswordHash) {
		return errId, ssoerrors.ErrInvalidCredentials
	}
	return user.ID, nil
}

// Получение профиля пользователя
func (s *UserService) GetProfile(ctx context.Context, userID string) (models.User, error) {
	//TO DO: use logger

	user, err := s.userGet.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			return models.User{}, ssoerrors.ErrUserNotFound
		}
		return models.User{}, ssoerrors.ErrInternal
	}

	// Скрываем хеш пароля
	user.PasswordHash = ""
	return user, nil
}

// Обновление профиля
func (s *UserService) UpdateProfile(
	ctx context.Context,
	fulName, userID, email, password string,
) (models.User, error) {
	//TO DO: use logger

	if err := s.userManage.UpdateUser(ctx, userID, email, fulName); err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			return models.User{}, ssoerrors.ErrUserExists
		}
		return models.User{}, ssoerrors.ErrInternal
	}
	return s.GetProfile(ctx, userID)
}
