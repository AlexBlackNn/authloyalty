package redis_sentinel

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"sso/internal/config"
	"time"
)

type Cache struct {
	client *redis.ClusterClient
}

func New(cfg *config.Config) *Cache {
	// NewFailoverClusterClient routes readonly commands to slave nodes
	redisClient := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName: cfg.RedisSentinel.MasterName,
		SentinelAddrs: []string{
			cfg.RedisSentinel.SentinelAddrs1,
			cfg.RedisSentinel.SentinelAddrs2,
			cfg.RedisSentinel.SentinelAddrs3,
		},
		Password: cfg.RedisSentinel.Password,
	})

	// Enable tracing instrumentation.
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		panic(err)
	}

	// Enable metrics instrumentation.
	if err := redisotel.InstrumentMetrics(redisClient); err != nil {
		panic(err)
	}
	return &Cache{client: redisClient}
}

var tracer = otel.Tracer("sso service")

func (s *Cache) SaveToken(
	ctx context.Context,
	token string,
	ttl time.Duration,
) (context.Context, error) {
	const op = "DATA LAYER: storage.redis.SaveToken"

	ctx, span := tracer.Start(ctx, "data layer RedisSentinel: SaveToken",
		trace.WithAttributes(attribute.String("handler", "SaveToken")))
	defer span.End()

	err := s.client.Set(ctx, token, true, ttl).Err()
	if err != nil {
		return ctx, fmt.Errorf("%s: %w", op, err)
	}

	return ctx, nil
}

func (s *Cache) GetToken(
	ctx context.Context,
	token string,
) (context.Context, string, error) {
	const op = "DATA LAYER: storage.redis.GetToken"

	ctx, span := tracer.Start(ctx, "data layer RedisSentinel: GetToken",
		trace.WithAttributes(attribute.String("handler", "GetToken")))
	defer span.End()

	val, err := s.client.Get(ctx, token).Result()
	if err != nil {
		return ctx, "", fmt.Errorf("%s: %w", op, err)
	}
	return ctx, val, nil
}

func (s *Cache) CheckTokenExists(
	ctx context.Context,
	token string,
) (context.Context, int64, error) {
	const op = "DATA LAYER: storage.redis.CheckTokenExists"

	ctx, span := tracer.Start(ctx, "data layer RedisSentinel: CheckTokenExists",
		trace.WithAttributes(attribute.String("handler", "CheckTokenExists")))
	defer span.End()

	val, err := s.client.Exists(ctx, token).Result()
	if err != nil {
		return ctx, 0, fmt.Errorf("%s: %w", op, err)
	}
	return ctx, val, nil
}
