package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
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

		assert.NoError(t, err)
		assert.NotZero(t, res.LastCommitHash)
		assert.NotZero(t, res.LastCommitTimestamp)
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

			assert.NoError(t, err, "%s in case %v", err, tc)
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

		assert.Equal(t, req.User.Name, res.User.Name)
		assert.Equal(t, req.User.Email, res.User.Email)
		// TODO: Delete the user after the test
	})

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

			assert.NoError(t, err, "%s in case %v", err, tc)
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

		assert.NoError(t, err)
		assert.Equal(t, req.User.Email, res.User.Email)
		assert.NotZero(t, res.User.Token)

		// TODO: Delete the user after the test
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
