package domain

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// WebhookEvent types are stable identifiers sent in the X-Brocados-Event header.
// O cabeçalho continua genérico: criação usa order.created e qualquer transição
// persistida usa order.updated, mesmo com a seleção por status do painel.
const (
	WebhookEventOrderCreated = "order.created"
	WebhookEventOrderUpdated = "order.updated"
)

// Chaves de seleção persistidas em store_settings.webhook.events. Os valores
// são estáveis: o painel permite habilitar cada um deles e o notificador filtra
// o disparo antes de qualquer envio externo.
const (
	WebhookEventKeyCreated        = WebhookEventOrderCreated
	WebhookEventKeyConfirmed      = "order." + OrderStatusConfirmed
	WebhookEventKeyPreparing      = "order." + OrderStatusPreparing
	WebhookEventKeyReady          = "order." + OrderStatusReady
	WebhookEventKeyOutForDelivery = "order." + OrderStatusOutForDelivery
	WebhookEventKeyPickedUp       = "order." + OrderStatusPickedUp
	WebhookEventKeyDelivered      = "order." + OrderStatusDelivered
	WebhookEventKeyCancelled      = "order." + OrderStatusCancelled
)

// WebhookEventKeys devolve as chaves na ordem exibida no painel. Uma função
// evita expor uma fatia global mutável.
func WebhookEventKeys() []string {
	return []string{
		WebhookEventKeyCreated,
		WebhookEventKeyConfirmed,
		WebhookEventKeyPreparing,
		WebhookEventKeyReady,
		WebhookEventKeyOutForDelivery,
		WebhookEventKeyPickedUp,
		WebhookEventKeyDelivered,
		WebhookEventKeyCancelled,
	}
}

// IsWebhookEventKey informa se o valor pode ser habilitado no painel.
func IsWebhookEventKey(key string) bool {
	switch key {
	case WebhookEventKeyCreated, WebhookEventKeyConfirmed, WebhookEventKeyPreparing,
		WebhookEventKeyReady, WebhookEventKeyOutForDelivery, WebhookEventKeyPickedUp,
		WebhookEventKeyDelivered, WebhookEventKeyCancelled:
		return true
	default:
		return false
	}
}

// MaxWebhookURLLength limita o tamanho do endpoint configurado sem impor
// allowlist de host: a loja pode apontar para n8n, Make, Zapier ou API própria.
const MaxWebhookURLLength = 2048

// WebhookSettings é a configuração de notificações externas da loja. A URL
// carrega telefone do cliente no corpo do pedido, por isso o painel que a edita
// é restrito a owner e a auditoria guarda apenas se a URL está preenchida.
type WebhookSettings struct {
	URL    string
	Active bool
	// Events são as chaves habilitadas. Nil representa a configuração gravada
	// antes da seleção por evento e equivale a todos habilitados; uma fatia
	// vazia (não nil) significa explicitamente nenhum evento habilitado.
	Events    []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// EnabledEvents devolve o conjunto efetivo de eventos. A lista devolvida é uma
// cópia, para que o chamador não altere a configuração gravada.
func (s WebhookSettings) EnabledEvents() []string {
	if s.Events == nil {
		return WebhookEventKeys()
	}
	return append([]string(nil), s.Events...)
}

// EventEnabled aplica a seleção do painel a um evento disparado. Configuração
// legada (nil) permite qualquer chave, preservando o envio genérico anterior;
// com seleção explícita, status futuros sem chave cadastrada não são enviados.
func (s WebhookSettings) EventEnabled(key string) bool {
	if s.Events == nil {
		return true
	}
	for _, enabled := range s.Events {
		if enabled == key {
			return true
		}
	}
	return false
}

// WebhookEvent é o sinal mínimo entregue ao notificador. O pedido completo é
// necessário para montar a mensagem congelada, mas nunca entra em log.
type WebhookEvent struct {
	Type  string
	Order Order
}

// Key devolve a chave selecionável do evento. A criação mantém a mesma chave do
// cabeçalho; transições usam o status final do pedido, sem alterar o cabeçalho
// genérico X-Brocados-Event.
func (e WebhookEvent) Key() string {
	if e.Type == WebhookEventOrderCreated {
		return WebhookEventKeyCreated
	}
	return "order." + e.Order.Status
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
	if err := validateWebhookEvents(s.Events); err != nil {
		return err
	}
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

func validateWebhookEvents(events []string) error {
	if len(events) > len(WebhookEventKeys()) {
		return invalid("events", CodeOutOfRange, fmt.Sprintf("must have at most %d items", len(WebhookEventKeys())))
	}
	seen := make(map[string]struct{}, len(events))
	for index, event := range events {
		field := fmt.Sprintf("events[%d]", index)
		key := strings.TrimSpace(event)
		if !IsWebhookEventKey(key) {
			return invalid(field, CodeInvalid, "must be a supported webhook event")
		}
		if _, duplicate := seen[key]; duplicate {
			return invalid(field, CodeDuplicate, "duplicates another webhook event")
		}
		seen[key] = struct{}{}
	}
	return nil
}

// WebhookSettingsPatch contém apenas os campos editáveis pelo painel. Events
// nil preserva a seleção atual; uma fatia vazia desabilita todos os eventos.
type WebhookSettingsPatch struct {
	URL    *string
	Active *bool
	Events *[]string
}

// Empty indica que o PATCH não trouxe nenhum campo conhecido.
func (p WebhookSettingsPatch) Empty() bool {
	return p.URL == nil && p.Active == nil && p.Events == nil
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
	if p.Events != nil {
		updated.Events = make([]string, 0, len(*p.Events))
		for _, event := range *p.Events {
			updated.Events = append(updated.Events, strings.TrimSpace(event))
		}
	}
	return updated
}
