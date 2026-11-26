package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func orderRequestFixture() OrderRequest {
	request, _ := quoteFixture()
	changeForCents := int64(10000)
	return OrderRequest{
		Items:       request.Items,
		Fulfillment: request.Fulfillment,
		ZoneID:      request.ZoneID,
		Customer:    OrderCustomer{Name: "Maria Silva", Phone: "+55 85 99999-9999"},
		Address: &OrderAddress{
			Street:   "Rua das Flores",
			Number:   "450",
			District: "Tapera",
			City:     "Aquiraz",
			State:    "CE",
			ZipCode:  "61700-000",
		},
		Payment: OrderPayment{Method: PaymentMethodCash, ChangeForCents: &changeForCents},
		Note:    "Deixar na portaria",
	}
}

func TestOrderRequestValidate(t *testing.T) {
	if err := orderRequestFixture().Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestOrderRequestValidateRequiresDeliveryAddress(t *testing.T) {
	request := orderRequestFixture()
	request.Address = nil

	err := request.Validate()
	var validationError *ValidationError
	if !errors.As(err, &validationError) || validationError.Field != "address" {
		t.Fatalf("error = %v, want address required", err)
	}
}

func TestOrderRequestValidateRejectsInvalidPhone(t *testing.T) {
	request := orderRequestFixture()
	request.Customer.Phone = "123"

	err := request.Validate()
	var validationError *ValidationError
	if !errors.As(err, &validationError) || validationError.Field != "customer.phone" {
		t.Fatalf("error = %v, want invalid phone", err)
	}
}

func TestOrderRequestValidateRejectsChangeForNonCash(t *testing.T) {
	request := orderRequestFixture()
	request.Payment.Method = PaymentMethodPix

	err := request.Validate()
	var validationError *ValidationError
	if !errors.As(err, &validationError) || validationError.Field != "payment.changeForCents" {
		t.Fatalf("error = %v, want invalid change", err)
	}
}

func TestOrderRequestValidateWithTotal(t *testing.T) {
	request := orderRequestFixture()
	changeForCents := int64(6100)
	request.Payment.ChangeForCents = &changeForCents

	if err := request.ValidateWithTotal(6200); err == nil {
		t.Fatal("ValidateWithTotal() error = nil, want insufficient change")
	}
	if err := request.ValidateWithTotal(6000); err != nil {
		t.Fatalf("ValidateWithTotal() error = %v", err)
	}
}

func TestNewOrderFreezesQuoteValues(t *testing.T) {
	_, catalog := quoteFixture()
	request := orderRequestFixture()
	quote, err := BuildQuote(QuoteRequest{Items: request.Items, Fulfillment: request.Fulfillment, ZoneID: request.ZoneID}, catalog)
	if err != nil {
		t.Fatalf("BuildQuote() error = %v", err)
	}

	now := time.Date(2026, 11, 22, 22, 0, 0, 0, time.UTC)
	order := NewOrder(request, quote, quote.Zone, "order-1", 1042, "public-id", "idempotency-key", "hash", now)

	if order.Status != OrderStatusPending || order.Number != 1042 || order.PublicID != "public-id" {
		t.Errorf("order identity = %+v", order)
	}
	if order.TotalCents != 6200 || order.SubtotalCents != 5400 || order.DeliveryFeeCents != 800 {
		t.Errorf("order totals = %+v", order)
	}
	if order.Zone == nil || order.Zone.Name != "Centro" || order.Zone.FeeCents != 800 {
		t.Errorf("zone snapshot = %+v", order.Zone)
	}
	if len(order.Items) != 1 || order.Items[0].Note != request.Items[0].Note {
		t.Fatalf("items = %+v", order.Items)
	}
	if len(order.Items[0].Options) != 2 {
		t.Fatalf("options = %+v", order.Items[0].Options)
	}
	if order.Items[0].Options[0].GroupID != "doneness" || order.Items[0].Options[0].GroupName != "Ponto da carne" {
		t.Errorf("option group = %+v", order.Items[0].Options[0])
	}
	if order.Items[0].Options[1].Kind != QuoteOptionKindAddon || order.Items[0].Options[1].GroupID != "addons" {
		t.Errorf("addon metadata = %+v", order.Items[0].Options[1])
	}
	if len(order.StatusHistory) != 1 || order.StatusHistory[0].Status != OrderStatusPending {
		t.Errorf("status history = %+v", order.StatusHistory)
	}
	if !order.CreatedAt.Equal(now) || !order.UpdatedAt.Equal(now) {
		t.Errorf("timestamps = %s/%s", order.CreatedAt, order.UpdatedAt)
	}
}

func TestCanonicalRequestHash(t *testing.T) {
	first, err := CanonicalRequestHash(orderRequestFixture())
	if err != nil {
		t.Fatalf("CanonicalRequestHash() error = %v", err)
	}
	second, err := CanonicalRequestHash(orderRequestFixture())
	if err != nil {
		t.Fatalf("CanonicalRequestHash() error = %v", err)
	}
	if first != second || len(first) != 64 {
		t.Errorf("hash = %q, second = %q", first, second)
	}

	different := orderRequestFixture()
	different.Note = "Outra observação"
	third, err := CanonicalRequestHash(different)
	if err != nil {
		t.Fatalf("CanonicalRequestHash() error = %v", err)
	}
	if third == first {
		t.Error("hash did not change with request body")
	}

	// A preferência de perfil não altera o pedido nem a idempotência: um replay
	// com outra escolha continua devolvendo o mesmo pedido gravado.
	withOptOut := orderRequestFixture()
	save := false
	withOptOut.SaveCustomerProfile = &save
	fourth, err := CanonicalRequestHash(withOptOut)
	if err != nil {
		t.Fatalf("CanonicalRequestHash() opt-out error = %v", err)
	}
	if fourth != first {
		t.Error("hash changed with customer profile preference")
	}
}

func TestOrderRequestValidateRejectsLongNote(t *testing.T) {
	request := orderRequestFixture()
	request.Note = strings.Repeat("a", MaxOrderNoteLength+1)

	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want note too long")
	}
}
