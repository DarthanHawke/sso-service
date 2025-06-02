package main

import (
	"os"
	"sso-service/internal/config"
	"sso-service/internal/logger"

	"go.uber.org/zap"
)

const (
	example     = "example"
	development = "development"
	production  = "production"
)

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

	_ = cfg
	// TODO: init app

	// TODO: start grps
}
