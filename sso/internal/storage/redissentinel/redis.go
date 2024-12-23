package redissentinel

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexBlackNn/authloyalty/sso/internal/config"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Cache struct {
	client *redis.ClusterClient
}

func New(cfg *config.Config) (*Cache, error) {
	// NewFailoverClusterClient routes readonly commands to slave nodes.
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
		return nil, err
	}

	// Enable metrics instrumentation.
	if err := redisotel.InstrumentMetrics(redisClient); err != nil {
		return nil, err
	}
	return &Cache{client: redisClient}, nil
}

var tracer = otel.Tracer("sso service")

func (s *Cache) SaveToken(
	ctx context.Context,
	token string,
	ttl time.Duration,
) error {
	const op = "DATA LAYER: storage.redis.SaveToken"

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.String("handler", "SaveToken")))
	defer span.End()

	err := s.client.Set(ctx, token, true, ttl).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Cache) GetToken(
	ctx context.Context,
	token string,
) (string, error) {
	const op = "DATA LAYER: storage.redis.GetToken"

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.String("handler", "GetToken")))
	defer span.End()

	val, err := s.client.Get(ctx, token).Result()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return val, nil
}

func (s *Cache) CheckTokenExists(
	ctx context.Context,
	token string,
) (int64, error) {
	const op = "DATA LAYER: storage.redis.CheckTokenExists"

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.String("handler", "CheckTokenExists")))
	defer span.End()

	val, err := s.client.Exists(ctx, token).Result()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return val, nil
}
