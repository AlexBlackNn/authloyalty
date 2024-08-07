package suite

import (
	"context"
	"fmt"
	"github.com/AlexBlackNn/authloyalty/internal/config"
	ssov1 "github.com/AlexBlackNn/authloyalty/protos/proto/sso/gen"
	"github.com/AlexBlackNn/authloyalty/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
	"strconv"
	"testing"
	"time"
)

const (
	grpcHost = "localhost"
)

type Suite struct {
	*testing.T                  // потребуется для вызова методов *testing.T внутри Suite
	Cfg        *config.Config   // Конфигурация приложения
	AuthClient ssov1.AuthClient // Клиент для взаимодействия с grpc_transport - сервером
}

var tracer = otel.Tracer("testing client")

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()   // in case of failure in one of tests to form rightly stack trace
	t.Parallel() //can run test in parallel to increase performance

	cfg := config.MustLoadByPath("../config/local.yaml")
	tp, err := tracing.Init("testing client", cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)
	ctx = metadata.NewOutgoingContext(context.Background(), md)

	_, span := tracer.Start(ctx, "Testing Client")
	defer span.End()
	traceId := fmt.Sprintf("%s", span.SpanContext().TraceID())
	ctx = metadata.AppendToOutgoingContext(ctx, "x-trace-id", traceId)

	// context will be canceled when tests are stopped
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	// create grpc_transport client
	cc, err := grpc.DialContext(
		context.Background(),
		grpcAddress(cfg),
		//use insecure connection during test
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		t.Fatalf("grpc_transport server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:          t,
		Cfg:        cfg,
		AuthClient: ssov1.NewAuthClient(cc),
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
