package realtime

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// waitDrained espera o consumidor assíncrono processar pelo menos um evento.
// O runner do Go executa o corpo com N=1 antes da medição, então a asserção não
// pode depender de agendamento imediato da goroutine.
func waitDrained(drained *atomic.Int64) {
	deadline := time.Now().Add(200 * time.Millisecond)
	for drained.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
}

var benchmarkEvent = Event{
	Type:        EventOrderUpdated,
	PublicID:    "publico-bench",
	OrderNumber: 42,
	Status:      "preparing",
}

// BenchmarkHubSubscribeAndClose mede o custo de abrir e fechar uma assinatura,
// que é o par executado a cada conexão WebSocket recebida.
func BenchmarkHubSubscribeAndClose(b *testing.B) {
	hub := NewHub()
	b.Cleanup(hub.Close)

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		subscription, ok := hub.Subscribe(TopicAdmin)
		if !ok {
			b.Fatal("Subscribe(admin) = false")
		}
		subscription.Close()
	}
}

// BenchmarkHubPublish mede o fan-out para 1, 10 e 100 assinantes, com
// consumidor rápido (drena os canais) e lento (deixa o buffer descartar).
func BenchmarkHubPublish(b *testing.B) {
	for _, subscribers := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("assinantes=%d", subscribers), func(b *testing.B) {
			b.Run("consumidor-rapido", func(b *testing.B) {
				benchmarkPublishFastConsumer(b, subscribers)
			})
			b.Run("consumidor-lento", func(b *testing.B) {
				benchmarkPublishSlowConsumer(b, subscribers)
			})
		})
	}
}

func benchmarkPublishFastConsumer(b *testing.B, subscribers int) {
	hub := NewHub()
	subscriptions := make([]*Subscription, 0, subscribers)
	var drained atomic.Int64
	var waitGroup sync.WaitGroup

	for index := 0; index < subscribers; index++ {
		subscription, ok := hub.Subscribe(TopicAdmin)
		if !ok {
			b.Fatalf("Subscribe #%d = false", index+1)
		}
		subscriptions = append(subscriptions, subscription)
		waitGroup.Add(1)
		go func(events <-chan Event) {
			defer waitGroup.Done()
			for range events {
				drained.Add(1)
			}
		}(subscription.Events)
	}

	defer func() {
		for _, subscription := range subscriptions {
			subscription.Close()
		}
		waitGroup.Wait()
		hub.Close()
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		hub.Publish(TopicAdmin, benchmarkEvent)
	}
	b.StopTimer()

	waitDrained(&drained)
	if drained.Load() == 0 {
		b.Fatal("nenhum evento drenado pelos assinantes")
	}
}

func benchmarkPublishSlowConsumer(b *testing.B, subscribers int) {
	hub := NewHub()
	subscriptions := make([]*Subscription, 0, subscribers)
	for index := 0; index < subscribers; index++ {
		subscription, ok := hub.Subscribe(TopicAdmin)
		if !ok {
			b.Fatalf("Subscribe #%d = false", index+1)
		}
		subscriptions = append(subscriptions, subscription)
	}
	defer func() {
		for _, subscription := range subscriptions {
			subscription.Close()
		}
		hub.Close()
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		hub.Publish(TopicAdmin, benchmarkEvent)
	}
	b.StopTimer()

	// O buffer é limitado: publicações extras descartam as antigas sem bloquear.
	expected := b.N
	if expected > eventBufferSize {
		expected = eventBufferSize
	}
	if len(subscriptions[0].Events) != expected {
		b.Fatalf("buffer retido = %d, want %d", len(subscriptions[0].Events), expected)
	}
}

// BenchmarkHubPublishParallel mede a contenção do mutex do hub quando vários
// pedidos são publicados ao mesmo tempo em tópicos públicos diferentes.
func BenchmarkHubPublishParallel(b *testing.B) {
	const topics = 8

	hub := NewHub()
	subscriptions := make([]*Subscription, 0, topics)
	var drained atomic.Int64
	var waitGroup sync.WaitGroup

	for index := 0; index < topics; index++ {
		subscription, ok := hub.Subscribe(TopicForPublic(fmt.Sprintf("publico-bench-%d", index)))
		if !ok {
			b.Fatalf("Subscribe(tópico %d) = false", index)
		}
		subscriptions = append(subscriptions, subscription)
		waitGroup.Add(1)
		go func(events <-chan Event) {
			defer waitGroup.Done()
			for range events {
				drained.Add(1)
			}
		}(subscription.Events)
	}

	defer func() {
		for _, subscription := range subscriptions {
			subscription.Close()
		}
		waitGroup.Wait()
		hub.Close()
	}()

	var topicIndex atomic.Uint64
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		topic := int((topicIndex.Add(1) - 1) % topics)
		name := TopicForPublic(fmt.Sprintf("publico-bench-%d", topic))
		for pb.Next() {
			hub.Publish(name, benchmarkEvent)
		}
	})
	b.StopTimer()

	waitDrained(&drained)
	if drained.Load() == 0 {
		b.Fatal("nenhum evento drenado pelos assinantes")
	}
}

// BenchmarkHubSubscribeParallel mede a contenção em aberturas e fechamentos
// simultâneos no tópico administrativo, que tem limite maior que o público.
func BenchmarkHubSubscribeParallel(b *testing.B) {
	hub := NewHub()
	defer hub.Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			subscription, ok := hub.Subscribe(TopicAdmin)
			if !ok {
				b.Fatal("Subscribe(admin) = false")
			}
			subscription.Close()
		}
	})
}
