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
	client *redis.Client
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
