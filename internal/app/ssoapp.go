package ssoapp

import (
	grpcapp "sso-service/internal/app/grpc"
	"sso-service/internal/storage"
	"time"

	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	logger *zap.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	// Подключаемся к БД
	dataBase, err := storage.NewDatabase(storagePath)
	if err != nil {
		panic(err)
	}
	// TODO: authService
	// TODO: roleService
	// TODO: userService
	_ = dataBase
	gRPCApp := grpcapp.New(logger, grpcPort)

	return &App{
		GRPCServer: gRPCApp,
	}
}
