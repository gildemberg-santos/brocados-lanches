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
	"github.com/brocados-lanches/api/internal/platform/token"
	"github.com/brocados-lanches/api/internal/service"
)

var testAccessToken = token.DeriveOrderAccess("test-secret", "test-source")

type fakeOrderService struct {
	quote    domain.Quote
	err      error
	received domain.QuoteRequest

	created       domain.Order
	createdErr    error
	replayed      bool
	receivedOrder domain.OrderRequest
	receivedKey   string
	receivedToken string

	tracked      domain.Order
	trackedErr   error
	history      service.OrderHistoryPage
	historyErr   error
	historyCalls int
	store        domain.StoreSettings
	storeErr     error

	quoteCalls  int
	createCalls int
	trackCalls  int
}

func (f *fakeOrderService) Quote(_ context.Context, request domain.QuoteRequest) (domain.Quote, error) {
	f.quoteCalls++
	f.received = request
	return f.quote, f.err
}

func (f *fakeOrderService) Create(_ context.Context, request domain.OrderRequest, key, accessToken string) (domain.Order, bool, error) {
	f.createCalls++
	f.receivedOrder = request
	f.receivedKey = key
	f.receivedToken = accessToken
	return f.created, f.replayed, f.createdErr
}

func (f *fakeOrderService) GetByPublicID(context.Context, string) (domain.Order, error) {
	f.trackCalls++
	return f.tracked, f.trackedErr
}

func (f *fakeOrderService) ListByAccessToken(context.Context, string, domain.OrderHistoryQuery) (service.OrderHistoryPage, error) {
	f.historyCalls++
	return f.history, f.historyErr
}

func (f *fakeOrderService) StoreSettings(context.Context) (domain.StoreSettings, error) {
	return f.store, f.storeErr
}

const validQuoteBody = `{
	"items": [{
		"productId": "prod-1",
		"quantity": 2,
		"optionIds": ["medium"],
		"addonIds": ["addon-1"],
		"note": "sem cebola"
	}],
	"fulfillment": "delivery",
	"zoneId": "zone-1"
}`

func newOrderTestRouter(orders OrderService) http.Handler {
	return NewRouter(Dependencies{
		Pinger:   fakePinger{},
		Orders:   orders,
		Notifier: &fakeOrderNotifier{},
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

type fakeOrderNotifier struct {
	events []domain.WebhookEvent
}

func (f *fakeOrderNotifier) Notify(_ context.Context, event domain.WebhookEvent) {
	f.events = append(f.events, event)
}

func performQuoteRequest(handler http.Handler, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders/quote", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func quoteFixture() domain.Quote {
	return domain.Quote{
		SubtotalCents:    5400,
		DeliveryFeeCents: 800,
		TotalCents:       6200,
		Currency:         domain.CurrencyBRL,
		Items: []domain.QuoteLine{{
			ProductID:         "prod-1",
			Name:              "Bacon Cheddar",
			Quantity:          2,
			UnitPriceCents:    2200,
			OptionsPriceCents: 500,
			LineTotalCents:    5400,
			Options: []domain.QuoteOption{
				{Name: "Ao ponto", PriceCents: 0},
				{Name: "Bacon", PriceCents: 500},
			},
		}},
		Warnings: []string{},
	}
}

func TestQuoteSuccess(t *testing.T) {
	orders := &fakeOrderService{quote: quoteFixture()}
	response := performQuoteRequest(newOrderTestRouter(orders), validQuoteBody)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}

	var body quoteEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.SubtotalCents != 5400 || body.Data.DeliveryFeeCents != 800 || body.Data.TotalCents != 6200 {
		t.Errorf("totals = %+v", body.Data)
	}
	if body.Data.Currency != "BRL" || len(body.Data.Items) != 1 || len(body.Data.Items[0].Options) != 2 {
		t.Errorf("data = %+v", body.Data)
	}
	if len(body.Data.Warnings) != 0 {
		t.Errorf("warnings = %v, want empty", body.Data.Warnings)
	}

	if orders.received.Items[0].ProductID != "prod-1" || orders.received.Items[0].Quantity != 2 {
		t.Errorf("received = %+v", orders.received)
	}
	if orders.received.Items[0].Note != "sem cebola" || orders.received.ZoneID != "zone-1" {
		t.Errorf("received = %+v", orders.received)
	}
}

func TestQuoteMalformedBody(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "JSON inválido", body: `{"items":`},
		{name: "campo desconhecido", body: `{"items": [], "priceCents": 100}`},
		{name: "campo desconhecido no item", body: `{"items": [{"productId": "p", "quantity": 1, "unitPriceCents": 100}], "fulfillment": "pickup"}`},
		{name: "corpo vazio", body: ``},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performQuoteRequest(newOrderTestRouter(&fakeOrderService{}), test.body)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body = %s", response.Code, response.Body.String())
			}

			var body errorEnvelope
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Error.Code != codeValidationError {
				t.Errorf("error code = %q, want %q", body.Error.Code, codeValidationError)
			}
			if len(body.Error.Details) != 1 || body.Error.Details[0].Field != "body" {
				t.Errorf("details = %+v", body.Error.Details)
			}
		})
	}
}

