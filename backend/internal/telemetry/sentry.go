package telemetry

import (
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"

	"github.com/create-go-app/chi-go-template/internal/config"
)

// InitSentry initializes Sentry SDK and returns middleware handler.
func InitSentry(cfg *config.SentryConfig) (*sentryhttp.Handler, error) {
	if !cfg.Enable {
		return nil, nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		TracesSampleRate: cfg.TracesSampleRate,
	})
	if err != nil {
		return nil, err
	}

	return sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	}), nil
}

// FlushSentry waits to send pending Sentry events.
func FlushSentry() {
	sentry.Flush(2 * time.Second)
}

// Middleware returns net/http compatible middleware from sentry handler.
func Middleware(handler *sentryhttp.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return handler.Handle(next)
	}
}
