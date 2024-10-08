package cache

import (
	"context"
	"time"

	rcache "github.com/frain-dev/newcloud-migrator/convoy-23.9.2/cache/redis"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/config"
)

type Cache interface {
	Set(ctx context.Context, key string, data interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, data interface{}) error
	Delete(ctx context.Context, key string) error
}

func NewCache(cfg config.RedisConfiguration) (Cache, error) {
	ca, err := rcache.NewRedisCache(cfg.BuildDsn())
	if err != nil {
		return nil, err
	}

	return ca, nil
}
