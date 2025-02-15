package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	totalRequests    prometheus.CounterVec
	totalErrors      prometheus.CounterOpts
	requestsDuration prometheus.HistogramVec
}

type MetricsHooks struct {
	OnShortenUrlCalledFn         func(ctx context.Context, longUrl string) context.Context
	OnShortenUrlFinishedFn       func(ctx context.Context, longUrl string, err error)
	OnGetLongUrlCalledFn         func(ctx context.Context, shortenUrl string) context.Context
	OnGetLongUrlFinishedFn       func(ctx context.Context, shortenUrl string, err error)
	OnDeleteShortenUrlCalledFn   func(ctx context.Context, shortenUrl string) context.Context
	OnDeleteShortenUrlFinishedFn func(ctx context.Context, shortenUrl string, err error)
}

func (m *MetricsHooks) OnShortenUrlCalled(ctx context.Context, longUrl string) context.Context {
	if m != nil && m.OnShortenUrlCalledFn != nil {
		return m.OnShortenUrlCalledFn(ctx, longUrl)
	}
	return ctx
}

func (m *MetricsHooks) OnShortenUrlFinished(ctx context.Context, longUrl string, err error) {
	if m != nil && m.OnShortenUrlFinishedFn != nil {
		m.OnShortenUrlFinishedFn(ctx, longUrl, err)
	}
}

func (m *MetricsHooks) OnGetLongUrlCalled(ctx context.Context, shortenUrl string) context.Context {
	if m != nil && m.OnGetLongUrlCalledFn != nil {
		return m.OnGetLongUrlCalledFn(ctx, shortenUrl)
	}
	return ctx
}

func (m *MetricsHooks) OnGetLongUrlFinished(ctx context.Context, shortenUrl string, err error) {
	if m != nil && m.OnGetLongUrlFinishedFn != nil {
		m.OnGetLongUrlFinishedFn(ctx, shortenUrl, err)
	}
}

func (m *MetricsHooks) OnDeleteShortenUrlCalled(ctx context.Context, shortenUrl string) context.Context {
	if m != nil && m.OnDeleteShortenUrlCalledFn != nil {
		return m.OnDeleteShortenUrlCalledFn(ctx, shortenUrl)
	}
	return ctx
}

func (m *MetricsHooks) OnDeleteShortenUrlFinished(ctx context.Context, shortenUrl string, err error) {
	if m != nil && m.OnDeleteShortenUrlFinishedFn != nil {
		m.OnDeleteShortenUrlFinishedFn(ctx, shortenUrl, err)
	}
}
