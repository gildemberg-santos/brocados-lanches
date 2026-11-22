package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/ratelimit"
	"github.com/brocados-lanches/api/internal/platform/token"
	"github.com/go-chi/chi/v5/middleware"
)

const maxAdminAuthBodyBytes = 1 << 16

// AdminCookieConfig descreve o cookie da sessão administrativa.
type AdminCookieConfig struct {
	Name   string
	Secure bool
}

// AdminAuthService é a dependência dos handlers de autenticação do painel.
type AdminAuthService interface {
	Login(ctx context.Context, email, password string) (domain.AdminUser, string, error)
	Authenticate(ctx context.Context, rawToken string) (domain.AdminUser, error)
	Logout(ctx context.Context, rawToken string) error
}

type adminAuthEnvelope struct {
	Data adminAuthData `json:"data"`
}

type adminAuthData struct {
	User adminUserData `json:"user"`
}

type adminUserData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type adminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type adminPrincipalKey struct{}

func handleAdminLogin(auth AdminAuthService, cfg AdminCookieConfig, limiter *ratelimit.Limiter, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth == nil {
			writeError(w, r, logger, errors.New("admin auth service is not configured"))
			return
		}

		request, err := decodeAdminLoginRequest(w, r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		// A chave usa o hash do e-mail para não manter PII em memória.
		emailKey := strings.ToLower(strings.TrimSpace(request.Email))
		rateKey := clientIP(r) + "|" + token.Hash(emailKey)
		if limiter != nil && !limiter.Allow(rateKey) {
			w.Header().Set("Retry-After", "60")
			writeError(w, r, logger, domain.ErrRateLimited)
			return
		}

		user, rawToken, err := auth.Login(r.Context(), request.Email, request.Password)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}

		http.SetCookie(w, newAdminCookie(cfg, rawToken, false))
		writeJSON(w, http.StatusOK, adminAuthEnvelope{Data: adminAuthData{User: newAdminUserData(user)}})
	}
}

func handleAdminLogout(auth AdminAuthService, cfg AdminCookieConfig, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cfg.Name)
		if auth == nil {
			writeError(w, r, logger, errors.New("admin auth service is not configured"))
			return
		}
		if err == nil {
			if err := auth.Logout(r.Context(), cookie.Value); err != nil {
				writeError(w, r, logger, err)
				return
			}
		}

		http.SetCookie(w, newAdminCookie(cfg, "", true))
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleAdminMe(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		writeJSON(w, http.StatusOK, adminAuthEnvelope{Data: adminAuthData{User: newAdminUserData(user)}})
	}
}

// csrfProtection valida Origin/Referer em métodos mutáveis. Requisições sem
// esses cabeçalhos (clientes não navegador) dependem do SameSite=Lax do cookie.
func csrfProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		candidate := r.Header.Get("Origin")
		if candidate == "" {
			candidate = r.Header.Get("Referer")
		}
		if candidate == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !isSameOrigin(candidate, r) {
			writeError(w, r, nil, domain.ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAdmin autentica a sessão e, quando roles é informado, exige um dos papéis.
func requireAdmin(auth AdminAuthService, cfg AdminCookieConfig, logger *slog.Logger, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if auth == nil {
				writeError(w, r, logger, errors.New("admin auth service is not configured"))
				return
			}
			cookie, err := r.Cookie(cfg.Name)
			if err != nil || cookie.Value == "" {
				writeError(w, r, logger, domain.ErrUnauthorized)
				return
			}

			user, err := auth.Authenticate(r.Context(), cookie.Value)
			if err != nil {
				if !errors.Is(err, domain.ErrUnauthorized) {
					logger.Warn("admin session authentication failed", "request_id", middleware.GetReqID(r.Context()))
				}
				writeError(w, r, logger, domain.ErrUnauthorized)
				return
			}
			if len(roles) > 0 && !hasRole(user.Role, roles) {
				writeError(w, r, logger, domain.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(withAdminPrincipal(r.Context(), user)))
		})
	}
}

func withAdminPrincipal(ctx context.Context, user domain.AdminUser) context.Context {
	return context.WithValue(ctx, adminPrincipalKey{}, user)
}

func adminFromContext(ctx context.Context) (domain.AdminUser, bool) {
	user, ok := ctx.Value(adminPrincipalKey{}).(domain.AdminUser)
	return user, ok
}

func hasRole(role string, allowed []string) bool {
	for _, candidate := range allowed {
		if role == candidate {
			return true
		}
	}
	return false
}

func isSameOrigin(candidate string, r *http.Request) bool {
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	// O Host vem do navegador e não pode ser forjado por script cross-origin.
	// X-Forwarded-Proto é aceito apenas para o esquema, pois o proxy pode
	// terminar TLS; mesmo assim o host ainda precisa conferir.
	scheme := firstHeaderValue(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = "http"
		if r.TLS != nil {
			scheme = "https"
		}
	}

	return strings.EqualFold(parsed.Scheme, scheme) && strings.EqualFold(parsed.Host, r.Host)
}

func firstHeaderValue(value string) string {
	if index := strings.IndexByte(value, ','); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(value)
}

func clientIP(r *http.Request) string {
	if ip := middleware.GetClientIP(r.Context()); ip != "" {
		return ip
	}
	host, _, found := strings.Cut(r.RemoteAddr, ":")
	if !found {
		return r.RemoteAddr
	}
	return host
}

func newAdminCookie(cfg AdminCookieConfig, value string, clear bool) *http.Cookie {
	cookie := &http.Cookie{
		Name:     cfg.Name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	}
	if clear {
		cookie.MaxAge = -1
		cookie.Expires = time.Unix(1, 0).UTC()
		cookie.Value = ""
	}
	return cookie
}

func decodeAdminLoginRequest(w http.ResponseWriter, r *http.Request) (adminLoginRequest, error) {
	var request adminLoginRequest
	body := http.MaxBytesReader(w, r.Body, maxAdminAuthBodyBytes)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return adminLoginRequest{}, &malformedRequestError{
			field:   "body",
			code:    domain.CodeInvalid,
			message: "invalid JSON body",
		}
	}
	return request, nil
}

func newAdminUserData(user domain.AdminUser) adminUserData {
	return adminUserData{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}
}
