package apperr

import "fmt"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Details)
	}
	return e.Message
}

// WithDetails добавляет детали к ошибке (создаёт копию)
func (e *AppError) WithDetails(details any) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// Ошибки пользователя (users)
var (
	// ErrUserNotFound — пользователь не найден
	ErrUserNotFound = &AppError{Code: "USER_NOT_FOUND", Message: "user not found"}

	// ErrUserExists — пользователь с таким email уже существует
	ErrUserExists = &AppError{Code: "USER_EXISTS", Message: "user with this email already exists"}

	// ErrUserDisabled — пользователь заблокирован
	ErrUserDisabled = &AppError{Code: "USER_DISABLED", Message: "user is disabled"}

	// ErrEmailNotVerified — email не подтверждён
	ErrEmailNotVerified = &AppError{Code: "EMAIL_NOT_VERIFIED", Message: "email not verified"}

	// ErrEmailAlreadyVerified — email уже подтверждён
	ErrEmailAlreadyVerified = &AppError{Code: "EMAIL_ALREADY_VERIFIED", Message: "email already verified"}
)

// Ошибки аутентификации (credentials)
var (
	// ErrInvalidCredentials — неверный email или пароль
	ErrInvalidCredentials = &AppError{Code: "INVALID_CREDENTIALS", Message: "invalid email or password"}

	// ErrInvalidOldPassword — неверный старый пароль при смене
	ErrInvalidOldPassword = &AppError{Code: "INVALID_OLD_PASSWORD", Message: "invalid old password"}

	// ErrPasswordTooWeak — пароль не соответствует требованиям сложности
	ErrPasswordTooWeak = &AppError{Code: "PASSWORD_TOO_WEAK", Message: "password does not meet complexity requirements"}

	// ErrInvalidEmail — неверный формат email
	ErrInvalidEmail = &AppError{Code: "INVALID_EMAIL", Message: "invalid email format"}
)

// Ошибки токенов (tokens)
var (
	// ErrTokenExpired — токен истёк
	ErrTokenExpired = &AppError{Code: "TOKEN_EXPIRED", Message: "token expired"}

	// ErrTokenRevoked — токен отозван (в blacklist)
	ErrTokenRevoked = &AppError{Code: "TOKEN_REVOKED", Message: "token revoked"}

	// ErrTokenInvalid — токен недействителен (подпись, формат)
	ErrTokenInvalid = &AppError{Code: "TOKEN_INVALID", Message: "token invalid"}

	// ErrRefreshTokenNotFound — refresh-токен не найден (сессия удалена или истекла)
	ErrRefreshTokenNotFound = &AppError{Code: "REFRESH_TOKEN_NOT_FOUND", Message: "refresh token not found"}
)

// Ошибки сессий (sessions)
var (
	// ErrSessionNotFound — сессия не найдена
	ErrSessionNotFound = &AppError{Code: "SESSION_NOT_FOUND", Message: "session not found"}
)

// Ошибки верификации (verification)
var (
	// ErrVerificationTokenInvalid — недействительный токен верификации
	ErrVerificationTokenInvalid = &AppError{Code: "VERIFICATION_TOKEN_INVALID", Message: "verification token invalid or expired"}

	// ErrResetTokenInvalid — недействительный токен сброса пароля
	ErrResetTokenInvalid = &AppError{Code: "RESET_TOKEN_INVALID", Message: "reset token invalid or expired"}
)

// Внутренние ошибки
var (
	// ErrInternal — внутренняя ошибка сервера
	ErrInternal = &AppError{Code: "INTERNAL_ERROR", Message: "internal server error"}
)
