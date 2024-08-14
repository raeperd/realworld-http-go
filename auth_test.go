package realworld_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/raeperd/realworld"
)

func TestAuth(t *testing.T) {
	claim := realworld.JWTClaim{Email: "user@email.com", Exp: time.Now()}
	service := realworld.NewJWTService([]byte("secret"))

	token, err := service.Serialize(claim)
	assert.NoError(t, err)
	assert.NotZero(t, token)
	assert.True(t, strings.HasPrefix(token, service.Header()), "token should start with header but got '%s'", token)

	after, err := service.Deserialize(token)
	assert.NoError(t, err)
	assert.Equal(t, claim, after)

	token2, err := service.Serialize(after)
	assert.NoError(t, err)
	assert.Equal(t, token, token2)
}
