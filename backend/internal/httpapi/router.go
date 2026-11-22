package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/ratelimit"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Pinger é a dependência mínima de readiness.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Dependencies reúne o que o roteador precisa para funcionar.
type Dependencies struct {
	Pinger       Pinger
	Catalog      CatalogService
	Orders       OrderService
	AdminAuth    AdminAuthService
	AdminCookie  AdminCookieConfig
	AdminCatalog AdminCatalogService
	AdminStore   AdminStoreService
	AdminOrders  AdminOrderService
	LoginLimiter *ratelimit.Limiter
	Logger       *slog.Logger
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
	// ClientIPFromRemoteAddr não confia em X-Forwarded-For/X-Real-IP, que podem
	// ser forjados pelo cliente e burlar o rate limit de login.
	router.Use(middleware.ClientIPFromRemoteAddr)
	router.Use(middleware.Recoverer)
	router.Use(requestLogger(logger))

	router.Get("/healthz", handleHealth)
	router.Get("/readyz", handleReady(deps.Pinger))

	router.Route("/api/v1", func(router chi.Router) {
		router.Get("/", handleAPIRoot)
		router.Get("/menu", handleMenu(deps.Catalog, logger))
		router.Get("/delivery-zones", handleDeliveryZones(deps.Catalog, logger))
		router.Post("/orders/quote", handleQuote(deps.Orders, logger))
		router.Post("/orders", handleCreateOrder(deps.Orders, logger))
		router.Get("/orders/{publicId}", handleTrackOrder(deps.Orders, logger))
	})

	if deps.AdminAuth != nil {
		cookieConfig := deps.AdminCookie
		if cookieConfig.Name == "" {
			cookieConfig.Name = "brocados_admin"
		}
		limiter := deps.LoginLimiter
		if limiter == nil {
			// Parâmetros constantes e válidos; erro seria apenas de programação.
			limiter, _ = ratelimit.New(5, time.Minute)
		}

		router.Route("/api/v1/admin", func(admin chi.Router) {
			admin.Use(csrfProtection)
			admin.Post("/auth/login", handleAdminLogin(deps.AdminAuth, cookieConfig, limiter, logger))
			admin.Post("/auth/logout", handleAdminLogout(deps.AdminAuth, cookieConfig, logger))

			admin.Group(func(protected chi.Router) {
				protected.Use(requireAdmin(deps.AdminAuth, cookieConfig, logger))
				protected.Get("/auth/me", handleAdminMe(logger))
			})

			admin.Group(func(protected chi.Router) {
				protected.Use(requireAdmin(deps.AdminAuth, cookieConfig, logger, domain.AdminRoleOwner, domain.AdminRoleOperator))

				protected.Get("/categories", handleAdminListCategories(deps.AdminCatalog, logger))
				protected.Post("/categories", handleAdminCreateCategory(deps.AdminCatalog, logger))
				protected.Patch("/categories/{id}", handleAdminUpdateCategory(deps.AdminCatalog, logger))

				protected.Get("/products", handleAdminListProducts(deps.AdminCatalog, logger))
				protected.Post("/products", handleAdminCreateProduct(deps.AdminCatalog, logger))
				protected.Patch("/products/{id}", handleAdminUpdateProduct(deps.AdminCatalog, logger))

				protected.Get("/addons", handleAdminListAddons(deps.AdminCatalog, logger))
				protected.Post("/addons", handleAdminCreateAddon(deps.AdminCatalog, logger))
				protected.Patch("/addons/{id}", handleAdminUpdateAddon(deps.AdminCatalog, logger))

				protected.Get("/delivery-zones", handleAdminListZones(deps.AdminCatalog, logger))
				protected.Post("/delivery-zones", handleAdminCreateZone(deps.AdminCatalog, logger))
				protected.Patch("/delivery-zones/{id}", handleAdminUpdateZone(deps.AdminCatalog, logger))

				protected.Get("/store", handleAdminGetStore(deps.AdminStore, logger))
				protected.Patch("/store", handleAdminUpdateStore(deps.AdminStore, logger))

				protected.Get("/orders", handleAdminListOrders(deps.AdminOrders, logger))
				protected.Get("/orders/{id}", handleAdminGetOrder(deps.AdminOrders, logger))
				protected.Patch("/orders/{id}/status", handleAdminUpdateOrderStatus(deps.AdminOrders, logger))
			})
		})
	}

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
