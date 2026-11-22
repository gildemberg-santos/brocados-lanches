package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeAdminStoreHandlerService struct {
	settings domain.StoreSettings
	patch    domain.StoreSettingsPatch
	actorID  string
	err      error
	calls    int
}

func (f *fakeAdminStoreHandlerService) Get(context.Context) (domain.StoreSettings, error) {
	f.calls++
	return f.settings, f.err
}

func (f *fakeAdminStoreHandlerService) Update(_ context.Context, actorID string, patch domain.StoreSettingsPatch) (domain.StoreSettings, error) {
	f.calls++
	f.actorID = actorID
	f.patch = patch
	if f.err != nil {
		return domain.StoreSettings{}, f.err
	}
	return patch.Apply(f.settings), nil
}

func testStoreSettings() domain.StoreSettings {
	return domain.StoreSettings{
		Name: "Brocados Lanches", IsOpen: true, DeliveryTime: "30-90 min",
		WhatsApp: "+5585992358708", Address: "Rua das Palmeiras, 123, Aquiraz-CE",
		OpeningHours: []domain.OpeningHours{{
			Day: "sunday", Intervals: []domain.OpeningInterval{{Open: "18:00", Close: "23:30"}},
		}},
		CreatedAt: time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC),
	}
}

func TestAdminGetStore(t *testing.T) {
	store := &fakeAdminStoreHandlerService{settings: testStoreSettings()}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), nil, store)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/store", ""))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data adminStoreResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	settings := envelope.Data
	if settings.Name != "Brocados Lanches" || !settings.IsOpen || settings.WhatsApp != "+5585992358708" {
		t.Errorf("settings = %+v", settings)
	}
	if len(settings.OpeningHours) != 1 || settings.OpeningHours[0].Intervals[0].Open != "18:00" {
		t.Errorf("openingHours = %+v", settings.OpeningHours)
	}
	if settings.CreatedAt != "2026-11-20T12:00:00Z" {
		t.Errorf("createdAt = %q", settings.CreatedAt)
	}
}

func TestAdminPatchStore(t *testing.T) {
	store := &fakeAdminStoreHandlerService{settings: testStoreSettings()}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), nil, store)

	body := `{"isOpen":false,"openingHours":[{"day":"sunday","closed":false,"intervals":[{"open":"18:00","close":"23:00"}]}]}`
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodPatch, "/api/v1/admin/store", body))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if store.actorID != testActorIDForHTTP || store.patch.IsOpen == nil || *store.patch.IsOpen ||
		store.patch.OpeningHours == nil || len(*store.patch.OpeningHours) != 1 {
		t.Fatalf("patch = %+v, actor = %q", store.patch, store.actorID)
	}
	var envelope struct {
		Data adminStoreResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.IsOpen || envelope.Data.OpeningHours[0].Intervals[0].Close != "23:00" {
		t.Errorf("response settings = %+v", envelope.Data)
	}
}

func TestAdminStoreRoutesRequireSessionAndCSRF(t *testing.T) {
	store := &fakeAdminStoreHandlerService{settings: testStoreSettings()}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), nil, store)

	anonymous := httptest.NewRecorder()
	request := newAdminRequest(http.MethodPatch, "/api/v1/admin/store", `{"isOpen":false}`)
	request.Header.Del("Cookie")
	router.ServeHTTP(anonymous, request)
	if anonymous.Code != http.StatusUnauthorized {
		t.Errorf("anonymous status = %d, want 401", anonymous.Code)
	}

	csrf := httptest.NewRecorder()
	csrfRequest := newAdminRequest(http.MethodPatch, "/api/v1/admin/store", `{"isOpen":false}`)
	csrfRequest.Header.Set("Origin", "https://evil.example")
	router.ServeHTTP(csrf, csrfRequest)
	if csrf.Code != http.StatusForbidden {
		t.Errorf("CSRF status = %d, want 403", csrf.Code)
	}

	badBody := httptest.NewRecorder()
	router.ServeHTTP(badBody, newAdminRequest(http.MethodPatch, "/api/v1/admin/store", `{"isOpen":false,"unknown":1}`))
	if badBody.Code != http.StatusBadRequest {
		t.Errorf("unknown field status = %d, want 400", badBody.Code)
	}
}

func TestAdminStoreServiceValidationErrorIsMapped(t *testing.T) {
	store := &fakeAdminStoreHandlerService{err: &domain.ValidationError{Field: "whatsapp", Code: domain.CodeInvalid, Message: "inválido"}}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), nil, store)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodPatch, "/api/v1/admin/store", `{"whatsapp":"85992358708"}`))
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body = %s", response.Code, response.Body.String())
	}
	var envelope errorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(envelope.Error.Details) != 1 || envelope.Error.Details[0].Field != "whatsapp" {
		t.Errorf("details = %+v", envelope.Error.Details)
	}
}
