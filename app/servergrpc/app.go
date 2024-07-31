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

	storage, err := patroni.New(cfg) // Use cfg from the closure
	if err != nil {
		return nil, err
	}
	tokenCache := redis.New(cfg)

	producer, err := broker.NewProducer(cfg.Kafka.KafkaURL, cfg.Kafka.SchemaRegistryURL)
	if err != nil {
		return nil, err
	}
	authService := authservice.New(cfg, log, storage, tokenCache, producer)

	boot := rkboot.NewBoot()
	// Get grpc entry with name
	grpcEntry := boot.GetEntry("sso").(*rkgrpc.GrpcEntry)
	// Register grpc registration function
	registerAuth := registerAuthFunc(authService)
	grpcEntry.AddRegFuncGrpc(registerAuth)
	// Register grpc-gateway registration function
	grpcEntry.AddRegFuncGw(authgen.RegisterAuthHandlerFromEndpoint)
	// Bootstrap
	boot.Bootstrap(context.Background())
	return &App{}, nil
}

func registerAuthFunc(authService *authservice.Auth) func(server *grpc.Server) {
	return func(server *grpc.Server) { // Use the provided server
		authtransport.Register(server, authService)
	}
}
