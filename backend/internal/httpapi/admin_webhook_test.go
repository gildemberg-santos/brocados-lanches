package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeAdminWebhookHandlerService struct {
	settings domain.WebhookSettings
	err      error
	received domain.WebhookSettingsPatch
	actorID  string
	calls    int
}

func (f *fakeAdminWebhookHandlerService) Get(context.Context) (domain.WebhookSettings, error) {
	return f.settings, f.err
}

func (f *fakeAdminWebhookHandlerService) Update(_ context.Context, actorID string, patch domain.WebhookSettingsPatch) (domain.WebhookSettings, error) {
	f.calls++
	f.actorID = actorID
	f.received = patch
	updated := f.settings
	if patch.URL != nil {
		updated.URL = *patch.URL
	}
	if patch.Active != nil {
		updated.Active = *patch.Active
	}
	return updated, f.err
}

func TestAdminGetWebhookReturnsSettingsEnvelope(t *testing.T) {
	service := &fakeAdminWebhookHandlerService{settings: domain.WebhookSettings{
		URL: "https://hooks.exemplo.invalid/path", Active: true,
	}}
	handler := handleAdminGetWebhook(service, slog.Default())

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/integrations/webhook", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), `"url":"https://hooks.exemplo.invalid/path"`) {
		t.Fatalf("resposta inesperada: %s", response.Body.String())
	}
}

func TestAdminUpdateWebhookUsesAuthenticatedActor(t *testing.T) {
	service := &fakeAdminWebhookHandlerService{settings: domain.WebhookSettings{}}
	router := NewRouter(Dependencies{
		AdminAuth:   newAdminOwnerAuth(),
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		Webhooks:    service,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	request := newAdminRequest(
		http.MethodPatch,
		"/api/v1/admin/integrations/webhook",
		`{"url":"http://127.0.0.1:5678/hook","active":true}`,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.actorID != testActorIDForHTTP {
		t.Fatalf("actorID = %q", service.actorID)
	}
	if service.received.URL == nil || *service.received.URL != "http://127.0.0.1:5678/hook" {
		t.Fatalf("url não recebida: %+v", service.received)
	}
}

func TestAdminWebhookRouteIsOwnerOnly(t *testing.T) {
	router := NewRouter(Dependencies{
		AdminAuth:   newAdminOperatorAuth(),
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		Webhooks:    &fakeAdminWebhookHandlerService{},
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	for _, method := range []string{http.MethodGet, http.MethodPatch} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, newAdminRequest(method, "/api/v1/admin/integrations/webhook", ""))
		if response.Code != http.StatusForbidden {
			t.Fatalf("%s status = %d, esperado 403", method, response.Code)
		}
	}
}
