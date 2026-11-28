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

type fakeAdminPixHandlerService struct {
	settings domain.PixSettings
	err      error
	received domain.PixSettingsPatch
	actorID  string
	calls    int
}

func (f *fakeAdminPixHandlerService) Get(context.Context) (domain.PixSettings, error) {
	return f.settings, f.err
}

func (f *fakeAdminPixHandlerService) Update(_ context.Context, actorID string, patch domain.PixSettingsPatch) (domain.PixSettings, error) {
	f.calls++
	f.actorID = actorID
	f.received = patch
	updated := f.settings
	if patch.Active != nil {
		updated.Active = *patch.Active
	}
	if patch.Key != nil {
		updated.Key = *patch.Key
	}
	if patch.MerchantName != nil {
		updated.MerchantName = *patch.MerchantName
	}
	if patch.MerchantCity != nil {
		updated.MerchantCity = *patch.MerchantCity
	}
	return updated, f.err
}

func TestAdminGetPixReturnsSettingsEnvelope(t *testing.T) {
	service := &fakeAdminPixHandlerService{settings: domain.PixSettings{
		Active: true, Key: "123e4567-e89b-12d3-a456-426614174000",
		MerchantName: "Brocados Lanches", MerchantCity: "Aquiraz",
	}}
	handler := handleAdminGetPix(service, slog.Default())

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/admin/integrations/pix", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	body := response.Body.String()
	if !strings.Contains(body, `"key":"123e4567-e89b-12d3-a456-426614174000"`) {
		t.Fatalf("resposta inesperada: %s", body)
	}
	if !strings.Contains(body, `"merchantCity":"Aquiraz"`) {
		t.Fatalf("cidade ausente: %s", body)
	}
}

func TestAdminUpdatePixUsesAuthenticatedOwner(t *testing.T) {
	service := &fakeAdminPixHandlerService{settings: domain.PixSettings{}}
	router := NewRouter(Dependencies{
		AdminAuth:   newAdminOwnerAuth(),
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		Pix:         service,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	request := newAdminRequest(
		http.MethodPatch,
		"/api/v1/admin/integrations/pix",
		`{"active":true,"key":"123e4567-e89b-12d3-a456-426614174000","merchantName":"Brocados Lanches","merchantCity":"Aquiraz"}`,
	)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.actorID != testActorIDForHTTP {
		t.Fatalf("actorID = %q", service.actorID)
	}
	if service.received.Key == nil || *service.received.Key != "123e4567-e89b-12d3-a456-426614174000" {
		t.Fatalf("chave não recebida: %+v", service.received)
	}
	if service.received.MerchantCity == nil || *service.received.MerchantCity != "Aquiraz" {
		t.Fatalf("cidade não recebida: %+v", service.received)
	}
}

func TestAdminPixRouteIsOwnerOnly(t *testing.T) {
	router := NewRouter(Dependencies{
		AdminAuth:   newAdminOperatorAuth(),
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		Pix:         &fakeAdminPixHandlerService{},
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	for _, method := range []string{http.MethodGet, http.MethodPatch} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, newAdminRequest(method, "/api/v1/admin/integrations/pix", ""))
		if response.Code != http.StatusForbidden {
			t.Fatalf("%s status = %d, esperado 403", method, response.Code)
		}
	}
}
