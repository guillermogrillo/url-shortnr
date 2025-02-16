package api

import (
	"bytes"
	"errors"
	"github.com/bwmarrin/snowflake"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"urlshortn/pkg/hash"
	"urlshortn/pkg/metrics"
	"urlshortn/pkg/storage"
	"urlshortn/pkg/token"
)

func TestUrlHandler_ShortenUrl(t *testing.T) {
	type fields struct {
		TokenGen              token.TokenGenerator
		TokenHasher           hash.TokenHasher
		UrlStore              storage.Store
		ShortUrlEventProducer interface {
			Produce(content string) error
		}
		MetricsHooks *metrics.MetricsHooks
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantCode int
	}{
		{
			name:   "when the request body is not parseable, the response is bad request",
			fields: fields{},
			args: args{
				r: httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(""))),
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "when there is an error generating a token, the response is internal server error",
			fields: fields{
				TokenGen: &token.FakeTokenGenerator{GenerateTokenFn: func() (snowflake.ID, error) {
					return 0, errors.New("expected error")
				}},
			},
			args: args{
				r: httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{\"url\":\"http://google.com\"}"))),
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "when there is an error hashing a token, response is internal server error",
			fields: fields{
				TokenGen: &token.FakeTokenGenerator{GenerateTokenFn: func() (snowflake.ID, error) {
					return 1234, nil
				}},
				TokenHasher: &hash.FakeTokenHasher{HashFn: func(n int64) (string, error) {
					return "", errors.New("expected error")
				}},
			},
			args: args{
				r: httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{\"url\":\"http://google.com\"}"))),
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "when the short url is generated, return a status ok",
			fields: fields{
				TokenGen: &token.FakeTokenGenerator{GenerateTokenFn: func() (snowflake.ID, error) {
					return 1234, nil
				}},
				TokenHasher: &hash.FakeTokenHasher{HashFn: func(n int64) (string, error) {
					return "1234", nil
				}},
				ShortUrlEventProducer: &FakeShortUrlEventProducer{
					ProduceFn: func(content string) error {
						return nil
					},
				},
			},
			args: args{
				r: httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{\"url\":\"http://google.com\"}"))),
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			h := &UrlHandler{
				TokenGen:              tt.fields.TokenGen,
				TokenHasher:           tt.fields.TokenHasher,
				UrlStore:              tt.fields.UrlStore,
				ShortUrlEventProducer: tt.fields.ShortUrlEventProducer,
				MetricsHooks:          tt.fields.MetricsHooks,
				logger:                logger,
			}
			rr := httptest.NewRecorder()
			h.ShortenUrl(rr, tt.args.r)
			assert.Equal(t, tt.wantCode, rr.Code, "http status code does not match")
		})
	}
}

func TestUrlHandler_GetLongUrl(t *testing.T) {
	type fields struct {
		TokenGen              token.TokenGenerator
		TokenHasher           hash.TokenHasher
		UrlStore              storage.Store
		ShortUrlEventProducer interface {
			Produce(content string) error
		}
		MetricsHooks *metrics.MetricsHooks
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantCode int
	}{
		{
			name:   "when the url is not correct, response is bad request",
			fields: fields{},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "/shortn/", nil),
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "when there is an error fetching the long url, response is internal server error",
			fields: fields{
				UrlStore: &storage.FakeUrlStore{
					FetchFn: func(s string) (string, error) {
						return "", errors.New("expected error")
					},
				},
			},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "/shortn/1234", nil),
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "when there is an error fetching the long url because the short url does not exist, response is bad request",
			fields: fields{
				UrlStore: &storage.FakeUrlStore{
					FetchFn: func(s string) (string, error) {
						return "", redis.Nil
					},
				},
			},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "/shortn/1234", nil),
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "when the long url is found, response is moved temporarily",
			fields: fields{
				UrlStore: &storage.FakeUrlStore{
					FetchFn: func(s string) (string, error) {
						return "1234567890", nil
					},
				},
			},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "/shortn/1234", nil),
			},
			wantCode: http.StatusFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			h := &UrlHandler{
				TokenGen:              tt.fields.TokenGen,
				TokenHasher:           tt.fields.TokenHasher,
				UrlStore:              tt.fields.UrlStore,
				ShortUrlEventProducer: tt.fields.ShortUrlEventProducer,
				MetricsHooks:          tt.fields.MetricsHooks,
				logger:                logger,
			}
			rr := httptest.NewRecorder()
			h.GetLongUrl(rr, tt.args.r)
			assert.Equal(t, tt.wantCode, rr.Code, "http status code does not match")
		})
	}
}

func TestUrlHandler_DeleteShortenUrl(t *testing.T) {
	type fields struct {
		TokenGen              token.TokenGenerator
		TokenHasher           hash.TokenHasher
		UrlStore              storage.Store
		ShortUrlEventProducer interface {
			Produce(content string) error
		}
		MetricsHooks *metrics.MetricsHooks
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantCode int
	}{
		{
			name:   "when the url is not correct, response is bad request",
			fields: fields{},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "/shortn/", nil),
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "when there is an error deleting the long url, response is internal server error",
			fields: fields{
				UrlStore: &storage.FakeUrlStore{
					RemoveFn: func(s string) error {
						return errors.New("expected error")
					},
				},
			},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "/shortn/1234", nil),
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "when there is an error deleting the long url because the short url does not exist, response is bad request",
			fields: fields{
				UrlStore: &storage.FakeUrlStore{
					RemoveFn: func(s string) error {
						return redis.Nil
					},
				},
			},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "/shortn/1234", nil),
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "when the short url is deleted, response is OK",
			fields: fields{
				UrlStore: &storage.FakeUrlStore{
					RemoveFn: func(s string) error {
						return nil
					},
				},
			},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "/shortn/1234", nil),
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			h := &UrlHandler{
				TokenGen:              tt.fields.TokenGen,
				TokenHasher:           tt.fields.TokenHasher,
				UrlStore:              tt.fields.UrlStore,
				ShortUrlEventProducer: tt.fields.ShortUrlEventProducer,
				MetricsHooks:          tt.fields.MetricsHooks,
				logger:                logger,
			}
			rr := httptest.NewRecorder()
			h.DeleteShortenUrl(rr, tt.args.r)
			assert.Equal(t, tt.wantCode, rr.Code, "http status code does not match")
		})
	}
}

type FakeShortUrlEventProducer struct {
	ProduceFn func(content string) error
}

func (f *FakeShortUrlEventProducer) Produce(content string) error {
	return f.ProduceFn(content)
}
