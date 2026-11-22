package httpapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/ratelimit"
)

type fakeAdminAuthService struct {
	user       domain.AdminUser
	loginErr   error
	authErr    error
	logoutErr  error
	email      string
	password   string
	token      string
	loginCalls int
	logoutCall int
}

func (f *fakeAdminAuthService) Login(_ context.Context, email, password string) (domain.AdminUser, string, error) {
	f.loginCalls++
	f.email = email
	f.password = password
	if f.loginErr != nil {
		return domain.AdminUser{}, "", f.loginErr
	}
	return f.user, "token-opaco", nil
}

func (f *fakeAdminAuthService) Authenticate(_ context.Context, rawToken string) (domain.AdminUser, error) {
	f.token = rawToken
	if f.authErr != nil {
		return domain.AdminUser{}, f.authErr
	}
	return f.user, nil
}

func (f *fakeAdminAuthService) Logout(_ context.Context, rawToken string) error {
	f.logoutCall++
	f.token = rawToken
	return f.logoutErr
}

func newAdminAuthTestRouter(auth AdminAuthService, cfg AdminCookieConfig, limiter *ratelimit.Limiter) http.Handler {
	return NewRouter(Dependencies{
		AdminAuth:    auth,
		AdminCookie:  cfg,
		LoginLimiter: limiter,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

func newAdminLoginRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://example.com")
	request.Host = "example.com"
	return request
}

func TestAdminLoginSuccess(t *testing.T) {
	auth := &fakeAdminAuthService{user: domain.AdminUser{
		ID:    "665f1c2a9b1e4a0012a3b4f1",
		Name:  "Operador",
		Email: "operador@example.com",
		Role:  domain.AdminRoleOperator,
	}}
	router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin"}, nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, newAdminLoginRequest(t, `{"email":"operador@example.com","password":"senha-forte-com-12-caracteres"}`))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", recorder.Code, recorder.Body.String())
	}
	if auth.email != "operador@example.com" || auth.password != "senha-forte-com-12-caracteres" {
		t.Errorf("service received email = %q, password = %q", auth.email, auth.password)
	}
	if strings.Contains(recorder.Body.String(), "passwordHash") {
		t.Error("response leaked passwordHash")
	}
	var envelope adminAuthEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.User.Role != domain.AdminRoleOperator || envelope.Data.User.ID != auth.user.ID {
		t.Errorf("user data = %+v", envelope.Data.User)
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != "brocados_admin" || cookie.Value != "token-opaco" {
		t.Errorf("cookie = %+v", cookie)
	}
	if !cookie.HttpOnly || cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" {
		t.Errorf("cookie attributes = %+v", cookie)
	}
}

func TestAdminLoginSetsSecureHostCookie(t *testing.T) {
	auth := &fakeAdminAuthService{user: domain.AdminUser{Role: domain.AdminRoleOwner}}
	cfg := AdminCookieConfig{Name: "__Host-brocados_admin", Secure: true}
	router := newAdminAuthTestRouter(auth, cfg, nil)

	recorder := httptest.NewRecorder()
	request := newAdminLoginRequest(t, `{"email":"dono@example.com","password":"senha-forte-com-12-caracteres"}`)
	request.Header.Set("Origin", "https://example.com")
	request.TLS = &tls.ConnectionState{}
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", recorder.Code, recorder.Body.String())
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].Secure || cookies[0].Name != "__Host-brocados_admin" {
		t.Errorf("secure cookie = %+v", cookies)
	}
}

func TestAdminLoginRejectsBadInputAndCredentials(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		loginErr   error
		wantStatus int
		wantCode   string
	}{
		{name: "invalid JSON", body: `{`, wantStatus: http.StatusBadRequest, wantCode: codeValidationError},
		{name: "unknown field", body: `{"email":"a@b.com","password":"123456789012","role":"owner"}`, wantStatus: http.StatusBadRequest, wantCode: codeValidationError},
		{name: "bad credentials", body: `{"email":"a@b.com","password":"errada"}`, loginErr: domain.ErrUnauthorized, wantStatus: http.StatusUnauthorized, wantCode: codeUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			auth := &fakeAdminAuthService{loginErr: test.loginErr}
			router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin"}, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, newAdminLoginRequest(t, test.body))

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			var envelope errorEnvelope
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode error body: %v", err)
			}
			if envelope.Error.Code != test.wantCode {
				t.Errorf("code = %q, want %q", envelope.Error.Code, test.wantCode)
			}
			if len(recorder.Result().Cookies()) != 0 {
				t.Error("failed login must not set a session cookie")
			}
		})
	}
}

func TestAdminLoginRateLimit(t *testing.T) {
	limiter, err := ratelimit.New(2, time.Minute)
	if err != nil {
		t.Fatalf("ratelimit.New() error = %v", err)
	}
	auth := &fakeAdminAuthService{loginErr: domain.ErrUnauthorized}
	router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin"}, limiter)

	for attempt := 1; attempt <= 2; attempt++ {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, newAdminLoginRequest(t, `{"email":"a@b.com","password":"errada"}`))
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want 401", attempt, recorder.Code)
		}
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, newAdminLoginRequest(t, `{"email":"a@b.com","password":"errada"}`))
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429; body = %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Retry-After") != "60" {
		t.Error("missing Retry-After header")
	}
	if auth.loginCalls != 2 {
		t.Errorf("login calls = %d, want 2", auth.loginCalls)
	}
}

