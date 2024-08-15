package main

import (
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
