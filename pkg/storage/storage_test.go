package storage

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestRedisStore_Fetch(t *testing.T) {
	type fields struct {
		client interface {
			Get(ctx context.Context, key string) *redis.StringCmd
			Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
			Del(ctx context.Context, keys ...string) *redis.IntCmd
		}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "when fetching if there is an error, return it",
			fields: fields{
				client: &FakeRedisStore{
					GetFn: func(ctx context.Context, key string) *redis.StringCmd {
						result := &redis.StringCmd{}
						result.SetVal("")
						result.SetErr(errors.New("expected error"))
						return result
					},
				},
			},
			args: args{
				key: "something",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "when fetching if there is no error, return val",
			fields: fields{
				client: &FakeRedisStore{
					GetFn: func(ctx context.Context, key string) *redis.StringCmd {
						result := &redis.StringCmd{}
						result.SetVal("value")
						result.SetErr(nil)
						return result
					},
				},
			},
			args: args{
				key: "something",
			},
			want:    "value",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			store := &RedisStore{
				client: tt.fields.client,
				logger: logger,
			}
			got, err := store.Fetch(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Fetch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisStore_Store(t *testing.T) {
	type fields struct {
		client interface {
			Get(ctx context.Context, key string) *redis.StringCmd
			Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
			Del(ctx context.Context, keys ...string) *redis.IntCmd
		}
	}
	type args struct {
		key  string
		data string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "when storing, if there's an error, return it",
			fields: fields{
				client: &FakeRedisStore{
					SetFn: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
						result := &redis.StatusCmd{}
						result.SetErr(errors.New("expected error"))
						return result
					},
				},
			},
			args: args{
				key:  "key",
				data: "value",
			},
			wantErr: true,
		},
		{
			name: "when storing, if there's no error, return nil",
			fields: fields{
				client: &FakeRedisStore{
					SetFn: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
						result := &redis.StatusCmd{}
						result.SetErr(nil)
						return result
					},
				},
			},
			args: args{
				key:  "key",
				data: "value",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			store := &RedisStore{
				client: tt.fields.client,
				logger: logger,
			}
			if err := store.Store(tt.args.key, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisStore_Remove(t *testing.T) {
	type fields struct {
		client interface {
			Get(ctx context.Context, key string) *redis.StringCmd
			Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
			Del(ctx context.Context, keys ...string) *redis.IntCmd
		}
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "when there is an error removing a key, return it",
			fields: fields{
				client: &FakeRedisStore{
					DelFn: func(ctx context.Context, keys ...string) *redis.IntCmd {
						result := &redis.IntCmd{}
						result.SetErr(errors.New("expected error"))
						return result
					},
				},
			},
			args: args{
				key: "something",
			},
			wantErr: true,
		},
		{
			name: "when there is an error removing a key, return it",
			fields: fields{
				client: &FakeRedisStore{
					DelFn: func(ctx context.Context, keys ...string) *redis.IntCmd {
						result := &redis.IntCmd{}
						result.SetErr(nil)
						return result
					},
				},
			},
			args: args{
				key: "something",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			store := &RedisStore{
				client: tt.fields.client,
				logger: logger,
			}
			if err := store.Remove(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type FakeRedisStore struct {
	GetFn func(ctx context.Context, key string) *redis.StringCmd
	SetFn func(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	DelFn func(ctx context.Context, keys ...string) *redis.IntCmd
}

func (f *FakeRedisStore) Get(ctx context.Context, key string) *redis.StringCmd {
	return f.GetFn(ctx, key)
}
func (f *FakeRedisStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return f.SetFn(ctx, key, value, expiration)
}
func (f *FakeRedisStore) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return f.DelFn(ctx, keys...)
}
