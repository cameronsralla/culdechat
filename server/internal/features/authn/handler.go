package authn

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/cameronsralla/culdechat/server/internal/auth"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
	"github.com/cameronsralla/culdechat/server/internal/middleware"
)

type Module struct {
	Svc *Service
	// Strict limits credential endpoints by IP (brute-force protection).
	Strict *middleware.Limiter
}

func (m *Module) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(m.Strict.ByIP)
		r.Post("/auth/login", m.login)
		r.Post("/auth/refresh", m.refresh)
		r.Post("/auth/logout", m.logout)
		r.Get("/auth/invite", m.peekInvite)
		r.Post("/auth/complete-invite", m.completeInvite)
	})
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireUser)
		r.Post("/auth/logout-all", m.logoutAll)
		r.Post("/auth/change-password", m.changePassword)
	})
}

func client(r *http.Request) Client {
	return Client{UserAgent: r.UserAgent(), IP: middleware.ClientIP(r)}
}

func (m *Module) login(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	sess, err := m.Svc.Login(r.Context(), in, client(r))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, sess)
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token"`
}

func (m *Module) refresh(w http.ResponseWriter, r *http.Request) {
	var in refreshBody
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if in.RefreshToken == "" {
		httpx.Fail(w, r, httpx.ErrBadRequest.WithMessage("refresh_token is required"))
		return
	}
	sess, err := m.Svc.Refresh(r.Context(), in.RefreshToken, client(r))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, sess)
}

func (m *Module) logout(w http.ResponseWriter, r *http.Request) {
	var in refreshBody
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if err := m.Svc.Logout(r.Context(), in.RefreshToken); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.NoContent(w)
}

func (m *Module) logoutAll(w http.ResponseWriter, r *http.Request) {
	if err := m.Svc.LogoutAll(r.Context(), auth.MustIdentity(r.Context()).UserID); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.NoContent(w)
}

func (m *Module) peekInvite(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("token")
	if tok == "" {
		httpx.Fail(w, r, httpx.ErrBadRequest.WithMessage("token is required"))
		return
	}
	u, err := m.Svc.PeekInvite(r.Context(), tok)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"email": u.Email, "unit_number": u.UnitNumber, "display_name": u.DisplayName})
}

func (m *Module) completeInvite(w http.ResponseWriter, r *http.Request) {
	var in CompleteInviteInput
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	sess, err := m.Svc.CompleteInvite(r.Context(), in, client(r))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, sess)
}

func (m *Module) changePassword(w http.ResponseWriter, r *http.Request) {
	var in ChangePasswordInput
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	sess, err := m.Svc.ChangePassword(r.Context(), auth.MustIdentity(r.Context()).UserID, in, client(r))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, sess)
}