func TestQuoteOversizedBody(t *testing.T) {
	body := `{"items": [{"productId": "` + strings.Repeat("a", maxQuoteBodyBytes+1) + `"}]}`
	response := performQuoteRequest(newOrderTestRouter(&fakeOrderService{}), body)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestQuoteValidationError(t *testing.T) {
	orders := &fakeOrderService{err: &domain.ValidationError{
		Field:   "items[0].quantity",
		Code:    domain.CodeOutOfRange,
		Message: "must be between 1 and 99",
	}}
	response := performQuoteRequest(newOrderTestRouter(orders), validQuoteBody)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}

	var body errorEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Code != codeValidationError || len(body.Error.Details) != 1 {
		t.Fatalf("error = %+v", body.Error)
	}
	if body.Error.Details[0].Field != "items[0].quantity" {
		t.Errorf("detail = %+v", body.Error.Details[0])
	}
}

func TestQuoteConflictErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
	}{
		{name: "loja fechada", err: domain.ErrStoreClosed, code: codeStoreClosed},
		{name: "zona indisponível", err: domain.ErrZoneUnavailable, code: codeZoneUnavailable},
		{name: "produto indisponível", err: domain.ErrProductUnavailable, code: codeProductUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			orders := &fakeOrderService{err: test.err}
			response := performQuoteRequest(newOrderTestRouter(orders), validQuoteBody)

			if response.Code != http.StatusConflict {
				t.Fatalf("status = %d, want 409; body = %s", response.Code, response.Body.String())
			}

			var body errorEnvelope
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Error.Code != test.code {
				t.Errorf("error code = %q, want %q", body.Error.Code, test.code)
			}
			if body.Error.Message == "" || strings.Contains(body.Error.Message, test.err.Error()) {
				t.Errorf("message = %q", body.Error.Message)
			}
		})
	}
}

const validOrderBody = `{
	"items": [{
		"productId": "prod-1",
		"quantity": 2,
		"optionIds": ["medium"],
		"addonIds": ["addon-1"],
		"note": "sem cebola"
	}],
	"fulfillment": "delivery",
	"zoneId": "zone-1",
	"customer": { "name": "Maria Silva", "phone": "+5585999999999" },
	"address": {
		"street": "Rua das Flores", "number": "450", "district": "Tapera",
		"city": "Aquiraz", "state": "CE", "zipCode": "61700-000"
	},
	"payment": { "method": "cash", "changeForCents": 10000 },
	"note": "Deixar na portaria"
}`

const testOrderHeaderKey = "4d1f6b1e-7c2a-4e2f-9f2a-1c3d5e7f9a0b"

func createdOrderFixture() domain.Order {
	return domain.Order{
		ID:               "order-1",
		Number:           1042,
		PublicID:         "public-123",
		Status:           domain.OrderStatusPending,
		Fulfillment:      domain.FulfillmentDelivery,
		SubtotalCents:    5400,
		DeliveryFeeCents: 800,
		TotalCents:       6200,
		CreatedAt:        time.Date(2026, 11, 22, 22, 0, 0, 0, time.UTC),
		AccessToken:      testAccessToken,
	}
}

