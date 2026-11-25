package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Status do pedido. Os estados finais não transitam.
const (
	OrderStatusPending        = "pending"
	OrderStatusConfirmed      = "confirmed"
	OrderStatusPreparing      = "preparing"
	OrderStatusReady          = "ready"
	OrderStatusOutForDelivery = "out_for_delivery"
	OrderStatusPickedUp       = "picked_up"
	OrderStatusDelivered      = "delivered"
	OrderStatusCancelled      = "cancelled"
)

// Formas de pagamento aceitas na V1, sempre na entrega ou retirada.
const (
	PaymentMethodCash = "cash"
	PaymentMethodCard = "card"
	PaymentMethodPix  = "pix"
)

// Limites de entrada do pedido. Não são dados de negócio.
const (
	MinIdempotencyKeyLength = 16
	MaxIdempotencyKeyLength = 128
	MaxCustomerNameLength   = 120
	MaxPhoneLength          = 20
	MaxAddressFieldLength   = 120
	MaxZipCodeLength        = 12
	MaxOrderNoteLength      = 280
	MaxChangeForCents       = 10000000 // R$ 100.000,00
)

// PhonePattern aceita números com ou sem "+", espaços, parênteses e hífen.
var phonePattern = regexp.MustCompile(`^\+?[0-9][0-9()\s-]{7,23}$`)

// OrderRequest contém apenas escolhas e dados necessários à execução do pedido.
// Preço, taxa, total e disponibilidade nunca vêm do cliente.
type OrderRequest struct {
	Items       []QuoteItem   `json:"items"`
	Fulfillment Fulfillment   `json:"fulfillment"`
	ZoneID      string        `json:"zoneId,omitempty"`
	Customer    OrderCustomer `json:"customer"`
	Address     *OrderAddress `json:"address,omitempty"`
	Payment     OrderPayment  `json:"payment"`
	Note        string        `json:"note,omitempty"`
}

// OrderCustomer são dados pessoais mínimos para executar o pedido.
type OrderCustomer struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// OrderAddress é obrigatório em delivery e ignorado em pickup.
type OrderAddress struct {
	Street     string `json:"street"`
	Number     string `json:"number"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district"`
	City       string `json:"city"`
	State      string `json:"state"`
	ZipCode    string `json:"zipCode"`
	Reference  string `json:"reference,omitempty"`
}

// OrderPayment não aceita dado de cartão. Troco só é válido em dinheiro.
type OrderPayment struct {
	Method         string `json:"method"`
	ChangeForCents *int64 `json:"changeForCents,omitempty"`
}

