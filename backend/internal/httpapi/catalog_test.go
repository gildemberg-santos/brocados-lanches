package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/service"
)

type fakeCatalogService struct {
	menu     service.Menu
	menuErr  error
	zones    []domain.DeliveryZone
	zonesErr error
}

func (f fakeCatalogService) Menu(context.Context) (service.Menu, error) {
	return f.menu, f.menuErr
}

func (f fakeCatalogService) DeliveryZones(context.Context) ([]domain.DeliveryZone, error) {
	return f.zones, f.zonesErr
}

func TestMenu(t *testing.T) {
	updatedAt := time.Date(2026, 11, 21, 12, 0, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	menu := service.Menu{
		UpdatedAt: updatedAt,
		Categories: []service.MenuCategory{
			{
				Category: domain.Category{
					ID:        "cat-1",
					Name:      "Hambúrgueres",
					Slug:      "hamburgueres",
					SortOrder: 1,
				},
				Products: []service.MenuProduct{
					{
						Product: domain.Product{
							ID:          "prod-1",
							Name:        "Bacon Cheddar",
							Slug:        "bacon-cheddar",
							Description: "Pão brioche, carne e bacon.",
							PriceCents:  2200,
							ImageURL:    "https://cdn.example.com/burger.webp",
							ImageAlt:    "Hambúrguer com bacon",
							Allergens:   []string{"glúten", "leite"},
							Available:   true,
							Featured:    true,
							SortOrder:   1,
						},
						OptionGroups: []domain.OptionGroup{
							{
								ID:            "doneness",
								Name:          "Ponto da carne",
								Required:      true,
								MinSelections: 1,
								MaxSelections: 1,
								Options: []domain.Option{
									{ID: "medium", Name: "Ao ponto", PriceCents: 0, Available: true},
								},
							},
						},
						Addons: []domain.Addon{
							{ID: "addon-1", Name: "Bacon", PriceCents: 500, Active: true},
						},
					},
				},
			},
		},
	}

	response := performRequest(t, newCatalogTestRouter(fakeCatalogService{menu: menu}, nil), "/api/v1/menu")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}

	raw := response.Body.String()
	if strings.Contains(raw, `"_id"`) {
		t.Error("response exposes _id")
	}
	if strings.Contains(raw, `"active"`) {
		t.Error("response exposes active")
	}

	var body menuEnvelope
	if err := json.NewDecoder(strings.NewReader(raw)).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.UpdatedAt == nil {
		t.Fatal("updatedAt = nil, want RFC 3339 string")
	}
	if *body.Data.UpdatedAt != "2026-11-21T15:00:00Z" {
		t.Errorf("updatedAt = %q, want 2026-11-21T15:00:00Z", *body.Data.UpdatedAt)
	}
	if len(body.Data.Categories) != 1 {
		t.Fatalf("len(categories) = %d, want 1", len(body.Data.Categories))
	}

	category := body.Data.Categories[0]
	if category.ID != "cat-1" || category.Slug != "hamburgueres" || category.SortOrder != 1 {
		t.Errorf("category = %+v", category)
	}
	if len(category.Products) != 1 {
		t.Fatalf("len(products) = %d, want 1", len(category.Products))
	}

	product := category.Products[0]
	if product.ID != "prod-1" || product.PriceCents != 2200 || !product.Available || !product.Featured {
		t.Errorf("product = %+v", product)
	}
	if len(product.Allergens) != 2 || product.Allergens[0] != "glúten" {
		t.Errorf("allergens = %#v", product.Allergens)
	}
	if len(product.OptionGroups) != 1 {
		t.Fatalf("len(optionGroups) = %d, want 1", len(product.OptionGroups))
	}
	group := product.OptionGroups[0]
	if group.ID != "doneness" || !group.Required || group.MinSelections != 1 || group.MaxSelections != 1 {
		t.Errorf("option group = %+v", group)
	}
	if len(group.Options) != 1 || group.Options[0].ID != "medium" {
		t.Errorf("options = %+v", group.Options)
	}
	if len(product.Addons) != 1 || product.Addons[0].ID != "addon-1" || !product.Addons[0].Available {
		t.Errorf("addons = %+v", product.Addons)
	}
}

func TestMenuEmpty(t *testing.T) {
	response := performRequest(t, newCatalogTestRouter(fakeCatalogService{
		menu: service.Menu{Categories: []service.MenuCategory{}},
	}, nil), "/api/v1/menu")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}

	raw := strings.TrimSpace(response.Body.String())
	if raw != `{"data":{"categories":[]}}` {
		t.Errorf("body = %s, want {\"data\":{\"categories\":[]}}", raw)
	}
}

