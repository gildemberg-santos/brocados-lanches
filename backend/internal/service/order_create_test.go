package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

const testIdempotencyKey = "4d1f6b1e-7c2a-4e2f-9f2a-1c3d5e7f9a0b"

func createOrderRequestFixture() domain.OrderRequest {
	changeForCents := int64(10000)
	return domain.OrderRequest{
		Items: []domain.QuoteItem{{
			ProductID: "prod-1",
			Quantity:  2,
			OptionIDs: []string{"medium"},
			AddonIDs:  []string{"addon-1"},
			Note:      "sem cebola",
		}},
		Fulfillment: domain.FulfillmentDelivery,
		ZoneID:      "zone-1",
		Customer:    domain.OrderCustomer{Name: "Maria Silva", Phone: "+5585999999999"},
		Address: &domain.OrderAddress{
			Street:   "Rua das Flores",
			Number:   "450",
			District: "Tapera",
			City:     "Aquiraz",
			State:    "CE",
			ZipCode:  "61700-000",
		},
		Payment: domain.OrderPayment{Method: domain.PaymentMethodCash, ChangeForCents: &changeForCents},
	}
}

func TestOrderServiceCreatePersistsAuthoritativeValues(t *testing.T) {
	reader, service := orderServiceFixture()
	service.publicID = func() (string, error) { return "public-123", nil }
	reader.created = domain.Order{
		ID:        "order-1",
		PublicID:  "public-123",
		Number:    1042,
		Status:    domain.OrderStatusPending,
		CreatedAt: time.Date(2026, 11, 22, 22, 0, 0, 0, time.UTC),
	}

	order, replayed, err := service.Create(context.Background(), createOrderRequestFixture(), testIdempotencyKey)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if replayed {
		t.Error("replayed = true, want false")
	}
	if order.Number != 1042 {
		t.Errorf("number = %d, want 1042", order.Number)
	}
	if reader.receivedOrder.SubtotalCents != 5400 || reader.receivedOrder.DeliveryFeeCents != 800 || reader.receivedOrder.TotalCents != 6200 {
		t.Errorf("stored totals = %+v", reader.receivedOrder)
	}
	if reader.receivedOrder.PublicID != "public-123" || reader.receivedOrder.IdempotencyKey != testIdempotencyKey {
		t.Errorf("stored identity = %+v", reader.receivedOrder)
	}
	if reader.receivedOrder.RequestHash == "" || len(reader.receivedOrder.RequestHash) != 64 {
		t.Errorf("request hash = %q", reader.receivedOrder.RequestHash)
	}
}

func TestOrderServiceCreateReplaysSameKey(t *testing.T) {
	reader, service := orderServiceFixture()
	reader.existingFound = true
	reader.existing = domain.Order{ID: "order-1", PublicID: "public-123", Number: 1042}
	hash, err := domain.CanonicalRequestHash(createOrderRequestFixture())
	if err != nil {
		t.Fatal(err)
	}
	reader.existing.RequestHash = hash

	order, replayed, err := service.Create(context.Background(), createOrderRequestFixture(), testIdempotencyKey)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !replayed || order.ID != "order-1" {
		t.Errorf("order = %+v, replayed = %v", order, replayed)
	}
	for _, call := range reader.calls {
		if call == "create" {
			t.Fatal("Create() called repository insert on replay")
		}
	}
}

func TestOrderServiceCreateRejectsDifferentBodyForSameKey(t *testing.T) {
	reader, service := orderServiceFixture()
	reader.existingFound = true
	reader.existing = domain.Order{RequestHash: "hash-de-outra-requisicao"}

	_, _, err := service.Create(context.Background(), createOrderRequestFixture(), testIdempotencyKey)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestOrderServiceCreateRequiresIdempotencyKey(t *testing.T) {
	_, service := orderServiceFixture()

	_, _, err := service.Create(context.Background(), createOrderRequestFixture(), "curta")
	var validationError *domain.ValidationError
	if !errors.As(err, &validationError) || validationError.Field != "Idempotency-Key" {
		t.Fatalf("error = %v, want invalid key", err)
	}
}

func TestOrderServiceCreateValidatesChangeAgainstServerTotal(t *testing.T) {
	_, service := orderServiceFixture()
	request := createOrderRequestFixture()
	changeForCents := int64(6100)
	request.Payment.ChangeForCents = &changeForCents

	_, _, err := service.Create(context.Background(), request, testIdempotencyKey)
	var validationError *domain.ValidationError
	if !errors.As(err, &validationError) || validationError.Field != "payment.changeForCents" {
		t.Fatalf("error = %v, want insufficient change", err)
	}
}

func TestOrderServiceGetByPublicID(t *testing.T) {
	reader, service := orderServiceFixture()
	reader.byPublicID = domain.Order{PublicID: "public-123", Number: 1042}

	order, err := service.GetByPublicID(context.Background(), "public-123")
	if err != nil {
		t.Fatalf("GetByPublicID() error = %v", err)
	}
	if order.PublicID != "public-123" || order.Number != 1042 {
		t.Errorf("order = %+v", order)
	}
}

func TestBuildWhatsAppURL(t *testing.T) {
	store := domain.StoreSettings{WhatsApp: "(85) 99999-9999", DeliveryTime: "35-50 min"}
	order := domain.Order{
		Number:      1042,
		Fulfillment: domain.FulfillmentDelivery,
		Customer:    domain.OrderCustomer{Name: "Maria Silva", Phone: "+5585999999999"},
		Address: &domain.OrderAddress{
			Street: "Rua das Flores", Number: "450", District: "Tapera",
			City: "Aquiraz", State: "CE", ZipCode: "61700-000",
		},
		Items: []domain.OrderItem{{Name: "Bacon Cheddar", Quantity: 2}},
		Payment: domain.OrderPayment{
			Method: domain.PaymentMethodCash,
		},
		SubtotalCents: 5400, DeliveryFeeCents: 800, TotalCents: 6200,
	}

	link, err := BuildWhatsAppURL(store, order)
	if err != nil {
		t.Fatalf("BuildWhatsAppURL() error = %v", err)
	}
	if !strings.HasPrefix(link, "https://wa.me/5585999999999?text=") {
		t.Errorf("link prefix = %q", link)
	}
	if !strings.Contains(link, url.QueryEscape("Pedido #1042")) {
		t.Errorf("link = %q", link)
	}
	if !strings.Contains(link, url.QueryEscape("Total: R$ 62.00")) {
		t.Errorf("link missing total: %q", link)
	}
}

func TestBuildWhatsAppURLRejectsInvalidStoreNumber(t *testing.T) {
	_, err := BuildWhatsAppURL(domain.StoreSettings{WhatsApp: "123"}, domain.Order{})
	if err == nil {
		t.Fatal("BuildWhatsAppURL() error = nil, want invalid number")
	}
}
