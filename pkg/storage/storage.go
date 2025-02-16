package storage

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

const (
	defaultTTL = time.Hour * 24 * 31 //assuming max number of days in a month
)

type Store interface {
	Fetch(string) (string, error)
	Store(string, string) error
	Remove(string) error
}

type RedisStore struct {
	client interface {
		Get(ctx context.Context, key string) *redis.StringCmd
		Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
		Del(ctx context.Context, keys ...string) *redis.IntCmd
	}
	logger *slog.Logger
}

func NewRedisStore(redisClientAddr string, redisClientPassword string, logger *slog.Logger) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     redisClientAddr,
		Password: redisClientPassword,
		DB:       0,
	})
	return &RedisStore{client: client, logger: logger}
}

func (store *RedisStore) Fetch(key string) (string, error) {
	return store.client.Get(context.Background(), key).Result()
}

func (store *RedisStore) Store(key string, data string) error {
	return store.client.Set(context.Background(), key, data, defaultTTL).Err()
}

func (store *RedisStore) Remove(key string) error {
	return store.client.Del(context.Background(), key).Err()
}

type FakeUrlStore struct {
	FetchFn  func(string) (string, error)
	StoreFn  func(string, string) error
	RemoveFn func(string) error
}

func (store *FakeUrlStore) Fetch(key string) (string, error) {
	return store.FetchFn(key)
}
func (store *FakeUrlStore) Store(key string, data string) error {
	return store.StoreFn(key, data)
}
func (store *FakeUrlStore) Remove(key string) error {
	return store.RemoveFn(key)
}
