package realworld_test

import (
	"strings"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/raeperd/realworld"
)

func TestAuth(t *testing.T) {
	claim := realworld.JWTClaim{Email: "user@email.com", Exp: time.Now().Unix()}
	service := realworld.NewJWTService([]byte("secret"))

	token, err := service.Serialize(claim)
	be.NilErr(t, err)
	be.Nonzero(t, token)
	be.True(t, strings.HasPrefix(token, service.Header()))

	after, err := service.Deserialize(token)
	be.NilErr(t, err)
	be.Equal(t, claim, after)

	token2, err := service.Serialize(after)
	be.NilErr(t, err)
	be.Equal(t, token, token2)
}
