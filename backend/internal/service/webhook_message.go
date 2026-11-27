package service

import (
	"fmt"

	"github.com/brocados-lanches/api/internal/domain"
)

// BuildWebhookMessage reaproveita o snapshot congelado do pedido e acrescenta
// o status quando o evento é uma transição. Status futuros entram sem lista
// fixa: o rótulo cai para o próprio identificador persistido.
func BuildWebhookMessage(event domain.WebhookEvent) string {
	message := BuildOrderMessage(event.Order)
	if event.Type == domain.WebhookEventOrderUpdated {
		message += fmt.Sprintf("\n✅ Status: %s\n", statusLabel(event.Order.Status))
	}
	return message
}

func statusLabel(status string) string {
	labels := map[string]string{
		domain.OrderStatusPending:        "Pendente",
		domain.OrderStatusConfirmed:      "Confirmado",
		domain.OrderStatusPreparing:      "Em preparação",
		domain.OrderStatusReady:          "Pronto",
		domain.OrderStatusOutForDelivery: "Saiu para entrega",
		domain.OrderStatusPickedUp:       "Retirado",
		domain.OrderStatusDelivered:      "Entregue",
		domain.OrderStatusCancelled:      "Cancelado",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return status
}
