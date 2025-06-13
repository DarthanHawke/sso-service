package grpcapp

import (
	"fmt"
	"net"
	grpcrole "sso-service/internal/grpc/role"
	grpcsession "sso-service/internal/grpc/session"
	grpcuser "sso-service/internal/grpc/user"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	logger     *zap.Logger
	gRPCServer *grpc.Server
	gRPCport   int
}

func New(
	logger *zap.Logger,
	gRPCport int,
	roleService grpcrole.Role,
	sessionService grpcsession.Session,
	userService grpcuser.User,
) *App {
	gRPCServer := grpc.NewServer()
	grpcrole.NewRoleServer(gRPCServer, roleService)
	grpcsession.NewSessionServer(gRPCServer, sessionService)
	grpcuser.NewUserServer(gRPCServer, userService)
	return &App{
		logger:     logger,
		gRPCServer: gRPCServer,
		gRPCport:   gRPCport,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	a.logger.Info("Starting gRPC server", zap.Int("port", a.gRPCport))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.gRPCport))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	a.logger.Info("gRPC server is running", zap.String("addres", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (a *App) Stop() {
	a.logger.Info("Stopping gRPC server")
	a.gRPCServer.GracefulStop()
}
