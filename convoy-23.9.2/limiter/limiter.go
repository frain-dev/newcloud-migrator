package limiter

import (
	"context"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/config"
	rlimiter "github.com/frain-dev/newcloud-migrator/convoy-23.9.2/limiter/redis"
	"github.com/go-redis/redis_rate/v10"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit, duration int) (*redis_rate.Result, error)
	ShouldAllow(ctx context.Context, key string, limit, duration int) (*redis_rate.Result, error)
}

func NewLimiter(cfg config.RedisConfiguration) (RateLimiter, error) {
	ra, err := rlimiter.NewRedisLimiter(cfg.BuildDsn())
	if err != nil {
		return nil, err
	}

	return ra, nil
}
