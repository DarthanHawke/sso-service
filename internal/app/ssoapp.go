package ssoapp

import (
	grpcapp "sso-service/internal/app/grpc"
	"sso-service/internal/lib/hash"
	"sso-service/internal/lib/jwt"
	"sso-service/internal/service/role"
	"sso-service/internal/service/session"
	"sso-service/internal/service/user"
	"sso-service/internal/storage"

	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	logger *zap.Logger,
	grpcPort int,
	storagePath string,
	jwtManager jwt.JWTManager,
	hasher hash.Argon2Manager,
) *App {
	// Подключаемся к БД
	dataBase, err := storage.NewDatabase(storagePath)
	if err != nil {
		panic(err)
	}
	roleDataBase := storage.NewRoleDataBase(dataBase)
	sessionDataBase := storage.NewSessionDataBase(dataBase)
	userDataBase := storage.NewUserDataBase(dataBase)
	roleService := role.NewRoleService(logger, roleDataBase, roleDataBase)
	sessionService := session.NewSessionService(logger, sessionDataBase, sessionDataBase, jwtManager, hasher)
	userService := user.NewUserService(logger, userDataBase, userDataBase, hasher)

	_ = dataBase
	gRPCApp := grpcapp.New(logger, grpcPort, roleService, sessionService, userService)

	return &App{
		GRPCServer: gRPCApp,
	}
}
