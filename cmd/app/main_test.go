package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	port := getFreePort(t)
	go func() {
		err := run(ctx, os.Stdout, []string{"realworld", "--port", port})
		if err != nil {
			fmt.Printf("failed to run in test %s\n", err)
		}
	}()

	address := "http://localhost:" + port
	waitForHealthy(t, ctx, 2*time.Second, address+"/health")

	t.Run("GET /health", func(t *testing.T) {
		var res HealthCheckResponse
		err := requests.URL(address).Path("./health").ToJSON(&res).Fetch(ctx)

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
			err := requests.URL(address).Path("./api/users").
				BodyJSON(&PostUserRequestBody{User: tc}).
				ToJSON(&res).
				CheckStatus(422).
				Fetch(ctx)

			be.NilErr(t, err)
		}

		req := PostUserRequestBody{
			User: PostUserRequest{
				Name:     "username",
				Email:    "user@email.com",
				Password: "some-password",
			},
		}
		var res PostUserResponseBody
		err := requests.URL(address).Path("./api/users").
			BodyJSON(&req).ToJSON(&res).
			Fetch(ctx)
		if err != nil {
			t.Fatal(err)
		}

		be.Equal(t, req.User.Name, res.User.Name)
		be.Equal(t, req.User.Email, res.User.Email)
		// TODO: Delete the user after the test
	})

	// TODO: improve token handling to test be independent
	token := ""
	t.Run("POST /api/users/login", func(t *testing.T) {
		badcases := []PostUserLoginRequest{
			{Email: "", Password: "password"},
			{Email: "user@email.com", Password: ""},
			{Email: "", Password: ""},
		}

		for _, tc := range badcases {
			var res ErrorResponseBody
			err := requests.URL(address).Path("./api/users/login").
				BodyJSON(&PostUserLoginRequestBody{User: tc}).ToJSON(&res).
				CheckStatus(422).Fetch(ctx)

			be.NilErr(t, err)
		}

		req := PostUserLoginRequestBody{
			User: PostUserLoginRequest{
				Email:    "user@email.com",
				Password: "some-password",
			},
		}
		var res PostUserResponseBody
		err := requests.URL(address).Path("./api/users/login").
			BodyJSON(&req).ToJSON(&res).Fetch(ctx)

		be.NilErr(t, err)
		be.Equal(t, req.User.Email, res.User.Email)
		be.Nonzero(t, res.User.Token)

		token = res.User.Token
		// TODO: Delete the user after the test
	})

	t.Run("GET /api/user", func(t *testing.T) {
		err := requests.URL(address).Path("./api/user").CheckStatus(401).Fetch(ctx)
		be.NilErr(t, err)

		err = requests.URL(address).Path("./api/user").
			Header("Authorization", "Token invalid-token").
			CheckStatus(422).Fetch(ctx)
		be.NilErr(t, err)

		var res PostUserResponseBody
		err = requests.URL(address).Path("./api/user").
			Header("Authorization", "Token "+token).
			CheckStatus(200).ToJSON(&res).Fetch(ctx)

		be.NilErr(t, err)
		be.Equal(t, token, res.User.Token)
	})

	t.Run("GET /api/profiles/{username}", func(t *testing.T) {
		username := "username"
		err := requests.URL(address).Path("./api/profiles/" + username).
			CheckStatus(200).Fetch(ctx)
		be.NilErr(t, err)

		err = requests.URL(address).Path("./api/profiles/invalid-username").
			CheckStatus(404).Fetch(ctx)
		be.NilErr(t, err)
	})

}

func getFreePort(t *testing.T) string {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to get a free port: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return strconv.Itoa(addr.Port)
}

func waitForHealthy(t *testing.T, ctx context.Context, timeout time.Duration, endpoint string) {
	startTime := time.Now()
	for {
		err := requests.URL(endpoint).Fetch(ctx)
		if err == nil {
			t.Log("endpoint is ready")
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			if timeout <= time.Since(startTime) {
				t.Fatalf("timeout reached white waitForHealthy")
				return
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}
