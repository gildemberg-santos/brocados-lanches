package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/realtime"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
)

// O servidor HTTP usa timeouts de 15 s. Um WebSocket é de longa duração, então
// as deadlines são removidas depois do upgrade e passam a ser controladas pelo
// contexto de cada escrita.
const wsWriteTimeout = 5 * time.Second

// publicIDPattern evita caracteres especiais no tópico e aceita o identificador
// opaco de 128 bits em base64url (22 caracteres), com folga para versões futuras.
var publicIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{16,64}$`)

// handlePublicOrderEvents expõe apenas mudanças de status de um publicId.
// Não exige sessão e nunca devolve dados pessoais, valores ou whatsappUrl.
func handlePublicOrderEvents(hub *realtime.Hub, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		publicID := chi.URLParam(r, "publicId")
		if !publicIDPattern.MatchString(publicID) {
			writeError(w, r, logger, domain.ErrNotFound)
			return
		}
		serveOrderEvents(w, r, hub, realtime.TopicForPublic(publicID), "public", logger)
	}
}

// handleAdminOrderEvents exige sessão válida: a rota é registrada dentro do
// grupo protegido do painel, com CSRF e autorização por papel no servidor.
func handleAdminOrderEvents(hub *realtime.Hub, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveOrderEvents(w, r, hub, realtime.TopicAdmin, "admin", logger)
	}
}

func serveOrderEvents(w http.ResponseWriter, r *http.Request, hub *realtime.Hub, topic, kind string, logger *slog.Logger) {
	if hub == nil {
		writeError(w, r, logger, errors.New("realtime hub is not configured"))
		return
	}

	controller := http.NewResponseController(w)
	_ = controller.SetReadDeadline(time.Time{})
	_ = controller.SetWriteDeadline(time.Time{})

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		logger.Debug("websocket handshake failed", "channel", kind, "error", err)
		return
	}
	defer func() { _ = conn.CloseNow() }()

	rootContext, cancelRoot := context.WithCancel(context.Background())
	defer cancelRoot()
	// CloseRead cancela o contexto quando o cliente fecha e responde pings.
	ctx := conn.CloseRead(rootContext)

	subscription, ok := hub.Subscribe(topic)
	if !ok {
		logger.Warn("websocket subscription limit reached", "channel", kind)
		_ = conn.Close(websocket.StatusPolicyViolation, "limite de conexões atingido")
		return
	}
	defer subscription.Close()

	if err := writeWSEvent(ctx, conn, realtime.Event{Type: realtime.EventConnected}); err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, open := <-subscription.Events:
			if !open {
				return
			}
			if err := writeWSEvent(ctx, conn, event); err != nil {
				return
			}
		}
	}
}

func writeWSEvent(ctx context.Context, conn *websocket.Conn, event realtime.Event) error {
	writeContext, cancel := context.WithTimeout(ctx, wsWriteTimeout)
	defer cancel()
	return wsjson.Write(writeContext, conn, event)
}

// publishOrderEvent converte o pedido do domínio para o evento mínimo. O tipo
// distingue criação de transição para o painel reorganizar a fila sem precisar
// de polling.
func publishOrderEvent(hub *realtime.Hub, eventType string, order domain.Order) {
	if hub == nil {
		return
	}
	event := realtime.Event{
		Type:        eventType,
		OrderID:     order.ID,
		PublicID:    order.PublicID,
		OrderNumber: order.Number,
		Status:      order.Status,
		UpdatedAt:   order.UpdatedAt.UTC().Format(time.RFC3339),
	}
	hub.Publish(realtime.TopicAdmin, event)
	publicEvent := event
	publicEvent.OrderID = ""
	hub.Publish(realtime.TopicForPublic(order.PublicID), publicEvent)
}
