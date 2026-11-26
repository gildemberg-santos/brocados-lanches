package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/ratelimit"
)

func rateLimitTestRouter(orders OrderService, limits Limiters, customers ...CustomerLookupService) http.Handler {
	var customerService CustomerLookupService
	if len(customers) > 0 {
		customerService = customers[0]
	}
	return NewRouter(Dependencies{
		Pinger:           fakePinger{},
		Orders:           orders,
		Catalog:          fakeCatalogService{},
		CustomerProfiles: customerService,
		Limiters:         limits,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

func rateLimitRequest(handler http.Handler, method, path, body, remoteAddr string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.RemoteAddr = remoteAddr
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", testOrderHeaderKey)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestPublicRateLimits(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		path      string
		body      string
		limit     int
		configure func(*Limiters, *ratelimit.Limiter)
		varyIP    bool
		wantCalls func(*fakeOrderService) int
	}{
		{
			name:   "pedidos por IP",
			method: http.MethodPost,
			path:   "/api/v1/orders",
			body:   validOrderBody,
			limit:  orderPerIPLimit,
			configure: func(limits *Limiters, limiter *ratelimit.Limiter) {
				limits.Orders = limiter
			},
			wantCalls: func(orders *fakeOrderService) int { return orders.createCalls },
		},
		{
			name:   "pedidos por telefone",
			method: http.MethodPost,
			path:   "/api/v1/orders",
			body:   validOrderBody,
			limit:  orderPerPhoneLimit,
			configure: func(limits *Limiters, limiter *ratelimit.Limiter) {
				limits.OrderPhones = limiter
			},
			varyIP:    true,
			wantCalls: func(orders *fakeOrderService) int { return orders.createCalls },
		},
		{
			name:   "cotação por IP",
			method: http.MethodPost,
			path:   "/api/v1/orders/quote",
			body:   validQuoteBody,
			limit:  quotePerIPLimit,
			configure: func(limits *Limiters, limiter *ratelimit.Limiter) {
				limits.Quotes = limiter
			},
			wantCalls: func(orders *fakeOrderService) int { return orders.quoteCalls },
		},
		{
			name:   "acompanhamento por IP",
			method: http.MethodGet,
			path:   "/api/v1/orders/public-123",
			limit:  trackingPerIPLimit,
			configure: func(limits *Limiters, limiter *ratelimit.Limiter) {
				limits.Tracking = limiter
			},
			wantCalls: func(orders *fakeOrderService) int { return orders.trackCalls },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			limiter, err := ratelimit.New(test.limit, rateLimitWindow)
			if err != nil {
				t.Fatal(err)
			}
			// As demais rotas recebem limites altos para isolar a rota testada.
			limits := Limiters{
				Orders:      newLimiter(1000),
				OrderPhones: newLimiter(1000),
				Quotes:      newLimiter(1000),
				Tracking:    newLimiter(1000),
				Menu:        newLimiter(1000),
				Admin:       newLimiter(1000),
			}
			test.configure(&limits, limiter)

			orders := &fakeOrderService{
				created: createdOrderFixture(),
				quote:   quoteFixture(),
				tracked: createdOrderFixture(),
				store:   domain.StoreSettings{WhatsApp: "(85) 99999-9999"},
			}
			handler := rateLimitTestRouter(orders, limits)

			for attempt := 0; attempt < test.limit; attempt++ {
				remoteAddr := "192.0.2.1:1234"
				if test.varyIP {
					remoteAddr = fmt.Sprintf("192.0.2.%d:1234", attempt+1)
				}
				response := rateLimitRequest(handler, test.method, test.path, test.body, remoteAddr)
				if response.Code == http.StatusTooManyRequests {
					t.Fatalf("attempt %d was rate limited before the limit", attempt+1)
				}
			}

			response := rateLimitRequest(handler, test.method, test.path, test.body, "192.0.2.1:1234")
			if response.Code != http.StatusTooManyRequests {
				t.Fatalf("status = %d, want 429; body = %s", response.Code, response.Body.String())
			}
			if got := response.Header().Get("Retry-After"); got != rateLimitRetryAfter {
				t.Errorf("Retry-After = %q, want %q", got, rateLimitRetryAfter)
			}
			var body errorEnvelope
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Error.Code != codeRateLimited {
				t.Errorf("error code = %q, want %q", body.Error.Code, codeRateLimited)
			}
			if calls := test.wantCalls(orders); calls != test.limit {
				t.Errorf("service calls = %d, want %d", calls, test.limit)
			}
		})
	}
}

func TestCustomerProfileRateLimits(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		varyIP    bool
		configure func(*Limiters, *ratelimit.Limiter)
	}{
		{
			name:  "consulta por IP",
			limit: customerPerIPLimit,
			configure: func(limits *Limiters, limiter *ratelimit.Limiter) {
				limits.CustomerLookups = limiter
			},
		},
		{
			name:   "consulta por telefone",
			limit:  customerPerPhoneLimit,
			varyIP: true,
			configure: func(limits *Limiters, limiter *ratelimit.Limiter) {
				limits.CustomerPhones = limiter
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			limiter, err := ratelimit.New(test.limit, rateLimitWindow)
			if err != nil {
				t.Fatal(err)
			}
			limits := Limiters{
				Orders:          newLimiter(1000),
				OrderPhones:     newLimiter(1000),
				CustomerLookups: newLimiter(1000),
				CustomerPhones:  newLimiter(1000),
				Quotes:          newLimiter(1000),
				Tracking:        newLimiter(1000),
				Menu:            newLimiter(1000),
				Admin:           newLimiter(1000),
			}
			test.configure(&limits, limiter)

			customers := &fakeCustomerLookupService{}
			handler := rateLimitTestRouter(&fakeOrderService{}, limits, customers)
			for attempt := 0; attempt < test.limit; attempt++ {
				remoteAddr := "192.0.2.1:1234"
				if test.varyIP {
					remoteAddr = fmt.Sprintf("192.0.2.%d:1234", attempt+1)
				}
				response := rateLimitRequest(
					handler,
					http.MethodPost,
					"/api/v1/customers/profile",
					`{"phone":"85999999999"}`,
					remoteAddr,
				)
				if response.Code == http.StatusTooManyRequests {
					t.Fatalf("attempt %d was rate limited before the limit", attempt+1)
				}
			}

			response := rateLimitRequest(
				handler,
				http.MethodPost,
				"/api/v1/customers/profile",
				`{"phone":"85999999999"}`,
				"192.0.2.1:1234",
			)
			if response.Code != http.StatusTooManyRequests {
				t.Fatalf("status = %d, want 429; body = %s", response.Code, response.Body.String())
			}
			if customers.calls != test.limit {
				t.Errorf("service calls = %d, want %d", customers.calls, test.limit)
			}
		})
	}
}