func TestMenuServiceError(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	response := performRequest(t, newCatalogTestRouter(fakeCatalogService{
		menuErr: errors.New("mongo down"),
	}, logger), "/api/v1/menu")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	if strings.Contains(response.Body.String(), "mongo down") {
		t.Error("response exposes internal error")
	}

	var body errorEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Code != codeInternalError {
		t.Errorf("error code = %q, want %q", body.Error.Code, codeInternalError)
	}
	if !strings.Contains(logs.String(), "mongo down") {
		t.Error("internal error was not logged")
	}
}

func TestMenuNilService(t *testing.T) {
	response := performRequest(t, newCatalogTestRouter(nil, nil), "/api/v1/menu")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
}

func TestDeliveryZones(t *testing.T) {
	zones := []domain.DeliveryZone{
		{ID: "zone-1", Name: "Tapera", FeeCents: 800, SortOrder: 1, Active: true},
		{ID: "zone-2", Name: "Centro", FeeCents: 1000, SortOrder: 2, Active: true},
	}

	response := performRequest(t, newCatalogTestRouter(fakeCatalogService{zones: zones}, nil), "/api/v1/delivery-zones")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}

	var body deliveryZonesEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.Data) != 2 {
		t.Fatalf("len(data) = %d, want 2", len(body.Data))
	}
	if body.Data[0].ID != "zone-1" || body.Data[0].FeeCents != 800 || body.Data[0].SortOrder != 1 {
		t.Errorf("first zone = %+v", body.Data[0])
	}
	if body.Data[1].ID != "zone-2" || body.Data[1].FeeCents != 1000 || body.Data[1].SortOrder != 2 {
		t.Errorf("second zone = %+v", body.Data[1])
	}
}

func TestDeliveryZonesEmpty(t *testing.T) {
	response := performRequest(t, newCatalogTestRouter(fakeCatalogService{
		zones: []domain.DeliveryZone{},
	}, nil), "/api/v1/delivery-zones")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}

	if raw := strings.TrimSpace(response.Body.String()); raw != `{"data":[]}` {
		t.Errorf("body = %s, want {\"data\":[]}", raw)
	}
}

func TestDeliveryZonesServiceError(t *testing.T) {
	response := performRequest(t, newCatalogTestRouter(fakeCatalogService{
		zonesErr: errors.New("mongo down"),
	}, nil), "/api/v1/delivery-zones")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantDetails int
		wantLogged  bool
	}{
		{
			name: "validação com detalhe",
			err: &domain.ValidationError{
				Field:   "items[0].quantity",
				Code:    domain.CodeOutOfRange,
				Message: "must be greater than zero",
			},
			wantStatus:  http.StatusBadRequest,
			wantCode:    codeValidationError,
			wantDetails: 1,
		},
		{
			name:       "validação sem detalhe",
			err:        domain.ErrValidation,
			wantStatus: http.StatusBadRequest,
			wantCode:   codeValidationError,
		},
		{
			name:       "não encontrado",
			err:        fmt.Errorf("find: %w", domain.ErrNotFound),
			wantStatus: http.StatusNotFound,
			wantCode:   codeNotFound,
		},
		{
			name:       "conflito",
			err:        fmt.Errorf("save: %w", domain.ErrConflict),
			wantStatus: http.StatusConflict,
			wantCode:   codeConflict,
		},
		{
			name:       "erro interno",
			err:        errors.New("mongo down"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   codeInternalError,
			wantLogged: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logs, nil))
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/menu", nil)

			writeError(response, request, logger, test.err)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if got := response.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}

			var body errorEnvelope
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Error.Code != test.wantCode {
				t.Errorf("error code = %q, want %q", body.Error.Code, test.wantCode)
			}
			if len(body.Error.Details) != test.wantDetails {
				t.Errorf("len(details) = %d, want %d", len(body.Error.Details), test.wantDetails)
			}
			if test.wantDetails == 1 {
				detail := body.Error.Details[0]
				if detail.Field != "items[0].quantity" || detail.Code != domain.CodeOutOfRange || detail.Message != "must be greater than zero" {
					t.Errorf("detail = %+v", detail)
				}
			}
			if test.wantStatus >= 500 && body.Error.Message != "Não foi possível processar a solicitação." {
				t.Errorf("internal error message = %q", body.Error.Message)
			}
			if strings.Contains(body.Error.Message, test.err.Error()) && test.wantStatus >= 500 {
				t.Error("response message exposes internal error")
			}
			if test.wantLogged && !strings.Contains(logs.String(), test.err.Error()) {
				t.Error("internal error was not logged")
			}
			if !test.wantLogged && logs.Len() != 0 {
				t.Errorf("unexpected log: %s", logs.String())
			}
		})
	}
}

func newCatalogTestRouter(catalog CatalogService, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return NewRouter(Dependencies{
		Pinger:  fakePinger{},
		Catalog: catalog,
		Logger:  logger,
	})
}

func performRequest(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	handler.ServeHTTP(response, request)
	return response
}
