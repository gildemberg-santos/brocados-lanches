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
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/service"
)

type fakeAdminCatalogHandlerService struct {
	category      domain.Category
	product       domain.Product
	addon         domain.Addon
	zone          domain.DeliveryZone
	pageMeta      adminMeta
	actorID       string
	entityID      string
	categoryPatch domain.CategoryPatch
	query         domain.AdminListQuery
	err           error
	calls         int
}

func (f *fakeAdminCatalogHandlerService) ListCategories(_ context.Context, query domain.AdminListQuery) (service.AdminPage[domain.Category], error) {
	f.calls++
	f.query = query
	return service.AdminPage[domain.Category]{Items: []domain.Category{f.category}, Page: 2, PageSize: 10, Total: 11, TotalPages: 2}, f.err
}

func (f *fakeAdminCatalogHandlerService) CreateCategory(_ context.Context, actorID string, category domain.Category) (domain.Category, error) {
	f.calls++
	f.actorID = actorID
	f.category = category
	return f.fillCategory(category), f.err
}

func (f *fakeAdminCatalogHandlerService) UpdateCategory(_ context.Context, actorID, id string, patch domain.CategoryPatch) (domain.Category, error) {
	f.calls++
	f.actorID = actorID
	f.entityID = id
	f.categoryPatch = patch
	f.category.Name = "Hambúrgueres"
	return f.fillCategory(f.category), f.err
}

func (f *fakeAdminCatalogHandlerService) ListProducts(_ context.Context, query domain.AdminListQuery) (service.AdminPage[domain.Product], error) {
	f.calls++
	f.query = query
	return service.AdminPage[domain.Product]{Items: []domain.Product{f.product}, Page: 1, PageSize: 20, Total: 1, TotalPages: 1}, f.err
}

func (f *fakeAdminCatalogHandlerService) CreateProduct(_ context.Context, actorID string, product domain.Product) (domain.Product, error) {
	f.calls++
	f.actorID = actorID
	f.product = product
	return product, f.err
}

func (f *fakeAdminCatalogHandlerService) UpdateProduct(_ context.Context, actorID, id string, _ domain.ProductPatch) (domain.Product, error) {
	f.calls++
	f.actorID = actorID
	f.entityID = id
	return f.product, f.err
}

func (f *fakeAdminCatalogHandlerService) ListAddons(_ context.Context, query domain.AdminListQuery) (service.AdminPage[domain.Addon], error) {
	f.calls++
	f.query = query
	return service.AdminPage[domain.Addon]{Items: []domain.Addon{f.addon}, Page: 1, PageSize: 20, Total: 1, TotalPages: 1}, f.err
}

func (f *fakeAdminCatalogHandlerService) CreateAddon(_ context.Context, actorID string, addon domain.Addon) (domain.Addon, error) {
	f.calls++
	f.actorID = actorID
	f.addon = addon
	return addon, f.err
}

func (f *fakeAdminCatalogHandlerService) UpdateAddon(_ context.Context, actorID, id string, _ domain.AddonPatch) (domain.Addon, error) {
	f.calls++
	f.actorID = actorID
	f.entityID = id
	return f.addon, f.err
}

func (f *fakeAdminCatalogHandlerService) ListDeliveryZones(_ context.Context, query domain.AdminListQuery) (service.AdminPage[domain.DeliveryZone], error) {
	f.calls++
	f.query = query
	return service.AdminPage[domain.DeliveryZone]{Items: []domain.DeliveryZone{f.zone}, Page: 1, PageSize: 20, Total: 1, TotalPages: 1}, f.err
}

func (f *fakeAdminCatalogHandlerService) CreateDeliveryZone(_ context.Context, actorID string, zone domain.DeliveryZone) (domain.DeliveryZone, error) {
	f.calls++
	f.actorID = actorID
	f.zone = zone
	return zone, f.err
}

func (f *fakeAdminCatalogHandlerService) UpdateDeliveryZone(_ context.Context, actorID, id string, _ domain.DeliveryZonePatch) (domain.DeliveryZone, error) {
	f.calls++
	f.actorID = actorID
	f.entityID = id
	return f.zone, f.err
}