func TestAdminRoutesRejectCrossOriginMutation(t *testing.T) {
	auth := &fakeAdminAuthService{}
	router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin"}, nil)

	request := newAdminLoginRequest(t, `{"email":"a@b.com","password":"senha-forte-com-12-caracteres"}`)
	request.Header.Set("Origin", "https://evil.example")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", recorder.Code, recorder.Body.String())
	}
	if auth.loginCalls != 0 {
		t.Error("service must not be called on cross-origin request")
	}
}

func TestAdminLoginAcceptsSameOriginReferer(t *testing.T) {
	auth := &fakeAdminAuthService{user: domain.AdminUser{Role: domain.AdminRoleOwner}}
	router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin"}, nil)

	request := newAdminLoginRequest(t, `{"email":"dono@example.com","password":"senha-forte-com-12-caracteres"}`)
	request.Header.Del("Origin")
	request.Header.Set("Referer", "http://example.com/admin/login")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestAdminMeRequiresSession(t *testing.T) {
	auth := &fakeAdminAuthService{user: domain.AdminUser{
		ID:    "665f1c2a9b1e4a0012a3b4f1",
		Name:  "Dono",
		Email: "dono@example.com",
		Role:  domain.AdminRoleOwner,
	}}
	router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin"}, nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d, want 401", recorder.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: "brocados_admin", Value: "token-opaco"})
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("authenticated status = %d, want 200; body = %s", recorder.Code, recorder.Body.String())
	}
	if auth.token != "token-opaco" {
		t.Errorf(" Authenticate token = %q", auth.token)
	}
	var envelope adminAuthEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.User.Name != "Dono" || envelope.Data.User.Role != domain.AdminRoleOwner {
		t.Errorf("user data = %+v", envelope.Data.User)
	}
	if strings.Contains(recorder.Body.String(), "passwordHash") {
		t.Error("response leaked passwordHash")
	}
}

func TestAdminMeRejectsExpiredSession(t *testing.T) {
	auth := &fakeAdminAuthService{authErr: domain.ErrUnauthorized}
	router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin"}, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: "brocados_admin", Value: "token-expirado"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401; body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestAdminLogoutClearsCookie(t *testing.T) {
	auth := &fakeAdminAuthService{}
	router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin", Secure: true}, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/logout", nil)
	request.Header.Set("Origin", "https://example.com")
	request.TLS = &tls.ConnectionState{}
	request.AddCookie(&http.Cookie{Name: "brocados_admin", Value: "token-opaco"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body = %s", recorder.Code, recorder.Body.String())
	}
	if auth.logoutCall != 1 || auth.token != "token-opaco" {
		t.Errorf("Logout calls = %d, token = %q", auth.logoutCall, auth.token)
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 || !cookies[0].Secure {
		t.Errorf("cleared cookie = %+v", cookies)
	}
}

func TestAdminLogoutWithoutCookieIsNoContent(t *testing.T) {
	auth := &fakeAdminAuthService{}
	router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin"}, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/logout", nil)
	request.Header.Set("Origin", "http://example.com")
	request.Host = "example.com"
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
	if auth.logoutCall != 0 {
		t.Error("Logout must not be called without a cookie")
	}
}

func TestRequireAdminRoles(t *testing.T) {
	auth := &fakeAdminAuthService{user: domain.AdminUser{Role: domain.AdminRoleOperator}}
	handler := requireAdmin(auth, AdminCookieConfig{Name: "brocados_admin"}, nil, domain.AdminRoleOwner)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		}),
	)
	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.AddCookie(&http.Cookie{Name: "brocados_admin", Value: "token-opaco"})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestWriteErrorMapsAdminErrors(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{domain.ErrUnauthorized, codeUnauthorized},
		{domain.ErrForbidden, codeForbidden},
		{domain.ErrRateLimited, codeRateLimited},
	}
	for _, test := range tests {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		writeError(recorder, request, nil, test.err)

		var envelope errorEnvelope
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode error body: %v", err)
		}
		if envelope.Error.Code != test.code {
			t.Errorf("code = %q, want %q", envelope.Error.Code, test.code)
		}
	}
}

func TestAdminAuthServiceFailureIsGeneric(t *testing.T) {
	auth := &fakeAdminAuthService{loginErr: errors.New("detalhe interno do banco")}
	router := newAdminAuthTestRouter(auth, AdminCookieConfig{Name: "brocados_admin"}, nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, newAdminLoginRequest(t, `{"email":"a@b.com","password":"senha-forte-com-12-caracteres"}`))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "detalhe interno") {
		t.Error("internal error detail leaked to client")
	}
}
