package service

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	"github.com/brocados-lanches/api/internal/domain"
)

// BuildWebhookMessage reaproveita o snapshot congelado do pedido, acrescenta o
// status quando o evento é uma transição e inclui o link do painel quando a base
// foi configurada. Status futuros entram sem lista fixa: o rótulo cai para o
// próprio identificador persistido.
func BuildWebhookMessage(event domain.WebhookEvent, adminBaseURL string) string {
	message := BuildOrderMessage(event.Order)
	if event.Type == domain.WebhookEventOrderUpdated {
		message += fmt.Sprintf("\n✅ Status: %s\n", statusLabel(event.Order.Status))
	}
	if link := adminOrderURL(adminBaseURL, event.Order.ID); link != "" {
		message += fmt.Sprintf("\n🔗 Ver pedido no painel: %s\n", link)
	}
	return message
}

// NewWebhookMessageBuilder fixa a base do painel na composição da aplicação e
// devolve o construtor exigido pelo notificador.
func NewWebhookMessageBuilder(adminBaseURL string) func(domain.WebhookEvent) string {
	return func(event domain.WebhookEvent) string {
		return BuildWebhookMessage(event, adminBaseURL)
	}
}

// adminOrderURL monta o link administrativo com o ID hexadecimal do pedido.
// Base vazia (produção sem domínio definido) ou ID inválido omitem o link: o
// webhook continua válido e nenhuma URL quebrada é enviada ao destino.
func adminOrderURL(baseURL, orderID string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" || !isObjectIDHex(orderID) {
		return ""
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	path := strings.TrimRight(parsed.Path, "/")
	return parsed.Scheme + "://" + parsed.Host + path + "/admin/pedidos/" + orderID
}

func isObjectIDHex(id string) bool {
	if len(id) != 24 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
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
