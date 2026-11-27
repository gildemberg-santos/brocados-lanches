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

type fakeAdminOrdersHandlerService struct {
	item         domain.AdminOrderListItem
	order        domain.Order
	profile      *domain.CustomerProfile
	profileErr   error
	query        domain.AdminOrderListQuery
	actorID      string
	id           string
	profilePhone string
	request      domain.OrderStatusChangeRequest
	err          error
	calls        int
	profileCalls int
}

func (f *fakeAdminOrdersHandlerService) List(_ context.Context, query domain.AdminOrderListQuery) (service.AdminPage[domain.AdminOrderListItem], error) {
	f.calls++
	f.query = query
	return service.AdminPage[domain.AdminOrderListItem]{
		Items: []domain.AdminOrderListItem{f.item}, Page: query.Page, PageSize: query.PageSize, Total: 1, TotalPages: 1,
	}, f.err
}

func (f *fakeAdminOrdersHandlerService) Get(_ context.Context, id string) (domain.Order, error) {
	f.calls++
	f.id = id
	return f.order, f.err
}

func (f *fakeAdminOrdersHandlerService) GetCustomerProfile(_ context.Context, phone string) (*domain.CustomerProfile, error) {
	f.profileCalls++
	f.profilePhone = phone
	return f.profile, f.profileErr
}

func (f *fakeAdminOrdersHandlerService) UpdateStatus(_ context.Context, actorID, id string, request domain.OrderStatusChangeRequest) (domain.Order, error) {
	f.calls++
	f.actorID = actorID
	f.id = id
	f.request = request
	updated := f.order
	updated.Status = request.Status
	return updated, f.err
}

func (f *fakeAdminOrdersHandlerService) Delete(_ context.Context, actorID, id string) (domain.Order, error) {
	f.calls++
	f.actorID = actorID
	f.id = id
	if f.err != nil {
		return domain.Order{}, f.err
	}
	return f.order, nil
}

