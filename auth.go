package realworld

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type UserAuthService struct {
	repo       UserRepository
	jwtService JWTService
}

type AuthorizedUser struct {
	User
	Token string
}

func NewUserAuthService(repo UserRepository, jwtService JWTService) UserAuthService {
	return UserAuthService{repo: repo, jwtService: jwtService}
}

func (u UserAuthService) Login(ctx context.Context, email, password string) (AuthorizedUser, error) {
	user, err := u.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return AuthorizedUser{}, err
	}
	// TODO: Add password hashing
	// TODO: Add error type
	if user.Password != password {
		return AuthorizedUser{}, fmt.Errorf("invalid password")
	}
	token, err := u.jwtService.Serialize(JWTClaim{Email: email, Exp: time.Now().Add(time.Hour).Unix()})
	if err != nil {
		return AuthorizedUser{}, err
	}
	return AuthorizedUser{User: user, Token: token}, nil
}

type JWTSerializer interface {
	Header() string
	Serialize(c JWTClaim) (string, error)
}

type JWTDeserializer interface {
	Deserialize(token string) (JWTClaim, error)
}

type JWTClaim struct {
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
}

type JWTService struct {
	header string
	secret []byte
}

func NewJWTService(secret []byte) JWTService {
	header := base64.URLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	return JWTService{header: header, secret: secret}
}

func (j JWTService) Serialize(c JWTClaim) (string, error) {
	claims, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	payload := base64.URLEncoding.EncodeToString(claims)

	h := hmac.New(sha256.New, j.secret)
	if _, err := h.Write([]byte(j.header + "." + payload)); err != nil {
		return "", err
	}

	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return j.header + "." + payload + "." + signature, nil
}

func (j JWTService) Header() string {
	return j.header
}

func (j JWTService) Deserialize(token string) (JWTClaim, error) {
	parts := strings.Split(token, ".")
	// TODO: Add error for this type
	if len(parts) != 3 {
		return JWTClaim{}, fmt.Errorf("invalid token too many parts")
	}

	if parts[0] != j.header {
		return JWTClaim{}, fmt.Errorf("invalid token header want: %s got: %s", j.header, parts[0])
	}
	_, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return JWTClaim{}, fmt.Errorf("invalid token header")
	}

	payload, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return JWTClaim{}, fmt.Errorf("invalid token payload")
	}

	signature, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil {
		return JWTClaim{}, fmt.Errorf("invalid token signature")
	}

	h := hmac.New(sha256.New, j.secret)
	if _, err := h.Write([]byte(parts[0] + "." + parts[1])); err != nil {
		return JWTClaim{}, err
	}
	expectedSignature := h.Sum(nil)
	if !hmac.Equal(signature, expectedSignature) {
		return JWTClaim{}, fmt.Errorf("invalid token signature mismatch")
	}

	// Log the payload to inspect it
	fmt.Printf("Decoded payload: %s\n", string(payload))

	var claim JWTClaim
	if err := json.Unmarshal(payload, &claim); err != nil {
		return JWTClaim{}, err
	}
	return claim, nil
}
