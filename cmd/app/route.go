package main

import (
	"log"
	"net/http"
	"time"

	"github.com/carlmjohnson/versioninfo"
	"github.com/raeperd/realworld"
)

func newRouter(userService realworld.UserService, authService realworld.UserAuthService) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /health", handleHealthCheck())
	mux.Handle("POST /api/users", handlePostUsers(userService, authService))
	mux.Handle("POST /api/users/login", handlePostUsersLogin(authService))
	mux.Handle("GET /api/user", handleGetUser(authService))
	mux.Handle("GET /api/profiles/{username}", handleGetProfile(userService))
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
