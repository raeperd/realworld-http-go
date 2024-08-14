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
	if user.Password != password {
		return AuthorizedUser{}, fmt.Errorf("%w with email %s", ErrPasswordNotMatched, email)
	}
	token, err := u.jwtService.Serialize(JWTClaim{Email: email, Exp: time.Now().Add(time.Hour).Unix()})
	if err != nil {
		return AuthorizedUser{}, err
	}
	return AuthorizedUser{User: user, Token: token}, nil
}

// TODO: Authenticate or Authorize?
func (u UserAuthService) Authenticate(ctx context.Context, token string) (AuthorizedUser, error) {
	claim, err := u.jwtService.Deserialize(token)
	if err != nil {
		return AuthorizedUser{}, err
	}
	user, err := u.repo.FindUserByEmail(ctx, claim.Email)
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
	if token == "" {
		return JWTClaim{}, ErrTokenNotFound
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return JWTClaim{}, fmt.Errorf("%w invalid token with %d parts: '%s'", ErrInvalidToken, len(parts), token)
	}

	if parts[0] != j.header {
		return JWTClaim{}, fmt.Errorf("%w header want: %s got: %s", ErrInvalidToken, j.header, parts[0])
	}
	_, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return JWTClaim{}, fmt.Errorf("%w header", ErrInvalidToken)
	}

	payload, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return JWTClaim{}, fmt.Errorf("%w payload", ErrInvalidToken)
	}

	signature, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil {
		return JWTClaim{}, fmt.Errorf("%w signature", ErrInvalidToken)
	}

	h := hmac.New(sha256.New, j.secret)
	if _, err := h.Write([]byte(parts[0] + "." + parts[1])); err != nil {
		return JWTClaim{}, err
	}
	expectedSignature := h.Sum(nil)
	if !hmac.Equal(signature, expectedSignature) {
		return JWTClaim{}, fmt.Errorf("%w signature mismatch", ErrInvalidToken)
	}

	var claim JWTClaim
	if err := json.Unmarshal(payload, &claim); err != nil {
		return JWTClaim{}, err
	}
	return claim, nil
}
