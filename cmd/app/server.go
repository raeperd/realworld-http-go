package main

import (
	"net/http"
	"strings"

	"github.com/raeperd/realworld"
)

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

func handleGetProfile(service realworld.UserService) http.Handler {
	type profile struct {
		Username  string `json:"username"`
		Bio       string `json:"bio"`
		Image     string `json:"image"`
		Following bool   `json:"following"`
	}
	type profileResponse struct {
		Profile profile `json:"profile"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")
		found, err := service.FindProfileByUsername(r.Context(), username)
		if err != nil {
			_ = encodeError(w, err)
			return
		}
		_ = encode(w, 200, profileResponse{
			Profile: profile{
				Username:  found.Username,
				Bio:       found.Bio,
				Image:     found.Image,
				Following: false,
			},
		})
	})
}
