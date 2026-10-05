package users

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/cameronsralla/culdechat/server/internal/auth"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
)

type Module struct {
	Svc     *Service
	DevMode bool
}

// Routes mounts under /api. Caller wraps with Authenticate; guards are applied here.
func (m *Module) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireUser)
		r.Get("/me", m.me)
		r.Patch("/me", m.updateMe)
		r.Get("/directory", m.directory)
	})
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAdmin)
		r.Get("/admin/users", m.list)
		r.Post("/admin/users/invite", m.invite)
		r.Post("/admin/users/{id}/reinvite", m.reinvite)
		r.Put("/admin/users/{id}/admin", m.setAdmin)
		r.Put("/admin/users/{id}/status", m.setStatus)
	})
}

func (m *Module) me(w http.ResponseWriter, r *http.Request) {
	id := auth.MustIdentity(r.Context())
	u, err := m.Svc.Get(r.Context(), id.UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, ToUser(u))
}

func (m *Module) updateMe(w http.ResponseWriter, r *http.Request) {
	var in UpdateProfileInput
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	u, err := m.Svc.UpdateProfile(r.Context(), auth.MustIdentity(r.Context()).UserID, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, ToUser(u))
}

func (m *Module) directory(w http.ResponseWriter, r *http.Request) {
	list, err := m.Svc.Directory(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

func (m *Module) list(w http.ResponseWriter, r *http.Request) {
	rows, err := m.Svc.List(r.Context())
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	out := make([]User, len(rows))
	for i, u := range rows {
		out[i] = ToUser(u)
	}
	httpx.JSON(w, http.StatusOK, out)
}

func (m *Module) invite(w http.ResponseWriter, r *http.Request) {
	var in InviteInput
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	res, err := m.Svc.Invite(r.Context(), auth.MustIdentity(r.Context()).UserID, in, m.DevMode)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, res)
}

func (m *Module) reinvite(w http.ResponseWriter, r *http.Request) {
	target, ok := pathUUID(w, r)
	if !ok {
		return
	}
	res, err := m.Svc.Reinvite(r.Context(), auth.MustIdentity(r.Context()).UserID, target, m.DevMode)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, res)
}

func (m *Module) setAdmin(w http.ResponseWriter, r *http.Request) {
	target, ok := pathUUID(w, r)
	if !ok {
		return
	}
	var in struct {
		IsAdmin bool `json:"is_admin"`
	}
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	u, err := m.Svc.SetAdmin(r.Context(), auth.MustIdentity(r.Context()).UserID, target, in.IsAdmin)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, ToUser(u))
}

func (m *Module) setStatus(w http.ResponseWriter, r *http.Request) {
	target, ok := pathUUID(w, r)
	if !ok {
		return
	}
	var in struct {
		Active bool `json:"active"`
	}
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	u, err := m.Svc.SetStatus(r.Context(), auth.MustIdentity(r.Context()).UserID, target, in.Active)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, ToUser(u))
}

func pathUUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, httpx.ErrBadRequest.WithMessage("invalid id"))
		return uuid.Nil, false
	}
	return id, true
}
