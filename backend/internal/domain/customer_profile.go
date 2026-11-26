package domain

import (
	"strings"
	"time"
)

// CustomerProfileRetentionMonths é o prazo proposto para o perfil de cliente,
// alinhado ao prazo dos pedidos em docs/10-operacao-e-lancamento.md.
const CustomerProfileRetentionMonths = 24

// CustomerProfile reúne, por telefone, os dados já informados pelo cliente para
// agilizar o próximo pedido. Não é uma API pública e nunca contém observação,
// valor de troco, senha ou token.
type CustomerProfile struct {
	ID                string
	PhoneKey          string
	Phone             string
	Name              string
	Address           *OrderAddress
	Fulfillment       Fulfillment
	Zone              *OrderZone
	PaymentMethod     string
	OrderCount        int
	LastOrderPublicID string
	LastOrderAt       time.Time
	OptedOut          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ExpiresAt         time.Time
}

// NormalizePhoneKey devolve a forma canônica usada para reconhecer o cliente
// por telefone. Números locais brasileiros recebem o prefixo 55; o valor
// guardado é apenas dígitos, sem "+", espaços ou pontuação.
func NormalizePhoneKey(phone string) string {
	var builder strings.Builder
	for _, character := range phone {
		if character >= '0' && character <= '9' {
			builder.WriteRune(character)
		}
	}
	digits := builder.String()
	switch {
	case len(digits) == 10 || len(digits) == 11:
		return "55" + digits
	case (len(digits) == 12 || len(digits) == 13) && strings.HasPrefix(digits, "55"):
		return digits
	default:
		return digits
	}
}

// NewCustomerProfileFromOrder congela no perfil apenas o que ajuda no próximo
// pedido. Observação e troco ficam de fora mesmo quando presentes no pedido.
func NewCustomerProfileFromOrder(order Order, now time.Time) (CustomerProfile, error) {
	phoneKey := NormalizePhoneKey(order.Customer.Phone)
	if phoneKey == "" {
		return CustomerProfile{}, invalid("customer.phone", CodeInvalid, "unsupported phone format")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	lastOrderAt := order.CreatedAt.UTC()
	if lastOrderAt.IsZero() {
		lastOrderAt = now
	}

	profile := CustomerProfile{
		PhoneKey:          phoneKey,
		Phone:             order.Customer.Phone,
		Name:              order.Customer.Name,
		Fulfillment:       order.Fulfillment,
		PaymentMethod:     order.Payment.Method,
		OrderCount:        1,
		LastOrderPublicID: order.PublicID,
		LastOrderAt:       lastOrderAt,
		CreatedAt:         now.UTC(),
		UpdatedAt:         now.UTC(),
		ExpiresAt:         now.UTC().AddDate(0, CustomerProfileRetentionMonths, 0),
	}
	if order.Address != nil {
		address := *order.Address
		profile.Address = &address
	}
	if order.Zone != nil {
		zone := *order.Zone
		profile.Zone = &zone
	}
	return profile, nil
}

// Validate protege a chave do perfil contra valores vazios vindos de dados
// antigos ou de chamadas internas incorretas.
func (p CustomerProfile) Validate() error {
	if p.PhoneKey == "" {
		return invalid("phoneKey", CodeRequired, "is required")
	}
	if p.PhoneKey != NormalizePhoneKey(p.PhoneKey) {
		return invalid("phoneKey", CodeInvalid, "must be normalized digits")
	}
	return nil
}