func TestMenuRateLimit(t *testing.T) {
	limiter, err := ratelimit.New(2, rateLimitWindow)
	if err != nil {
		t.Fatal(err)
	}
	handler := rateLimitTestRouter(&fakeOrderService{}, Limiters{Menu: limiter})

	for attempt := 0; attempt < 2; attempt++ {
		response := rateLimitRequest(handler, http.MethodGet, "/api/v1/menu", "", "198.51.100.7:1234")
		if response.Code != http.StatusOK {
			t.Fatalf("attempt %d status = %d, want 200", attempt+1, response.Code)
		}
	}

	response := rateLimitRequest(handler, http.MethodGet, "/api/v1/menu", "", "198.51.100.7:1234")
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429; body = %s", response.Code, response.Body.String())
	}
}

func TestAdminSessionRateLimit(t *testing.T) {
	limiter, err := ratelimit.New(2, rateLimitWindow)
	if err != nil {
		t.Fatal(err)
	}
	auth := &fakeAdminAuthService{user: domain.AdminUser{
		ID:    "665f1c2a9b1e4a0012a3b4f1",
		Name:  "Dono",
		Email: "dono@teste.local",
		Role:  domain.AdminRoleOwner,
	}}
	handler := NewRouter(Dependencies{
		AdminAuth:   auth,
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		Limiters:    Limiters{Admin: limiter},
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	for attempt := 0; attempt < 2; attempt++ {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
		request.AddCookie(&http.Cookie{Name: "brocados_admin", Value: "token-opaco"})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("attempt %d status = %d, want 200", attempt+1, response.Code)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: "brocados_admin", Value: "token-opaco"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429; body = %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Retry-After"); got != rateLimitRetryAfter {
		t.Errorf("Retry-After = %q, want %q", got, rateLimitRetryAfter)
	}
}
