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

type fakeOrderService struct {
	quote    domain.Quote
	err      error
	received domain.QuoteRequest
}

func (f *fakeOrderService) Quote(_ context.Context, request domain.QuoteRequest) (domain.Quote, error) {
	f.received = request
	return f.quote, f.err
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
		Pinger: fakePinger{},
		Orders: orders,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
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

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
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
