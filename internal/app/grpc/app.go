package grpcapp

import (
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"sso/internal/config"
	authtransport "sso/internal/grpc_transport/auth"
	authservice "sso/internal/services/auth_service"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	cfg        *config.Config
}

// New creates new gRPC server app
func New(
	log *slog.Logger,
	authService authservice.AuthorizationInterface,
	cfg *config.Config,
) *App {
	gRPCServer := grpc.NewServer()
	authtransport.Register(gRPCServer, authService)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		cfg:        cfg,
	}
}

func (a *App) Run() error {
	const op = "APP LAYER: grpcapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.cfg.GRPC.Port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("grpc_transport server is running", slog.String("address", l.Addr().String()))
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.log.With(
		slog.String("op", op),
		slog.Int("port", a.cfg.GRPC.Port),
	).Info("stopping grpc_transport server", slog.Int("port", a.cfg.GRPC.Port))
	a.gRPCServer.GracefulStop()
}
