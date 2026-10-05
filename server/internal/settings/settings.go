// Package settings is the operator-editable runtime configuration stored in
// the settings table (JSONB values). Reads are cached briefly in-process.
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cameronsralla/culdechat/server/internal/auth"
	"github.com/cameronsralla/culdechat/server/internal/db/dbq"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
)

// Well-known keys. Adding a setting = add a key here, a default in the
// migration, and (if needed) a validator in validate().
const (
	CommunityName     = "community_name"
	CapabilityTier    = "capability_tier"
	InviteExpiryHours = "invite_expiry_hours"
)

// publicKeys are readable by any authenticated resident; everything else is admin-only.
var publicKeys = map[string]bool{CommunityName: true, CapabilityTier: true}

type Service struct {
	q     *dbq.Queries
	mu    sync.RWMutex
	cache map[string]json.RawMessage
	at    time.Time
	ttl   time.Duration
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{q: dbq.New(pool), cache: map[string]json.RawMessage{}, ttl: 30 * time.Second}
}

func (s *Service) all(ctx context.Context) (map[string]json.RawMessage, error) {
	s.mu.RLock()
	if time.Since(s.at) < s.ttl && len(s.cache) > 0 {
		c := s.cache
		s.mu.RUnlock()
		return c, nil
	}
	s.mu.RUnlock()

	rows, err := s.q.ListSettings(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]json.RawMessage, len(rows))
	for _, r := range rows {
		m[r.Key] = json.RawMessage(r.Value)
	}
	s.mu.Lock()
	s.cache, s.at = m, time.Now()
	s.mu.Unlock()
	return m, nil
}

func (s *Service) invalidate() {
	s.mu.Lock()
	s.at = time.Time{}
	s.mu.Unlock()
}

// Get unmarshals a setting into out; returns false if missing.
func (s *Service) Get(ctx context.Context, key string, out any) bool {
	m, err := s.all(ctx)
	if err != nil {
		return false
	}
	raw, ok := m[key]
	if !ok {
		return false
	}
	return json.Unmarshal(raw, out) == nil
}

func (s *Service) String(ctx context.Context, key, def string) string {
	var v string
	if s.Get(ctx, key, &v) {
		return v
	}
	return def
}

func (s *Service) Int(ctx context.Context, key string, def int) int {
	var v int
	if s.Get(ctx, key, &v) {
		return v
	}
	return def
}

// Set validates and writes a setting.
func (s *Service) Set(ctx context.Context, key string, value json.RawMessage, by uuid.UUID) (dbq.Setting, error) {
	if err := validate(key, value); err != nil {
		return dbq.Setting{}, httpx.ErrBadRequest.WithMessage(err.Error())
	}
	row, err := s.q.UpsertSetting(ctx, dbq.UpsertSettingParams{Key: key, Value: value, UpdatedBy: &by})
	if err != nil {
		return dbq.Setting{}, err
	}
	s.invalidate()
	return row, nil
}

func validate(key string, raw json.RawMessage) error {
	switch key {
	case CommunityName:
		var v string
		if json.Unmarshal(raw, &v) != nil || v == "" || len(v) > 80 {
			return errors.New("community_name must be a non-empty string up to 80 characters")
		}
	case CapabilityTier:
		var v string
		if json.Unmarshal(raw, &v) != nil || (v != "CORE" && v != "STANDARD" && v != "PLUS") {
			return errors.New("capability_tier must be CORE, STANDARD, or PLUS")
		}
	case InviteExpiryHours:
		var v int
		if json.Unmarshal(raw, &v) != nil || v < 1 || v > 24*90 {
			return errors.New("invite_expiry_hours must be between 1 and 2160")
		}
	default:
		return fmt.Errorf("unknown setting %q", key)
	}
	return nil
}

// ---- HTTP ----

type Module struct{ Svc *Service }

type setting struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (m *Module) Routes(r chi.Router) {
	r.With(auth.RequireUser).Get("/settings", m.public)
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAdmin)
		r.Get("/admin/settings", m.list)
		r.Put("/admin/settings/{key}", m.set)
	})
}

func (m *Module) public(w http.ResponseWriter, r *http.Request) {
	all, err := m.Svc.all(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	out := map[string]json.RawMessage{}
	for k := range publicKeys {
		if v, ok := all[k]; ok {
			out[k] = v
		}
	}
	httpx.JSON(w, http.StatusOK, out)
}

func (m *Module) list(w http.ResponseWriter, r *http.Request) {
	rows, err := m.Svc.q.ListSettings(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	out := make([]setting, len(rows))
	for i, s := range rows {
		out[i] = setting{Key: s.Key, Value: json.RawMessage(s.Value), UpdatedAt: s.UpdatedAt}
	}
	httpx.JSON(w, http.StatusOK, out)
}

func (m *Module) set(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Value json.RawMessage `json:"value"`
	}
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if len(in.Value) == 0 {
		httpx.Fail(w, r, httpx.ErrBadRequest.WithMessage("value is required"))
		return
	}
	row, err := m.Svc.Set(r.Context(), chi.URLParam(r, "key"), in.Value, auth.MustIdentity(r.Context()).UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = httpx.ErrNotFound
	}
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, setting{Key: row.Key, Value: json.RawMessage(row.Value), UpdatedAt: row.UpdatedAt})
}
