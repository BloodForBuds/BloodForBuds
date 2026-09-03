package httpserver

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/BloodForBuds/BloodForBuds/services/server/internal/identity"
)

const (
	csrfCookieName       = "bfb_csrf"
	sessionCookieName    = "bfb_session"
	secureCSRFCookieName = "__Host-bfb_csrf"
	secureSessionName    = "__Host-bfb_session"
	maxRequestBodyBytes  = 16 << 10
)

type handler struct {
	sessions identity.SessionManager
	config   Config
}

type createSessionRequest struct {
	CSRFToken string `json:"csrfToken"`
	IDToken   string `json:"idToken"`
}

type csrfResponse struct {
	CSRFToken string `json:"csrfToken"`
}

type meResponse struct {
	User identity.Principal `json:"user"`
}

type principalContextKey struct{}

func (h *handler) firebaseConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.config.FirebaseWeb)
}

func (h *handler) csrf(w http.ResponseWriter, _ *http.Request) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create CSRF token")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	http.SetCookie(w, &http.Cookie{
		Name:     h.csrfCookieName(),
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		Secure:   h.config.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusOK, csrfResponse{CSRFToken: token})
}

func (h *handler) createSession(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "cross-origin request rejected")
		return
	}
	if !isJSON(r.Header.Get("Content-Type")) {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	var request createSessionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.IDToken == "" || !h.validCSRFToken(r, request.CSRFToken) {
		writeError(w, http.StatusUnauthorized, "invalid login request")
		return
	}

	sessionCookie, principal, err := h.sessions.CreateSession(
		r.Context(),
		request.IDToken,
		h.config.SessionDuration,
		h.config.MaxAuthenticationAge,
	)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired login")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.sessionCookieName(),
		Value:    sessionCookie,
		Path:     "/",
		Secure:   h.config.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusCreated, meResponse{User: principal})
}

func (h *handler) deleteSession(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		writeError(w, http.StatusForbidden, "cross-origin request rejected")
		return
	}
	csrfToken := r.Header.Get("X-CSRF-Token")
	if !h.validCSRFToken(r, csrfToken) {
		writeError(w, http.StatusUnauthorized, "invalid logout request")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.sessionCookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   h.config.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) me(w http.ResponseWriter, r *http.Request) {
	principal, ok := r.Context().Value(principalContextKey{}).(identity.Principal)
	if !ok {
		writeError(w, http.StatusInternalServerError, "authenticated user unavailable")
		return
	}
	writeJSON(w, http.StatusOK, meResponse{User: principal})
}

func (h *handler) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(h.sessionCookieName())
		if err != nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		principal, err := h.sessions.VerifySession(r.Context(), cookie.Value)
		if err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:     h.sessionCookieName(),
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				Secure:   h.config.CookieSecure,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		ctx := context.WithValue(r.Context(), principalContextKey{}, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *handler) validCSRFToken(r *http.Request, token string) bool {
	if token == "" {
		return false
	}
	cookie, err := r.Cookie(h.csrfCookieName())
	if err != nil || len(cookie.Value) != len(token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) == 1
}

func (h *handler) csrfCookieName() string {
	if h.config.CookieSecure {
		return secureCSRFCookieName
	}
	return csrfCookieName
}

func (h *handler) sessionCookieName() string {
	if h.config.CookieSecure {
		return secureSessionName
	}
	return sessionCookieName
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return strings.EqualFold(parsed.Host, r.Host)
}

func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "application/json"
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
