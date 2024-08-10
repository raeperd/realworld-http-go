package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/carlmjohnson/versioninfo"
	"github.com/raeperd/realworld"
)

func newServer(userService realworld.UserService) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /health", handleHealthCheck())
	mux.Handle("POST /api/users", handlePostUsers(userService))
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

func handlePostUsers(service realworld.UserService) http.Handler {
	// TODO: handle error
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := decode(r)
		user := req.toUser()
		user, _ = service.CreateUser(r.Context(), user)
		_ = encode(w, 201, newPostUserResponseBody(user))
	})
}

func decode[T RequestBody](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
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

type RequestBody interface {
	PostUserRequestBody
}

type ResponseBody interface {
	PostUserResponseBody | HealthCheckResponse
}

type PostUserRequestBody UserWrapper[PostUserRequest]

type PostUserResponseBody UserWrapper[PostUserResponse]

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

func newPostUserResponseBody(user realworld.User) PostUserResponseBody {
	return PostUserResponseBody{User: PostUserResponse{
		Name:  user.Name,
		Email: user.Email,
		Token: "",
		Bio:   "",
		Image: new(string)},
	}
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

type HealthCheckResponse struct {
	BuildId             string
	LastCommitHash      string
	LastCommitTimestamp int64
}
