package queue

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	rdb         *redis.Client
	defaultRate float64
}

func New(addr string, defaultRate float64) *Queue {
	return &Queue{
		rdb:         redis.NewClient(&redis.Options{Addr: addr}),
		defaultRate: defaultRate,
	}
}

func zkey(eventID string) string     { return "queue:" + eventID }
func startKey(eventID string) string { return "queue:" + eventID + ":start" }
func rateKey(eventID string) string  { return "queue:" + eventID + ":rate" }

func (q *Queue) Join(ctx context.Context, eventID, userID string) (int64, error) {
	now := time.Now().UnixMilli()
	q.rdb.SetNX(ctx, startKey(eventID), now, 0)
	if err := q.rdb.ZAddNX(ctx, zkey(eventID), redis.Z{Score: float64(now), Member: userID}).Err(); err != nil {
		return 0, err
	}
	_, position, err := q.rank(ctx, eventID, userID)
	return position, err
}

func (q *Queue) Status(ctx context.Context, eventID, userID string) (bool, int64, error) {
	rank, err := q.rdb.ZRank(ctx, zkey(eventID), userID).Result()
	if errors.Is(err, redis.Nil) {
		return true, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	admitted, err := q.admittedCount(ctx, eventID)
	if err != nil {
		return false, 0, err
	}
	if rank < admitted {
		_ = q.rdb.ZRem(ctx, zkey(eventID), userID).Err()
		return true, 0, nil
	}
	return false, rank - admitted + 1, nil
}

func (q *Queue) SetRate(ctx context.Context, eventID string, rate float64) error {
	return q.rdb.Set(ctx, rateKey(eventID), rate, 0).Err()
}

func (q *Queue) rank(ctx context.Context, eventID, userID string) (bool, int64, error) {
	rank, err := q.rdb.ZRank(ctx, zkey(eventID), userID).Result()
	if err != nil {
		return false, 0, err
	}
	admitted, err := q.admittedCount(ctx, eventID)
	if err != nil {
		return false, 0, err
	}
	if rank < admitted {
		return true, 0, nil
	}
	return false, rank - admitted + 1, nil
}

func (q *Queue) admittedCount(ctx context.Context, eventID string) (int64, error) {
	start, err := q.rdb.Get(ctx, startKey(eventID)).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	rate := q.defaultRate
	if v, err := q.rdb.Get(ctx, rateKey(eventID)).Float64(); err == nil {
		rate = v
	}
	elapsed := float64(time.Now().UnixMilli()-start) / 1000
	return int64(elapsed * rate), nil
}
