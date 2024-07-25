package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"sso/internal/config"
	"time"
)

type Cache struct {
	client *redis.Client
}

func New(cfg *config.Config) *Cache {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
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

//
//func TestNew() *Cache {
//	redisClient := redis.NewClient(&redis.Options{
//		Addr:     "localhost:6379",
//		Password: "", // no password set
//		DB:       0,  // use default DB
//	})
//	return &Cache{client: redisClient}
//}

func (s *Cache) SaveToken(
	ctx context.Context,
	token string,
	ttl time.Duration,
) error {
	const op = "DATA LAYER: storage.redis.SaveToken"

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

	val, err := s.client.Exists(ctx, token).Result()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return val, nil
}

//func main() {
//	storage := TestNew()
//	ctx := context.Background()
//
//	err := storage.SaveToken(
//		ctx,
//		"e7bw24hj8bhz8-4w",
//		0,
//	)
//	if err != nil {
//		fmt.Println("Error SAVE:", err.Error())
//	}
//
//	result, err := storage.GetToken(ctx, "e7bw24hj8bhz1-4w")
//	if err != nil {
//		fmt.Println("Error GET:", err.Error())
//	}
//	fmt.Println("result: ", result)
//
//	exists, err := storage.CheckTokenExists(ctx, "e7bw24hj8bhz1-4w")
//	if err != nil {
//		fmt.Println("Error CheckTokenExists:", err.Error())
//	}
//	fmt.Println("exists: ", exists)
//
//	result, err = storage.GetToken(ctx, "e7bw24hj8bhz8-4w")
//	if err != nil {
//		fmt.Println("Error GET:", err.Error())
//	}
//	fmt.Println("result: ", result)
//
//	exists, err = storage.CheckTokenExists(ctx, "e7bw24hj8bhz8-4w")
//	if err != nil {
//		fmt.Println("Error CheckTokenExists:", err.Error())
//	}
//	fmt.Println("exists: ", exists)
//}
