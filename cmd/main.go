package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"log/slog"
	"net/http"
	"os"
	"urlshortn/cmd/instrumentation"
	"urlshortn/pkg/api"
	"urlshortn/pkg/hash"
	"urlshortn/pkg/storage"
	"urlshortn/pkg/token"
)

const (
	appName      = "shortn"
	defaultEpoch = "2010-11-04T00:00:00Z" //this seems to be twitter's default epoch. Using the same
)

func main() {
	os.Exit(runApp(appName, os.Args[1:]...))
}

func runApp(name string, args ...string) int {
	port := getEnvVarOrDefault("PORT", "8080")
	fmt.Println("Starting http server on port " + port)

	redisAddr := getEnvVarOrDefault("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnvVarOrDefault("REDIS_PASSWORD", "")

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	metrics := instrumentation.NewMetrics()
	metricsHooks := metrics.GetHooks()

	tokenGen := token.NewSnowflakeTokenGenerator(defaultEpoch, logger)

	urlTokenHasher := hash.NewUrlTokenHash(logger)

	urlStore := storage.NewRedisStore(redisAddr, redisPassword, logger)

	urlHandler := api.NewUrlHandler(tokenGen, urlTokenHasher, urlStore, metricsHooks, logger)

	http.HandleFunc("/shortn", func(w http.ResponseWriter, r *http.Request) {
		urlHandler.ShortenUrl(w, r)
	})
	http.HandleFunc("/shortn/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			urlHandler.GetLongUrl(w, r)
		case http.MethodDelete:
			urlHandler.DeleteShortenUrl(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
		return 1
	}

	return 0
}

func getEnvVarOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