func (f *fakeAdminCatalogHandlerService) fillCategory(category domain.Category) domain.Category {
	category.ID = "665f1c2a9b1e4a0012a3b4f1"
	category.CreatedAt = fixedTestTime()
	category.UpdatedAt = fixedTestTime()
	return category
}

func fixedTestTime() (result time.Time) {
	return time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
}

func newAdminCatalogTestRouter(auth AdminAuthService, catalog AdminCatalogService, store AdminStoreService) http.Handler {
	return NewRouter(Dependencies{
		AdminAuth:    auth,
		AdminCookie:  AdminCookieConfig{Name: "brocados_admin"},
		AdminCatalog: catalog,
		AdminStore:   store,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

func newAdminOperatorAuth() *fakeAdminAuthService {
	return &fakeAdminAuthService{user: domain.AdminUser{
		ID: testActorIDForHTTP, Name: "Operador", Email: "operador@example.com", Role: domain.AdminRoleOperator,
	}}
}

const testActorIDForHTTP = "665f1c2a9b1e4a0012a3b4f9"

func newAdminRequest(method, path, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://example.com")
	request.Host = "example.com"
	request.AddCookie(&http.Cookie{Name: "brocados_admin", Value: "token-opaco"})
	return request
}

func TestAdminCatalogRoutesRequireSessionRoleAndCSRF(t *testing.T) {
	paths := []string{"/api/v1/admin/categories", "/api/v1/admin/products", "/api/v1/admin/addons", "/api/v1/admin/delivery-zones"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			catalog := &fakeAdminCatalogHandlerService{}
			router := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)

			anonymous := httptest.NewRecorder()
			request := newAdminRequest(http.MethodGet, path, "")
			request.Header.Del("Cookie")
			router.ServeHTTP(anonymous, request)
			if anonymous.Code != http.StatusUnauthorized {
				t.Errorf("anonymous GET status = %d, want 401", anonymous.Code)
			}

			forbiddenAuth := &fakeAdminAuthService{user: domain.AdminUser{ID: "outro", Role: "financeiro"}}
			forbidden := httptest.NewRecorder()
			forbiddenRouter := newAdminCatalogTestRouter(forbiddenAuth, catalog, nil)
			forbiddenRouter.ServeHTTP(forbidden, newAdminRequest(http.MethodGet, path, ""))
			if forbidden.Code != http.StatusForbidden {
				t.Errorf("wrong role GET status = %d, want 403", forbidden.Code)
			}

			csrfRouter := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)
			csrf := httptest.NewRecorder()
			csrfRequest := newAdminRequest(http.MethodPost, path, `{"name":"x","slug":"x"}`)
			csrfRequest.Header.Set("Origin", "http://evil.example")
			csrfRouter.ServeHTTP(csrf, csrfRequest)
			if csrf.Code != http.StatusForbidden {
				t.Errorf("CSRF POST status = %d, want 403", csrf.Code)
			}
			if catalog.calls != 0 {
				t.Errorf("service calls = %d, want 0 for unauthorized requests", catalog.calls)
			}
		})
	}
}

