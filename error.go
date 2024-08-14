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
)

func ErrorIfEmpty[T comparable](name string, value T) error {
	var v T
	if v == value {
		return fmt.Errorf("%w: %s is required but empty", ErrBadRequest, name)
	}
	return nil
}

func StatusFromError(err error) int {
	if err == nil {
		return 200
	}
	if errors.Is(err, ErrUserNotFound) {
		return 404
	}
	if errors.Is(err, ErrBadRequest) || errors.Is(err, ErrPasswordNotMatched) {
		return 422
	}
	return 500
}
