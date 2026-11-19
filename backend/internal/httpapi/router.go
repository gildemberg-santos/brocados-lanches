package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Pinger é a dependência mínima de readiness.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Dependencies reúne o que o roteador precisa para funcionar.
type Dependencies struct {
	Pinger Pinger
	Logger *slog.Logger
}

type healthResponse struct {
	Status string `json:"status"`
}

type readyResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// NewRouter monta as rotas públicas da API.
func NewRouter(deps Dependencies) http.Handler {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(requestLogger(logger))

	router.Get("/healthz", handleHealth)
	router.Get("/readyz", handleReady(deps.Pinger))

	router.Route("/api/v1", func(router chi.Router) {
		router.Get("/", handleAPIRoot)
	})

	return router
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func handleReady(pinger Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := http.StatusOK
		body := readyResponse{
			Status: "ok",
			Checks: map[string]string{"mongo": "ok"},
		}

		if pinger == nil {
			status = http.StatusServiceUnavailable
			body.Status = "degraded"
			body.Checks["mongo"] = "unavailable"
		} else if err := pinger.Ping(ctx); err != nil {
			status = http.StatusServiceUnavailable
			body.Status = "degraded"
			body.Checks["mongo"] = "error"
		}

		writeJSON(w, status, body)
	}
}

func handleAPIRoot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"name":    "Brocados Lanches API",
			"version": "v1",
		},
	})
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wrapped, r)

			logger.Info("http request",
				"request_id", middleware.GetReqID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.Status(),
				"duration_ms", time.Since(started).Milliseconds(),
			)
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
