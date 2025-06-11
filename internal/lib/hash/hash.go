package hash

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrHashingFailed = errors.New("password hashing failed")
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password, hash string) bool
}

type BcryptHasher struct {
	cost int // Сложность хеширования (4-31)
}

func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{cost: cost}
}

// Hash создает bcrypt-хеш пароля
func (h *BcryptHasher) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", ErrHashingFailed
	}
	return string(hashedBytes), nil
}

// Compare проверяет пароль против хеша
func (h *BcryptHasher) Compare(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
