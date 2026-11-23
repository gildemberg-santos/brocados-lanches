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

	var message strings.Builder
	fmt.Fprintf(&message, "Pedido #%d\n", order.Number)
	message.WriteString("Itens:\n")
	for _, item := range order.Items {
		fmt.Fprintf(&message, "- %dx %s", item.Quantity, item.Name)
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
		fmt.Fprintf(&message, "Observação: %s\n", order.Note)
	}
	fmt.Fprintf(&message, "Subtotal: R$ %s\n", formatCents(order.SubtotalCents))
	fmt.Fprintf(&message, "Taxa de entrega: R$ %s\n", formatCents(order.DeliveryFeeCents))
	fmt.Fprintf(&message, "Total: R$ %s\n", formatCents(order.TotalCents))
	fmt.Fprintf(&message, "Pagamento: %s\n", paymentLabel(order.Payment))
	fmt.Fprintf(&message, "Cliente: %s — %s\n", order.Customer.Name, order.Customer.Phone)
	if order.Fulfillment == domain.FulfillmentDelivery && order.Address != nil {
		fmt.Fprintf(&message, "Entrega: %s, %s — %s, %s/%s — CEP %s\n",
			order.Address.Street,
			order.Address.Number,
			order.Address.District,
			order.Address.City,
			order.Address.State,
			order.Address.ZipCode,
		)
		if order.Address.Complement != "" {
			fmt.Fprintf(&message, "Complemento: %s\n", order.Address.Complement)
		}
	} else {
		message.WriteString("Retirada na loja\n")
	}

	return "https://wa.me/" + number + "?text=" + url.QueryEscape(message.String()), nil
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
		return "Pix na entrega"
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
	return fmt.Sprintf("%d.%02d", reais, resto)
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
