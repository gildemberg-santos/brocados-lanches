// Package webhook entrega notificações HTTP para integrações externas. O envio
// é best-effort e assíncrono: falha de rede ou resposta 5xx nunca altera o
// resultado da criação ou da transição do pedido.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

const (
	defaultTimeout    = 3 * time.Second
	maxResponseBytes  = 1024
	eventHeader       = "X-Brocados-Event"
	contentTypeHeader = "application/json"
	userAgent         = "Brocados-Lanches-Webhook/1.0"
)

var (
	errSettingsRequired = errors.New("webhook settings reader is required")
	errMessageRequired  = errors.New("webhook message builder is required")
)

// Payload é o corpo exato exigido pelo contrato de integração. Nenhum campo
// adicional é enviado.
type Payload struct {
	Number  string `json:"number"`
	Message string `json:"message"`
}

// SettingsReader é a parcela do serviço administrativo necessária ao disparo.
type SettingsReader interface {
	Get(ctx context.Context) (domain.WebhookSettings, error)
}

// MessageBuilder monta o texto a partir do pedido persistido.
type MessageBuilder func(domain.WebhookEvent) string

// Notifier publica eventos de pedido. Notify devolve imediatamente ao handler.
type Notifier struct {
	settings SettingsReader
	client   *http.Client
	logger   *slog.Logger
	message  MessageBuilder
	timeout  time.Duration
}

// NewNotifier valida as dependências. Sem leitor de configuração ou construtor
// de mensagem, o notificador não é criado; sem logger, usa o padrão do processo.
func NewNotifier(settings SettingsReader, logger *slog.Logger, message MessageBuilder) (*Notifier, error) {
	if settings == nil {
		return nil, errSettingsRequired
	}
	if message == nil {
		return nil, errMessageRequired
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Notifier{
		settings: settings,
		// O timeout é aplicado pelo contexto em deliver. CheckRedirect impede
		// que um 307/308 repita o POST com PII em destino não configurado.
		client: &http.Client{
			Timeout: 0,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		logger:  logger,
		message: message,
		timeout: defaultTimeout,
	}, nil
}

// Notify agenda o envio fora do contexto da requisição, que é cancelado logo
// após a resposta HTTP. Nenhum dado pessoal é registrado em log.
func (n *Notifier) Notify(ctx context.Context, event domain.WebhookEvent) {
	if n == nil {
		return
	}
	base := context.WithoutCancel(ctx)
	go n.deliver(base, event)
}

// deliver executa uma tentativa. É separado de Notify para permitir teste
// determinístico sem depender de agendamento de goroutine.
func (n *Notifier) deliver(base context.Context, event domain.WebhookEvent) {
	ctx, cancel := context.WithTimeout(base, n.timeout)
	defer cancel()

	settings, err := n.settings.Get(ctx)
	if err != nil {
		n.logger.Warn("webhook skipped: settings unavailable")
		return
	}
	if !settings.Active || strings.TrimSpace(settings.URL) == "" {
		return
	}

	body, err := json.Marshal(Payload{
		Number:  event.Order.Customer.Phone,
		Message: n.message(event),
	})
	if err != nil {
		n.logger.Warn("webhook skipped: payload encoding failed")
		return
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, settings.URL, bytes.NewReader(body))
	if err != nil {
		// O erro do http inclui a URL; por isso só a categoria é registrada.
		n.logger.Warn("webhook skipped: invalid configured URL")
		return
	}
	request.Header.Set("Content-Type", contentTypeHeader)
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set(eventHeader, event.Type)

	response, err := n.client.Do(request)
	if err != nil {
		n.logger.Warn("webhook delivery failed", "reason", failureReason(err))
		return
	}
	defer func() { _ = response.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))
	// Redirecionamentos não são seguidos; antes de 400, ainda indicam que o
	// destino configurado não aceitou a notificação.
	if response.StatusCode >= http.StatusMultipleChoices {
		n.logger.Warn("webhook delivery failed", "status", response.StatusCode)
		return
	}
	n.logger.Info("webhook delivered", "status", response.StatusCode, "event", event.Type)
}

func failureReason(err error) string {
	if err == nil {
		return "unknown"
	}
	if isTimeout(err) {
		return "timeout"
	}
	return "transport"
}

func isTimeout(err error) bool {
	type timeout interface{ Timeout() bool }
	if checker, ok := err.(timeout); ok && checker.Timeout() {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded)
}
