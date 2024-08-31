package main

import (
	"context"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
)

func TestUser(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	testUser, password := postTestUser(t, ctx, endpoint())

	t.Run("GET /health", func(t *testing.T) {
		var res HealthCheckResponse
		err := requests.URL(endpoint()).Path("./health").ToJSON(&res).Fetch(ctx)

		be.NilErr(t, err)
		be.Nonzero(t, res.LastCommitHash)
		be.Nonzero(t, res.LastCommitTimestamp)
	})

	t.Run("POST /api/users", func(t *testing.T) {
		badcases := []PostUserRequest{
			{Name: "", Email: "user@email.com", Password: "password"},
			{Name: "username", Email: "", Password: "password"},
			{Name: "username", Email: "user@email.com", Password: ""},
			{Name: "username", Email: "", Password: ""},
			{Name: "", Email: "user@email.com", Password: ""},
			{Name: "", Email: "", Password: "password"},
			{Name: "", Email: "", Password: ""},
		}

		for _, tc := range badcases {
			var res ErrorResponseBody
			err := requests.URL(endpoint()).Path("./api/users").
				BodyJSON(&PostUserRequestBody{User: tc}).
				ToJSON(&res).
				CheckStatus(422).
				Fetch(ctx)

			be.NilErr(t, err)
		}

		req := PostUserRequestBody{User: PostUserRequest{
			Name:     testUser.Name,
			Email:    testUser.Email,
			Password: password,
		}}
		var res PostUserResponseBody
		err := requests.URL(endpoint()).Path("./api/users").
			BodyJSON(&req).ToJSON(&res).
			Fetch(ctx)
		if err != nil {
			t.Fatal(err)
		}

		be.Equal(t, req.User.Name, res.User.Name)
		be.Equal(t, req.User.Email, res.User.Email)
	})

	t.Run("POST /api/users/login", func(t *testing.T) {
		badcases := []PostUserLoginRequest{
			{Email: "", Password: "password"},
			{Email: "user@email.com", Password: ""},
			{Email: "", Password: ""},
		}

		for _, tc := range badcases {
			var res ErrorResponseBody
			err := requests.URL(endpoint()).Path("./api/users/login").
				BodyJSON(&PostUserLoginRequestBody{User: tc}).ToJSON(&res).
				CheckStatus(422).Fetch(ctx)

			be.NilErr(t, err)
		}

		req := PostUserLoginRequestBody{User: PostUserLoginRequest{
			Email:    testUser.Email,
			Password: password,
		}}
		var res PostUserResponseBody
		err := requests.URL(endpoint()).Path("./api/users/login").
			BodyJSON(&req).ToJSON(&res).Fetch(ctx)

		be.NilErr(t, err)
		be.Equal(t, req.User.Email, res.User.Email)
		be.Nonzero(t, res.User.Token)
	})

	t.Run("GET /api/user", func(t *testing.T) {
		err := requests.URL(endpoint()).Path("./api/user").CheckStatus(401).Fetch(ctx)
		be.NilErr(t, err)

		err = requests.URL(endpoint()).Path("./api/user").
			Header("Authorization", "Token invalid-token").
			CheckStatus(422).Fetch(ctx)
		be.NilErr(t, err)

		var res PostUserResponseBody
		err = requests.URL(endpoint()).Path("./api/user").
			Header("Authorization", "Token "+testUser.Token).
			CheckStatus(200).ToJSON(&res).Fetch(ctx)

		be.NilErr(t, err)
		be.Equal(t, testUser.Token, res.User.Token)
	})

	t.Run("GET /api/profiles/{username}", func(t *testing.T) {
		var errRes ErrorResponseBody
		err := requests.URL(endpoint()).Path("./api/profiles/unknown-user").
			CheckStatus(404).ToJSON(&errRes).Fetch(ctx)
		be.NilErr(t, err)
		be.Nonzero(t, errRes.Errors)

		var res GetProfilesResponseBody
		err = requests.URL(endpoint()).Path("./api/profiles/" + testUser.Name).
			CheckStatus(200).ToJSON(&res).Fetch(ctx)
		be.NilErr(t, err)
		be.Equal(t, testUser.Name, res.Profile.Username)
	})

}

func postTestUser(t *testing.T, ctx context.Context, address string) (PostUserResponse, string) {
	req := PostUserRequestBody{
		User: PostUserRequest{
			Name:     "testuser",
			Email:    "testuser@email.com",
			Password: "testuser-password",
		},
	}
	var res PostUserResponseBody
	err := requests.URL(address).Path("./api/users").
		BodyJSON(&req).ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return res.User, req.User.Password
}
