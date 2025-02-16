package instrumentation

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"urlshortn/pkg/metrics"
)

const (
	shortenUrlStartName          = "shorten_started"
	shortenEndpointName          = "shorten"
	getLongUrlStartName          = "get_long_url_started"
	getLongUrlEndpointName       = "get_long_url"
	deleteShortenUrlStartName    = "delete_shorten_url"
	deleteShortenUrlEndpointName = "delete_shorten"
)

type Metrics struct {
	totalRequests    *prometheus.CounterVec
	totalErrors      *prometheus.CounterVec
	requestsDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	totalRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "url"},
	)
	totalErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_errors",
			Help: "Total number of errors",
		},
		[]string{"method", "endpoint", "url"},
	)
	requestsDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	prometheus.MustRegister(totalRequests)
	prometheus.MustRegister(totalErrors)
	prometheus.MustRegister(requestsDuration)

	return &Metrics{
		totalRequests:    totalRequests,
		totalErrors:      totalErrors,
		requestsDuration: requestsDuration,
	}
}

func (m *Metrics) GetHooks() *metrics.MetricsHooks {
	return &metrics.MetricsHooks{
		OnShortenUrlCalledFn: func(ctx context.Context, longUrl string) context.Context {
			return context.WithValue(ctx, shortenUrlStartName, time.Now())
		},
		OnShortenUrlFinishedFn: func(ctx context.Context, longUrl string, err error) {
			fmt.Println(fmt.Printf("OnShortenUrlFinishedFn with context %v", ctx))
			if startedAt, ok := ctx.Value(shortenUrlStartName).(time.Time); ok {
				fmt.Printf("Roundtrip for %s time is %v\n", shortenEndpointName, time.Since(startedAt).Seconds())
				m.requestsDuration.WithLabelValues("POST", shortenEndpointName).Observe(time.Since(startedAt).Seconds())
			}
			if err != nil {
				m.totalErrors.WithLabelValues("POST", shortenEndpointName, longUrl).Inc()
			}
			m.totalRequests.WithLabelValues("POST", shortenEndpointName, longUrl).Inc()
		},
		OnGetLongUrlCalledFn: func(ctx context.Context, shortenUrl string) context.Context {
			return context.WithValue(ctx, getLongUrlStartName, time.Now())
		},
		OnGetLongUrlFinishedFn: func(ctx context.Context, shortenUrl string, err error) {
			fmt.Println(fmt.Printf("OnGetLongUrlFinishedFn with context %v", ctx))
			if startedAt, ok := ctx.Value(getLongUrlStartName).(time.Time); ok {
				fmt.Printf("Roundtrip for %s time is %v\n", getLongUrlEndpointName, time.Since(startedAt).Seconds())
				m.requestsDuration.WithLabelValues("GET", getLongUrlEndpointName).Observe(time.Since(startedAt).Seconds())
			}
			if err != nil {
				m.totalErrors.WithLabelValues("GET", getLongUrlEndpointName, shortenUrl).Inc()
			}
			m.totalRequests.WithLabelValues("GET", getLongUrlEndpointName, shortenUrl).Inc()
		},
		OnDeleteShortenUrlCalledFn: func(ctx context.Context, shortenUrl string) context.Context {
			return context.WithValue(ctx, deleteShortenUrlStartName, time.Now())
		},
		OnDeleteShortenUrlFinishedFn: func(ctx context.Context, shortenUrl string, err error) {
			fmt.Println(fmt.Printf("OnDeleteShortenUrlFinishedFn with context %v", ctx))
			if startedAt, ok := ctx.Value(deleteShortenUrlStartName).(time.Time); ok {
				fmt.Printf("Roundtrip for %s time is %v\n", deleteShortenUrlEndpointName, time.Since(startedAt).Seconds())
				m.requestsDuration.WithLabelValues("DELETE", deleteShortenUrlEndpointName).Observe(time.Since(startedAt).Seconds())
			}
			if err != nil {
				m.totalErrors.WithLabelValues("DELETE", deleteShortenUrlEndpointName, shortenUrl).Inc()
			}
			m.totalRequests.WithLabelValues("DELETE", deleteShortenUrlEndpointName, shortenUrl).Inc()
		},
	}
}
