package services

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrNotSubscribed      = errors.New("subscribe to this board before posting")
	ErrCommentsDisabled   = errors.New("comments are disabled on bulletin posts")
	ErrInactiveAccount    = errors.New("account is not active")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
