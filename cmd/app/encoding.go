package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/raeperd/realworld"
)

func decode[T RequestBody](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("%w failed to decode json: %w", realworld.ErrBadRequest, err)
	}
	return v, nil
}

func encode[T ResponseBody](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func encodeError(w http.ResponseWriter, err error) error {
	if errors, ok := err.(interface{ Unwrap() []error }); ok {
		return encode(w, realworld.StatusFromError(err), NewErrorResponseBody(errors.Unwrap()...))
	}
	return encode(w, realworld.StatusFromError(err), NewErrorResponseBody(err))
}

type RequestBody interface {
	PostUserRequestBody | PostUserLoginRequestBody
}

type ResponseBody interface {
	ErrorResponseBody | PostUserResponseBody | HealthCheckResponse
}

type ErrorResponseBody struct {
	Errors struct {
		Body []string `json:"body"`
	} `json:"errors"`
}

func NewErrorResponseBody(errors ...error) ErrorResponseBody {
	responses := make([]string, len(errors))
	for i, err := range errors {
		responses[i] = err.Error()
	}
	return ErrorResponseBody{
		Errors: struct {
			Body []string `json:"body"`
		}{responses},
	}
}

type PostUserRequestBody UserWrapper[PostUserRequest]

type PostUserResponseBody UserWrapper[PostUserResponse]

type PostUserLoginRequestBody UserWrapper[PostUserLoginRequest]

type GetProfilesResponseBody struct {
	Profile struct {
		Username  string  `json:"username"`
		Bio       string  `json:"bio"`
		Image     *string `json:"image"`
		Following bool    `json:"following"`
	} `json:"profile"`
}

// TODO: Remove this struct by implementing [json.Marshaler]
type UserWrapper[T any] struct {
	User T `json:"user"`
}

func (r PostUserRequestBody) toUser() realworld.User {
	return realworld.User{
		Name:     r.User.Name,
		Email:    r.User.Email,
		Password: r.User.Password,
	}
}

// TODO: Add detailed validation such as email format, password strength etc
func (r PostUserRequestBody) Valid() error {
	return errors.Join(
		realworld.ErrorIfEmpty("username", r.User.Name),
		realworld.ErrorIfEmpty("email", r.User.Email),
		realworld.ErrorIfEmpty("password", r.User.Password),
	)
}

func newPostUserResponseBody(user realworld.AuthenticatedUser) PostUserResponseBody {
	return PostUserResponseBody{User: PostUserResponse{
		Name:  user.Name,
		Email: user.Email,
		Token: user.Token,
		Bio:   "",
		Image: new(string)},
	}
}

func (r PostUserLoginRequestBody) Valid() error {
	return errors.Join(
		realworld.ErrorIfEmpty("email", r.User.Email),
		realworld.ErrorIfEmpty("password", r.User.Password),
	)

}

type PostUserRequest struct {
	Name     string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PostUserResponse struct {
	Name  string  `json:"username"`
	Email string  `json:"email"`
	Token string  `json:"token"`
	Bio   string  `json:"bio"`
	Image *string `json:"image"`
}

type PostUserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type HealthCheckResponse struct {
	BuildId             string
	LastCommitHash      string
	LastCommitTimestamp int64
}
