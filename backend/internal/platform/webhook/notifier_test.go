package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

const testStoreNumber = "+5585992358708"

type stubSettings struct {
	settings      domain.WebhookSettings
	storeNumber   string
	err           error
	noStoreNumber bool
}

func (s *stubSettings) GetNotification(context.Context) (domain.WebhookNotification, error) {
	storeNumber := s.storeNumber
	if !s.noStoreNumber && storeNumber == "" {
		storeNumber = testStoreNumber
	}
	return domain.WebhookNotification{Settings: s.settings, StoreNumber: storeNumber}, s.err
}

func testOrder() domain.Order {
	return domain.Order{
		Number:   1042,
		Status:   domain.OrderStatusConfirmed,
		Customer: domain.OrderCustomer{Name: "Cliente", Phone: "+5511999991234"},
		Items: []domain.OrderItem{{
			Name:           "Hambúrguer",
			Quantity:       1,
			UnitPriceCents: 2000,
			LineTotalCents: 2000,
		}},
		SubtotalCents: 2000,
		TotalCents:    2000,
	}
}

func TestDeliverPostsExactPayloadAndEventHeader(t *testing.T) {
	var received Payload
	var event string
	var method string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		event = r.Header.Get(eventHeader)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if err := json.Unmarshal(body, &received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	notifier, err := NewNotifier(
		&stubSettings{
			settings:    domain.WebhookSettings{URL: server.URL, Active: true},
			storeNumber: testStoreNumber,
		},
		slog.Default(),
		func(domain.WebhookEvent) string { return "Mensagem formatada para envio ao cliente" },
	)
	if err != nil {
		t.Fatalf("NewNotifier: %v", err)
	}

	notifier.deliver(context.Background(), domain.WebhookEvent{
		Type:  domain.WebhookEventOrderUpdated,
		Order: testOrder(),
	})

	if method != http.MethodPost {
		t.Fatalf("method = %q, want POST", method)
	}
	if event != domain.WebhookEventOrderUpdated {
		t.Fatalf("event = %q, want %q", event, domain.WebhookEventOrderUpdated)
	}
	if received.Number != "+5511999991234" {
		t.Fatalf("number = %q", received.Number)
	}
	if received.StoreNumber != testStoreNumber {
		t.Fatalf("storeNumber = %q, want %q", received.StoreNumber, testStoreNumber)
	}
	if received.Message != "Mensagem formatada para envio ao cliente" {
		t.Fatalf("message = %q", received.Message)
	}
}

func TestDeliverSkipsWhenStoreNumberUnavailable(t *testing.T) {
	var deliveries int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		deliveries++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	var logs bytes.Buffer
	notifier, err := NewNotifier(
		&stubSettings{
			settings:      domain.WebhookSettings{URL: server.URL, Active: true},
			noStoreNumber: true,
		},
		slog.New(slog.NewTextHandler(&logs, nil)),
		func(domain.WebhookEvent) string { return "mensagem" },
	)
	if err != nil {
		t.Fatalf("NewNotifier: %v", err)
	}
	notifier.deliver(context.Background(), domain.WebhookEvent{
		Type:  domain.WebhookEventOrderCreated,
		Order: testOrder(),
	})

	if deliveries != 0 {
		t.Fatalf("webhook incompleto foi entregue: %d chamada(s)", deliveries)
	}
	if !strings.Contains(logs.String(), "webhook skipped: store number unavailable") {
		t.Fatalf("falta do número da loja não foi registrada: %s", logs.String())
	}
	if strings.Contains(logs.String(), "+5511999991234") {
		t.Fatalf("log expôs telefone do cliente: %s", logs.String())
	}
}

func TestDeliverDoesNotFollowRedirects(t *testing.T) {
	var redirected int
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		redirected++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer destination.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer origin.Close()

	var logs bytes.Buffer
	notifier, err := NewNotifier(
		&stubSettings{settings: domain.WebhookSettings{URL: origin.URL, Active: true}},
		slog.New(slog.NewTextHandler(&logs, nil)),
		func(domain.WebhookEvent) string { return "mensagem" },
	)
	if err != nil {
		t.Fatalf("NewNotifier: %v", err)
	}
	notifier.deliver(context.Background(), domain.WebhookEvent{
		Type:  domain.WebhookEventOrderUpdated,
		Order: testOrder(),
	})

	if redirected != 0 {
		t.Fatalf("redirecionamento entregou PII a destino não configurado: %d chamada(s)", redirected)
	}
	if !strings.Contains(logs.String(), "webhook delivery failed") || strings.Contains(logs.String(), "webhook delivered") {
		t.Fatalf("redirecionamento foi registrado como sucesso: %s", logs.String())
	}
}

func TestDeliverSkipsInactiveOrMissingSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings domain.WebhookSettings
		err      error
	}{
		{name: "inativo", settings: domain.WebhookSettings{URL: "http://127.0.0.1:1", Active: false}},
		{name: "sem url", settings: domain.WebhookSettings{Active: true}},
		{name: "erro ao ler", err: errors.New("database unavailable")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			notifier, err := NewNotifier(
				&stubSettings{settings: test.settings, err: test.err},
				slog.Default(),
				func(domain.WebhookEvent) string { return "mensagem" },
			)
			if err != nil {
				t.Fatalf("NewNotifier: %v", err)
			}
			// Nenhum servidor é contatado; a chamada termina sem pânico e sem log de PII.
			notifier.deliver(context.Background(), domain.WebhookEvent{
				Type:  domain.WebhookEventOrderCreated,
				Order: testOrder(),
			})
		})
	}
}

