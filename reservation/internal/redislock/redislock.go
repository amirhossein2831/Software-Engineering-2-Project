package redislock

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Locker struct {
	rdb *redis.Client
	ttl time.Duration
}

func New(addr string, ttl time.Duration) *Locker {
	return &Locker{
		rdb: redis.NewClient(&redis.Options{Addr: addr}),
		ttl: ttl,
	}
}

func (l *Locker) TTL() time.Duration { return l.ttl }

func SeatKey(eventID, seatID uuid.UUID) string {
	return fmt.Sprintf("lock:seat:%s:%s", eventID, seatID)
}

var acquireScript = redis.NewScript(`
for i=1,#KEYS do
  if redis.call('GET', KEYS[i]) then return i end
end
for i=1,#KEYS do
  redis.call('SET', KEYS[i], ARGV[1], 'PX', ARGV[2])
end
return 0
`)

func (l *Locker) Acquire(ctx context.Context, keys []string, holdID string) (bool, error) {
	res, err := acquireScript.Run(ctx, l.rdb, keys, holdID, l.ttl.Milliseconds()).Int()
	if err != nil {
		return false, err
	}
	return res == 0, nil
}

func (l *Locker) Release(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	return l.rdb.Del(ctx, keys...).Err()
}
