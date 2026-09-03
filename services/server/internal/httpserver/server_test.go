package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BloodForBuds/BloodForBuds/services/server/internal/identity"
)

var testPrincipal = identity.Principal{
	UID:           "firebase-user",
	Email:         "donor@example.com",
	EmailVerified: true,
}

type fakeSessions struct {
	create func(context.Context, string, time.Duration, time.Duration) (string, identity.Principal, error)
	verify func(context.Context, string) (identity.Principal, error)
}

func (f fakeSessions) CreateSession(
	ctx context.Context,
	token string,
	duration time.Duration,
	maxAge time.Duration,
) (string, identity.Principal, error) {
	if f.create != nil {
		return f.create(ctx, token, duration, maxAge)
	}
	return "firebase-session", testPrincipal, nil
}

func (f fakeSessions) VerifySession(ctx context.Context, cookie string) (identity.Principal, error) {
	if f.verify != nil {
		return f.verify(ctx, cookie)
	}
	return testPrincipal, nil
}

func testHandler(sessions identity.SessionManager) http.Handler {
	return NewHandler(sessions, Config{
		CookieSecure:         true,
		SessionDuration:      12 * time.Hour,
		MaxAuthenticationAge: 5 * time.Minute,
		FirebaseWeb: FirebaseWebConfig{
			APIKey:      "fake-api-key",
			AppID:       "fake-app-id",
			AuthDomain:  "example.test",
			ProjectID:   "demo-bloodforbuds",
			UseEmulator: true,
		},
	})
}

func TestHealthz(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	testHandler(fakeSessions{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected application/json content type, got %q", contentType)
	}
	if body := response.Body.String(); body != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected response body: %q", body)
	}
}

func TestFirebaseConfig(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/auth/config", nil)
	response := httptest.NewRecorder()

	testHandler(fakeSessions{}).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, `"projectId":"demo-bloodforbuds"`) {
		t.Fatalf("unexpected response body: %q", body)
	}
}

func TestCreateSession(t *testing.T) {
	handler := testHandler(fakeSessions{
		create: func(_ context.Context, token string, duration, maxAge time.Duration) (string, identity.Principal, error) {
			if token != "firebase-id-token" {
				t.Fatalf("unexpected ID token %q", token)
			}
			if duration != 12*time.Hour || maxAge != 5*time.Minute {
				t.Fatalf("unexpected session policy duration=%s maxAge=%s", duration, maxAge)
			}
			return "firebase-session", testPrincipal, nil
		},
	})
	csrfToken, csrfCookie := requestCSRFToken(t, handler)

	body := `{"idToken":"firebase-id-token","csrfToken":"` + csrfToken + `"}`
	request := httptest.NewRequest(http.MethodPost, "https://example.test/auth/session", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://example.test")
	request.AddCookie(csrfCookie)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	sessionCookie := findCookie(t, response.Result().Cookies(), secureSessionName)
	if sessionCookie.Value != "firebase-session" || !sessionCookie.HttpOnly || !sessionCookie.Secure {
		t.Fatalf("unexpected session cookie: %#v", sessionCookie)
	}
	if sessionCookie.SameSite != http.SameSiteStrictMode || sessionCookie.Path != "/" {
		t.Fatalf("session cookie is missing host protections: %#v", sessionCookie)
	}
}

func TestCreateSessionRejectsCrossOriginRequest(t *testing.T) {
	handler := testHandler(fakeSessions{})
	csrfToken, csrfCookie := requestCSRFToken(t, handler)
	body := `{"idToken":"firebase-id-token","csrfToken":"` + csrfToken + `"}`
	request := httptest.NewRequest(http.MethodPost, "https://example.test/auth/session", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://attacker.test")
	request.AddCookie(csrfCookie)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestMeRequiresAValidSession(t *testing.T) {
	handler := testHandler(fakeSessions{
		verify: func(_ context.Context, cookie string) (identity.Principal, error) {
			if cookie != "firebase-session" {
				return identity.Principal{}, errors.New("invalid cookie")
			}
			return testPrincipal, nil
		},
	})

	unauthenticated := httptest.NewRecorder()
	handler.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/auth/me", nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated status %d, got %d", http.StatusUnauthorized, unauthenticated.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: secureSessionName, Value: "firebase-session"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if body := response.Body.String(); !strings.Contains(body, `"uid":"firebase-user"`) {
		t.Fatalf("unexpected response body: %q", body)
	}
}

func TestDeleteSessionRequiresCSRFToken(t *testing.T) {
	handler := testHandler(fakeSessions{})
	csrfToken, csrfCookie := requestCSRFToken(t, handler)
	request := httptest.NewRequest(http.MethodDelete, "https://example.test/auth/session", nil)
	request.Header.Set("Origin", "https://example.test")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(csrfCookie)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, response.Code, response.Body.String())
	}
	cleared := findCookie(t, response.Result().Cookies(), secureSessionName)
	if cleared.MaxAge != -1 {
		t.Fatalf("expected session cookie to be removed, got %#v", cleared)
	}
}

func requestCSRFToken(t *testing.T, handler http.Handler) (string, *http.Cookie) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "https://example.test/auth/csrf", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("get CSRF token: status %d: %s", response.Code, response.Body.String())
	}
	var body csrfResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode CSRF response: %v", err)
	}
	return body.CSRFToken, findCookie(t, response.Result().Cookies(), secureCSRFCookieName)
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found in %#v", name, cookies)
	return nil
}