func TestDeliverDoesNotFailOnTransportError(t *testing.T) {
	notifier, err := NewNotifier(
		&stubSettings{settings: domain.WebhookSettings{URL: "http://127.0.0.1:1", Active: true}},
		slog.Default(),
		func(domain.WebhookEvent) string { return "mensagem" },
	)
	if err != nil {
		t.Fatalf("NewNotifier: %v", err)
	}
	notifier.timeout = 50 * time.Millisecond

	started := time.Now()
	notifier.deliver(context.Background(), domain.WebhookEvent{
		Type:  domain.WebhookEventOrderCreated,
		Order: testOrder(),
	})
	if time.Since(started) > time.Second {
		t.Fatal("deliver bloqueou apesar de falha de transporte")
	}
}

func TestDeliverHonorsSelectedEventsAndKeepsGenericHeader(t *testing.T) {
	var events []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		events = append(events, r.Header.Get(eventHeader))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	notifier, err := NewNotifier(
		&stubSettings{settings: domain.WebhookSettings{
			URL: server.URL, Active: true,
			Events: []string{domain.WebhookEventKeyCreated, domain.WebhookEventKeyConfirmed},
		}},
		slog.Default(),
		func(domain.WebhookEvent) string { return "mensagem" },
	)
	if err != nil {
		t.Fatalf("NewNotifier: %v", err)
	}

	notifier.deliver(context.Background(), domain.WebhookEvent{Type: domain.WebhookEventOrderCreated, Order: testOrder()})
	notifier.deliver(context.Background(), domain.WebhookEvent{Type: domain.WebhookEventOrderUpdated, Order: testOrder()})
	notifier.deliver(context.Background(), domain.WebhookEvent{
		Type: domain.WebhookEventOrderUpdated,
		Order: domain.Order{
			Status: domain.OrderStatusPreparing,
		},
	})

	if len(events) != 2 {
		t.Fatalf("entregas = %d, want 2 (evento não selecionado deveria ser ignorado)", len(events))
	}
	if events[0] != domain.WebhookEventOrderCreated || events[1] != domain.WebhookEventOrderUpdated {
		t.Fatalf("cabeçalhos = %v, want created/updated", events)
	}
}

func TestDeliverWithLegacySettingsSendsEveryEvent(t *testing.T) {
	var deliveries int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		deliveries++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	notifier, err := NewNotifier(
		&stubSettings{settings: domain.WebhookSettings{URL: server.URL, Active: true}},
		slog.Default(),
		func(domain.WebhookEvent) string { return "mensagem" },
	)
	if err != nil {
		t.Fatalf("NewNotifier: %v", err)
	}
	notifier.deliver(context.Background(), domain.WebhookEvent{
		Type: domain.WebhookEventOrderUpdated, Order: testOrder(),
	})
	notifier.deliver(context.Background(), domain.WebhookEvent{
		Type: domain.WebhookEventOrderUpdated, Order: domain.Order{Status: "status_futuro"},
	})
	if deliveries != 2 {
		t.Fatalf("entregas = %d, want 2 (configuração legada envia qualquer status)", deliveries)
	}
}

func TestNewNotifierRequiresDependencies(t *testing.T) {
	if _, err := NewNotifier(nil, slog.Default(), func(domain.WebhookEvent) string { return "" }); err == nil {
		t.Fatal("esperava erro sem leitor de configuração")
	}
	if _, err := NewNotifier(&stubSettings{}, slog.Default(), nil); err == nil {
		t.Fatal("esperava erro sem construtor de mensagem")
	}
}
