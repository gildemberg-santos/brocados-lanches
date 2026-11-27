package realtime

import (
	"testing"
)

func TestHubPublishEntregaApenasAssinantesDoTopico(t *testing.T) {
	hub := NewHub()
	defer hub.Close()

	admin, ok := hub.Subscribe(TopicAdmin)
	if !ok {
		t.Fatal("Subscribe(admin) = false")
	}
	defer admin.Close()
	public, ok := hub.Subscribe(TopicForPublic("publico-1"))
	if !ok {
		t.Fatal("Subscribe(publico-1) = false")
	}
	defer public.Close()
	other, ok := hub.Subscribe(TopicForPublic("publico-2"))
	if !ok {
		t.Fatal("Subscribe(publico-2) = false")
	}
	defer other.Close()

	adminEvent := Event{Type: EventOrderUpdated, PublicID: "publico-1", Status: "ready"}
	hub.Publish(TopicAdmin, adminEvent)
	assertEvent(t, admin.Events, adminEvent)
	assertNoEvent(t, public.Events)
	assertNoEvent(t, other.Events)

	publicEvent := Event{Type: EventOrderUpdated, OrderNumber: 42, Status: "ready"}
	hub.Publish(TopicForPublic("publico-1"), publicEvent)
	assertEvent(t, public.Events, publicEvent)
	assertNoEvent(t, admin.Events)
	assertNoEvent(t, other.Events)
}

func TestHubDescartaEventoAntigoQuandoAssinanteEstaLento(t *testing.T) {
	hub := NewHub()
	defer hub.Close()

	subscription, ok := hub.Subscribe(TopicAdmin)
	if !ok {
		t.Fatal("Subscribe(admin) = false")
	}
	defer subscription.Close()

	for index := 0; index < eventBufferSize+3; index++ {
		hub.Publish(TopicAdmin, Event{Type: EventOrderUpdated, Status: "preparing"})
	}
	if len(subscription.Events) != eventBufferSize {
		t.Fatalf("eventos retidos = %d, want %d", len(subscription.Events), eventBufferSize)
	}

	for index := 0; index < eventBufferSize; index++ {
		<-subscription.Events
	}
	hub.Publish(TopicAdmin, Event{Type: EventOrderUpdated, Status: "ready"})
	last := <-subscription.Events
	if last.Status != "ready" {
		t.Fatalf("último evento = %q, want ready", last.Status)
	}
}

func TestHubRespeitaLimiteDeAssinaturas(t *testing.T) {
	hub := NewHub()
	defer hub.Close()

	subscriptions := make([]*Subscription, 0, maxPublicSubscribers)
	for index := 0; index < maxPublicSubscribers; index++ {
		subscription, ok := hub.Subscribe(TopicForPublic("publico-lotado"))
		if !ok {
			t.Fatalf("Subscribe #%d = false", index+1)
		}
		subscriptions = append(subscriptions, subscription)
	}
	if _, ok := hub.Subscribe(TopicForPublic("publico-lotado")); ok {
		t.Fatal("Subscribe acima do limite retornou true")
	}
	subscriptions[0].Close()
	subscription, ok := hub.Subscribe(TopicForPublic("publico-lotado"))
	if !ok {
		t.Fatal("Subscribe após liberação retornou false")
	}
	subscription.Close()
}

func TestHubCloseFechaCanaisEÉIdempotente(t *testing.T) {
	hub := NewHub()
	subscription, ok := hub.Subscribe(TopicAdmin)
	if !ok {
		t.Fatal("Subscribe(admin) = false")
	}

	hub.Close()
	hub.Close()
	subscription.Close()

	if _, open := <-subscription.Events; open {
		t.Fatal("canal deveria estar fechado após Close")
	}
}

func assertEvent(t *testing.T, events <-chan Event, want Event) {
	t.Helper()
	select {
	case got := <-events:
		if got != want {
			t.Fatalf("evento = %+v, want %+v", got, want)
		}
	default:
		t.Fatal("nenhum evento recebido")
	}
}

func assertNoEvent(t *testing.T, events <-chan Event) {
	t.Helper()
	select {
	case got := <-events:
		t.Fatalf("evento inesperado: %+v", got)
	default:
	}
}
