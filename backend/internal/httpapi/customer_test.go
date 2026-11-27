package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeCustomerLookupService struct {
	profile domain.CustomerProfile
	err     error
	calls   int
	phone   string
}

func (f *fakeCustomerLookupService) Lookup(_ context.Context, phone string) (*domain.CustomerProfile, error) {
	f.calls++
	f.phone = phone
	if f.err != nil {
		return nil, f.err
	}
	if f.profile.PhoneKey == "" {
		return nil, nil
	}
	profile := f.profile
	return &profile, nil
}

func newCustomerTestRouter(customers CustomerLookupService) http.Handler {
	return NewRouter(Dependencies{
		Pinger:           fakePinger{},
		CustomerProfiles: customers,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

func performCustomerLookup(handler http.Handler, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/customers/profile", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestCustomerProfileLookupSuccess(t *testing.T) {
	customers := &fakeCustomerLookupService{profile: domain.CustomerProfile{
		PhoneKey:      "5585999999999",
		Name:          "Maria Silva",
		Phone:         "(85) 99999-9999",
		Fulfillment:   domain.FulfillmentDelivery,
		PaymentMethod: "pix",
		Address: &domain.OrderAddress{
			Street:   "Rua das Flores",
			Number:   "450",
			District: "Tapera",
			City:     "Aquiraz",
			State:    "CE",
			ZipCode:  "61700000",
		},
		Zone: &domain.OrderZone{ID: "zona-1", Name: "Centro", FeeCents: 800},
	}}
	response := performCustomerLookup(
		newCustomerTestRouter(customers),
		`{"phone":"(85) 99999-9999"}`,
	)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
	if customers.calls != 1 || customers.phone != "(85) 99999-9999" {
		t.Errorf("chamadas = %d, telefone = %q", customers.calls, customers.phone)
	}

	var body customerLookupEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.Data.Found || body.Data.Profile == nil {
		t.Fatalf("data = %+v", body.Data)
	}
	if body.Data.Profile.Name != "Maria Silva" || body.Data.Profile.Zone.ID != "zona-1" {
		t.Errorf("profile = %+v", body.Data.Profile)
	}
	if strings.Contains(response.Body.String(), "feeCents") ||
		strings.Contains(response.Body.String(), "phoneKey") ||
		strings.Contains(response.Body.String(), "orderCount") {
		t.Errorf("resposta expôs campo interno: %s", response.Body.String())
	}
}

func TestCustomerProfileLookupNotFound(t *testing.T) {
	customers := &fakeCustomerLookupService{}
	response := performCustomerLookup(newCustomerTestRouter(customers), `{"phone":"85988888888"}`)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var body customerLookupEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.Found || body.Data.Profile != nil {
		t.Errorf("data = %+v, want found=false", body.Data)
	}
}

func TestCustomerProfileLookupValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "JSON inválido", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "campo extra", body: `{"phone":"85999999999","name":"Maria"}`, wantStatus: http.StatusBadRequest},
		{name: "telefone inválido", body: `{"phone":"123"}`, wantStatus: http.StatusUnprocessableEntity},
		{name: "DDI duplicado", body: `{"phone":"+5555859913655"}`, wantStatus: http.StatusUnprocessableEntity},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			customers := &fakeCustomerLookupService{}
			response := performCustomerLookup(newCustomerTestRouter(customers), test.body)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			if customers.calls != 0 {
				t.Errorf("service calls = %d, want 0", customers.calls)
			}
		})
	}
}

func TestCustomerProfileLookupServiceError(t *testing.T) {
	customers := &fakeCustomerLookupService{err: errors.New("mongo indisponível")}
	response := performCustomerLookup(newCustomerTestRouter(customers), `{"phone":"85999999999"}`)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body = %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "mongo indisponível") {
		t.Errorf("erro interno vazou: %s", response.Body.String())
	}
}

func TestCustomerProfileLookupRequiresService(t *testing.T) {
	response := performCustomerLookup(newCustomerTestRouter(nil), `{"phone":"85999999999"}`)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body = %s", response.Code, response.Body.String())
	}
}