func TestAdminListCategoriesEnvelopeAndQuery(t *testing.T) {
	catalog := &fakeAdminCatalogHandlerService{category: domain.Category{ID: "665f1c2a9b1e4a0012a3b4f1", Name: "Hambúrgueres"}}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/categories?page=2&pageSize=10&active=true&sort=name&order=desc&q=hamb", ""))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data []adminCategoryResponse `json:"data"`
		Meta adminMeta               `json:"meta"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(envelope.Data) != 1 || envelope.Data[0].ID != "665f1c2a9b1e4a0012a3b4f1" {
		t.Errorf("data = %+v", envelope.Data)
	}
	if envelope.Meta.Total != 11 || envelope.Meta.TotalPages != 2 || envelope.Meta.Page != 2 {
		t.Errorf("meta = %+v", envelope.Meta)
	}
	if catalog.query.Page != 2 || catalog.query.PageSize != 10 || catalog.query.Sort != "name" ||
		catalog.query.Order != "desc" || catalog.query.Search != "hamb" || catalog.query.Active == nil || !*catalog.query.Active {
		t.Errorf("query = %+v", catalog.query)
	}
}

func TestAdminCreateAndUpdateCategory(t *testing.T) {
	catalog := &fakeAdminCatalogHandlerService{}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)

	created := httptest.NewRecorder()
	router.ServeHTTP(created, newAdminRequest(http.MethodPost, "/api/v1/admin/categories",
		`{"name":"Hambúrgueres","slug":"hamburgueres","active":true}`))
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body = %s", created.Code, created.Body.String())
	}
	if catalog.actorID != testActorIDForHTTP || !catalog.category.Active || catalog.category.Slug != "hamburgueres" {
		t.Errorf("service received actor = %q, category = %+v", catalog.actorID, catalog.category)
	}

	updated := httptest.NewRecorder()
	router.ServeHTTP(updated, newAdminRequest(http.MethodPatch, "/api/v1/admin/categories/665f1c2a9b1e4a0012a3b4f1", `{"active":false}`))
	if updated.Code != http.StatusOK {
		t.Fatalf("patch status = %d, want 200; body = %s", updated.Code, updated.Body.String())
	}
	if catalog.entityID != "665f1c2a9b1e4a0012a3b4f1" || catalog.categoryPatch.Active == nil || *catalog.categoryPatch.Active {
		t.Errorf("patch = %+v, id = %q", catalog.categoryPatch, catalog.entityID)
	}
}

func TestAdminCreateRejectsMalformedBodyAndServiceErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
		wantCode   string
	}{
		{name: "JSON inválido", body: `{`, wantStatus: http.StatusBadRequest, wantCode: codeValidationError},
		{name: "campo desconhecido", body: `{"name":"x","slug":"x","preco":10}`, wantStatus: http.StatusBadRequest, wantCode: codeValidationError},
		{name: "regra de negócio", body: `{"name":"x","slug":"x"}`, serviceErr: &domain.ValidationError{Field: "slug", Code: domain.CodeDuplicate, Message: "duplicado"}, wantStatus: http.StatusUnprocessableEntity, wantCode: codeValidationError},
		{name: "slug duplicado", body: `{"name":"x","slug":"x"}`, serviceErr: domain.ErrConflict, wantStatus: http.StatusConflict, wantCode: codeConflict},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog := &fakeAdminCatalogHandlerService{err: test.serviceErr}
			router := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, newAdminRequest(http.MethodPost, "/api/v1/admin/categories", test.body))

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			var envelope errorEnvelope
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if envelope.Error.Code != test.wantCode {
				t.Errorf("code = %q, want %q", envelope.Error.Code, test.wantCode)
			}
		})
	}
}

func TestAdminProductRequestMapsNestedFields(t *testing.T) {
	catalog := &fakeAdminCatalogHandlerService{}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)

	body := `{
		"categoryId":"665f1c2a9b1e4a0012a3b4f1","name":"Hambúrguer","slug":"hamburguer",
		"description":"Descrição","priceCents":2300,"available":true,"active":true,
		"allergens":["glúten"],"addonIds":["665f1c2a9b1e4a0012a3b4f2"],
		"optionGroups":[{"id":"ponto","name":"Ponto","required":true,"minSelections":1,"maxSelections":1,
		"options":[{"id":"mal passado","name":"Mal passado","priceCents":0,"available":true}]}]
	}`
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodPost, "/api/v1/admin/products", body))
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	product := catalog.product
	if product.PriceCents != 2300 || !product.Available || len(product.Allergens) != 1 ||
		len(product.AddonIDs) != 1 || len(product.OptionGroups) != 1 || len(product.OptionGroups[0].Options) != 1 {
		t.Fatalf("product = %+v", product)
	}
	if product.OptionGroups[0].Options[0].ID != "mal passado" {
		t.Errorf("option = %+v", product.OptionGroups[0].Options[0])
	}

	list := httptest.NewRecorder()
	router.ServeHTTP(list, newAdminRequest(http.MethodGet, "/api/v1/admin/products?categoryId=665f1c2a9b1e4a0012a3b4f1", ""))
	if list.Code != http.StatusOK || catalog.query.CategoryID != "665f1c2a9b1e4a0012a3b4f1" {
		t.Errorf("list status = %d, query = %+v", list.Code, catalog.query)
	}
}

func TestAdminProductResponseUsesEmptyArrays(t *testing.T) {
	tests := []struct {
		name      string
		allergens []string
		addonIDs  []string
	}{
		{name: "listas nulas"},
		{name: "listas vazias não nulas", allergens: []string{}, addonIDs: []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			catalog := &fakeAdminCatalogHandlerService{product: domain.Product{
				ID:           "665f1c2a9b1e4a0012a3b4f1",
				Name:         "Hambúrguer",
				Slug:         "hamburguer",
				PriceCents:   2300,
				Available:    true,
				Active:       true,
				Allergens:    test.allergens,
				AddonIDs:     test.addonIDs,
				OptionGroups: []domain.OptionGroup{{ID: "ponto", Name: "Ponto"}},
			}}
			router := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/products", ""))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
			}

			var payload struct {
				Data []map[string]json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(payload.Data) != 1 {
				t.Fatalf("data length = %d, want 1", len(payload.Data))
			}
			for _, field := range []string{"allergens", "addonIds"} {
				if got := string(payload.Data[0][field]); got != "[]" {
					t.Errorf("%s = %s, want []", field, got)
				}
			}
			if got := string(payload.Data[0]["optionGroups"]); got == "null" || !strings.HasPrefix(got, "[") {
				t.Errorf("optionGroups = %s, want an array", got)
			}

			var groups []map[string]json.RawMessage
			if err := json.Unmarshal(payload.Data[0]["optionGroups"], &groups); err != nil {
				t.Fatalf("decode option groups: %v", err)
			}
			if len(groups) != 1 || string(groups[0]["options"]) != "[]" {
				t.Errorf("option groups = %s, want one group with options []", payload.Data[0]["optionGroups"])
			}
		})
	}
}

func TestAdminCreateAddonAndZone(t *testing.T) {
	catalog := &fakeAdminCatalogHandlerService{}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)

	addon := httptest.NewRecorder()
	router.ServeHTTP(addon, newAdminRequest(http.MethodPost, "/api/v1/admin/addons",
		`{"name":"Queijo","slug":"queijo","priceCents":300,"sortOrder":1}`))
	if addon.Code != http.StatusCreated || catalog.addon.PriceCents != 300 || !catalog.addon.Active {
		t.Errorf("addon status = %d, addon = %+v", addon.Code, catalog.addon)
	}

	zone := httptest.NewRecorder()
	router.ServeHTTP(zone, newAdminRequest(http.MethodPost, "/api/v1/admin/delivery-zones",
		`{"name":"Aquiraz","slug":"aquiraz","feeCents":500,"active":false}`))
	if zone.Code != http.StatusCreated || catalog.zone.FeeCents != 500 || catalog.zone.Active {
		t.Errorf("zone status = %d, zone = %+v", zone.Code, catalog.zone)
	}
}

func TestAdminListQueryRejectsInvalidValues(t *testing.T) {
	catalog := &fakeAdminCatalogHandlerService{}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)
	for _, path := range []string{
		"/api/v1/admin/categories?page=abc",
		"/api/v1/admin/categories?pageSize=abc",
		"/api/v1/admin/categories?active=maybe",
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, newAdminRequest(http.MethodGet, path, ""))
		if response.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s status = %d, want 422", path, response.Code)
		}
	}
}

func TestAdminCatalogHandlerErrorsDoNotLeakInternals(t *testing.T) {
	catalog := &fakeAdminCatalogHandlerService{err: errors.New("mongo: connection refused")}
	router := newAdminCatalogTestRouter(newAdminOperatorAuth(), catalog, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/categories", ""))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
	if strings.Contains(response.Body.String(), "connection refused") {
		t.Errorf("response leaked internal error: %s", response.Body.String())
	}
}
