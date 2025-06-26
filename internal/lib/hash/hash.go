package hash

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrHashingFailed      = errors.New("hashing failed")
	ErrInvalidHashFormat  = errors.New("invalid hash format")
	ErrIncompatibleParams = errors.New("incompatible argon2 parameters")
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type Argon2Hasher struct {
	params Argon2Params
}

func NewArgon2Hasher(params Argon2Params) *Argon2Hasher {
	// Устанавливаем значения по умолчанию, если какие-то параметры не заданы
	if params.Iterations == 0 {
		params.Iterations = 3
	}
	if params.Memory == 0 {
		params.Memory = 64 * 1024 // 64MB
	}
	if params.Parallelism == 0 {
		params.Parallelism = 4
	}
	if params.SaltLength == 0 {
		params.SaltLength = 16
	}
	if params.KeyLength == 0 {
		params.KeyLength = 32
	}

	return &Argon2Hasher{params: params}
}

// GenerateHash создает хеш с использованием Argon2id
func (h *Argon2Hasher) GenerateHash(data string) (string, error) {
	// Генерируем случайную соль
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("%w: %v", ErrHashingFailed, err)
	}

	// Генерируем хеш с помощью Argon2id
	hash := argon2.IDKey(
		[]byte(data),
		salt,
		h.params.Iterations,
		h.params.Memory,
		h.params.Parallelism,
		h.params.KeyLength,
	)

	// Кодируем хеш и соль в строку
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Формат: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.Memory,
		h.params.Iterations,
		h.params.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

// CompareHashAndData сравнивает данные с хешем
func (h *Argon2Hasher) CompareHashAndData(data, encodedHash string) (bool, error) {
	// Парсим закодированный хеш
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, ErrInvalidHashFormat
	}

	if parts[1] != "argon2id" {
		return false, ErrInvalidHashFormat
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, ErrIncompatibleParams
	}

	var params Argon2Params
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// Генерируем хеш для сравнения
	comparisonHash := argon2.IDKey(
		[]byte(data),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		uint32(len(hash)),
	)

	// Сравниваем хеши с защитой от атак по времени
	if subtle.ConstantTimeCompare(hash, comparisonHash) == 1 {
		return true, nil
	}

	return false, nil
}

func (h *Argon2Hasher) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
