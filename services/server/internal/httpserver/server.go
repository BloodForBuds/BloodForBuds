package httpserver

import (
	"net/http"
	"time"

	"github.com/BloodForBuds/BloodForBuds/services/server/internal/identity"
)

type Config struct {
	CookieSecure         bool
	SessionDuration      time.Duration
	MaxAuthenticationAge time.Duration
	FirebaseWeb          FirebaseWebConfig
}

type FirebaseWebConfig struct {
	APIKey      string `json:"apiKey"`
	AppID       string `json:"appId"`
	AuthDomain  string `json:"authDomain"`
	ProjectID   string `json:"projectId"`
	UseEmulator bool   `json:"useEmulator"`
}

func NewHandler(sessions identity.SessionManager, config Config) http.Handler {
	handler := &handler{
		sessions: sessions,
		config:   config,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /auth/config", handler.firebaseConfig)
	mux.HandleFunc("GET /auth/csrf", handler.csrf)
	mux.HandleFunc("POST /auth/session", handler.createSession)
	mux.HandleFunc("DELETE /auth/session", handler.deleteSession)
	mux.Handle("GET /auth/me", handler.requireSession(http.HandlerFunc(handler.me)))
	return mux
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
}
