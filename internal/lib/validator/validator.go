package validator

import (
	ssoerrors "sso-service/internal/lib/errors"

	"regexp"
	"unicode"
)

// Проверка email с помощью регулярного выражения
func ValidateEmail(email string) error {
	// Регулярка для стандартной проверки email
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	if !matched {
		return ssoerrors.ErrInvalidEmail
	}
	return nil
}

// Проверка сложности пароля
func ValidatePassword(password string) error {
	var (
		hasMinLen  = len(password) >= 8
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasMinLen || !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ssoerrors.ErrPasswordTooWeak
	}

	return nil
}
