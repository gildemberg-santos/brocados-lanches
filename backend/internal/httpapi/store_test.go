package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/service"
)

type fakeStoreHandlerService struct {
	view service.StoreView
	err  error
}

func (f *fakeStoreHandlerService) Get(context.Context) (service.StoreView, error) {
	return f.view, f.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestGetStoreReturnsEffectiveStatus(t *testing.T) {
	next := time.Date(2026, 11, 23, 2, 0, 0, 0, time.UTC)
	view := service.StoreView{
		Settings: domain.StoreSettings{
			Name: "Brocados Lanches", IsOpen: true, DeliveryTime: "30-90 min",
			WhatsApp: "+5585992358708", Instagram: "@brocados", Address: "Rua das Palmeiras, 123",
			OpeningHours: []domain.OpeningHours{{
				Day: "sunday", Intervals: []domain.OpeningInterval{{Open: "18:00", Close: "23:30"}},
			}},
			Pix: domain.PixSettings{
				Active: true, Key: "123e4567-e89b-12d3-a456-426614174000",
				MerchantName: "Brocados Lanches", MerchantCity: "Aquiraz",
			},
			UpdatedAt: time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC),
		},
		Status:        "closed",
		NextOpeningAt: &next,
	}
	router := NewRouter(Dependencies{Store: &fakeStoreHandlerService{view: view}, Logger: testLogger()})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/store", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var envelope storeEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data := envelope.Data
	if data.Status != "closed" || !data.IsOpenManual || data.Timezone != "America/Fortaleza" {
		t.Errorf("store data = %+v", data)
	}
	if data.NextOpeningAt == nil || *data.NextOpeningAt != "2026-11-23T02:00:00Z" {
		t.Errorf("nextOpeningAt = %v", data.NextOpeningAt)
	}
	if len(data.OpeningHours) != 1 || data.OpeningHours[0].Intervals[0].Close != "23:30" {
		t.Errorf("openingHours = %+v", data.OpeningHours)
	}
	if !data.Pix.Active || data.Pix.Key != "123e4567-e89b-12d3-a456-426614174000" ||
		data.Pix.MerchantName != "Brocados Lanches" || data.Pix.MerchantCity != "Aquiraz" {
		t.Errorf("pix = %+v", data.Pix)
	}
	if data.UpdatedAt != "2026-11-22T12:00:00Z" {
		t.Errorf("updatedAt = %q", data.UpdatedAt)
	}
}

func TestGetStoreHidesInactivePixConfiguration(t *testing.T) {
	view := service.StoreView{Settings: domain.StoreSettings{
		Pix: domain.PixSettings{
			Active: false, Key: "segredo@example.com",
			MerchantName: "Brocados", MerchantCity: "Aquiraz",
		},
	}}
	router := NewRouter(Dependencies{Store: &fakeStoreHandlerService{view: view}, Logger: testLogger()})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/store", nil))

	var envelope storeEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.Pix.Active || envelope.Data.Pix.Key != "" ||
		envelope.Data.Pix.MerchantName != "" || envelope.Data.Pix.MerchantCity != "" {
		t.Errorf("pix inativo foi exposto: %+v", envelope.Data.Pix)
	}
}

func TestGetStoreWithoutServiceReturnsInternalError(t *testing.T) {
	router := NewRouter(Dependencies{Logger: testLogger()})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/store", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
}

func TestGetStoreMapsServiceError(t *testing.T) {
	service := &fakeStoreHandlerService{err: errors.New("mongo down")}
	router := NewRouter(Dependencies{Store: service, Logger: testLogger()})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/store", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
}
