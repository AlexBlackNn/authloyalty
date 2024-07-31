package servergrpc

import (
	"context"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	authtransport "github.com/AlexBlackNn/authloyalty/internal/grpc_transport/auth"
	authservice "github.com/AlexBlackNn/authloyalty/internal/services/auth_service"
	"github.com/AlexBlackNn/authloyalty/pkg/broker"
	patroni "github.com/AlexBlackNn/authloyalty/pkg/storage/patroni"
	redis "github.com/AlexBlackNn/authloyalty/pkg/storage/redis-sentinel"
	authgen "github.com/AlexBlackNn/authloyalty/protos/proto/sso/gen"
	rkboot "github.com/rookie-ninja/rk-boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/boot"
	"google.golang.org/grpc"
	"log/slog"
)

type App struct {
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {
	boot := rkboot.NewBoot()
	// Get grpc entry with name
	grpcEntry := boot.GetEntry("sso").(*rkgrpc.GrpcEntry)
	// Register grpc registration function
	registerAuth := registerGreeterFunc(log, cfg)
	grpcEntry.AddRegFuncGrpc(registerAuth)
	// Register grpc-gateway registration function
	grpcEntry.AddRegFuncGw(authgen.RegisterAuthHandlerFromEndpoint)
	// Bootstrap
	boot.Bootstrap(context.Background())
	//grpcApp := NewAppInner(log, authService, cfg)
	return &App{}, nil
}

func registerGreeterFunc(log *slog.Logger, cfg *config.Config) func(server *grpc.Server) {
	return func(server *grpc.Server) { // Use the provided server
		storage, err := patroni.New(cfg) // Use cfg from the closure
		if err != nil {
			log.Error("Failed to create storage", "error", err) // Use log from the closure
			panic(err)
		}
		tokenCache := redis.New(cfg) // Use cfg from the closure

		kafkaURL := "localhost:9094"
		schemaRegistryURL := "http://localhost:8081"

		producer, err := broker.NewProducer(kafkaURL, schemaRegistryURL)
		authService := authservice.New(cfg, log, storage, tokenCache, producer) // Use log and cfg from the closure
		authtransport.Register(server, authService)                             // Register the service on the provided server
	}
}
