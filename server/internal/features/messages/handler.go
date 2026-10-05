package messages

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/cameronsralla/culdechat/server/internal/auth"
	"github.com/cameronsralla/culdechat/server/internal/httpx"
)

type Module struct {
	Svc *Service
}

func (m *Module) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireUser)
		r.Get("/messages/conversations", m.list)
		r.Get("/messages/conversations/{id}", m.get)
		r.Post("/messages/conversations/{id}/accept", m.accept)
		r.Post("/messages/conversations/{id}/decline", m.decline)
		r.Post("/messages", m.send)
	})
}

func (m *Module) list(w http.ResponseWriter, r *http.Request) {
	items, err := m.Svc.List(r.Context(), auth.MustIdentity(r.Context()).UserID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (m *Module) get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	view, err := m.Svc.Get(r.Context(), auth.MustIdentity(r.Context()).UserID, id)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, view)
}

func (m *Module) send(w http.ResponseWriter, r *http.Request) {
	var in SendInput
	if err := httpx.Bind(r, &in); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	view, err := m.Svc.Send(r.Context(), auth.MustIdentity(r.Context()).UserID, in)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, view)
}

func (m *Module) accept(w http.ResponseWriter, r *http.Request) {
	m.decide(w, r, true)
}

func (m *Module) decline(w http.ResponseWriter, r *http.Request) {
	m.decide(w, r, false)
}

func (m *Module) decide(w http.ResponseWriter, r *http.Request, accept bool) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	me := auth.MustIdentity(r.Context()).UserID
	var (
		view ConversationView
		err  error
	)
	if accept {
		view, err = m.Svc.Accept(r.Context(), me, id)
	} else {
		view, err = m.Svc.Decline(r.Context(), me, id)
	}
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, view)
}

func pathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.Fail(w, r, httpx.ErrBadRequest.WithMessage("invalid id"))
		return uuid.Nil, false
	}
	return id, true
}
