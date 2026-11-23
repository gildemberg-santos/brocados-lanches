package service

import (
	"net/url"
	"strings"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
)

func whatsappOrderFixture() domain.Order {
	changeForCents := int64(10000)
	return domain.Order{
		Number:           1042,
		Fulfillment:      domain.FulfillmentDelivery,
		SubtotalCents:    5400,
		DeliveryFeeCents: 800,
		TotalCents:       6200,
		Customer: domain.OrderCustomer{
			Name:  "Maria Silva",
			Phone: "+5585999999999",
		},
		Address: &domain.OrderAddress{
			Street:   "Rua das Flores",
			Number:   "450",
			District: "Tapera",
			City:     "Aquiraz",
			State:    "CE",
			ZipCode:  "61700-000",
		},
		Items: []domain.OrderItem{{
			Name:           "Bacon Cheddar",
			Quantity:       2,
			UnitPriceCents: 2200,
			LineTotalCents: 5400,
			Note:           "sem cebola",
			Options: []domain.OrderItemOption{
				{GroupName: "Ponto da carne", Name: "Ao ponto", PriceCents: 0},
				{GroupName: "Adicionais", Name: "Bacon", PriceCents: 500},
			},
		}},
		Note:    "Deixar na portaria",
		Payment: domain.OrderPayment{Method: domain.PaymentMethodCash, ChangeForCents: &changeForCents},
	}
}

func decodedWhatsAppMessage(t *testing.T, store domain.StoreSettings, order domain.Order, wantNumber string) string {
	t.Helper()
	link, err := BuildWhatsAppURL(store, order)
	if err != nil {
		t.Fatalf("BuildWhatsAppURL() error = %v", err)
	}
	prefix := "https://wa.me/" + wantNumber + "?text="
	if !strings.HasPrefix(link, prefix) {
		t.Fatalf("link = %q, want prefix %q", link, prefix)
	}

	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse link: %v", err)
	}
	return parsed.Query().Get("text")
}

func TestBuildWhatsAppURLIncludesFrozenOrderDetails(t *testing.T) {
	message := decodedWhatsAppMessage(t, domain.StoreSettings{WhatsApp: "(85) 99999-9999"}, whatsappOrderFixture(), "5585999999999")

	for _, expected := range []string{
		"Pedido #1042",
		"2x Bacon Cheddar (Ao ponto, Bacon) — sem cebola",
		"Observação: Deixar na portaria",
		"Subtotal: R$ 54.00",
		"Taxa de entrega: R$ 8.00",
		"Total: R$ 62.00",
		"Pagamento: dinheiro, troco para R$ 100.00",
		"Cliente: Maria Silva — +5585999999999",
		"Entrega: Rua das Flores, 450 — Tapera, Aquiraz/CE — CEP 61700-000",
	} {
		if !strings.Contains(message, expected) {
			t.Errorf("message missing %q:\n%s", expected, message)
		}
	}
}

func TestBuildWhatsAppURLPaymentAndPickupLabels(t *testing.T) {
	tests := []struct {
		name        string
		payment     domain.OrderPayment
		fulfillment domain.Fulfillment
		wantPayment string
		want        string
	}{
		{
			name:        "cartão na retirada",
			payment:     domain.OrderPayment{Method: domain.PaymentMethodCard},
			fulfillment: domain.FulfillmentPickup,
			wantPayment: "Pagamento: cartão na entrega",
			want:        "Retirada na loja",
		},
		{
			name:        "pix na entrega",
			payment:     domain.OrderPayment{Method: domain.PaymentMethodPix},
			fulfillment: domain.FulfillmentDelivery,
			want:        "Pagamento: Pix na entrega",
		},
		{
			name:        "dinheiro sem troco",
			payment:     domain.OrderPayment{Method: domain.PaymentMethodCash},
			fulfillment: domain.FulfillmentDelivery,
			want:        "Pagamento: dinheiro",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			order := whatsappOrderFixture()
			order.Payment = test.payment
			order.Fulfillment = test.fulfillment

			message := decodedWhatsAppMessage(t, domain.StoreSettings{WhatsApp: "+558599999999"}, order, "558599999999")
			if test.wantPayment != "" && !strings.Contains(message, test.wantPayment) {
				t.Errorf("message missing %q:\n%s", test.wantPayment, message)
			}
			if !strings.Contains(message, test.want) {
				t.Errorf("message missing %q:\n%s", test.want, message)
			}
		})
	}
}

func TestBuildWhatsAppURLRejectsMalformedStoreNumbers(t *testing.T) {
	tests := []string{"", "abc", "12345", "12345678901234"}
	for _, number := range tests {
		t.Run(number, func(t *testing.T) {
			if _, err := BuildWhatsAppURL(domain.StoreSettings{WhatsApp: number}, whatsappOrderFixture()); err == nil {
				t.Errorf("BuildWhatsAppURL(%q) error = nil, want error", number)
			}
		})
	}
}
