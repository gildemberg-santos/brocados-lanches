package service

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/brocados-lanches/api/internal/domain"
)

// BuildWhatsAppURL gera o handoff depois de o pedido estar persistido. A URL
// contém dados pessoais e nunca deve ser registrada em log.
func BuildWhatsAppURL(store domain.StoreSettings, order domain.Order) (string, error) {
	number := digitsOnly(store.WhatsApp)
	if len(number) == 10 || len(number) == 11 {
		number = "55" + number
	}
	if len(number) < 12 || len(number) > 13 {
		return "", fmt.Errorf("invalid store whatsapp number")
	}

	return "https://wa.me/" + number + "?text=" + url.QueryEscape(BuildOrderMessage(order)), nil
}

// BuildOrderMessage monta o texto congelado do pedido. É usado no handoff do
// WhatsApp e no corpo dos webhooks, sem depender de catálogo atual.
func BuildOrderMessage(order domain.Order) string {
	var message strings.Builder
	fmt.Fprintf(&message, "*Pedido #%d* 🍔\n\n", order.Number)

	message.WriteString("*Itens* 🛒\n")
	for _, item := range order.Items {
		fmt.Fprintf(&message, "• %dx %s", item.Quantity, item.Name)
		if len(item.Options) > 0 {
			message.WriteString(" (")
			for index, option := range item.Options {
				if index > 0 {
					message.WriteString(", ")
				}
				message.WriteString(option.Name)
			}
			message.WriteString(")")
		}
		if item.Note != "" {
			fmt.Fprintf(&message, " — %s", item.Note)
		}
		message.WriteString("\n")
	}
	if order.Note != "" {
		fmt.Fprintf(&message, "\n*Observação* 📝\n%s\n", order.Note)
	}

	fmt.Fprintf(&message, "\n*Resumo* 💰\n")
	fmt.Fprintf(&message, "Subtotal: R$ %s\n", formatCents(order.SubtotalCents))
	fmt.Fprintf(&message, "Taxa de entrega: R$ %s\n", formatCents(order.DeliveryFeeCents))
	fmt.Fprintf(&message, "*Total: R$ %s*\n", formatCents(order.TotalCents))
	fmt.Fprintf(&message, "\nPagamento: %s\n", paymentLabel(order.Payment))

	fmt.Fprintf(&message, "\n*Cliente* 👤\n%s\n%s\n",
		order.Customer.Name,
		formatPhoneForMessage(order.Customer.Phone),
	)

	if order.Fulfillment == domain.FulfillmentDelivery && order.Address != nil {
		fmt.Fprintf(&message, "\n*Entrega* 📍\n%s, %s\n", order.Address.Street, order.Address.Number)
		if order.Address.Complement != "" {
			fmt.Fprintf(&message, "Complemento: %s\n", order.Address.Complement)
		}
		fmt.Fprintf(&message, "%s, %s/%s\nCEP %s\n",
			order.Address.District,
			order.Address.City,
			order.Address.State,
			formatZipCode(order.Address.ZipCode),
		)
	} else {
		fmt.Fprintf(&message, "\n*Retirada na loja* 🏪\n")
	}
	return message.String()
}

func paymentLabel(payment domain.OrderPayment) string {
	switch payment.Method {
	case domain.PaymentMethodCash:
		if payment.ChangeForCents != nil {
			return fmt.Sprintf("dinheiro, troco para R$ %s", formatCents(*payment.ChangeForCents))
		}
		return "dinheiro"
	case domain.PaymentMethodCard:
		return "cartão na entrega"
	case domain.PaymentMethodPix:
		return "Pix a pagar na confirmação (BR Code estático)"
	default:
		return payment.Method
	}
}

func formatCents(cents int64) string {
	reais := cents / 100
	resto := cents % 100
	if resto < 0 {
		resto = -resto
	}
	return fmt.Sprintf("%d,%02d", reais, resto)
}

// formatPhoneForMessage deixa o telefone congelado no pedido legível no texto.
// Qualquer formato inesperado é mantido como veio do snapshot.
func formatPhoneForMessage(value string) string {
	digits := digitsOnly(value)
	if len(digits) == 12 || len(digits) == 13 {
		digits = strings.TrimPrefix(digits, "55")
	}
	switch len(digits) {
	case 11:
		return fmt.Sprintf("+55 (%s) %s-%s", digits[:2], digits[2:7], digits[7:])
	case 10:
		return fmt.Sprintf("+55 (%s) %s-%s", digits[:2], digits[2:6], digits[6:])
	default:
		return value
	}
}

func formatZipCode(value string) string {
	digits := digitsOnly(value)
	if len(digits) == 8 {
		return digits[:5] + "-" + digits[5:]
	}
	return value
}

func digitsOnly(value string) string {
	var builder strings.Builder
	for _, character := range value {
		if unicode.IsDigit(character) {
			builder.WriteRune(character)
		}
	}
	return builder.String()
}
