package users

import (
	"net/http"
	"strconv"

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
		r.Get("/admin/units", m.listUnits)
		r.Post("/admin/units", m.createUnit)
		r.Patch("/admin/units/{id}", m.renameUnit)
		r.Delete("/admin/units/{id}", m.deleteUnit)
		r.Get("/admin/users", m.list)
		r.Post("/admin/users/invite", m.invite)
		r.Post("/admin/users/{id}/reinvite", m.reinvite)
		r.Post("/admin/users/{id}/reset-password", m.resetPassword)
		r.Put("/admin/users/{id}/admin", m.setAdmin)
		r.Put("/admin/users/{id}/status", m.setStatus)
	})
}

func (m *Module) listUnits(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := m.Svc.ListUnits(r.Context(), UnitParams{
		Q:        q.Get("q"),
		Dir:      q.Get("dir"),
		Vacant:   queryBool(q.Get("vacant")),
		Page:     queryInt(q.Get("page"), 1),
		PageSize: queryInt(q.Get("page_size"), 25),
	})
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, page)
}

func (m *Module) createUnit(w http.ResponseWriter, r *http.Request) {
	var in UnitInput
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	unit, err := m.Svc.CreateUnit(r.Context(), in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, unit)
}

func (m *Module) renameUnit(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r)
	if !ok {
		return
	}
	var in UnitInput
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	unit, err := m.Svc.RenameUnit(r.Context(), id, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, unit)
}

func (m *Module) deleteUnit(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r)
	if !ok {
		return
	}
	if err := m.Svc.DeleteUnit(r.Context(), id); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.NoContent(w)
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
	q := r.URL.Query()
	page, err := m.Svc.DirectoryPage(r.Context(), DirectoryParams{
		Q:        q.Get("q"),
		Name:     q.Get("name"),
		Email:    q.Get("email"),
		Unit:     q.Get("unit"),
		Sort:     q.Get("sort"),
		Dir:      q.Get("dir"),
		Page:     queryInt(q.Get("page"), 1),
		PageSize: queryInt(q.Get("page_size"), 25),
	}, auth.MustIdentity(r.Context()).UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, page)
}

func (m *Module) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := m.Svc.ListPage(r.Context(), ListParams{
		Q:         q.Get("q"),
		Name:      q.Get("name"),
		Email:     q.Get("email"),
		Unit:      q.Get("unit"),
		Status:    q.Get("status"),
		Admin:     queryBool(q.Get("admin")),
		Directory: queryBool(q.Get("directory")),
		Sort:      q.Get("sort"),
		Dir:       q.Get("dir"),
		Page:      queryInt(q.Get("page"), 1),
		PageSize:  queryInt(q.Get("page_size"), 25),
	})
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, page)
}

func queryInt(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func queryBool(raw string) *bool {
	switch raw {
	case "true", "1":
		v := true
		return &v
	case "false", "0":
		v := false
		return &v
	default:
		return nil
	}
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

func (m *Module) resetPassword(w http.ResponseWriter, r *http.Request) {
	target, ok := pathUUID(w, r)
	if !ok {
		return
	}
	res, err := m.Svc.ResetPassword(r.Context(), auth.MustIdentity(r.Context()).UserID, target)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
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