func newAdminOrdersTestRouter(auth AdminAuthService, orders AdminOrderService) http.Handler {
	return NewRouter(Dependencies{
		AdminAuth:   auth,
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		AdminOrders: orders,
		Notifier:    &fakeOrderNotifier{},
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

func adminOrderDetailFixture() domain.Order {
	created := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	return domain.Order{
		ID: "665f1c2a9b1e4a0012a3b4f1", Number: 42, PublicID: "opaco",
		Status: domain.OrderStatusPending, Fulfillment: domain.FulfillmentDelivery,
		Customer: domain.OrderCustomer{Name: "Maria Silva", Phone: "+5585992358708"},
		Address: &domain.OrderAddress{
			Street: "Rua das Flores", Number: "123", District: "Centro",
			City: "Aquiraz", State: "CE", ZipCode: "61700000",
		},
		Zone: &domain.OrderZone{ID: "665f1c2a9b1e4a0012a3b4f2", Name: "Centro", FeeCents: 500},
		Items: []domain.OrderItem{{
			ProductID: "665f1c2a9b1e4a0012a3b4f3", Name: "Burguer",
			UnitPriceCents: 2300, Quantity: 1, LineTotalCents: 2300,
			Options: []domain.OrderItemOption{{OptionID: "medium", Name: "Ao ponto", Kind: "option", PriceCents: 0}},
		}},
		SubtotalCents: 2300, DeliveryFeeCents: 500, TotalCents: 2800,
		Payment:       domain.OrderPayment{Method: domain.PaymentMethodCash},
		StatusHistory: []domain.OrderStatusChange{{Status: domain.OrderStatusPending, At: created}},
		CreatedAt:     created, UpdatedAt: created,
	}
}

func TestAdminListOrdersParsesFiltersAndReturnsMaskedList(t *testing.T) {
	orders := &fakeAdminOrdersHandlerService{item: domain.NewAdminOrderListItem(adminOrderDetailFixture())}
	router := newAdminOrdersTestRouter(newAdminOperatorAuth(), orders)

	path := "/api/v1/admin/orders?status=pending&fulfillment=delivery&from=2026-11-22T00:00:00Z&to=2026-11-23T00:00:00Z&q=8708&sort=number&order=asc&page=2&pageSize=10"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, path, ""))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if orders.query.Status != domain.OrderStatusPending || orders.query.Fulfillment != domain.FulfillmentDelivery ||
		orders.query.Search != "8708" || orders.query.Sort != "number" || orders.query.Order != "asc" ||
		orders.query.Page != 2 || orders.query.PageSize != 10 || orders.query.From == nil || orders.query.To == nil {
		t.Fatalf("query = %+v", orders.query)
	}

	var envelope struct {
		Data []adminOrderListItemResponse `json:"data"`
		Meta adminMeta                    `json:"meta"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(envelope.Data) != 1 || envelope.Data[0].CustomerName != "M*** S***" || envelope.Data[0].Phone != "****8708" {
		t.Fatalf("data = %+v", envelope.Data)
	}
	if envelope.Meta.Total != 1 || envelope.Meta.TotalPages != 1 || envelope.Meta.PageSize != 10 {
		t.Errorf("meta = %+v", envelope.Meta)
	}
}

func TestAdminListOrdersRejectsInvalidTimestamp(t *testing.T) {
	orders := &fakeAdminOrdersHandlerService{}
	router := newAdminOrdersTestRouter(newAdminOperatorAuth(), orders)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/orders?from=ontem", ""))
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body = %s", response.Code, response.Body.String())
	}
	var envelope errorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(envelope.Error.Details) != 1 || envelope.Error.Details[0].Field != "from" {
		t.Errorf("details = %+v", envelope.Error.Details)
	}
}

func TestAdminGetOrderReturnsFullDetail(t *testing.T) {
	orders := &fakeAdminOrdersHandlerService{order: adminOrderDetailFixture()}
	router := newAdminOrdersTestRouter(newAdminOperatorAuth(), orders)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/orders/"+adminOrderDetailFixture().ID, ""))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if orders.id != adminOrderDetailFixture().ID {
		t.Errorf("id = %q", orders.id)
	}
	var envelope struct {
		Data adminOrderDetailResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	detail := envelope.Data
	if detail.Customer.Name != "Maria Silva" || detail.Customer.Phone != "+5585992358708" ||
		detail.Address == nil || detail.Address.ZipCode != "61700000" || detail.Zone == nil ||
		detail.Payment.Method != domain.PaymentMethodCash || len(detail.Items) != 1 ||
		len(detail.Items[0].Options) != 1 || len(detail.StatusHistory) != 1 {
		t.Fatalf("detail = %+v", detail)
	}
	if detail.TotalCents != 2800 || detail.CreatedAt != "2026-11-22T12:00:00Z" {
		t.Errorf("totals/timestamps = %+v", detail)
	}
}

func TestAdminGetOrderIncludesCustomerProfile(t *testing.T) {
	created := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	orders := &fakeAdminOrdersHandlerService{
		order: adminOrderDetailFixture(),
		profile: &domain.CustomerProfile{
			PhoneKey:      "5585992358708",
			Phone:         "+5585992358708",
			Name:          "Maria Silva",
			OrderCount:    3,
			LastOrderAt:   created,
			Fulfillment:   domain.FulfillmentDelivery,
			Zone:          &domain.OrderZone{Name: "Centro"},
			PaymentMethod: domain.PaymentMethodPix,
			UpdatedAt:     created,
			ExpiresAt:     created.AddDate(0, 24, 0),
		},
	}
	router := newAdminOrdersTestRouter(newAdminOperatorAuth(), orders)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/orders/"+adminOrderDetailFixture().ID, ""))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data adminOrderDetailResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	profile := envelope.Data.CustomerProfile
	if profile == nil || profile.OrderCount != 3 || profile.Name != "Maria Silva" || profile.ZoneName != "Centro" {
		t.Fatalf("customerProfile = %+v", profile)
	}
	if profile.LastOrderAt != "2026-11-20T12:00:00Z" || profile.RetainedUntil != "2028-11-20T12:00:00Z" {
		t.Errorf("profile timestamps = %+v", profile)
	}
	if orders.profilePhone != "+5585992358708" {
		t.Errorf("profile phone = %q", orders.profilePhone)
	}
}

func TestAdminGetOrderDoesNotFailWhenProfileLookupFails(t *testing.T) {
	orders := &fakeAdminOrdersHandlerService{
		order:      adminOrderDetailFixture(),
		profileErr: errors.New("mongo indisponível"),
	}
	router := newAdminOrdersTestRouter(newAdminOperatorAuth(), orders)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/orders/"+adminOrderDetailFixture().ID, ""))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if orders.profileCalls != 1 {
		t.Errorf("profileCalls = %d, want 1", orders.profileCalls)
	}
}

func TestAdminUpdateOrderStatus(t *testing.T) {
	orders := &fakeAdminOrdersHandlerService{order: adminOrderDetailFixture()}
	notifier := &fakeOrderNotifier{}
	router := NewRouter(Dependencies{
		AdminAuth:   newAdminOperatorAuth(),
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		AdminOrders: orders,
		Notifier:    notifier,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	body := `{"status":"cancelled","reason":"cliente desistiu"}`
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodPatch, "/api/v1/admin/orders/"+adminOrderDetailFixture().ID+"/status", body))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if len(notifier.events) != 1 {
		t.Fatalf("eventos = %d, esperado 1", len(notifier.events))
	}
	if notifier.events[0].Type != domain.WebhookEventOrderUpdated || notifier.events[0].Order.Status != domain.OrderStatusCancelled {
		t.Fatalf("evento inesperado: %+v", notifier.events[0])
	}
	if orders.actorID != testActorIDForHTTP || orders.id != adminOrderDetailFixture().ID ||
		orders.request.Status != domain.OrderStatusCancelled || orders.request.Reason != "cliente desistiu" {
		t.Fatalf("request = %+v, actor = %q, id = %q", orders.request, orders.actorID, orders.id)
	}
	var envelope struct {
		Data adminOrderDetailResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.Status != domain.OrderStatusCancelled {
		t.Errorf("response status = %q", envelope.Data.Status)
	}
}

func TestAdminOrderStatusErrorsAreMapped(t *testing.T) {
	path := "/api/v1/admin/orders/" + adminOrderDetailFixture().ID + "/status"
	body := `{"status":"confirmed"}`

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"transição inválida", &domain.OrderStatusTransitionError{From: "pending", To: "preparing"}, http.StatusConflict, codeConflict},
		{"não encontrado", domain.ErrNotFound, http.StatusNotFound, codeNotFound},
		{"sem sessão", domain.ErrUnauthorized, http.StatusUnauthorized, codeUnauthorized},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			orders := &fakeAdminOrdersHandlerService{order: adminOrderDetailFixture(), err: test.err}
			router := newAdminOrdersTestRouter(newAdminOperatorAuth(), orders)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, newAdminRequest(http.MethodPatch, path, body))
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

	validation := &fakeAdminOrdersHandlerService{
		order: adminOrderDetailFixture(),
		err:   &domain.ValidationError{Field: "reason", Code: domain.CodeRequired, Message: "obrigatório"},
	}
	router := newAdminOrdersTestRouter(newAdminOperatorAuth(), validation)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodPatch, path, body))
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("validation status = %d, want 422", response.Code)
	}
}

func TestAdminOrderRoutesRequireSessionRoleAndCSRF(t *testing.T) {
	router := newAdminOrdersTestRouter(newAdminOperatorAuth(), &fakeAdminOrdersHandlerService{order: adminOrderDetailFixture()})
	path := "/api/v1/admin/orders"

	anonymous := httptest.NewRecorder()
	request := newAdminRequest(http.MethodGet, path, "")
	request.Header.Del("Cookie")
	router.ServeHTTP(anonymous, request)
	if anonymous.Code != http.StatusUnauthorized {
		t.Errorf("anonymous status = %d, want 401", anonymous.Code)
	}

	csrf := httptest.NewRecorder()
	csrfRequest := newAdminRequest(http.MethodPatch, path+"/"+adminOrderDetailFixture().ID+"/status", `{"status":"confirmed"}`)
	csrfRequest.Header.Set("Origin", "https://evil.example")
	router.ServeHTTP(csrf, csrfRequest)
	if csrf.Code != http.StatusForbidden {
		t.Errorf("CSRF status = %d, want 403", csrf.Code)
	}

	unknownField := httptest.NewRecorder()
	router.ServeHTTP(unknownField, newAdminRequest(http.MethodPatch, path+"/"+adminOrderDetailFixture().ID+"/status", `{"status":"confirmed","totalCents":1}`))
	if unknownField.Code != http.StatusBadRequest {
		t.Errorf("unknown field status = %d, want 400", unknownField.Code)
	}

	deleteAsOperator := httptest.NewRecorder()
	id := adminOrderDetailFixture().ID
	router.ServeHTTP(deleteAsOperator, newAdminRequest(http.MethodDelete, path+"/"+id, ""))
	if deleteAsOperator.Code != http.StatusForbidden {
		t.Errorf("operator delete status = %d, want 403", deleteAsOperator.Code)
	}
}

func TestAdminDeleteOrderRequiresOwnerAndReturnsNoContent(t *testing.T) {
	orders := &fakeAdminOrdersHandlerService{order: adminOrderDetailFixture()}
	notifier := &fakeOrderNotifier{}
	router := NewRouter(Dependencies{
		AdminAuth:   newAdminOwnerAuth(),
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		AdminOrders: orders,
		Notifier:    notifier,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	id := adminOrderDetailFixture().ID
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodDelete, "/api/v1/admin/orders/"+id, ""))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body = %s", response.Code, response.Body.String())
	}
	if response.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", response.Body.String())
	}
	if orders.actorID != testActorIDForHTTP || orders.id != id {
		t.Fatalf("actor = %q, id = %q", orders.actorID, orders.id)
	}
	if len(notifier.events) != 0 {
		t.Fatalf("delete must not notify webhook: %+v", notifier.events)
	}
}

func TestAdminDeleteOrderErrorsAreMapped(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"não encontrado", domain.ErrNotFound, http.StatusNotFound, codeNotFound},
		{"sem sessão", domain.ErrUnauthorized, http.StatusUnauthorized, codeUnauthorized},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			orders := &fakeAdminOrdersHandlerService{err: test.err}
			router := newAdminOrdersTestRouter(newAdminOwnerAuth(), orders)
			path := "/api/v1/admin/orders/" + adminOrderDetailFixture().ID

			response := httptest.NewRecorder()
			router.ServeHTTP(response, newAdminRequest(http.MethodDelete, path, ""))
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

func TestAdminDeleteOrderRejectsBadCSRF(t *testing.T) {
	router := newAdminOrdersTestRouter(newAdminOwnerAuth(), &fakeAdminOrdersHandlerService{})
	request := newAdminRequest(http.MethodDelete, "/api/v1/admin/orders/"+adminOrderDetailFixture().ID, "")
	request.Header.Set("Origin", "https://evil.example")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}
