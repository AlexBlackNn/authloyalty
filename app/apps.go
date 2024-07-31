package app

import (
	"github.com/AlexBlackNn/authloyalty/app/servergrpc"
	"github.com/AlexBlackNn/authloyalty/app/serverhttp"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	"github.com/AlexBlackNn/authloyalty/internal/logger"
)

type App struct {
	ServerHttp *serverhttp.App
	ServerGrpc *servergrpc.App
}

func New() (*App, error) {
	cfg := config.New()
	log := logger.New(cfg.Env)

	// http server
	serverHttp, err := serverhttp.New(cfg, log)
	if err != nil {
		return nil, err
	}

	// grpc server
	serverGrpc, err := servergrpc.New(cfg, log)
	if err != nil {
		return nil, err
	}
	return &App{ServerHttp: serverHttp, ServerGrpc: serverGrpc}, nil
}
