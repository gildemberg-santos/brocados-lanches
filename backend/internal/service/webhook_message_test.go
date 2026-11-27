package service

import (
	"strings"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
)

func TestBuildWebhookMessageUsesFrozenOrderAndStatus(t *testing.T) {
	order := domain.Order{
		Number:        1042,
		Status:        domain.OrderStatusConfirmed,
		Fulfillment:   domain.FulfillmentDelivery,
		Customer:      domain.OrderCustomer{Name: "Cliente", Phone: "+5511999991234"},
		SubtotalCents: 2000,
		TotalCents:    2000,
	}
	event := domain.WebhookEvent{Type: domain.WebhookEventOrderUpdated, Order: order}

	message := BuildWebhookMessage(event)
	if !strings.Contains(message, "Pedido #1042") {
		t.Fatalf("mensagem sem número: %q", message)
	}
	if !strings.Contains(message, "Status: Confirmado") {
		t.Fatalf("mensagem sem status: %q", message)
	}
	if strings.Contains(message, domain.WebhookEventOrderUpdated) {
		// Apenas o rótulo amigável deve aparecer; o identificador do evento é do cabeçalho.
		t.Fatalf("identificador do evento vazou na mensagem: %q", message)
	}
}

func TestBuildWebhookMessageAcceptsFutureStatus(t *testing.T) {
	order := domain.Order{Number: 1, Status: "novo_status_futuro"}
	message := BuildWebhookMessage(domain.WebhookEvent{
		Type:  domain.WebhookEventOrderUpdated,
		Order: order,
	})
	if !strings.Contains(message, "Status: novo_status_futuro") {
		t.Fatalf("mensagem não cobriu status futuro: %q", message)
	}
}
