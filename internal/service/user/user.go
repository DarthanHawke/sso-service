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
	errIdInt    = -1
	errIdString = ""
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

func (s *UserService) Register(ctx context.Context, email, password, fullName string) (int64, error) {
	//TO DO: use logger

	// Валидация
	if err := validator.ValidateEmail(email); err != nil {
		return errIdInt, err
	}

	if err := validator.ValidatePassword(password); err != nil {
		return errIdInt, err
	}

	// Хеширование пароля
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return errIdInt, ssoerrors.ErrInternal
	}

	// Создание пользователя
	id, err := s.userManage.CreateUser(ctx, email, passwordHash, fullName)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			return errIdInt, ssoerrors.ErrUserExists
		}
		return errIdInt, ssoerrors.ErrInternal
	}
	return id, nil
}

// Аутентификация пользователя
func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	//TO DO: use logger

	user, err := s.userGet.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			return errIdString, ssoerrors.ErrInvalidCredentials
		}
		return errIdString, ssoerrors.ErrInternal
	}

	// Проверка пароля
	if !s.hasher.Compare(password, user.PasswordHash) {
		return errIdString, ssoerrors.ErrInvalidCredentials
	}
	return user.ID, nil
}

// Получение профиля пользователя
func (s *UserService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	//TO DO: use logger

	user, err := s.userGet.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ssoerrors.ErrUserNotFound) {
			return nil, ssoerrors.ErrUserNotFound
		}
		return nil, ssoerrors.ErrInternal
	}

	// Скрываем хеш пароля
	user.PasswordHash = ""
	return &user, nil
}

// Обновление профиля
func (s *UserService) UpdateProfile(
	ctx context.Context,
	userID, email, password, fulName string,
) error {
	//TO DO: use logger

	if err := s.userManage.UpdateUser(ctx, userID, email, fulName); err != nil {
		if errors.Is(err, ssoerrors.ErrUserExists) {
			return ssoerrors.ErrUserExists
		}
		return ssoerrors.ErrInternal
	}

	return nil
}
