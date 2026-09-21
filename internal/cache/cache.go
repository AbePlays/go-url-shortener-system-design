package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type UrlCache struct {
	client *redis.Client
}

func New(client *redis.Client) *UrlCache {
	return &UrlCache{client: client}
}

func (c *UrlCache) Get(ctx context.Context, code string) (string, error) {
	val, err := c.client.Get(ctx, code).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrCacheMiss
		}

		return "", err
	}

	return val, nil
}

func (c *UrlCache) Set(ctx context.Context, code string, url string, ttl time.Duration) error {
	err := c.client.Set(ctx, code, url, ttl).Err()
	return err
}
