package main

import (
	"os"
	"os/signal"
	ssoapp "sso-service/internal/app"
	"sso-service/internal/config"
	"sso-service/internal/lib/hash"
	"sso-service/internal/lib/jwt"
	"sso-service/internal/lib/logger"
	"syscall"

	"go.uber.org/zap"
)

const (
	example     = "example"
	development = "development"
	production  = "production"
)

func newArgon2Params() hash.Argon2Params {
	return hash.Argon2Params{
		Memory:      64 * 1024, // 64MB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func main() {
	// Подключаем логгер Zap в настраиваемой конфигурации (>=LevelInfo выводит в консоль, >=DebugLevel в файл)
	log := logger.SetupLogger()
	defer log.Sync()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = example // значение по умолчанию (например, production)
	}
	log.Info("ENV applied", zap.String("env", env))

	// Загружаем конфиг
	cfg, err := config.LoadConfig(env)
	if err != nil {
		log.Error("Failed to load config", zap.Error(err))
		return
	}

	// JWT
	jwtManager, err := jwt.NewTokenGenerator(
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.Issuer,
		cfg.PrivateKeyPath,
		cfg.PublicKeyPath,
	)
	if err != nil {
		log.Error("Failed to init JWT", zap.Error(err))
		return
	}

	// Hash
	hasher := hash.NewArgon2Hasher(newArgon2Params())

	// Канал для graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// инициализируем сервер sso
	application := ssoapp.New(log, cfg.GRPSServer.Port, cfg.DataBase.DSN(), jwtManager, hasher)

	// запускаем gRPC сервер
	go application.GRPCServer.MustRun()

	// Ожидание сигнала завершения
	<-done

	// Graceful shutdown
	application.GRPCServer.Stop()
	log.Info("Gracefully stopped")
}
