package identity

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var ErrAuthenticationTooOld = errors.New("authentication is too old")

type Principal struct {
	UID           string `json:"uid"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"emailVerified"`
}

type SessionManager interface {
	CreateSession(context.Context, string, time.Duration, time.Duration) (string, Principal, error)
	VerifySession(context.Context, string) (Principal, error)
}

type Firebase struct {
	client *firebaseauth.Client
	now    func() time.Time
}

type Config struct {
	ProjectID             string
	CredentialsBase64JSON string
}

func NewFirebase(ctx context.Context, config Config) (*Firebase, error) {
	if config.ProjectID == "" {
		return nil, errors.New("Firebase project ID is required")
	}

	var options []option.ClientOption
	if config.CredentialsBase64JSON != "" {
		credentialsJSON, err := base64.StdEncoding.DecodeString(config.CredentialsBase64JSON)
		if err != nil {
			return nil, fmt.Errorf("decode Firebase Admin credentials: %w", err)
		}
		options = append(options, option.WithCredentialsJSON(credentialsJSON))
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: config.ProjectID}, options...)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase app: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize Firebase Auth client: %w", err)
	}

	return &Firebase{client: client, now: time.Now}, nil
}

func (f *Firebase) CreateSession(
	ctx context.Context,
	idToken string,
	duration time.Duration,
	maxAuthenticationAge time.Duration,
) (string, Principal, error) {
	token, err := f.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", Principal{}, fmt.Errorf("verify Firebase ID token: %w", err)
	}

	authenticatedAt := time.Unix(token.AuthTime, 0)
	if authenticatedAt.Before(f.now().Add(-maxAuthenticationAge)) {
		return "", Principal{}, ErrAuthenticationTooOld
	}

	cookie, err := f.client.SessionCookie(ctx, idToken, duration)
	if err != nil {
		return "", Principal{}, fmt.Errorf("create Firebase session cookie: %w", err)
	}

	return cookie, principalFromToken(token), nil
}

func (f *Firebase) VerifySession(ctx context.Context, sessionCookie string) (Principal, error) {
	token, err := f.client.VerifySessionCookie(ctx, sessionCookie)
	if err != nil {
		return Principal{}, fmt.Errorf("verify Firebase session cookie: %w", err)
	}
	return principalFromToken(token), nil
}

func principalFromToken(token *firebaseauth.Token) Principal {
	email, _ := token.Claims["email"].(string)
	emailVerified, _ := token.Claims["email_verified"].(bool)
	return Principal{
		UID:           token.UID,
		Email:         email,
		EmailVerified: emailVerified,
	}
}
