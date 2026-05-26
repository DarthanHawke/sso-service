package main

import (
	"authn-gate-service/internal/config"
	"authn-gate-service/internal/lib/logger"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

const (
	example     = "example"
	development = "development"
	production  = "production"
)

// RunMigrations применяет все pending миграции
func main() {
	// Подключаем логгер Zap в настраиваемой конфигурации (>=LevelInfo выводит в консоль, >=DebugLevel в файл)
	log := logger.SetupLogger()
	defer log.Sync()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = example
	}
	log.Info("ENV applied", zap.String("env", env))

	// Загружаем конфиг
	cfg, err := config.LoadConfig(env)
	if err != nil {
		log.Error("Failed to load config", zap.Error(err))
		return
	}

	// Инициализируем мигратор
	mgrt, err := migrate.New("file://./migrations", cfg.DataBase.DSN())
	if err != nil {
		log.Error("migrate init failed:", zap.Error(err))
	}
	defer mgrt.Close()

	// Применяем миграции
	if err := mgrt.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error("migrate up failed:", zap.Error(err))
	}

}
