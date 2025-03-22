package services

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
)

type IDTokenValidator interface {
	Validate(ctx context.Context, idToken string) (*idtoken.Payload, error)
}

type GoogleIDTokenValidator struct {
	googleClientID string
}

func NewGoogleIDTokenValidator() IDTokenValidator {
	return &GoogleIDTokenValidator{
		googleClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	}
}

func (tv *GoogleIDTokenValidator) Validate(ctx context.Context, idToken string) (*idtoken.Payload, error) {
	return idtoken.Validate(ctx, idToken, tv.googleClientID)
}

type MockIDTokenValidator struct{}

func NewMockIDTokenValidator() IDTokenValidator {
	return &MockIDTokenValidator{}
}

func (tv *MockIDTokenValidator) Validate(ctx context.Context, idToken string) (*idtoken.Payload, error) {
	if email, ok := strings.CutPrefix(idToken, "valid-token:"); ok {
		now := time.Now().Unix()
		return &idtoken.Payload{
			Issuer:   "https://accounts.google.com",
			Subject:  "1234",
			Audience: "myapp",
			IssuedAt: now - 1,
			Expires:  now + 59,
			Claims: map[string]interface{}{
				"email":          email,
				"email_verified": true,
				"at_hash":        "h5BZSFLbvXb8jd3ZSIX8nn",
				"name":           "John Doe",
				"picture":        "https://example.com/avatar.jpg",
				"given_name":     "John",
				"family_name":    "Doe",
			},
		}, nil
	}
	return nil, errors.New("invalid token")
}
