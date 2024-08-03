package app

import (
	"fmt"
	"github.com/AlexBlackNn/authloyalty/app/servergrpc"
	"github.com/AlexBlackNn/authloyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
	"github.com/AlexBlackNn/authloyalty/pkg/broker"
	patroni "github.com/AlexBlackNn/authloyalty/pkg/storage/patroni"
	redis "github.com/AlexBlackNn/authloyalty/pkg/storage/redissentinel"
)

type App struct {
	ServerHttp *serverhttp.App
	ServerGrpc *servergrpc.App
}

func New() (*App, error) {
	cfg := config.New()
	log := logger.New(cfg.Env)

	userStorage, err := patroni.New(cfg)
	if err != nil {
		return nil, err
	}

	tokenStorage := redis.New(cfg)

	producer, kafkaResponseChan, err := broker.NewProducer(cfg.Kafka.KafkaURL, cfg.Kafka.SchemaRegistryURL)

	go func() {
		for kafkaResponse := range kafkaResponseChan {
			fmt.Println("http", kafkaResponse)
		}
	}()

	// http server
	serverHttp, err := serverhttp.New(cfg, log, userStorage, tokenStorage, producer)
	if err != nil {
		return nil, err
	}

	// grpc server
	serverGrpc, err := servergrpc.New(cfg, log, userStorage, tokenStorage, producer)
	if err != nil {
		return nil, err
	}
	return &App{ServerHttp: serverHttp, ServerGrpc: serverGrpc}, nil
}
