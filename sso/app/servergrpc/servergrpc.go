package servergrpc

import (
	"fmt"
	"go.opentelemetry.io/otel"
	"log/slog"
	"net"

	"github.com/AlexBlackNn/authloyalty/sso/internal/config"
	v1 "github.com/AlexBlackNn/authloyalty/sso/internal/handlersgrpc/grpc/v1"
	"github.com/AlexBlackNn/authloyalty/sso/internal/interceptors"
	"github.com/AlexBlackNn/authloyalty/sso/internal/services/authservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// App service consists all entities needed to work.
type App struct {
	Cfg         *config.Config
	Log         *slog.Logger
	Server      *grpc.Server
	authService *authservice.Auth
}

// New creates App collecting grpc server and its handlers
func New(
	cfg *config.Config,
	log *slog.Logger,
	authService *authservice.Auth,
) (*App, error) {
	// Создаем gRPC сервер с опциями
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.NewTracing(otel.Tracer("sso service")).GetInterceptor()),
	)

	// Регистрируем gRPC сервисы
	v1.Register(server, authService)

	// Включаем gRPC Reflection для удобства тестирования
	reflection.Register(server)

	return &App{
		Cfg:         cfg,
		Log:         log,
		Server:      server,
		authService: authService,
	}, nil
}

// Start starts gRPC server
func (a *App) Start() error {
	// Запускаем gRPC сервер на указанном порту
	a.Log.Info("Starting gRPC server", slog.String("address", "44044"))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.Cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("%s: %w", "app start error", err)
	}
	if err := a.Server.Serve(l); err != nil {
		return fmt.Errorf("app start error: %w", err)
	}
	return nil
}
