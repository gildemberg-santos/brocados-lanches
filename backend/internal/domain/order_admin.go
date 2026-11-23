package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

// AdminOrderListQuery descreve os filtros da fila administrativa. A ordenação
// é conferida contra allowlist no service; o repositório nunca recebe um campo
// arbitrário vindo da requisição.
type AdminOrderListQuery struct {
	Page        int
	PageSize    int
	Sort        string
	Order       string
	Status      string
	Fulfillment Fulfillment
	From        *time.Time
	To          *time.Time
	Search      string
}

// AdminOrderListItem é a visão da fila. Os dados do cliente já chegam mascarados
// ao handler, para que a listagem nunca exponha PII completa por acidente.
type AdminOrderListItem struct {
	ID               string
	Number           int64
	PublicID         string
	Status           string
	Fulfillment      Fulfillment
	CustomerName     string
	Phone            string
	ItemCount        int
	SubtotalCents    int64
	DeliveryFeeCents int64
	TotalCents       int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewAdminOrderListItem converte um pedido para a visão mascarada da fila.
func NewAdminOrderListItem(order Order) AdminOrderListItem {
	return AdminOrderListItem{
		ID:               order.ID,
		Number:           order.Number,
		PublicID:         order.PublicID,
		Status:           order.Status,
		Fulfillment:      order.Fulfillment,
		CustomerName:     MaskName(order.Customer.Name),
		Phone:            MaskPhone(order.Customer.Phone),
		ItemCount:        len(order.Items),
		SubtotalCents:    order.SubtotalCents,
		DeliveryFeeCents: order.DeliveryFeeCents,
		TotalCents:       order.TotalCents,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}
}

// MaskPhone mantém apenas os últimos quatro dígitos, permitindo conferir um
// pedido sem expor o telefone completo na fila.
func MaskPhone(phone string) string {
	visible := 0
	for index := len(phone) - 1; index >= 0; index-- {
		if phone[index] < '0' || phone[index] > '9' {
			continue
		}
		visible++
		if visible == 4 {
			return "****" + phone[index:]
		}
	}
	return strings.Repeat("*", len(phone))
}

// MaskName preserva a inicial de cada parte do nome; o detalhe autorizado é
// o único lugar que mostra o nome completo.
func MaskName(name string) string {
	parts := strings.Fields(name)
	for index, part := range parts {
		letter, _ := utf8.DecodeRuneInString(part)
		parts[index] = string(letter) + "***"
	}
	return strings.Join(parts, " ")
}
