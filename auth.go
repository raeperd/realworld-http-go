package realworld

import (
	"encoding/base64"
	"time"
)

type JWTSerializer interface {
	Header() string
	Serialize(c JWTClaim) (string, error)
}

type JWTDeserializer interface {
	Deserialize(token string) (JWTClaim, error)
}

type JWTClaim struct {
	Email string
	Exp   time.Time
}

type JWTService struct {
	header string
	secret []byte
}

func NewJWTService(secret []byte) JWTService {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	return JWTService{header: header, secret: secret}
}

func (j JWTService) Serialize(c JWTClaim) (string, error) {
	return "", nil
}

func (j JWTService) Header() string {
	return j.header
}

func (j JWTService) Deserialize(token string) (JWTClaim, error) {
	return JWTClaim{}, nil
}
