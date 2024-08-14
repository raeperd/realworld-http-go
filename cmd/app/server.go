package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/carlmjohnson/versioninfo"
	"github.com/raeperd/realworld"
)

func newServer(userService realworld.UserService, authService realworld.UserAuthService) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /health", handleHealthCheck())
	mux.Handle("POST /api/users", handlePostUsers(userService, authService))
	mux.Handle("POST /api/users/login", handlePostUsersLogin(authService))
	mux.Handle("GET /api/user", handleGetUser(authService))
	return loggingMiddleware(mux)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s %s %s", r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
	})
}

var BuildId string

func handleHealthCheck() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = encode(w, 200, HealthCheckResponse{
			BuildId:             BuildId,
			LastCommitHash:      versioninfo.Revision,
			LastCommitTimestamp: versioninfo.LastCommit.Unix(),
		})
	})
}

func handlePostUsers(service realworld.UserService, auth realworld.UserAuthService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[PostUserRequestBody](r)
		if err != nil {
			_ = encodeError(w, err)
			return
		}
		if err := req.Valid(); err != nil {
			_ = encodeError(w, err)
			return
		}
		user, err := service.CreateUser(r.Context(), req.toUser())
		if err != nil {
			_ = encodeError(w, err)
			return
		}
		authUser, err := auth.Login(r.Context(), user.Email, user.Password)
		if err != nil {
			_ = encodeError(w, err)
			return
		}
		_ = encode(w, 201, newPostUserResponseBody(authUser))
	})
}

func handlePostUsersLogin(service realworld.UserAuthService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[PostUserLoginRequestBody](r)
		if err != nil {
			_ = encodeError(w, err)
			return
		}
		if err := req.Valid(); err != nil {
			_ = encodeError(w, err)
			return
		}
		user, err := service.Login(r.Context(), req.User.Email, req.User.Password)
		if err != nil {
			_ = encodeError(w, err)
			return
		}
		_ = encode(w, 200, newPostUserResponseBody(user))
		// TODO: validate req
	})
}

func handleGetUser(auth realworld.UserAuthService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Token ")
		user, err := auth.Authenticate(r.Context(), token)
		if err != nil {
			_ = encodeError(w, err)
			return
		}
		_ = encode(w, 200, newPostUserResponseBody(user))
	})
}

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
