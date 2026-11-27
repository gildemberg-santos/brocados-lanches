package service

import (
	"strings"
	"testing"

	"github.com/brocados-lanches/api/internal/domain"
)

const testAdminBaseURL = "http://localhost:5173"

func TestBuildWebhookMessageUsesFrozenOrderAndStatus(t *testing.T) {
	order := domain.Order{
		ID:            "6b0913ed4c6162f4586e7774",
		Number:        1042,
		Status:        domain.OrderStatusConfirmed,
		Fulfillment:   domain.FulfillmentDelivery,
		Customer:      domain.OrderCustomer{Name: "Cliente", Phone: "+5511999991234"},
		SubtotalCents: 2000,
		TotalCents:    2000,
	}
	event := domain.WebhookEvent{Type: domain.WebhookEventOrderUpdated, Order: order}

	message := BuildWebhookMessage(event, testAdminBaseURL)
	if !strings.Contains(message, "Pedido #1042") {
		t.Fatalf("mensagem sem número: %q", message)
	}
	if !strings.Contains(message, "Status: Confirmado") {
		t.Fatalf("mensagem sem status: %q", message)
	}
	if !strings.Contains(message, "http://localhost:5173/admin/pedidos/6b0913ed4c6162f4586e7774") {
		t.Fatalf("mensagem sem link do pedido: %q", message)
	}
	if strings.Contains(message, domain.WebhookEventOrderUpdated) {
		// Apenas o rótulo amigável deve aparecer; o identificador do evento é do cabeçalho.
		t.Fatalf("identificador do evento vazou na mensagem: %q", message)
	}
}

func TestBuildWebhookMessageAcceptsFutureStatus(t *testing.T) {
	order := domain.Order{ID: "6b0913ed4c6162f4586e7774", Number: 1, Status: "novo_status_futuro"}
	message := BuildWebhookMessage(domain.WebhookEvent{
		Type:  domain.WebhookEventOrderUpdated,
		Order: order,
	}, testAdminBaseURL)
	if !strings.Contains(message, "Status: novo_status_futuro") {
		t.Fatalf("mensagem não cobriu status futuro: %q", message)
	}
}

func TestAdminOrderURLLink(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		orderID string
		want    string
	}{
		{
			name:    "desenvolvimento com barra final",
			baseURL: "http://localhost:5173/",
			orderID: "6b0913ed4c6162f4586e7774",
			want:    "http://localhost:5173/admin/pedidos/6b0913ed4c6162f4586e7774",
		},
		{
			name:    "producao com subpath",
			baseURL: "https://admin.exemplo.com/painel",
			orderID: "6b0913ed4c6162f4586e7774",
			want:    "https://admin.exemplo.com/painel/admin/pedidos/6b0913ed4c6162f4586e7774",
		},
		{name: "base vazia", baseURL: "", orderID: "6b0913ed4c6162f4586e7774", want: ""},
		{name: "id publico nao serve", baseURL: testAdminBaseURL, orderID: "AQAZ922yT5jN8RPCva248g", want: ""},
		{name: "id vazio", baseURL: testAdminBaseURL, orderID: "", want: ""},
		{name: "base invalida", baseURL: "localhost:5173", orderID: "6b0913ed4c6162f4586e7774", want: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := adminOrderURL(test.baseURL, test.orderID); got != test.want {
				t.Fatalf("adminOrderURL = %q, want %q", got, test.want)
			}
		})
	}
}

func TestBuildWebhookMessageOmitsLinkWithoutConfiguredBase(t *testing.T) {
	message := BuildWebhookMessage(domain.WebhookEvent{
		Type:  domain.WebhookEventOrderCreated,
		Order: domain.Order{ID: "6b0913ed4c6162f4586e7774", Number: 1042},
	}, "")
	if strings.Contains(message, "/admin/pedidos/") {
		t.Fatalf("link inesperado sem base configurada: %q", message)
	}
}

func TestNewWebhookMessageBuilderUsesFixedBase(t *testing.T) {
	builder := NewWebhookMessageBuilder("https://admin.exemplo.com")
	message := builder(domain.WebhookEvent{
		Type:  domain.WebhookEventOrderCreated,
		Order: domain.Order{ID: "6b0913ed4c6162f4586e7774", Number: 1042},
	})
	if !strings.Contains(message, "https://admin.exemplo.com/admin/pedidos/6b0913ed4c6162f4586e7774") {
		t.Fatalf("construtor não usou a base configurada: %q", message)
	}
}
