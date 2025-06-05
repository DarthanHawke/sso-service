package ssoapp

import (
	grpcapp "sso-service/internal/app/grpc"
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
	// TODO: storage
	// TODO: authService
	// TODO: roleService
	// TODO: userService

	gRPCApp := grpcapp.New(logger, grpcPort)

	return &App{
		GRPCServer: gRPCApp,
	}
}
