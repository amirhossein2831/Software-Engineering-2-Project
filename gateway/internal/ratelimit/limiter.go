package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

func New(addr string, limit int, window time.Duration) *Limiter {
	if addr == "" || limit <= 0 {
		return nil
	}
	return &Limiter{
		rdb:    redis.NewClient(&redis.Options{Addr: addr}),
		limit:  limit,
		window: window,
	}
}

func (l *Limiter) Allow(ctx context.Context, key string) bool {
	full := "ratelimit:" + key
	count, err := l.rdb.Incr(ctx, full).Result()
	if err != nil {
		return true
	}
	if count == 1 {
		l.rdb.Expire(ctx, full, l.window)
	}
	return count <= int64(l.limit)
}
