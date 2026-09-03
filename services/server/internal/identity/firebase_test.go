package identity

import (
	"context"
	"testing"

	firebaseauth "firebase.google.com/go/v4/auth"
)

func TestNewFirebaseValidatesConfiguration(t *testing.T) {
	t.Run("project ID", func(t *testing.T) {
		if _, err := NewFirebase(context.Background(), Config{}); err == nil {
			t.Fatal("expected an empty project ID to be rejected")
		}
	})

	t.Run("credentials", func(t *testing.T) {
		if _, err := NewFirebase(context.Background(), Config{
			ProjectID:             "test-project",
			CredentialsBase64JSON: "not-base64",
		}); err == nil {
			t.Fatal("expected malformed base64 credentials to be rejected")
		}
	})
}

func TestPrincipalFromToken(t *testing.T) {
	principal := principalFromToken(&firebaseauth.Token{
		UID: "firebase-user",
		Claims: map[string]any{
			"email":          "donor@example.com",
			"email_verified": true,
		},
	})

	if principal.UID != "firebase-user" || principal.Email != "donor@example.com" || !principal.EmailVerified {
		t.Fatalf("unexpected principal: %#v", principal)
	}
}
