package realworld

import (
	"errors"
	"fmt"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrBadRequest         = Error("bad request")
	ErrUserNotFound       = Error("user not found")
	ErrPasswordNotMatched = Error("password not matched")
	ErrTokenNotFound      = Error("token not found")
	ErrInvalidToken       = Error("invalid token")
)

func ErrorIfEmpty[T comparable](name string, value T) error {
	var v T
	if v == value {
		return fmt.Errorf("%w: %s is required but empty", ErrBadRequest, name)
	}
	return nil
}

func StatusFromError(err error) int {
	switch {
	case err == nil:
		return 200
	case errors.Is(err, ErrTokenNotFound):
		return 401
	case errors.Is(err, ErrUserNotFound):
		return 404
	case errors.Is(err, ErrBadRequest) || errors.Is(err, ErrPasswordNotMatched) || errors.Is(err, ErrInvalidToken):
		return 422
	default:
		return 500
	}
}
