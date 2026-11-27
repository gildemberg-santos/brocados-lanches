package repository

import (
	"reflect"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestWebhookEventsBSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		events  []string
		wantNil bool
	}{
		{name: "configuração legada", events: nil, wantNil: true},
		{name: "nenhum evento habilitado", events: []string{}, wantNil: false},
		{name: "seleção parcial", events: []string{domain.WebhookEventKeyCreated, domain.WebhookEventKeyCancelled}, wantNil: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := bson.Marshal(webhookSettingsDocument{
				URL: "https://hooks.exemplo.invalid", Active: true,
				Events: test.events, CreatedAt: now, UpdatedAt: now,
			})
			if err != nil {
				t.Fatalf("bson.Marshal: %v", err)
			}
			var decoded webhookSettingsDocument
			if err := bson.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("bson.Unmarshal: %v", err)
			}
			if (decoded.Events == nil) != test.wantNil {
				t.Fatalf("events após leitura = %#v, wantNil = %t", decoded.Events, test.wantNil)
			}
			settings := (storeSettingsDocument{Webhook: decoded}).toDomain()
			if (settings.Webhook.Events == nil) != test.wantNil {
				t.Fatalf("events no domínio = %#v, wantNil = %t", settings.Webhook.Events, test.wantNil)
			}
			if !reflect.DeepEqual(settings.Webhook.Events, test.events) {
				t.Fatalf("events = %v, want %v", settings.Webhook.Events, test.events)
			}
		})
	}
}

func TestNewStoreSettingsDocumentCopiesWebhookEvents(t *testing.T) {
	events := []string{domain.WebhookEventKeyConfirmed}
	settings := domain.StoreSettings{
		Name: "Brocados",
		Webhook: domain.WebhookSettings{
			URL: "https://hooks.exemplo.invalid", Active: true, Events: events,
		},
	}

	document := newStoreSettingsDocument(settings)
	events[0] = "alterada"
	if document.Webhook.Events[0] != domain.WebhookEventKeyConfirmed {
		t.Fatal("documento reutilizou a fatia da configuração recebida")
	}

	settings.Webhook.Events = []string{}
	emptyDocument := newStoreSettingsDocument(settings)
	if emptyDocument.Webhook.Events == nil {
		t.Fatal("seleção vazia foi convertida em configuração legada")
	}
	settings.Webhook.Events = nil
	legacyDocument := newStoreSettingsDocument(settings)
	if legacyDocument.Webhook.Events != nil {
		t.Fatal("configuração legada não foi preservada como nil")
	}
}
