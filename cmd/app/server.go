package main

import (
	"net/http"

	"github.com/raeperd/realworld"
)

func newServer(userService realworld.UserService) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /api/users", handlePostUsers(userService))
	return mux
}

func handlePostUsers(_ realworld.UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: implements this
	})
}

// TODO: Remove this struct by implementing [json.Marshaler]
type UserWrapper[T any] struct {
	User T `json:"user"`
}

type PostUserRequestBody UserWrapper[PostUserRequest]

type PostUserResponseBody UserWrapper[PostUserResponse]

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