// Order é o pedido persistido, com valores congelados no momento da criação.
type Order struct {
	ID               string
	Number           int64
	PublicID         string
	Status           string
	Fulfillment      Fulfillment
	Customer         OrderCustomer
	Address          *OrderAddress
	Zone             *OrderZone
	Items            []OrderItem
	SubtotalCents    int64
	DeliveryFeeCents int64
	TotalCents       int64
	Payment          OrderPayment
	Note             string
	IdempotencyKey   string
	RequestHash      string
	StatusHistory    []OrderStatusChange
	ConfirmedAt      *time.Time
	CancelledAt      *time.Time
	CancelReason     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// OrderZone é o snapshot da zona no momento do pedido.
type OrderZone struct {
	ID       string
	Name     string
	FeeCents int64
}

// OrderItem é o snapshot de uma linha do pedido.
type OrderItem struct {
	ProductID      string
	Name           string
	UnitPriceCents int64
	Quantity       int
	Options        []OrderItemOption
	Note           string
	LineTotalCents int64
}

// OrderItemOption congela o nome e o preço de uma opção ou adicional.
type OrderItemOption struct {
	GroupID    string
	GroupName  string
	OptionID   string
	Name       string
	Kind       string
	PriceCents int64
}

// OrderStatusChange registra uma transição de status.
type OrderStatusChange struct {
	Status  string
	At      time.Time
	ActorID string
}

// Validate confere a forma da requisição antes de qualquer leitura do banco.
func (r OrderRequest) Validate() error {
	quoteRequest := QuoteRequest{Items: r.Items, Fulfillment: r.Fulfillment, ZoneID: r.ZoneID}
	if err := quoteRequest.Validate(); err != nil {
		return err
	}

	if err := validateRequiredString("customer.name", r.Customer.Name, MaxCustomerNameLength); err != nil {
		return err
	}
	if err := validatePhone(r.Customer.Phone); err != nil {
		return err
	}
	if err := validateMaxLength("note", r.Note, MaxOrderNoteLength); err != nil {
		return err
	}

	if r.Fulfillment == FulfillmentDelivery {
		if r.Address == nil {
			return invalid("address", CodeRequired, "is required for delivery")
		}
		if err := r.Address.Validate(); err != nil {
			return prefixValidationError("address", err)
		}
	}

	switch r.Payment.Method {
	case PaymentMethodCash, PaymentMethodCard, PaymentMethodPix:
	default:
		return invalid("payment.method", CodeInvalid, "must be cash, card or pix")
	}
	if r.Payment.Method != PaymentMethodCash && r.Payment.ChangeForCents != nil {
		return invalid("payment.changeForCents", CodeInvalid, "is only valid for cash payments")
	}

	return nil
}

// ValidateWithTotal aplica as regras que dependem do total calculado no servidor.
func (r OrderRequest) ValidateWithTotal(totalCents int64) error {
	if r.Payment.Method != PaymentMethodCash || r.Payment.ChangeForCents == nil {
		return nil
	}
	if *r.Payment.ChangeForCents < totalCents {
		return invalid("payment.changeForCents", CodeOutOfRange, "must be at least the order total")
	}
	if *r.Payment.ChangeForCents > MaxChangeForCents || *r.Payment.ChangeForCents > math.MaxInt64-totalCents {
		return invalid("payment.changeForCents", CodeOutOfRange, "exceeds the supported limit")
	}
	return nil
}

// Validate confere os campos do endereço de entrega.
func (a OrderAddress) Validate() error {
	fields := []struct {
		name  string
		value string
	}{
		{"street", a.Street},
		{"number", a.Number},
		{"district", a.District},
		{"city", a.City},
		{"state", a.State},
		{"zipCode", a.ZipCode},
	}
	for _, field := range fields {
		limit := MaxAddressFieldLength
		if field.name == "state" {
			limit = 2
		}
		if field.name == "zipCode" {
			limit = MaxZipCodeLength
		}
		if err := validateRequiredString(field.name, field.value, limit); err != nil {
			return err
		}
	}
	if err := validateMaxLength("complement", a.Complement, MaxAddressFieldLength); err != nil {
		return err
	}
	return validateMaxLength("reference", a.Reference, MaxAddressFieldLength)
}

// NewOrder congela os valores da cotação em um pedido pendente.
func NewOrder(request OrderRequest, quote Quote, zone *OrderZone, id string, number int64, publicID, idempotencyKey, requestHash string, now time.Time) Order {
	items := make([]OrderItem, 0, len(quote.Items))
	for index, line := range quote.Items {
		options := make([]OrderItemOption, 0, len(line.Options))
		for _, option := range line.Options {
			options = append(options, OrderItemOption(option))
		}
		items = append(items, OrderItem{
			ProductID:      line.ProductID,
			Name:           line.Name,
			UnitPriceCents: line.UnitPriceCents,
			Quantity:       line.Quantity,
			Options:        options,
			Note:           request.Items[index].Note,
			LineTotalCents: line.LineTotalCents,
		})
	}

	return Order{
		ID:               id,
		Number:           number,
		PublicID:         publicID,
		Status:           OrderStatusPending,
		Fulfillment:      request.Fulfillment,
		Customer:         request.Customer,
		Address:          request.Address,
		Zone:             zone,
		Items:            items,
		SubtotalCents:    quote.SubtotalCents,
		DeliveryFeeCents: quote.DeliveryFeeCents,
		TotalCents:       quote.TotalCents,
		Payment:          request.Payment,
		Note:             request.Note,
		IdempotencyKey:   idempotencyKey,
		RequestHash:      requestHash,
		StatusHistory: []OrderStatusChange{{
			Status: OrderStatusPending,
			At:     now.UTC(),
		}},
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}
}

// CanonicalRequestHash gera o hash do corpo usado na idempotência.
func CanonicalRequestHash(request OrderRequest) (string, error) {
	payload, err := json.Marshal(struct {
		Items       []QuoteItem   `json:"items"`
		Fulfillment Fulfillment   `json:"fulfillment"`
		ZoneID      string        `json:"zoneId"`
		Customer    OrderCustomer `json:"customer"`
		Address     *OrderAddress `json:"address"`
		Payment     OrderPayment  `json:"payment"`
		Note        string        `json:"note"`
	}{
		Items:       request.Items,
		Fulfillment: request.Fulfillment,
		ZoneID:      request.ZoneID,
		Customer:    request.Customer,
		Address:     request.Address,
		Payment:     request.Payment,
		Note:        request.Note,
	})
	if err != nil {
		return "", fmt.Errorf("marshal order request: %w", err)
	}
	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:]), nil
}

func validatePhone(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return invalid("customer.phone", CodeRequired, "is required")
	}
	if utf8.RuneCountInString(trimmed) > MaxPhoneLength || !phonePattern.MatchString(trimmed) {
		return invalid("customer.phone", CodeInvalid, "must be a valid phone number")
	}
	return nil
}
