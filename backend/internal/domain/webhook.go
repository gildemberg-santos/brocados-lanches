package domain

import (
	"net/url"
	"strings"
	"time"
)

// WebhookEvent types are stable identifiers sent in the X-Brocados-Event header.
// O disparo é genérico: qualquer status persistido, inclusive um futuro status
// personalizado, usa EventOrderUpdated sem depender de uma lista fixa.
const (
	WebhookEventOrderCreated = "order.created"
	WebhookEventOrderUpdated = "order.updated"
)

// MaxWebhookURLLength limita o tamanho do endpoint configurado sem impor
// allowlist de host: a loja pode apontar para n8n, Make, Zapier ou API própria.
const MaxWebhookURLLength = 2048

// WebhookSettings é a configuração de notificações externas da loja. A URL
// carrega telefone do cliente no corpo do pedido, por isso o painel que a edita
// é restrito a owner e a auditoria guarda apenas se a URL está preenchida.
type WebhookSettings struct {
	URL       string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// WebhookEvent é o sinal mínimo entregue ao notificador. O pedido completo é
// necessário para montar a mensagem congelada, mas nunca entra em log.
type WebhookEvent struct {
	Type  string
	Order Order
}

// WebhookNotification reúne, em uma única leitura do documento da loja, a
// configuração do webhook e o WhatsApp usado como destinatário interno (D-046).
type WebhookNotification struct {
	Settings    WebhookSettings
	StoreNumber string
}

// Validate confere o endpoint antes de gravar. Qualquer esquema absoluto é
// aceito, sem allowlist de host ou de IP; esquemas não HTTP falham no envio
// best-effort sem bloquear o pedido.
func (s WebhookSettings) Validate() error {
	value := strings.TrimSpace(s.URL)
	if s.Active && value == "" {
		return invalid("url", CodeRequired, "is required when webhook is active")
	}
	if value == "" {
		return nil
	}
	if err := validateMaxLength("url", value, MaxWebhookURLLength); err != nil {
		return err
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" {
		return invalid("url", CodeInvalid, "must be an absolute URL with a scheme")
	}
	return nil
}

// WebhookSettingsPatch contém apenas os campos editáveis pelo painel.
type WebhookSettingsPatch struct {
	URL    *string
	Active *bool
}

// Empty indica que o PATCH não trouxe nenhum campo conhecido.
func (p WebhookSettingsPatch) Empty() bool {
	return p.URL == nil && p.Active == nil
}

// Apply devolve a configuração resultante sem alterar a atual.
func (p WebhookSettingsPatch) Apply(current WebhookSettings) WebhookSettings {
	updated := current
	if p.URL != nil {
		updated.URL = strings.TrimSpace(*p.URL)
	}
	if p.Active != nil {
		updated.Active = *p.Active
	}
	return updated
}
