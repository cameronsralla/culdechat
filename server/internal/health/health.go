// Package health exposes liveness and readiness probes.
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cameronsralla/culdechat/server/internal/db"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
)

type Module struct {
	Pool    *pgxpool.Pool
	Version string
}

func (m *Module) Routes(r chi.Router) {
	r.Get("/healthz", m.live)
	r.Get("/readyz", m.ready)
}

func (m *Module) live(w http.ResponseWriter, _ *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok", "version": m.Version})
}

func (m *Module) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := m.Pool.Ping(ctx); err != nil {
		httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db_unreachable"})
		return
	}
	pending, err := db.Pending(ctx, m.Pool)
	if err != nil || pending {
		httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "migrations_pending"})
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready", "version": m.Version})
}
