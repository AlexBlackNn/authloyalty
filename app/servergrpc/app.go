package servergrpc

import (
	"context"
	"fmt"
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
	"net"
)

type App struct {
	GRPCSrv *appInner
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {

	// TODO: seems to need factory here
	//storage, err := postgres.New(cfg.StoragePath) //uncomment for postgres
	storage, err := patroni.New(cfg)
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
	registerAuth := registerGreeterFunc(log, cfg)
	grpcEntry.AddRegFuncGrpc(registerAuth)
	// Register grpc-gateway registration function
	grpcEntry.AddRegFuncGw(authgen.RegisterAuthHandlerFromEndpoint)

	// Bootstrap
	boot.Bootstrap(context.Background())

	// Wait for shutdown sig
	boot.WaitForShutdownSig(context.Background())

	grpcApp := NewAppInner(log, authService, cfg)
	return &App{GRPCSrv: grpcApp}, nil
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

type appInner struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	cfg        *config.Config
}

// New creates new gRPC server app
func NewAppInner(
	log *slog.Logger,
	authService authservice.AuthorizationInterface,
	cfg *config.Config,
) *appInner {
	gRPCServer := grpc.NewServer()
	authtransport.Register(gRPCServer, authService)
	return &appInner{
		log:        log,
		gRPCServer: gRPCServer,
		cfg:        cfg,
	}
}

func (a *appInner) Run() error {
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

func (a *appInner) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *appInner) Stop() {
	const op = "grpcapp.Stop"
	a.log.With(
		slog.String("op", op),
		slog.Int("port", a.cfg.GRPC.Port),
	).Info("stopping grpc_transport server", slog.Int("port", a.cfg.GRPC.Port))
	a.gRPCServer.GracefulStop()
}
