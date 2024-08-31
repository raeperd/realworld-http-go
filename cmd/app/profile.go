package main

import (
	"net/http"

	"github.com/raeperd/realworld"
)

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