func performCreateOrderRequest(handler http.Handler, key, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func performCreateOrderRequestWithToken(handler http.Handler, key, accessToken, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", key)
	request.Header.Set("Authorization", "Bearer "+accessToken)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func TestCreateOrderForwardsBrowserAccessToken(t *testing.T) {
	orders := &fakeOrderService{
		created: createdOrderFixture(),
		store:   domain.StoreSettings{WhatsApp: "(85) 99999-9999"},
	}

	performCreateOrderRequestWithToken(newOrderTestRouter(orders), testOrderHeaderKey, testAccessToken, validOrderBody)

	if orders.receivedToken != testAccessToken {
		t.Errorf("received token = %q, want %q", orders.receivedToken, testAccessToken)
	}
}

func TestCreateOrderRejectsInvalidAuthorizationHeader(t *testing.T) {
	response := performCreateOrderRequestWithToken(newOrderTestRouter(&fakeOrderService{}), testOrderHeaderKey, "invalido", validOrderBody)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestOrderHistoryReturnsPagedSummaries(t *testing.T) {
	orders := &fakeOrderService{
		history: service.OrderHistoryPage{
			Items: []domain.Order{{
				Number:      1042,
				PublicID:    "public-123",
				Status:      domain.OrderStatusPending,
				Fulfillment: domain.FulfillmentDelivery,
				Items:       []domain.OrderItem{{Name: "Bacon Cheddar", Quantity: 2}},
				TotalCents:  6200,
				CreatedAt:   time.Date(2026, 11, 22, 22, 0, 0, 0, time.UTC),
				UpdatedAt:   time.Date(2026, 11, 22, 22, 0, 0, 0, time.UTC),
			}},
			Page: 1, PageSize: 20, Total: 1, TotalPages: 1,
		},
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/orders/history?page=1&pageSize=20", nil)
	request.Header.Set("Authorization", "Bearer "+testAccessToken)
	response := httptest.NewRecorder()
	newOrderTestRouter(orders).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Error("history response must not be cached")
	}
	var body orderHistoryEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.Data) != 1 || body.Data[0].PublicID != "public-123" || body.Meta.Total != 1 {
		t.Errorf("body = %+v", body)
	}
	if strings.Contains(response.Body.String(), "Maria Silva") {
		t.Error("history must not expose customer PII")
	}
}

func TestOrderHistoryRequiresAccessToken(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/orders/history", nil)
	newOrderTestRouter(&fakeOrderService{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
}

func TestCreateOrderSuccess(t *testing.T) {
	orders := &fakeOrderService{
		created: createdOrderFixture(),
		store:   domain.StoreSettings{WhatsApp: "(85) 99999-9999"},
	}
	response := performCreateOrderRequest(newOrderTestRouter(orders), testOrderHeaderKey, validOrderBody)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	if location := response.Header().Get("Location"); location != "/api/v1/orders/public-123" {
		t.Errorf("Location = %q", location)
	}
	if orders.receivedKey != testOrderHeaderKey {
		t.Errorf("idempotency key = %q", orders.receivedKey)
	}
	if orders.receivedOrder.Customer.Phone != "+5585999999999" {
		t.Errorf("customer = %+v", orders.receivedOrder.Customer)
	}

	var body orderEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Data.OrderNumber != 1042 || body.Data.TotalCents != 6200 || body.Data.Status != "pending" {
		t.Errorf("data = %+v", body.Data)
	}
	if !strings.HasPrefix(body.Data.WhatsAppURL, "https://wa.me/5585999999999?text=") {
		t.Errorf("whatsappUrl = %q", body.Data.WhatsAppURL)
	}
	if body.Data.AccessToken != testAccessToken {
		t.Errorf("accessToken = %q, want %q", body.Data.AccessToken, testAccessToken)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Error("order response must not be cached")
	}
}

func TestCreateOrderKeepsPersistedOrderWhenHandoffUnavailable(t *testing.T) {
	tests := []struct {
		name   string
		orders *fakeOrderService
	}{
		{
			name: "falha ao carregar configurações da loja",
			orders: &fakeOrderService{
				created:  createdOrderFixture(),
				storeErr: errors.New("mongo down"),
			},
		},
		{
			name: "número de WhatsApp inválido",
			orders: &fakeOrderService{
				created: createdOrderFixture(),
				store:   domain.StoreSettings{WhatsApp: "invalid"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performCreateOrderRequest(newOrderTestRouter(test.orders), testOrderHeaderKey, validOrderBody)

			if response.Code != http.StatusCreated {
				t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
			}
			if location := response.Header().Get("Location"); location != "/api/v1/orders/public-123" {
				t.Errorf("Location = %q", location)
			}

			body := response.Body.String()
			if strings.Contains(body, "mongo down") || strings.Contains(body, "invalid store whatsapp") {
				t.Errorf("response exposes internal error: %s", body)
			}
			var envelope orderEnvelope
			if err := json.NewDecoder(strings.NewReader(body)).Decode(&envelope); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if envelope.Data.OrderNumber != 1042 || envelope.Data.WhatsAppURL != "" {
				t.Errorf("data = %+v", envelope.Data)
			}
		})
	}
}

func TestCreateOrderIdempotentReplay(t *testing.T) {
	orders := &fakeOrderService{
		created:  createdOrderFixture(),
		replayed: true,
		store:    domain.StoreSettings{WhatsApp: "(85) 99999-9999"},
	}
	notifier := &fakeOrderNotifier{}
	router := NewRouter(Dependencies{
		Pinger:   fakePinger{},
		Orders:   orders,
		Notifier: notifier,
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	response := performCreateOrderRequest(router, testOrderHeaderKey, validOrderBody)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if response.Header().Get("Idempotent-Replayed") != "true" {
		t.Error("missing Idempotent-Replayed header")
	}
	if len(notifier.events) != 0 {
		t.Fatalf("replay não deve gerar webhook: %d eventos", len(notifier.events))
	}
}

func TestCreateOrderNotifiesWebhookWithPersistedOrder(t *testing.T) {
	orders := &fakeOrderService{
		created: createdOrderFixture(),
		store:   domain.StoreSettings{WhatsApp: "(85) 99999-9999"},
	}
	notifier := &fakeOrderNotifier{}
	router := NewRouter(Dependencies{
		Pinger:   fakePinger{},
		Orders:   orders,
		Notifier: notifier,
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	response := performCreateOrderRequest(router, testOrderHeaderKey, validOrderBody)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", response.Code)
	}
	if len(notifier.events) != 1 {
		t.Fatalf("eventos = %d, esperado 1", len(notifier.events))
	}
	if notifier.events[0].Type != domain.WebhookEventOrderCreated {
		t.Fatalf("tipo = %q", notifier.events[0].Type)
	}
	if notifier.events[0].Order.Number != 1042 {
		t.Fatalf("pedido = %+v", notifier.events[0].Order)
	}
}

func TestCreateOrderRequiresIdempotencyKey(t *testing.T) {
	response := performCreateOrderRequest(newOrderTestRouter(&fakeOrderService{}), "", validOrderBody)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	var body errorEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Details[0].Field != "Idempotency-Key" {
		t.Errorf("detail = %+v", body.Error.Details)
	}
}

func TestCreateOrderMapsConflict(t *testing.T) {
	orders := &fakeOrderService{createdErr: domain.ErrConflict}
	response := performCreateOrderRequest(newOrderTestRouter(orders), testOrderHeaderKey, validOrderBody)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", response.Code, response.Body.String())
	}
}

func TestTrackOrderHidesPersonalData(t *testing.T) {
	order := createdOrderFixture()
	order.Customer = domain.OrderCustomer{Name: "Maria Silva", Phone: "+5585999999999"}
	order.Address = &domain.OrderAddress{Street: "Rua das Flores", Number: "450"}
	order.Items = []domain.OrderItem{{Name: "Bacon Cheddar", Quantity: 2}}
	orders := &fakeOrderService{
		tracked: order,
		store:   domain.StoreSettings{DeliveryTime: "35-50 min"},
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/orders/public-123", nil)
	response := httptest.NewRecorder()
	newOrderTestRouter(orders).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	body := response.Body.String()
	for _, forbidden := range []string{"Maria Silva", "+5585999999999", "Rua das Flores", "whatsappUrl"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("tracking response leaked %q: %s", forbidden, body)
		}
	}
	var envelope trackingEnvelope
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&envelope); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if envelope.Data.EstimatedTime != "35-50 min" || len(envelope.Data.Items) != 1 {
		t.Errorf("data = %+v", envelope.Data)
	}
}

func TestQuoteServiceError(t *testing.T) {
	orders := &fakeOrderService{err: errors.New("mongo down")}
	response := performQuoteRequest(newOrderTestRouter(orders), validQuoteBody)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}

	var body errorEnvelope
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Code != codeInternalError {
		t.Errorf("error code = %q, want %q", body.Error.Code, codeInternalError)
	}
	if strings.Contains(body.Error.Message, "mongo down") {
		t.Error("response exposes internal error")
	}
}

func TestQuoteServiceNotConfigured(t *testing.T) {
	response := performQuoteRequest(newOrderTestRouter(nil), validQuoteBody)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.Code)
	}
}
