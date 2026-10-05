// Package httpx holds the HTTP conventions every handler uses: JSON responses,
// a typed error model mapped to status codes, and request body binding.
package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	applog "github.com/cameronsralla/culdechat/server/internal/log"
)

// Error is the typed application error. Services return these; handlers pass
// them to Fail and the client receives {"error": {"code": "...", "message": "..."}}.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.cause }

// Is lets errors.Is(err, httpx.ErrNotFound) match copies made by Wrap/WithMessage.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t.Code == e.Code
}

// Wrap attaches an internal cause (logged, never sent to the client).
func (e *Error) Wrap(err error) *Error {
	c := *e
	c.cause = err
	return &c
}

// WithMessage returns a copy with a different client-facing message.
func (e *Error) WithMessage(msg string) *Error {
	c := *e
	c.Message = msg
	return &c
}

var (
	ErrBadRequest   = &Error{Status: http.StatusBadRequest, Code: "bad_request", Message: "invalid request"}
	ErrUnauthorized = &Error{Status: http.StatusUnauthorized, Code: "unauthorized", Message: "authentication required"}
	ErrForbidden    = &Error{Status: http.StatusForbidden, Code: "forbidden", Message: "not allowed"}
	ErrNotFound     = &Error{Status: http.StatusNotFound, Code: "not_found", Message: "not found"}
	ErrConflict     = &Error{Status: http.StatusConflict, Code: "conflict", Message: "already exists"}
	ErrTooLarge     = &Error{Status: http.StatusRequestEntityTooLarge, Code: "too_large", Message: "request too large"}
	ErrRateLimited  = &Error{Status: http.StatusTooManyRequests, Code: "rate_limited", Message: "too many requests"}
	ErrInternal     = &Error{Status: http.StatusInternalServerError, Code: "internal", Message: "something went wrong"}
)

type errorBody struct {
	Error *Error `json:"error"`
}

// JSON writes v as JSON with the given status.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// NoContent writes 204.
func NoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

// Fail maps err to an HTTP error response. Unknown errors become 500 and are logged.
func Fail(w http.ResponseWriter, r *http.Request, err error) {
	var e *Error
	if !errors.As(err, &e) {
		e = ErrInternal.Wrap(err)
	}
	if e.Status >= 500 {
		applog.From(r.Context()).Error("request failed", "err", err, "code", e.Code)
	}
	JSON(w, e.Status, errorBody{Error: e})
}

// Validator is implemented by request types that validate themselves.
type Validator interface{ Validate() error }

// Bind decodes a JSON body into v, rejecting unknown fields and trailing data,
// then runs Validate if v implements Validator.
func Bind(r *http.Request, v any) error {
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/json") {
		return ErrBadRequest.WithMessage("content-type must be application/json")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return ErrTooLarge
		}
		return ErrBadRequest.WithMessage("malformed json").Wrap(err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return ErrBadRequest.WithMessage("unexpected trailing data")
	}
	if val, ok := v.(Validator); ok {
		if err := val.Validate(); err != nil {
			return ErrBadRequest.WithMessage(err.Error())
		}
	}
	return nil
}

// Fields collects per-field validation problems.
type Fields struct{ problems []string }

func (f *Fields) Add(field, problem string) { f.problems = append(f.problems, field+": "+problem) }

func (f *Fields) Err() error {
	if len(f.problems) == 0 {
		return nil
	}
	return errors.New(strings.Join(f.problems, "; "))
}
