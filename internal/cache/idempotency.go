package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTransactionStore struct {
	dbr *redis.Client
}

func NewRedisTransactionStore(dbr *redis.Client) *RedisTransactionStore {
	return &RedisTransactionStore{
		dbr: dbr,
	}
}

func (rts *RedisTransactionStore) IsNew(ctx context.Context, key string, value string) (bool, error) {

	success, err := rts.dbr.SetNX(ctx, key, "PROCESSING", 30*time.Second).Result()

	if err != nil {
		return false, err
	}

	return success, nil
}

func (rts *RedisTransactionStore) Delete(ctx context.Context, key string) error {
	return rts.dbr.Del(ctx, key).Err()
}

func (rts *RedisTransactionStore) Get(ctx context.Context, key string) *redis.StringCmd {
	return rts.dbr.Get(ctx, key)
}

func (rts *RedisTransactionStore) SetStatusCompleted(ctx context.Context, key string) *redis.StatusCmd {
	return rts.dbr.Set(ctx, key, "COMPLETED", 10*time.Second)
}
