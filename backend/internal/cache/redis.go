package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/kntjspr/fullstack-golang-next-template/internal/config"
)

// OpenRedis creates and validates a Redis connection.
func OpenRedis(cfg *config.RedisConfig, log zerolog.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Info().
		Str("addr", cfg.Addr()).
		Int("db", cfg.DB).
		Msg("redis connection established")

	return client, nil
}
