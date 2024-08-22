package servergrpc

import (
	"log/slog"

	authgen "github.com/AlexBlackNn/authloyalty/commands/proto/sso/gen"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	v1 "github.com/AlexBlackNn/authloyalty/internal/handlersgrpc/grpc/v1"
	"github.com/AlexBlackNn/authloyalty/internal/services/authservice"
	rkboot "github.com/rookie-ninja/rk-boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/boot"
	"google.golang.org/grpc"
)

// App service consists all entities needed to work.
type App struct {
	Cfg         *config.Config
	Log         *slog.Logger
	Srv         *rkboot.Boot
	authService *authservice.Auth
}

// New creates App collecting grpc server and its handlers
func New(
	authService *authservice.Auth,
) (*App, error) {
	boot := rkboot.NewBoot()
	// Get grpc entry with name
	grpcEntry := boot.GetEntry("sso").(*rkgrpc.GrpcEntry)
	// Register grpc registration function
	registerAuth := registerAuthFunc(authService)
	grpcEntry.AddRegFuncGrpc(registerAuth)
	// Register grpc-gateway registration function
	grpcEntry.AddRegFuncGw(authgen.RegisterAuthHandlerFromEndpoint)
	// Bootstrap
	return &App{Srv: boot}, nil
}

func registerAuthFunc(authService *authservice.Auth) func(server *grpc.Server) {
	return func(server *grpc.Server) { // Use the provided server
		v1.Register(server, authService)
	}
}
