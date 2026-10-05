// Package app wires configuration, infrastructure, and feature modules into
// an http.Handler. main.go and tests both build the server through here.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cameronsralla/culdechat/server/internal/auth"
	"github.com/cameronsralla/culdechat/server/internal/config"
	"github.com/cameronsralla/culdechat/server/internal/db"
	"github.com/cameronsralla/culdechat/server/internal/features/authn"
	"github.com/cameronsralla/culdechat/server/internal/features/users"
	"github.com/cameronsralla/culdechat/server/internal/health"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
	"github.com/cameronsralla/culdechat/server/internal/mail"
	"github.com/cameronsralla/culdechat/server/internal/middleware"
	"github.com/cameronsralla/culdechat/server/internal/settings"
)

// Module is a feature that mounts routes. Every feature package exports one.
type Module interface {
	Routes(r chi.Router)
}

// App holds live infrastructure for the lifetime of the process.
type App struct {
	Cfg     config.Config
	Log     *slog.Logger
	Pool    *pgxpool.Pool
	Handler http.Handler
	Version string
}

// New connects the database, runs migrations, bootstraps the admin, and
// builds the router.
func New(ctx context.Context, cfg config.Config, log *slog.Logger, version string) (*App, error) {
	pool, err := db.Open(ctx, cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	tokens := auth.NewTokens(cfg.JWTSecret, cfg.AccessTokenTTL)
	mailer := mail.New(cfg.SMTP, log)
	settingsSvc := settings.NewService(pool)
	usersSvc := users.NewService(pool, mailer, settingsSvc, cfg.PublicURL)
	authSvc := authn.NewService(pool, tokens, cfg.RefreshTokenTTL)

	ba := cfg.BootstrapAdmin
	if err := usersSvc.EnsureBootstrapAdmin(ctx, ba.Email, ba.Password, ba.Name, ba.Unit); err != nil {
		pool.Close()
		return nil, fmt.Errorf("bootstrap admin: %w", err)
	}

	general := middleware.NewLimiter(300, 60, cfg.RateLimitEnabled) // per user/IP
	strict := middleware.NewLimiter(10, 5, cfg.RateLimitEnabled)    // credential endpoints per IP

	modules := []Module{
		&authn.Module{Svc: authSvc, Strict: strict},
		&users.Module{Svc: usersSvc, DevMode: !cfg.IsProd()},
		&settings.Module{Svc: settingsSvc},
	}

	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP(cfg.TrustedProxies),
		middleware.Logger,
		middleware.Recover,
		middleware.Timeout(cfg.RequestTimeout),
		middleware.BodyLimit(cfg.MaxJSONBody),
		middleware.SecurityHeaders(cfg.IsProd()),
		middleware.CORS(cfg.CORSOrigins),
	)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) { httpx.Fail(w, r, httpx.ErrNotFound) })
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpx.Fail(w, r, httpx.ErrBadRequest.WithMessage("method not allowed"))
	})

	(&health.Module{Pool: pool, Version: version}).Routes(r)

	r.Route("/api", func(api chi.Router) {
		api.Use(tokens.Authenticate, general.ByUser)
		for _, m := range modules {
			m.Routes(api)
		}
	})

	return &App{Cfg: cfg, Log: log, Pool: pool, Handler: r, Version: version}, nil
}

func (a *App) Close() { a.Pool.Close() }
