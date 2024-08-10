package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/carlmjohnson/requests"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	port := getFreePort(t)
	go func() {
		err := run(ctx, os.Stdout, []string{"--port", port})
		if err != nil {
			fmt.Printf("failed to run in test %s\n", err)
		}
	}()

	address := "http://localhost:" + port
	t.Run("POST /api/users", func(t *testing.T) {
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
