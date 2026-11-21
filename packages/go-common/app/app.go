package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/auth"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/config"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/db"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/httpx"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/logging"
	"github.com/fardanwardhana/qoder-hackathon/packages/go-common/server"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Cfg     config.Config
	Log     *slog.Logger
	Pool    *pgxpool.Pool
	Metrics *httpx.Metrics
	Router  *chi.Mux
}

func New(ctx context.Context, service string) (*App, error) {
	cfg := config.Load(service)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	logger := logging.New(service, cfg.LogLevel)
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	metrics := httpx.NewMetrics()
	r := chi.NewRouter()
	r.Use(httpx.RequestID)
	r.Use(httpx.Recoverer(logger))
	r.Use(httpx.AccessLog(logger))
	r.Use(metrics.Middleware)
	r.Get("/health", httpx.Health)
	r.Get("/ready", httpx.Ready(func(ctx context.Context) error { return db.Ping(ctx, pool) }))
	r.Method(http.MethodGet, "/metrics", metrics.Handler())

	return &App{Cfg: cfg, Log: logger, Pool: pool, Metrics: metrics, Router: r}, nil
}

func (a *App) Close() {
	a.Pool.Close()
}

func (a *App) Run(ctx context.Context) error {
	return server.Run(ctx, a.Cfg.HTTPAddr, a.Router, a.Log)
}

func (a *App) Verifier() *auth.HTTPVerifier {
	return auth.NewHTTPVerifier(a.Cfg.IdentityURL, a.Cfg.InternalToken)
}
