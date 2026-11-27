// Package realtime distribui sinais de mudança de pedido para conexões
// WebSocket. Os eventos carregam apenas identificador opaco, número, status e
// timestamp: nenhum dado pessoal, valor monetário ou link de WhatsApp trafega
// por este canal.
package realtime

import (
	"sync"
)

// Tipos de evento publicados pelo servidor.
const (
	EventConnected    = "connected"
	EventOrderCreated = "order.created"
	EventOrderUpdated = "order.updated"
	EventOrderDeleted = "order.deleted"
)

// TopicAdmin recebe os sinais de todos os pedidos; o tópico público recebe
// apenas os sinais de um publicId.
const TopicAdmin = "admin"

const (
	eventBufferSize      = 8
	maxAdminSubscribers  = 200
	maxPublicSubscribers = 10
	maxTopics            = 2000
)

// TopicForPublic devolve o tópico restrito a um pedido público.
func TopicForPublic(publicID string) string {
	return "public:" + publicID
}

// Event é o sinal mínimo enviado ao cliente. Campos ausentes são omitidos para
// que o canal público nunca devolva o ID interno do pedido.
type Event struct {
	Type        string `json:"type"`
	OrderID     string `json:"orderId,omitempty"`
	PublicID    string `json:"publicId,omitempty"`
	OrderNumber int64  `json:"orderNumber,omitempty"`
	Status      string `json:"status,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// Subscription é a assinatura de um cliente. Close é idempotente.
type Subscription struct {
	Events chan Event

	hub   *Hub
	topic string
	once  sync.Once
}

// Close remove a assinatura do hub e fecha o canal.
func (s *Subscription) Close() {
	s.once.Do(func() {
		s.hub.unsubscribe(s.topic, s.Events)
	})
}

// Hub é um barramento em memória, sem persistência e sem estado compartilhado
// entre instâncias. Em uma implantação com mais de uma réplica é necessário um
// barramento externo; isso está registrado em docs/04-arquitetura.md.
type Hub struct {
	mu     sync.Mutex
	topics map[string]map[chan Event]struct{}
	total  int
}

// NewHub cria um hub vazio.
func NewHub() *Hub {
	return &Hub{topics: make(map[string]map[chan Event]struct{})}
}

// Subscribe registra um canal no tópico. O segundo valor é false quando o
// limite de assinaturas foi atingido.
func (h *Hub) Subscribe(topic string) (*Subscription, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	limit := maxPublicSubscribers
	if topic == TopicAdmin {
		limit = maxAdminSubscribers
	}
	subscribers := h.topics[topic]
	if len(subscribers) >= limit || (len(h.topics) >= maxTopics && subscribers == nil) {
		return nil, false
	}
	if subscribers == nil {
		subscribers = make(map[chan Event]struct{})
		h.topics[topic] = subscribers
	}

	events := make(chan Event, eventBufferSize)
	subscribers[events] = struct{}{}
	h.total++
	return &Subscription{hub: h, topic: topic, Events: events}, true
}

// Publish entrega o evento a todas as assinaturas do tópico. O envio é não
// bloqueante: quando o canal está cheio, o evento mais antigo é descartado.
func (h *Hub) Publish(topic string, event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for events := range h.topics[topic] {
		select {
		case events <- event:
		default:
			select {
			case <-events:
			default:
			}
			select {
			case events <- event:
			default:
			}
		}
	}
}

func (h *Hub) unsubscribe(topic string, events chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subscribers, ok := h.topics[topic]
	if !ok {
		return
	}
	if _, ok := subscribers[events]; !ok {
		return
	}
	delete(subscribers, events)
	h.total--
	if len(subscribers) == 0 {
		delete(h.topics, topic)
	}
	close(events)
}

// Close encerra todas as assinaturas. Chamado no shutdown do processo.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for topic, subscribers := range h.topics {
		for events := range subscribers {
			delete(subscribers, events)
			close(events)
			h.total--
		}
		delete(h.topics, topic)
	}
}
